package dot

import (
	"strings"
	"testing"
	"unicode/utf8"
)

func TestPackDiffIncludesInventoryStatsAndTailFiles(t *testing.T) {
	diff := testDiff(
		"first.txt", "@@ -1 +1 @@\n-old\n+new\n"+strings.Repeat("+bulk\n", 40),
		"tail.txt", "@@ -1 +1 @@\n-before\n+after\n",
	)
	packed, err := PackDiff(diff, 700)
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"files: 2", "first.txt", "tail.txt", "added_lines: 42", "deleted_lines: 2", "diff --git a/tail.txt b/tail.txt"} {
		if !strings.Contains(packed, want) {
			t.Errorf("packed diff missing %q:\n%s", want, packed)
		}
	}
	if len(packed) > 700 {
		t.Fatalf("packed diff is %d bytes, budget is 700", len(packed))
	}
}

func TestPackDiffPrioritizesSecurityChanges(t *testing.T) {
	diff := testDiff(
		"large.txt", "@@ -1 +1 @@\n-old\n+new\n"+strings.Repeat("+bulk\n", 30),
		"auth/policy.go", "@@ -1 +1 @@\n-allowAll()\n+requirePermission()\n",
		"tail.txt", "@@ -1 +1 @@\n-old\n+new\n",
	)
	manifest, err := PackDiff(diff, 1)
	if err == nil || manifest != "" {
		t.Fatalf("budget below manifest must fail, payload=%q err=%v", manifest, err)
	}

	full, err := PackDiff(diff, DefaultMaxDiffSize)
	if err != nil {
		t.Fatal(err)
	}
	budget := strings.Index(full, "diff --git a/tail.txt")
	if budget <= 0 {
		t.Fatal("full payload does not contain tail file")
	}
	packed, err := PackDiff(diff, budget)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(packed, "diff --git a/auth/policy.go") {
		t.Fatalf("security-sensitive file was not prioritized:\n%s", packed)
	}
	if !strings.Contains(packed, "tail.txt |") || !strings.Contains(packed, "status=omitted-file") {
		t.Fatalf("omitted tail file was not explicit:\n%s", packed)
	}
}

func TestPackDiffMarksEveryOmittedHunkAndPreservesLines(t *testing.T) {
	diff := testDiff("unicode.txt", "@@ -1 +1 @@\n-a\n+é\n@@ -10 +10 @@\n-b\n+tail\n")
	full, err := PackDiff(diff, DefaultMaxDiffSize)
	if err != nil {
		t.Fatal(err)
	}
	budget := len(full) - len("@@ -10 +10 @@\n-b\n+tail\n")
	packed, err := PackDiff(diff, budget)
	if err != nil {
		t.Fatal(err)
	}
	if !utf8.ValidString(packed) || strings.Contains(packed, "@@ -10 +10 @@") {
		t.Fatalf("payload split or retained the omitted hunk:\n%s", packed)
	}
	if !strings.Contains(packed, "omitted_hunks=2") || !strings.Contains(packed, "omitted_bytes=23") {
		t.Fatalf("hunk omission is not auditable:\n%s", packed)
	}
}

func TestPackDiffRejectsInvalidInput(t *testing.T) {
	if _, err := PackDiff(string([]byte{0xff}), 100); err == nil {
		t.Fatal("invalid UTF-8 must fail")
	}
	if _, err := PackDiff("plain text", 100); err == nil {
		t.Fatal("non-diff input must fail")
	}
}

func TestPackDiffInventoriesQuotedDeletedPath(t *testing.T) {
	diff := "diff --git \"a/old file.txt\" \"b/old file.txt\"\n--- \"a/old file.txt\"\n+++ /dev/null\n@@ -1 +0,0 @@\n-old\n"
	packed, err := PackDiff(diff, DefaultMaxDiffSize)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(packed, "- old file.txt |") {
		t.Fatalf("deleted path was not decoded in inventory:\n%s", packed)
	}
}

func TestPackDiffInventoriesQuotedBinaryPath(t *testing.T) {
	diff := "diff --git \"a/image file.png\" \"b/image file.png\"\nindex 1111111..2222222 100644\nBinary files \"a/image file.png\" and \"b/image file.png\" differ\n"
	packed, err := PackDiff(diff, DefaultMaxDiffSize)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(packed, "- image file.png |") || !strings.Contains(packed, "status=complete") {
		t.Fatalf("binary path was not decoded in inventory:\n%s", packed)
	}
}

func testDiff(parts ...string) string {
	var out strings.Builder
	for i := 0; i < len(parts); i += 2 {
		path, body := parts[i], parts[i+1]
		out.WriteString("diff --git a/" + path + " b/" + path + "\n")
		out.WriteString("--- a/" + path + "\n+++ b/" + path + "\n")
		out.WriteString(body)
	}
	return out.String()
}
