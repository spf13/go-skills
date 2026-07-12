# Idiomatic Go: The Pragmatic Playbook

A definitive guide to building robust, efficient, and maintainable Go applications—without the enterprise baggage.

## Another Go project layout guide? No thanks.

As Go has grown in popularity, developers are importing structural baggage from other languages—specifically Java and Spring Boot—and rebranding them as "best practices."

If you search for "Go project layout," you will inevitably be pointed to repositories like the notorious `golang-standards/project-layout`. These templates promote deeply nested directories, arbitrary architectural layers (`service/`, `repository/`, `pkg/`), heavy worker pools, and complex mocking frameworks.

**Those are not Go patterns. They are Java patterns translated into Go syntax.** They actively fight against Go's design. Go was engineered for simplicity, readability, and flat, discoverable APIs. When you force a layered, Spring-Boot-style architecture onto Go, you introduce circular dependencies, obscure the control flow, and destroy the beauty of the language its breathtaking simplicity.

This problem has gotten significantly worse with LLMs. AI coding assistants have been trained on the full corpus of the internet—including all of those misguided "Go best practices" guides. The result is that they confidently generate Java-in-Go-syntax by default, then argue with you when you push back. These skills exist to correct that. They give the LLM authoritative, first-principles guidance so it stops reaching for `internal/` junk drawers, BDD frameworks, and static worker pools when writing your Go code.

## A Course Correction

As the creator of tools like Hugo, Cobra, and Viper, and having spent years on the core Go team, I feel I have a good sense of what a well structured Go codebase should look like and I'm tired of arguing with LLMs trained on Java style codebases. This playbook is to save my sanity.

It strips away the noise and focuses on the actual architectures used by the Go standard library and the most successful, high-performance open-source projects in the world.

**This playbook advocates for:**

- **Domains over Layers:** Deleting the `internal/` junk drawer and the `pkg/` anti-pattern. Organize code by what it *does*, not by what *kind of file* it is.
- **Standard Library over Frameworks:** Using `testing`, table-driven tests, and simple stubs instead of heavy BDD or mock-generation frameworks.
- **Channels over Mutexes:** Using Go's native concurrency primitives for orchestration rather than rigid, static worker pools.
- **Command-First Architecture:** Treating your application binary as a router for commands, entirely decoupled from your core business logic.

## The Playbooks

### 1. [General Go Patterns](./go/SKILL.md)

The core language idioms. This covers fundamental Go philosophy, proper interface design, standard-library-compliant testing, and concurrency patterns. **Stop writing Java in Go.** Read this first.

### 2. [CLI Architecture: Cobra & Viper](./cobra-viper/SKILL.md)

The definitive guide to building modern command-line applications. This covers the "Command-First" architecture, 12-factor configuration, and how to structure your application using Cobra and Viper exactly as they were intended to be used.

### 3. [Spec Review](./go-spec-reviewer/SKILL.md)

Review a design document *before* implementation begins. Channels Rob Pike, the stdlib authors, and spf13 to catch over-engineering, missing error paths, interface misuse, and Cobra/Viper convention violations while a plan is still cheap to change.

### 4. [Release Engineering](./go-release/SKILL.md)

How to cut a release without breaking your users. Semantic versioning promises, mechanically detecting breaking changes, `Deprecated:` conventions, go.mod hygiene, and shipping binaries with GoReleaser.

### 5. [Desktop Apps: Wails](./wails/SKILL.md)

Building native desktop applications in Go with Wails. Covers both v2 (stable) and v3 (alpha) — and, critically, how to tell them apart, since their APIs are incompatible and easily blended by mistake.

### 6. [File Operations: fileflow & pathologize](./fileflow-pathologize/SKILL.md)

Safe cross-filesystem file operations using `spf13/fileflow` (move, copy, rename with conflict-safe naming) and `spf13/pathologize` (sanitize filenames and path segments for every OS). Use whenever Go code moves, copies, renames, downloads, or extracts files, or builds paths from untrusted input.

## Installation

### Claude Code

```text
/plugin marketplace add spf13/go-skills
/plugin install go@go-skills
/plugin install cobra-viper@go-skills
/plugin install go-spec-reviewer@go-skills
/plugin install go-release@go-skills
/plugin install wails@go-skills
/plugin install fileflow-pathologize@go-skills
```

### Codex

```bash
codex plugin marketplace add spf13/go-skills
codex plugin add go-skills@go-skills
```

Start a new Codex thread after installation. The bundle provides `$go-skills:go`, `$go-skills:cobra-viper`, `$go-skills:go-spec-reviewer`, `$go-skills:go-release`, `$go-skills:wails`, and `$go-skills:fileflow-pathologize`.

Codex caches plugins by version. When changing any bundled skill, also bump `version` in [`.codex-plugin/plugin.json`](./.codex-plugin/plugin.json) so installed copies can be updated.

### Other AI Agents (Copilot, Cursor, etc.)

Place the skills where your AI coding agent can find them:

```powershell
# Windows — directory junction (run as Administrator)
New-Item -ItemType Junction -Path "$env:USERPROFILE\.agents\skills\go" -Target "$PWD\go"
New-Item -ItemType Junction -Path "$env:USERPROFILE\.agents\skills\cobra-viper" -Target "$PWD\cobra-viper"
New-Item -ItemType Junction -Path "$env:USERPROFILE\.agents\skills\go-spec-reviewer" -Target "$PWD\go-spec-reviewer"
New-Item -ItemType Junction -Path "$env:USERPROFILE\.agents\skills\go-release" -Target "$PWD\go-release"
New-Item -ItemType Junction -Path "$env:USERPROFILE\.agents\skills\wails" -Target "$PWD\wails"
New-Item -ItemType Junction -Path "$env:USERPROFILE\.agents\skills\fileflow-pathologize" -Target "$PWD\fileflow-pathologize"
```

```bash
# macOS / Linux — symlink
ln -s "$PWD/go"                        "$HOME/.agents/skills/go"
ln -s "$PWD/cobra-viper"               "$HOME/.agents/skills/cobra-viper"
ln -s "$PWD/go-spec-reviewer"          "$HOME/.agents/skills/go-spec-reviewer"
ln -s "$PWD/go-release"                "$HOME/.agents/skills/go-release"
ln -s "$PWD/wails"                     "$HOME/.agents/skills/wails"
ln -s "$PWD/fileflow-pathologize"      "$HOME/.agents/skills/fileflow-pathologize"
```

After linking, restart VS Code. The skills will appear in the Copilot customizations index and be invoked automatically when relevant Go or CLI work is detected.

## The Golden Rule

*Clear is better than clever.* Go code should be boring in the best possible way—predictable, consistent, and immediately understandable to a new developer opening the file for the first time. When in doubt, delete the abstraction.

## License

MIT
