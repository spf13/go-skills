# Ratchet log

Process note: implementer passes run in the main session (it holds the inventory context); every judge
is a fresh subagent that never wrote the change.

### S1 · fidelity — KEPT — 68→80 — 3fe48f0
- Critique: 88 Go fences and ~80 std version claims had no mechanical check (BASELINE); go-release/SKILL.md:157 fenced go.mod as ```go.
- Change: add .github/scripts/check_skill_code.py (gofmt parse of every go fence; std version claims vs the toolchain's stdversion manifest), a `skill code` CI job on setup-go stable, CONTRIBUTING rule; refence go.mod as text.
- Verified: `GOTOOLCHAIN=go1.27.3 python3 .github/scripts/check_skill_code.py` → 87 fences, 80 claims OK; invariants, manifests, go_release go1.27.3, `claude plugin validate` ×7 green. Mutation: 7/8 caught (implementer), 16/22 version + 5/5 fence (judge).
- Judge: KEEP — real, regression-free; follow-ups: nonexistent symbols skipped, context-package subject gap, indented fences skipped, tracebacks on missing go/manifest, old-manifest false errors, CONTRIBUTING wording, no Unreleased entry, no unit tests.

### S1 · correctness — KEPT — 70→86 — 825fdb8
- Critique: cobra-viper/SKILL.md:384-405, :486, go/SKILL.md:794, go-spec-reviewer/SKILL.md:59 claimed Go 1.27 encoding/json rejects duplicate names and invalid UTF-8; false on go1.27.3 (v1 API keeps v1 semantics; only encoding/json/v2 is strict — src/encoding/json/v2_options.go:89-96). go/SKILL.md:914 and go-release/SKILL.md:114 forecast nogreenteagc removal in 1.27; it still builds on 1.27.3.
- Change: rewrite the four passages; cobra-viper shows validating config bytes with encoding/json/v2 before Viper (compiled + run vs viper v1.21.0); add .github/claims stdlib test module (json v1 lenient / v2 strict, unbuffered timers) + CI step; CHANGELOG [Unreleased].
- Verified: failing test of old claim → FAIL on go1.27.3; claims suite ok; check_skill_code 88/80, invariants, go_release, `claude plugin validate` green.
- Judge: KEEP — old claim reproduced false, replacement verified true. Minor: nojsonv2 breaks the json/v2 snippet (unmentioned); timer test doesn't pin asynctimerchan removal; json.Options mention dropped.

### S1 · completeness — KEPT — 40→66 — d69106e
- Critique: check_skill_code.py skipped nonexistent std symbols, indented fences, context-package subjects; tracebacks on missing go/manifest (judge 1).
- Change: unknown-symbol check in prose and fences (non-std module names skipped), subject = first symbol after context package, indented fences, clean one-line errors, CONTRIBUTING rules rewritten.
- Verified: 89 fences / 222 symbols / 81 claim tokens on go1.27.3; invariants, go_release, claims, `claude plugin validate` green; implementer mutation 9/9.
- Judge: KEEP — judge mutation 18/28 in scope. Follow-ups → S2: skill-wide foreign names hide std `uuid`/`cmp` (go/SKILL.md mentions google/uuid, go-cmp); `range pkg.X` read as a declaration (hides go/SKILL.md:858); filename `errors.go` false positive; cross-fence variables; `go env` failure traceback; bare `Type.Method` without context unchecked (document).

### S1 · docs/voice — KEPT — 86→92 — 646717b
- Critique: go-spec-reviewer/SKILL.md:106 named nonexistent `cobra.Args`; go/SKILL.md version index lacked 1.27 `encoding/json/v2`/`jsontext`; cobra-viper didn't say nojsonv2 removes json/v2.
- Change: point to `Args` field + `cobra.PositionalArgs`; add index row; one clause in cobra-viper; CHANGELOG lines.
- Verified: `go doc` vs cobra v1.10.2; manifest shows both packages at 27; nojsonv2 build of the snippet fails; smoke green.
- Judge: KEEP — fixes a nonexistent identifier; all claims verified.

### S1 · simplify — CONVERGED — n/a
- Critique: the brief's conservative appetite rules out restructuring skill content; no dead or duplicated skill text surfaced in this sweep. Script simplification belongs to rock 2 (S2).

### S1 · reality — KEPT — gate passed — 646717b
- Verified (clean tree, go1.27.3): validate_manifests 7 OK; invariants OK v1.1.0; go_release go1.27.3 OK; check_skill_code 89 fences / 225 symbols / 83 claim tokens; claims `go vet` + `go test` OK; `claude plugin validate` root + 6 skills OK.

## Sweep 1 scorecard (rock: content fidelity)
| Lens | Result | Score |
| --- | --- | --- |
| fidelity | KEPT | 68→80 |
| correctness | KEPT | 70→86 |
| completeness | KEPT | 40→66 |
| docs/voice | KEPT | 86→92 |
| simplify | CONVERGED | — |
| reality | PASS | — |

| Metric | Baseline | S1 |
| --- | --- | --- |
| Go fences mechanically parsed | 0 / 88 | 89 / 89 |
| Std symbol mentions verified to exist | 0 | 225 |
| Std version claim tokens verified | 0 | 83 |
| Behavioral claim tests (go test) | 0 | 3 |
| Python unit tests | 0 | 0 |
| Factual errors fixed in skills | — | 5 (strict-json ×4 passages, nogreenteagc ×2, cobra.Args, go.mod fence) |
| SKILL.md total bytes | 125486 | 126065 |
| CI jobs | 3 | 4 |

## Sweep 2 (rock: CI guardrails)

### S2 · correctness — KEPT — 68→76 — cc58d0c
- Critique: module-level error lists leaked across main() calls (check_invariants.py:23, check_go_release.py:33); check_go_release's private skill discovery ignored the root; judge-3 checker gaps (std uuid/cmp hidden, `range pkg.Fn` as declaration, `errors.go` flagged, cross-fence vars, `go env` traceback).
- Change: errors local to main; check_go_release reuses declared_skills; checker fixes; .github/scripts/test_checks.py (39 tests via main() on fixture catalogs).
- Verified: suite on HEAD^ scripts → 7 failures + 1 error; on HEAD → 39 OK; smoke green.
- Judge: KEEP — fixes outweigh one regression: skill-wide `declares` hides typos (go/SKILL.md:512 `fs afero.Fs` hides every `fs.*`); variadic params lost; Go-dependent tests fail instead of skip on go<1.27; InvariantsTest runs git in an unmanaged temp dir.

### S2 · tests — KEPT — 55→66 — d965586
- Critique: suite not in CI; skill-wide `declares` hid typos (judge 5); Go-dependent tests failed (not skipped) below 1.27; invariants tests ran git in an unmanaged temp dir.
- Change: unittest step in the skill-code CI job; per-fence declarations; variadic params; skip below Go 1.27; `git init` fixture.
- Verified: 40 tests OK on go1.27.3; 40 run / 20 skipped on go1.24.7; fs.PathErrorz mutation at go/SKILL.md:444 caught. (Commit message says 41/21 — off by one; recorded here as the correction.)
- Judge: KEEP — mutation 35/58 killed. Survivors: invariants (dup tags, source, dup entries, entry w/o dir, plugin.json name, missing description, MIN_DESCRIPTION, README link mismatch, set-vs-list, missing CHANGELOG, EXCLUDED), go_release (claim ahead, env fallback), skill_code (non-subject exist rule, prose foreign exemption, std-import override, variadic test can't fail, var/for decls, t./b. receivers, code_only stripping, main exit code).

### S2 · tests (pass 2) — KEPT — 52→68 — 7476fec
- Critique: judge 6's 23 surviving mutants; no test asserted a non-zero exit.
- Change: run() asserts exit code agrees with reported problems; 21 targeted tests; `var` declaration regex narrowed to the one form nothing else covers.
- Verified: 61 tests OK (go1.27.3), 26 skipped on go1.24.7; implementer re-ran judge-6 survivors: 22/22 killed.
- Judge: KEEP — fresh campaign 33/49 (parent 25/48). Defects: real-git-tag test breaks under host `commit.gpgsign=true`; `b.Loop()` half of a test can't fail. Remaining survivors (off-by-ones, (Go 1.X+) headings, value-receiver keys, TEST_RECEIVERS width) → backlog.

### S2 · architecture — KEPT — 62→74 — ac6c74c
- Critique: skill discovery copied 3× (check_invariants.py:54, check_go_release.py:40 [removed in cc58d0c], validate_manifests.py:41); EXCLUDED twice with dead entries; fixture inherited host git config (judge 7).
- Change: validate_manifests uses declared_skills; EXCLUDED = {"templates"}; fixture isolated from host git config; b.Loopz assertion.
- Verified: same 7 manifests validated before/after; 61 tests OK; tests pass under a signing global config (parent: 1 error).
- Judge: KEEP — equivalence proven. Minor: a hidden dir with plugin.json now counts as a skill → fixed next commit (skip dot-dirs).

### S2 · security — KEPT — 35→95 — 27b238c
- Critique: AGENTS.md published private hostnames, absolute paths, machine names, ssh commands (AGENTS.md:15-41 at 9adc400).
- Change: generalized AGENTS.md (layout, CI checks, serving from a local clone); skill list points to README.
- Verified: every listed command run green; grep of all tracked files for the removed names: no private hits.
- Judge: KEEP — none remain outside git history (9adc400, no credentials). Nit: "CI runs the same set" omits the lychee link check → fixed in the docs follow-up.

### S2 · architecture (follow-up) — 57c4e02
- Change: skip every dot-directory in skill discovery (judge 8 minor); test added. 62 tests OK.

### S2 · docs/voice — KEPT — 80→87 — 5261cc3, fb3fc0e (+ 57c4e02 reviewed)
- Critique: README.md:68 "Channels over mutexes" overstated the go skill (go/SKILL.md:123,317); CONTRIBUTING release lookup depended on go.dev only; AGENTS omitted the link check.
- Change: README line states the skill's rule; proxy.golang.org fallback command; CHANGELOG Changed section; AGENTS CI wording.
- Verified: fallback command run as written → go1.27.3, pass; 62 tests; smoke green.
- Judge: KEEP — voice kept. Minor: test suite belongs under Added; AGENTS omits the scheduled release check → both fixed in the next commit.

### S2 · simplify — CONVERGED — n/a
- Critique: discovery deduped (ac6c74c), dead EXCLUDED entries and a redundant declares() branch removed (7476fec). Remaining script length is rules plus tests; further cuts would be churn.

### S2 · reality — PASS — e1c950b
- Verified on a fresh clone of origin/claude/festive-curie-d6zqps (go1.27.3): unittest OK; validate_manifests 8 OK; invariants 7 skills OK; go_release OK; check_skill_code 93 fences / 269 symbols / 84 claim tokens; claims vet+test OK; `claude plugin validate` root + 7 skills OK. (Counts include the afero skill, 6e76008, under S3 review.)

## Sweep 2 scorecard (rock: CI guardrails)
| Lens | Result | Score |
| --- | --- | --- |
| correctness | KEPT | 68→76 |
| tests | KEPT ×2 | 55→66, 52→68 |
| architecture | KEPT | 62→74 |
| security | KEPT | 35→95 |
| docs/voice | KEPT | 80→87 |
| simplify | CONVERGED | — |
| reality | PASS | — |

| Metric | Baseline | S1 | S2 |
| --- | --- | --- | --- |
| Python unit tests | 0 | 0 | 62 |
| Judge mutation kill rate (scripts) | n/a | n/a | 33/49 fresh campaign; 22/22 prior survivors |
| Skill-discovery implementations | 3 | 3 | 1 |
| Module-level mutable error state | 2 | 2 | 0 |
| Private host/path mentions in tracked files | 15+ (AGENTS.md) | same | 0 |
| CI jobs / steps | 3 / 4 | 4 / 6 | 4 / 7 |

## Sweep 3 (rock: additional skills + refinement)

### S3 · completeness — KEPT — 78→84 — 6e76008 (+ wording fix below)
- Critique: the catalog had no skill for spf13/afero although the go skill names it the standard (go/SKILL.md:500-516) with one line of advice. A Hugo skill was the first choice; the sandbox classifier denied building/running a Hugo binary, so it is backlogged rather than written unverified.
- Change: afero/SKILL.md + plugin.json + marketplace entry (development), README row, CONTRIBUTING category, CHANGELOG; "six" removed where a count was not needed; go skill cross-reference.
- Verified: implementer scratch module (afero v1.15.0, go1.27.3, -race): 11 tests covering every claim and both code examples; all repo checks + `claude plugin validate ./afero` green.
- Judge: KEEP — independent tests: 27/30 claims exact, 0 false, 3 imprecise (CopyOnWriteFs dry run cannot remove/rename base-only files; ReadAt wording; NewIOFS ReadDir lists rooted dirs). All three reproduced by the implementer (RemoveAll returns nil and the file stays visible) and reworded in the next commit.
