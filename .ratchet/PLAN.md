# Plan (refreshed after sweep 3)

## Brief recap
See BRIEF.md (locked). Rocks: 1 content fidelity ✔ (S1), 2 CI guardrails ✔ (S2), 3 additional skills + refinement ✔ (S3: afero).
Final bookend sweep: architecture, simplify, ux, design/craft, docs, reality across the product.

## Done (see RATCHET_LOG.md for shas and judge verdicts)
- F1 full audit of Go fences + version claims → check_skill_code.py (93 fences, 269 std symbols, 84 claim tokens on go1.27.3).
- F2 tests → test_checks.py (62 tests, CI) + .github/claims (4 behavioral tests, CI).
- F4 discovery ×3 → 1. F7 go.mod fence. F8 AGENTS.md private details. F5 README mutex line. F6 "six". F10 proxy fallback.
- Content errors fixed: strict-JSON claim ×4 passages, nogreenteagc forecast ×2, `cobra.Args`, missing json/v2 index row, nojsonv2 note.
- New skill: afero.

## Remaining backlog (ranked)
| # | Item | Why not done | Effort |
| --- | --- | --- | --- |
| B1 | `hugo` skill (spf13's own project; heavy version-blending risk like wails) | Sandbox classifier denied building/running a Hugo binary; brief requires every claim verified | M (needs permission to run hugo, or user-supplied verification) |
| B2 | Pinned third-party API checks (cobra, viper, fileflow, pathologize, wails, afero) — `go doc`/compile tests against pinned module versions | Adds Go module deps to the repo; brief is stdlib-only | M |
| B3 | Push the `v1.0.0`/`v1.1.0` tags (absent locally and on origin; CHANGELOG links 404; the "behind latest tag" rule no-ops) | Agent may not create tags | S (user) |
| B4 | Re-run `sage sync` / dotagents on your machines for the new `afero` skill (old AGENTS.md protocol) | Cloud session cannot reach your hosts | S (user) |
| B5 | check_skill_code survivors from judge 7: off-by-one variants, `(Go 1.X+)` heading tests, value-receiver keys, TEST_RECEIVERS width | Diminishing returns | S |
| B6 | Bare `Type.Method` in prose without its package on the line is unchecked (documented exemption) | Needs a context model across lines | M |
| B7 | `go/SKILL.md` context cost (1188 lines, ~50 KB) → `references/` | Non-goal under conservative appetite | M |
| B8 | Skill triggering not exercised (needs paid model calls) | Out of smoke boundary | — |
| B9 | Private AGENTS.md text remains in history (9adc400); hostnames/paths, no credentials | History rewrite of main is forbidden | — |
