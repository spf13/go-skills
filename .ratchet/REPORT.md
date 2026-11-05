# Ratchet report: go-skills

**Done signal: met.** The run asked for a better test harness, mechanical validation of the skills, refined content, and new skills where they fit. Each is shipped and verified below.

Branch `claude/festive-curie-d6zqps`, base `9adc400`. Three sweeps plus a final bookend. Every pass was reviewed by an independent judge running in a fresh context. Reference toolchain: go1.27.3, the latest stable release.

## Baseline vs final

| Metric | Baseline (9adc400) | Final (29aeaff) |
| --- | --- | --- |
| Go fences mechanically parsed | 0 / 88 | 93 / 93 |
| Std symbol mentions verified to exist | 0 | 269 |
| Std version-claim tokens verified (stdversion manifest) | 0 | 84 |
| Behavioral claim tests (`.github/claims`, `go test`) | 0 | 4 |
| Python unit tests for the check scripts | 0 | 62 |
| Factual errors fixed in skills | — | 7 (see below) |
| Skills | 6 | 7 (`afero`) |
| Skill-discovery implementations in CI scripts | 3 | 1 |
| Private hosts/paths in tracked files | 15+ (AGENTS.md) | 0 |
| CI: jobs / steps | 3 / 4 | 4 / 7 |
| Full local gate, clean clone | ~3 s (3 scripts) | 9.8 s (everything above + `claude plugin validate` × 8) |
| SKILL.md bytes | 125,486 | 134,020 (+7.4 KB afero, +1.1 KB fixes) |

## Factual errors found and fixed in the skills

All seven were reproduced on go1.27.3 before the fix, then verified again by a judge.

1. **Strict JSON, the biggest one.** `cobra-viper` §6 and its Common Mistakes entry, plus `go` and `go-spec-reviewer`, said Go 1.27's `encoding/json` rejects duplicate object names and invalid UTF-8, so users' working config files would "start failing".
   - That is false. The v1 API keeps v1 behavior on the new engine; only `encoding/json/v2` is strict (`src/encoding/json/v2_options.go:89-96`).
   - The rewrite shows how to validate a config with `encoding/json/v2` before Viper reads it. The snippet was compiled and run against viper v1.21.0.
2. **Green Tea opt-out forecast** (`go`, `go-release`). The skills said `GOEXPERIMENT=nogreenteagc` would be removed in Go 1.27. It still builds on 1.27.3.
3. **Nonexistent Cobra identifier** (`go-spec-reviewer`). The skill named `cobra.Args`, which doesn't exist; the validators are `cobra.PositionalArgs` values set on the `Args` field.
4. **Missing index rows** (`go`). The version index lacked the Go 1.27 `encoding/json/v2` and `encoding/json/jsontext` packages.
5. **Unstated build break** (`cobra-viper`). It didn't say that `GOEXPERIMENT=nojsonv2` also removes `encoding/json/v2`, which breaks the strict-config snippet.
6. **Mislabelled fence** (`go-release`). A `go.mod` `retract` block was fenced as Go code.
7. **Overstated README line.** "Channels over mutexes" contradicted the go skill, which calls `sync.Mutex` "the gold standard" for guarding state. It now reads "Channels before mutexes. Pass data over a channel where you can; keep the mutex for guarding state."

## Kept changes

| SHA | Lens | Change | Judge |
| --- | --- | --- | --- |
| 3fe48f0 | S1 fidelity | `check_skill_code.py`: gofmt-parse every Go fence; std version claims vs the toolchain's stdversion manifest; CI job | 68→80 |
| 825fdb8 | S1 correctness | Strict-JSON fix (4 passages), nogreenteagc, `.github/claims` test module | 70→86 |
| d69106e | S1 completeness | Catch invented std APIs (prose + fences), indented fences, context-package claims, clean errors | 40→66 |
| 646717b | S1 docs | `cobra.Args`, json/v2 index row, nojsonv2 note | 86→92 |
| cc58d0c | S2 correctness | Module-level error state leaked across runs; duplicate discovery; 5 checker gaps; 39-test suite (8 red on parent) | 68→76 |
| d965586 | S2 tests | Unit tests in CI; per-fence declarations (regression fix); skip below Go 1.27; hermetic git fixture | 55→66 |
| 7476fec | S2 tests | Exit-code invariant in every test; 21 tests killing 22/22 surviving mutants | 52→68 |
| ac6c74c, 57c4e02 | S2 architecture | One skill discovery; dead EXCLUDED entries; dot-dirs never skills; host git config isolated | 62→74 |
| 27b238c | S2 security | AGENTS.md generalized; private hosts and paths removed | 35→95 |
| 5261cc3, fb3fc0e, e1c950b | S2 docs | README concurrency line; module-proxy fallback for the latest Go; CHANGELOG; AGENTS CI wording | 80→87 |
| 6e76008 | S3 completeness | New `afero` skill (every claim tested against afero v1.15.0, `-race`) | 78→84 |
| cbc6e3e | S3 tests | Claim test: `godebug asynctimerchan=1` fails the build on Go 1.27 | 80→91 |
| 51b35d1, 29aeaff | S3 correctness | afero dry-run caveats (CopyOnWriteFs cannot stage deletes/moves), NewIOFS/ReadAt wording | 80→91 |
| 4a35172 | Final craft | One failure format across all four checks | 82→86 |

`.ratchet/RATCHET_LOG.md` has the critique, verification commands, and judge reasoning for each pass.

## Reverted attempts

None. Two regressions introduced mid-run were caught by judges and fixed in a later commit, not reverted, because each commit still beat its parent:
- **Skill-wide declarations** (cc58d0c). One `fs afero.Fs` parameter hid every `fs.*` typo in the go skill. Judge 5 caught it; d965586 fixed it, with a test.
- **Untestable variadic test and host-git fragility** (d965586, 7476fec). Judges 6 and 7 caught these; 7476fec and ac6c74c fixed them.

## Remaining backlog (ranked)

| # | Item | Why not done | Effort |
| --- | --- | --- | --- |
| B1 | A `hugo` skill (your project; the same version-blending risk `wails` exists for) | The sandbox classifier denied building and running a Hugo binary, and the brief requires every claim verified | M |
| B2 | Pinned third-party API checks (cobra, viper, fileflow, pathologize, wails, afero): `go doc` and compile tests at fixed versions. A one-off scan found `cobra.Args`, the only invented identifier in about 100 refs | Adds Go module dependencies to the repo; the brief is stdlib-only | M |
| B3 | Push the `v1.0.0`/`v1.1.0` tags. They are absent locally and on origin, so the CHANGELOG release links 404 and the "never behind the latest tag" rule does nothing | Agents may not create tags | S, you |
| B4 | Run your skill-sync tool on your machines so `afero` gets linked | A cloud session can't reach your hosts | S, you |
| B5 | Checker edge mutants that survive: off-by-one variants, tests for `(Go 1.X+)` headings, value-receiver keys | Diminishing returns | S |
| B6 | A bare `Type.Method` in prose without its package on the same line goes unchecked (a documented exemption) | Needs context across lines | M |
| B7 | `go/SKILL.md` context cost (1188 lines, about 50 KB) | Non-goal under the conservative appetite | M |

## Risks to review by hand

- **Old AGENTS.md is still in git history** (`9adc400`, already on `main`). It holds hostnames and paths, no credentials. The rewrite also drops your private sync protocol from this repo; it belongs in your private `dotagents` repo.
- **CI's `skill code` job tracks `go-version: stable`.** When Go 1.28 ships, the `.github/claims` tests may go red if behavior changes, and the monthly drift check will flag `current through Go 1.27`. That is intended, but expect it.
- **Afero behavior is pinned only by this run's scratch tests**, not in CI (see B2). A future afero release could change a documented quirk unnoticed.
- **Skill triggering wasn't exercised** [UNVERIFIED]; that needs paid model calls. `claude plugin validate` passes for all 7 skills on CLI 2.1.318.
- **The lychee link check wasn't run locally** [UNVERIFIED locally]; CI runs it. The new README row links `./afero/SKILL.md`, which exists.
- **The new `afero` skill is written in your voice.** Read `afero/SKILL.md` before it ships under your name.
