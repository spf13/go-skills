package check

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// maxSkillLines is the progressive-disclosure budget for a SKILL.md: the whole
// file has to fit in one context load, and anything longer belongs in
// `references/`, loaded only when the task needs it.
const maxSkillLines = 500

// overBudget records skills whose SKILL.md predates that budget. Each entry caps
// the file at the length it had when the cap was recorded: the file may shrink
// but never grow, and the entry goes away with the change that moves the excess
// into `references/`. A skill absent from this table must satisfy the budget.
var overBudget = map[string]int{
	"cobra-viper":          609,
	"fileflow-pathologize": 527,
}

// defersSection is the heading that resolves overlapping triggers without a
// priority table.
const defersSection = "## When NOT to Use This Skill"

// placeholderRE matches an unresolved fill-in token such as [PLACEHOLDER],
// [SPEC_FILE_PATH], [TODO], or [TBD]: bracketed capitals a model reads as
// content instead of as an instruction.
var placeholderRE = regexp.MustCompile(`\[[A-Z][A-Z0-9_ -]{2,}\]`)

var referenceLinkRE = regexp.MustCompile(`(references/[^)\s"']+\.md)`)

var fencedBlockRE = regexp.MustCompile("(?s)```.*?```")

// skillContent runs the SKILL.md rules that are about prose rather than catalog
// metadata.
func (c *checker) skillContent() error {
	if err := c.ensureSkills(); err != nil {
		return err
	}
	for _, s := range c.skills {
		path := filepath.Join(s.dir, "SKILL.md")
		raw, ok := c.read(path)
		if !ok {
			continue
		}
		src := string(raw)
		rel := c.rel(path)

		c.checkBudget(rel, s.name, src)
		if !strings.Contains(src, defersSection) {
			c.fail(rel, 0, "%q: missing %q section naming the sibling skills it defers to", s.name, defersSection)
		}
		c.checkPlaceholders(rel, s.name, src)
		c.checkReferences(rel, s.name, s.dir, src)
	}
	return nil
}

// checkBudget enforces the line budget, allowing a recorded cap for a skill that
// predates it.
func (c *checker) checkBudget(path, name, src string) {
	lines := strings.Count(src, "\n")
	if !strings.HasSuffix(src, "\n") {
		lines++
	}
	if lines <= maxSkillLines {
		return
	}
	if cap, recorded := overBudget[name]; recorded {
		if lines > cap {
			c.fail(path, 0, "%q: SKILL.md is %d lines, above its recorded cap of %d; it may only shrink until it fits the %d-line budget", name, lines, cap, maxSkillLines)
		}
		return
	}
	c.fail(path, 0, "%q: SKILL.md is %d lines; maximum is %d — move detail into references/", name, lines, maxSkillLines)
}

// checkPlaceholders flags fill-in tokens a reader could mistake for content.
// Fenced code is skipped: a sample may legitimately show bracketed capitals.
func (c *checker) checkPlaceholders(path, name, src string) {
	prose := fencedBlockRE.ReplaceAllString(src, "")
	for i, line := range strings.Split(prose, "\n") {
		for _, token := range placeholderRE.FindAllString(line, -1) {
			c.fail(path, i+1, "%q: unresolved placeholder %q", name, token)
		}
	}
}

// checkReferences asserts every `references/…md` link resolves to a file.
func (c *checker) checkReferences(path, name, dir, src string) {
	for _, match := range referenceLinkRE.FindAllStringSubmatch(src, -1) {
		link := match[1]
		if idx := strings.Index(link, "#"); idx >= 0 {
			link = link[:idx]
		}
		if _, err := os.Stat(filepath.Join(dir, filepath.FromSlash(link))); err != nil {
			c.fail(path, 0, "%q: reference %q does not exist", name, match[1])
		}
	}
}
