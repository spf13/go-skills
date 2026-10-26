package check

import (
	"fmt"
	"regexp"
	"strings"
)

// claimRE extracts the release a skill says it covers, e.g. "current through Go
// 1.27", from its description prose.
var claimRE = regexp.MustCompile(`(?i)current through Go (\d+\.\d+)`)

var minorRE = regexp.MustCompile(`(\d+)\.(\d+)`)

// claimingSkills is the recorded baseline for "the claim must not silently
// disappear". Only `go` claims: it is the only skill that enumerates the whole
// language surface. A skill that later asserts release coverage adds itself here
// in the same change that adds the claim.
var claimingSkills = map[string]bool{"go": true}

// releaseClaims asserts that every "current through Go X.Y" claim names the
// latest stable release. The comparison needs a caller-supplied latest; the
// claim-must-not-disappear rule holds with or without one.
func (c *checker) releaseClaims() error {
	var latestMinor [2]int
	comparing := c.latest != ""
	if comparing {
		minor, ok := parseMinor(c.latest)
		if !ok {
			return fmt.Errorf("cannot parse a Go version from %q", c.latest)
		}
		latestMinor = minor
	}
	if err := c.ensureSkills(); err != nil {
		return err
	}

	for _, s := range c.skills {
		path := s.dir + "/SKILL.md"
		raw, ok := c.read(path)
		if !ok {
			continue
		}
		fm, ok := parseFrontmatter(string(raw))
		if !ok {
			continue
		}
		rel := c.rel(path)
		match := claimRE.FindStringSubmatch(fm.Fields["description"])
		if match == nil {
			if claimingSkills[s.name] {
				c.fail(rel, 0, "%q: description lost its 'current through Go X.Y' claim; it is a known claiming skill, so the claim must not disappear", s.name)
			}
			continue
		}
		if !comparing {
			continue
		}
		claimed, ok := parseMinor(match[1])
		if !ok {
			c.fail(rel, 0, "%q: cannot parse the claimed Go line from 'current through Go %s'", s.name, match[1])
			continue
		}
		if claimed != latestMinor {
			relation := "behind"
			if claimed[0] > latestMinor[0] || (claimed[0] == latestMinor[0] && claimed[1] > latestMinor[1]) {
				relation = "ahead of"
			}
			c.fail(rel, 0, "%q: claims 'current through Go %s' but the latest stable Go release is %s — the claim is %s the release", s.name, match[1], c.latest, relation)
		}
	}
	return nil
}

// parseMinor reads the release line from "go1.27.3", "1.27", or "1.27.3".
func parseMinor(version string) ([2]int, bool) {
	match := minorRE.FindStringSubmatch(strings.TrimSpace(version))
	if match == nil {
		return [2]int{}, false
	}
	var parts [2]int
	for i := range parts {
		if _, err := fmt.Sscanf(match[i+1], "%d", &parts[i]); err != nil {
			return [2]int{}, false
		}
	}
	return parts, true
}
