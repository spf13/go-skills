---
name: go
description: >
  Expert Go programming skill authored by spf13 (former Go team lead, author of Cobra, Viper, Hugo, Afero).
  Covers idiomatic Go — package design, error handling, interfaces, concurrency, testing, and
  project layout, current through Go 1.27. Use whenever Go code is written, reviewed, debugged, or
  refactored — any .go file, go.mod, CLI tool, or web service, and whenever the user mentions
  Go or golang, even if they don't ask for "idiomatic" code.
---

# Idiomatic Go: The Go Way

Idiomatic Go patterns and best practices for building robust, efficient, and maintainable applications.

## When to Activate

- Writing new Go code
- Reviewing or auditing existing Go code
- Refactoring Go code (especially code that looks like Java/Spring Boot patterns in Go)
- Designing Go packages, modules, or APIs
- Choosing between stdlib and third-party libraries
- Any question about Go project structure, error handling, concurrency, or testing

## Version Gating: Check the Module's Floor First

Every version-gated feature is a floor: the module's packages must compile on the release that introduced it. Before recommending such a feature, establish the floor and recommend only what it supports.

1. Read the `go` directive in the module's `go.mod`. That version is the floor for the module's own packages — a module declaring `go 1.21` must still compile on Go 1.21, whatever toolchain you happen to be running.
2. Check the file's build constraints. A file behind `//go:build go1.24` may use Go 1.24 features even when the module's `go` directive is older.
3. If the floor is unknown, read `go.mod` or ask. Never guess, and never assume the reader is on the latest release.

The inverse error is just as real: do not flag an idiom as unavailable in a module whose floor already permits it, and never raise a module's `go` directive merely to reach a convenience helper — a `go` directive is a floor imposed on every dependent of the module.

### The Version → Feature Index

One line per feature, so availability is looked up here rather than guessed.

| Since | Feature |
| --- | --- |
| 1.17 | `//go:build` constraints, required in place of `// +build` lines |
| 1.18 | Type parameters (generics) |
| 1.18 | `any` as an alias for `interface{}` |
| 1.19 | `sync/atomic` typed values (`atomic.Int64`, `atomic.Pointer[T]`) |
| 1.20 | `errors.Join` |
| 1.21 | `slices` package |
| 1.21 | `maps` package |
| 1.21 | `cmp` package (`cmp.Compare`) |
| 1.21 | `min`, `max`, `clear` builtins |
| 1.21 | `log/slog` |
| 1.21 | `context.WithoutCancel` |
| 1.22 | `math/rand/v2` |
| 1.22 | `net/http` ServeMux method and wildcard patterns |
| 1.22 | `for i := range n` (range over integer) |
| 1.22 | Per-iteration loop variables |
| 1.22 | `slices.Concat` |
| 1.22 | `cmp.Or` |
| 1.23 | Range-over-func iterators (`iter.Seq`, `iter.Seq2`) |
| 1.23 | `maps.Keys` / `maps.Values` as iterators |
| 1.23 | `slices.Collect` / `slices.Values` / `slices.Sorted` |
| 1.24 | `encoding/json` `omitzero` |
| 1.24 | `tool` directive in `go.mod` |
| 1.24 | Generic type aliases |
| 1.24 | `t.Context()`, `b.Loop()`, `t.Chdir()` |
| 1.25 | `testing/synctest` (`synctest.Test`) |
| 1.25 | `sync.WaitGroup.Go` |
| 1.26 | `new(expr)` |
| 1.26 | Self-referential type-parameter constraints |
| 1.26 | `errors.AsType[T]` |
| 1.26 | `log/slog.NewMultiHandler` |
| 1.26 | `reflect` iterators (`Type.Fields`, `Type.Methods`, `Type.Ins`, `Type.Outs`, `Value.Fields`, `Value.Methods`) |
| 1.26 | `t.ArtifactDir()` / `b.ArtifactDir()` / `f.ArtifactDir()` with `go test -artifacts` |
| 1.26 | `B.Loop` no longer inhibits inlining |
| 1.26 | `testing/cryptotest.SetGlobalRandom` |
| 1.26 | Green Tea GC on by default |
| 1.26 | Heap base address randomization on 64-bit platforms |
| 1.26 | `runtime/pprof` `goroutineleak` profile (experiment: `GOEXPERIMENT=goroutineleakprofile`) |
| 1.26 | `simd/archsimd` (experiment: `GOEXPERIMENT=simd`) |
| 1.26 | `runtime/secret` (experiment: `GOEXPERIMENT=runtimesecret`) |
| 1.27 | Generic methods |
| 1.27 | Struct literal keys may be any field selector |
| 1.27 | Generalized function type inference |
| 1.27 | `stdversion` vet check on by default in `go test` |
| 1.27 | `uuid` package |
| 1.27 | `strings.CutLast` / `bytes.CutLast` |
| 1.27 | `net/url` `URL.Clone` / `Values.Clone` |
| 1.27 | `math/rand/v2` `(*Rand).N` |
| 1.27 | `testing/synctest.Sleep` |
| 1.27 | `net/http/httptest.NewTestServer` |
| 1.27 | `runtime/pprof` `goroutineleak` profile generally available |
| 1.27 | `time` timers and tickers always unbuffered; `asynctimerchan` removed |
| 1.27 | Size-specialized small allocations (below 80 bytes) |
| 1.27 | `encoding/json` backed by the `encoding/json/v2` implementation |
| 1.27 | `simd` (experiment: `GOEXPERIMENT=simd`) |
| 1.27 | `simd/archsimd` on arm64 and WebAssembly (experiment: `GOEXPERIMENT=simd`) |

### Enforcement: `stdversion`

Since Go 1.27, `go test` runs the `stdversion` vet check by default. It reports the use of standard library symbols that are too new for the Go version in force in the referring file, as determined by the `go` directive in `go.mod` and the build tags on the file.

- Run `go vet ./...` (or `go test ./...`) before calling version-gated code correct.
- Treat a `stdversion` finding as a version-gating error, not a false positive: either the code is too new for the module's floor, or the floor is deliberately being raised and every dependent of the module must be told.
- It covers standard library symbols only. Language features (`new(expr)`, generic methods) and behavior changes that introduce no new symbol are still yours to gate by hand.

## Core Principles

### 1. Clear is Better than Clever

Go favors readability and simplicity over abstraction and cleverness. Code should be obvious. If you have to read a function three times to understand its control flow, it needs to be rewritten.

```go
// Idiomatic: Direct, linear control flow. Note: no "Get" prefix — Go omits it.
func LookupUser(id string) (*User, error) {
    user, err := db.FindUser(id)
    if err != nil {
        return nil, fmt.Errorf("finding user %s: %w", id, err)
    }
    return user, nil
}
```

### 2. Make the Zero Value Useful

Design types so their zero value is immediately usable without initialization. This eliminates boilerplate constructors. `sync.Mutex` and `bytes.Buffer` are the gold standard for this.

```go
// Idiomatic: Ready to use immediately
type Counter struct {
    mu    sync.Mutex
    count int
}

func (c *Counter) Inc() {
    c.mu.Lock()
    c.count++
    c.mu.Unlock()
}
```

### 3. Return Early, Keep the Happy Path Left

Handle errors and edge cases immediately and return. Do not use `else` blocks for the main logic. The "happy path" of your function should never be indented.

## Write Less: The Necessity Ladder

Before writing code, stop at the first rung that holds. Climb **after** you understand the problem — read the code the change touches and trace the real flow first. If two rungs both work, take the higher one and move on.

1. **Does this need to exist?** "We'll want it later" is not a need. Skip it, and say so in one line.
2. **Does the codebase already have it?** Search before you write — the helper, type, or pattern may already live a package over. Re-implementing what is already here is the most common form of this mistake.
3. **Does the standard library do it?** In Go this rung catches more than any other: `net/http`'s ServeMux over a router, `slog` over a logging module, `slices`/`maps`/`cmp` over hand-rolled loops, `errors.Join` over `multierr`, `math/rand/v2` over an RNG helper. See **Standard Library: Use the New Packages**.
4. **Does the toolchain or the platform already do it?** These are free and already installed: `//go:embed` over an asset pipeline or a loader, `os.Root` over hand-rolled path confinement, `sync.OnceValue` over a lazy-init helper, the `tool` directive over a script, build tags over a shell wrapper, a `CHECK` constraint or a unique index over application code that re-implements it. In a service's templates, a native input beats a component library and its bundle.
5. **Is it already in `go.mod`?** Use the dependency you already pay for. Never add a module for what a few lines or an existing import can do — but this is not a stdlib-only rule: `golang.org/x/sync/errgroup`, `afero`, and `go-cmp` are installed dependencies this skill recommends (see **Concurrency Patterns** and **Testing Patterns**), not things to hand-roll.
6. **Can it be one line?** Write the one line.
7. **Only then:** the minimum that works — no extra files, layers, or knobs. An interface with one implementation, a factory for one product, and a config field for a value that never changes are all rung-7 failures.

Rungs 3 and 4 name version-gated features (`sync.OnceValue` arrives in Go 1.21; `os.Root` and the `tool` directive in 1.24), so establish the module's floor before recommending one — see **Version Gating**.

**Never cut:** validation at a trust boundary, error handling that prevents data loss, security, cleanup, and anything explicitly requested. In Go that means the error is still checked and wrapped, the `defer` that releases the resource stays, the mutex that makes the map safe stays, and a non-trivial branch or parser still leaves one runnable test behind (see **Testing Patterns**). Laziness shortens the solution, never the reading, and never the checks.

**Fix the root cause, not the reported path.** A bug report names a symptom. Grep every caller of the function before touching it: one guard where all callers route through is a smaller diff than a guard at each call site, and it fixes the callers the ticket never mentioned.

**Name the ceiling you accepted.** When the minimal version has a known limit, record it where the limit lives: `// yagni: whole-table lock — per-key locks if write throughput matters`. A stated ceiling is a simplification; an unstated one is a landmine.

## Package Organization: Flat by Default

**Anti-Pattern:** Using deeply nested directory trees or relying heavily on an `internal/` folder by default to artificially enforce "Clean Architecture" layers. This leads to circular dependencies and difficult navigation.

### 1. The Single-Package Default

Start flat. If you are building a microservice or a simple tool, put everything in the root directory (or alongside your `main.go`). Only create a new package when you truly need a new namespace to clarify the code, or when you need to decouple a strictly independent domain.

One binary means `main.go` at the module root. Reach for `cmd/<name>/` only when the module ships more than one binary — one directory per binary, nothing else in there.

### 2. The Proper Use of `internal/`

The `internal/` directory has a specific compiler enforcement: it prevents other modules from importing the code inside it.

- **For Applications:** If you are building an executable binary, nobody can import your code anyway. Using `internal/` here is usually just adding unnecessary path depth.
- **For Libraries:** Use `internal/` sparingly. It should be reserved for complex subsystems where you need to share exported types between your own packages, but absolutely must prevent end-users from relying on those types.

```
// Idiomatic: A flat, feature-focused library or simple app
myproject/
├── main.go           # Entry point (if application)
├── server.go         # Core logic
├── config.go         # Configuration
├── parser.go         # Domain specific parsing
├── parser_test.go
├── go.mod
└── go.sum
```

### 3. Domain Packages for Service Applications

When an application has genuinely distinct, independently-testable domains, give each its own top-level package. The rule is still **one level deep** — no `internal/` nesting, no Clean Architecture layers. Each package owns one concern. `main.go` wires them together.

The signal to create a package: can this domain be described in one sentence, and can it do its job without knowing the packages that call it? If yes, it earns its own package.

```
// Idiomatic: A web service with distinct domain packages
myservice/
├── main.go        # wires everything together; no business logic here
├── config/        # Config struct, env loading
├── auth/          # identity verification, session middleware
├── db/            # data store client + all queries
├── storage/       # blob storage (S3, R2, GCS)
├── billing/       # payment provider + credit ledger
├── jobs/          # job lifecycle + queue dispatch + worker handler (same domain, one package)
├── web/           # HTTP handlers + HTML templates + static assets (tightly coupled by design)
├── transcribe/    # domain-specific processing — independent pure functions
├── Makefile
├── Dockerfile
└── .env.example
```

**Rules for domain packages:**
- Each package has **one clear purpose** — name it after what it does, not what layer it is (`jobs/`, not `service/`)
- Imports point one way: `web` may import `jobs`, and `jobs` may import `db`, but nothing below imports anything above it, and nothing imports `main`. Sibling imports are fine — cycles and upward imports are the smell. `main` is the wiring point, not a library
- Related sub-concerns that always travel together stay in one package (e.g. job creation + worker handler both live in `jobs/` because they share the job lifecycle domain)
- HTTP handlers and the templates they render belong together in `web/` — they are tightly coupled by design, and `//go:embed` can only read files inside its own package directory, so templates and static assets have to live there
- Do not create `utils/`, `helpers/`, or `common/` packages — these are symptoms of unclear ownership

**Anti-pattern to reject:** Clean Architecture / DDD layers (`service/`, `repository/`, `controller/`, `domain/`). These layer-named packages cause circular imports, force interface proliferation, and add zero clarity in Go. Domain-named packages (`auth/`, `billing/`, `jobs/`) are the correct middle ground between "too flat" and "over-engineered."

### 4. When Flat Stops Working

Flat is a default, not a vow. Split when the code shows you the seam is already there:

- `main.go` has grown past a few hundred lines, or one file now holds several unrelated concerns.
- A struct carries state for two unrelated concerns, or one mutex guards two of them.
- Testing one concern means building the whole program first — the server, the config, the database handle — before the test can run.
- A file's imports are disjoint: half of them exist for one half of the file.
- HTTP handlers, embedded assets, or shell-outs to the OS sit beside pure logic that needs none of them.

These are symptoms, not a target. Split along domain lines and stay one level deep; never split into layers to satisfy the list. A package you can never change without changing its neighbor is one package wearing two names.

### 5. Splitting a God Object

The mechanical version, in order:

1. Move the state and the mutex that guards it **together** — a field separated from its lock is a new bug, not a refactor.
2. Move the methods that touch that state, and nothing else.
3. Keep the fields unexported. The constructor (`NewX`) is the seam: it takes what the package needs, returns the type, and hides the lock.
4. Let `main` construct it and hand it to whoever needs it. The new package must not know who calls it; if it needs to call back, the boundary is on the wrong line.
5. Interfaces stay with the consumer (see **Interface Design**) — the new package does not define one to describe its own callers.

Tests move with the state. If the new package still cannot be tested on its own, the split is not finished.

## Interface Design

### 1. Interfaces are Discovered, Not Designed Upfront

Write concrete types first. Only define an interface when you discover that multiple types need to be used interchangeably by a consumer.

### 2. Define Interfaces Where They Are Used

Interfaces belong in the package that *consumes* them, not the package that *implements* them. This decouples your packages.

```go
// processor/processor.go

// Idiomatic: The consumer defines exactly what it needs.
// The concrete 'UserStore' doesn't even need to know this interface exists.
type UserFetcher interface {
    FetchUser(id string) (*User, error)
}

type Processor struct {
    fetcher UserFetcher
}
```

### 3. Accept Interfaces, Return Structs

Require the smallest interface possible as an input parameter (e.g., `io.Reader` instead of `*os.File`), but return a concrete struct so callers aren't forced to use type assertions to access specific fields or methods.

## Library API Design

### Domain Object as Entry Point

When designing a Go library that wraps a stateful resource (a vault, a database connection, a config store), make that resource struct the primary object. Its methods return domain-typed sub-objects. **Avoid:**

- Passing a config/resource struct as the first argument to every package-level function
- Package-level global state (like pflag's default `FlagSet`) for library code that may be embedded

**Do this instead:**

```go
// Open the primary resource once
v, err := vault.Open(path)

// Domain operations are methods on the primary object
// Each call is stateless (reloads fresh) — no cached state on the struct
idx, err := v.People()             // returns *people.Index, error
note, err := v.Daily(time.Now())   // returns *daily.Note, error
mtgs, err := v.Meetings()          // returns *meetings.Index, error

// Domain types live in sub-packages — callers use type inference
p, err := idx.FindOne("Steve")     // *people.Person
```

**Why this pattern:**

- The primary struct (`Vault`) is the single entry point — callers need just one import to get started
- Sub-packages define the rich domain types (`people.Person`, `daily.Note`) — each type lives with the logic that owns it
- No global state means the library is safe for concurrent use, multiple instances, and testing
- Stateless method calls (reload fresh each time) keep the struct simple — no cache invalidation logic needed
- Type inference (`:=`) means callers rarely need to explicitly import sub-packages for variable declarations

**When to use a global instance instead:** Only for CLI-only tools (like `pflag` itself) where there is truly only ever one instance and ease of use for end-users outweighs library correctness.

## Concurrency Patterns

**Anti-Pattern:** Heavy, static Worker Pools. Go's scheduler is incredibly efficient; you don't need to manually manage pools of workers like OS threads in other languages.

### 1. Share Memory by Communicating

Don't use mutexes to protect shared data if you can pass that data over a channel instead. Channels orchestrate execution; mutexes serialize execution.

### 2. Bounded Concurrency

To limit concurrency, use `errgroup` with `SetLimit`. Don't hand-roll a semaphore channel or a rigid worker pool when this exists.

```go
func FetchAll(ctx context.Context, urls []string, maxConcurrent int) error {
    g, ctx := errgroup.WithContext(ctx)
    g.SetLimit(maxConcurrent)

    for _, url := range urls {
        g.Go(func() error {
            return fetch(ctx, url)
        })
    }

    return g.Wait()
}
```

Loop variables are per-iteration since Go 1.22 — never emit the old `url := url` capture line.

When you don't need error propagation, `sync.WaitGroup.Go` (Go 1.25) removes the Add/Done boilerplate:

```go
var wg sync.WaitGroup
for _, url := range urls {
    wg.Go(func() { process(url) })
}
wg.Wait()
```

### 3. Never Start a Goroutine Without Knowing How It Stops

Every `go func()` must have a clear exit condition, usually governed by a `context.Context` or a closed channel.

Since Go 1.27 that rule is measurable rather than aspirational: the `goroutineleak` profile is generally available in `runtime/pprof` and as the `/debug/pprof/goroutineleak` endpoint.

```go
import (
    "os"
    "runtime/pprof"
)

if p := pprof.Lookup("goroutineleak"); p != nil {
    _ = p.WriteTo(os.Stderr, 1)
}
```

A leaked goroutine is one blocked on a concurrency primitive (channel, `sync.Mutex`, `sync.Cond`, and so on) that cannot possibly become unblocked — the runtime proves it via reachability from the garbage collector. Because the technique is reachability-based, it cannot see leaks caused by blocking on primitives reachable through global variables or through the local variables of runnable goroutines, so a clean profile is evidence, not proof. In Go 1.26 the profile was experimental behind `GOEXPERIMENT=goroutineleakprofile`; that experiment was deleted in Go 1.27.

## Configuration and Struct Design

### Functional Options for Complex Initialization

When a struct has many optional configuration parameters, avoid massive constructors. Use the Functional Options pattern.

```go
type Server struct {
    addr    string
    timeout time.Duration
}

type Option func(*Server)

func WithTimeout(d time.Duration) Option {
    return func(s *Server) { s.timeout = d }
}

func NewServer(addr string, opts ...Option) *Server {
    s := &Server{
        addr:    addr,
        timeout: 30 * time.Second, // Sane default
    }
    for _, opt := range opts {
        opt(s)
    }
    return s
}
```

### Optional Pointer Fields with `new(expr)` (Go 1.26)

A pointer field is how Go distinguishes "absent" from "present but zero". Since Go 1.26, `new` takes an expression, so the temporary variable disappears:

```go
type Person struct {
    Name string
    Age  *int // nil means "not recorded"; a pointer separates absent from 0
}

// Old
age := yearsSince(born)
p := Person{Name: name, Age: &age}

// New (Go 1.26) — the canonical use of new(expr)
p := Person{Name: name, Age: new(yearsSince(born))}
```

Below Go 1.26, keep the `v := expr; &v` form.

## Error Handling

### 1. Errors are Values

Errors aren't exceptions to be caught; they are values to be handled. Check them explicitly.

### 2. Wrap for Context, Not for Stack Traces

When returning an error, add context about what you were trying to do.

```go
// Idiomatic
data, err := os.ReadFile(path)
if err != nil {
    return fmt.Errorf("loading config file %s: %w", path, err)
}
```

### 3. `errors.AsType[T]` over `errors.As` (Go 1.26)

For modules whose floor is Go 1.26 or later, use `errors.AsType[T](err)` — generic, type-safe, and faster, because it needs no pre-declared target and no pointer-to-pointer argument:

```go
// Classic errors.As
var pathErr *fs.PathError
if errors.As(err, &pathErr) {
    // use pathErr
}

// Go 1.26 — the value falls straight out of the call
if pathErr, ok := errors.AsType[*fs.PathError](err); ok {
    // use pathErr
}
```

Below Go 1.26, `errors.As` with `&target` remains the required form; it is correct and supported on every version.

## Testing Patterns

**Anti-Pattern:** Relying on heavy BDD frameworks (like Ginkgo) or complex mocking generation tools. Go testing should just be Go programming.

**Layout decides what is testable.** `package main` cannot be imported, so code that lives there can only be exercised by building the whole binary. A domain that cannot be tested without constructing the server is a domain that wants its own package — that is the practical force behind the split triggers in **Package Organization**.

### 1. Table-Driven Tests

The absolute standard for unit testing in Go. Iterate over a slice of structs containing inputs and expected outputs using `t.Run()`.

```go
func TestParseConfig(t *testing.T) {
    tests := []struct {
        name    string
        input   string
        wantErr bool
    }{
        {"valid config", "port=8080", false},
        {"invalid format", "port=abc", true},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            _, err := ParseConfig(tt.input)
            if (err != nil) != tt.wantErr {
                t.Fatalf("ParseConfig() error = %v, wantErr %v", err, tt.wantErr)
            }
        })
    }
}
```

### 2. Meaningful Helpers with `t.Helper()`

When extracting repeated assertion logic, always call `t.Helper()` to ensure failures point to the actual test case, not the helper function line.

### 3. Fakes and Stubs over Heavy Mocks

Leverage Go's implicit interfaces to write simple, manual fakes. This keeps test dependencies lightweight and test logic transparent.

### 4. Golden Files and the `testdata` Directory

For tests requiring complex inputs or producing large outputs, use a directory named `testdata`. The `go test` tool explicitly ignores these directories.

### 5. Filesystem Abstraction (The Afero Pattern)

Do not hardcode `os` package calls deep within business logic. Accept an interface for the filesystem so tests can run in memory without touching the disk. `github.com/spf13/afero` is the industry standard for this.

```go
import "github.com/spf13/afero"

type FileProcessor struct {
    fs afero.Fs
}

func NewFileProcessor(fs afero.Fs) *FileProcessor {
    return &FileProcessor{fs: fs}
}
```

In tests, inject `afero.NewMemMapFs()` to completely eliminate disk I/O and prevent flaky, slow tests.

### 6. `cmp` over DeepEqual

For comparing complex structs or maps, use `github.com/google/go-cmp/cmp` for rich, readable diffs instead of the strict `reflect.DeepEqual`.

### 7. Modern `testing` Additions (Go 1.24+)

```go
// Test-scoped context — canceled automatically when the test ends.
// Use instead of context.Background() in tests.
func TestFetch(t *testing.T) {
    ctx := t.Context()
    ...
}

// t.Chdir(dir) changes working directory for the test and restores it after.
```

**Benchmarks: always write `b.Loop()`.** `B.Loop` arrived in Go 1.24 and is the mandated benchmark form; since Go 1.26 it no longer inhibits inlining of the loop body, so migrating off `b.N` is unconditionally safe:

```go
func BenchmarkParse(b *testing.B) {
    for b.Loop() {
        Parse(input)
    }
}
```

Arguments, results, and variables assigned inside a `for b.Loop()` body are still kept alive, so the compiler cannot optimize the benchmarked call away — but the loop condition must be written exactly as `b.Loop()`. Never emit the `for i := 0; i < b.N; i++` form, and never mix the two.

**Test output files go in `t.ArtifactDir()` (Go 1.26).** `T.ArtifactDir`, `B.ArtifactDir`, and `F.ArtifactDir` return a per-test directory for files a test produces. With `go test -artifacts` that directory sits under `-outputdir` (or the current directory); without the flag it is a temporary directory removed when the test finishes, so writing there never pollutes the tree.

```go
func TestRender(t *testing.T) {
    out := filepath.Join(t.ArtifactDir(), "render.png")
    if err := render(out); err != nil {
        t.Fatal(err)
    }
}
```

Do not invent your own `os.MkdirTemp` cleanup for test output, and do not write artifacts next to the fixtures in `testdata/`.

**Deterministic crypto randomness: `cryptotest.SetGlobalRandom` (Go 1.26).** It seeds a global, deterministic cryptographic randomness source for the duration of the test, affecting `crypto/rand` and the implicit randomness inside the `crypto/...` packages:

```go
func TestKeyGeneration(t *testing.T) {
    cryptotest.SetGlobalRandom(t, 42) // fixed seed for this test
    ...
}
```

It changes process-wide state, so it cannot be used in parallel tests or tests with parallel ancestors. The `GODEBUG=cryptocustomrand=1` setting temporarily restores the old behavior.

### 8. `testing/synctest` for Concurrent Code (Go 1.25)

Test goroutines and timeouts deterministically with a fake clock — no `time.Sleep` in tests, no flaky timing assumptions:

```go
func TestTimeout(t *testing.T) {
    synctest.Test(t, func(t *testing.T) {
        ctx, cancel := context.WithTimeout(t.Context(), time.Second)
        defer cancel()
        time.Sleep(2 * time.Second) // instant — fake clock advances when all goroutines block
        if ctx.Err() == nil {
            t.Fatal("expected timeout")
        }
    })
}
```

Never write `time.Sleep(100 * time.Millisecond)` to "wait for a goroutine" in a test. Use synctest, channels, or explicit synchronization.

Since Go 1.27, `synctest.Sleep` combines `time.Sleep` with `synctest.Wait` — it blocks until the bubble's fake clock has advanced by the duration *and* every other goroutine in the bubble is durably blocked:

```go
synctest.Test(t, func(t *testing.T) {
    go worker(t) // system under test
    synctest.Sleep(time.Second) // advances the clock, then waits for it to settle
    // assertions see a quiescent system
})
```

Prefer it to a bare `time.Sleep` inside a bubble: when the test and the code under test sleep for the same duration, plain `time.Sleep` makes the wakeup order unpredictable.

Also Go 1.27: `httptest.NewTestServer(t, handler)` runs an `http.Server` on an in-memory fake network with no ports and no real socket, which makes HTTP tests usable inside a synctest bubble:

```go
srv := httptest.NewTestServer(t, mux)
defer srv.Close()
resp, err := srv.Client().Get("http://example.com/") // routed to srv, address ignored
```

The client returned by `srv.Client()` directs HTTP and HTTPS requests to that server regardless of destination host, so tests need no `srv.URL` base. Below Go 1.27, `httptest.NewServer` binds a real localhost port and cannot be used inside a synctest bubble.

## Generics (Go 1.18+)

Generics exist to eliminate duplicated algorithms, not to create type hierarchies. If you are thinking about generics in terms of inheritance or polymorphism, stop — you are writing Java.

### When to Use Generics

Use generics when you have the **same algorithm** that needs to operate on **multiple concrete types**:

```go
// Good: generic algorithm, concrete types as inputs
func Map[S, T any](slice []S, f func(S) T) []T {
    result := make([]T, len(slice))
    for i, v := range slice {
        result[i] = f(v)
    }
    return result
}

// Good: constraint expresses a meaningful requirement
func Min[T cmp.Ordered](a, b T) T {
    if a < b {
        return a
    }
    return b
}
```

### When NOT to Use Generics

```go
// Bad: generic interface for polymorphism — this is Java
type Repository[T any] interface {
    Find(id string) (T, error)
    Save(entity T) error
}

// Good: a concrete interface for what you actually need
type UserStore interface {
    FindUser(id string) (*User, error)
    SaveUser(u *User) error
}
```

- **Do not** create generic base types, generic services, or generic repositories.
- **Do not** use `any` as a constraint to mean "I don't know the type yet." That's a design smell.
- **Do** use `comparable` when you need map keys or equality checks.
- **Do** use `cmp.Ordered` when you need `<`, `>`, `<=`, `>=`.
- Start with a concrete implementation. Generify only when you have the same logic repeated across 3+ types.
- Generic type aliases are fully supported since Go 1.24.

## Standard Library: Use the New Packages

LLMs frequently suggest third-party utilities or write manual helpers that have been in the standard library since Go 1.21. **Always check stdlib first.**

### `slices` package (Go 1.21)

```go
import "slices"

// Searching and testing
slices.Contains(s, "value")
slices.Index(s, "value")           // returns -1 if not found
slices.ContainsFunc(s, func(v string) bool { return v == "x" })

// Sorting
slices.Sort(s)                     // sorts in place, works on any ordered type
slices.SortFunc(s, func(a, b T) int { return cmp.Compare(a.Name, b.Name) })
slices.IsSorted(s)

// Manipulation
slices.Reverse(s)
slices.Compact(s)                  // removes consecutive duplicates
slices.Delete(s, i, j)            // removes elements [i, j)
slices.Clone(s)                   // shallow copy
slices.Concat(s1, s2, s3)         // concatenate multiple slices (1.22)

// Iterator bridging (1.23)
slices.Collect(it)                // iterator -> slice
slices.Sorted(it)                 // iterator -> sorted slice
slices.Values(s)                  // slice -> iterator
```

Never write `sort.Slice(s, func(i, j int) bool { return s[i] < s[j] })` when `slices.Sort(s)` exists.

### `maps` package (Go 1.21; iterators 1.23)

```go
import "maps"

maps.Keys(m)        // iterator over keys — collect with slices.Collect / slices.Sorted
maps.Values(m)      // iterator over values
maps.Clone(m)       // shallow copy
maps.Copy(dst, src) // copies all entries from src into dst
maps.DeleteFunc(m, func(k K, v V) bool { ... })  // delete entries matching predicate
maps.Equal(m1, m2)  // reports whether two maps are equal

// Common idiom: sorted keys in one line
keys := slices.Sorted(maps.Keys(m))
```

### `cmp` package (Go 1.21)

```go
import "cmp"

cmp.Compare(a, b)    // returns -1, 0, or 1; works on any cmp.Ordered type
cmp.Or(a, b, c)      // returns first non-zero value — replaces ternary workarounds (Go 1.22)
min(a, b)            // built-in since Go 1.21
max(a, b)            // built-in since Go 1.21
```

`cmp.Or` is especially useful for default-value patterns:

```go
// Instead of: if cfg.Timeout == 0 { cfg.Timeout = 30 * time.Second }
cfg.Timeout = cmp.Or(cfg.Timeout, 30*time.Second)
```

### `errors.Join` (Go 1.20)

```go
// Combine multiple errors — no third-party library needed
err := errors.Join(err1, err2, err3)

// Works correctly with errors.Is and errors.As
if errors.Is(err, ErrNotFound) { ... }
```

Use this instead of `fmt.Errorf("%w; %w", err1, err2)` or any `multierr` package.

### Iterators: `iter` and Range-over-Func (Go 1.23)

The standard way to expose a sequence from an API without allocating a slice. Return `iter.Seq[T]` or `iter.Seq2[K, V]`; callers use plain `range`:

```go
func (idx *Index) All() iter.Seq[*Person] {
    return func(yield func(*Person) bool) {
        for _, p := range idx.people {
            if !yield(p) {
                return
            }
        }
    }
}

// Caller — just a normal range loop
for p := range idx.All() { ... }
```

Prefer returning iterators over slices for large or lazily-produced sequences. Don't invent a custom `Next()/HasNext()` iterator type — that's Java.

### `math/rand/v2` (Go 1.22)

Always import `math/rand/v2`, never the old `math/rand`. Auto-seeded, cleaner API:

```go
import "math/rand/v2"

rand.IntN(100)            // was rand.Intn — note the capital N
rand.N(10 * time.Second)  // generic: random value in [0, n) for any integer type
```

The generic `N` is also a method since Go 1.27, so a seeded source no longer needs the package-level generic:

```go
r := rand.New(rand.NewPCG(seed1, seed2))
n := r.N(100) // Go 1.27 — in [0, 100), type inferred as int
```

Below Go 1.27, use the package-level `rand.N(100)` for the default source, or `r.IntN(100)` where an `int` will do.

### `encoding/json`: `omitzero` (Go 1.24)

`omitzero` omits any zero value — including `time.Time{}` and zero structs, which `omitempty` never handled correctly:

```go
type Event struct {
    Name      string    `json:"name"`
    StartedAt time.Time `json:"started_at,omitzero"` // omitempty would emit "0001-01-01T..."
}
```

At Go 1.27, `encoding/json` is backed by the `encoding/json/v2` implementation, whose defaults are stricter — it rejects invalid UTF-8 in JSON strings and duplicate names within an object; v1 marshaling and unmarshaling behavior is preserved but error message text may differ. The `cobra-viper` skill covers the config-decoding consequences.

### `uuid` package (Go 1.27)

The standard library now generates and parses UUIDs. For a module that can require Go 1.27, this replaces the reflexive `github.com/google/uuid` recommendation:

```go
import "uuid"

u := uuid.New()      // algorithm suitable for most purposes
v7 := uuid.NewV7()   // time-ordered; good for database keys
id, err := uuid.Parse(s)
```

`MustParse` panics on bad input; use `Parse` anywhere user input reaches it. Random components come from a cryptographically secure generator. Below Go 1.27, keep `github.com/google/uuid`.

### `strings.CutLast` / `bytes.CutLast` (Go 1.27)

Cut around the *last* occurrence of a separator, replacing the `LastIndex` + slicing idiom:

```go
// Old
i := strings.LastIndex(path, "/")
if i < 0 {
    return path, "", false
}
return path[:i], path[i+1:], true

// New (Go 1.27)
before, after, found := strings.CutLast(path, "/")
```

Both return slices of the original (no copy) and report `found`. Below Go 1.27, the `LastIndex` form is the fallback.

### `log/slog.NewMultiHandler` (Go 1.26)

Fan a record out to several handlers — for example text to stderr and JSON to a file — without writing a logging wrapper:

```go
handler := slog.NewMultiHandler(
    slog.NewTextHandler(os.Stderr, nil),
    slog.NewJSONHandler(file, nil),
)
logger := slog.New(handler)
```

`MultiHandler.Enabled` reports true if *any* of the handlers is enabled. Below Go 1.26, write a small `slog.Handler` that forwards to a slice of handlers.

### `net/url` Clones (Go 1.27)

`URL.Clone` and `Values.Clone` return deep copies. Use them before mutating a URL or query map you did not construct — a plain struct copy still shares the userinfo pointer and the query map:

```go
u2 := u.Clone()      // deep copy of all URL fields
q2 := q.Clone()      // deep copy of Values
q2.Set("page", "2")  // q is untouched
```

Below Go 1.27, copy manually (and remember `url.Userinfo` and the `Values` map each need their own copy).

### `reflect` Iterators (Go 1.26)

`Type.Fields`, `Type.Methods`, `Type.Ins`, `Type.Outs`, `Value.Fields`, and `Value.Methods` return iterators, so they compose with the range-over-func style used elsewhere in this skill:

```go
for field := range reflect.TypeOf(cfg).Fields() {
    fmt.Println(field.Name, field.Type)
}
```

Below Go 1.26, index with `NumField()`/`Field(i)`.

## Concurrency: Modern Patterns

### `sync/atomic` Typed Values (Go 1.19)

Use the typed atomic values instead of the function-based API:

```go
// Old (still works but avoid for new code)
var count int64
atomic.AddInt64(&count, 1)
val := atomic.LoadInt64(&count)

// New — type-safe, no pointer arithmetic
var count atomic.Int64
count.Add(1)
val := count.Load()

// Other typed atomics
var flag  atomic.Bool
var ptr   atomic.Pointer[MyStruct]
var val32 atomic.Int32
var val64 atomic.Uint64
```

### `context.WithoutCancel` (Go 1.21)

When you need to detach a context's cancellation but preserve its values (e.g., for a background task that should outlive a request):

```go
// The background job should keep running even after the HTTP request context cancels
bgCtx := context.WithoutCancel(requestCtx)
go doBackgroundWork(bgCtx)
```

### `time` Channels Are Always Unbuffered (Go 1.27)

Channels created by package `time` (timers and tickers) are now always unbuffered (synchronous), irrespective of GODEBUG settings. Go 1.23 made capacity-0 timer channels the behavior for modules declaring `go 1.23` or later, with `asynctimerchan=1` as the opt-out; Go 1.27 removed that setting permanently, so there is no longer a way back.

Consequences when consuming timers and tickers:

- `Stop` and `Reset` guarantee that no value prepared before the call is sent or received after it. The buffered-channel dance around `Reset` is obsolete, and a timer's fired value is never silently discarded.
- `len` and `cap` of a timer or ticker channel are `0`, so never poll the length to decide whether a receive will succeed — use a non-blocking receive with `select` and `default`.
- A ticker adjusts its interval or drops ticks to make up for a slow receiver, so consume `Ticker.C` promptly or `Stop` the ticker.

### Runtime Changes with No Code Impact (Go 1.26 / 1.27)

Behavior to know about so you neither chase it nor "fix" it:

- **Green Tea GC (Go 1.26)** — enabled by default; 10–40% less GC overhead. Opt out with `GOEXPERIMENT=nogreenteagc` (expected to be removed in Go 1.27).
- **Heap base address randomization (Go 1.26)** — on 64-bit platforms allocated objects no longer sit at predictable addresses. Opt out with `GOEXPERIMENT=norandomizedheapbase64`.
- **Size-specialized small allocations (Go 1.27)** — allocations below 80 bytes are up to 30% cheaper, at the cost of roughly 60 KB of binary size. Opt out with `GOEXPERIMENT=nosizespecializedmalloc` (expected to be removed in Go 1.28).

## HTTP: Use the Improved stdlib Router (Go 1.22)

LLMs reflexively recommend gorilla/mux or chi for any routing beyond the trivial. Since Go 1.22, the standard `net/http` ServeMux handles method and path-parameter routing natively.

```go
mux := http.NewServeMux()

// Method-scoped routes
mux.HandleFunc("GET /users", listUsers)
mux.HandleFunc("POST /users", createUser)

// Path parameters — accessed via r.PathValue
mux.HandleFunc("GET /users/{id}", func(w http.ResponseWriter, r *http.Request) {
    id := r.PathValue("id")
    // ...
})

// Wildcard
mux.HandleFunc("GET /files/{path...}", serveFile)
```

Reach for chi or gorilla/mux only when you need named route generation or regex constraints — middleware doesn't justify a framework (see below). For pure method + path routing, the stdlib is sufficient.

### Production Servers: Always Set Timeouts

`http.ListenAndServe(addr, mux)` ships with **no timeouts** — a single slow client can hold a connection open forever (slow-loris). LLM-generated servers almost never set these. Never emit a production server without them:

```go
srv := &http.Server{
    Addr:              ":8080",
    Handler:           mux,
    ReadHeaderTimeout: 5 * time.Second,   // slow-loris protection
    ReadTimeout:       10 * time.Second,  // full request read
    WriteTimeout:      30 * time.Second,  // response write (covers handler time)
    IdleTimeout:       120 * time.Second, // keep-alive connections
}
```

For per-route control beyond `WriteTimeout`, use `http.TimeoutHandler` or handler-level `context.WithTimeout`. Outbound calls need the same discipline: `http.DefaultClient` has no timeout either — construct a client with one.

### Graceful Shutdown

Every production server needs a shutdown path that stops accepting connections and drains in-flight requests. The pattern:

```go
func run(ctx context.Context) error {
    ctx, stop := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
    defer stop()

    srv := &http.Server{ /* ... timeouts as above ... */ }

    errCh := make(chan error, 1)
    go func() { errCh <- srv.ListenAndServe() }()

    select {
    case err := <-errCh:
        return err // ListenAndServe failed at startup
    case <-ctx.Done():
        shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
        defer cancel()
        return srv.Shutdown(shutdownCtx) // drain in-flight, then exit
    }
}
```

- `Shutdown` needs a fresh context with its own deadline — the signal context is already canceled.
- Long-lived connections (SSE, websockets) must watch `r.Context()` or they'll hold shutdown until the drain deadline.

### Middleware Is Just a Function

No framework needed. A middleware is `func(http.Handler) http.Handler`:

```go
func withRequestLog(logger *slog.Logger, next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        start := time.Now()
        next.ServeHTTP(w, r)
        logger.Info("request", "method", r.Method, "path", r.URL.Path, "dur", time.Since(start))
    })
}

// Compose by wrapping — outermost runs first
handler := withRequestLog(logger, withAuth(mux))
```

Don't import a middleware framework for what function composition already does.

## Syntax: Use Current Go

LLMs frequently generate outdated syntax. Know the current idioms:

### Range over Integer (Go 1.22)

```go
// Old
for i := 0; i < 10; i++ { ... }

// New
for i := range 10 { ... }
```

### `new(expr)` (Go 1.26)

`new` accepts an expression and returns a pointer to a new variable holding that value, replacing the `v := expr; &v` dance:

```go
// Old
v := yearsSince(born)
p := &v

// New (Go 1.26)
p := new(yearsSince(born))
```

The canonical case is an optional pointer struct field — see Configuration and Struct Design above.

### Generic Methods (Go 1.27)

A method declaration may declare its own type parameters, so a generic helper can be scoped to a type instead of the whole package. `math/rand/v2` gains the method form `(*Rand) N[Int intType](Int) Int` alongside the existing package-level generic `N`:

```go
import "math/rand/v2"

r := rand.New(rand.NewPCG(seed1, seed2))
n := r.N(100) // method form, Go 1.27
```

Two hard limits:

- A method of an interface may not declare type parameters.
- An interface method cannot be implemented by a generic method.

Below Go 1.27, hoist the helper to a package-level generic function.

### Self-Referential Type-Parameter Constraints (Go 1.26)

A generic type may now refer to itself in its own type-parameter list:

```go
type Adder[A Adder[A]] interface {
    Add(A) A
}
```

Before Go 1.26 this was rejected by the compiler.

### Struct Literal Keys: Any Field Selector (Go 1.27)

A key in a struct literal may be any valid field selector for the struct type, not just a top-level field name — so a promoted field can be named directly:

```go
type Base struct{ ID int }

type User struct {
    Base
    Name string
}

u := User{ID: 1, Name: "steve"} // Go 1.27 — Base.ID named directly
```

Below Go 1.27, initialize the embedded value explicitly: `User{Base: Base{ID: 1}, Name: "steve"}`.

### Generalized Function Type Inference (Go 1.27)

Function type inference applies in every context where a generic function is assigned to, or converted to, a value of a matching function type:

```go
type Less func(a, b string) int

// Go 1.26 and earlier: "cannot use generic function cmp.Compare without instantiation"
var less Less = Less(cmp.Compare) // Go 1.27 infers T = string
```

Below Go 1.27, instantiate explicitly: `cmp.Compare[string]`.

### Build Constraints

```go
// Old (deprecated — do not generate)
// +build linux darwin

// Current
//go:build linux || darwin
```

The `//go:build` form is required since Go 1.17. Never emit the old `// +build` syntax.

### `any` Instead of `interface{}`

```go
// Old
func Print(v interface{}) { ... }

// Current — `any` is a built-in alias for interface{} since Go 1.18
func Print(v any) { ... }
```

### Tool Dependencies in `go.mod` (Go 1.24)

Never generate a `tools.go` file with blank imports. Track dev tools with the `tool` directive:

```bash
go get -tool golang.org/x/tools/cmd/stringer
go tool stringer -type=Color   # run it
```

## Structured Logging with `log/slog` (Go 1.21)

The skill note above covers basic setup. The patterns LLMs most often get wrong:

```go
// Pass a logger via context for request-scoped logging
func HandleRequest(ctx context.Context, r *Request) {
    logger := slog.With("request_id", r.ID, "user_id", r.UserID)
    // All subsequent log calls carry these fields automatically
    logger.Info("handling request")
    process(ctx, logger, r)
}

// Group related fields
logger.With(slog.Group("http",
    slog.String("method", r.Method),
    slog.String("path", r.URL.Path),
    slog.Int("status", status),
))

// Log at the right level — LLMs over-use Info
logger.Debug("cache miss", "key", key)     // internal state, high volume
logger.Info("server started", "addr", addr) // lifecycle events
logger.Warn("retrying", "attempt", n)       // recoverable problems
logger.Error("request failed", "err", err)  // needs attention
```

- **Never** use a package-level `log` or `slog` global beyond `main`. Pass `*slog.Logger` as a dependency.
- **Never** log and return an error. Log at the boundary, return the error through the call stack.
- Use `slog.Default()` as the fallback only in `main` or in libraries when no logger is provided.

## Experimental Packages: Not for Production

These APIs ship behind `GOEXPERIMENT` build-time flags, and their APIs are explicitly **not stable**. Prototype with them, or use them to prepare for a future release — never put them in production code or in a library others depend on.

- **`simd` (Go 1.27)** — portable, vector-size-agnostic SIMD behind `GOEXPERIMENT=simd`, available on all architectures. Write vector code once and let the compiler map it to the hardware width.
- **`simd/archsimd` (Go 1.26; extended in Go 1.27)** — architecture-specific SIMD behind `GOEXPERIMENT=simd`. Introduced for amd64 in Go 1.26; Go 1.27 extends it to arm64 Neon and WebAssembly 128-bit SIMD. It is intentionally non-portable.
- **`runtime/secret` (Go 1.26)** — behind `GOEXPERIMENT=runtimesecret`, amd64 and arm64 Linux only. Goroutines created while in secret mode also execute in secret mode as of Go 1.27.

## Debugging: The Go Toolchain Is Not the Problem

**The Go tool is extremely reliable. It is almost never the source of a bug.**

When debugging, do not waste time suspecting `go run`, `go build`, `go test`, or the build cache. The Go toolchain does what it says:

- `go run` always recompiles from source. It does not use a stale cached binary.
- `go build` is deterministic and correct.
- `go test` runs the actual compiled test binary.
- The build cache is keyed by source content — if the source changed, the cache is invalidated automatically.

**If an error persists after you edit the code, the explanation is one of these — in order of likelihood:**

1. The edit did not fix the underlying logic error.
2. The edit was made in the wrong file, wrong function, or wrong package.
3. There is a second call site with the same bug that was not updated.
4. The error is coming from a different code path than the one being edited.

**What to do instead of blaming the tool:**

- Re-read the error message carefully. Go's error messages are accurate.
- Confirm the file you edited is actually the file being compiled (`go list -f '{{.GoFiles}}' .`).
- Add a `fmt.Println` or `t.Log` at the exact site to verify execution reaches it.
- Check that all call sites of a changed function were updated.

Do not suggest clearing the build cache (`go clean -cache`), restarting the Go toolchain, or any other tool-level intervention before first exhausting all code-level explanations. The tool is not lying to you.