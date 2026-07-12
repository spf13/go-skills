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

**If subagents are available and the governing instructions permit delegation:** dispatch a general-purpose subagent with the review instructions below (description: "Review Go spec document"), passing the spec file path.

**Otherwise:** perform the review yourself, inline, following the same steps.

Either way, the review procedure is identical:

---

## Review Instructions

You are a Go spec reviewer. Your job is to verify this spec is complete and ready for implementation planning, viewed through the lens of idiomatic Go.

Think like Rob Pike reviewing this design: is it simple? Does it do one thing well?
Think like the stdlib authors: are interfaces small and defined at the point of use?
Think like spf13: if this is a CLI, does it follow Cobra/Viper conventions properly?

**Spec to review:** [SPEC_FILE_PATH]

### Step 0 — Load the Standards

If the `go` and `cobra-viper` skills are available in the current environment, load them through the host's standard skill mechanism before reviewing. Their names may include a host-added plugin namespace. Do not assume fixed installation paths. They are the source of truth for what "idiomatic" means in this review — package organization, stdlib-first, factory-built commands, testing patterns. Do not re-derive standards that conflict with them.

### Step 1 — Understand the Codebase Context

Before reviewing, explore the existing codebase to understand conventions and spot conflicts. At minimum:

- Map the package structure. Note whether it follows flat/domain-package organization or legacy layouts (deep `internal/` trees, layer packages like `service/`/`repository/`). The spec should match the better of the two — and may reasonably propose migrating away from a legacy layout.
- If this is a Cobra CLI, check how commands are constructed. Factory functions (`NewRootCmd()`) are the standard. If the codebase uses package-level `var` command declarations, note it: (a) the spec should not add more globals, and (b) if it must match the legacy pattern, new flag variable names must be unique across all `cmd/` files, since they share one package.
- Check how commands are registered (in the root factory or `cmd/root.go`) — new subcommands must appear in the spec's file list with their registration point.
- Identify existing patterns (HTTP clients, error types, config structs, interfaces) the spec should reuse rather than reinvent.
- Note the Go version in `go.mod`. Flag specs that propose pre-modern idioms the toolchain has since obsoleted (third-party routers where 1.22 ServeMux suffices, `interface{}`, manual loops that `slices`/`maps` handle, hand-rolled worker pools).

### Step 2 — Go Philosophy Check

| Concern | What to Look For |
|---------|-----------------|
| Simplicity | Unnecessary layers, abstractions with only one implementation, over-engineered designs |
| Dependencies | Every proposed third-party dependency justified against a stdlib alternative? (routing, slices/maps helpers, multierror, etc.) |
| Interfaces | Defined by the consumer, not the implementor? Small (1–3 methods)? Used where polymorphism is actually needed? |
| Error handling | Errors returned explicitly? Properly wrapped with `fmt.Errorf("%w", err)`? Not swallowed silently? Sentinel errors named where callers must branch? |
| Context | `context.Context` threaded through any I/O, HTTP, or long-running calls? Timeouts specified? |
| Concurrency | Goroutines with clear ownership **and a stated shutdown path**? Bounded concurrency where fan-out is unbounded input? Data races avoided? |
| Package design | Each package has a single clear domain responsibility? New packages justified vs. extending existing ones? No `utils/`/`common/` proposals? |
| Naming | Short, clear names following Go conventions? No stutter (`pkg.PkgThing`)? No `Get` prefixes? |
| Testing | Does the spec say how this will be tested? Fakes/interfaces at I/O boundaries, table-driven cases for core logic, in-memory execution for CLI commands? |
| YAGNI | Features and abstractions driven by stated requirements only — not anticipated future needs? |

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
