---
name: go-spec-reviewer
description: >
  Review design specification documents for Go programs before implementation begins. Use this
  skill when a user has a Go spec, design doc, RFC, or PRD to review, wants feedback on a planned
  feature in a Go codebase, is about to start implementing from a spec, or asks "is this spec
  ready?" or "review this before I build it". Applies Go philosophy — simplicity, composition,
  explicit errors, context propagation — plus Cobra/Viper CLI conventions where applicable.
---

# Go Spec Reviewer

## Purpose

Verify that a Go design document is complete, consistent, and idiomatic **before implementation begins**. The reviewer channels the perspective of Rob Pike, the Go standard library authors, and spf13 — people who would reject unnecessary abstractions, demand explicit error handling, and expect the simplest design that actually works.

## How to Run the Review

**If a subagent/Task tool is available:** dispatch a general-purpose subagent with the review instructions below (description: "Review Go spec document"), passing the spec file path.

**If not (Claude.ai, or no Task tool):** perform the review yourself, inline, following the same steps.

Either way, the review procedure is identical:

---

## Review Instructions

You are a Go spec reviewer. Your job is to verify this spec is complete and ready for implementation planning, viewed through the lens of idiomatic Go.

Think like Rob Pike reviewing this design: is it simple? Does it do one thing well?
Think like the stdlib authors: are interfaces small and defined at the point of use?
Think like spf13: if this is a CLI, does it follow Cobra/Viper conventions properly?

**Spec to review:** [SPEC_FILE_PATH]

### Step 0 — Load the Standards

If the `go` and `cobra-viper` skills are installed (check `/mnt/skills/user/go/SKILL.md` and `/mnt/skills/user/cobra-viper/SKILL.md`), read them first. They are the source of truth for what "idiomatic" means in this review — package organization, stdlib-first, factory-built commands, testing patterns. Do not re-derive standards that conflict with them.

### Step 1 — Understand the Codebase Context

Before reviewing, explore the existing codebase to understand conventions and spot conflicts. At minimum:

- Map the package structure. Note whether it follows flat/domain-package organization or legacy layouts (deep `internal/` trees, layer packages like `service/`/`repository/`). The spec should match the better of the two — and may reasonably propose migrating away from a legacy layout.
- If this is a Cobra CLI, check how commands are constructed. Factory functions (`NewRootCmd()`) are the standard. If the codebase uses package-level `var` command declarations, note it: (a) the spec should not add more globals, and (b) if it must match the legacy pattern, new flag variable names must be unique across all `cmd/` files, since they share one package.
- Check how commands are registered (in the root factory or `cmd/root.go`) — new subcommands must appear in the spec's file list with their registration point.
- Identify existing patterns (HTTP clients, error types, config structs, interfaces) the spec should reuse rather than reinvent.
- Note the Go version in `go.mod` — the module's own `go` directive and the toolchain. The floor rule and every feature's introducing release are stated once, in the `go` skill's *Version Gating* section and *The Version → Feature Index*; apply them, do not restate them. Flag specs that propose idioms the toolchain has since obsoleted:
  - Third-party routers where the `net/http.ServeMux` method and wildcard patterns suffice.
  - `interface{}` where `any` applies.
  - Manual loops that `slices`/`maps` handle.
  - Hand-rolled worker pools where `sync.WaitGroup.Go` applies.
  - Hand-rolled `v := expr; p := &v` where `new(expr)` applies — canonical for an optional pointer struct field, `Age: new(yearsSince(born))`.
  - `errors.As` where `errors.AsType[T]` applies — generic, type-safe, and faster.
  - `LastIndex`-based splitting where `strings.CutLast` or `bytes.CutLast` applies — a slice around the last occurrence of a separator.
  - `b.N` benchmark loops. `B.Loop` stopped inhibiting inlining of the loop body, so migrating `b.N` to it is unconditionally safe; parameters, results, and variables assigned inside the loop are still kept alive.
  - Third-party UUID libraries where the stdlib `uuid` package applies.
  - `encoding/json` v1 defaults where the spec targets `go 1.27` or later and v2 semantics are wanted — `encoding/json` is now backed by the v2 implementation, which rejects invalid UTF-8 in JSON strings and duplicate names within an object; v1 behavior is preserved via `Options`.

#### Version Gate — Check Both Directions

Apply the `go` skill's *Version Gating: Check the Module's Floor First* section — the `go` directive is the floor, availability is looked up in *The Version → Feature Index*, and *Enforcement: `stdversion`* is the mechanical backstop. That section owns the rule; this one enforces it.

Two reviewer-only obligations:

- **Require the spec to state its floor.** A spec that raises the module's `go` directive must record the decision and the feature that forced it, because the bump is a compatibility statement about every dependent.
- **State the introducing release when you recommend a feature.** A spec-level recommendation is only actionable if it says which floor the feature needs and what the spec must do when the module is below it.

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

### Calibration

**Only flag issues that would cause real problems during implementation.**
A missing error path, a flag naming conflict that won't compile, an abstraction that adds complexity without enabling anything — those are Issues. Minor wording preferences, formatting inconsistencies, and "could be more detailed" are not.

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