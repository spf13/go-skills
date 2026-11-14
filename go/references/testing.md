## Testing Patterns

**Anti-Pattern:** Relying on heavy BDD frameworks (like Ginkgo) or complex mocking generation tools. Go testing should just be Go programming.

**Layout decides what is testable.** `package main` cannot be imported, so code that lives there can only be exercised by building the whole binary. A domain that cannot be tested without constructing the server is a domain that wants its own package — that is the practical force behind the split triggers in **Package Organization**.

**The test loop is `go test -race ./...`.** Run it as the ordinary loop, not a special occasion: the race detector is the only reliable way to catch unsynchronized access to shared memory, and it proves the race rather than waiting for a flake. Keep the suite fast enough that `-race` stays affordable.

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

In tests, inject `afero.NewMemMapFs()` to completely eliminate disk I/O and prevent flaky, slow tests. The `afero` skill covers where `MemMapFs` differs from a real disk and when read-only code should accept `fs.FS` instead.

This rule governs **your** code's filesystem access. It does not reach inside a library that owns the syscalls: `fileflow.Move`/`Copy` take real path strings and use `MoveFileEx` (Windows) or link-and-remove/`os.Rename` (Unix), none of which a memory FS can model — so it is not afero-injectable, and its file operations are integration-tested instead. Do not try to fake a library's syscalls; wrap the library behind an interface you own (`type Mover interface { Move(src, dst string) (string, error) }`), fake that in unit tests, and keep one integration test on a real `t.TempDir()`. The seam is discussed on both sides: see `fileflow-pathologize` for the library's testability contract.

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

**Benchmarks: `b.Loop()` on a Go 1.24+ floor.** `B.Loop` arrived in Go 1.24. For modules whose floor is Go 1.24 or later, write the loop as `for b.Loop()` — it runs the body the right number of times and keeps arguments, results, and variables assigned in the body alive, so no sink variable is needed. Since Go 1.26 it no longer inhibits inlining of the loop body, so on that floor migrating off `b.N` is unconditionally safe:

```go
func BenchmarkParse(b *testing.B) {
    for b.Loop() {
        Parse(input)
    }
}
```

Where `b.Loop()` is the form in use, the loop condition must be written exactly as `b.Loop()`, and it must not be mixed with a `b.N` counter in the same benchmark.

Below Go 1.24, `b.Loop` does not exist and `for i := 0; i < b.N; i++` remains the required form. That form does not keep results alive, so assign the benchmarked call to a package-level sink:

```go
var sink string

func BenchmarkParse(b *testing.B) {
    for i := 0; i < b.N; i++ {
        sink = Parse(input)
    }
}
```

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
