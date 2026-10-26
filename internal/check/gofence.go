package check

import (
	"go/parser"
	"go/token"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

// goBlock is one fenced Go sample lifted out of a Markdown file.
type goBlock struct {
	file string // repository-relative path
	line int    // 1-based line of the opening fence
	src  string
}

// packageClauseRE marks a self-contained sample. A fence carrying its own
// package clause can be compiled and vetted as written. Fences without one are
// deliberately partial — a statement, a signature, an excerpt — and are not
// compiled, because wrapping them in a synthetic func would invent context the
// author never wrote and report failures that are not real.
var packageClauseRE = regexp.MustCompile(`(?m)^package\s+[A-Za-z_]\w*\s*$`)

// compileEnv keeps snippet builds offline and independent of any workspace.
var compileEnv = []string{"GOWORK=off", "GOFLAGS=-mod=mod", "GO111MODULE=on"}

// goFences compiles and vets every complete Go sample in the shipped skills.
func (c *checker) goFences() error {
	if err := c.ensureSkills(); err != nil {
		return err
	}
	for _, s := range c.skills {
		for _, path := range markdownFiles(s.dir) {
			raw, err := os.ReadFile(path)
			if err != nil {
				continue // reported by skillContent when it matters
			}
			rel := c.rel(path)
			for _, block := range goBlocks(rel, string(raw)) {
				if !packageClauseRE.MatchString(block.src) {
					continue // partial sample: a statement or an excerpt
				}
				if !stdlibOnly(block.src) {
					// The check has no network and no module to resolve against,
					// so a sample importing any other module cannot be built here.
					c.skippedSamples++
					continue
				}
				if problem := compileSnippet(block.src); problem != "" {
					c.fail(block.file, block.line, "complete Go sample does not compile: %s", problem)
					continue
				}
				c.compiledSamples++
			}
		}
	}
	return nil
}

// markdownFiles lists SKILL.md plus every Markdown file in references/.
func markdownFiles(dir string) []string {
	refs, err := filepath.Glob(filepath.Join(dir, "references", "*.md"))
	if err != nil {
		refs = nil
	}
	return append([]string{filepath.Join(dir, "SKILL.md")}, refs...)
}

// goBlocks extracts every ```go fence, remembering where it started.
func goBlocks(file, src string) []goBlock {
	var blocks []goBlock
	var current goBlock
	inside := false
	for i, line := range strings.Split(src, "\n") {
		trimmed := strings.TrimSpace(line)
		if !inside {
			if strings.HasPrefix(trimmed, "```") && strings.TrimSpace(strings.TrimPrefix(trimmed, "```")) == "go" {
				inside = true
				current = goBlock{file: file, line: i + 1}
			}
			continue
		}
		if strings.HasPrefix(trimmed, "```") {
			blocks = append(blocks, current)
			inside = false
			continue
		}
		current.src += line + "\n"
	}
	return blocks
}

// stdlibOnly reports whether a self-contained sample imports nothing but the
// standard library. Every stdlib path lacks a dot in its first element, so the
// test needs no module graph and no network.
func stdlibOnly(src string) bool {
	file, err := parser.ParseFile(token.NewFileSet(), "snippet.go", src, parser.SkipObjectResolution)
	if err != nil {
		return true // malformed: let the compile step report it
	}
	for _, spec := range file.Imports {
		path, err := strconv.Unquote(spec.Path.Value)
		if err != nil {
			return false
		}
		first, _, _ := strings.Cut(path, "/")
		if strings.Contains(first, ".") {
			return false
		}
	}
	return true
}

// compileSnippet builds and vets one self-contained sample in a throwaway
// module. It returns a one-line description of the first failure, or "".
func compileSnippet(src string) string {
	dir, err := os.MkdirTemp("", "go-skills-snippet")
	if err != nil {
		return err.Error()
	}
	defer os.RemoveAll(dir)

	if err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module snippet\n\ngo 1.22\n"), 0o644); err != nil {
		return err.Error()
	}
	if err := os.WriteFile(filepath.Join(dir, "snippet.go"), []byte(src), 0o644); err != nil {
		return err.Error()
	}
	for _, args := range [][]string{{"build", "./..."}, {"vet", "./..."}} {
		cmd := exec.Command("go", args...)
		cmd.Dir = dir
		cmd.Env = append(os.Environ(), compileEnv...)
		out, err := cmd.CombinedOutput()
		if err != nil {
			return firstFailure(string(out), dir)
		}
	}
	return ""
}

// firstFailure reduces toolchain output to one readable line, with the throwaway
// directory stripped so the message points at the sample.
func firstFailure(output, dir string) string {
	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimSpace(strings.ReplaceAll(line, dir+string(filepath.Separator), ""))
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		return line
	}
	return "toolchain reported failure with no output"
}
