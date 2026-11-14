# Brief (locked 2026-11-04)

## Intent
A Claude Code marketplace of Go Agent Skills by spf13 that steers models toward idiomatic,
version-correct Go. Golden path: `/plugin marketplace add spf13/go-skills` → `/plugin install <skill>@go-skills`
→ user asks a Go question → skill fires from its `description` → advice is correct for the module's `go` floor.

## Golden path smoke (every pass)
1. `python3 .github/scripts/validate_manifests.py`
2. `python3 .github/scripts/check_invariants.py`
3. `python3 .github/scripts/check_go_release.py go1.27.3`
4. `claude plugin validate .` and `claude plugin validate ./<skill>` for every skill
5. Any test suite / skill-validation harness added during the run (must pass).

## Quality bar (fixtures)
- Every Go example parses with go1.27.3 `gofmt`, unless it is an intentional fragment.
- Every API or version claim matches go1.27.3 (`go doc`, `$GOROOT/api/go1.X.txt`).
- Reference toolchain: go1.27.3 via `GOTOOLCHAIN=go1.27.3` (proxy.golang.org). No user fixtures supplied.

## Invariants (sacred)
- Skill directory name = plugin.json `name` = SKILL.md frontmatter `name`. No renames.
- marketplace.json entries keep exactly four keys; plugin.json is the source of truth.
- One catalog version across all plugins. **No version bump**; changes go under `## [Unreleased]` in CHANGELOG.md.
- `go/SKILL.md` keeps the literal `current through Go 1.27`.
- spf13's opinions are the product voice (flat packages, no `pkg/`, no mocks, command-first CLIs,
  stdlib over frameworks). Edit skill text for factual errors and clarity only, never to flip a position.

## Non-goals
- No splitting `go/SKILL.md` into `references/` (conservative appetite).
- No rewriting skills wholesale. No renaming or removing skills.

## Appetite
Conservative: fix facts, add tests, dedupe scripts, refine text. **Exception (done signal): new skills may be added**
when they fit the catalog and every claim in them can be verified here.

## Boundaries
- Dependencies: standard library only (Python stdlib `unittest`; `jsonschema` stays as the existing CI dep).
- Network during smoke: proxy.golang.org and pkg.go.dev only. go.dev is unreachable from the sandbox.
- CI: may add a unit-test job. The done signal asks for "validation of the skills", so a skill-validation job
  that installs Go with `actions/setup-go` is allowed too. No other CI or release-config edits. No tags.
- AGENTS.md: generalize it and remove private hostnames, absolute paths, and machine names (user answer to Q9).

## Contradictions resolved
- Missing `v*` tags: flag in the report; the user pushes tags.
- README "Channels over mutexes": the skill is right; soften the README line.
- Hard-coded "six": say "every skill" where a count isn't needed; keep "Six" in the README pitch.

## Ranked rocks
1. Content fidelity: every Go example and version claim verified on go1.27.3.
2. CI guardrails: unittest suite, dedupe skill discovery, a skill-validation harness.
3. Additional skills + refinement (the done signal).
(go/SKILL.md context-cost split: cut by conservative appetite.)

## Done signal
A better test harness and mechanical validation of the skills; skill content refined and verified on go1.27.3;
additional skills where they fit; everything green; a draft PR to review.
