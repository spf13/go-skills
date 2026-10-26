## Structured Logging with `log/slog` (Go 1.21)

These are the patterns LLMs most often get wrong:

```go
// Take a *slog.Logger as a dependency; derive a request-scoped child from it
func HandleRequest(ctx context.Context, logger *slog.Logger, r *Request) {
    logger = logger.With("request_id", r.ID, "user_id", r.UserID)

    // Group related fields under one key
    logger = logger.With(slog.Group("http",
        slog.String("method", r.Method),
        slog.String("path", r.URL.Path),
    ))

    // All subsequent log calls carry these fields automatically
    logger.Info("handling request")
    process(ctx, logger, r)
}

// Log at the right level — LLMs over-use Info
func logLevels(logger *slog.Logger, addr, key string, err error) {
    logger.Debug("cache miss", "key", key)      // internal state, high volume
    logger.Info("server started", "addr", addr) // lifecycle events
    logger.Warn("retrying", "err", err)         // recoverable problems
    logger.Error("request failed", "err", err)  // needs attention
}
```

- **One logger policy: inject `*slog.Logger`.** Any package that logs takes a `*slog.Logger` parameter (or holds one on a struct) — never a package-level `log` or `slog` global, and never a `slog.Default()` fallback inside library code. `slog.Default()` is created and used only at the `main` boundary, which passes that logger, or one derived from it, down the call chain.
- **Never** log and return an error. Log at the boundary, return the error through the call stack.
