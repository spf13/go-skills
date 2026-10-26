package check

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const demoDescription = "Use when writing demo Go code, testing demo packages, or reviewing demo layout. Not for other skills."

// fixture is a throwaway repository the checks run against.
type fixture struct {
	root string
}

// newFixture writes the smallest repository that satisfies every rule, so each
// test can break exactly one rule and assert the matching message.
func newFixture(t *testing.T) *fixture {
	t.Helper()
	f := &fixture{root: t.TempDir()}
	f.write(t, ".claude-plugin/marketplace.json", `{
  "name": "demo-market",
  "owner": {"name": "Demo"},
  "plugins": [
    {"name": "demo", "source": "./demo", "category": "development", "tags": ["go", "demo"]}
  ]
}`)
	f.write(t, "demo/.claude-plugin/plugin.json", `{
  "name": "demo",
  "description": "`+demoDescription+`",
  "version": "1.0.0",
  "keywords": ["go", "demo"]
}`)
	f.write(t, "demo/SKILL.md", skillBody(demoDescription))
	f.write(t, "README.md", "# Demo\n\n| Skill | Use when |\n| --- | --- |\n| [`demo`](./demo/SKILL.md) | demo work |\n")
	f.write(t, "CHANGELOG.md", "# Changelog\n\n## [1.0.0] - 2026-10-26\n\nFirst release.\n")
	return f
}

func skillBody(description string) string {
	return "---\nname: demo\ndescription: >\n  " + description + "\n---\n\n# Demo\n\nBody.\n\n" +
		defersSection + "\n\nUse `cobra-viper` instead.\n"
}

func (f *fixture) write(t *testing.T, path, content string) {
	t.Helper()
	full := filepath.Join(f.root, filepath.FromSlash(path))
	if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
		t.Fatalf("mkdir %s: %v", path, err)
	}
	if err := os.WriteFile(full, []byte(content), 0o644); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}

// check runs every rule against the fixture.
func (f *fixture) check(t *testing.T, opts Options) Report {
	t.Helper()
	opts.Root = f.root
	report, err := Run(opts)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	return report
}

// TestBaselineFixture proves the fixture is green: every other test is a
// single-rule deviation from it.
func TestBaselineFixture(t *testing.T) {
	f := newFixture(t)
	report := f.check(t, Options{})
	if !report.OK() {
		t.Fatalf("baseline fixture should pass, got %v", report.Issues)
	}
	if report.Version != "1.0.0" {
		t.Errorf("version = %q, want 1.0.0", report.Version)
	}
	if len(report.Skills) != 1 || report.Skills[0] != "demo" {
		t.Errorf("skills = %v, want [demo]", report.Skills)
	}
}

func TestSchemaValidation(t *testing.T) {
	f := newFixture(t)
	// `source` is required by the marketplace schema.
	f.write(t, ".claude-plugin/marketplace.json", `{
  "name": "demo-market",
  "owner": {"name": "Demo"},
  "plugins": [{"name": "demo", "category": "development", "tags": ["go"]}]
}`)
	report := f.check(t, Options{})
	assertIssue(t, report, ".claude-plugin/marketplace.json", `missing required property "source"`)

	// Wrong type inside a nested manifest property.
	f = newFixture(t)
	f.write(t, "demo/.claude-plugin/plugin.json", `{"name": "demo", "version": 7}`)
	report = f.check(t, Options{})
	assertIssue(t, report, "demo/.claude-plugin/plugin.json", "/version: want string, got number")
}

func TestManifestInvariants(t *testing.T) {
	t.Run("tags must equal keywords", func(t *testing.T) {
		f := newFixture(t)
		f.write(t, "demo/.claude-plugin/plugin.json", `{"name": "demo", "description": "`+demoDescription+`", "version": "1.0.0", "keywords": ["go"]}`)
		assertIssue(t, f.check(t, Options{}), "demo/.claude-plugin/plugin.json", "marketplace tags")
	})

	t.Run("catalog and directories correspond", func(t *testing.T) {
		f := newFixture(t)
		f.write(t, "extra/.claude-plugin/plugin.json", `{"name": "extra", "description": "`+demoDescription+`", "version": "1.0.0", "keywords": ["go", "demo"]}`)
		f.write(t, "extra/SKILL.md", skillBody(demoDescription))
		assertIssue(t, f.check(t, Options{}), "marketplace.json", `skill directory "extra" has no matching marketplace entry`)
	})

	t.Run("README lists every skill", func(t *testing.T) {
		f := newFixture(t)
		f.write(t, "README.md", "# Demo\n")
		assertIssue(t, f.check(t, Options{}), "README.md", "skills table must list every skill exactly once")
	})

	t.Run("versions move together", func(t *testing.T) {
		f := newFixture(t)
		f.write(t, ".claude-plugin/marketplace.json", `{
  "name": "demo-market",
  "owner": {"name": "Demo"},
  "plugins": [
    {"name": "demo", "source": "./demo", "category": "development", "tags": ["go", "demo"]},
    {"name": "other", "source": "./other", "category": "development", "tags": ["go", "demo"]}
  ]
}`)
		f.write(t, "other/.claude-plugin/plugin.json", `{"name": "other", "description": "`+demoDescription+`", "version": "2.0.0", "keywords": ["go", "demo"]}`)
		f.write(t, "other/SKILL.md", strings.Replace(skillBody(demoDescription), "name: demo", "name: other", 1))
		f.write(t, "README.md", "| [`demo`](./demo/SKILL.md) | x |\n| [`other`](./other/SKILL.md) | y |\n")
		assertIssue(t, f.check(t, Options{}), "marketplace.json", "plugin versions must be identical")
	})

	t.Run("CHANGELOG entry exists", func(t *testing.T) {
		f := newFixture(t)
		f.write(t, "CHANGELOG.md", "# Changelog\n\n## [0.9.0] - 2026-10-26\n")
		assertIssue(t, f.check(t, Options{}), "CHANGELOG.md", "no '## [1.0.0]' entry")
	})

	t.Run("entry keys are exact", func(t *testing.T) {
		f := newFixture(t)
		f.write(t, ".claude-plugin/marketplace.json", `{
  "name": "demo-market",
  "owner": {"name": "Demo"},
  "plugins": [{"name": "demo", "source": "./demo", "category": "development", "tags": ["go", "demo"], "description": "x"}]
}`)
		assertIssue(t, f.check(t, Options{}), "marketplace.json", "keys must be exactly")
	})
}

func TestSkillContentRules(t *testing.T) {
	t.Run("line budget", func(t *testing.T) {
		f := newFixture(t)
		body := skillBody(demoDescription) + strings.Repeat("filler line\n", maxSkillLines)
		f.write(t, "demo/SKILL.md", body)
		assertIssue(t, f.check(t, Options{}), "demo/SKILL.md", "maximum is 500")
	})

	t.Run("token budget", func(t *testing.T) {
		f := newFixture(t)
		line := strings.Repeat("word ", 20) + "\n"
		body := skillBody(demoDescription) + strings.Repeat(line, 200)
		f.write(t, "demo/SKILL.md", body)
		assertIssue(t, f.check(t, Options{}), "demo/SKILL.md", "tokens; maximum is 5000")
	})

	t.Run("sandbox path rejected", func(t *testing.T) {
		f := newFixture(t)
		f.write(t, "demo/SKILL.md", skillBody(demoDescription)+"\nRead /mnt/skills/user/go/SKILL.md first.\n")
		assertIssue(t, f.check(t, Options{}), "demo/SKILL.md", "machine-specific sandbox path")
	})

	t.Run("negative trigger section", func(t *testing.T) {
		f := newFixture(t)
		f.write(t, "demo/SKILL.md", strings.Replace(skillBody(demoDescription), defersSection, "## Related Skills", 1))
		assertIssue(t, f.check(t, Options{}), "demo/SKILL.md", "missing \"## When NOT to Use This Skill\"")
	})

	t.Run("placeholders", func(t *testing.T) {
		f := newFixture(t)
		body := skillBody(demoDescription) + "\nDrop the spec into [SPEC_FILE_PATH] first.\n"
		f.write(t, "demo/SKILL.md", body)
		line := lineOf(body, "[SPEC_FILE_PATH]")
		assertIssueAt(t, f.check(t, Options{}), "demo/SKILL.md", line, `unresolved placeholder "[SPEC_FILE_PATH]"`)
	})

	t.Run("placeholders inside fenced code are allowed", func(t *testing.T) {
		f := newFixture(t)
		f.write(t, "demo/SKILL.md", skillBody(demoDescription)+"\n```text\nreplace [FILL_THIS_IN]\n```\n")
		if report := f.check(t, Options{}); !report.OK() {
			t.Fatalf("fenced tokens should be ignored, got %v", report.Issues)
		}
	})

	t.Run("references resolve", func(t *testing.T) {
		f := newFixture(t)
		f.write(t, "demo/SKILL.md", skillBody(demoDescription)+"\nSee [details](references/details.md).\n")
		assertIssue(t, f.check(t, Options{}), "demo/SKILL.md", `reference "references/details.md" does not exist`)

		f.write(t, "demo/references/details.md", "# Details\n")
		if report := f.check(t, Options{}); !report.OK() {
			t.Fatalf("existing reference should pass, got %v", report.Issues)
		}
	})
}

func TestDescriptionStandard(t *testing.T) {
	cases := map[string]string{
		"trigger-only opening": "Reviews demo Go packages and validates demo layout conventions.",
		"minimum length":       "Use when Go.",
		"coercive activation":  demoDescription + " Always use this skill.",
	}
	for name, description := range cases {
		t.Run(name, func(t *testing.T) {
			f := newFixture(t)
			f.write(t, "demo/SKILL.md", skillBody(description))
			f.write(t, "demo/.claude-plugin/plugin.json", `{"name": "demo", "description": "`+description+`", "version": "1.0.0", "keywords": ["go", "demo"]}`)
			if report := f.check(t, Options{}); report.OK() {
				t.Fatal("expected a description violation")
			}
		})
	}
}

func TestGoReleaseClaims(t *testing.T) {
	claiming := strings.Replace(skillBody(demoDescription), "name: demo", "name: go", 1)
	claiming = strings.Replace(claiming, demoDescription, demoDescription+" current through Go 1.27.", 1)

	f := newFixture(t)
	if err := os.RemoveAll(filepath.Join(f.root, "demo")); err != nil {
		t.Fatalf("remove demo: %v", err)
	}
	f.write(t, ".claude-plugin/marketplace.json", `{
  "name": "demo-market",
  "owner": {"name": "Demo"},
  "plugins": [{"name": "go", "source": "./go", "category": "languages", "tags": ["go", "demo"]}]
}`)
	f.write(t, "go/.claude-plugin/plugin.json", `{"name": "go", "description": "`+demoDescription+` current through Go 1.27.", "version": "1.0.0", "keywords": ["go", "demo"]}`)
	f.write(t, "go/SKILL.md", claiming)
	f.write(t, "README.md", "| [`go`](./go/SKILL.md) | x |\n")

	if report := f.check(t, Options{Latest: "go1.27.3"}); !report.OK() {
		t.Fatalf("matching claim should pass, got %v", report.Issues)
	}
	assertIssue(t, f.check(t, Options{Latest: "go1.28.0"}), "go/SKILL.md", "the claim is behind the release")
	assertIssue(t, f.check(t, Options{Latest: "1.26"}), "go/SKILL.md", "the claim is ahead of the release")

	// The claim must not disappear from a known claiming skill.
	f.write(t, "go/SKILL.md", claimingWithoutClaim(t, claiming))
	assertIssue(t, f.check(t, Options{}), "go/SKILL.md", "lost its 'current through Go X.Y' claim")
}

// claimingWithoutClaim strips the release claim from a skill body.
func claimingWithoutClaim(t *testing.T, body string) string {
	t.Helper()
	return strings.Replace(body, " current through Go 1.27.", ".", 1)
}

func TestSnippetCompilation(t *testing.T) {
	if testing.Short() {
		t.Skip("compiles Go samples")
	}

	t.Run("stdlib sample builds", func(t *testing.T) {
		f := newFixture(t)
		f.write(t, "demo/SKILL.md", skillBody(demoDescription)+"\n```go\npackage main\n\nimport \"fmt\"\n\nfunc main() { fmt.Println(\"hi\") }\n```\n")
		report := f.check(t, Options{})
		if !report.OK() {
			t.Fatalf("valid sample should build, got %v", report.Issues)
		}
		if report.CompiledSamples != 1 {
			t.Errorf("compiled samples = %d, want 1", report.CompiledSamples)
		}
	})

	t.Run("broken stdlib sample fails", func(t *testing.T) {
		f := newFixture(t)
		f.write(t, "demo/SKILL.md", skillBody(demoDescription)+"\n```go\npackage main\n\nfunc main() { fmt.Println(\"hi\") }\n```\n")
		assertIssue(t, f.check(t, Options{}), "demo/SKILL.md", "complete Go sample does not compile")
	})

	t.Run("sample importing another module is skipped", func(t *testing.T) {
		f := newFixture(t)
		f.write(t, "demo/SKILL.md", skillBody(demoDescription)+"\n```go\npackage main\n\nimport \"github.com/spf13/cobra\"\n\nfunc main() { _ = cobra.Command{} }\n```\n")
		report := f.check(t, Options{})
		if !report.OK() {
			t.Fatalf("unresolvable import should be skipped, got %v", report.Issues)
		}
		if report.SkippedSamples != 1 || report.CompiledSamples != 0 {
			t.Errorf("compiled=%d skipped=%d, want 0/1", report.CompiledSamples, report.SkippedSamples)
		}
	})

	t.Run("partial sample is not compiled", func(t *testing.T) {
		f := newFixture(t)
		f.write(t, "demo/SKILL.md", skillBody(demoDescription)+"\n```go\nfor _, x := range xs {\n\t...\n}\n```\n")
		report := f.check(t, Options{})
		if !report.OK() || report.CompiledSamples != 0 || report.SkippedSamples != 0 {
			t.Fatalf("partial sample should be ignored, got %v", report.Issues)
		}
	})
}

func TestFrontmatterFolding(t *testing.T) {
	fm, ok := parseFrontmatter("---\nname: demo\ndescription: >\n  Use when a\n  b happens.\nmetadata:\n  version: 3\n---\nbody\n")
	if !ok {
		t.Fatal("frontmatter not parsed")
	}
	if got, want := fm.Fields["description"], "Use when a b happens."; got != want {
		t.Errorf("description = %q, want %q", got, want)
	}
	if _, ok := parseFrontmatter("# no frontmatter\n"); ok {
		t.Error("frontmatter parsed from a file without one")
	}
}

// TestRepository runs every rule against this repository, which is what CI does.
func TestRepository(t *testing.T) {
	report, err := Run(Options{})
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if !report.OK() {
		t.Fatalf("repository violates its own rules:\n%v", report.Issues)
	}
}

// lineOf reports the 1-based line holding fragment.
func lineOf(body, fragment string) int {
	for i, line := range strings.Split(body, "\n") {
		if strings.Contains(line, fragment) {
			return i + 1
		}
	}
	return 0
}

// assertIssueAt fails unless an issue at path and line contains fragment.
func assertIssueAt(t *testing.T, report Report, path string, line int, fragment string) {
	t.Helper()
	for _, issue := range report.Issues {
		if strings.Contains(issue.Path, path) && issue.Line == line && strings.Contains(issue.Msg, fragment) {
			return
		}
	}
	t.Fatalf("no issue at %s:%d containing %q; got %v", path, line, fragment, report.Issues)
}

// assertIssue fails unless some issue points at path and contains fragment.
func assertIssue(t *testing.T, report Report, path, fragment string) {
	t.Helper()
	for _, issue := range report.Issues {
		if strings.Contains(issue.Path, path) && strings.Contains(issue.Msg, fragment) {
			return
		}
	}
	t.Fatalf("no issue at %q containing %q; got %v", path, fragment, report.Issues)
}
