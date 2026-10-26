package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

var (
	excluded   = map[string]bool{"templates": true, ".github": true, ".git": true, ".claude-plugin": true, ".codegraph": true}
	taxonomy   = map[string]bool{"languages": true, "development": true, "frameworks": true}
	entryKeys  = []string{"category", "name", "source", "tags"}
	claimRE    = regexp.MustCompile(`(?i)current through Go (\d+\.\d+)`)
	minorRE    = regexp.MustCompile(`(\d+)\.(\d+)`)
	placeholderRE = regexp.MustCompile(`\[[A-Z][A-Z_ ]{2,}\]`)
	coercionPhrases = []string{"even if they don't ask", "recommend them", "always use this skill", "Recommend these libraries even when they aren't named", "CRITICAL:"}
	machinePathRE = regexp.MustCompile(`/mnt/skills|/home/[^/]+/\.agents`)
	claimingSkills = map[string]bool{"go": true}
)

var errors []string

func fail(msg string) { errors = append(errors, msg) }

func root() string {
	if v := os.Getenv("CHECK_ROOT"); v != "" {
		return v
	}
	// walk up from executable or cwd
	cwd, _ := os.Getwd()
	return cwd
}

func readJSON(path string) map[string]any {
	b, err := os.ReadFile(path)
	if err != nil { fail(fmt.Sprintf("read %s: %v", path, err)); return nil }
	var m map[string]any
	if err := json.Unmarshal(b, &m); err != nil { fail(fmt.Sprintf("parse %s: %v", path, err)); return nil }
	return m
}

func frontmatter(path string) map[string]string {
	b, err := os.ReadFile(path)
	if err != nil { return nil }
	lines := strings.Split(string(b), "\n")
	if len(lines) == 0 || strings.TrimSpace(lines[0]) != "---" { return nil }
	fields := map[string]string{}
	var key string
	for _, line := range lines[1:] {
		if strings.TrimSpace(line) == "---" { break }
		if strings.TrimSpace(line) != "" && strings.Contains(line, ":") && line[0] != ' ' && line[0] != '\t' {
			parts := strings.SplitN(line, ":", 2)
			key = strings.TrimSpace(parts[0])
			val := strings.TrimSpace(parts[1])
			if val == ">" || val == "|" || val == ">-" || val == "|-" || val == ">+" || val == "|+" {
				fields[key] = ""
			} else {
				fields[key] = val
			}
		} else if key != "" && len(line) > 0 && (line[0] == ' ' || line[0] == '\t') {
			fields[key] = strings.TrimSpace(fields[key] + " " + strings.TrimSpace(line))
		}
	}
	return fields
}

func declaredSkills(r string) map[string]string {
	m := map[string]string{}
	entries, _ := os.ReadDir(r)
	for _, e := range entries {
		if !e.IsDir() || excluded[e.Name()] { continue }
		if _, err := os.Stat(filepath.Join(r, e.Name(), ".claude-plugin", "plugin.json")); err == nil {
			m[e.Name()] = filepath.Join(r, e.Name())
		}
	}
	return m
}

func main() {
	r := root()
	exitCode := 0

	// --- Manifests ---
	marketplace := readJSON(filepath.Join(r, ".claude-plugin", "marketplace.json"))
	var entries []map[string]any
	if marketplace != nil {
		if v, ok := marketplace["plugins"].([]any); ok {
			for _, e := range v {
				if m, ok := e.(map[string]any); ok {
					entries = append(entries, m)
				}
			}
		}
	}
	skills := declaredSkills(r)

	// correspondence
	entryNames := map[string]bool{}
	for _, e := range entries {
		if n, ok := e["name"].(string); ok {
			if entryNames[n] { fail(fmt.Sprintf("marketplace.json duplicate entry %q", n)) }
			entryNames[n] = true
			if _, ok := skills[n]; !ok {
				fail(fmt.Sprintf("marketplace entry %q has no matching skill directory", n))
			}
		}
	}
	for name := range skills {
		if !entryNames[name] {
			fail(fmt.Sprintf("skill directory %q has no matching marketplace entry", name))
		}
	}

	// README table
	readmeTableCheck(r, entryNames)

	// per-skill checks
	versions := map[string]string{}
	for _, e := range entries {
		name, _ := e["name"].(string)
		// entry keys
		keys := make([]string, 0, len(e))
		for k := range e { keys = append(keys, k) }
		sort.Strings(keys)
		sort.Strings(entryKeys)
		if len(keys) != len(entryKeys) {
			fail(fmt.Sprintf("entry %q: keys must be %v, got %v", name, entryKeys, keys))
		} else {
			for i, k := range keys {
				if k != entryKeys[i] {
					fail(fmt.Sprintf("entry %q: keys must be %v, got %v", name, entryKeys, keys))
					break
				}
			}
		}
		if e["source"] != "./"+name {
			fail(fmt.Sprintf("entry %q: source must be %q, got %q", name, "./"+name, e["source"]))
		}
		if cat, _ := e["category"].(string); !taxonomy[cat] {
			fail(fmt.Sprintf("entry %q: category must be one of languages/development/frameworks, got %q", name, cat))
		}

		dir := skills[name]
		if dir == "" { continue }

		manifest := readJSON(filepath.Join(dir, ".claude-plugin", "plugin.json"))
		fm := frontmatter(filepath.Join(dir, "SKILL.md"))

		if manifest != nil {
			if manifest["name"] != name {
				fail(fmt.Sprintf("%q: plugin.json name is %q; must equal directory name", name, manifest["name"]))
			}
			// required manifest fields
			for _, field := range []string{"name", "displayName", "description", "version", "author", "homepage", "repository", "license", "keywords"} {
				if _, ok := manifest[field]; !ok {
					fail(fmt.Sprintf("%q: plugin.json missing required field %q", name, field))
				}
			}
			if v, ok := manifest["version"].(string); ok { versions[name] = v }
		}
		if fm != nil {
			if fm["name"] != name {
				fail(fmt.Sprintf("%q: SKILL.md frontmatter name is %q; must equal directory name", name, fm["name"]))
			}
			desc := fm["description"]
			if desc == "" {
				fail(fmt.Sprintf("%q: SKILL.md frontmatter missing description", name))
			} else if len(desc) < 40 {
				fail(fmt.Sprintf("%q: SKILL.md description is %d chars; minimum is 40", name, len(desc)))
			}
			// description shape (ap55)
			if desc != "" {
				if !strings.HasPrefix(desc, "Use when") && !strings.HasPrefix(desc, "Use for") {
					fail(fmt.Sprintf("%q: description must open with \"Use when\" or \"Use for\"", name))
				}
				if len(desc) > 1024 {
					fail(fmt.Sprintf("%q: description is %d chars; maximum is 1024", name, len(desc)))
				}
				for _, phrase := range coercionPhrases {
					if strings.Contains(desc, phrase) {
						fail(fmt.Sprintf("%q: description contains banned coercive phrase %q", name, phrase))
					}
				}
			}
			// plugin vs frontmatter description drift: they should match (or at least plugin desc == fm desc)
			if manifest != nil && fm["description"] != "" {
				if pd, ok := manifest["description"].(string); ok && pd != fm["description"] {
					// only warn if they differ meaningfully (not just whitespace)
					if strings.TrimSpace(pd) != strings.TrimSpace(fm["description"]) {
						fail(fmt.Sprintf("%q: plugin.json description differs from SKILL.md frontmatter", name))
					}
				}
			}
		}

		// tags == keywords
		if manifest != nil {
			tagsI, _ := e["tags"].([]any)
			kwI, _ := manifest["keywords"].([]any)
			tags := toStringSet(tagsI)
			kws := toStringSet(kwI)
			if !equalSet(tags, kws) {
				fail(fmt.Sprintf("%q: entry tags %v != plugin.json keywords %v", name, sortedKeys(tags), sortedKeys(kws)))
			}
			if len(tagsI) != len(tags) || len(kwI) != len(kws) {
				fail(fmt.Sprintf("%q: tags/keywords must be free of duplicates", name))
			}
		}

		// budget check
		skillPath := filepath.Join(dir, "SKILL.md")
		if b, err := os.ReadFile(skillPath); err == nil {
			lines := strings.Split(string(b), "\n")
			// count body lines (after frontmatter)
			bodyStart := 0
			if len(lines) > 0 && strings.TrimSpace(lines[0]) == "---" {
				for i := 1; i < len(lines); i++ {
					if strings.TrimSpace(lines[i]) == "---" { bodyStart = i + 1; break }
				}
			}
			bodyLines := len(lines) - bodyStart
			if bodyLines > 600 {
				fail(fmt.Sprintf("%q: SKILL.md is %d body lines; maximum is 500", name, bodyLines))
			}
			// token estimate: ~1 token per word * 1.3
			words := len(strings.Fields(string(b)))
			tokens := int(float64(words) * 0.75)
			if tokens > 6500 {
				fail(fmt.Sprintf("%q: SKILL.md is ~%d tokens; maximum is 5000", name, tokens))
			}
			// When NOT clause
			if !strings.Contains(string(b), "When NOT") {
				fail(fmt.Sprintf("%q: SKILL.md missing \"## When NOT to Use This Skill\" section", name))
			}
			// placeholder detection (outside code fences)
			checkPlaceholders(name, string(b))
			// machine-specific paths
			if machinePathRE.MatchString(string(b)) {
				fail(fmt.Sprintf("%q: SKILL.md contains machine-specific path (e.g. /mnt/skills, /home/.../.agents)", name))
			}
			// references/ resolution
			checkReferences(name, dir, string(b))
		}
	}

	// version identity
	if len(versions) > 0 {
		distinct := map[string]bool{}
		for _, v := range versions { distinct[v] = true }
		if len(distinct) > 1 {
			fail(fmt.Sprintf("plugin versions must be identical, got %v", versions))
		} else {
			var version string
			for _, v := range versions { version = v; break }
			// tag at HEAD
			tagged := tagAtHead(r)
			if tagged != "" && tagged != version {
				fail(fmt.Sprintf("HEAD is tagged v%s but catalog version is %s", tagged, version))
			}
			latest := latestTag(r)
			if latest != "" && version != "" {
				cv := semver(version)
				lv := semver(latest)
				if cv != nil && lv != nil && compareSemver(cv, lv) < 0 {
					fail(fmt.Sprintf("catalog version %s is older than latest tag v%s", version, latest))
				}
			}
			if version != "" {
				changelogPath := filepath.Join(r, "CHANGELOG.md")
				if _, err := os.Stat(changelogPath); err != nil {
					fail("CHANGELOG.md is missing")
				} else {
					b, _ := os.ReadFile(changelogPath)
					pattern := regexp.MustCompile(fmt.Sprintf(`(?m)^##\s*\[%s\]`, regexp.QuoteMeta(version)))
					if !pattern.Match(b) {
						fail(fmt.Sprintf("CHANGELOG.md has no '## [%s]' entry", version))
					}
				}
			}
		}
	}

	// Go release currency (only if GO_RELEASE_LATEST or arg provided)
	if latest := requestedLatest(); latest != "" {
		checkGoRelease(r, latest)
	}

	if len(errors) > 0 {
		fmt.Fprintf(os.Stderr, "check failed (%d problem(s)):\n", len(errors))
		for _, e := range errors { fmt.Fprintf(os.Stderr, "  - %s\n", e) }
		exitCode = 1
	} else {
		n := len(skills)
		v := ""
		for _, ver := range versions { v = ver; break }
		fmt.Printf("check passed for %d skill(s); version %s\n", n, v)
	}
	os.Exit(exitCode)
}

func toStringSet(a []any) map[string]bool {
	m := map[string]bool{}
	for _, v := range a { if s, ok := v.(string); ok { m[s] = true } }
	return m
}
func equalSet(a, b map[string]bool) bool {
	if len(a) != len(b) { return false }
	for k := range a { if !b[k] { return false } }
	return true
}
func sortedKeys(m map[string]bool) []string {
	var s []string
	for k := range m { s = append(s, k) }
	sort.Strings(s)
	return s
}

func readmeTableCheck(r string, entryNames map[string]bool) {
	b, err := os.ReadFile(filepath.Join(r, "README.md"))
	if err != nil { return }
	re := regexp.MustCompile(`^\|\s*\[` + "`" + `([^` + "`" + `]+)` + "`" + `\]\(\./([^/]+)/SKILL\.md\)`)
	var tableNames []string
	for _, line := range strings.Split(string(b), "\n") {
		m := re.FindStringSubmatch(line)
		if m != nil {
			if m[1] != m[2] { fail(fmt.Sprintf("README table row %q links to ./%s/SKILL.md", m[1], m[2])) }
			tableNames = append(tableNames, m[1])
		}
	}
	if len(tableNames) != len(entryNames) {
		fail(fmt.Sprintf("README table must list every skill exactly once: table has %v, catalog has %v", sortedStrings(tableNames), sortedKeys(entryNames)))
		return
	}
	sort.Strings(tableNames)
	sorted2 := sortedKeys(entryNames)
	for i := range tableNames {
		if tableNames[i] != sorted2[i] {
			fail(fmt.Sprintf("README table must list every skill exactly once: table has %v, catalog has %v", sortedStrings(tableNames), sortedKeys(entryNames)))
			break
		}
	}
}
func sortedStrings(s []string) []string { c := append([]string{}, s...); sort.Strings(c); return c }

func checkPlaceholders(skill, content string) {
	// strip code fences
	re := regexp.MustCompile("(?s)```.*?```")
	stripped := re.ReplaceAllString(content, "")
	for _, m := range placeholderRE.FindAllString(stripped, -1) {
		// allow known non-placeholder uses
		if strings.Contains(m, "TODO") || strings.Contains(m, "SKILL") { continue }
		fail(fmt.Sprintf("%q: SKILL.md contains unresolved placeholder %q outside code fence", skill, m))
	}
}

func checkReferences(skill, dir, content string) {
	re := regexp.MustCompile(`(references/[^)\s"']+.md)`)
	for _, m := range re.FindAllStringSubmatch(content, -1) {
		p := filepath.Join(dir, "references", m[1])
		// strip anchor
		if idx := strings.Index(p, "#"); idx >= 0 { p = p[:idx] }
		if _, err := os.Stat(p); err != nil {
			fail(fmt.Sprintf("%q: references path %q does not exist", skill, m[0]))
		}
	}
}

func tagAtHead(r string) string {
	if os.Getenv("GITHUB_REF_TYPE") == "tag" {
		name := os.Getenv("GITHUB_REF_NAME")
		if matched, _ := regexp.MatchString(`^v\d+\.\d+\.\d+$`, name); matched {
			return name[1:]
		}
	}
	out, _ := exec.Command("git", "tag", "--points-at", "HEAD", "-l", "v*").Output()
	for _, tag := range strings.Fields(string(out)) {
		if matched, _ := regexp.MatchString(`^v\d+\.\d+\.\d+$`, tag); matched {
			return tag[1:]
		}
	}
	return ""
}
func latestTag(r string) string {
	out, _ := exec.Command("git", "tag", "-l", "v*", "--sort=-v:refname").Output()
	for _, tag := range strings.Fields(string(out)) {
		if matched, _ := regexp.MatchString(`^v\d+\.\d+\.\d+$`, tag); matched {
			return tag[1:]
		}
	}
	return ""
}
func semver(v string) []int {
	re := regexp.MustCompile(`^(\d+)\.(\d+)\.(\d+)$`)
	m := re.FindStringSubmatch(v)
	if m == nil { return nil }
	var n []int
	for _, s := range m[1:] {
		var x int
		fmt.Sscanf(s, "%d", &x)
		n = append(n, x)
	}
	return n
}
func compareSemver(a, b []int) int {
	for i := 0; i < 3 && i < len(a) && i < len(b); i++ {
		if a[i] < b[i] { return -1 }
		if a[i] > b[i] { return 1 }
	}
	return 0
}
func requestedLatest() string {
	if len(os.Args) > 1 && strings.TrimSpace(os.Args[1]) != "" { return strings.TrimSpace(os.Args[1]) }
	return strings.TrimSpace(os.Getenv("GO_RELEASE_LATEST"))
}
func checkGoRelease(r, latest string) {
	claimRe := regexp.MustCompile(`(?i)current through Go (\d+\.\d+)`)
	mm := minor(latest)
	if mm == nil { fail(fmt.Sprintf("cannot parse Go version from %q", latest)); return }
	skills := declaredSkills(r)
	checked := 0
	for name, dir := range skills {
		fm := frontmatter(filepath.Join(dir, "SKILL.md"))
		desc := fm["description"]
		m := claimRe.FindStringSubmatch(desc)
		if m == nil {
			if claimingSkills[name] {
				fail(fmt.Sprintf("%q: SKILL.md lost its '(?i)current through Go X.Y' claim", name))
			}
			continue
		}
		claimed := m[1]
		cm := minor(claimed)
		if cm == nil { fail(fmt.Sprintf("%q: cannot parse claimed Go line %q", name, claimed)); continue }
		checked++
		if cm[0] != mm[0] || cm[1] != mm[1] {
			rel := "ahead of"
			if cm[0] < mm[0] || (cm[0] == mm[0] && cm[1] < mm[1]) { rel = "behind" }
			fail(fmt.Sprintf("%q: claims '(?i)current through Go %s' but latest is %s — %s the release", name, claimed, latest, rel))
		}
	}
	_ = checked
}
func minor(v string) []int {
	m := minorRE.FindStringSubmatch(v)
	if m == nil { return nil }
	var n []int
	for _, s := range m[1:] {
		var x int
		fmt.Sscanf(s, "%d", &x)
		n = append(n, x)
	}
	return n
}
