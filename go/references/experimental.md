## Experimental Packages: Not for Production

These APIs ship behind `GOEXPERIMENT` build-time flags, and their APIs are explicitly **not stable**. Prototype with them, or use them to prepare for a future release — never put them in production code or in a library others depend on.

- **`simd` (Go 1.27)** — portable, vector-size-agnostic SIMD behind `GOEXPERIMENT=simd`, available on all architectures. Write vector code once and let the compiler map it to the hardware width.
- **`simd/archsimd` (Go 1.26; extended in Go 1.27)** — architecture-specific SIMD behind `GOEXPERIMENT=simd`. Introduced for amd64 in Go 1.26; Go 1.27 extends it to arm64 Neon and WebAssembly 128-bit SIMD. It is intentionally non-portable.
- **`runtime/secret` (Go 1.26)** — behind `GOEXPERIMENT=runtimesecret`, amd64 and arm64 Linux only. Goroutines created while in secret mode also execute in secret mode as of Go 1.27.


## Static Analysis Baseline

Beyond `go vet` (and the `stdversion` check it now includes), the baseline is `gofmt -l .` for formatting and `staticcheck ./...` — or `golangci-lint run`, which embeds staticcheck — before review, because they catch dead code, misused stdlib calls, and nil-safety bugs that vet does not.


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
