package dot

import (
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"unicode/utf8"
)

type diffHunk struct {
	text string
}

type diffFile struct {
	path     string
	preamble string
	hunks    []diffHunk
	added    int
	deleted  int
}

// PackDiff turns a unified Git diff into an auditable payload within maxSize.
// The complete inventory is reserved first; patch units then enter in a stable,
// security-first round-robin so later files are represented before any file dominates.
func PackDiff(diff string, maxSize int) (string, error) {
	if maxSize <= 0 {
		maxSize = DefaultMaxDiffSize
	}
	if !utf8.ValidString(diff) {
		return "", errors.New("diff is not valid UTF-8")
	}

	files, err := parseDiff(diff)
	if err != nil {
		return "", err
	}
	selected := make([]map[int]bool, len(files))
	for i := range selected {
		selected[i] = make(map[int]bool)
	}

	payload := renderPackedDiff(diff, files, selected)
	if len(payload) > maxSize {
		return "", fmt.Errorf("diff budget %d bytes cannot hold the %d-byte inventory and omission manifest", maxSize, len(payload))
	}

	for _, candidate := range prioritizedDiffUnits(files) {
		selected[candidate.file][candidate.unit] = true
		trial := renderPackedDiff(diff, files, selected)
		if len(trial) <= maxSize {
			payload = trial
			continue
		}
		delete(selected[candidate.file], candidate.unit)
	}
	return payload, nil
}

type diffUnit struct {
	file int
	unit int
}

func prioritizedDiffUnits(files []diffFile) []diffUnit {
	var security, fair, remaining []diffUnit
	seen := make(map[diffUnit]bool)
	maxUnits := 0
	for i, file := range files {
		unitCount := max(1, len(file.hunks))
		maxUnits = max(maxUnits, unitCount)
		for unit := range unitCount {
			candidate := diffUnit{file: i, unit: unit}
			if isSecuritySensitive(file, unit) {
				security = append(security, candidate)
				seen[candidate] = true
			}
		}
	}
	for unit := range maxUnits {
		for file := range files {
			if unit >= max(1, len(files[file].hunks)) {
				continue
			}
			candidate := diffUnit{file: file, unit: unit}
			if seen[candidate] {
				continue
			}
			if unit == 0 {
				fair = append(fair, candidate)
			} else {
				remaining = append(remaining, candidate)
			}
		}
	}
	return append(append(security, fair...), remaining...)
}

func isSecuritySensitive(file diffFile, unit int) bool {
	text := strings.ToLower(file.path)
	if len(file.hunks) > 0 {
		text += "\n" + strings.ToLower(file.hunks[unit].text)
	}
	for _, marker := range []string{"auth", "credential", "permission", "secret", "security", ".github/workflows/"} {
		if strings.Contains(text, marker) {
			return true
		}
	}
	return false
}

func parseDiff(diff string) ([]diffFile, error) {
	var files []diffFile
	for _, line := range strings.SplitAfter(diff, "\n") {
		if strings.HasPrefix(line, "diff --git ") {
			files = append(files, diffFile{path: pathFromDiffHeader(line), preamble: line})
			continue
		}
		if len(files) == 0 {
			if strings.TrimSpace(line) == "" {
				continue
			}
			return nil, errors.New("input is not a unified Git diff")
		}
		file := &files[len(files)-1]
		if strings.HasPrefix(line, "@@ ") || strings.HasPrefix(line, "@@@ ") {
			file.hunks = append(file.hunks, diffHunk{text: line})
			continue
		}
		if len(file.hunks) == 0 {
			file.preamble += line
			if path := pathFromMarker(line); path != "" {
				file.path = path
			}
			continue
		}
		file.hunks[len(file.hunks)-1].text += line
		if strings.HasPrefix(line, "+") && !strings.HasPrefix(line, "+++") {
			file.added++
		}
		if strings.HasPrefix(line, "-") && !strings.HasPrefix(line, "---") {
			file.deleted++
		}
	}
	if len(files) == 0 {
		return nil, errors.New("input contains no changed files")
	}
	return files, nil
}

func pathFromDiffHeader(line string) string {
	line = strings.TrimSpace(strings.TrimPrefix(line, "diff --git "))
	if split := strings.LastIndex(line, " \"b/"); split >= 0 {
		return decodeDiffPath(line[split+1:])
	}
	if split := strings.LastIndex(line, " b/"); split >= 0 {
		return decodeDiffPath(line[split+1:])
	}
	return line
}

func pathFromMarker(line string) string {
	marker := ""
	if strings.HasPrefix(line, "+++ ") {
		marker = "+++ "
	} else if strings.HasPrefix(line, "--- ") {
		marker = "--- "
	} else {
		return ""
	}
	path := strings.TrimSpace(strings.TrimPrefix(line, marker))
	if path == "/dev/null" {
		return ""
	}
	return decodeDiffPath(path)
}

func decodeDiffPath(path string) string {
	if strings.HasPrefix(path, "\"") {
		if decoded, err := strconv.Unquote(path); err == nil {
			path = decoded
		}
	}
	return strings.TrimPrefix(strings.TrimPrefix(path, "a/"), "b/")
}

func renderPackedDiff(original string, files []diffFile, selected []map[int]bool) string {
	var added, deleted int
	for _, file := range files {
		added += file.added
		deleted += file.deleted
	}

	var out strings.Builder
	fmt.Fprintf(&out, "# Diff summary\nfiles: %d\nadded_lines: %d\ndeleted_lines: %d\noriginal_bytes: %d\n\n# Changed files\n", len(files), added, deleted, len(original))
	for i, file := range files {
		includedBytes := 0
		if len(selected[i]) > 0 {
			includedBytes += len(file.preamble)
		}
		omittedHunks := make([]int, 0, len(file.hunks))
		for hunk, content := range file.hunks {
			if selected[i][hunk] {
				includedBytes += len(content.text)
			} else {
				omittedHunks = append(omittedHunks, hunk+1)
			}
		}
		totalBytes := len(file.preamble)
		for _, hunk := range file.hunks {
			totalBytes += len(hunk.text)
		}
		status := "partial"
		if len(selected[i]) == 0 {
			status = "omitted-file"
		} else if totalBytes == includedBytes {
			status = "complete"
		}
		fmt.Fprintf(&out, "- %s | +%d -%d | status=%s | omitted_hunks=%s | omitted_bytes=%d\n", file.path, file.added, file.deleted, status, formatIntRanges(omittedHunks), totalBytes-includedBytes)
	}
	out.WriteString("\n# Packed patch\n")
	for i, file := range files {
		if len(selected[i]) == 0 {
			continue
		}
		out.WriteString(file.preamble)
		for hunk, content := range file.hunks {
			if selected[i][hunk] {
				out.WriteString(content.text)
			}
		}
	}
	return out.String()
}

func formatIntRanges(values []int) string {
	if len(values) == 0 {
		return "none"
	}
	sort.Ints(values)
	parts := make([]string, 0, len(values))
	for start := 0; start < len(values); {
		end := start
		for end+1 < len(values) && values[end+1] == values[end]+1 {
			end++
		}
		if start == end {
			parts = append(parts, strconv.Itoa(values[start]))
		} else {
			parts = append(parts, fmt.Sprintf("%d-%d", values[start], values[end]))
		}
		start = end + 1
	}
	return strings.Join(parts, ",")
}
