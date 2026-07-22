# CODEBUDDY.md

This file provides guidance to CodeBuddy Code when working with code in this repository.

## What this repository is

This is **not** a Go application. It is a **Claude Code plugin marketplace** that distributes six authoritative Go-related skills authored by Steve Francia (spf13) — creator of Cobra, Viper, Hugo. The stated mission (see `README.md`): counter the Java-in-Go-syntax defaults that LLMs reach for, and steer them toward idiomatic patterns used by the Go standard library and spf13's own projects.

There is no Go source code, no `go.mod`, no Makefile, no CI, and no test suite. Every change is a content change to a markdown skill or a JSON manifest.

## Repository layout

```
.
├── README.md                           # Top-level pitch + per-skill table of contents
├── CODEBUDDY.md                        # This file — guidance for CodeBuddy Code agents
├── .claude-plugin/
│   └── marketplace.json                # Claude Code marketplace manifest (all 6 plugins aggregated)
├── .codebuddy/
│   └── skills/                         # CodeBuddy project-skill discovery — symlinks below
│       ├── README.md
│       ├── go/            -> ../../go
│       ├── cobra-viper/   -> ../../cobra-viper
│       ├── go-spec-reviewer/ -> ../../go-spec-reviewer
│       ├── go-release/    -> ../../go-release
│       ├── wails/         -> ../../wails
│       └── fileflow-pathologize/ -> ../../fileflow-pathologize
├── go/                                 # Skill: idiomatic Go patterns (read first)
│   ├── SKILL.md
│   └── .claude-plugin/plugin.json
├── cobra-viper/                        # Skill: CLI architecture with Cobra & Viper
│   ├── SKILL.md
│   └── .claude-plugin/plugin.json
├── go-spec-reviewer/                   # Skill: review Go design docs pre-implementation
│   ├── SKILL.md
│   └── .claude-plugin/plugin.json
├── go-release/                         # Skill: release engineering (semver, GoReleaser)
│   ├── SKILL.md
│   └── .claude-plugin/plugin.json
├── wails/                              # Skill: Wails v2/v3 desktop apps
│   ├── SKILL.md
│   └── .claude-plugin/plugin.json
└── fileflow-pathologize/               # Skill: safe file ops (spf13/fileflow + pathologize)
    ├── SKILL.md
    └── .claude-plugin/plugin.json
```

Each skill directory at the repo root is **self-contained and independent** — sibling directories are not dependencies. The two parallel discovery surfaces are:

- **`.claude-plugin/marketplace.json`** — Claude Code catalog (one entry per skill).
- **`.codebuddy/skills/`** — CodeBuddy project-skill index, implemented as symlinks to the same root directories. Edit the root; the symlinks follow.

CodeBuddy's required format (per `https://www.codebuddy.cn/docs/ide/Features/Skills`) is just `SKILL.md` with YAML frontmatter (`name`, `description`); optional per-skill `scripts/`, `references/`, `assets/` should be added **next to the symlink target** at the repo root, not inside `.codebuddy/skills/`.

## File roles

- **`SKILL.md`** — the actual skill content. Begins with YAML frontmatter (`name`, `description`); the `description` is the activation trigger and must be specific enough to fire on the right requests. Body is markdown organized by `## When to Activate`, philosophy, then patterns. SKILL.md sizes range from ~7 KB to ~27 KB (see `README.md` table).
- **`<skill>/.claude-plugin/plugin.json`** — the per-plugin manifest. Validates against `https://json.schemastore.org/claude-code-plugin-manifest.json`. Fields: `name`, `displayName`, `description`, `version`, `author`, `homepage`, `repository`, `license`, `keywords`. `name` must match the directory name.
- **`.claude-plugin/marketplace.json`** — the root marketplace catalog. Validates against `https://json.schemastore.org/claude-code-marketplace.json`. Each entry's `source` is a relative path (`./<skill>`) and `name` must match the entry's directory.

## Development workflow

There is no `make` / `go build` / `go test`. To develop on a skill:

1. **Edit** `<skill>/SKILL.md` (content) or `<skill>/.claude-plugin/plugin.json` (manifest).
2. **Validate JSON** against the schemas declared in their `$schema` field. Both schemas are public — fetch and validate with any JSON Schema validator, e.g.:
   ```bash
   # quick syntactic check
   for f in .claude-plugin/marketplace.json */.claude-plugin/plugin.json; do
     python3 -c "import json,sys; json.load(open('$f'))" && echo "OK $f"
   done
   ```
3. **Verify frontmatter** in `SKILL.md`: the top of every `SKILL.md` is YAML between `---` lines. `name` must equal the directory name; `description` is the activation prompt — keep it specific so the skill fires on relevant requests and stays silent on others.
4. **Cross-check `marketplace.json`**: when adding or renaming a plugin, update `.claude-plugin/marketplace.json` to match. The `name` and `source` there must agree with the directory.
5. **Update `README.md`** if you add/remove a skill or change the per-skill pitch.

## Editing principles (from `README.md`)

These apply to skill content, not to runtime code:

- **Domains over layers** — no `internal/` junk drawer, no `pkg/` anti-pattern; organize by what code *does*.
- **Standard library over frameworks** — `testing`, table-driven tests, simple stubs over BDD / mock-generation frameworks.
- **Channels over mutexes** — native concurrency primitives, not static worker pools.
- **Command-first architecture** — CLI as a router, business logic decoupled from Cobra.
- *Clear is better than clever.* Delete the abstraction when in doubt.

## Git workflow

Single long-lived `main` branch (see `git log`). Commits follow `<type>(<scope>): <subject>` — e.g. `feat(wails): add Wails desktop app skill`, `docs(fileflow): sync skill with current API`. PRs are merged into `main` and deployed by the marketplace manifest.
