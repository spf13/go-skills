---
name: go
description: >
  Expert Go programming skill authored by spf13 (former Go team lead, author of Cobra, Viper, Hugo, Afero).
  Covers idiomatic Go — package design, error handling, interfaces, concurrency, testing, and
  project layout. Use when writing, reviewing, or refactoring any Go code.
---

# Idiomatic Go: The Go Way

Idiomatic Go patterns and best practices for building robust, efficient, and maintainable applications.

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
