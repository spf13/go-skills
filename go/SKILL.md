---
name: go
description: >
  Use when the user writes, reviews, debugs, or refactors Go code — a `.go` file, `go.mod`, a Go
  package or web service, "golang", a goroutine or interface design question, or an error-handling
  decision. Not for Cobra/Viper CLI wiring (use cobra-viper), cutting a release (use go-release),
  reviewing a design doc or spec (use go-spec-reviewer), a Wails desktop app (use wails), or
  cross-filesystem moves and untrusted path segments (use fileflow-pathologize). Current through
  Go 1.27.
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

## When NOT to Use This Skill

This skill covers general Go idioms and project design. For CLI framework patterns use `cobra-viper`; for release, versioning, and distribution use `go-release`; for pre-implementation spec review use `go-spec-reviewer`; for desktop/webview apps use `wails`; for cross-filesystem file moves and untrusted path sanitization use `fileflow-pathologize`.

## Version Gating: Check the Module's Floor First

Every version-gated feature is a floor: the module's packages must compile on the release that introduced it. Before recommending such a feature, establish the floor and recommend only what it supports.

1. Read the `go` directive in the module's `go.mod`. That version is the floor for the module's own packages — a module declaring `go 1.21` must still compile on Go 1.21, whatever toolchain you happen to be running.
2. Check the file's build constraints. A file behind `//go:build go1.24` may use Go 1.24 features even when the module's `go` directive is older.
3. If the floor is unknown, read `go.mod` or ask. Never guess, and never assume the reader is on the latest release.

The inverse errors are just as real: do not flag an idiom as unavailable in a module whose floor already permits it, and do not flag an idiom as obsolete when the floor cannot reach its replacement — a module below a feature's release is entitled to the older idiom. And never raise a module's `go` directive merely to reach a convenience helper — a `go` directive is a floor imposed on every dependent of the module.


Full table: [references/version-index.md](references/version-index.md).

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

Before writing code, stop at the first rung that holds. Climb **after** you understand the problem — read the code the change touches and trace the real flow first. If two rungs both work, take the earlier (smaller-numbered) rung and move on.

1. **Does this need to exist?** "We'll want it later" is not a need. Skip it, and say so in one line.
2. **Does the codebase already have it?** Search before you write — the helper, type, or pattern may already live a package over. Re-implementing what is already here is the most common form of this mistake.
3. **Does the standard library do it?** In Go this rung catches more than any other: `net/http`'s ServeMux over a router, `slog` over a logging module, `slices`/`maps`/`cmp` over hand-rolled loops, `errors.Join` over `multierr`, `math/rand/v2` over an RNG helper. See **Standard Library: Use the New Packages**.
4. **Does the toolchain or the platform already do it?** These are free and already installed: `//go:embed` over an asset pipeline or a loader, `os.Root` over hand-rolled path confinement, `sync.OnceValue` over a lazy-init helper, the `tool` directive over a script, build tags over a shell wrapper, a `CHECK` constraint or a unique index over application code that re-implements it. In a service's templates, a native input beats a component library and its bundle.
5. **Is it already in `go.mod`?** Use the dependency you already pay for. Never add a module for what a few lines or an existing import can do — but this is not a stdlib-only rule: `golang.org/x/sync/errgroup`, `afero`, and `go-cmp` are installed dependencies this skill recommends (see **Concurrency Patterns** and **Testing Patterns**), not things to hand-roll.
6. **Can it be one line?** Write the one line.
7. **Only then:** the minimum that works — no extra files, layers, or knobs. An interface with one implementation, a factory for one product, and a config field for a value that never changes are all rung-7 failures.

Rungs 3 and 4 name version-gated features (`sync.OnceValue` arrives in Go 1.21; `os.Root` and the `tool` directive in 1.24), so establish the module's floor before recommending one — see **Version Gating**.

**Containment: `os.Root` or `pathologize.Join`?** They answer different questions, so pick by threat rather than habit.

- `os.Root` (Go 1.24+) is kernel-enforced confinement: `root.OpenRoot(dir)` makes every later open, create, and rename resolve from a directory handle, so a symlink or an absolute path cannot escape the tree. This is the stdlib rung when the threat is *escape from a trusted directory* and the floor permits it.
- `pathologize.Join(root, parts...)` is lexical, pre-open sanitization: each untrusted part is stripped of separators and reserved/illegal characters, so the result stays under `root` **and** is a valid name on every OS (Windows `CON`/`NUL`, `:` `*` `?`). Containment here belongs to the library when portability is the problem, or when the floor is below 1.24, or when you only need a string to hand to some other API.

Neither substitutes for the other: `Join` knows nothing about symlinks, and `os.Root` is not sanitization — it will happily create `CON` or `a:b` inside its root. On a 1.24+ floor with untrusted names, do both: `Join` the path, then open it under a `*os.Root`. See the `fileflow-pathologize` skill.

**Never cut:** validation at a trust boundary, error handling that prevents data loss, security, cleanup, and anything explicitly requested. In Go that means the error is still checked and wrapped, the `defer` that releases the resource stays, the mutex that makes the map safe stays, and a non-trivial branch or parser still leaves one runnable test behind (see **Testing Patterns**). Laziness shortens the solution, never the reading, and never the checks.

**Fix the root cause, not the reported path.** A bug report names a symptom. Grep every caller of the function before touching it: one guard where all callers route through is a smaller diff than a guard at each call site, and it fixes the callers the ticket never mentioned.

**Name the ceiling you accepted.** When the minimal version has a known limit, record it where the limit lives: `// yagni: whole-table lock — per-key locks if write throughput matters`. A stated ceiling is a simplification; an unstated one is a landmine.

## Package Organization: Flat by Default

**Anti-Pattern:** Using deeply nested directory trees or relying heavily on an `internal/` folder by default to artificially enforce "Clean Architecture" layers. This leads to circular dependencies and difficult navigation.

### 1. The Single-Package Default

Start flat. If you are building a microservice or a simple tool, put everything in the root directory (or alongside your `main.go`). Only create a new package when you truly need a new namespace to clarify the code, or when you need to decouple a strictly independent domain.

One binary means `main.go` at the module root. Reach for `cmd/<name>/` only when the module ships more than one binary — one directory per binary, nothing else in there.

Note the two different things spelled `cmd/`. `cmd/<name>/` is a binary entrypoint: its own `package main`, holding a `main.go` and nothing else. A bare `cmd/` is an ordinary library package and never holds `main.go`.

**One exception, owned by `cobra-viper`:** a Cobra CLI with subcommands uses a single bare `cmd/` as its command-routing layer even when it ships one binary, with `main.go` at the root as a shim. Commands are the CLI's public surface, not a second binary. That skill states the rule and the layout; this one governs every other module.

### 2. The Proper Use of `internal/`

The `internal/` directory has a specific compiler enforcement: it prevents other modules from importing the code inside it. That is its only job, so the rule is scoped by what you are building:

- **Applications: no `internal/`.** Nobody can import an executable's code anyway, so it buys nothing but path depth.
- **Libraries: at most one level.** Reserve it for a subsystem whose exported types must be shared between your own packages but must not be relied on by end-users. `internal/store` is fine; `internal/store/sql/repo` is not.

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

When an application has genuinely distinct, independently-testable domains, give each its own top-level package. Each package sits **one level deep** — no Clean Architecture layers, and no `internal/` (§2). Each package owns one concern. `main.go` wires them together.

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

Loop variables are per-iteration since Go 1.22 — on a floor of Go 1.22 or later, never emit the old `url := url` capture line. Below that floor the capture line is still required, or goroutines will read the loop variable after it has advanced.

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

### 3. Sentinel Errors: `errors.Is` and `errors.As`

Declare sentinels once at package level, return or wrap that same value (`%w`), and match with `errors.Is` — never `==` on a returned error, never string comparison.

```go
var ErrNotFound = errors.New("not found")

// errors.Is walks the wrap chain; == does not.
if _, err := store.Find(id); errors.Is(err, ErrNotFound) {
    return fmt.Errorf("loading user %s: %w", id, err)
}

// errors.As extracts a typed error and its data.
var pathErr *fs.PathError
if errors.As(err, &pathErr) {
    use(pathErr.Path)
}
```

Returning `errors.New("not found")` at the call site destroys identity: the caller can no longer match it.

### 4. `errors.AsType[T]` over `errors.As` (Go 1.26)

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



