# Project Audit: `go-skills`

- **Date**: 2026-10-26
- **Target**: `/share/steve/code/go-skills` (`spf13/go-skills` — 6 skills, `2,406 LOC` Markdown/JSON + 3 Python CI scripts)
- **Surfaces Audited**: `Agent Skills (go, cobra-viper, go-release, go-spec-reviewer, fileflow-pathologize, wails)`, `Claude Plugin Marketplace Manifests`, `CI Validation Scripts`
- **Lenses Run**: `PM, UX/AX, Architecture, Engineering` + `Adversarial Skeptic`
- **Verification Summary**: `20 raw findings → 2 refuted / 11 merged by skeptic → 7 confirmed recommendations`
- **Estimated Simplification Leverage**: `go/SKILL.md split from 1,189 → ~350 core lines + references/, 3 Python scripts (~450 LOC) replaced by 1 hermetic Go validator, 7-file version fan-out collapsed`

---

## 1. Executive Diagnosis & Goal Alignment

- **Stated Goal**: Authoritative Go agent skills marketplace (`spf13/go-skills`) by Steve Francia, teaching idiomatic Go (through Go 1.27), Cobra/Viper CLI architecture, design-spec review, release engineering, safe file operations, and Wails desktop apps.
- **Codebase Reality**: 6 `SKILL.md` files (`2,572 lines` total, led by `go/SKILL.md` at `1,189 lines`), 7 `.claude-plugin/*.json` manifests, and 3 Python CI scripts (`.github/scripts/{check_invariants.py,check_go_release.py,validate_manifests.py}`) plus a 307-line empty `.roborev.toml`.
- **Core Alignment Gaps**:
  1. **Cross-Skill Contradictions & Non-Compiling Samples**: `go/SKILL.md:171` prescribes `main.go` at module root for single-binary apps while `cobra-viper/SKILL.md:48-53` prescribes `cmd/`; `go/SKILL.md:1135` calls non-existent top-level `slog.With(...)` (`undefined: slog.With` — only `slog.Default().With(...)` or `logger.With(...)` exists); `cobra-viper/SKILL.md:74,271` suppresses errors with `SilenceErrors: true` then calls `os.Exit(1)` in `main` without printing the error and declares package-level `var cfg Config`; and `go-spec-reviewer/SKILL.md:19-35` ships a literal `[SPEC_FILE_PATH]` placeholder.
  2. **Trigger Over-Broadness & Progressive Disclosure Violation**: `go/SKILL.md:3-6` triggers on any mention of `Go` or `go.mod` with zero negative trigger boundaries (`not for...`), starving the 5 specialist skills, and loads all `1,189 lines` into context with zero `references/` split (`skill-creator` budget is ≤500 lines).
  3. **Python CI Scripts & 7-File Release Fan-Out in a Flagship Go Repo**: Every version bump requires editing 6 `<skill>/.claude-plugin/plugin.json` files + `.claude-plugin/marketplace.json` + `CHANGELOG.md` in lockstep, validated by 3 Python scripts (`pip install jsonschema` + live `urllib` fetches to `schemastore.org`) instead of a hermetic Go `go test` / `go run` check.
- **Strongest Existing Assets**: Deep, authoritative Go 1.26/1.27 coverage (`stdversion`, `os.Root`, `testing/synctest`, `b.Loop()`, `errors.AsType[T]`, `omitzero`) and well-structured existing Kata Epics (`j6sd`, `6c4t`, `d96p`, `g6k7`, `zr5c`).

---

## 2. Prioritized Recommendations Matrix

Ordered by the **Leverage Hierarchy** (`DELETE/COLLAPSE` → `FIX_FLOW` → `SIMPLIFY/ARCHITECT` → `FEATURE_GAP`):

| ID | Category | Recommendation | Lenses | Severity | Effort | Net Delta |
|---|---|---|---|---|---|---|
| **R-01** | `DELETE` | Replace 3 Python CI scripts & remote `schemastore.org` fetch with 1 hermetic Go validator (`cmd/check`) & trim 307-LOC empty `.roborev.toml` | ARCH, ENG | `HIGH` | `M` | `−450 LOC Py, +150 LOC Go` |
| **R-02** | `COLLAPSE` | Deduplicate `go-spec-reviewer` against `go` + `cobra-viper` and assign single rule owners across the 6 skills (Kata `6c4t`, `j6sd`) | PM, ARCH | `HIGH` | `M` | `−120 LOC duplicated rules` |
| **R-03** | `FIX_FLOW` | Fix non-compiling `slog.With` sample (`go:1135`), `[SPEC_FILE_PATH]` placeholder (`go-spec-reviewer:35`), & `cobra-viper` `main.go` silent exit (`j6sd`, `6c4t`) | UX, ENG | `FATAL` | `S` | `100% vet-clean samples` |
| **R-04** | `FIX_FLOW` | Reconcile `cmd/` layout (`go` vs `cobra-viper`), `os.Root` vs `pathologize.Join`, and module-floor qualification (`6c4t`) | PM, ENG | `HIGH` | `S` | `+35 LOC disambiguation` |
| **R-05** | `FIX_FLOW` | Add mutual negative trigger boundaries (`Not for...`) to all 6 `SKILL.md` and `plugin.json` descriptions (`g6k7`) | PM, UX | `FATAL` | `S` | `+24 lines across 12 files` |
| **R-06** | `SIMPLIFY` | Split `go/SKILL.md` (`1,189 lines`) and `fileflow-pathologize/SKILL.md` into ≤400-line core + `references/*.md` (`j6sd`) | UX, ARCH | `HIGH` | `M` | `1,189 → ~370 core lines` |
| **R-07** | `FEATURE_GAP` | Close high-priority content gaps (`go` race/sentinel/context rules, `fileflow` verification/ENOSPC, `wails` v3 pin, `README`/`CHANGELOG` claims) (`zr5c`, `d96p`) | PM | `MEDIUM` | `M` | `+180 LOC targeted guidance` |

---

## 3. Detailed Recommendations & Alignment with Existing Kata Epics

1. **[R-01] Replace Python CI Scripts with Hermetic Go Validator (`d96p` extension)**:
   - Port `.github/scripts/{check_invariants.py,check_go_release.py,validate_manifests.py}` to a zero-dependency Go checker (`internal/check` or `cmd/check`) that validates manifest invariants, `SKILL.md` line budgets (≤500 lines), negative trigger clauses, unresolved `[PLACEHOLDER]` tokens, and compiles/vets complete Go code blocks.
2. **[R-02] Deduplicate Cross-Skill Rules (`j6sd#9st8`, `j6sd#r395`)**:
   - Have `go-spec-reviewer/SKILL.md` reference `go/SKILL.md`'s Version → Feature Index and Necessity Ladder rather than maintaining a second copy of the 11-row obsolete-idiom table.
3. **[R-03] Fix Broken Code Samples & Template Placeholders (`j6sd#xgy4`, `j6sd#zqg7`, `6c4t#2p95`)**:
   - In `go/SKILL.md:1135`, replace `slog.With(...)` with `logger.With(...)` (or `slog.Default().With(...)`).
   - In `go-spec-reviewer/SKILL.md:19-35`, remove the broken subagent fork and `[SPEC_FILE_PATH]` placeholder.
   - In `cobra-viper/SKILL.md:74,271`, print `err` to `os.Stderr` before `os.Exit(1)` when `SilenceErrors: true` is set, and move `var cfg Config` inside `RunE`.
4. **[R-04] Reconcile Cross-Skill Architectural Rules (`6c4t#bk9y`, `6c4t#eqdh`, `6c4t#f18b`, `6c4t#s6rp`)**:
   - State the `main.go`-at-root vs `cmd/` rule once (`main.go` at root for single-binary services/tools without subcommands; `cmd/` routing layer for multi-command Cobra CLIs).
   - State when to use stdlib `os.Root` (Go 1.24+ OS-enforced directory handle) vs `pathologize.Join` (cross-OS untrusted filename sanitization).
5. **[R-05] Add Negative Trigger Boundaries Across All 6 Skills (`g6k7#92nx`, `g6k7#z6q6`, `g6k7#mkvx`)**:
   - Add explicit `Not for...` exclusions to `go/SKILL.md` and all 6 `.claude-plugin/plugin.json` descriptions so `go` does not swallow `cobra-viper`, `go-release`, `go-spec-reviewer`, `fileflow-pathologize`, or `wails` prompts.
6. **[R-06] Progressive Disclosure Split for `go/SKILL.md` & `fileflow-pathologize/SKILL.md` (`j6sd#p1wv`, `j6sd#6gtp`)**:
   - Split `go/SKILL.md` (`1,189 lines`) into a ~370-line core + `go/references/` files, and move `fileflow-pathologize`'s exhaustive API reference tables into `fileflow-pathologize/references/`.
7. **[R-07] Close Content & Currency Gaps (`zr5c`, `d96p#8y0w`)**:
   - Complete the open `zr5c` and `d96p` items (`README`/`CHANGELOG` claims, `go` race/context/sentinel-error coverage, `fileflow` destructive-move verification).
