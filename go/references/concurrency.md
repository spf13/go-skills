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

### `context.Context`: First Parameter, Never a Struct Field

- `ctx context.Context` is the first parameter of every function that does I/O, blocks, or starts a goroutine. Name it `ctx`, never `context`, never nil — start from `context.Background()` (or `t.Context()` in tests) and thread it down.
- **Never store a context in a struct field.** The struct outlives the request that created the context, so cancellation silently stops working and one request's deadline leaks into another's work. Store the values you need, not the context.
- Context carries cancellation, deadlines, and request-scoped metadata only — never optional parameters, a logger, a config, or a DB handle. See **Structured Logging** for passing `*slog.Logger` explicitly.

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

- **Green Tea GC (Go 1.26)** — enabled by default; 10–40% less GC overhead. Opt out with `GOEXPERIMENT=nogreenteagc`; announced as temporary, it still works in Go 1.27.
- **Heap base address randomization (Go 1.26)** — on 64-bit platforms allocated objects no longer sit at predictable addresses. Opt out with `GOEXPERIMENT=norandomizedheapbase64`.
- **Size-specialized small allocations (Go 1.27)** — allocations below 80 bytes are up to 30% cheaper, at the cost of roughly 60 KB of binary size. Opt out with `GOEXPERIMENT=nosizespecializedmalloc` (expected to be removed in Go 1.28).
