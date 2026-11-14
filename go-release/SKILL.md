---
name: go-release
description: >
  Use when the user cuts, plans, or reviews a Go release — a version tag, semver, breaking-change
  detection, a v2 module path, deprecating an API, go.mod hygiene before tagging, a changelog, or
  shipping binaries with GoReleaser. Also use when an exported API change to a published module
  needs a patch, minor, or major decision. Not for general Go code (use go), CLI construction (use
  cobra-viper), or design-doc review (use go-spec-reviewer).
---

# Go Release Engineering: Versions, Compatibility, and Distribution

A release is a promise. Once a version is tagged and fetched through the module proxy, it is immutable and someone depends on it forever. This skill covers how to make that promise correctly — for libraries (API compatibility) and applications (binary distribution).

## Scope

Two artifact kinds, and only these: **published Go modules** (tags fetched through the module proxy) and **shipped binaries and container images**. Both are promises to strangers, so both get the same discipline.

This skill does not cover releasing the skill catalog this file ships in. That release is Markdown plus a version bump, a `CHANGELOG.md` entry, and an annotated tag, governed by the catalog's `CONTRIBUTING.md` — the module-proxy and artifact rules below do not apply to it. Reviewing an unshipped design is `go-spec-reviewer`'s job; this skill starts when published API has to become a version number.

## When to Activate

- Cutting or planning a release of a Go module or binary
- Reviewing whether a change is breaking for a published library
- Deprecating or removing exported API
- Editing `go.mod` in a way that affects dependents (go directive, dependencies)
- Setting up binary distribution (GoReleaser, ldflags, multi-platform builds)
- Recovering from a bad release

## When NOT to Use This Skill

For writing Go packages use `go`. For CLI patterns use `cobra-viper`. For spec review use `go-spec-reviewer`. For desktop packaging use `wails` with `go-release` for versioning.

## The Compatibility Promise

### Versioning Semantics in Go

- **v0.x.y** — no compatibility promise. Break freely, but document breaks in release notes. Staying on v0 indefinitely is a cop-out for a widely-used library: dependents deserve a promise.
- **v1.x.y** — the promise: code written against v1.0.0 compiles against every later v1 release. Minor versions add, patches fix. Nothing is removed or changed incompatibly. Ever.
- **v2+** — a *different module* with a different import path (`module github.com/you/lib/v2`, imported as `.../lib/v2`). This is semantic import versioning and it is not optional — a v2 tag without the `/v2` module path is broken for module-mode consumers.

**Plan to stay on v1 forever.** A major version bump splits your ecosystem: every dependent must migrate manually, and transitive dependencies mix v1 and v2 in the same build. Cobra, pflag, and the standard library have held v1 for a decade by expanding rather than replacing. Reach for v2 only when the API is genuinely unfixable additively — and treat it as launching a new library, because that's what it costs.

### What Is Actually a Breaking Change

LLMs and humans both misjudge this. The compiler-visible breaks:

| Change | Breaking? | Why |
|--------|-----------|-----|
| Removing/renaming any exported identifier | **Yes** | Obvious |
| Changing a function/method signature | **Yes** | Includes adding a variadic parameter — the signature changes |
| **Adding a method to an exported interface** | **Yes** | Every external implementor no longer satisfies it |
| Adding a field to a struct | Usually no | **Breaks callers using unkeyed composite literals** and exact-type assignments; safe if you've documented keyed-literals-only |
| Adding a method to a struct | No | Unless it collides with a method promoted from an embedded type in *their* code — rare, accepted risk |
| Adding a new function, type, or package | No | This is what minor versions are for |
| Changing an untyped constant's value | **Yes** | Compiled into caller binaries |
| Tightening input validation | **Behaviorally yes** | Code that worked now errors — treat as breaking |
| Changing error *messages* | No | But breaking for anyone string-matching; give them `errors.Is`-able sentinels instead |
| Raising the `go` directive in go.mod | **Effectively yes** | Dependents on older toolchains can no longer build you — see go.mod hygiene |

**The interface trap deserves emphasis:** if your library exports an interface that users implement, it is frozen at v1. Design for this — either keep exported interfaces tiny and stable, or add an unexported method so only you can implement it, or accept functions/structs instead of defining the interface at all (interfaces belong to consumers — see the `go` skill).

### Verify Mechanically, Not by Eyeball

```bash
# Before every tag — compares against the latest released version
go run golang.org/x/exp/cmd/gorelease@latest

# Or compare two specific API surfaces
go run golang.org/x/exp/cmd/apidiff@latest old/ new/
```

`gorelease` reports incompatible changes and suggests the correct next version number. Run it in CI on every PR that touches a published module's exported identifiers — catching a break at review time is free; catching it after the tag costs a major version or a retraction. For an application module with no importers there is nothing to compare, so skip it there.

## Deprecation: The Additive Escape Hatch

You can't remove, but you can deprecate. The convention is machine-readable and respected by gopls, staticcheck, and pkg.go.dev:

```go
// Deprecated: Use [ParseContext] instead, which supports cancellation.
// This function will be maintained but not extended.
func Parse(input string) (*Result, error) { ... }
```

- The paragraph must begin exactly with `Deprecated:` to be recognized.
- Deprecated API **keeps working forever** on v1 — deprecation redirects new code; it does not license removal.
- Always name the replacement. A deprecation without a migration path is just an insult.
- Deprecate in a minor release, never silently in a patch.

### Ship the Migration, Not Just the Replacement Name

Naming a replacement is only half a deprecation. On **Go 1.26+**, `go fix` is the home of Go's *modernizers* — a `go vet`-style analyzer suite plus a source-level inliner driven by `//go:fix inline` directives (the historical fixers were removed in 1.26). Annotate a thin deprecated wrapper and `go fix` rewrites callers to the replacement for them:

```go
// Deprecated: Use [ParseContext] instead, which supports cancellation.
//
//go:fix inline
func Parse(input string) (*Result, error) {
	return ParseContext(context.Background(), input)
}
```

- Keep the directive on a wrapper whose body is a single delegation to the replacement; the inliner rewrites each call site in terms of the replacement.
- Put `go fix ./...` in the release notes as the migration step for dependents on Go 1.26+. A deprecation that ships an executable rewrite gets cleared in one commit; one that ships only a prose suggestion hangs around for years.
- **Go 1.27** adds the `atomictypes`, `embedlit`, `slicesbackward`, and `unsafefuncs` modernizers, and renames the `waitgroup` analyzer to `waitgroupgo`. The `fmtappendf` analyzer was removed on stylistic grounds, so it no longer fires — don't wait on it.

## go.mod Hygiene for Releases

Your go.mod is part of your API. Through minimal version selection, your minimums become your dependents' minimums.

- **The `go` directive is a floor you impose on every dependent.** Set it to the oldest version you actually need, not the newest you happen to run. Bumping it just to use one new stdlib helper forces an upgrade on your whole downstream tree — for a library, that's a decision, not a default. On **Go 1.26+**, `go mod init` already errs downward: a `1.N.X` toolchain writes `go 1.(N-1).0`, and a pre-release of `1.N` writes `go 1.(N-2).0`. The generated floor is conservative by construction — never assume the directive matches the toolchain installed on your machine, and raise it only deliberately, with the change called out in the release notes.
- **The `toolchain` directive is local preference, not a floor.** It records the toolchain you'd like to build with and does not constrain dependents. Pinning one in CI is fine, but keep a reachable bootstrap path: Go 1.26 requires Go 1.24.6 or later to bootstrap, and the stated expectation is that Go 1.28 will require a minor release of Go 1.26 or later.
- **Never tag a release containing `replace` directives.** They are ignored in dependents' builds, so the module you tested is not the module they get. Same for `go.work` assumptions — workspaces don't ship.
- **Run `go mod tidy` before tagging.** A dirty go.mod ships phantom requirements to everyone downstream. On **Go 1.27+**, tidy also merges duplicate `require` blocks for modules whose `go` directive is 1.27 or later, enforcing at most two blocks (direct and indirect) while preserving attached comment blocks. The first tidy after adopting 1.27 can therefore rewrite a large stretch of go.mod — that reorganization is expected, not a review finding. Commit it as its own change so the diff stays readable.
- **Keep dependency minimums low.** Requiring the latest patch of every dependency forces churn on dependents; require the oldest version that has what you need.

## GODEBUG and GOEXPERIMENT: Deadlines, Not Escapes

When Go changes a behavior it cannot change silently, it stages the change behind a `GODEBUG` setting or a `GOEXPERIMENT` flag, then removes the escape hatch a release or two later. The removals are scheduled, not negotiable:

- `asynctimerchan` (added in Go 1.23) is removed permanently in **Go 1.27** — `time` package channels are now always unbuffered.
- The `goroutineleak` profile graduated: experimental via `GOEXPERIMENT=goroutineleakprofile` in **Go 1.26**, generally available in `runtime/pprof` and `/debug/pprof/goroutineleak` in **Go 1.27**, with the GOEXPERIMENT deleted.
- `GOEXPERIMENT=nogreenteagc` (**Go 1.26**, Green Tea GC opt-out) and `GOEXPERIMENT=nosizespecializedmalloc` (**Go 1.27**, size-specialized allocation opt-out) are both announced as temporary. The Green Tea opt-out was slated for removal in Go 1.27 but still works there; the allocation opt-out is expected to go in Go 1.28. Treat both as deadlines either way.
- `GOEXPERIMENT=nojsonv2` (**Go 1.27**) restores the pre-v2 `encoding/json` implementation and is likewise expected to be removed in a future release.

An opt-out is a migration deadline, never a permanent escape:

- **A library cannot pin behavior for its dependents.** A `godebug` stanza in your go.mod or a `//go:debug` comment in your source takes effect for the main module; your dependents' builds do not inherit it. If your release notes rely on an opt-out to keep old behavior for consumers, that promise is already broken — use the window to migrate your own code, and ship the new behavior.
- **Use `godebug` entries to preview the change.** Set the setting to its *new* value in your go.mod `godebug` stanza and develop against the behavior that will become the default — the window exists so you can migrate on your own schedule instead of the release's.
- **Go 1.27 turns stale pins into build failures.** The `go` command now recognizes a `GODEBUG` setting whose support was removed only when it is set to the final default value; a `godebug` entry or `//go:debug` comment pinned to any older value fails the build. That is the enforcement arm of the deadline — pinning a removed escape hatch does not preserve behavior, it stops compiling.
- **Document the removal release in your own notes.** When you adopt a new behavior at the expense of an old one, say which Go release removes the escape hatch and what dependents must change. That is the same migration-path obligation you owe for a deprecated API.

Testing knobs follow the same rule: `testing/cryptotest.SetGlobalRandom` is the supported way to get deterministic cryptographic randomness for a test (**Go 1.26**), with the `cryptocustomrand=1` GODEBUG existing only to restore the old behavior temporarily — do not build test infrastructure on it.

## The Release Checklist

```bash
go mod tidy && git diff --exit-code go.mod go.sum   # no surprises
go build ./...
go vet ./...
go test -race ./...
go run golang.org/x/vuln/cmd/govulncheck@latest ./...
go run golang.org/x/exp/cmd/gorelease@latest        # libraries: API check + version suggestion
```

On Go 1.27+ with a `go 1.27` directive, `go mod tidy` may consolidate duplicate `require` blocks (at most direct and indirect, comments preserved) — expect it, review it, and don't read the reorganization as an unrelated dependency change.

Then:

```bash
git tag -a v1.4.0 -m "v1.4.0"    # annotated tag, full semver, leading v
git push origin v1.4.0
```

- Tag the module root. In a multi-module repo, prefix with the module directory: `submod/v1.4.0`.
- Write release notes for humans: what's new, what's deprecated, any behavior changes. Link migration guidance for anything deprecated.
- For deprecations that shipped `//go:fix inline` wrappers, the migration instruction is one line: dependents on Go 1.26+ run `go fix ./...`. Say it explicitly — the directive does nothing for anyone who doesn't know to run the command.
- Note any platform or minimum-OS floor change in the same release (see below); it is a support-statement change even when no API moved.

### The Proxy Is Forever: Fixing a Bad Release

The module proxy caches every tagged version permanently. **Never delete or re-tag a published version** — the proxy still serves the original, and now your tag and the proxy disagree, which breaks checksum verification for everyone.

The correct fix is `retract`:

```text
// go.mod of a NEW release (e.g. v1.4.2)
retract (
    v1.4.1 // Broke Parse on empty input; use v1.4.2.
)
```

Ship the retraction in the next release. `go get` skips retracted versions and `go list -m -u` warns users already on them. A version can even retract itself (`retract v1.5.0` inside v1.5.0's go.mod, published as v1.5.1... which retracts both). Retraction is advisory withdrawal — the only kind that exists.

## Prereleases: Testing a Release Without Promising It

A semver prerelease tag (`v1.5.0-rc.1`, `v1.5.0-beta.2`) is how real users get a real artifact without the ecosystem inheriting a promise:

- **`go get` will not find it by accident.** Prerelease tags are not selected as `latest` unless the repository has no other semver tags, and `go get -u` never upgrades a dependent onto one. Adoption is explicit: `go get example.com/lib@v1.5.0-rc.1`.
- **Ordering is semver ordering.** `v1.4.9 < v1.5.0-rc.1 < v1.5.0-rc.2 < v1.5.0`. A prerelease belongs to the version it will become — never cut `v1.4.10-rc.1` to "test the next patch" and then tag `v1.5.0`, which is a different version with different ordering.
- **RCs are as immutable as releases.** The proxy caches `v1.5.0-rc.1` exactly as it caches `v1.5.0`. A broken RC gets `-rc.2`; deleting the tag only desynchronizes you from the proxy.
- **Cut one when the risk moves.** Anything in the "yes" column of the breaking-change table, a raised `go` directive, or a bumped dependency minimum deserves an RC — those are the failures that surface in someone else's build, not yours.
- **For binaries, say so at the release.** GoReleaser's `release: {prerelease: auto}` flags any tag carrying a prerelease suffix as a prerelease on the SCM, so it is not served as "latest", and `make_latest: false` keeps it out of `releases/latest`. `{{.Version}}` still carries the suffix into ldflags, so `app --version` prints `v1.5.0-rc.1` and a bug report names the RC.
- **An RC is not a substitute for a version number.** `v1.5.0-rc.1` is a rehearsal, not a home: leave it standing for months and you have a v1.5.0-sized change that `go get` still resolves past, because it keeps handing people v1.4.x.

## Application Releases: Shipping Binaries

For CLIs and services, the release artifact is a binary, and **GoReleaser is the standard** — don't hand-roll cross-compilation matrices in shell scripts or Makefiles.

### Version Injection

Wire build-time version info into the variables your CLI already exposes (see the `cobra-viper` skill's Version handling):

```yaml
# .goreleaser.yaml
builds:
  - env: [CGO_ENABLED=0]
    flags: [-trimpath]
    ldflags:
      - -s -w
      - -X github.com/you/app/cmd.version={{.Version}}
      - -X github.com/you/app/cmd.commit={{.ShortCommit}}
      - -X github.com/you/app/cmd.date={{.CommitDate}}
    goos: [linux, darwin, windows]
    goarch: [amd64, arm64]
```

- `CGO_ENABLED=0` for static, portable binaries — enable cgo only when a dependency truly requires it.
- `-trimpath` strips your local paths from the binary: smaller, reproducible, no leaked `/home/you/src/...` in panics.
- `-s -w` strips symbol tables — standard for release binaries; omit if you want usable core dumps.
- `{{.CommitDate}}`, not `{{.Date}}`, for the build-date variable. `{{.Date}}` is the moment the build ran, so the same tag produces a different binary on every run and no one can reproduce your release from source.

Simpler fallback: `debug.ReadBuildInfo()` returns the module version and VCS revision with zero ldflags when installed via `go install` — a reasonable default before you need GoReleaser at all.

### Distribution Checklist

Work down the list; each item assumes the ones above it.

1. **Checksums.** GoReleaser emits `checksums.txt` by default — publish it. Sign or attest that one file rather than each archive: one signature then covers every artifact you shipped.
2. **Reproducibility.** `-trimpath` is necessary, not sufficient. Pin the toolchain (`toolchain` in go.mod, `GOTOOLCHAIN=local` in the release job), keep `CGO_ENABLED=0`, and — the part everyone skips — **never inject a build clock into the binary**. `{{.Date}}` is the build time, so it makes every build unique; use `{{.CommitDate}}` and the build becomes repeatable from the tag alone. Then verify the claim instead of asserting it:

   ```bash
   git checkout v1.4.0
   CGO_ENABLED=0 go build -trimpath -o /tmp/a .
   CGO_ENABLED=0 go build -trimpath -o /tmp/b .
   cmp /tmp/a /tmp/b && echo reproducible
   ```

3. **SBOM.** GoReleaser catalogs archives with Syft by default (`sboms: [{artifacts: archive}]`, SPDX JSON). Ship those files next to the checksums — they make "does this CVE affect me?" answerable from the release page instead of from your source tree. The archive step does not see container images, but `dockers_v2` attaches an image SBOM by default (`sbom: true`).
4. **Provenance and signing.** Keyless cosign over the checksum file, so the signature names the workflow that produced it:

   ```yaml
   checksum:
     name_template: checksums.txt
   signs:
     - cmd: cosign
       signature: "${artifact}.sigstore.json"
       args: ["sign-blob", "--bundle=${signature}", "${artifact}", "--yes"]
       artifacts: checksum
   ```

   Consumers then verify the publisher's *identity*, not just integrity, and only then the bytes:

   ```bash
   cosign verify-blob --bundle checksums.txt.sigstore.json \
     --certificate-identity 'https://github.com/you/app/.github/workflows/release.yml@refs/tags/v1.4.0' \
     --certificate-oidc-issuer 'https://token.actions.githubusercontent.com' checksums.txt
   sha256sum --check --ignore-missing checksums.txt
   ```

   Add build provenance as well (`actions/attest-build-provenance`, checked with `gh attestation verify`) once downstream users start asking how they know the binary is yours. Don't cargo-cult the full chain into a weekend project; put it in place before a security-sensitive consumer asks.
5. **`go install` must always work.** `go install github.com/you/app@latest` is the zero-infrastructure install path. Keep `main` at the repo root or under `cmd/app` so the path is guessable, and never break it with replace directives.
6. **Containers.** Build images from the binaries GoReleaser already produced — `dockers_v2` (v2.12+; the older `dockers:` is deprecated) or `ko:` for a Dockerfile-free build. Base on `scratch` or `distroless/static` since the binary is static, tag by version *and* commit SHA with `latest` as an alias rather than the only tag, and let `platforms` produce one multi-arch manifest instead of one image per platform users must choose between. A Dockerfile that runs `go build` duplicates the build you already did.
7. **Package managers.** Homebrew tap/formula, nfpm for deb/rpm, Scoop: GoReleaser generates all of these from config — add them when users ask, not before.
8. **GoReleaser version drift.** `.goreleaser.yaml` is versioned configuration: v2 requires `version: 2` at the top and rejects a v1 file outright ("only version: 2 configuration files are supported"). Pin both halves — the action *and* the GoReleaser release it runs (`version: '~> v2'`, not the action's `latest` default) — so a schema change surfaces in a PR instead of at tag time. Check and dry-run before the real run:

   ```bash
   goreleaser check
   goreleaser release --snapshot --clean
   ```

   Deprecation warnings are scheduled work, not noise: in v2 a signer that writes no signature file is a warning, and v3 turns it into an error.

### Platform Floors Belong in the Release Notes

Supported targets and minimum OS versions move every Go release, and a cross-compilation matrix that names a dead target ships artifacts nobody can run. Check these before publishing a matrix or a support statement:

- **macOS:** Go 1.26 is the last release whose toolchain runs on macOS 12 Monterey; **Go 1.27 requires macOS 13 Ventura or later**. If you build `darwin` artifacts in CI, say which floor you promise.
- **`windows/arm` (32-bit) was removed in Go 1.26** — drop it from `goos`/`goarch` in `.goreleaser.yaml`.
- **`linux/ppc64` (big-endian) moved to the ELFv2 ABI in Go 1.27** — cgo and hand-written assembly for that target may need updating before you claim support.
- **`freebsd/riscv64` is marked broken** — don't advertise or ship it.
- Other target changes worth tracking: `linux/riscv64` gained the race detector in Go 1.26, and `s390x` passes arguments and results in registers in Go 1.26.

## Common Mistakes

- **Adding a method to an exported interface in a minor release** — the classic silent break. Run `gorelease`; it catches this.
- **Re-tagging after a botched release** — the proxy already has the old bytes; you've now broken checksums globally. Retract and roll forward.
- **Tagging v2.0.0 without changing the module path to `/v2`** — the tag exists but module-mode consumers can't use it.
- **Bumping the go directive casually** — you just forced a toolchain upgrade on every dependent.
- **Shipping a release with `replace` in go.mod** — dependents get different dependency versions than you tested.
- **Treating error message text as stable API in docs/tests** — export sentinel errors; message wording is yours to change.
- **Hand-rolled `GOOS`/`GOARCH` build loops in CI** — GoReleaser exists; a bespoke matrix is complexity with no payoff.
- **Injecting `{{.Date}}` (build time) into ldflags** — no two builds match, so nobody can rebuild your binary from the tag; use `{{.CommitDate}}`.
- **Letting the GoReleaser version float in CI** — the action's `latest` default turns a config-schema change into a failed release at tag time; pin it and run `goreleaser check` on PRs.
- **Skipping `go mod tidy` before the tag** — the one hygiene step that can't be fixed after publishing.