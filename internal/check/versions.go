package check

import (
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"
)

var (
	releaseTagRE = regexp.MustCompile(`^v(\d+)\.(\d+)\.(\d+)$`)
	semverRE     = regexp.MustCompile(`^(\d+)\.(\d+)\.(\d+)$`)
)

// catalogVersions asserts the catalog moves as one version, agrees with the tag
// at HEAD, never falls behind the newest tag, and has a CHANGELOG entry.
func (c *checker) catalogVersions() error {
	if len(c.pluginVersions) == 0 {
		return nil
	}
	distinct := map[string]bool{}
	for _, version := range c.pluginVersions {
		distinct[version] = true
	}
	if len(distinct) > 1 {
		c.fail(".claude-plugin/marketplace.json", 0, "plugin versions must be identical across the catalog, got %v", c.pluginVersions)
		return nil
	}
	version := c.version()

	if tagged := tagAtHead(c.root); tagged != "" && tagged != version {
		c.fail(".claude-plugin/marketplace.json", 0, "HEAD is tagged v%s but the catalog version is %s", tagged, version)
	}
	if newest := newestTag(c.root); newest != "" {
		current, okCurrent := semver(version)
		latest, okLatest := semver(newest)
		if okCurrent && okLatest && compareSemver(current, latest) < 0 {
			c.fail(".claude-plugin/marketplace.json", 0, "catalog version %s is older than the latest tag v%s", version, newest)
		}
	}

	raw, err := os.ReadFile(c.root + "/CHANGELOG.md")
	if err != nil {
		c.fail("CHANGELOG.md", 0, "%v", err)
		return nil
	}
	entry := regexp.MustCompile(`(?m)^##\s*\[` + regexp.QuoteMeta(version) + `\]`)
	if !entry.Match(raw) {
		c.fail("CHANGELOG.md", 0, "no '## [%s]' entry", version)
	}
	return nil
}

// tagAtHead reports the semver of a v* tag pointing at HEAD, locally or in CI.
func tagAtHead(root string) string {
	if os.Getenv("GITHUB_REF_TYPE") == "tag" {
		if releaseTagRE.MatchString(os.Getenv("GITHUB_REF_NAME")) {
			return strings.TrimPrefix(os.Getenv("GITHUB_REF_NAME"), "v")
		}
	}
	return firstReleaseTag(root, "tag", "--points-at", "HEAD", "-l", "v*")
}

// newestTag reports the highest vX.Y.Z tag in the repository.
func newestTag(root string) string {
	return firstReleaseTag(root, "tag", "-l", "v*", "--sort=-v:refname")
}

// firstReleaseTag returns the first semver-looking v* tag git lists. A repository
// without tags, or without git, has no tagged release to compare against.
func firstReleaseTag(root string, args ...string) string {
	cmd := exec.Command("git", args...)
	cmd.Dir = root
	out, err := cmd.Output()
	if err != nil {
		return ""
	}
	for _, tag := range strings.Fields(string(out)) {
		if releaseTagRE.MatchString(tag) {
			return strings.TrimPrefix(tag, "v")
		}
	}
	return ""
}

func semver(value string) ([3]int, bool) {
	match := semverRE.FindStringSubmatch(value)
	if match == nil {
		return [3]int{}, false
	}
	var parts [3]int
	for i := range parts {
		if _, err := fmt.Sscanf(match[i+1], "%d", &parts[i]); err != nil {
			return [3]int{}, false
		}
	}
	return parts, true
}

func compareSemver(a, b [3]int) int {
	for i := range a {
		switch {
		case a[i] < b[i]:
			return -1
		case a[i] > b[i]:
			return 1
		}
	}
	return 0
}
