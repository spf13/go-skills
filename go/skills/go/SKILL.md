---
name: go
description: >
  Expert Go programming skill authored by spf13 (former Go team lead, author of Cobra, Viper, Hugo, Afero).
  Covers idiomatic Go — package design, error handling, interfaces, concurrency, testing, and
  project layout. Use when writing, reviewing, or refactoring any Go code.
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

## Core Principles

### 1. Clear is Better than Clever

Go favors readability and simplicity over abstraction and cleverness. Code should be obvious. If you have to read a function three times to understand its control flow, it needs to be rewritten.

```go
// Idiomatic: Direct, linear control flow
func GetUser(id string) (*User, error) {
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

## Package Organization: Flat by Default

**Anti-Pattern:** Using deeply nested directory trees or relying heavily on an `internal/` folder by default to artificially enforce "Clean Architecture" layers. This leads to circular dependencies and difficult navigation.

### 1. The Single-Package Default

Start flat. If you are building a microservice or a simple tool, put everything in the root directory (or alongside your `main.go`). Only create a new package when you truly need a new namespace to clarify the code, or when you need to decouple a strictly independent domain.

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

## Interface Design

### 1. Interfaces are Discovered, Not Designed Upfront

Write concrete types first. Only define an interface when you discover that multiple types need to be used interchangeably by a consumer.

### 2. Define Interfaces Where They Are Used

Interfaces belong in the package that *consumes* them, not the package that *implements* them. This decouples your packages.

```go
// internal/processor/processor.go

// Idiomatic: The consumer defines exactly what it needs.
// The concrete 'UserStore' doesn't even need to know this interface exists.
type UserFetcher interface {
    GetUser(id string) (*User, error)
}

type Processor struct {
    fetcher UserFetcher
}
```

### 3. Accept Interfaces, Return Structs

Require the smallest interface possible as an input parameter (e.g., `io.Reader` instead of `*os.File`), but return a concrete struct so callers aren't forced to use type assertions to access specific fields or methods.

## Concurrency Patterns

**Anti-Pattern:** Heavy, static Worker Pools. Go's scheduler is incredibly efficient; you don't need to manually manage pools of workers like OS threads in other languages.

### 1. Share Memory by Communicating

Don't use mutexes to protect shared data if you can pass that data over a channel instead. Channels orchestrate execution; mutexes serialize execution.

### 2. Bounded Concurrency (The Semaphore Pattern)

If you need to limit concurrency, use a buffered channel as a semaphore rather than a rigid worker pool.

```go
func FetchAll(urls []string, maxConcurrent int) error {
    sem := make(chan struct{}, maxConcurrent)
    g, ctx := errgroup.WithContext(context.Background())

    for _, url := range urls {
        url := url // Note: Go 1.22+ handles this natively

        sem <- struct{}{} // Block if we hit max concurrency

        g.Go(func() error {
            defer func() { <-sem }() // Release token
            return fetch(ctx, url)
        })
    }

    return g.Wait()
}
```

### 3. Never Start a Goroutine Without Knowing How It Stops

Every `go func()` must have a clear exit condition, usually governed by a `context.Context` or a closed channel.

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

## Testing Patterns

**Anti-Pattern:** Relying on heavy BDD frameworks (like Ginkgo) or complex mocking generation tools. Go testing should just be Go programming.

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
```

Never write `sort.Slice(s, func(i, j int) bool { return s[i] < s[j] })` when `slices.Sort(s)` exists.

### `maps` package (Go 1.21)

```go
import "maps"

maps.Keys(m)        // returns an iterator over keys (1.23) / []K (earlier)
maps.Values(m)      // returns an iterator over values
maps.Clone(m)       // shallow copy
maps.Copy(dst, src) // copies all entries from src into dst
maps.Delete(m, func(k, v T) bool { ... })  // delete entries matching predicate
maps.Equal(m1, m2)  // reports whether two maps are equal
```

### `cmp` package (Go 1.21)

```go
import "cmp"

cmp.Compare(a, b)    // returns -1, 0, or 1; works on any cmp.Ordered type
cmp.Or(a, b, c)      // returns first non-zero value — replaces ternary workarounds
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

Reach for chi or gorilla/mux only when you need middleware chaining, named route generation, or regex constraints. For pure method + path routing, the stdlib is sufficient.

## Syntax: Use Current Go

LLMs frequently generate outdated syntax. Know the current idioms:

### Range over Integer (Go 1.22)

```go
// Old
for i := 0; i < 10; i++ { ... }

// New
for i := range 10 { ... }
```

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
