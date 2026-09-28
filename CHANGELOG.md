# Changelog

All notable changes to this catalog are documented here.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/), and this catalog adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html). One version covers all six plugins; they move together.

## [1.1.0] - 2026-09-28

The catalog now covers **Go 1.26 and Go 1.27**, and every Go skill states the release it covers and gates its advice on the module's own `go` directive.

### Added

- `go` — a version-gating rule ahead of every feature section: derive the floor from the `go.mod` `go` directive and any `//go:build go1.X` constraints, and recommend only what that floor supports.
- `go` — a version-to-feature index covering Go 1.17 through Go 1.27, and the `stdversion` vet check (`go vet ./...`, `go test`) as the mechanical backstop for standard-library symbols.
- `go` — the Go 1.26 and 1.27 language and standard-library additions: `new(expr)`, generic methods and their two hard limits, self-referential type-parameter constraints, struct-literal field selectors, generalized function type inference, `errors.AsType`, the `uuid` package, `strings`/`bytes.CutLast`, `log/slog.NewMultiHandler`, the `net/url` clones, `reflect` iterators, and `math/rand/v2` `(*Rand).N`.
- `go` — the Go 1.26 and 1.27 operational changes: the `goroutineleak` profile, unbuffered `time` channels, `b.Loop` no longer inhibiting inlining, `ArtifactDir` with `go test -artifacts`, `testing/cryptotest.SetGlobalRandom`, `testing/synctest.Sleep`, `httptest.NewTestServer`, and one clearly labelled non-production block covering `simd`, `simd/archsimd`, and `runtime/secret`.
- `go-release` — Go 1.26's rebuilt `go fix` and its modernizers as a concrete deprecation-to-migration path, `//go:fix inline` for rewriting callers, the conservative `go mod init` default, Go 1.27's duplicate-`require`-block merging in `go mod tidy`, GODEBUG/GOEXPERIMENT removal discipline, platform floors (macOS 13, `linux/ppc64` ELFv2), and the bootstrap requirement.
- `go-spec-reviewer` — the 1.26/1.27 obsolete-idiom replacements, a bidirectional version gate (never recommend what the module cannot compile, never flag as fresh what it already has), and a modern-language-feature checklist.
- `cobra-viper` — `errors.AsType` for config-not-found inspection, Go 1.27's stricter JSON decoding for `.json` config files with both escape hatches, and `omitzero` versus `omitempty` for config structs.
- `fileflow-pathologize` — `errors.AsType` for structured error inspection, and a recorded audit of `LastIndex`/`CutLast` against extension parsing.
- Repository — `.github/workflows/go-release-check.yml` and `.github/scripts/check_go_release.py`: a monthly scheduled check that fails when a skill's `current through Go X.Y` claim falls behind the latest stable Go release, opening or updating a single standing `Go release drift` issue.

### Changed

- `go` — the frontmatter claim moved from "current through Go 1.25" to "current through Go 1.27", in the same changes that added the release's content.
- `CONTRIBUTING.md` — documents the frontmatter claim the release-currency check parses, and lists the check among the ones you can run locally.

[1.1.0]: https://github.com/spf13/go-skills/releases/tag/v1.1.0

## [1.0.0] - 2026-09-17

### Added

- `go` — idiomatic Go patterns: package design, error handling, interfaces, concurrency, testing, and project layout.
- `cobra-viper` — CLI architecture with Cobra and Viper: command-first design, 12-factor configuration, environment-variable binding, and in-memory CLI testing.
- `go-spec-reviewer` — review a Go design document before implementation begins.
- `go-release` — release engineering: semantic versioning, breaking-change detection, deprecation, `go.mod` hygiene, and GoReleaser.
- `wails` — desktop applications in Go with Wails v2 (stable) and v3 (alpha).
- `fileflow-pathologize` — safe cross-filesystem file operations with `spf13/fileflow` and `spf13/pathologize`.
- Claude Code marketplace support: `/plugin marketplace add spf13/go-skills`.
- `LICENSE` (MIT), `CONTRIBUTING.md`, and a new-skill scaffold under `templates/skill/`.
- CI validating the manifests, the cross-file invariants, and documentation links.

[1.0.0]: https://github.com/spf13/go-skills/releases/tag/v1.0.0
