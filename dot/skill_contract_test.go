package dot

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	stdhtml "html"
	"io"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"sort"
	"strings"
	"syscall"
	"testing"
	"time"
	"unicode"
	"unicode/utf8"

	"github.com/urfave/cli/v3"
	"github.com/yuin/goldmark"
	markdownast "github.com/yuin/goldmark/ast"
	markdowntext "github.com/yuin/goldmark/text"
	nethtml "golang.org/x/net/html"
	"gopkg.in/yaml.v3"
)

const (
	maxSkillCompatibility     = 500
	maxSkillDescriptionLength = 240
	maxSkillLines             = 500
	maxSkillNameLength        = 64
	maxSkillParsedFileBytes   = 1 << 20
	maxSkillShortDescription  = 64
	minSkillShortDescription  = 25
)

type skillManifest struct {
	Skills  map[string][]string `json:"skills"`
	Version int                 `json:"version"`
}

var (
	remoteLiteralPattern = regexp.MustCompile(`(?i)\b(?:https?|mailto)(?::|\\:|&#(?:0*58|x0*3a);)[^\s<>"']+`)
	remoteURLPattern     = regexp.MustCompile(`(?i)\b(?:https?|mailto):[^\s<>"']+`)
	brandColorPattern    = regexp.MustCompile(`^#[0-9A-Fa-f]{6}$`)
	skillNamePattern     = regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)
	toolNamePattern      = regexp.MustCompile(`^[a-z0-9][a-z0-9+.-]*$`)
)

type skillContractFrontmatter struct {
	AllowedTools  string            `yaml:"allowed-tools"`
	Compatibility string            `yaml:"compatibility"`
	Description   string            `yaml:"description"`
	License       string            `yaml:"license"`
	Metadata      map[string]string `yaml:"metadata"`
	Name          string            `yaml:"name"`
}

type skillOpenAIConfig struct {
	Interface    skillOpenAIInterface    `yaml:"interface"`
	Policy       skillOpenAIPolicy       `yaml:"policy"`
	Dependencies skillOpenAIDependencies `yaml:"dependencies"`
}

type skillOpenAIDependencies struct {
	Tools []skillOpenAITool `yaml:"tools"`
}

type skillOpenAIInterface struct {
	BrandColor       string `yaml:"brand_color"`
	DefaultPrompt    string `yaml:"default_prompt"`
	DisplayName      string `yaml:"display_name"`
	IconLarge        string `yaml:"icon_large"`
	IconSmall        string `yaml:"icon_small"`
	ShortDescription string `yaml:"short_description"`
}

type skillOpenAIPolicy struct {
	AllowImplicitInvocation *bool `yaml:"allow_implicit_invocation"`
}

type skillOpenAITool struct {
	Description string `yaml:"description"`
	Transport   string `yaml:"transport"`
	Type        string `yaml:"type"`
	URL         string `yaml:"url"`
	Value       string `yaml:"value"`
}

func TestSkillContracts(t *testing.T) {
	repo := skillRepositoryRoot(t)
	manifest := readSkillManifest(t, repo)
	if manifest.Version != 1 {
		t.Fatalf("skills/contracts.json: unsupported version %d; replace with 1", manifest.Version)
	}

	skillFiles, discoveryFindings := discoverSkillFiles(repo)
	for _, finding := range discoveryFindings {
		t.Error(finding)
	}
	globalCount := 0
	for _, skillFile := range skillFiles {
		relative := skillRelativePath(repo, skillFile)
		if strings.HasPrefix(relative, "skills/") {
			globalCount++
		}
	}
	checkSkillCountDocumentation(t, repo, globalCount)
	seen := make(map[string]string, len(skillFiles))
	for _, skillFile := range skillFiles {
		name := filepath.Base(filepath.Dir(skillFile))
		if previous, duplicate := seen[name]; duplicate {
			t.Errorf("duplicate skill name %q in %s and %s; global and repository-scoped discovery must stay unambiguous", name, previous, skillFile)
			continue
		}
		seen[name] = skillFile
		tools, ok := manifest.Skills[name]
		if !ok {
			t.Errorf("skills/contracts.json: missing required-tools declaration for skill %q; add an explicit array, including [] when no CLI is required", name)
			continue
		}
		checkSkillContract(t, repo, skillFile, tools)
	}
	for name := range manifest.Skills {
		if _, ok := seen[name]; !ok {
			t.Errorf("skills/contracts.json: unknown skill %q; remove it or add SKILL.md under skills/ or .agents/skills/", name)
		}
	}
}

func TestSkillContractManifestDecoding(t *testing.T) {
	valid := []byte(`{"version":1,"skills":{"fixture":[]}}`)
	manifest, err := decodeSkillManifest(valid)
	if err != nil {
		t.Fatalf("valid manifest: %v", err)
	}
	if manifest.Version != 1 || !slices.Equal(manifest.Skills["fixture"], []string{}) {
		t.Fatalf("decoded manifest = %#v", manifest)
	}

	for _, test := range []struct {
		name string
		data string
	}{
		{name: "unknown field", data: `{"version":1,"skills":{},"skils":{}}`},
		{name: "second document", data: `{"version":1,"skills":{}} {"version":1,"skills":{}}`},
	} {
		t.Run(test.name, func(t *testing.T) {
			if _, err := decodeSkillManifest([]byte(test.data)); err == nil {
				t.Fatal("invalid manifest decoded without an error")
			}
		})
	}
}

func TestSkillContractParsedContentLimit(t *testing.T) {
	if _, err := readSkillParsedContent(bytes.NewReader(bytes.Repeat([]byte("x"), maxSkillParsedFileBytes))); err != nil {
		t.Fatalf("content at parsed-file limit: %v", err)
	}
	if _, err := readSkillParsedContent(bytes.NewReader(bytes.Repeat([]byte("x"), maxSkillParsedFileBytes+1))); err == nil {
		t.Fatal("oversized parsed content was accepted")
	}
}

func discoverSkillFiles(repo string) ([]string, []string) {
	skillFiles := make([]string, 0)
	findings := make([]string, 0)
	for _, root := range []string{filepath.Join(repo, "skills"), filepath.Join(repo, ".agents", "skills")} {
		err := filepath.WalkDir(root, func(path string, entry os.DirEntry, walkErr error) error {
			if walkErr != nil {
				return walkErr
			}
			if entry.Type()&os.ModeSymlink != 0 {
				findings = append(findings, skillRelativePath(repo, path)+": symbolic link is not allowed in first-party skill discovery")
				return nil
			}
			if entry.IsDir() || entry.Name() != "SKILL.md" {
				return nil
			}
			relative := filepath.ToSlash(skillRelativePath(repo, path))
			parts := strings.Split(relative, "/")
			validGlobal := len(parts) == 3 && parts[0] == "skills" && parts[2] == "SKILL.md"
			validRepository := len(parts) == 4 && parts[0] == ".agents" && parts[1] == "skills" && parts[3] == "SKILL.md"
			if !validGlobal && !validRepository {
				findings = append(findings, relative+": unsupported nested skill layout; use skills/<name>/SKILL.md or .agents/skills/<name>/SKILL.md")
				return nil
			}
			skillFiles = append(skillFiles, path)
			return nil
		})
		if err != nil && !errors.Is(err, os.ErrNotExist) {
			findings = append(findings, fmt.Sprintf("%s: discover skills: %v", skillRelativePath(repo, root), err))
		}
	}
	sort.Strings(skillFiles)
	sort.Strings(findings)
	return skillFiles, findings
}

func TestSkillContractRejectsUnsafePackageState(t *testing.T) {
	tests := []struct {
		configure func(*testing.T, string)
		name      string
		resource  string
		want      string
		content   []byte
		mode      os.FileMode
	}{
		{
			name:     "python bytecode cache",
			resource: "scripts/__pycache__/payload.pyc",
			content:  []byte("generated bytecode"),
			mode:     0o644,
			want:     "Python bytecode cache",
		},
		{
			name:     "tool cache",
			resource: "tests/.pytest_cache/README.md",
			content:  []byte("generated test cache"),
			mode:     0o644,
			want:     "generated cache or metadata",
		},
		{
			name:     "bidirectional marker",
			resource: "references/note.md",
			content:  []byte("safe prefix\u202Ehidden suffix\n"),
			mode:     0o644,
			want:     "bidirectional or invisible Unicode",
		},
		{
			name:     "terminal escape",
			resource: "references/note.md",
			content:  []byte("safe prefix\x1b[2Jspoofed suffix\n"),
			mode:     0o644,
			want:     "unsafe control character",
		},
		{
			name:     "binary script",
			resource: "scripts/payload.py",
			content:  []byte{'p', 'r', 'i', 'n', 't', 0, '(', ')'},
			mode:     0o755,
			want:     "non-text executable or instruction resource",
		},
		{
			name:     "oversized parsed resource",
			resource: "references/note.md",
			content:  bytes.Repeat([]byte("x"), maxSkillParsedFileBytes+1),
			mode:     0o644,
			want:     "exceeds the 1048576-byte parsed-file limit",
		},
		{
			name:     "executable outside scripts",
			resource: "assets/payload",
			content:  []byte("#!/bin/sh\nexit 0\n"),
			mode:     0o755,
			want:     "executable outside scripts/",
		},
		{
			name:     "symbolic link",
			resource: "scripts/payload.py",
			configure: func(t *testing.T, path string) {
				t.Helper()
				target := filepath.Join(filepath.Dir(filepath.Dir(path)), "outside.py")
				if err := os.WriteFile(target, []byte("print('outside')\n"), 0o644); err != nil {
					t.Fatal(err)
				}
				if err := os.Symlink(target, path); err != nil {
					t.Fatal(err)
				}
			},
			want: "symbolic link",
		},
		{
			name:     "non-regular resource",
			resource: "scripts/runtime.pipe",
			configure: func(t *testing.T, path string) {
				t.Helper()
				if err := syscall.Mkfifo(path, 0o600); err != nil {
					t.Fatal(err)
				}
				started := make(chan struct{})
				written := make(chan error, 1)
				go func() {
					close(started)
					file, err := os.OpenFile(path, os.O_WRONLY, 0)
					if err == nil {
						_, err = file.WriteString("runtime pipe\n")
						if closeErr := file.Close(); err == nil {
							err = closeErr
						}
					}
					written <- err
				}()
				t.Cleanup(func() {
					<-started
					select {
					case err := <-written:
						if err != nil {
							t.Errorf("write named pipe: %v", err)
						}
						return
					default:
					}
					file, err := os.OpenFile(path, os.O_RDONLY|syscall.O_NONBLOCK, 0)
					if err != nil {
						t.Errorf("open named pipe for cleanup: %v", err)
						return
					}
					if err := <-written; err != nil {
						t.Errorf("write named pipe: %v", err)
					}
					if err := file.Close(); err != nil {
						t.Errorf("close named pipe: %v", err)
					}
				})
			},
			want: "non-regular resource",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			repo := t.TempDir()
			skillDirectory := filepath.Join(repo, "skills", "fixture")
			resourcePath := filepath.Join(skillDirectory, filepath.FromSlash(test.resource))
			if err := os.MkdirAll(filepath.Dir(resourcePath), 0o755); err != nil {
				t.Fatal(err)
			}
			skillFile := filepath.Join(skillDirectory, "SKILL.md")
			skill := "---\nname: fixture\ndescription: Exercise the package contract with an isolated unsafe resource. Use for contract testing only.\nlicense: MIT\n---\n\n# Fixture\n\nUse " + test.resource + ".\n"
			if err := os.WriteFile(skillFile, []byte(skill), 0o644); err != nil {
				t.Fatal(err)
			}
			if test.configure != nil {
				test.configure(t, resourcePath)
			} else if err := os.WriteFile(resourcePath, test.content, test.mode); err != nil {
				t.Fatal(err)
			}

			assertSkillFinding(t, skillContractFindings(repo, skillFile, nil), test.want)
		})
	}
}

func TestSkillContractValidatesMetadataAndInterface(t *testing.T) {
	tests := []struct {
		name      string
		directory string
		skill     string
		openAI    string
		want      string
	}{
		{
			name:      "name matches directory",
			directory: "fixture",
			skill:     "---\nname: other\ndescription: Describe a useful fixture and when an agent should use it.\nlicense: MIT\n---\n\n# Fixture\n",
			want:      "must match its directory",
		},
		{
			name:      "description required",
			directory: "fixture",
			skill:     "---\nname: fixture\ndescription: \nlicense: MIT\n---\n\n# Fixture\n",
			want:      "description must contain",
		},
		{
			name:      "description discovery budget",
			directory: "fixture",
			skill:     "---\nname: fixture\ndescription: " + strings.Repeat("x", 241) + "\nlicense: MIT\n---\n\n# Fixture\n",
			want:      "description must contain 1-240 characters",
		},
		{
			name:      "repository license",
			directory: "fixture",
			skill:     "---\nname: fixture\ndescription: Describe a useful fixture and when an agent should use it.\nlicense: Apache-2.0\n---\n\n# Fixture\n",
			want:      "license must be MIT",
		},
		{
			name:      "unknown frontmatter field",
			directory: "fixture",
			skill:     "---\nname: fixture\ndescription: Describe a useful fixture and when an agent should use it.\nlicense: MIT\ntyop: rejected\n---\n\n# Fixture\n",
			want:      `unsupported frontmatter field "tyop"`,
		},
		{
			name:      "compatibility limit",
			directory: "fixture",
			skill:     "---\nname: fixture\ndescription: Describe a useful fixture and when an agent should use it.\nlicense: MIT\ncompatibility: " + strings.Repeat("x", maxSkillCompatibility+1) + "\n---\n\n# Fixture\n",
			want:      "compatibility must contain",
		},
		{
			name:      "metadata mapping",
			directory: "fixture",
			skill:     "---\nname: fixture\ndescription: Describe a useful fixture and when an agent should use it.\nlicense: MIT\nmetadata: [not, a, mapping]\n---\n\n# Fixture\n",
			want:      "decode frontmatter",
		},
		{
			name:      "interface prompt names skill",
			directory: "fixture",
			skill:     "---\nname: fixture\ndescription: Describe a useful fixture and when an agent should use it.\nlicense: MIT\n---\n\n# Fixture\n\nCodex interface metadata lives in [agents/openai.yaml](agents/openai.yaml).\n",
			openAI:    "interface:\n  display_name: \"Fixture\"\n  short_description: \"Exercise the fixture contract safely\"\n  default_prompt: \"Review this fixture without invoking its skill.\"\n",
			want:      "default_prompt must mention $fixture",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			repo := t.TempDir()
			skillDirectory := filepath.Join(repo, "skills", test.directory)
			if err := os.MkdirAll(skillDirectory, 0o755); err != nil {
				t.Fatal(err)
			}
			skillFile := filepath.Join(skillDirectory, "SKILL.md")
			if err := os.WriteFile(skillFile, []byte(test.skill), 0o644); err != nil {
				t.Fatal(err)
			}
			if test.openAI != "" {
				agents := filepath.Join(skillDirectory, "agents")
				if err := os.MkdirAll(agents, 0o755); err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(filepath.Join(agents, "openai.yaml"), []byte(test.openAI), 0o644); err != nil {
					t.Fatal(err)
				}
			}

			assertSkillFinding(t, skillContractFindings(repo, skillFile, nil), test.want)
		})
	}
}

func TestSkillContractRequiresExactToolDocumentation(t *testing.T) {
	t.Run("substring is not a tool token", func(t *testing.T) {
		repo := t.TempDir()
		_, skillFile := writeSkillContractFixture(t, repo, "Use Google Cloud for the service.")
		assertSkillFinding(t, skillContractFindings(repo, skillFile, []string{"go"}), `required tool "go" is undocumented`)
	})

	t.Run("command token is documented", func(t *testing.T) {
		repo := t.TempDir()
		_, skillFile := writeSkillContractFixture(t, repo, "Run `go test ./...` before handoff.")
		if findings := skillContractFindings(repo, skillFile, []string{"go"}); len(findings) != 0 {
			t.Fatalf("exact required-tool token produced findings: %v", findings)
		}
	})

	t.Run("visible remote-link label is documented", func(t *testing.T) {
		repo := t.TempDir()
		_, skillFile := writeSkillContractFixture(t, repo, "Install [Go](https://go.dev/doc/install) before continuing.")
		if findings := skillContractFindings(repo, skillFile, []string{"go"}); len(findings) != 0 {
			t.Fatalf("visible required-tool link label produced findings: %v", findings)
		}
	})
}

func TestSkillContractRequiresExactInvocationBoundaries(t *testing.T) {
	for _, test := range []struct {
		name   string
		prompt string
		want   bool
	}{
		{name: "standalone", prompt: "Use $fixture for this request.", want: true},
		{name: "punctuation", prompt: "Use ($fixture), then report.", want: true},
		{name: "left word", prompt: "Use not$fixture without invoking it.", want: false},
		{name: "uppercase suffix", prompt: "Use $fixtureX instead.", want: false},
		{name: "underscore suffix", prompt: "Use $fixture_extra instead.", want: false},
	} {
		t.Run(test.name, func(t *testing.T) {
			if got := containsExactSkillInvocation(test.prompt, "fixture"); got != test.want {
				t.Errorf("containsExactSkillInvocation(%q) = %v, want %v", test.prompt, got, test.want)
			}
		})
	}
}

func TestSkillContractLineCount(t *testing.T) {
	for _, test := range []struct {
		name    string
		content []byte
		want    int
	}{
		{name: "empty", content: nil, want: 0},
		{name: "one terminated line", content: []byte("line\n"), want: 1},
		{name: "one unterminated line", content: []byte("line"), want: 1},
		{name: "five hundred terminated lines", content: bytes.Repeat([]byte("line\n"), 500), want: 500},
		{name: "five hundred one terminated lines", content: bytes.Repeat([]byte("line\n"), 501), want: 501},
	} {
		t.Run(test.name, func(t *testing.T) {
			if got := skillLineCount(test.content); got != test.want {
				t.Errorf("skillLineCount() = %d, want %d", got, test.want)
			}
		})
	}
}

func TestSkillContractRejectsNestedLayout(t *testing.T) {
	repo := t.TempDir()
	path := filepath.Join(repo, "skills", "group", "fixture", "SKILL.md")
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte("---\nname: fixture\ndescription: Nested fixture.\nlicense: MIT\n---\n\n# Fixture\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	_, findings := discoverSkillFiles(repo)
	assertSkillFinding(t, findings, "unsupported nested skill layout")
}

func TestSkillContractRejectsRepositoryEscape(t *testing.T) {
	parent := t.TempDir()
	repo := filepath.Join(parent, "repository")
	skillDirectory := filepath.Join(repo, "skills", "fixture")
	if err := os.MkdirAll(skillDirectory, 0o755); err != nil {
		t.Fatal(err)
	}
	outside := filepath.Join(parent, "outside.md")
	if err := os.WriteFile(outside, []byte("outside\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	skillFile := filepath.Join(skillDirectory, "SKILL.md")
	skill := "---\nname: fixture\ndescription: Exercise repository path containment. Use for contract testing only.\nlicense: MIT\n---\n\n# Fixture\n\nRead [outside](../../../outside.md).\n"
	if err := os.WriteFile(skillFile, []byte(skill), 0o644); err != nil {
		t.Fatal(err)
	}

	assertSkillFinding(t, skillContractFindings(repo, skillFile, nil), "escapes the repository")
}

func TestSkillContractRejectsSymlinkedPackageRoots(t *testing.T) {
	content := []byte("---\nname: fixture\ndescription: Exercise symlinked package-root rejection. Use for contract testing only.\nlicense: MIT\n---\n\n# Fixture\n")
	t.Run("SKILL.md", func(t *testing.T) {
		parent := t.TempDir()
		repo := filepath.Join(parent, "repo")
		directory := filepath.Join(repo, "skills", "fixture")
		if err := os.MkdirAll(directory, 0o755); err != nil {
			t.Fatal(err)
		}
		target := filepath.Join(parent, "outside-skill.md")
		if err := os.WriteFile(target, content, 0o644); err != nil {
			t.Fatal(err)
		}
		skillFile := filepath.Join(directory, "SKILL.md")
		if err := os.Symlink(target, skillFile); err != nil {
			t.Fatal(err)
		}

		assertSkillFinding(t, skillContractFindings(repo, skillFile, nil), "symbolic link")
	})

	t.Run("skill directory", func(t *testing.T) {
		parent := t.TempDir()
		repo := filepath.Join(parent, "repo")
		if err := os.MkdirAll(filepath.Join(repo, "skills"), 0o755); err != nil {
			t.Fatal(err)
		}
		target := filepath.Join(parent, "fixture")
		if err := os.MkdirAll(target, 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(target, "SKILL.md"), content, 0o644); err != nil {
			t.Fatal(err)
		}
		directory := filepath.Join(repo, "skills", "fixture")
		if err := os.Symlink(target, directory); err != nil {
			t.Fatal(err)
		}

		assertSkillFinding(t, skillContractFindings(repo, filepath.Join(directory, "SKILL.md"), nil), "symbolic link")
	})
}

func TestSkillContractRejectsUnsafeInstructionLinks(t *testing.T) {
	t.Run("nested repository escape", func(t *testing.T) {
		parent := t.TempDir()
		repo := filepath.Join(parent, "repo")
		directory, skillFile := writeSkillContractFixture(t, repo, "Read [the detail](references/detail.md).")
		if err := os.MkdirAll(filepath.Join(directory, "references"), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(parent, "outside.md"), []byte("outside\n"), 0o644); err != nil {
			t.Fatal(err)
		}
		detail := "# Detail\n\nRead [outside](../../../../outside.md).\n"
		if err := os.WriteFile(filepath.Join(directory, "references", "detail.md"), []byte(detail), 0o644); err != nil {
			t.Fatal(err)
		}

		assertSkillFinding(t, skillContractFindings(repo, skillFile, nil), "escapes the repository")
	})

	for _, test := range []struct {
		name string
		body string
		want string
	}{
		{name: "inline file URL", body: "Read [outside](file:///etc/passwd).", want: "unsupported local link"},
		{name: "empty-label file URL", body: "Read [](file:///etc/passwd).", want: "unsupported local link"},
		{name: "escaped-label file URL", body: `Read [host\]](file:///etc/passwd).`, want: "unsupported local link"},
		{name: "home-relative", body: "Read [outside](~/.ssh/id_ed25519).", want: "home-relative link"},
		{name: "URI autolink", body: "Read <file:///etc/passwd>.", want: "unsupported local link"},
		{name: "reference-style link", body: "Read [outside][host].\n\n[host]: file:///etc/passwd", want: "unsupported local link"},
		{name: "HTML double-quoted link", body: `<a href="file:///etc/passwd">outside</a>`, want: "unsupported local link"},
		{name: "HTML single-quoted link", body: `<img src='~/.ssh/id_ed25519'>`, want: "home-relative link"},
		{name: "HTML entity-encoded link", body: `<a href="file&#58;///etc/passwd">outside</a>`, want: "unsupported local link"},
		{name: "HTML object data link", body: `<object data="file:///etc/passwd"></object>`, want: "unsupported local link"},
		{name: "HTML source-set link", body: `<img srcset="file:///etc/passwd 1x, https://example.com/image.png 2x">`, want: "unsupported local link"},
		{name: "HTML template link", body: "<template>\n<a href=\"file:///etc/passwd\">outside</a>\n</template>", want: "unsupported local link"},
	} {
		t.Run(test.name, func(t *testing.T) {
			repo := t.TempDir()
			_, skillFile := writeSkillContractFixture(t, repo, test.body)
			assertSkillFinding(t, skillContractFindings(repo, skillFile, nil), test.want)
		})
	}
}

func TestSkillContractIgnoresNonLinkHTMLDataAttributes(t *testing.T) {
	for _, test := range []struct {
		name string
		body string
	}{
		{name: "data-href", body: `<span data-href="file:///etc/passwd">metadata only</span>`},
		{name: "data on a non-object element", body: `<div data="metadata">example</div>`},
		{name: "commented-out link", body: `<!-- <a href="file:///etc/passwd">disabled example</a> -->`},
		{name: "script source text", body: "<script>\nconst href = \"file:///etc/passwd\";\n</script>"},
	} {
		t.Run(test.name, func(t *testing.T) {
			repo := t.TempDir()
			_, skillFile := writeSkillContractFixture(t, repo, test.body)
			if findings := skillContractFindings(repo, skillFile, nil); len(findings) != 0 {
				t.Fatalf("non-link HTML metadata produced findings: %v", findings)
			}
		})
	}
}

func TestSkillContractAcceptsCommaInsideSourceSetURL(t *testing.T) {
	repo := t.TempDir()
	_, skillFile := writeSkillContractFixture(t, repo, `<img srcset="https://example.com/a,b.png 1x">`)
	if findings := skillContractFindings(repo, skillFile, nil); len(findings) != 0 {
		t.Fatalf("valid srcset URL containing a comma produced findings: %v", findings)
	}
}

func TestSkillContractParsesSourceSetCandidates(t *testing.T) {
	for _, test := range []struct {
		name  string
		value string
		want  []string
	}{
		{name: "descriptors", value: "one.png 1x, two.png 2x", want: []string{"one.png", "two.png"}},
		{name: "URL comma", value: "https://example.com/a,b.png 1x, file:///etc/passwd 2x", want: []string{"https://example.com/a,b.png", "file:///etc/passwd"}},
		{name: "data URL", value: "data:image/png;base64,AAAA 1x", want: []string{"data:image/png;base64,AAAA"}},
		{name: "trailing comma", value: "one.png, two.png,", want: []string{"one.png", "two.png"}},
	} {
		t.Run(test.name, func(t *testing.T) {
			if got := skillHTMLSourceSetTargets(test.value); !slices.Equal(got, test.want) {
				t.Fatalf("srcset targets = %q, want %q", got, test.want)
			}
		})
	}
}

func TestSkillContractAcceptsCommonMarkDestinationWithSpaces(t *testing.T) {
	for _, test := range []struct {
		name        string
		destination string
	}{
		{name: "angle destination", destination: "<references/foo bar.md>"},
		{name: "percent-encoded destination", destination: "references/foo%20bar.md"},
	} {
		t.Run(test.name, func(t *testing.T) {
			repo := t.TempDir()
			directory, skillFile := writeSkillContractFixture(t, repo, "Read [the note]("+test.destination+").")
			if err := os.MkdirAll(filepath.Join(directory, "references"), 0o755); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(filepath.Join(directory, "references", "foo bar.md"), []byte("# Note\n"), 0o644); err != nil {
				t.Fatal(err)
			}
			if findings := skillContractFindings(repo, skillFile, nil); len(findings) != 0 {
				t.Fatalf("valid CommonMark destination with spaces produced findings: %v", findings)
			}
		})
	}
}

func TestSkillContractRejectsMissingNestedMarkdownLink(t *testing.T) {
	repo := t.TempDir()
	skillDirectory, skillFile := writeSkillContractFixture(t, repo, "Read [the guide](references/guide.md).")
	referencesDirectory := filepath.Join(skillDirectory, "references")
	if err := os.MkdirAll(referencesDirectory, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(referencesDirectory, "guide.md"), []byte("# Guide\n\nContinue with [the missing detail](missing.md).\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	assertSkillFinding(t, skillContractFindings(repo, skillFile, nil), `missing referenced resource "missing.md"`)
}

func TestSkillContractResolvesNestedMarkdownLinksFromSkillRoot(t *testing.T) {
	repo := t.TempDir()
	skillDirectory, skillFile := writeSkillContractFixture(t, repo, "Read [the guide](references/guide.md) and [the detail](references/detail.md).")
	referencesDirectory := filepath.Join(skillDirectory, "references")
	if err := os.MkdirAll(referencesDirectory, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(referencesDirectory, "guide.md"), []byte("# Guide\n\nContinue with [the detail](references/detail.md).\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(referencesDirectory, "detail.md"), []byte("# Detail\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	if findings := skillContractFindings(repo, skillFile, nil); len(findings) != 0 {
		t.Fatalf("valid skill-root-relative resource link produced findings: %v", findings)
	}
}

// A checkout is often reached through a symlinked path: macOS puts temporary
// directories behind /var -> /private/var, and a home directory can itself be a
// link. Containment must then compare two evaluated paths, or every in-package
// resource looks like an escape — which is exactly how this suite failed on
// macOS while passing on Linux.
func TestSkillContractAcceptsSymlinkedRepositoryRoot(t *testing.T) {
	repo := t.TempDir()
	skillDirectory, _ := writeSkillContractFixture(t, repo, "Read [the guide](references/guide.md).")
	referencesDirectory := filepath.Join(skillDirectory, "references")
	if err := os.MkdirAll(referencesDirectory, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(referencesDirectory, "guide.md"), []byte("# Guide\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	linkedRepo := filepath.Join(t.TempDir(), "linked-repo")
	if err := os.Symlink(repo, linkedRepo); err != nil {
		t.Skipf("symlinks unavailable: %v", err)
	}
	linkedSkill := filepath.Join(linkedRepo, "skills", "fixture", "SKILL.md")

	if findings := skillContractFindings(linkedRepo, linkedSkill, nil); len(findings) != 0 {
		t.Fatalf("skill reached through a symlinked repository root produced findings: %v", findings)
	}
}

// Companion to the test above: evaluating the root must not turn the containment
// check into a rubber stamp. A resource that truly leaves the repository through
// a linked directory still has to be reported.
func TestSkillContractRejectsResourceEscapingThroughLinkedDirectory(t *testing.T) {
	repo := t.TempDir()
	skillDirectory, skillFile := writeSkillContractFixture(t, repo, "Read [the guide](references/guide.md).")
	outside := t.TempDir()
	if err := os.WriteFile(filepath.Join(outside, "guide.md"), []byte("# Guide\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(outside, filepath.Join(skillDirectory, "references")); err != nil {
		t.Skipf("symlinks unavailable: %v", err)
	}

	assertSkillFinding(t, skillContractFindings(repo, skillFile, nil), "resolves outside the repository")
}

func TestSkillContractRejectsResourceDirectoryRelativeMarkdownLink(t *testing.T) {
	repo := t.TempDir()
	skillDirectory, skillFile := writeSkillContractFixture(t, repo, "Read [the guide](references/guide.md) and [the detail](references/detail.md).")
	referencesDirectory := filepath.Join(skillDirectory, "references")
	if err := os.MkdirAll(referencesDirectory, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(referencesDirectory, "guide.md"), []byte("# Guide\n\nContinue with [the detail](detail.md).\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(referencesDirectory, "detail.md"), []byte("# Detail\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	assertSkillFinding(t, skillContractFindings(repo, skillFile, nil), `missing referenced resource "detail.md"`)
}

func TestSkillContractRejectsLinksOutsideTheSkillCatalog(t *testing.T) {
	repo := t.TempDir()
	if err := os.WriteFile(filepath.Join(repo, "README.md"), []byte("repository documentation\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	_, skillFile := writeSkillContractFixture(t, repo, "Read [repository documentation](../../README.md).")
	assertSkillFinding(t, skillContractFindings(repo, skillFile, nil), "outside the first-party skill catalog")
}

func TestSkillContractRejectsResidualFormatControls(t *testing.T) {
	for _, test := range []struct {
		name    string
		content string
		want    string
	}{
		{name: "bare carriage return", content: "safe prefix\rspoofed suffix\n", want: "unsafe control character U+000D"},
		{name: "Arabic letter mark", content: "safe prefix\u061Cspoofed suffix\n", want: "bidirectional or invisible Unicode"},
		{name: "word joiner", content: "safe prefix\u2060spoofed suffix\n", want: "bidirectional or invisible Unicode"},
	} {
		t.Run(test.name, func(t *testing.T) {
			assertSkillFinding(t, validateSkillText("fixture.md", []byte(test.content)), test.want)
		})
	}

	if findings := validateSkillText("fixture.md", []byte("windows\r\nline endings\r\n")); len(findings) != 0 {
		t.Fatalf("CRLF text produced findings: %v", findings)
	}
}

func TestSkillContractRequiresExactResourceReferences(t *testing.T) {
	t.Run("substring is not a reference", func(t *testing.T) {
		repo := t.TempDir()
		directory, skillFile := writeSkillContractFixture(t, repo, "Mention scripts/payload.sh.bak only.")
		if err := os.MkdirAll(filepath.Join(directory, "scripts"), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(directory, "scripts", "payload.sh"), []byte("#!/bin/sh\nexit 0\n"), 0o755); err != nil {
			t.Fatal(err)
		}

		assertSkillFinding(t, skillContractFindings(repo, skillFile, nil), "not directly disclosed")
	})

	t.Run("wrong parent-relative path is not a reference", func(t *testing.T) {
		repo := t.TempDir()
		directory, skillFile := writeSkillContractFixture(t, repo, "Read ../scripts/payload.sh only.")
		if err := os.MkdirAll(filepath.Join(directory, "scripts"), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(directory, "scripts", "payload.sh"), []byte("#!/bin/sh\nexit 0\n"), 0o755); err != nil {
			t.Fatal(err)
		}

		assertSkillFinding(t, skillContractFindings(repo, skillFile, nil), "not directly disclosed")
	})

	t.Run("scripts must be a directory", func(t *testing.T) {
		repo := t.TempDir()
		directory, skillFile := writeSkillContractFixture(t, repo, "Run scripts.")
		if err := os.WriteFile(filepath.Join(directory, "scripts"), []byte("#!/bin/sh\nexit 0\n"), 0o755); err != nil {
			t.Fatal(err)
		}

		assertSkillFinding(t, skillContractFindings(repo, skillFile, nil), "unsupported progressive-disclosure path")
	})
}

func TestSkillContractRequiresDirectResourceDisclosure(t *testing.T) {
	repo := t.TempDir()
	directory, skillFile := writeSkillContractFixture(t, repo, "Read [the reference index](references/index.md).")
	references := filepath.Join(directory, "references")
	if err := os.MkdirAll(references, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(references, "index.md"), []byte("# Index\n\nRead [the detail](detail.md).\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(references, "detail.md"), []byte("# Detail\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	assertSkillFinding(t, skillContractFindings(repo, skillFile, nil), `resource "references/detail.md" is not directly disclosed`)
}

func TestSkillContractRemoteURLsDoNotReferenceLocalResources(t *testing.T) {
	for _, test := range []struct {
		name string
		body string
	}{
		{name: "inline link query", body: "Read [remote](https://example.com/?path=scripts/payload.sh)."},
		{name: "URI autolink path", body: "Read <https://example.com/scripts/payload.sh>."},
		{name: "reference link fragment", body: "Read [remote][docs].\n\n[docs]: https://example.com/#scripts/payload.sh"},
		{name: "HTML attribute query", body: `<a href="https://example.com/?path=scripts/payload.sh">remote</a>`},
		{name: "HTML entity scheme", body: `<a href="https&#58;//example.com/scripts/payload.sh">remote</a>`},
		{name: "bare URL query", body: "Read https://example.com/?path=scripts/payload.sh for context."},
		{name: "bare HTML entity URL query", body: "Read https&#58;//example.com/?path=scripts/payload.sh for context."},
		{name: "Markdown-escaped URL query", body: `Read https\://example.com/?path=scripts/payload.sh for context.`},
	} {
		t.Run(test.name, func(t *testing.T) {
			repo := t.TempDir()
			directory, skillFile := writeSkillContractFixture(t, repo, test.body)
			if err := os.MkdirAll(filepath.Join(directory, "scripts"), 0o755); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(filepath.Join(directory, "scripts", "payload.sh"), []byte("#!/bin/sh\nexit 0\n"), 0o755); err != nil {
				t.Fatal(err)
			}

			assertSkillFinding(t, skillContractFindings(repo, skillFile, nil), "not directly disclosed")
		})
	}
}

func TestSkillContractDoesNotDoubleDecodeHTMLAttributes(t *testing.T) {
	targets, err := skillLinkTargets("fixture.md", `<a href="file&amp;#58;///etc/passwd">literal encoding</a>`)
	if err != nil {
		t.Fatal(err)
	}
	if !slices.Equal(targets, []string{"file&#58;///etc/passwd"}) {
		t.Fatalf("HTML attribute targets = %q, want one decoding pass", targets)
	}

	repo := t.TempDir()
	directory, skillFile := writeSkillContractFixture(t, repo, `<a href="scripts&amp;#47;payload.sh">literal path encoding</a>`)
	if err := os.MkdirAll(filepath.Join(directory, "scripts"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(directory, "scripts", "payload.sh"), []byte("#!/bin/sh\nexit 0\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	assertSkillFinding(t, skillContractFindings(repo, skillFile, nil), `resource "scripts/payload.sh" is not directly disclosed`)
}

func TestSkillContractVisibleHTMLBlockReferencesLocalResource(t *testing.T) {
	for _, test := range []struct {
		name string
		body string
	}{
		{name: "plain text", body: "<div>\nRun scripts/payload.sh before handoff.\n</div>"},
		{name: "literal less-than", body: "<div>\nWhen x < y, run scripts/payload.sh before handoff.\n</div>"},
	} {
		t.Run(test.name, func(t *testing.T) {
			repo := t.TempDir()
			directory, skillFile := writeSkillContractFixture(t, repo, test.body)
			if err := os.MkdirAll(filepath.Join(directory, "scripts"), 0o755); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(filepath.Join(directory, "scripts", "payload.sh"), []byte("#!/bin/sh\nexit 0\n"), 0o755); err != nil {
				t.Fatal(err)
			}
			if findings := skillContractFindings(repo, skillFile, nil); len(findings) != 0 {
				t.Fatalf("visible HTML-block resource reference produced findings: %v", findings)
			}
		})
	}
}

func TestSkillContractEncodedCodeDoesNotReferenceLocalResource(t *testing.T) {
	for _, test := range []struct {
		name string
		body string
	}{
		{name: "HTML entity", body: "```text\nscripts&#47;payload.sh\n```"},
		{name: "backslash escape", body: "```text\nscripts\\/payload.sh\n```"},
		{name: "inline code entity", body: "The literal `scripts&#47;payload.sh` is not the package path."},
	} {
		t.Run(test.name, func(t *testing.T) {
			repo := t.TempDir()
			directory, skillFile := writeSkillContractFixture(t, repo, test.body)
			if err := os.MkdirAll(filepath.Join(directory, "scripts"), 0o755); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(filepath.Join(directory, "scripts", "payload.sh"), []byte("#!/bin/sh\nexit 0\n"), 0o755); err != nil {
				t.Fatal(err)
			}
			assertSkillFinding(t, skillContractFindings(repo, skillFile, nil), "not directly disclosed")
		})
	}
}

func TestSkillContractNonMarkdownEntityDoesNotReferenceNestedResource(t *testing.T) {
	repo := t.TempDir()
	directory, skillFile := writeSkillContractFixture(t, repo, "Run scripts/launcher.py.")
	if err := os.MkdirAll(filepath.Join(directory, "scripts"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(directory, "references"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(directory, "scripts", "launcher.py"), []byte("# Literal references&#47;payload.md is not a filesystem path.\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(directory, "references", "payload.md"), []byte("# Payload\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	assertSkillFinding(t, skillContractFindings(repo, skillFile, nil), `resource "references/payload.md" is not directly disclosed`)
}

func TestSkillContractValidatesOutputTemplateLinksWithoutTreatingThemAsPackageResources(t *testing.T) {
	for _, test := range []struct {
		name     string
		template string
		want     string
	}{
		{
			name:     "relative output destination",
			template: "Start with [Getting Started](getting-started) after this file is copied.\n",
		},
		{
			name:     "unsafe local scheme",
			template: "Open [host data](file:///etc/passwd).\n",
			want:     "unsupported local link scheme",
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			repo := t.TempDir()
			skillDirectory, skillFile := writeSkillContractFixture(t, repo, "Use [the output template](templates/page.md).")
			templateDirectory := filepath.Join(skillDirectory, "templates")
			if err := os.MkdirAll(templateDirectory, 0o755); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(filepath.Join(templateDirectory, "page.md"), []byte(test.template), 0o644); err != nil {
				t.Fatal(err)
			}

			findings := skillContractFindings(repo, skillFile, nil)
			if test.want == "" {
				if len(findings) != 0 {
					t.Fatalf("valid output template findings = %q", findings)
				}
				return
			}
			assertSkillFinding(t, findings, test.want)
		})
	}
}

func TestSkillContractValidatesOpenAIPaths(t *testing.T) {
	t.Run("missing icon", func(t *testing.T) {
		repo := t.TempDir()
		directory, skillFile := writeSkillContractFixture(t, repo, "Codex interface metadata is package-owned.")
		writeSkillOpenAIConfig(t, directory, "  icon_small: \"./assets/missing.svg\"\n  default_prompt: \"Use $fixture to inspect this package.\"\n")

		assertSkillFinding(t, skillContractFindings(repo, skillFile, nil), "assets/missing.svg")
	})

	t.Run("valid icon is an exact metadata reference", func(t *testing.T) {
		repo := t.TempDir()
		directory, skillFile := writeSkillContractFixture(t, repo, "Codex interface metadata is package-owned.")
		if err := os.MkdirAll(filepath.Join(directory, "assets"), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(directory, "assets", "icon.svg"), []byte("<svg xmlns=\"http://www.w3.org/2000/svg\"/>\n"), 0o644); err != nil {
			t.Fatal(err)
		}
		writeSkillOpenAIConfig(t, directory, "  icon_small: \"./assets/icon.svg\"\n  default_prompt: \"Use $fixture to inspect this package.\"\n")

		if findings := skillContractFindings(repo, skillFile, nil); len(findings) != 0 {
			t.Fatalf("valid Codex icon package produced findings: %v", findings)
		}
	})

	t.Run("prompt invocation token", func(t *testing.T) {
		repo := t.TempDir()
		directory, skillFile := writeSkillContractFixture(t, repo, "Codex interface metadata is package-owned.")
		writeSkillOpenAIConfig(t, directory, "  default_prompt: \"Use $fixture-extra instead.\"\n")

		assertSkillFinding(t, skillContractFindings(repo, skillFile, nil), "default_prompt")
	})
}

func TestSkillContractValidatesOpenAISchema(t *testing.T) {
	validConfig := "interface:\n" +
		"  display_name: \"Fixture\"\n" +
		"  short_description: \"Exercise the fixture package contract\"\n" +
		"  icon_small: \"./assets/icon.svg\"\n" +
		"  icon_large: \"./assets/icon.svg\"\n" +
		"  brand_color: \"#3B82F6\"\n" +
		"  default_prompt: \"Use $fixture to inspect this package.\"\n" +
		"dependencies:\n" +
		"  tools:\n" +
		"    - type: \"mcp\"\n" +
		"      value: \"github\"\n" +
		"      description: \"GitHub MCP server\"\n" +
		"      transport: \"streamable_http\"\n" +
		"      url: \"https://api.githubcopilot.com/mcp/\"\n" +
		"policy:\n" +
		"  allow_implicit_invocation: false\n"
	remoteFields := "      transport: \"streamable_http\"\n" +
		"      url: \"https://api.githubcopilot.com/mcp/\"\n"

	t.Run("full supported schema", func(t *testing.T) {
		repo := t.TempDir()
		directory, skillFile := writeSkillContractFixture(t, repo, "Codex interface metadata is package-owned.")
		if err := os.MkdirAll(filepath.Join(directory, "assets"), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(directory, "assets", "icon.svg"), []byte("<svg xmlns=\"http://www.w3.org/2000/svg\"/>\n"), 0o644); err != nil {
			t.Fatal(err)
		}
		writeSkillOpenAIConfigDocument(t, directory, validConfig)

		if findings := skillContractFindings(repo, skillFile, nil); len(findings) != 0 {
			t.Fatalf("valid full Codex interface schema produced findings: %v", findings)
		}
	})

	for _, test := range []struct {
		name   string
		config string
	}{
		{
			name:   "named MCP dependency",
			config: strings.Replace(validConfig, remoteFields, "", 1),
		},
		{
			name:   "CLI dependency",
			config: strings.Replace(strings.Replace(validConfig, remoteFields, "", 1), `type: "mcp"`, `type: "cli"`, 1),
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			repo := t.TempDir()
			directory, skillFile := writeSkillContractFixture(t, repo, "Codex interface metadata is package-owned.")
			if err := os.MkdirAll(filepath.Join(directory, "assets"), 0o755); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(filepath.Join(directory, "assets", "icon.svg"), []byte("<svg xmlns=\"http://www.w3.org/2000/svg\"/>\n"), 0o644); err != nil {
				t.Fatal(err)
			}
			writeSkillOpenAIConfigDocument(t, directory, test.config)

			if findings := skillContractFindings(repo, skillFile, nil); len(findings) != 0 {
				t.Fatalf("valid Codex dependency schema produced findings: %v", findings)
			}
		})
	}

	for _, test := range []struct {
		name   string
		config string
		want   string
	}{
		{
			name:   "unknown top-level field",
			config: validConfig + "unexpected: true\n",
			want:   `field unexpected not found`,
		},
		{
			name:   "unknown interface field",
			config: strings.Replace(validConfig, "  display_name:", "  display_nam: \"typo\"\n  display_name:", 1),
			want:   `field display_nam not found`,
		},
		{
			name:   "unknown dependencies field",
			config: strings.Replace(validConfig, "  tools:", "  toolz: []\n  tools:", 1),
			want:   `field toolz not found`,
		},
		{
			name:   "unknown tool field",
			config: strings.Replace(validConfig, "      value:", "      vale: \"typo\"\n      value:", 1),
			want:   `field vale not found`,
		},
		{
			name:   "unknown policy field",
			config: strings.Replace(validConfig, "  allow_implicit_invocation:", "  allow_implicit_invocaton: true\n  allow_implicit_invocation:", 1),
			want:   `field allow_implicit_invocaton not found`,
		},
		{
			name:   "invalid brand color",
			config: strings.Replace(validConfig, "#3B82F6", "blue", 1),
			want:   "interface.brand_color",
		},
		{
			name:   "unsupported tool type",
			config: strings.Replace(validConfig, `type: "mcp"`, `type: "shell"`, 1),
			want:   "dependencies.tools[0].type",
		},
		{
			name:   "empty tool type",
			config: strings.Replace(validConfig, `type: "mcp"`, `type: " "`, 1),
			want:   "dependencies.tools[0].type is required",
		},
		{
			name:   "empty tool value",
			config: strings.Replace(validConfig, `value: "github"`, `value: " "`, 1),
			want:   "dependencies.tools[0].value",
		},
		{
			name:   "empty tool description",
			config: strings.Replace(validConfig, `description: "GitHub MCP server"`, `description: ""`, 1),
			want:   "dependencies.tools[0].description",
		},
		{
			name:   "empty tool transport",
			config: strings.Replace(validConfig, `transport: "streamable_http"`, `transport: ""`, 1),
			want:   "dependencies.tools[0].transport",
		},
		{
			name:   "empty tool URL",
			config: strings.Replace(validConfig, `url: "https://api.githubcopilot.com/mcp/"`, `url: ""`, 1),
			want:   "dependencies.tools[0].url",
		},
		{
			name:   "invalid tool URL",
			config: strings.Replace(validConfig, "https://api.githubcopilot.com/mcp/", "relative/mcp", 1),
			want:   "absolute HTTP(S) URL",
		},
		{
			name:   "unsupported remote MCP transport",
			config: strings.Replace(validConfig, `transport: "streamable_http"`, `transport: "stdio"`, 1),
			want:   `transport must be "streamable_http"`,
		},
		{
			name:   "CLI rejects remote fields",
			config: strings.Replace(validConfig, `type: "mcp"`, `type: "cli"`, 1),
			want:   "CLI dependency must not declare transport or url",
		},
		{
			name:   "policy must be boolean",
			config: strings.Replace(validConfig, "allow_implicit_invocation: false", "allow_implicit_invocation: \"sometimes\"", 1),
			want:   "decode interface metadata",
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			repo := t.TempDir()
			directory, skillFile := writeSkillContractFixture(t, repo, "Codex interface metadata is package-owned.")
			if err := os.MkdirAll(filepath.Join(directory, "assets"), 0o755); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(filepath.Join(directory, "assets", "icon.svg"), []byte("<svg xmlns=\"http://www.w3.org/2000/svg\"/>\n"), 0o644); err != nil {
				t.Fatal(err)
			}
			writeSkillOpenAIConfigDocument(t, directory, test.config)

			assertSkillFinding(t, skillContractFindings(repo, skillFile, nil), test.want)
		})
	}
}

func TestSkillContractRejectsUnsafeOpenAIConfigPath(t *testing.T) {
	t.Run("symbolic link", func(t *testing.T) {
		repo := t.TempDir()
		directory, skillFile := writeSkillContractFixture(t, repo, "Codex interface metadata is package-owned.")
		agents := filepath.Join(directory, "agents")
		if err := os.MkdirAll(agents, 0o755); err != nil {
			t.Fatal(err)
		}
		target := filepath.Join(repo, "outside.yaml")
		if err := os.WriteFile(target, []byte("interface: {}\n"), 0o644); err != nil {
			t.Fatal(err)
		}
		if err := os.Symlink(target, filepath.Join(agents, "openai.yaml")); err != nil {
			t.Fatal(err)
		}

		assertSkillFinding(t, skillContractFindings(repo, skillFile, nil), "symbolic link")
	})

	t.Run("named pipe", func(t *testing.T) {
		repo := t.TempDir()
		directory, skillFile := writeSkillContractFixture(t, repo, "Codex interface metadata is package-owned.")
		agents := filepath.Join(directory, "agents")
		if err := os.MkdirAll(agents, 0o755); err != nil {
			t.Fatal(err)
		}
		path := filepath.Join(agents, "openai.yaml")
		if err := syscall.Mkfifo(path, 0o600); err != nil {
			t.Fatal(err)
		}

		completed := make(chan []string, 1)
		go func() {
			completed <- skillContractFindings(repo, skillFile, nil)
		}()
		select {
		case findings := <-completed:
			assertSkillFinding(t, findings, "non-regular")
		case <-time.After(2 * time.Second):
			// A writer releases a regressed blocking read so the test leaves no goroutine behind.
			writerDone := make(chan error, 1)
			go func() {
				file, err := os.OpenFile(path, os.O_WRONLY, 0)
				if err == nil {
					err = file.Close()
				}
				writerDone <- err
			}()
			<-completed
			if err := <-writerDone; err != nil {
				t.Errorf("release named-pipe reader: %v", err)
			}
			t.Fatal("skill contract blocked while reading a named-pipe agents/openai.yaml")
		}
	})
}

func writeSkillContractFixture(t *testing.T, repo, body string) (string, string) {
	t.Helper()
	directory := filepath.Join(repo, "skills", "fixture")
	if err := os.MkdirAll(directory, 0o755); err != nil {
		t.Fatal(err)
	}
	skillFile := filepath.Join(directory, "SKILL.md")
	content := "---\nname: fixture\ndescription: Exercise an isolated package contract. Use for contract testing only.\nlicense: MIT\n---\n\n# Fixture\n\n" + body + "\n"
	if err := os.WriteFile(skillFile, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	return directory, skillFile
}

func writeSkillOpenAIConfig(t *testing.T, directory, extra string) {
	t.Helper()
	config := "interface:\n  display_name: \"Fixture\"\n  short_description: \"Exercise the fixture package contract\"\n" + extra
	writeSkillOpenAIConfigDocument(t, directory, config)
}

func writeSkillOpenAIConfigDocument(t *testing.T, directory, config string) {
	t.Helper()
	agents := filepath.Join(directory, "agents")
	if err := os.MkdirAll(agents, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(agents, "openai.yaml"), []byte(config), 0o644); err != nil {
		t.Fatal(err)
	}
}

func assertSkillFinding(t *testing.T, findings []string, want string) {
	t.Helper()
	for _, finding := range findings {
		if strings.Contains(finding, want) {
			return
		}
	}
	t.Fatalf("findings %q do not contain %q", findings, want)
}

func skillRepositoryRoot(t *testing.T) string {
	t.Helper()
	workingDirectory, err := os.Getwd()
	if err != nil {
		t.Fatalf("resolve skill contract working directory: %v", err)
	}
	for candidate := filepath.Clean(workingDirectory); ; candidate = filepath.Dir(candidate) {
		if skillRepositoryMarkersExist(candidate) {
			return candidate
		}
		parent := filepath.Dir(candidate)
		if parent == candidate {
			break
		}
	}
	t.Fatalf("resolve repository root from %s: missing skills/contracts.json and dot/go.mod markers", workingDirectory)
	return ""
}

func skillRepositoryMarkersExist(root string) bool {
	for _, relative := range []string{"skills/contracts.json", "dot/go.mod"} {
		info, err := os.Stat(filepath.Join(root, filepath.FromSlash(relative)))
		if err != nil || !info.Mode().IsRegular() {
			return false
		}
	}
	return true
}

func readSkillManifest(t *testing.T, repo string) skillManifest {
	t.Helper()
	path := filepath.Join(repo, "skills", "contracts.json")
	data, err := readSkillParsedFile(path)
	if err != nil {
		t.Fatal(err)
	}
	manifest, err := decodeSkillManifest(data)
	if err != nil {
		t.Fatalf("skills/contracts.json: decode: %v", err)
	}
	return manifest
}

func decodeSkillManifest(data []byte) (skillManifest, error) {
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	var manifest skillManifest
	if err := decoder.Decode(&manifest); err != nil {
		return skillManifest{}, err
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		return skillManifest{}, fmt.Errorf("expected one JSON document: %w", err)
	}
	return manifest, nil
}

func readSkillParsedFile(path string) ([]byte, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	data, readErr := readSkillParsedContent(file)
	closeErr := file.Close()
	if readErr != nil {
		return nil, readErr
	}
	if closeErr != nil {
		return nil, closeErr
	}
	return data, nil
}

func readSkillParsedContent(reader io.Reader) ([]byte, error) {
	data, err := io.ReadAll(io.LimitReader(reader, maxSkillParsedFileBytes+1))
	if err != nil {
		return nil, err
	}
	if len(data) > maxSkillParsedFileBytes {
		return nil, fmt.Errorf("exceeds the %d-byte parsed-file limit", maxSkillParsedFileBytes)
	}
	return data, nil
}

func checkSkillContract(t *testing.T, repo, skillFile string, tools []string) {
	t.Helper()
	for _, finding := range skillContractFindings(repo, skillFile, tools) {
		t.Error(finding)
	}
}

func skillContractFindings(repo, skillFile string, tools []string) []string {
	if findings := validateSkillPackageRoot(repo, skillFile); len(findings) != 0 {
		return findings
	}
	data, err := readSkillParsedFile(skillFile)
	if err != nil {
		return []string{fmt.Sprintf("%s: read skill: %v", skillFile, err)}
	}
	relativeSkill := skillRelativePath(repo, skillFile)
	findings := validateSkillFrontmatter(relativeSkill, filepath.Base(filepath.Dir(skillFile)), data)
	if lines := skillLineCount(data); lines > maxSkillLines {
		findings = append(findings, fmt.Sprintf("%s: %d lines exceeds the %d-line progressive-disclosure limit; move detail into a referenced resource", relativeSkill, lines, maxSkillLines))
	}
	findings = append(findings, validateSkillText(relativeSkill, data)...)

	content := string(data)
	documentedContent, referenceErr := skillReferenceContent(skillFile, content, true)
	if referenceErr != nil {
		findings = append(findings, fmt.Sprintf("%s: parse documented tool references: %v", relativeSkill, referenceErr))
	}
	for _, tool := range tools {
		if !toolNamePattern.MatchString(tool) {
			findings = append(findings, fmt.Sprintf("skills/contracts.json: skill %q has invalid required tool %q; use its executable name", filepath.Base(filepath.Dir(skillFile)), tool))
		}
		if !containsExactToolName(documentedContent, tool) {
			findings = append(findings, fmt.Sprintf("%s: required tool %q is undocumented; name the prerequisite or remove it from skills/contracts.json", relativeSkill, tool))
		}
	}

	findings = append(findings, validateSkillLinks(repo, filepath.Dir(skillFile), skillFile, content, true)...)
	findings = append(findings, validateSkillResources(repo, skillFile, content)...)
	findings = append(findings, validateSkillOpenAIConfig(repo, skillFile)...)
	return findings
}

func containsExactToolName(content, tool string) bool {
	content = strings.ToLower(content)
	tool = strings.ToLower(tool)
	for offset := 0; offset < len(content); {
		index := strings.Index(content[offset:], tool)
		if index < 0 {
			return false
		}
		start := offset + index
		end := start + len(tool)
		leftBoundary := start == 0 || !isToolTokenCharacter(content[start-1])
		rightBoundary := end == len(content) || !isToolTokenCharacter(content[end])
		if leftBoundary && rightBoundary {
			return true
		}
		offset = start + 1
	}
	return false
}

func isToolTokenCharacter(value byte) bool {
	return value >= 'a' && value <= 'z' || value >= 'A' && value <= 'Z' || value >= '0' && value <= '9' || value == '_' || value == '+' || value == '-' || value == '.'
}

func skillLineCount(data []byte) int {
	if len(data) == 0 {
		return 0
	}
	lines := bytes.Count(data, []byte("\n"))
	if data[len(data)-1] != '\n' {
		lines++
	}
	return lines
}

func validateSkillPackageRoot(repo, skillFile string) []string {
	directory := filepath.Dir(skillFile)
	for _, candidate := range []struct {
		kind string
		path string
	}{
		{kind: "skill directory", path: directory},
		{kind: "SKILL.md", path: skillFile},
	} {
		info, err := os.Lstat(candidate.path)
		if err != nil {
			return []string{fmt.Sprintf("%s: inspect %s: %v", skillRelativePath(repo, candidate.path), candidate.kind, err)}
		}
		if info.Mode()&os.ModeSymlink != 0 {
			return []string{fmt.Sprintf("%s: symbolic link %s is not allowed for a skill package root", skillRelativePath(repo, candidate.path), candidate.kind)}
		}
		if candidate.kind == "skill directory" && !info.IsDir() {
			return []string{skillRelativePath(repo, candidate.path) + ": skill directory is not a directory"}
		}
		if candidate.kind == "SKILL.md" && !info.Mode().IsRegular() {
			return []string{skillRelativePath(repo, candidate.path) + ": SKILL.md is not a regular file"}
		}
	}

	evaluatedRepo, err := filepath.EvalSymlinks(repo)
	if err != nil {
		return []string{fmt.Sprintf("%s: resolve repository root: %v", repo, err)}
	}
	evaluatedDirectory, err := filepath.EvalSymlinks(directory)
	if err != nil {
		return []string{fmt.Sprintf("%s: resolve skill directory: %v", skillRelativePath(repo, directory), err)}
	}
	evaluatedSkill, err := filepath.EvalSymlinks(skillFile)
	if err != nil {
		return []string{fmt.Sprintf("%s: resolve SKILL.md: %v", skillRelativePath(repo, skillFile), err)}
	}
	if !pathWithin(evaluatedRepo, evaluatedDirectory) || !pathWithin(evaluatedDirectory, evaluatedSkill) {
		return []string{skillRelativePath(repo, skillFile) + ": skill package root resolves outside the repository"}
	}
	return nil
}

func validateSkillFrontmatter(relativeSkill, directoryName string, data []byte) []string {
	normalized := strings.ReplaceAll(string(data), "\r\n", "\n")
	if !strings.HasPrefix(normalized, "---\n") {
		return []string{relativeSkill + ": frontmatter must start with ---"}
	}
	end := strings.Index(normalized[4:], "\n---\n")
	if end < 0 {
		return []string{relativeSkill + ": frontmatter must end with ---"}
	}
	frontmatterEnd := 4 + end
	frontmatter := []byte(normalized[4:frontmatterEnd])
	var metadata skillContractFrontmatter
	if err := yaml.Unmarshal(frontmatter, &metadata); err != nil {
		return []string{fmt.Sprintf("%s: decode frontmatter: %v", relativeSkill, err)}
	}
	var fields map[string]yaml.Node
	if err := yaml.Unmarshal(frontmatter, &fields); err != nil {
		return []string{fmt.Sprintf("%s: decode frontmatter fields: %v", relativeSkill, err)}
	}

	findings := make([]string, 0, 7)
	allowedFields := []string{"allowed-tools", "compatibility", "description", "license", "metadata", "name"}
	fieldNames := make([]string, 0, len(fields))
	for field := range fields {
		fieldNames = append(fieldNames, field)
	}
	sort.Strings(fieldNames)
	for _, field := range fieldNames {
		if !slices.Contains(allowedFields, field) {
			findings = append(findings, fmt.Sprintf("%s: unsupported frontmatter field %q; use only %s", relativeSkill, field, strings.Join(allowedFields, ", ")))
		}
	}
	if len(metadata.Name) > maxSkillNameLength || !skillNamePattern.MatchString(metadata.Name) {
		findings = append(findings, fmt.Sprintf("%s: name must be 1-%d lowercase letters, digits, or single hyphens", relativeSkill, maxSkillNameLength))
	}
	if metadata.Name != directoryName {
		findings = append(findings, fmt.Sprintf("%s: name %q must match its directory %q", relativeSkill, metadata.Name, directoryName))
	}
	descriptionLength := utf8.RuneCountInString(strings.TrimSpace(metadata.Description))
	if descriptionLength < 1 || descriptionLength > maxSkillDescriptionLength {
		findings = append(findings, fmt.Sprintf("%s: description must contain 1-%d characters", relativeSkill, maxSkillDescriptionLength))
	}
	if metadata.License != "MIT" {
		findings = append(findings, fmt.Sprintf("%s: license must be MIT for this repository, got %q", relativeSkill, metadata.License))
	}
	if _, provided := fields["compatibility"]; provided {
		compatibilityLength := utf8.RuneCountInString(strings.TrimSpace(metadata.Compatibility))
		if compatibilityLength < 1 || compatibilityLength > maxSkillCompatibility {
			findings = append(findings, fmt.Sprintf("%s: compatibility must contain 1-%d characters when provided", relativeSkill, maxSkillCompatibility))
		}
	}
	if _, provided := fields["allowed-tools"]; provided && strings.TrimSpace(metadata.AllowedTools) == "" {
		findings = append(findings, relativeSkill+": allowed-tools must be a non-empty space-separated string when provided")
	}
	body := normalized[frontmatterEnd+5:]
	if strings.TrimSpace(body) == "" {
		findings = append(findings, relativeSkill+": Markdown body is empty")
	}
	return findings
}

func validateSkillLinks(repo, skillRoot, sourceFile, content string, requireExists bool) []string {
	findings := make([]string, 0)
	targets, err := skillLinkTargets(sourceFile, content)
	if err != nil {
		return []string{fmt.Sprintf("%s: parse Markdown links: %v", skillRelativePath(repo, sourceFile), err)}
	}
	// The symlink check below compares an evaluated resource with this root, so the
	// root has to be evaluated too: a checkout reached through a link (macOS keeps
	// temporary directories behind /var -> /private/var) would otherwise make every
	// in-package resource look like an escape.
	evaluatedRepo, evalErr := filepath.EvalSymlinks(repo)
	if evalErr != nil {
		return []string{fmt.Sprintf("%s: resolve repository root: %v", repo, evalErr)}
	}
	for _, target := range targets {
		if target == "" || strings.HasPrefix(target, "#") {
			continue
		}
		if strings.HasPrefix(target, "~") {
			findings = append(findings, fmt.Sprintf("%s: home-relative link %q is not allowed in a portable skill package", skillRelativePath(repo, sourceFile), target))
			continue
		}
		parsed, err := url.Parse(target)
		if err != nil {
			findings = append(findings, fmt.Sprintf("%s: parse link %q: %v", skillRelativePath(repo, sourceFile), target, err))
			continue
		}
		if parsed.Scheme != "" {
			switch strings.ToLower(parsed.Scheme) {
			case "http", "https", "mailto":
				continue
			default:
				findings = append(findings, fmt.Sprintf("%s: unsupported local link scheme in %q; use a repository-relative path or an approved remote URL", skillRelativePath(repo, sourceFile), target))
				continue
			}
		}
		pathTarget, err := url.PathUnescape(parsed.Path)
		if err != nil {
			findings = append(findings, fmt.Sprintf("%s: decode link path %q: %v", skillRelativePath(repo, sourceFile), target, err))
			continue
		}
		if pathTarget == "" {
			continue
		}
		if filepath.IsAbs(filepath.FromSlash(pathTarget)) {
			findings = append(findings, fmt.Sprintf("%s: absolute local link %q is not allowed in a portable skill package", skillRelativePath(repo, sourceFile), target))
			continue
		}
		// Agent Skill package paths stay rooted at the directory containing
		// SKILL.md, including when the reference appears in a support file.
		resolved := filepath.Clean(filepath.Join(skillRoot, filepath.FromSlash(pathTarget)))
		if !pathWithin(repo, resolved) {
			findings = append(findings, fmt.Sprintf("%s: referenced resource %q escapes the repository", skillRelativePath(repo, sourceFile), target))
			continue
		}
		if !pathInsideSkillCatalog(repo, resolved) {
			findings = append(findings, fmt.Sprintf("%s: referenced resource %q is outside the first-party skill catalog; keep package links under skills/<name>/ or .agents/skills/<name>/", skillRelativePath(repo, sourceFile), target))
			continue
		}
		_, statErr := os.Stat(resolved)
		if statErr != nil {
			if requireExists {
				findings = append(findings, fmt.Sprintf("%s: missing referenced resource %q; replace it with an existing relative path", skillRelativePath(repo, sourceFile), target))
			}
			continue
		}
		evaluated, err := filepath.EvalSymlinks(resolved)
		if err != nil || !pathWithin(evaluatedRepo, evaluated) {
			findings = append(findings, fmt.Sprintf("%s: referenced resource %q resolves outside the repository", skillRelativePath(repo, sourceFile), target))
		}
	}
	return findings
}

func validateSkillOutputTemplateLinks(repo, sourceFile, content string) []string {
	findings := make([]string, 0)
	targets, err := skillLinkTargets(sourceFile, content)
	if err != nil {
		return []string{fmt.Sprintf("%s: parse output-template links: %v", skillRelativePath(repo, sourceFile), err)}
	}
	for _, target := range targets {
		if target == "" || strings.HasPrefix(target, "#") {
			continue
		}
		if strings.HasPrefix(target, "~") {
			findings = append(findings, fmt.Sprintf("%s: home-relative link %q is not allowed in a portable output template", skillRelativePath(repo, sourceFile), target))
			continue
		}
		parsed, err := url.Parse(target)
		if err != nil {
			findings = append(findings, fmt.Sprintf("%s: parse output-template link %q: %v", skillRelativePath(repo, sourceFile), target, err))
			continue
		}
		if parsed.Scheme == "" {
			// Output-relative and origin-relative paths become meaningful only after
			// the template is copied, so they are not package resource references.
			continue
		}
		switch strings.ToLower(parsed.Scheme) {
		case "http", "https", "mailto":
			continue
		default:
			findings = append(findings, fmt.Sprintf("%s: unsupported local link scheme in %q; use an output-relative path or an approved remote URL", skillRelativePath(repo, sourceFile), target))
		}
	}
	return findings
}

func pathInsideSkillCatalog(repo, path string) bool {
	relative, err := filepath.Rel(repo, path)
	if err != nil || relative == "." || relative == ".." || strings.HasPrefix(relative, ".."+string(filepath.Separator)) {
		return false
	}
	parts := strings.Split(filepath.ToSlash(relative), "/")
	var root string
	switch {
	case len(parts) >= 2 && parts[0] == "skills" && parts[1] != "":
		root = filepath.Join(repo, "skills", parts[1])
	case len(parts) >= 3 && parts[0] == ".agents" && parts[1] == "skills" && parts[2] != "":
		root = filepath.Join(repo, ".agents", "skills", parts[2])
	default:
		return false
	}
	info, err := os.Stat(filepath.Join(root, "SKILL.md"))
	return err == nil && info.Mode().IsRegular()
}

func skillLinkTargets(sourceFile, content string) ([]string, error) {
	if !isSkillMarkdownFile(sourceFile) {
		return nil, nil
	}
	source := []byte(content)
	document := goldmark.DefaultParser().Parse(markdowntext.NewReader(source))
	rawTargets := make([]string, 0)
	decodedTargets := make([]string, 0)
	err := markdownast.Walk(document, func(node markdownast.Node, entering bool) (markdownast.WalkStatus, error) {
		if !entering {
			return markdownast.WalkContinue, nil
		}
		switch node := node.(type) {
		case *markdownast.Link:
			rawTargets = append(rawTargets, string(node.Destination))
		case *markdownast.Image:
			rawTargets = append(rawTargets, string(node.Destination))
		case *markdownast.AutoLink:
			rawTargets = append(rawTargets, string(node.URL(source)))
		case *markdownast.RawHTML:
			projection, err := analyzeSkillHTML(string(node.Segments.Value(source)))
			if err != nil {
				return markdownast.WalkStop, err
			}
			decodedTargets = append(decodedTargets, projection.targets...)
		case *markdownast.HTMLBlock:
			projection, err := analyzeSkillHTML(skillHTMLBlockText(node, source))
			if err != nil {
				return markdownast.WalkStop, err
			}
			decodedTargets = append(decodedTargets, projection.targets...)
		}
		return markdownast.WalkContinue, nil
	})
	if err != nil {
		return nil, err
	}

	targets := make([]string, 0, len(rawTargets)+len(decodedTargets))
	seen := make(map[string]struct{}, len(rawTargets)+len(decodedTargets))
	appendUnique := func(target string) {
		target = strings.TrimSpace(target)
		if _, duplicate := seen[target]; duplicate {
			return
		}
		seen[target] = struct{}{}
		targets = append(targets, target)
	}
	for _, raw := range rawTargets {
		appendUnique(stdhtml.UnescapeString(raw))
	}
	for _, target := range decodedTargets {
		appendUnique(target)
	}
	return targets, nil
}

type skillHTMLProjection struct {
	visible string
	targets []string
}

func analyzeSkillHTML(content string) (skillHTMLProjection, error) {
	tokenizer := nethtml.NewTokenizer(strings.NewReader(content))
	projection := skillHTMLProjection{targets: make([]string, 0)}
	var visible strings.Builder
	hiddenElements := make([]string, 0, 1)

	for {
		switch tokenType := tokenizer.Next(); tokenType {
		case nethtml.ErrorToken:
			if err := tokenizer.Err(); !errors.Is(err, io.EOF) {
				return skillHTMLProjection{}, fmt.Errorf("tokenize HTML: %w", err)
			}
			projection.visible = visible.String()
			return projection, nil
		case nethtml.TextToken:
			if len(hiddenElements) == 0 {
				visible.WriteByte(' ')
				visible.Write(tokenizer.Text())
			}
		case nethtml.StartTagToken, nethtml.SelfClosingTagToken:
			token := tokenizer.Token()
			// Hidden markup is not visible prose, but its URL-valued attributes can become active when cloned.
			projection.targets = append(projection.targets, skillHTMLTokenTargets(token)...)
			if tokenType == nethtml.StartTagToken && isHiddenSkillHTMLElement(token.Data) {
				hiddenElements = append(hiddenElements, token.Data)
			}
		case nethtml.EndTagToken:
			if len(hiddenElements) == 0 {
				continue
			}
			token := tokenizer.Token()
			if token.Data == hiddenElements[len(hiddenElements)-1] {
				hiddenElements = hiddenElements[:len(hiddenElements)-1]
			}
		}
	}
}

func skillHTMLTokenTargets(token nethtml.Token) []string {
	targets := make([]string, 0)
	for _, attribute := range token.Attr {
		key := attribute.Key
		if key == "data" && token.Data != "object" {
			continue
		}
		if key == "srcset" {
			for _, target := range skillHTMLSourceSetTargets(attribute.Val) {
				targets = appendSkillHTMLTarget(targets, target)
			}
			continue
		}
		if slices.Contains([]string{"action", "background", "cite", "data", "formaction", "href", "poster", "src", "xlink:href"}, key) {
			targets = appendSkillHTMLTarget(targets, attribute.Val)
		}
	}
	return targets
}

func skillHTMLSourceSetTargets(value string) []string {
	targets := make([]string, 0)
	for index := 0; index < len(value); {
		for index < len(value) && (isSkillHTMLSpace(value[index]) || value[index] == ',') {
			index++
		}
		start := index
		for index < len(value) && !isSkillHTMLSpace(value[index]) {
			index++
		}
		if start == index {
			break
		}
		target := value[start:index]
		if strings.HasSuffix(target, ",") {
			target = strings.TrimRight(target, ",")
			if target != "" {
				targets = append(targets, target)
			}
			continue
		}
		targets = append(targets, target)

		parentheses := 0
		for index < len(value) {
			switch value[index] {
			case '(':
				parentheses++
			case ')':
				if parentheses > 0 {
					parentheses--
				}
			case ',':
				if parentheses == 0 {
					index++
					goto nextCandidate
				}
			}
			index++
		}
	nextCandidate:
	}
	return targets
}

func isSkillHTMLSpace(value byte) bool {
	return value == '\t' || value == '\n' || value == '\f' || value == '\r' || value == ' '
}

func appendSkillHTMLTarget(targets []string, raw string) []string {
	// A leading slash in an HTML attribute is an origin-relative web path, not a host filesystem path.
	if raw == "" || strings.HasPrefix(raw, "/") {
		return targets
	}
	return append(targets, raw)
}

func isHiddenSkillHTMLElement(name string) bool {
	return name == "script" || name == "style" || name == "template"
}

func skillHTMLBlockText(node *markdownast.HTMLBlock, source []byte) string {
	content := append([]byte(nil), node.Lines().Value(source)...)
	if node.HasClosure() {
		content = append(content, node.ClosureLine.Value(source)...)
	}
	return string(content)
}

func skillReferenceContent(sourceFile, content string, includeRemoteLinkLabels bool) (string, error) {
	if !isSkillMarkdownFile(sourceFile) {
		return " " + remoteLiteralPattern.ReplaceAllString(content, " "), nil
	}
	source := []byte(content)
	document := goldmark.DefaultParser().Parse(markdowntext.NewReader(source))
	var references strings.Builder
	appendRawReference := func(value string) {
		references.WriteByte(' ')
		references.WriteString(remoteURLPattern.ReplaceAllString(value, " "))
	}
	appendReference := func(value string) {
		appendRawReference(normalizeSkillReferenceText(value))
	}
	appendDecodedTarget := func(target string) bool {
		target = strings.TrimSpace(target)
		localPath, ok := canonicalSkillLocalTarget(target)
		if !ok {
			return false
		}
		appendRawReference(localPath)
		return true
	}
	appendTarget := func(raw string) bool {
		return appendDecodedTarget(stdhtml.UnescapeString(raw))
	}
	err := markdownast.Walk(document, func(node markdownast.Node, entering bool) (markdownast.WalkStatus, error) {
		if !entering {
			return markdownast.WalkContinue, nil
		}
		switch node := node.(type) {
		case *markdownast.Link:
			if !appendTarget(string(node.Destination)) {
				if includeRemoteLinkLabels {
					break
				}
				return markdownast.WalkSkipChildren, nil
			}
		case *markdownast.Image:
			if !appendTarget(string(node.Destination)) {
				if includeRemoteLinkLabels {
					break
				}
				return markdownast.WalkSkipChildren, nil
			}
		case *markdownast.AutoLink:
			appendTarget(string(node.URL(source)))
			return markdownast.WalkSkipChildren, nil
		case *markdownast.RawHTML:
			projection, err := analyzeSkillHTML(string(node.Segments.Value(source)))
			if err != nil {
				return markdownast.WalkStop, err
			}
			for _, target := range projection.targets {
				appendDecodedTarget(target)
			}
			return markdownast.WalkSkipChildren, nil
		case *markdownast.HTMLBlock:
			projection, err := analyzeSkillHTML(skillHTMLBlockText(node, source))
			if err != nil {
				return markdownast.WalkStop, err
			}
			for _, target := range projection.targets {
				appendDecodedTarget(target)
			}
			appendRawReference(projection.visible)
			return markdownast.WalkSkipChildren, nil
		case *markdownast.Text:
			if _, isCode := node.Parent().(*markdownast.CodeSpan); isCode {
				appendRawReference(string(node.Segment.Value(source)))
			} else {
				appendReference(string(node.Segment.Value(source)))
			}
		case *markdownast.String:
			appendReference(string(node.Value))
		case *markdownast.CodeBlock:
			appendRawReference(string(node.Lines().Value(source)))
			return markdownast.WalkSkipChildren, nil
		case *markdownast.FencedCodeBlock:
			appendRawReference(string(node.Lines().Value(source)))
			return markdownast.WalkSkipChildren, nil
		}
		return markdownast.WalkContinue, nil
	})
	return references.String(), err
}

func isSkillMarkdownFile(path string) bool {
	switch strings.ToLower(filepath.Ext(path)) {
	case ".md", ".markdown":
		return true
	default:
		return false
	}
}

func normalizeSkillReferenceText(value string) string {
	value = stdhtml.UnescapeString(value)
	var normalized strings.Builder
	normalized.Grow(len(value))
	for index := 0; index < len(value); index++ {
		if value[index] == '\\' && index+1 < len(value) && isASCIIPunctuation(value[index+1]) {
			index++
		}
		normalized.WriteByte(value[index])
	}
	return normalized.String()
}

func isASCIIPunctuation(value byte) bool {
	return value >= 0x21 && value <= 0x2F || value >= 0x3A && value <= 0x40 || value >= 0x5B && value <= 0x60 || value >= 0x7B && value <= 0x7E
}

func canonicalSkillLocalTarget(target string) (string, bool) {
	if target == "" || strings.HasPrefix(target, "#") || strings.HasPrefix(target, "~") {
		return "", false
	}
	parsed, err := url.Parse(target)
	if err != nil || parsed.Scheme != "" || parsed.Host != "" {
		return "", false
	}
	pathTarget, err := url.PathUnescape(parsed.Path)
	if err != nil || pathTarget == "" {
		return "", false
	}
	return filepath.ToSlash(filepath.Clean(filepath.FromSlash(pathTarget))), true
}

func validateSkillResources(repo, skillFile, content string) []string {
	allowedDirectories := []string{"agents", "assets", "references", "resources", "scripts", "templates", "tests"}
	skillDirectory := filepath.Dir(skillFile)
	findings := make([]string, 0)
	resources := make([]string, 0)
	openAIReferenceText := ""
	err := filepath.WalkDir(skillDirectory, func(path string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if path == skillDirectory || path == skillFile {
			return nil
		}
		relative, relErr := filepath.Rel(skillDirectory, path)
		if relErr != nil {
			return relErr
		}
		relative = filepath.ToSlash(relative)
		if entry.Name() == "__pycache__" || strings.HasSuffix(entry.Name(), ".pyc") || strings.HasSuffix(entry.Name(), ".pyo") {
			findings = append(findings, fmt.Sprintf("%s: Python bytecode cache %q is generated state; remove it from the skill package", skillRelativePath(repo, skillFile), relative))
			if entry.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}
		if slices.Contains([]string{".DS_Store", ".mypy_cache", ".pytest_cache", ".ruff_cache"}, entry.Name()) {
			findings = append(findings, fmt.Sprintf("%s: generated cache or metadata %q is not package source; remove it from the skill package", skillRelativePath(repo, skillFile), relative))
			if entry.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}
		if entry.Type()&os.ModeSymlink != 0 {
			findings = append(findings, fmt.Sprintf("%s: symbolic link %q is not allowed in a portable skill package", skillRelativePath(repo, skillFile), relative))
			if entry.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}
		first, _, nested := strings.Cut(relative, "/")
		if !slices.Contains(allowedDirectories, first) {
			findings = append(findings, skillRelativePath(repo, path)+": unsupported progressive-disclosure path; move it under agents, assets, references, resources, scripts, templates, or tests")
			if entry.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}
		if entry.IsDir() {
			return nil
		}
		if !nested {
			findings = append(findings, fmt.Sprintf("%s: unsupported progressive-disclosure path; %q must be a directory containing package resources", skillRelativePath(repo, path), first))
			return nil
		}
		info, infoErr := entry.Info()
		if infoErr != nil {
			return infoErr
		}
		if !info.Mode().IsRegular() {
			findings = append(findings, fmt.Sprintf("%s: non-regular resource %q is not allowed in a portable skill package", skillRelativePath(repo, skillFile), relative))
			return nil
		}
		if info.Mode().Perm()&0o111 != 0 && first != "scripts" {
			findings = append(findings, fmt.Sprintf("%s: executable outside scripts/ at %q; move executable behavior under scripts/ or remove the mode", skillRelativePath(repo, skillFile), relative))
		}
		if first != "assets" {
			data, readErr := readSkillParsedFile(path)
			if readErr != nil {
				return readErr
			}
			findings = append(findings, validateSkillText(skillRelativePath(repo, path), data)...)
			if first == "templates" {
				findings = append(findings, validateSkillOutputTemplateLinks(repo, path, string(data))...)
			} else {
				findings = append(findings, validateSkillLinks(repo, skillDirectory, path, string(data), true)...)
			}
			if relative == "agents/openai.yaml" {
				referenceText, referenceErr := skillReferenceContent(path, string(data), false)
				if referenceErr != nil {
					return fmt.Errorf("parse OpenAI interface references: %w", referenceErr)
				}
				openAIReferenceText = referenceText
			}
		}
		resources = append(resources, relative)
		return nil
	})
	if err != nil {
		findings = append(findings, fmt.Sprintf("%s: inspect resources: %v", skillRelativePath(repo, skillFile), err))
		return findings
	}

	rootReferenceText, err := skillReferenceContent(skillFile, content, false)
	if err != nil {
		findings = append(findings, fmt.Sprintf("%s: parse resource references: %v", skillRelativePath(repo, skillFile), err))
		return findings
	}
	for _, relative := range resources {
		if relative == "agents/openai.yaml" || hasExactSkillPathReference(rootReferenceText, relative) {
			continue
		}
		resourcePath := filepath.Join(skillDirectory, filepath.FromSlash(relative))
		relativeFromOpenAI, relErr := filepath.Rel(filepath.Join(skillDirectory, "agents"), resourcePath)
		if relErr == nil && (hasExactSkillPathReference(openAIReferenceText, relative) || hasExactSkillPathReference(openAIReferenceText, filepath.ToSlash(relativeFromOpenAI))) {
			continue
		}
		findings = append(findings, fmt.Sprintf("%s: resource %q is not directly disclosed; reference it from SKILL.md or package-owned agents/openai.yaml metadata, or remove it", skillRelativePath(repo, skillFile), relative))
	}
	return findings
}

func hasExactSkillPathReference(content, path string) bool {
	path = filepath.ToSlash(filepath.Clean(filepath.FromSlash(path)))
	for _, candidate := range []string{path, "./" + path} {
		for offset := 0; offset < len(content); {
			index := strings.Index(content[offset:], candidate)
			if index < 0 {
				break
			}
			start := offset + index
			end := start + len(candidate)
			if !skillPathContinuesBefore(content, start) && !skillPathContinuesAfter(content, end) {
				return true
			}
			offset = start + 1
		}
	}
	return false
}

func skillPathContinuesBefore(content string, index int) bool {
	if index == 0 {
		return false
	}
	// A preceding dot may be part of ../path or .hidden/path, so it is never a
	// safe left boundary for an exact package-relative path reference.
	return isSkillPathCharacter(content[index-1])
}

func skillPathContinuesAfter(content string, index int) bool {
	if index == len(content) {
		return false
	}
	value := content[index]
	if value != '.' {
		return isSkillPathCharacter(value)
	}
	return index+1 < len(content) && isSkillPathWordCharacter(content[index+1])
}

func isSkillPathCharacter(value byte) bool {
	return isSkillPathWordCharacter(value) || value == '.' || value == '/' || value == '\\'
}

func isSkillPathWordCharacter(value byte) bool {
	return value >= 'a' && value <= 'z' || value >= 'A' && value <= 'Z' || value >= '0' && value <= '9' || value == '_' || value == '-'
}

func validateSkillText(path string, data []byte) []string {
	if !utf8.Valid(data) || bytes.IndexByte(data, 0) >= 0 {
		return []string{path + ": non-text executable or instruction resource is not allowed"}
	}
	for index, value := range data {
		if value == '\r' && (index+1 == len(data) || data[index+1] != '\n') {
			return []string{path + ": unsafe control character U+000D is not allowed outside CRLF line endings"}
		}
	}
	for _, value := range string(data) {
		if value < 0x20 && value != '\t' && value != '\n' && value != '\r' || value >= 0x7F && value <= 0x9F {
			return []string{fmt.Sprintf("%s: unsafe control character U+%04X is not allowed", path, value)}
		}
		if unicode.Is(unicode.Cf, value) {
			return []string{fmt.Sprintf("%s: bidirectional or invisible Unicode control U+%04X is not allowed", path, value)}
		}
	}
	return nil
}

func validateSkillOpenAIConfig(repo, skillFile string) []string {
	path := filepath.Join(filepath.Dir(skillFile), "agents", "openai.yaml")
	agentsDirectory := filepath.Dir(path)
	agentsInfo, err := os.Lstat(agentsDirectory)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	if err != nil {
		return []string{fmt.Sprintf("%s: inspect interface metadata directory: %v", skillRelativePath(repo, agentsDirectory), err)}
	}
	if agentsInfo.Mode()&os.ModeSymlink != 0 || !agentsInfo.IsDir() {
		return []string{skillRelativePath(repo, agentsDirectory) + ": interface metadata directory must be a real directory inside the skill package"}
	}
	info, err := os.Lstat(path)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	if err != nil {
		return []string{fmt.Sprintf("%s: inspect interface metadata: %v", skillRelativePath(repo, path), err)}
	}
	if info.Mode()&os.ModeSymlink != 0 || !info.Mode().IsRegular() {
		return []string{skillRelativePath(repo, path) + ": interface metadata must be a regular, non-symlink file"}
	}
	data, err := readSkillParsedFile(path)
	if err != nil {
		return []string{fmt.Sprintf("%s: read interface metadata: %v", skillRelativePath(repo, path), err)}
	}
	var config skillOpenAIConfig
	decoder := yaml.NewDecoder(bytes.NewReader(data))
	// Product metadata is machine-consumed, so a typo must fail rather than silently disable a policy or dependency.
	decoder.KnownFields(true)
	if err := decoder.Decode(&config); err != nil {
		return []string{fmt.Sprintf("%s: decode interface metadata: %v", skillRelativePath(repo, path), err)}
	}
	var trailing any
	if err := decoder.Decode(&trailing); !errors.Is(err, io.EOF) {
		if err == nil {
			err = errors.New("multiple YAML documents are not allowed")
		}
		return []string{fmt.Sprintf("%s: decode interface metadata: %v", skillRelativePath(repo, path), err)}
	}
	relativePath := skillRelativePath(repo, path)
	findings := make([]string, 0, 10)
	if strings.TrimSpace(config.Interface.DisplayName) == "" {
		findings = append(findings, relativePath+": interface.display_name is required")
	}
	shortLength := utf8.RuneCountInString(strings.TrimSpace(config.Interface.ShortDescription))
	if shortLength < minSkillShortDescription || shortLength > maxSkillShortDescription {
		findings = append(findings, fmt.Sprintf("%s: interface.short_description must contain %d-%d characters", relativePath, minSkillShortDescription, maxSkillShortDescription))
	}
	skillName := filepath.Base(filepath.Dir(skillFile))
	if !containsExactSkillInvocation(config.Interface.DefaultPrompt, skillName) {
		findings = append(findings, fmt.Sprintf("%s: interface.default_prompt must mention $%s", relativePath, skillName))
	}
	brandColor := strings.TrimSpace(config.Interface.BrandColor)
	if brandColor != "" && !brandColorPattern.MatchString(brandColor) {
		findings = append(findings, relativePath+": interface.brand_color must be a six-digit hexadecimal color such as #3B82F6")
	}
	for index, tool := range config.Dependencies.Tools {
		prefix := fmt.Sprintf("%s: dependencies.tools[%d]", relativePath, index)
		toolType := strings.TrimSpace(tool.Type)
		if strings.TrimSpace(tool.Value) == "" {
			findings = append(findings, prefix+".value is required")
		}
		if strings.TrimSpace(tool.Description) == "" {
			findings = append(findings, prefix+".description is required")
		}
		transport := strings.TrimSpace(tool.Transport)
		toolURL := strings.TrimSpace(tool.URL)
		switch toolType {
		case "":
			findings = append(findings, prefix+".type is required")
		case "mcp":
			// A named MCP dependency uses host configuration; a remote dependency must provide the complete endpoint pair.
			if transport == "" && toolURL != "" {
				findings = append(findings, prefix+".transport is required when url is set")
			}
			if transport != "" && toolURL == "" {
				findings = append(findings, prefix+".url is required when transport is set")
			}
			if transport != "" && toolURL != "" {
				if transport != "streamable_http" {
					findings = append(findings, prefix+`.transport must be "streamable_http" for a remote MCP dependency`)
				}
				parsedURL, parseErr := url.ParseRequestURI(toolURL)
				if parseErr != nil || parsedURL.Host == "" || parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
					findings = append(findings, prefix+".url must be an absolute HTTP(S) URL")
				}
			}
		case "cli":
			if transport != "" || toolURL != "" {
				findings = append(findings, prefix+" CLI dependency must not declare transport or url")
			}
		default:
			findings = append(findings, prefix+`.type must be "mcp" or "cli"`)
		}
	}
	for field, value := range map[string]string{
		"interface.icon_large": config.Interface.IconLarge,
		"interface.icon_small": config.Interface.IconSmall,
	} {
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}
		if strings.HasPrefix(value, "~") {
			findings = append(findings, fmt.Sprintf("%s: %s path %q must be relative to the skill directory", relativePath, field, value))
			continue
		}
		parsed, parseErr := url.Parse(value)
		if parseErr != nil || parsed.Scheme != "" || parsed.Host != "" {
			findings = append(findings, fmt.Sprintf("%s: %s path %q must be a local path relative to the skill directory", relativePath, field, value))
			continue
		}
		iconPath, unescapeErr := url.PathUnescape(parsed.Path)
		if unescapeErr != nil || filepath.IsAbs(filepath.FromSlash(iconPath)) {
			findings = append(findings, fmt.Sprintf("%s: %s path %q is invalid", relativePath, field, value))
			continue
		}
		resolved := filepath.Clean(filepath.Join(filepath.Dir(skillFile), filepath.FromSlash(iconPath)))
		if !pathWithin(filepath.Dir(skillFile), resolved) || !pathWithin(repo, resolved) {
			findings = append(findings, fmt.Sprintf("%s: %s path %q escapes the skill package", relativePath, field, value))
			continue
		}
		info, statErr := os.Lstat(resolved)
		if statErr != nil {
			findings = append(findings, fmt.Sprintf("%s: %s path %q does not exist", relativePath, field, value))
			continue
		}
		if info.Mode()&os.ModeSymlink != 0 || !info.Mode().IsRegular() {
			findings = append(findings, fmt.Sprintf("%s: %s path %q must be a regular, non-symlink file", relativePath, field, value))
		}
	}
	return findings
}

func containsExactSkillInvocation(prompt, skillName string) bool {
	needle := "$" + skillName
	for offset := 0; offset < len(prompt); {
		index := strings.Index(prompt[offset:], needle)
		if index < 0 {
			return false
		}
		start := offset + index
		end := start + len(needle)
		leftBoundary := start == 0 || !isSkillInvocationContinuation(prompt[start-1])
		rightBoundary := end == len(prompt) || !isSkillInvocationContinuation(prompt[end])
		if leftBoundary && rightBoundary {
			return true
		}
		offset = start + 1
	}
	return false
}

func isSkillInvocationContinuation(value byte) bool {
	return value >= 'a' && value <= 'z' || value >= 'A' && value <= 'Z' || value >= '0' && value <= '9' || value == '-' || value == '_' || value == '$'
}

func checkSkillCountDocumentation(t *testing.T, repo string, count int) {
	t.Helper()
	path := filepath.Join(repo, "README.md")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	want := fmt.Sprintf("library of %d reusable", count)
	if !strings.Contains(string(data), want) {
		t.Errorf("README.md: skill count drift; describe the current %d global skills with %q", count, want)
	}
}

func pathWithin(root, path string) bool {
	relative, err := filepath.Rel(filepath.Clean(root), filepath.Clean(path))
	return err == nil && relative != ".." && !strings.HasPrefix(relative, ".."+string(filepath.Separator)) && !filepath.IsAbs(relative)
}

func skillRelativePath(repo, path string) string {
	relative, err := filepath.Rel(repo, path)
	if err != nil {
		return path
	}
	return filepath.ToSlash(relative)
}

func TestSkillContractFullGateInstructionsPreserveDirtyWorktrees(t *testing.T) {
	repo := skillRepositoryRoot(t)
	skillFiles, findings := discoverSkillFiles(repo)
	if len(findings) != 0 {
		t.Fatalf("discover skills: %v", findings)
	}
	required := []string{
		"inspect the full gate's task definition and working-tree state",
		"whole-tree write-formatters",
		"isolated temporary worktree",
		"never reformat unrelated work",
	}
	for _, skillFile := range skillFiles {
		data, err := readSkillParsedFile(skillFile)
		if err != nil {
			t.Fatal(err)
		}
		content := strings.ToLower(string(data))
		if !strings.Contains(content, "`mise run all`") {
			continue
		}
		for _, phrase := range required {
			if !strings.Contains(content, phrase) {
				t.Errorf("%s: full-gate instruction is missing dirty-worktree safeguard %q", skillRelativePath(repo, skillFile), phrase)
			}
		}
	}
}

func TestHighRiskSkillSmokeContracts(t *testing.T) {
	repo := skillRepositoryRoot(t)
	tests := []struct {
		name     string
		path     string
		required []string
		ordered  []string
		commands [][]string
	}{
		{
			name:     "release gates publication",
			path:     "skills/release/SKILL.md",
			ordered:  []string{"Clean working tree", "git tag -a", "git push --atomic", "gh release create \"$tag\""},
			commands: [][]string{{"release"}, {"r"}},
		},
		{
			name:     "kubernetes verifies and stops local state",
			path:     "skills/k8s-local/SKILL.md",
			required: []string{"OFF by Default", "docker info", "--kubeconfig ~/.kube/dot/local.yaml --context k3d-local", "dot cluster start", "dot cluster stop"},
			commands: [][]string{{"cluster", "start"}, {"k", "s"}, {"cluster", "stop"}},
		},
		{
			name:    "git publication validates before push",
			path:    "skills/git-add-commit-push/SKILL.md",
			ordered: []string{"git status --short", "git commit -m", "git push", "Do not use `--no-verify`"},
		},
		{
			name:     "repository review preserves proof boundaries",
			path:     "skills/repository-review/SKILL.md",
			required: []string{"staged, unstaged, and untracked", "isolated temporary worktree", "mise run all", "source-ready", "local-green", "exact-head-CI", "runtime-proven", "deployed", "release-published", "Key findings", "Actions"},
		},
		{
			name:     "project backlog separates drafts from mutation",
			path:     "skills/project-backlog/SKILL.md",
			required: []string{"read-only", "explicitly authorizes", "open and closed", "verified finding", "trend opportunity", "confirmed repository", "private", "unavailable", "Unauthorized writes", "partial issue creation", "addBlockedBy", "Mutation receipt", "tests/behavioral-evaluations.md"},
		},
		{
			name:     "agent skills covers lifecycle and publication authority",
			path:     "skills/agent-skills/SKILL.md",
			required: []string{"# Agent Skills", "## Author", "## Validate", "## Publish", "## Install", "<candidate-root>/<slug>/SKILL.md", "gh skill publish --dry-run <candidate-root>", "gh skill publish <candidate-root> --tag <vX.Y.Z>", "exact reviewed local snapshot", "explicit publication authorization", "Publication receipt", "source-ready"},
			ordered:  []string{"## Author", "## Validate", "## Publish", "## Install"},
		},
		{
			name:     "product spec scales compact output without dropping invariants",
			path:     "skills/product-loop/SKILL.md",
			required: []string{"compact, low-risk", "combine or omit immaterial sections", "Always retain the decision summary", "acceptance and edge cases", "data and trust boundaries", "open decisions"},
		},
		{
			name:     "implementation plan scales compact output without fake topology",
			path:     "skills/implementation-plan/SKILL.md",
			required: []string{"two or three low-risk linear slices", "compact task table", "red-green proof", "one-line dependency chain", "Omit empty parallel-lane", "never omit a real risk or proof boundary"},
		},
		{
			name:     "terraform scaffold passes static scan variables",
			path:     "skills/terraform-stack/references/mise.toml",
			required: []string{"trivy config --exit-code 1 --tf-vars terraform.example.tfvars ."},
		},
	}
	app := NewApp()
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			data, err := readSkillParsedFile(filepath.Join(repo, test.path))
			if err != nil {
				t.Fatal(err)
			}
			content := string(data)
			for _, required := range test.required {
				if !strings.Contains(content, required) {
					t.Errorf("%s: missing offline smoke invariant %q; restore the documented safety step", test.path, required)
				}
			}
			previous := -1
			for _, required := range test.ordered {
				position := strings.Index(content, required)
				if position < 0 {
					t.Errorf("%s: missing offline smoke invariant %q; restore the documented safety step", test.path, required)
					continue
				}
				if position <= previous {
					t.Errorf("%s: safety step %q is out of order; keep validation before mutation and publication", test.path, required)
				}
				previous = position
			}
			for _, commandPath := range test.commands {
				if !hasCommandPath(app, commandPath) {
					t.Errorf("%s: documented dot command %q is unavailable; replace it with live command metadata", test.path, strings.Join(commandPath, " "))
				}
			}
		})
	}
}

func hasCommandPath(root *cli.Command, path []string) bool {
	current := root
	for _, part := range path {
		var found *cli.Command
		for _, command := range current.Commands {
			if command.Name == part || slices.Contains(command.Aliases, part) {
				found = command
				break
			}
		}
		if found == nil {
			return false
		}
		current = found
	}
	return true
}
