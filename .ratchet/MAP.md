# Map

**Archetype: mixed** (content product + small CI script engine).

## Entry points / user-facing surfaces
- `<skill>/SKILL.md` ×6: the product. Frontmatter `description` drives triggering; body loads into context on invoke.
- `.claude-plugin/marketplace.json`: catalog; `/plugin marketplace add spf13/go-skills`.
- `<skill>/.claude-plugin/plugin.json` ×6: source of truth for display metadata.
- `README.md` (install + manifesto), `CONTRIBUTING.md` (rules = CI contract), `CHANGELOG.md`.
- `templates/skill/`: new-skill scaffold.

## Core modules (CI)
- `.github/scripts/validate_manifests.py` (74 lines): schemastore validation, needs network + jsonschema.
- `.github/scripts/check_invariants.py` (197): cross-file rules; hand-rolled frontmatter parser; tag/version logic.
- `.github/scripts/check_go_release.py` (118): "current through Go X.Y" claim vs latest; imports from check_invariants.
- `.github/workflows/validate.yml` (push/PR), `go-release-check.yml` (monthly, opens issue).

## Tests / fixtures
- None. No unit tests, no fixtures. Only "test" is running scripts against the live repo.

## Churn hotspots (`git log --name-only`)
12 go/SKILL.md · 9 README.md · 6 CONTRIBUTING.md · 4 go-spec-reviewer/SKILL.md · 4 cobra-viper/SKILL.md ·
4 check_invariants.py · 4 marketplace.json · 3 .gitignore

## Other tracked files
- `AGENTS.md`: operator sync protocol (hostnames factory/wandamax, sage). `.roborev.toml` (307 lines, mostly empty defaults). `.kata.toml`.
