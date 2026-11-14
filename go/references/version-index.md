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
| 1.27 | `encoding/json/v2` and `encoding/json/jsontext` packages, strict by default |
| 1.27 | `encoding/json` backed by the `encoding/json/v2` implementation, v1 behavior kept |
| 1.27 | `simd` (experiment: `GOEXPERIMENT=simd`) |
| 1.27 | `simd/archsimd` on arm64 and WebAssembly (experiment: `GOEXPERIMENT=simd`) |

