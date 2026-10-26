# Topic Strategy: Uncovered Go Topics

- **Date**: 2026-10-26
- **Kata**: `1jwd` (parent epic `zr5c`)
- **Sources**: `docs/audits/2026-10-26-project-audit.md`; catalog evaluation §11 (coverage map), §12 (recommended structure), §13 (issue plan)
- **Status**: `DECIDED` — this document creates no plugin, no skill directory, and no marketplace entry.

## Purpose

The coverage map in the catalog evaluation (§11) names high-value Go topics with no home in the catalog. This document records one written decision per topic — **new skill**, **section**, **cross-reference**, or **defer** — so implementation issues can be filed against a settled plan instead of re-litigating ownership each time.

## Decision rules

Applied in this order; a topic takes the first form that holds.

1. **Section before skill.** New content goes into an existing skill unless it is a workflow or toolchain practice rather than the Go language, *and* it outgrows a section (~120 lines of rules), *and* its triggers are mutually exclusive with the six existing skills.
2. **One owner per rule** (spec invariant 4). A topic has exactly one canonical home; siblings cross-reference it and never restate it.
3. **`go` sections mean reference-tier files.** `go/SKILL.md` is split-pending (`j6sd#p1wv`): core ≤500 lines plus `go/references/*.md`. "Section in `go`" below means a new `go/references/*.md` file plus a pointer from core — never another 40 lines of core, and not a file created before the split lands.
4. **Name third-party surfaces, do not teach them.** OpenTelemetry, `sqlc`, `goose`, `protoc`, and golangci-lint change faster than this catalog. State the rule and the tool's role; leave the API to upstream docs.
5. **Decision now, build after the structural work.** The audit's §12 rule stands: no new skill is bundled with the restructure.

## Decisions

| # | Topic | Decision | Canonical home | Trigger boundary |
| --- | --- | --- | --- | --- |
| 1 | Profiling & performance | **New skill** — scoped in §1, built first | `go-perf/` (new) | `Not for` races/correctness bugs (`go`), releases (`go-release`), CLI flag design (`cobra-viper`) |
| 2 | `database/sql`, migrations, transactions | **New skill** — second, after `go-perf` ships | `go-database/` (new) | `Not for` schema/spec review (`go-spec-reviewer`) |
| 3 | Observability beyond `slog` | **Section** | `go/references/observability.md` | none — stays under `go`'s existing trigger |
| 4 | Static analysis & lint config | **Section** | `go/references/static-analysis.md` | none — stays under `go`'s existing trigger |
| 5 | Module & workspace management | **Section** | `go-release/SKILL.md`, cross-referenced from `go`'s version gating | none — `go-release` already owns `go.mod` hygiene |
| 6 | Dependency health & security | **Section** | `go-release/SKILL.md`, extending the existing `govulncheck` line | none — release-time concern |
| 7 | Code generation | **Section** | `go/references/codegen.md` | none — stays under `go`'s existing trigger |
| 8 | Post-implementation code review | **Cross-reference** | `pr-reviewer` (SteveSkills), pointed to from `go-spec-reviewer` and `go` | `go-spec-reviewer` keeps `Not for` post-implementation review |
| — | CLI design as a framework-agnostic discipline | **Cross-reference** | `cli-design` (SteveSkills), pointed to from `cobra-viper` | no duplication of `cli-design`'s 17 principles |

## 1. Profiling & performance — scoped first (top candidate)

Chosen as the first new skill: highest value in the coverage map, and the only topic that is both a toolchain workflow and large enough to justify its own budget. `go` today mentions `pprof` only for the `goroutineleak` profile and `b.Loop` only as a benchmark form — a foothold, not coverage.

### Identity

| Field | Value |
| --- | --- |
| Directory / `name` | `go-perf` |
| `displayName` | `Go Performance` |
| `category` | `development` |
| `version` | the catalog version at creation (currently `1.1.0`; one version covers every plugin) |
| Keywords | `go`, `golang`, `performance`, `profiling`, `pprof`, `benchmark`, `allocations` |

### Trigger description (draft, both halves required)

> Use when Go code is slow, allocates too much, leaks memory or goroutines, or must be measured: `pprof` CPU/heap/allocation/block/mutex/goroutine profiles, escape analysis, `go test -benchmem`, `benchstat` comparisons, `go tool trace`, and production `runtime/metrics`. Not for correctness bugs, data races, or goroutine lifetimes as a design question (use `go`); not for release versioning or `go.mod` hygiene (use `go-release`); not for CLI flag or exit-code design (use `cobra-viper`).

Carries `current through Go 1.27` because it teaches versioned runtime features (the `goroutineleak` profile). That means the skill is added to `CLAIMING_SKILLS` in `.github/scripts/check_go_release.py` in the same change — an unregistered claim is not currency-checked, and a registered claim may never lose the phrase.

### Sections, and what each one owns

| Section | Owns | Points at (never restates) |
| --- | --- | --- |
| Measure first | The loop: benchmark → profile → one change → `benchstat` on repeated runs. Never optimize from reading code | — |
| Benchmark hygiene | `-benchmem`, `-count=10`, `benchstat` A/B, `-cpuprofile` while benchmarking | `b.Loop` and table-driven benchmarks: `go` |
| CPU profiles | `go test -cpuprofile`, guarding `net/http/pprof` to an internal mux, `go tool pprof -http`, reading `top`/`list`/flame output | Server timeouts and graceful shutdown: `go` |
| Allocation & heap | `-memprofile`, `-alloc_space` vs `-inuse_space`, `go build -gcflags=-m` escape output, `testing.AllocsPerRun` for micro-asserts, the escape myths (returning a local pointer is fine; interface boxing and `string`↔`[]byte` conversions are not) | — |
| Goroutine & blocking | `goroutineleak` profile (GA in Go 1.27), `-blockprofile`/`-mutexprofile` with `runtime.SetBlockProfileRate` and `SetMutexProfileFraction` | "Never start a goroutine without knowing how it stops": `go` |
| Per-request attribution | `pprof.Do` and `pprof.Labels`, including the limit that labels affect only the CPU and goroutine profiles | Logger/context plumbing: `go` |
| Execution trace | `go tool trace` for scheduling, GC pauses, and latency attribution — and when pprof is the wrong tool | — |
| Production numbers | Percentiles over averages; `runtime/metrics` as the in-process hook | Metrics/tracing policy: `go/references/observability.md` |
| Anti-patterns | Micro-optimizing unmeasured code, `GOMAXPROCS`/`GOGC` knob-first tuning, `unsafe` tricks, benchmarks with unrealistic payloads | — |

### Non-goals

- No benchmark framework, harness, or CI performance gate design.
- No `unsafe`/`//go:nosplit`/linkname recipes.
- No restatement of `go`'s ownership: `b.Loop`, race detector usage, goroutine lifetime rules, and HTTP server tuning stay in `go`.

### Definition of done for the implementation issue

`SKILL.md` (≤500-line core; overflow to `go-perf/references/`), `.claude-plugin/plugin.json`, a four-key `marketplace.json` entry, a `README.md` table row, negative boundaries in both descriptions, the currency claim plus `CLAIMING_SKILLS` registration, and every Go sample compiling and vetting clean.

### Deliberately not created here

The `go-perf` directory is not created by this change. `j6sd#p1wv` (the `go` split) lands first so the new skill is authored against a settled core-plus-references convention rather than inventing a second one.

## 2. The remaining topics

**2 — `database/sql`, migrations, transactions: new skill (`go-database`), queued second.** High value and no home today. Kept honest by rule 4: the skill owns rules, not driver APIs — connection-pool sizing (`SetMaxOpenConns`, `SetMaxIdleConns`, `SetConnMaxLifetime`), `context` on every call and `QueryContext`/`ExecContext`, transactions with `BeginTx` and a deferred rollback, `sql.ErrNoRows` through `errors.Is`, `sql.Null*` versus pointer fields, prepared statements and when a generator (`sqlc`) beats them, migrations as ordered SQL files embedded with `go:embed` plus one runner (`goose` named, not taught), expand/contract discipline for zero-downtime changes, and ephemeral-database tests with `t.Cleanup`. Reassess before building: if it stabilizes at stdlib-only rules under ~120 lines, rule 1 collapses it into `go/references/database.md` and the skill is deleted from the plan.

**3 — Observability beyond `slog`: section in `go`.** `slog` is the only current foothold, and the missing piece is policy, not API: correlation IDs through `context`, the decision rule between logs, metrics, and traces, metric naming and cardinality, redaction, and `expvar`/`runtime/metrics` for cheap in-process counters. OpenTelemetry is named as the default vendor-neutral choice and left to its own docs. One reference file, not a seventh skill — the topic does not outgrow a section.

**4 — Static analysis & lint config: section in `go`.** Toolchain configuration, not language teaching. Covers `go vet` already running inside `go test`, staticcheck as the additional pass, a minimal `golangci-lint` enable list, `//nolint` requiring a stated reason, and formatting as non-negotiable. `stdversion` is already documented in `go` and is cross-referenced, not restated; `go fix` stays owned by `go-release`.

**5 — Module & workspace management: section in `go-release`.** Release engineering already owns `go.mod` hygiene, the `replace`-never-tagged rule, and `go mod tidy` behavior; the gap is boundaries and workspaces: when to split a module (independent versioning or distinct consumers, not directory count), `go.work` as a local-only convenience never assumed by a tag, how minimal version selection actually picks, committing `go.mod`/`go.sum` output, `retract` for a bad release, and the `go` directive as a compatibility promise. Cross-referenced from `go`'s version-gating section so the floor rule stays single-owned.

**6 — Dependency health & security: section in `go-release`.** Extends, rather than adds to, the existing `govulncheck` line: reachability-aware triage of findings, `govulncheck` in CI and on a schedule, `go version -m` to inventory the dependencies of a shipped binary, module checksum and `GOPRIVATE`/`GOFLAGS=-mod=readonly` hygiene, and a licence inventory step. Dependency admission stays with the Necessity Ladder in `go` (stdlib first).

**7 — Code generation: section in `go`.** Small, stable, and mostly convention: `//go:generate` in a build-tag-guarded `generate.go`, generated paths relative to the directive's file, reproducible tool invocation via the `tool` directive or `go run tool@vX.Y.Z`, committing generated output so builds need no generator, a regenerate-and-diff CI gate, the `// Code generated … DO NOT EDIT.` header so tooling skips the file, and never hand-editing generated code. `protoc`, `stringer`, and mock generators are named and pointed at. The `tool` directive and `//go:embed` are already owned by `go` and `go-spec-reviewer` respectively — cross-reference only.

**8 — Post-implementation code review: cross-reference, no new skill.** The Go-specific rules for reviewing finished code are the body of `go/SKILL.md`; a third review skill would re-derive them and violate rule 2. `go-spec-reviewer` stays the spec-time reviewer (`Not for` post-implementation review), `pr-reviewer` in the SteveSkills catalog carries the maintainer-style PR verdict, and `go` supplies the rules. The whole change is a pointer in `go-spec-reviewer` plus one line in `go`'s activation list.

**CLI design: cross-reference, no duplication.** `cli-design` (SteveSkills) already teaches framework-agnostic CLI and agent-experience design — 17 principles, exit codes, stream segregation, `--preview`. `cobra-viper` keeps Cobra/Viper wiring and points at `cli-design` for the discipline itself; it must not restate it.

**Recorded but outside this issue.** gRPC/RPC patterns and AI-agent-facing Go patterns stay deferred per §11. HTTP service/API design beyond routing carries §11's disposition — a section in `go` — and is not one of the eight topics decided here.

## 3. Sequencing

1. **Now** — decisions recorded in this file; no plugin, directory, or manifest touched.
2. **After `go/SKILL.md` is split** (`j6sd#p1wv`) — topics 3, 4, and 7 become `go/references/*.md` with core pointers; topics 5 and 6 land in `go-release`; the `cli-design` and `pr-reviewer` cross-references ship with the trigger-boundary work (`g6k7`), where descriptions are already being edited.
3. **Next release** — `go-perf`, authored to §1 of this document.
4. **Release after that** — `go-database`, after re-checking rule 1 against whatever `go-perf` cost.

## 4. What would reverse a decision

- `go-perf` needs more cross-references than rules → collapse it into `go/references/performance.md` and record that here.
- `database/sql` guidance turns out stdlib-only and small → collapse topic 2 as described above.
- `pr-reviewer` grows Go-specific rules → revisit topic 8's owner rather than letting two skills hold them.
- A new upstream that owns a topic authoritatively (for example a maintained Go performance skill) → cross-reference instead of shipping a competing copy.

## Acceptance mapping

- Written decision per topic, including the CLI-design cross-reference: `## Decisions`.
- Top candidate scoped first: `## 1. Profiling & performance`, ahead of every other topic.
- No new plugin before the decision: nothing outside this file is created by this change.
