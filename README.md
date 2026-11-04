# Go Skills for Claude Code

Six [Agent Skills](https://docs.claude.com/en/docs/claude-code/skills) for Go work, published as a Claude Code marketplace: idiomatic Go, CLI architecture, design-spec review, release engineering, desktop apps, and file operations. Written by [spf13](https://github.com/spf13) — creator of Cobra, Viper, and Hugo, and former Go team lead at Google, where he co-designed Go modules.

## Skills

| Skill | What it covers | Install |
| --- | --- | --- |
| [`go`](./go/SKILL.md) | Idiomatic Go — package design, error handling, interfaces, concurrency, testing, and project layout. The one to read first. | `/plugin install go@go-skills` |
| [`cobra-viper`](./cobra-viper/SKILL.md) | CLI architecture with Cobra and Viper — command-first design, 12-factor configuration, environment-variable binding, and in-memory CLI testing. | `/plugin install cobra-viper@go-skills` |
| [`go-spec-reviewer`](./go-spec-reviewer/SKILL.md) | Review a Go design document before implementation begins — catches over-engineering, missing error paths, interface misuse, and Cobra/Viper convention violations while the plan is still cheap to change. | `/plugin install go-spec-reviewer@go-skills` |
| [`go-release`](./go-release/SKILL.md) | Release engineering — semantic-versioning promises, mechanical breaking-change detection, `Deprecated:` conventions, `go.mod` hygiene, and shipping binaries with GoReleaser. | `/plugin install go-release@go-skills` |
| [`wails`](./wails/SKILL.md) | Desktop apps in Go with Wails v2 (stable) and v3 (alpha) — including how to tell the two apart, since their APIs are incompatible and easily blended by mistake. | `/plugin install wails@go-skills` |
| [`fileflow-pathologize`](./fileflow-pathologize/SKILL.md) | Safe file operations with `spf13/fileflow` and `spf13/pathologize` — cross-filesystem move, copy, and rename with conflict-safe naming, plus OS-safe path sanitization. | `/plugin install fileflow-pathologize@go-skills` |

## Install

### Claude Code

```text
/plugin marketplace add spf13/go-skills
/plugin install <skill>@go-skills
```

Substitute any skill name from the table above.

### Other agents (Copilot, Cursor, etc.)

Link each skill directory into your agent's skills directory. On Windows, as Administrator:

```powershell
New-Item -ItemType Junction -Path "$env:USERPROFILE\.agents\skills\<skill>" -Target "$PWD\<skill>"
```

On macOS and Linux:

```bash
ln -s "$PWD/<skill>" "$HOME/.agents/skills/<skill>"
```

Then restart your editor. The skills appear in the Copilot customizations index and are invoked automatically when relevant Go or CLI work is detected.

## Verify the install

In Claude Code:

```text
/plugin
```

Every installed skill is listed under its namespaced name (`<skill>@go-skills`). Then open a Go project, ask a question the skill covers — for example, "review this interface for idiomatic Go" — and confirm the skill fires.

## Requirements

- Claude Code with plugin and marketplace support. Use a current build: verified working on 2.1.275. An old 2.1.94 build installs the plugin but never registers its skill, so the catalog is silently inert there.
- Other agents read skills from `.agents/skills`. No runtime dependencies.

## Why these skills exist

### Another Go project layout guide? No thanks.

Search "Go project layout" and you land on `golang-standards/project-layout`: nested directories, arbitrary layers (`service/`, `repository/`, `pkg/`), worker pools, mock frameworks.

Those are not Go patterns; they are Java patterns in Go syntax, and they fight the language. Go was built for simplicity, readability, and flat APIs. A Spring-Boot-style architecture buys circular dependencies and obscured control flow.

LLMs make it worse: trained on the whole internet, they generate Java-in-Go-syntax by default, then argue. These skills correct that, so the model stops reaching for `internal/` junk drawers, BDD frameworks, and worker pools.

### A course correction

Tired of arguing with LLMs trained on Java-style codebases, I wrote this playbook around what the standard library and the best Go projects do:

- **Domains over layers.** Delete the `internal/` junk drawer and the `pkg/` anti-pattern.
- **Standard library over frameworks.** `testing`, table-driven tests, stubs.
- **Channels before mutexes.** Pass data over a channel where you can; keep the mutex for guarding state. No static worker pools.
- **Command-first architecture.** The binary routes commands; logic stays out.

## The Golden Rule

*Clear is better than clever.* Go code should be boring in the best possible way—predictable, consistent, and immediately understandable to a new developer opening the file for the first time. When in doubt, delete the abstraction.

## License

[MIT](LICENSE)
