---
name: go-spec-reviewer
description: >
  Use when reviewing a Go design document, spec, RFC, or PRD before implementation. Use for spec
  completeness, idiom checks, and Cobra/Viper convention validation at the design stage. Not for
  writing Go code (use go), building CLIs (use cobra-viper), or release decisions (use go-release).
---

# Go Spec Reviewer

## Purpose

Verify that a Go design document is complete, consistent, and idiomatic **before implementation begins**. The reviewer channels the perspective of Rob Pike, the Go standard library authors, and spf13 — people who would reject unnecessary abstractions, demand explicit error handling, and expect the simplest design that actually works.

## When NOT to Use This Skill

For writing or refactoring Go code use `go`. For CLI construction use `cobra-viper`. For release decisions use `go-release`. This skill reviews designs before implementation — it does not write the code.

## How to Run the Review

Run the review inline: read the spec file the user provides, explore the codebase, then work through the review instructions below and report in the Output Format at the end. The reviewer needs the codebase context open while reasoning about the spec, so there is nothing to hand off.

## Review Instructions

You are a Go spec reviewer. Your job is to verify this spec is complete and ready for implementation planning, viewed through the lens of idiomatic Go.

Think like Rob Pike reviewing this design: is it simple? Does it do one thing well?
Think like the stdlib authors: are interfaces small and defined at the point of use?
Think like spf13: if this is a CLI, does it follow Cobra/Viper conventions properly?

**Spec to review:** the spec file the user supplied (if none was named, ask for the path).

### Step 0 — Load the Standards

If the `go` and `cobra-viper` skills are installed, load and read them first. They are the source of truth for what "idiomatic" means in this review — package organization, stdlib-first, factory-built commands, testing patterns. Do not re-derive standards that conflict with them.

### Step 1 — Understand the Codebase Context

Before reviewing, explore the existing codebase to understand conventions and spot conflicts. At minimum:

- Map the package structure. Note whether it follows flat/domain-package organization or legacy layouts (deep `internal/` trees, layer packages like `service/`/`repository/`). The spec should match the better of the two — and may reasonably propose migrating away from a legacy layout.
- If this is a Cobra CLI, check how commands are constructed. Factory functions (`NewRootCmd()`) are the standard. If the codebase uses package-level `var` command declarations, note it: (a) the spec should not add more globals, and (b) if it must match the legacy pattern, new flag variable names must be unique across all `cmd/` files, since they share one package.
- Check how commands are registered (in the root factory or `cmd/root.go`) — new subcommands must appear in the spec's file list with their registration point.
- Identify existing patterns (HTTP clients, error types, config structs, interfaces) the spec should reuse rather than reinvent.
- Note the Go version in `go.mod` — the module's `go` directive and the toolchain. The `go` skill owns the rule and the data ("Version Gating: Check the Module's Floor First", its Version → Feature Index, and the `stdversion` enforcement) — apply it in both directions, flagging idioms the module's floor cannot compile *and* idioms that remain correct at that floor. Two obligations follow at spec level: name the introducing version whenever the review recommends a gated feature, and require the spec to justify any raise of its module's `go` directive by what it enables, since that floor binds every dependent of the module.

### Step 2 — Go Philosophy Check

| Concern | What to Look For |
|---------|-----------------|
| Simplicity | Unnecessary layers, abstractions with only one implementation, over-engineered designs |
| Dependencies | Every proposed third-party dependency justified against a stdlib alternative? (routing, slices/maps helpers, multierror, etc.) |
| Platform & toolchain | A dependency, background service, or hand-rolled mechanism doing what the compiler, runtime, toolchain, or platform already does? `//go:embed` over an asset pipeline, `os.Root` over path confinement, build tags or the `tool` directive over scripts, a database constraint over application code that re-implements it. The full necessity ladder is in the `go` skill — apply it at spec level, since a spec is where an over-build is cheapest to reject |
| Interfaces | Defined by the consumer, not the implementor? Small (1–3 methods)? Used where polymorphism is actually needed? |
| Error handling | Errors returned explicitly? Properly wrapped with `fmt.Errorf("%w", err)`? Not swallowed silently? Sentinel errors named where callers must branch? |
| Context | `context.Context` threaded through any I/O, HTTP, or long-running calls? Timeouts specified? |
| Concurrency | Goroutines with clear ownership **and a stated shutdown path**? Bounded concurrency where fan-out is unbounded input? Data races avoided? |
| Package design | Each package has a single clear domain responsibility? New packages justified vs. extending existing ones? No `utils/`/`common/` proposals? |
| Naming | Short, clear names following Go conventions? No stutter (`pkg.PkgThing`)? No `Get` prefixes? |
| Testing | Does the spec say how this will be tested? Fakes/interfaces at I/O boundaries, table-driven cases for core logic, in-memory execution for CLI commands? |
| YAGNI | Features and abstractions driven by stated requirements only — not anticipated future needs? |

### Step 2b — Modern Language Feature Check

Run this before approving a design that introduces interfaces, generics, or function-typed plumbing. A spec proposing a design the Go 1.26/1.27 language features make simpler is an Issue — the implementor will either build the workaround or silently simplify and drift from the spec.

| Feature | Introduced | What to Look For |
|---------|------------|------------------|
| `new(expr)` | 1.26 | A helper or temporary variable where `new(expr)` expresses the value directly |
| Self-referential type-parameter constraints | 1.26 | A spec asserting the self-reference is illegal, or carrying a workaround helper for it — the ban was lifted, `type Adder[A Adder[A]] interface { Add(A) A }` is legal |
| Generic methods | 1.27 | A method-heavy design that copies a generic package-level helper instead of declaring type parameters on the method, as `math/rand/v2` now does with `(*Rand).N` alongside the package-level `N`. Two hard limits to check against the spec: methods of interfaces may not declare type parameters, and interface methods cannot be implemented by generic methods — a design relying on either will not compile |
| Generalized function type inference | 1.27 | Explicit type arguments that inference now supplies — it applies wherever a generic function is assigned to, or converted to, a variable of a matching function type |
| Struct-literal field selectors | 1.27 | A spec insisting a struct-literal key be a top-level field name of the struct type; any valid field selector for the struct type is now allowed |

### Step 3 — Cobra/Viper CLI Check (skip if not a CLI)

| Concern | What to Look For |
|---------|-----------------|
| Command construction | New commands built via factory functions, not package-level `var`s? (Legacy-globals codebase: see Step 1 flag-name caveat) |
| Command registration | New subcommands registered with the parent — is this in the spec's file list? |
| RunE vs Run | `RunE` so errors propagate; `Run` silently swallows them |
| Persistent vs local flags | Cross-cutting flags on the root as persistent; operation-specific flags on the subcommand |
| Args validation | Positional argument counts handled via `cobra.Args` validators, not inside `RunE`? |
| Viper binding | Env var names and config keys explicitly bound? Defaults set for every key in the config struct (the `AutomaticEnv` + `Unmarshal` gotcha)? Business logic receives typed config structs, never Viper itself? |
| Output | Command output written via `cmd.OutOrStdout()` so it's testable? |
| I/O path collision | If the command takes `--input` and `--output`, is there a guard preventing them from resolving to the same path? |

### Step 4 — General Spec Completeness

| Category | What to Look For |
|----------|-----------------|
| Completeness | TODOs, placeholders, "TBD", incomplete sections, missing error paths |
| Consistency | Internal contradictions, conflicting requirements, types named differently in different sections |
| Clarity | Requirements ambiguous enough that two implementors would build different things |
| Scope | Focused enough for a single implementation plan? Not covering multiple independent subsystems? |
| Data flow | Clear what enters and exits each function or step? |
| Compatibility | If this changes existing behavior, flags, config keys, or APIs — is the migration/deprecation story stated? |
| Security | User-supplied values sanitized before use in shell commands, file paths, or external calls? |
| Observability | If this runs as a daemon, service, or scheduled job, does the spec name what an operator watches to tell working from stuck — the log fields on the key paths, the counter or health signal, the error text? A library or pure helper needs none; a long-running process with no stated way to observe it is an Issue |
| Schema & data migration | If persisted state changes shape (DB schema, config keys, on-disk files, cache format), is the migration stated — forward step, default or backfill for existing data, and whether old and new versions can read each other's writes during a rolling swap? |
| Rollout & rollback | Is there a way back if the new path misbehaves — flag to disable it, reversible migration, old code still reading new data? A one-way door with no rollback story is an Issue |
| Failure modes | For each external dependency and I/O path, does the spec say what happens on failure — retry or fail fast, timeout value, what partial state is left behind, and what the caller receives (`%w`-wrapped error, sentinel, degraded result)? |
| Performance targets | Where the spec claims a scale (requests/sec, item counts, batch or response sizes), is there a number, and does the design meet it — bounded concurrency, pagination, streaming instead of loading everything? "Should be fast" over an unbounded fan-out is an Issue |
| Test-plan realism | Does the plan exercise the paths the spec itself calls hard — error branches, concurrency limits, the migration — or only the happy path? A plan that would still pass with the feature removed is an Issue |

### Calibration

**Only flag issues that would cause real problems during implementation.**
A missing error path, a flag naming conflict that won't compile, an abstraction that adds complexity without enabling anything — those are Issues. Minor wording preferences, formatting inconsistencies, and "could be more detailed" are not.

Each Step 4 row beyond Security is conditional on the spec's own surface. No persisted state means no migration row applies; nothing long-running means no observability row applies; no scale claim means no performance row applies. Apply a row only where the spec touches that category, and flag it only when the design is silent **and** the implementor would have to invent the answer. Silence on a category the spec does not touch is not an Issue — it is the correct answer.

Ambiguity is not automatically an Issue. If a requirement is unclear but the spec is otherwise sound, raise it as a Question for the author rather than blocking.

Approve unless there are gaps that would lead to a flawed or incomplete implementation.

### Output Format

## Go Spec Review

**Status:** Approved | Approved with Questions | Issues Found

**Issues (block implementation):**
- [Section X]: [specific issue] — [why it matters for implementation]

**Questions for Author (need answers, don't block a sound design):**
- [Section Y]: [the ambiguity] — [the two+ readings an implementor could take]

**Recommendations (advisory):**
- [suggestions that improve correctness, idiomaticity, or clarity]