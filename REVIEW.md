# Review guidelines: go-skills

Steve Francia's Go Agent Skills marketplace: six skill directories (`go`, `cobra-viper`, `go-spec-reviewer`, `go-release`, `wails`, `fileflow-pathologize`) each with `SKILL.md` + optional `references/`, a `.claude-plugin/marketplace.json` catalog, `templates/skill/` scaffold, and a hermetic checker — `go run ./cmd/check` (go.mod floor `go 1.22`, `internal/check/`) — that enforces the `CONTRIBUTING.md` rules in CI.

## Flag (medium or higher)

- Checker/contract drift: a rule stated in `CONTRIBUTING.md` that `internal/check` does not enforce, or enforces differently — e.g. version equality that fails on an untagged version-bump PR (jobs 12, 13), tags↔keywords compared as ordered lists instead of sets, `templates/` not excluded, a `claimingSkills` claim added to a skill that only teaches individual features (§Go release currency).
- Frontmatter contract in any edited `SKILL.md`: `name` must equal the directory and `plugin.json.name`; `description` opens `Use when…`, stays third person, carries no coercive activation or author-credential framing, and keeps the `current through Go X.Y` claim as a trailing clause only for skills listed in `claimingSkills`.
- Catalog correspondence: a skill directory added/renamed/removed without matching `.claude-plugin/marketplace.json` and `plugin.json` versions/entries, or an entry whose `tags` diverge from `keywords`.
- Documented commands and flags that the checker or CI does not actually accept — a `go run ./cmd/check` invocation, exit code (0/1/2), or `-root`/`-latest` behavior that contradicts `CONTRIBUTING.md:161`.
- Compilable samples: a complete fenced `go` sample (with its own `package` clause) that would fail to build/vet, or a claim of Go-version support below the module floor.

## Don't flag

- Prose style, headings, table alignment, or wording in `SKILL.md`/`references/`; markdown nits and link styling.
- Partial code excerpts that are not meant to compile — the checker itself skips anything without a `package` clause.
- "Consider adding a seventh skill / more tests / more examples" suggestions, and hardening for inputs the hermetic checker never sees.
- Scaffold files under `templates/` being incomplete or inconsistent with a real skill; they are intentionally excluded.

## Severity

- high: reachable bug, CI-breaking regression, catalog/frontmatter invariant violation, or wrong documented command.
- medium: real defect with limited reach, or a missing check for a documented rule / missing test for new checker logic.
- low: everything else (not reported; review_min_severity is medium).
