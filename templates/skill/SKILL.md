---
name: go-profiling
description: >
  Use when a Go program is slow, allocation-heavy, or suspected of leaking — pprof CPU, heap,
  block, mutex and goroutine profiles, benchmark comparison with -benchmem, escape analysis via
  -gcflags=-m, and cutting allocation counts. Not for correctness bugs; use the go skill.
---

<!--
  This file is the scaffold, written as a worked example rather than a placeholder. Copy the
  directory, then replace every fact in it — the name, the description, and the rules — with your
  skill's. The description is the part that sets the shape: triggers only, third person, no
  credentials, no coercive wording, no workflow summary, under 1024 characters. CONTRIBUTING.md
  § SKILL.md is the standard.
-->

# go-profiling

## When this applies

- A benchmark regressed and you need to know whether the cause is allocations, GC, lock contention, or call overhead.
- A service's RSS climbs while heap profiles stay flat — a leak in goroutines, or memory retained off-heap.
- A hot path allocates inside a loop and you want the escape analysis answer before restructuring it.

## When NOT to Use This Skill

For correctness bugs, wrong output, or a failing test use `go`. For CLI flag and command design use `cobra-viper`. For cutting a release or versioning the module use `go-release`. Name the skills this one defers to — a skill with no negative trigger claims prompts its siblings own.

## Rules

- Measure before you change anything. `go test -bench . -benchmem` for the unit of work; `go tool pprof -http=:8080` over `-cpuprofile` / `-memprofile` output. A change with no before-and-after number is a guess.
- Read the profile in the right units. `-sample_index=alloc_objects` counts allocations, `alloc_space` counts bytes; the default `inuse_space` answers a different question than the one usually asked.
- Serve profiles from a long-running process with `net/http/pprof` on a private listener, never the public mux — the endpoints have no authentication, and a CPU profile is 30 seconds of `SIGPROF`.
- Confirm an escape with `go build -gcflags=-m` against the one package that changed. The analysis is per-package, and the compiler's `moved to heap` is the only answer that matters.
- Compare runs with `benchstat`, not by eyeballing `-bench` output. One run cannot tell a real 3% win from noise.
- Stop when the profile shows the remaining cost is proportional to work the program has to do. Optimization without a target is rework.

## Anti-patterns

- Optimizing from code reading. Profiles routinely disagree with intuition about which function is hot.
- Reaching for `sync.Pool` first. It trades an allocation for contention and retention, and loses on a low-churn path.
- Writing a benchmark without `-benchmem`. Allocations are invisible without it, and they are usually the answer.
- Leaving `net/http/pprof` mounted on the public mux "temporarily".
