// Package check enforces the rules documented in CONTRIBUTING.md.
//
// Every rule in CONTRIBUTING.md maps to a check here. If you add a rule there,
// add the check here too — and vice versa. The package is hermetic: it reads the
// working tree, the local git tags, and the Go toolchain, and it never touches
// the network.
package check

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// Options configures a run.
type Options struct {
	// Root is the repository root. Empty resolves by walking up from the working
	// directory to the directory holding the marketplace manifest.
	Root string
	// Latest is the latest stable Go release, such as "go1.27.3" or "1.27".
	// Empty skips the release-currency comparison: the scheduled workflow always
	// supplies the value it fetched from go.dev, so a local run without network
	// stays useful.
	Latest string
}

// Report is the outcome of a run.
type Report struct {
	// Skills lists the declared skills that were checked, sorted by name.
	Skills []string
	// Version is the single catalog version; empty when the plugins disagree.
	Version string
	// Latest is the Go release the claims were compared against; empty when the
	// caller supplied none and the comparison was skipped.
	Latest string
	// CompiledSamples counts the complete Go fences that built and vetted
	// cleanly.
	CompiledSamples int
	// SkippedSamples counts the complete Go fences not built because they import
	// another module, which this hermetic check cannot resolve.
	SkippedSamples int
	// Issues holds the violations, sorted by path, line, and message.
	Issues []Issue
}

// OK reports whether the repository satisfied every rule.
func (r Report) OK() bool { return len(r.Issues) == 0 }

// Issue is one rule violation, reported against a repository-relative path.
type Issue struct {
	Path string // slash-separated, relative to the repository root
	Line int    // 1-based; 0 when the issue is not tied to a line
	Msg  string
}

func (i Issue) String() string {
	if i.Line > 0 {
		return fmt.Sprintf("%s:%d: %s", i.Path, i.Line, i.Msg)
	}
	return fmt.Sprintf("%s: %s", i.Path, i.Msg)
}

// Run executes every check.
func Run(opts Options) (Report, error) {
	root := opts.Root
	if root == "" {
		resolved, err := FindRoot()
		if err != nil {
			return Report{}, err
		}
		root = resolved
	}
	c := &checker{root: root, latest: strings.TrimSpace(opts.Latest)}

	for _, stage := range []func() error{c.manifests, c.skillContent, c.catalogVersions, c.releaseClaims, c.goFences} {
		if err := stage(); err != nil {
			return Report{}, err
		}
	}

	sort.Slice(c.issues, func(i, j int) bool {
		a, b := c.issues[i], c.issues[j]
		if a.Path != b.Path {
			return a.Path < b.Path
		}
		if a.Line != b.Line {
			return a.Line < b.Line
		}
		return a.Msg < b.Msg
	})

	report := Report{
		Latest:          c.latest,
		Version:         c.version(),
		CompiledSamples: c.compiledSamples,
		SkippedSamples:  c.skippedSamples,
		Issues:          c.issues,
	}
	for _, s := range c.skills {
		report.Skills = append(report.Skills, s.name)
	}
	return report, nil
}

// FindRoot walks up from the working directory to the repository root.
func FindRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, ".claude-plugin", "marketplace.json")); err == nil {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf("no .claude-plugin/marketplace.json above %s", dir)
		}
		dir = parent
	}
}

type checker struct {
	root   string
	latest string
	issues []Issue

	catalog         []map[string]any  // marketplace entries
	skills          []skill           // declared skills, sorted by name
	pluginVersions  map[string]string // skill name -> plugin.json version
	compiledSamples int
	skippedSamples  int
}

// skill is a root-level directory published as a single-skill plugin.
type skill struct {
	name string
	dir  string
}

// taxonomy is the closed set of marketplace categories.
var taxonomy = map[string]bool{"languages": true, "development": true, "frameworks": true}

// entryKeys is the exact set of keys a marketplace entry may carry.
var entryKeys = []string{"category", "name", "source", "tags"}

func (c *checker) fail(path string, line int, format string, args ...any) {
	c.issues = append(c.issues, Issue{Path: path, Line: line, Msg: fmt.Sprintf(format, args...)})
}

// rel converts a filesystem path to the slash-separated relative form used in
// reports.
func (c *checker) rel(path string) string {
	rel, err := filepath.Rel(c.root, path)
	if err != nil {
		return filepath.ToSlash(path)
	}
	return filepath.ToSlash(rel)
}

// read returns the file contents, reporting an unreadable file as an issue.
func (c *checker) read(path string) ([]byte, bool) {
	raw, err := os.ReadFile(path)
	if err != nil {
		c.fail(c.rel(path), 0, "%v", err)
		return nil, false
	}
	return raw, true
}

// declaredSkills lists root-level directories that publish a plugin manifest.
func (c *checker) declaredSkills() error {
	dirents, err := os.ReadDir(c.root)
	if err != nil {
		return err
	}
	var found []skill
	for _, entry := range dirents {
		if !entry.IsDir() {
			continue
		}
		dir := filepath.Join(c.root, entry.Name())
		if _, err := os.Stat(filepath.Join(dir, ".claude-plugin", "plugin.json")); err == nil {
			found = append(found, skill{name: entry.Name(), dir: dir})
		}
	}
	sort.Slice(found, func(i, j int) bool { return found[i].name < found[j].name })
	c.skills = found
	return nil
}

// ensureSkills discovers the declared skills once per run.
func (c *checker) ensureSkills() error {
	if c.skills != nil {
		return nil
	}
	return c.declaredSkills()
}

// frontmatter is the parsed leading YAML block of a Markdown file.
type frontmatter struct {
	Fields map[string]string
}

// parseFrontmatter reads the leading `---` block. Folded (>) and literal (|)
// scalars have their continuation lines folded into one space-joined value,
// which is how the description is written in every shipped skill.
func parseFrontmatter(src string) (frontmatter, bool) {
	lines := strings.Split(src, "\n")
	if len(lines) == 0 || strings.TrimSpace(lines[0]) != "---" {
		return frontmatter{}, false
	}
	fm := frontmatter{Fields: map[string]string{}}
	key := ""
	for _, line := range lines[1:] {
		if strings.TrimSpace(line) == "---" {
			return fm, true
		}
		switch {
		case line != "" && line[0] != ' ' && line[0] != '\t' && strings.Contains(line, ":"):
			name, value, _ := strings.Cut(line, ":")
			key = strings.TrimSpace(name)
			value = strings.TrimSpace(value)
			switch value {
			case ">", "|", ">-", "|-", ">+", "|+":
				fm.Fields[key] = ""
			default:
				fm.Fields[key] = value
			}
		case key != "" && line != "" && (line[0] == ' ' || line[0] == '\t'):
			fm.Fields[key] = strings.TrimSpace(fm.Fields[key] + " " + strings.TrimSpace(line))
		}
	}
	return frontmatter{}, false
}

// version is the single catalog version shared by every plugin, or "" when the
// plugins disagree.
func (c *checker) version() string {
	var only string
	for _, version := range c.pluginVersions {
		if only != "" && only != version {
			return ""
		}
		only = version
	}
	return only
}

func keysOf(m map[string]any) []string {
	keys := make([]string, 0, len(m))
	for key := range m {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

func sortedNames(set map[string]bool) []string {
	names := make([]string, 0, len(set))
	for name := range set {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

func sortStrings(list []string) []string {
	sorted := append([]string{}, list...)
	sort.Strings(sorted)
	return sorted
}

func equalStrings(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func equalSets(a, b []string) bool {
	return equalStrings(sortStrings(a), sortStrings(b))
}

func hasDuplicates(list []string) bool {
	seen := map[string]bool{}
	for _, item := range list {
		if seen[item] {
			return true
		}
		seen[item] = true
	}
	return false
}
