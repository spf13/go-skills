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
