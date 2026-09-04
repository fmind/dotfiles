package dot

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"github.com/urfave/cli/v3"
)

const sessionExportSchema = "dot.agent.sessions/v1"

type SessionQuery struct {
	Since    time.Time
	Until    time.Time
	Agent    string
	CWD      string
	Identity string
}

type SessionSummary struct {
	sourceFingerprint string
	Agent             string              `json:"agent"`
	SessionID         string              `json:"session_id"`
	LineageID         string              `json:"lineage_id"`
	GenerationID      string              `json:"generation_id"`
	CWD               string              `json:"cwd,omitempty"`
	SourceType        string              `json:"source_type"`
	IngestedAt        string              `json:"ingested_at"`
	HighWaterMark     string              `json:"high_water_mark,omitempty"`
	Completeness      sessionCompleteness `json:"completeness"`
	Records           []SessionLogLine    `json:"records,omitempty"`
	Status            []string            `json:"status"`
	RecordCount       int                 `json:"record_count"`
	MalformedRecords  int                 `json:"malformed_records"`
	SkippedRecords    int                 `json:"skipped_records"`
}

type SessionExport struct {
	Schema   string           `json:"schema"`
	Sessions []SessionSummary `json:"sessions"`
}

func NewAgentSessionListCmd(state *GlobalState) *cli.Command {
	return &cli.Command{
		Name:    "list",
		Aliases: []string{"l"},
		Usage:   "List normalized session metadata",
		Flags:   sessionQueryFlags(),
		Action: func(ctx context.Context, cmd *cli.Command) error {
			query, err := sessionQueryFromCommand(cmd)
			if err != nil {
				return err
			}
			return RunAgentSessionList(ctx, state, query)
		},
	}
}

func NewAgentSessionShowCmd(state *GlobalState) *cli.Command {
	flags := append(sessionQueryFlags(), &cli.BoolFlag{Name: "content", Usage: "Include prompt and response content"})
	return &cli.Command{
		Name:    "show",
		Aliases: []string{"w"},
		Usage:   "Show one session generation by session or lineage identity",
		Flags:   flags,
		Action: func(ctx context.Context, cmd *cli.Command) error {
			query, err := sessionQueryFromCommand(cmd)
			if err != nil {
				return err
			}
			if query.Identity == "" {
				query.Identity = cmd.Args().First()
			}
			if query.Identity == "" {
				return errors.New("show requires a session or lineage identity")
			}
			return RunAgentSessionShow(ctx, state, query, cmd.Bool("content"))
		},
	}
}

func NewAgentSessionExportCmd(state *GlobalState) *cli.Command {
	flags := append(sessionQueryFlags(),
		&cli.StringFlag{Name: "format", Value: "json", Usage: "Export format: json or ndjson"},
		&cli.BoolFlag{Name: "content", Usage: "Include prompt and response content"},
		&cli.BoolFlag{Name: "redact-content", Usage: "Include records with content replaced by [redacted]"},
	)
	return &cli.Command{
		Name:    "export",
		Aliases: []string{"e"},
		Usage:   "Export sessions with a stable versioned schema",
		Flags:   flags,
		Action: func(ctx context.Context, cmd *cli.Command) error {
			query, err := sessionQueryFromCommand(cmd)
			if err != nil {
				return err
			}
			if cmd.Bool("content") && cmd.Bool("redact-content") {
				return errors.New("--content and --redact-content are mutually exclusive")
			}
			return RunAgentSessionExport(ctx, state, query, cmd.String("format"), cmd.Bool("content"), cmd.Bool("redact-content"))
		},
	}
}

func sessionQueryFlags() []cli.Flag {
	return []cli.Flag{
		&cli.StringFlag{Name: "agent", Usage: "Filter by agent"},
		&cli.StringFlag{Name: "cwd", Aliases: []string{"project"}, Usage: "Filter by exact project/CWD"},
		&cli.StringFlag{Name: "session", Usage: "Filter by session or lineage identity"},
		&cli.StringFlag{Name: "since", Usage: "Include ingestions on or after RFC3339 timestamp or YYYY-MM-DD"},
		&cli.StringFlag{Name: "until", Usage: "Include ingestions on or before RFC3339 timestamp or YYYY-MM-DD"},
	}
}

func sessionQueryFromCommand(cmd *cli.Command) (SessionQuery, error) {
	query := SessionQuery{Agent: cmd.String("agent"), CWD: cmd.String("cwd"), Identity: cmd.String("session")}
	var err error
	if query.Since, err = parseSessionDate(cmd.String("since"), false); err != nil {
		return SessionQuery{}, fmt.Errorf("invalid --since: %w", err)
	}
	if query.Until, err = parseSessionDate(cmd.String("until"), true); err != nil {
		return SessionQuery{}, fmt.Errorf("invalid --until: %w", err)
	}
	if !query.Since.IsZero() && !query.Until.IsZero() && query.Since.After(query.Until) {
		return SessionQuery{}, errors.New("--since must not be after --until")
	}
	return query, nil
}

func parseSessionDate(value string, endOfDay bool) (time.Time, error) {
	if value == "" {
		return time.Time{}, nil
	}
	if parsed, err := time.Parse(time.RFC3339, value); err == nil {
		return parsed, nil
	}
	parsed, err := time.Parse("2006-01-02", value)
	if err != nil {
		return time.Time{}, errors.New("expected RFC3339 or YYYY-MM-DD")
	}
	if endOfDay {
		parsed = parsed.Add(24*time.Hour - time.Nanosecond)
	}
	return parsed, nil
}

func requireOwnerOnly(path string, entry fs.DirEntry) error {
	if entry.Type()&os.ModeSymlink != 0 {
		return fmt.Errorf("session store contains a symbolic link: %s", path)
	}
	info, err := entry.Info()
	if err != nil {
		return err
	}
	if info.Mode().Perm()&0o077 != 0 {
		return fmt.Errorf("session store path is not owner-only: %s", path)
	}
	return nil
}

// sessionRecordsCWD reports the working directory of a generation: the first one
// any record carries, matching how the transcript was written.
func sessionRecordsCWD(records []SessionLogLine) string {
	for _, record := range records {
		if record.CWD != "" {
			return record.CWD
		}
	}
	return ""
}

// sessionGeneration pairs a stored generation with its manifest. Every field the
// cheap query predicates need lives in the manifest, so the walk can decide what
// to keep before anything reads, hashes and decodes a transcript.
type sessionGeneration struct {
	path     string
	manifest sessionManifest
	summary  SessionSummary
}

// discoverSessionGenerations walks the store and reads manifests only. This is the
// cheap half of a query: it touches one small JSON per generation instead of the
// whole transcript.
func discoverSessionGenerations() ([]sessionGeneration, error) {
	root, err := sessionStoreRoot()
	if err != nil {
		return nil, err
	}
	if _, statErr := os.Stat(root); errors.Is(statErr, os.ErrNotExist) {
		return []sessionGeneration{}, nil
	} else if statErr != nil {
		return nil, statErr
	}

	generations := make([]sessionGeneration, 0)
	err = filepath.WalkDir(root, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if permissionErr := requireOwnerOnly(path, entry); permissionErr != nil {
			return permissionErr
		}
		if entry.IsDir() || entry.Name() != "manifest.json" {
			return nil
		}
		generationPath := filepath.Dir(path)
		manifest, manifestErr := readSessionManifest(generationPath)
		if manifestErr != nil {
			return fmt.Errorf("failed to read session manifest %s: %w", path, manifestErr)
		}
		generations = append(generations, sessionGeneration{
			path:     generationPath,
			manifest: manifest,
			summary: SessionSummary{
				Agent: manifest.Agent, SessionID: manifest.SessionID, LineageID: manifest.LineageID,
				GenerationID: filepath.Base(generationPath), SourceType: manifest.SourceType,
				IngestedAt: manifest.IngestedAt, HighWaterMark: manifest.HighWaterMark,
				Completeness: manifest.Completeness, RecordCount: manifest.RecordCount,
				MalformedRecords: manifest.MalformedRecords, SkippedRecords: manifest.SkippedRecords,
				sourceFingerprint: manifest.SourceFingerprint,
			},
		})
		return nil
	})
	if err != nil {
		return nil, err
	}
	return generations, nil
}

// sessionLineageIndex records which generation is newest in each lineage and which
// source fingerprints repeat. It is built from every generation in the store, not
// just the queried subset, so a filter can never promote a stale generation to
// "current" by hiding its newer sibling.
type sessionLineageIndex struct {
	newest       map[string]string
	fingerprints map[string]int
}

func newSessionLineageIndex(generations []sessionGeneration) sessionLineageIndex {
	index := sessionLineageIndex{
		newest:       make(map[string]string, len(generations)),
		fingerprints: make(map[string]int, len(generations)),
	}
	newestIngestedAt := make(map[string]string, len(generations))
	for _, generation := range generations {
		lineage := generation.summary.Agent + "\x00" + generation.summary.LineageID
		if current, ok := newestIngestedAt[lineage]; !ok || generation.summary.IngestedAt > current {
			newestIngestedAt[lineage] = generation.summary.IngestedAt
			index.newest[lineage] = generation.summary.GenerationID
		}
		index.fingerprints[lineage+"\x00"+generation.summary.sourceFingerprint]++
	}
	return index
}

// apply adds the lineage-wide status flags to one summary. It runs after the
// per-generation flags so "current" still means "nothing else to report".
func (index sessionLineageIndex) apply(summary *SessionSummary) {
	lineage := summary.Agent + "\x00" + summary.LineageID
	if index.newest[lineage] != summary.GenerationID {
		summary.Status = append(summary.Status, "stale")
	}
	if index.fingerprints[lineage+"\x00"+summary.sourceFingerprint] > 1 {
		summary.Status = append(summary.Status, "duplicate")
	}
	if len(summary.Status) == 0 {
		summary.Status = []string{"current"}
	}
	slices.Sort(summary.Status)
}

// matchesManifestQuery applies every predicate that the manifest alone can decide.
// CWD is deliberately absent: it only exists inside the transcript.
func matchesManifestQuery(summary SessionSummary, query SessionQuery) bool {
	if query.Agent != "" && summary.Agent != query.Agent {
		return false
	}
	if query.Identity != "" && summary.SessionID != query.Identity && summary.LineageID != query.Identity && summary.GenerationID != query.Identity {
		return false
	}
	if query.Since.IsZero() && query.Until.IsZero() {
		return true
	}
	ingested, err := time.Parse(time.RFC3339Nano, summary.IngestedAt)
	if err != nil {
		return false
	}
	if !query.Since.IsZero() && ingested.Before(query.Since) {
		return false
	}
	return query.Until.IsZero() || !ingested.After(query.Until)
}

func querySessionSummaries(query SessionQuery, includeContent bool) ([]SessionSummary, error) {
	generations, err := discoverSessionGenerations()
	if err != nil {
		return nil, err
	}
	// Built over every generation, then applied only to the ones that survive the
	// filter, so lineage status stays independent of the query.
	lineage := newSessionLineageIndex(generations)

	summaries := make([]SessionSummary, 0, len(generations))
	for _, generation := range generations {
		if !matchesManifestQuery(generation.summary, query) {
			continue
		}
		summary := generation.summary
		manifest := generation.manifest
		// Only now, for a generation the query actually wants, is it worth reading,
		// hashing and decoding the transcript.
		if manifest.SchemaVersion != sessionSchemaVersion || manifest.ParserVersion != sessionParserVersion {
			summary.Status = append(summary.Status, "unsupported")
		} else if records, validationErr := validateSessionGenerationRecords(generation.path, manifest); validationErr != nil {
			summary.Status = append(summary.Status, "invalid")
		} else {
			summary.CWD = sessionRecordsCWD(records)
			if includeContent {
				summary.Records = records
			}
		}
		if manifest.Completeness == sessionPartial || manifest.MalformedRecords > 0 || manifest.SkippedRecords > 0 {
			summary.Status = append(summary.Status, "partial")
		}
		lineage.apply(&summary)
		if query.CWD != "" && summary.CWD != query.CWD {
			continue
		}
		summaries = append(summaries, summary)
	}

	slices.SortFunc(summaries, func(left, right SessionSummary) int {
		if result := strings.Compare(right.IngestedAt, left.IngestedAt); result != 0 {
			return result
		}
		return strings.Compare(left.LineageID+left.GenerationID, right.LineageID+right.GenerationID)
	})
	return summaries, nil
}

func RunAgentSessionList(_ context.Context, state *GlobalState, query SessionQuery) error {
	summaries, err := querySessionSummaries(query, false)
	if err != nil {
		return err
	}
	for _, summary := range summaries {
		_, _ = fmt.Fprintf(state.Stdout, "%s %s %s records=%d status=%s cwd=%s\n", summary.IngestedAt, summary.Agent, summary.SessionID, summary.RecordCount, strings.Join(summary.Status, ","), summary.CWD)
	}
	return nil
}

func RunAgentSessionShow(_ context.Context, state *GlobalState, query SessionQuery, includeContent bool) error {
	summaries, err := querySessionSummaries(query, includeContent)
	if err != nil {
		return err
	}
	if len(summaries) == 0 {
		return errors.New("session not found")
	}
	if len(summaries) > 1 {
		return ambiguousSessionError(summaries)
	}
	return writeSessionJSON(state.Stdout, summaries[0])
}

// ambiguousSessionErrorSamples caps how many generation identities the ambiguity
// error prints, so one long-lived session cannot bury the instruction in hashes.
const ambiguousSessionErrorSamples = 3

// ambiguousSessionError explains what would actually narrow this result. Suggesting
// --agent unconditionally is wrong whenever the matches are generations of one
// session: they all share an agent, so the flag the message asks for cannot change
// the outcome. Offer the agent filter only when the candidates really do span agents,
// and otherwise name the generation identities the caller can pick from.
func ambiguousSessionError(summaries []SessionSummary) error {
	agents := make([]string, 0, len(summaries))
	for _, summary := range summaries {
		if !slices.Contains(agents, summary.Agent) {
			agents = append(agents, summary.Agent)
		}
	}
	if len(agents) > 1 {
		slices.Sort(agents)
		return fmt.Errorf("session identity is ambiguous: %d matches across agents %s; add --agent or use a generation identity",
			len(summaries), strings.Join(agents, ", "))
	}

	// Newest first, matching how list presents them, so the samples are the ones a
	// caller is most likely to want.
	generations := make([]string, 0, len(summaries))
	for _, summary := range summaries {
		generations = append(generations, summary.GenerationID)
	}
	shown := min(len(generations), ambiguousSessionErrorSamples)
	suffix := ""
	if len(generations) > shown {
		suffix = fmt.Sprintf(" (%d more)", len(generations)-shown)
	}
	return fmt.Errorf("session identity is ambiguous: %d generations of this %s session; pass --session with a generation identity, e.g. %s%s",
		len(summaries), agents[0], strings.Join(generations[:shown], ", "), suffix)
}

func RunAgentSessionExport(_ context.Context, state *GlobalState, query SessionQuery, format string, includeContent, redactContent bool) error {
	summaries, err := querySessionSummaries(query, includeContent || redactContent)
	if err != nil {
		return err
	}
	if redactContent {
		for summaryIndex := range summaries {
			for recordIndex := range summaries[summaryIndex].Records {
				summaries[summaryIndex].Records[recordIndex].Content = "[redacted]"
			}
		}
	}
	switch format {
	case "json":
		return writeSessionJSON(state.Stdout, SessionExport{Schema: sessionExportSchema, Sessions: summaries})
	case "ndjson":
		encoder := json.NewEncoder(state.Stdout)
		for _, summary := range summaries {
			if err := encoder.Encode(struct {
				Schema  string         `json:"schema"`
				Session SessionSummary `json:"session"`
			}{Schema: sessionExportSchema, Session: summary}); err != nil {
				return err
			}
		}
		return nil
	default:
		return fmt.Errorf("unsupported export format %q: expected json or ndjson", format)
	}
}

func writeSessionJSON(writer io.Writer, value any) error {
	encoder := json.NewEncoder(writer)
	encoder.SetIndent("", "  ")
	return encoder.Encode(value)
}
