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
		Name:  "list",
		Usage: "List normalized session metadata",
		Flags: sessionQueryFlags(),
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
		Name:  "show",
		Usage: "Show one session generation by session or lineage identity",
		Flags: flags,
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
		Name:  "export",
		Usage: "Export sessions with a stable versioned schema",
		Flags: flags,
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

func loadSessionRecords(path string, includeContent bool) (string, []SessionLogLine, error) {
	file, err := os.Open(filepath.Join(path, "transcript.jsonl"))
	if err != nil {
		return "", nil, err
	}
	defer func() { _ = file.Close() }()

	var cwd string
	var records []SessionLogLine
	decoder := json.NewDecoder(file)
	for {
		var record SessionLogLine
		if err := decoder.Decode(&record); errors.Is(err, io.EOF) {
			break
		} else if err != nil {
			return "", nil, err
		}
		if cwd == "" && record.CWD != "" {
			cwd = record.CWD
		}
		if includeContent {
			records = append(records, record)
		}
	}
	return cwd, records, nil
}

func discoverSessionSummaries(includeContent bool) ([]SessionSummary, error) {
	root, err := sessionStoreRoot()
	if err != nil {
		return nil, err
	}
	if _, statErr := os.Stat(root); errors.Is(statErr, os.ErrNotExist) {
		return []SessionSummary{}, nil
	} else if statErr != nil {
		return nil, statErr
	}

	summaries := make([]SessionSummary, 0)
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
		summary := SessionSummary{
			Agent: manifest.Agent, SessionID: manifest.SessionID, LineageID: manifest.LineageID,
			GenerationID: filepath.Base(generationPath), SourceType: manifest.SourceType,
			IngestedAt: manifest.IngestedAt, HighWaterMark: manifest.HighWaterMark,
			Completeness: manifest.Completeness, RecordCount: manifest.RecordCount,
			MalformedRecords: manifest.MalformedRecords, SkippedRecords: manifest.SkippedRecords,
			sourceFingerprint: manifest.SourceFingerprint,
		}
		if manifest.SchemaVersion != sessionSchemaVersion || manifest.ParserVersion != sessionParserVersion {
			summary.Status = append(summary.Status, "unsupported")
		} else if validationErr := validateSessionGeneration(generationPath, manifest); validationErr != nil {
			summary.Status = append(summary.Status, "invalid")
		} else {
			summary.CWD, summary.Records, err = loadSessionRecords(generationPath, includeContent)
			if err != nil {
				summary.Status = append(summary.Status, "invalid")
			}
		}
		if manifest.Completeness == sessionPartial || manifest.MalformedRecords > 0 || manifest.SkippedRecords > 0 {
			summary.Status = append(summary.Status, "partial")
		}
		summaries = append(summaries, summary)
		return nil
	})
	if err != nil {
		return nil, err
	}
	markSessionGenerationStatus(summaries)
	slices.SortFunc(summaries, func(left, right SessionSummary) int {
		if result := strings.Compare(right.IngestedAt, left.IngestedAt); result != 0 {
			return result
		}
		return strings.Compare(left.LineageID+left.GenerationID, right.LineageID+right.GenerationID)
	})
	return summaries, nil
}

func markSessionGenerationStatus(summaries []SessionSummary) {
	latest := make(map[string]int)
	fingerprints := make(map[string]int)
	for index := range summaries {
		lineage := summaries[index].Agent + "\x00" + summaries[index].LineageID
		if current, ok := latest[lineage]; !ok || summaries[index].IngestedAt > summaries[current].IngestedAt {
			latest[lineage] = index
		}
		fingerprints[lineage+"\x00"+summaries[index].sourceFingerprint]++
	}
	for index := range summaries {
		lineage := summaries[index].Agent + "\x00" + summaries[index].LineageID
		if latest[lineage] != index {
			summaries[index].Status = append(summaries[index].Status, "stale")
		}
		if fingerprints[lineage+"\x00"+summaries[index].sourceFingerprint] > 1 {
			summaries[index].Status = append(summaries[index].Status, "duplicate")
		}
		if len(summaries[index].Status) == 0 {
			summaries[index].Status = []string{"current"}
		}
		slices.Sort(summaries[index].Status)
	}
}

func filterSessionSummaries(summaries []SessionSummary, query SessionQuery) []SessionSummary {
	filtered := make([]SessionSummary, 0, len(summaries))
	for _, summary := range summaries {
		if query.Agent != "" && summary.Agent != query.Agent || query.CWD != "" && summary.CWD != query.CWD {
			continue
		}
		if query.Identity != "" && summary.SessionID != query.Identity && summary.LineageID != query.Identity && summary.GenerationID != query.Identity {
			continue
		}
		ingested, err := time.Parse(time.RFC3339Nano, summary.IngestedAt)
		if err != nil && (!query.Since.IsZero() || !query.Until.IsZero()) {
			continue
		}
		if !query.Since.IsZero() && ingested.Before(query.Since) || !query.Until.IsZero() && ingested.After(query.Until) {
			continue
		}
		filtered = append(filtered, summary)
	}
	return filtered
}

func querySessionSummaries(query SessionQuery, includeContent bool) ([]SessionSummary, error) {
	summaries, err := discoverSessionSummaries(includeContent)
	if err != nil {
		return nil, err
	}
	return filterSessionSummaries(summaries, query), nil
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
