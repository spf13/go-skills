package check

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// minDescription is the shortest description a model can route on.
const minDescription = 40

// maxDescription is the routing ceiling from the Agent Skills spec.
const maxDescription = 1024

// coercivePhrases are invocation instructions that read as triggers but are not:
// they fire on prompts that have nothing to do with the skill.
var coercivePhrases = []string{
	"even if they don't ask",
	"even if they aren't named",
	"recommend them",
	"always use this skill",
	"recommend these libraries",
	"critical:",
	"you must use this skill",
}

var readmeRowRE = regexp.MustCompile("^\\|\\s*\\[`([^`]+)`\\]\\(\\./([^/]+)/SKILL\\.md\\)")

// manifests validates every JSON manifest against its vendored schemastore
// schema, then asserts the catalog invariants a schema cannot express.
func (c *checker) manifests() error {
	marketplaceSchema, err := loadSchema(marketplaceSchemaJSON, "marketplace")
	if err != nil {
		return err
	}
	pluginSchema, err := loadSchema(pluginSchemaJSON, "plugin manifest")
	if err != nil {
		return err
	}
	if err := c.ensureSkills(); err != nil {
		return err
	}
	c.pluginVersions = map[string]string{}

	marketplacePath := filepath.Join(c.root, ".claude-plugin", "marketplace.json")
	marketplace, ok := c.validateManifest(marketplacePath, marketplaceSchema, "marketplace")
	if !ok {
		return nil
	}
	if c.catalog, err = pluginsOf(marketplace); err != nil {
		c.fail(c.rel(marketplacePath), 0, "%v", err)
		return nil
	}

	for _, entry := range c.catalog {
		name, _ := entry["name"].(string)
		if name == "" {
			c.fail(c.rel(marketplacePath), 0, "entry without a name")
			continue
		}
		if keys := keysOf(entry); !equalStrings(keys, entryKeys) {
			c.fail(c.rel(marketplacePath), 0, "entry %q: keys must be exactly %v, got %v", name, entryKeys, keys)
		}
		if source, _ := entry["source"].(string); source != "./"+name {
			c.fail(c.rel(marketplacePath), 0, "entry %q: source must be %q, got %q", name, "./"+name, source)
		}
		if category, _ := entry["category"].(string); !taxonomy[category] {
			c.fail(c.rel(marketplacePath), 0, "entry %q: category must be languages, development, or frameworks, got %q", name, category)
		}
	}

	c.checkCorrespondence(c.rel(marketplacePath))
	c.checkReadmeTable()

	for _, entry := range c.catalog {
		name, _ := entry["name"].(string)
		if dir := c.skillDir(name); dir != "" {
			c.checkSkillManifest(dir, name, entry, pluginSchema)
		}
	}
	return nil
}

func (c *checker) skillDir(name string) string {
	for _, s := range c.skills {
		if s.name == name {
			return s.dir
		}
	}
	return ""
}

// validateManifest decodes path, validates it against schema, and reports every
// violation. It returns the document and whether it was usable.
func (c *checker) validateManifest(path string, schema map[string]any, label string) (map[string]any, bool) {
	raw, err := os.ReadFile(path)
	if err != nil {
		c.fail(c.rel(path), 0, "%v", err)
		return nil, false
	}
	var doc map[string]any
	if err := json.Unmarshal(raw, &doc); err != nil {
		c.fail(c.rel(path), 0, "invalid JSON: %v", err)
		return nil, false
	}
	for _, problem := range validateDocument(schema, doc) {
		c.fail(c.rel(path), 0, "[%s schema] %s", label, problem)
	}
	return doc, true
}

func pluginsOf(marketplace map[string]any) ([]map[string]any, error) {
	raw, present := marketplace["plugins"]
	if !present {
		return nil, fmt.Errorf("marketplace manifest has no plugins list")
	}
	list, ok := raw.([]any)
	if !ok {
		return nil, fmt.Errorf("marketplace manifest plugins must be a list")
	}
	entries := make([]map[string]any, 0, len(list))
	for i, item := range list {
		entry, ok := item.(map[string]any)
		if !ok {
			return nil, fmt.Errorf("plugins/%d must be an object", i)
		}
		entries = append(entries, entry)
	}
	return entries, nil
}

// checkCorrespondence asserts the catalog and the skill directories are in
// one-to-one correspondence, with no duplicate entry names.
func (c *checker) checkCorrespondence(path string) {
	declared := map[string]bool{}
	for _, entry := range c.catalog {
		name, _ := entry["name"].(string)
		if declared[name] {
			c.fail(path, 0, "duplicate entry name %q", name)
		}
		declared[name] = true
	}
	present := map[string]bool{}
	for _, s := range c.skills {
		present[s.name] = true
	}
	for _, name := range sortedNames(declared) {
		if !present[name] {
			c.fail(path, 0, "entry %q has no matching skill directory", name)
		}
	}
	for _, name := range sortedNames(present) {
		if !declared[name] {
			c.fail(path, 0, "skill directory %q has no matching marketplace entry", name)
		}
	}
}

// checkReadmeTable asserts the README skills table lists every catalog skill
// exactly once and links each row to its own directory.
func (c *checker) checkReadmeTable() {
	raw, ok := c.read(filepath.Join(c.root, "README.md"))
	if !ok {
		return
	}
	var table []string
	for i, line := range strings.Split(string(raw), "\n") {
		match := readmeRowRE.FindStringSubmatch(line)
		if match == nil {
			continue
		}
		if match[1] != match[2] {
			c.fail("README.md", i+1, "row %q links to ./%s/SKILL.md", match[1], match[2])
		}
		table = append(table, match[1])
	}
	var catalog []string
	for _, entry := range c.catalog {
		name, _ := entry["name"].(string)
		catalog = append(catalog, name)
	}
	if !equalSets(table, catalog) || hasDuplicates(table) {
		c.fail("README.md", 0, "skills table must list every skill exactly once: table has %v, catalog has %v", sortStrings(table), sortStrings(catalog))
	}
}

// checkSkillManifest asserts per-skill manifest and frontmatter agreement.
func (c *checker) checkSkillManifest(dir, name string, entry, pluginSchema map[string]any) {
	manifestPath := filepath.Join(dir, ".claude-plugin", "plugin.json")
	manifest, ok := c.validateManifest(manifestPath, pluginSchema, "plugin")
	if !ok {
		return
	}
	if got, _ := manifest["name"].(string); got != name {
		c.fail(c.rel(manifestPath), 0, "name is %q; must equal the directory name %q", got, name)
	}
	if version, _ := manifest["version"].(string); version != "" {
		c.pluginVersions[name] = version
	}

	skillPath := filepath.Join(dir, "SKILL.md")
	raw, ok := c.read(skillPath)
	if !ok {
		return
	}
	fm, ok := parseFrontmatter(string(raw))
	if !ok {
		c.fail(c.rel(skillPath), 0, "missing YAML frontmatter")
		return
	}
	if got := fm.Fields["name"]; got != name {
		c.fail(c.rel(skillPath), 0, "frontmatter name is %q; must equal the directory name %q", got, name)
	}
	description := fm.Fields["description"]
	c.checkDescription(c.rel(skillPath), name, description)
	if pluginDescription, _ := manifest["description"].(string); pluginDescription != description {
		c.fail(c.rel(skillPath), 0, "plugin.json description differs from the SKILL.md frontmatter description")
	}

	tags, keywords := stringsOf(entry["tags"]), stringsOf(manifest["keywords"])
	if !equalSets(tags, keywords) {
		c.fail(c.rel(manifestPath), 0, "marketplace tags %v != plugin.json keywords %v", sortStrings(tags), sortStrings(keywords))
	} else if hasDuplicates(tags) || hasDuplicates(keywords) {
		c.fail(c.rel(manifestPath), 0, "tags and keywords must each be free of duplicates")
	}
}

// checkDescription enforces the machine-checkable half of the description
// standard in CONTRIBUTING.md: length, trigger-only opening, routing ceiling,
// and the ban on coercive activation language. Wording stays a review call.
func (c *checker) checkDescription(path, name, description string) {
	switch {
	case description == "":
		c.fail(path, 0, "%q: frontmatter is missing a description", name)
		return
	case len(description) < minDescription:
		c.fail(path, 0, "%q: description is %d characters; minimum is %d", name, len(description), minDescription)
	case len(description) > maxDescription:
		c.fail(path, 0, "%q: description is %d characters; maximum is %d", name, len(description), maxDescription)
	}
	if !strings.HasPrefix(description, "Use when") && !strings.HasPrefix(description, "Use for") {
		c.fail(path, 0, "%q: description must open with \"Use when\" or \"Use for\"", name)
	}
	for _, phrase := range coercivePhrases {
		if strings.Contains(strings.ToLower(description), phrase) {
			c.fail(path, 0, "%q: description contains coercive activation phrase %q", name, phrase)
		}
	}
}

func stringsOf(raw any) []string {
	list, ok := raw.([]any)
	if !ok {
		return nil
	}
	out := make([]string, 0, len(list))
	for _, item := range list {
		if text, ok := item.(string); ok {
			out = append(out, text)
		}
	}
	return out
}
