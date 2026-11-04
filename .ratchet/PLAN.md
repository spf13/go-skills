# Plan (draft, pre-brief)

## Inferred brief
- **Product:** a Claude Code plugin marketplace of six Go Agent Skills. Each `SKILL.md` is loaded into a
  model's context to steer it toward spf13's idiomatic, version-correct Go. (README.md:1-3, marketplace.json)
- **Primary user:** Go developers using Claude Code (and other agents via `~/.agents/skills` symlinks),
  first of all spf13 across galactus/factory/wandamax (AGENTS.md).
- **Golden path:** `/plugin marketplace add spf13/go-skills` → `/plugin install go@go-skills` → user asks a
  Go question → skill fires (frontmatter `description`) → advice is correct for the module's `go` floor.
  Smoke = `claude plugin validate` + the three CI scripts + compile check of skill examples on go1.27.3.
- **Invariants I think exist:**
  - Skill directory names = plugin names = frontmatter `name` (check_invariants.py:139-142). Renames break installs.
  - marketplace entry has exactly 4 keys; plugin.json is source of truth (CONTRIBUTING.md:30-43, check_invariants.py:124).
  - One catalog version across all plugins; CHANGELOG entry per version (check_invariants.py:158-177).
  - `go/SKILL.md` description keeps the literal `current through Go 1.27` (check_go_release.py:27-29).
  - The opinions (flat packages, no `pkg/`, no mocks, command-first CLIs) are the product's voice, not defects (README.md:55-80).

## Top 3 rocks
1. **Content fidelity: every Go example and API/version claim is true on go1.27.3.** (archetype: content engine)
   Evidence: 88 Go fences, 0 mechanically checked (BASELINE); go/SKILL.md is the churn leader (12 commits) and
   carries 1.26/1.27 claims that LLM-authored edits are most likely to get wrong. Spot check passed 10/10, so a
   full sweep is needed to find the misses. Lenses: fidelity → correctness → completeness → docs/voice → simplify.
2. **CI guardrails are untested and duplicated.** (archetype: engine)
   Evidence: 389 lines across 3 scripts, 0 tests; skill discovery written 3 times
   (check_invariants.py:54, check_go_release.py:40, validate_manifests.py:41); `EXCLUDED` twice
   (check_invariants.py:20, validate_manifests.py:22); hand-rolled YAML parser (check_invariants.py:35-51).
   Lenses: tests → correctness → architecture → simplify.
3. **go/SKILL.md context cost.** (archetype: content)
   Evidence: 1188 lines / 50 KB (~12k tokens) loaded whenever Go is mentioned; the 1.17–1.27 version index
   (go/SKILL.md:34-102) and per-release sections are reference material that suits `go/references/` with
   progressive disclosure. Lenses: simplify → architecture → docs/voice.

## Findings backlog (ranked)
| # | Lens | Finding | Evidence |
| --- | --- | --- | --- |
| F1 | fidelity | Full audit of 88 Go blocks + version-index rows against go1.27.3 not done | BASELINE |
| F2 | tests | No tests for frontmatter parser, invariants, or release claim logic | .github/scripts/ |
| F3 | correctness/docs | No `v*` tags locally or on origin; CHANGELOG links to releases/tag/v1.0.0, v1.1.0; the "never behind latest tag" rule silently no-ops | CHANGELOG.md:30,44; check_invariants.py:85-93 |
| F4 | architecture | Skill discovery ×3, EXCLUDED ×2 | see rock 2 |
| F5 | docs | README "Channels over mutexes" overstates the skill, which calls `sync.Mutex` "the gold standard" and keeps mutexes where they belong | README.md:68 vs go/SKILL.md:123,157,317 |
| F6 | docs | "six" hard-coded in 5 places; every new skill must edit all | README.md:3; CONTRIBUTING.md:3,23,64,72; CHANGELOG.md:5 |
| F7 | fidelity | go.mod `retract` block fenced as ```go | go-release/SKILL.md:157 |
| F8 | security | AGENTS.md publishes private hostnames and paths (factory, wandamax, /share/steve, /root/.agents) in a public repo | AGENTS.md:15-41 |
| F9 | simplify | `.roborev.toml` is 307 lines, nearly all empty defaults | .roborev.toml |
| F10 | docs | Local check step 3 depends on go.dev, which some networks block; proxy.golang.org toolchain list is an alternative | CONTRIBUTING.md:107 |
| F11 | reality | lychee link check not reproducible locally | BASELINE |
