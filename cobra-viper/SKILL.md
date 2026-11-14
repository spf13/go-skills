---
name: cobra-viper
description: >
  Use when a Go CLI is built, reviewed, or refactored with Cobra and Viper — "cobra", "viper", a
  subcommand, a persistent or local flag, flag parsing, a viper config file, environment or flag
  binding, or a test that executes a `cobra.Command`. Not for general Go package design (use go),
  shipping a CLI binary (use go-release), or pre-implementation spec review (use
  go-spec-reviewer).
---

# Go CLI Architecture: Cobra & Viper

Idiomatic patterns and best practices for building robust, configuration-driven command-line interfaces using Cobra and Viper.

**Respect the module's `go` directive.** Read it in `go.mod` before recommending any version-gated feature named in this skill — a module declaring `go 1.21` must still compile on Go 1.21, whatever toolchain you happen to be running. If the floor is unknown, read `go.mod` or ask; never guess, and never assume the reader is on the latest release. Never raise the directive merely to reach a convenience API quoted here — a `go` directive is a floor imposed on every dependent of the module. Cobra and Viper API availability is a second axis: gate that on the versions pinned in `go.mod`, not on the language version. See the `go` skill's **Version Gating**.

## When to Activate

- Writing a new CLI application in Go
- Adding commands, subcommands, or flags to an existing Cobra application
- Integrating Viper for configuration file, environment variable, or flag management
- Reviewing or refactoring CLI code that uses Cobra and/or Viper
- Designing the command structure or configuration schema for a CLI tool
- Testing CLI commands

## When NOT to Use This Skill

For general Go package design, testing, and idioms use `go`. For release mechanics — tagging, versioning, and shipping a CLI binary — use `go-release`. For spec review use `go-spec-reviewer`.

## Core Philosophy

### The Command-First Architecture

Treat your application binary as a router for commands. The CLI framework (Cobra) should solely handle flags, arguments, and routing. Your core business logic should remain completely unaware of the CLI layer, making it highly testable and reusable.

### Unified Configuration

Configuration should be environment-aware and unified. Viper acts as the single source of truth, merging defaults, config files, environment variables, and command-line flags into a cohesive state before passing it to the application logic.

### Commands Are Built, Not Declared

Construct the command tree with factory functions (`NewRootCmd()`), not package-level `var` declarations. Globals make commands untestable (state leaks between test runs) and unusable as a library.

**The globals boundary — there is exactly one, and it does not scale with project size.** Package-level state is legitimate in exactly two cases:

1. **Build-time constants** stamped by `-ldflags`: `var version, commit, date string` (see *Version Handling*).
2. **`main.go` itself**, which holds nothing but `run()` and `os.Exit(run())` — no commands, no flags, no Viper.

A `*cobra.Command`, a flag, or a `*viper.Viper` declared at package level is always wrong: in a two-command tool as much as in a forty-command one. This rule has no small-tool exception — a tool small enough that factories feel like overhead is also small enough that a single in-process test pays for them, and a tool that will never be tested still has to be readable, which package-level wiring is not.

## CLI Package Organization

**Anti-Pattern:** Hiding all your commands and core logic deep inside an `internal/` directory tree, or shoving everything into `main.go`.

### Discoverable, Flat Structures

Command routing and business logic should live in standard, logically named packages. The `cmd/` package handles the CLI surface area, while other top-level packages handle the domain logic.

**Layout rule.** The general rule is owned by the `go` skill ("Package Organization: Flat by Default"): one binary means `main.go` at the module root, and `cmd/<name>/` only appears when the module ships more than one binary. This skill owns the single exception, and it applies to every CLI here. A CLI with subcommands puts its commands in one bare `cmd/` package even when it ships a single binary. The exception exists because the `cmd/` directory is keyed to the CLI surface, not the binary count: each command is a separate entry path with its own flags, help text, and Viper bindings, and keeping them at the root would put the CLI's public surface in the same package as the domain logic this skill insists on decoupling. One binary therefore means exactly one `cmd/`, and it is a library package holding the command factories — `main.go` stays at the root as the minimal shim shown below. A CLI with no subcommands earns neither: flags and one action go in `main.go` at the root.

```
mycli/
├── main.go               # Minimal entry point: run() + os.Exit(run())
├── cmd/                  # The Cobra routing layer
│   ├── root.go           # Root command factory, global flags, Viper setup
│   ├── serve.go          # The 'serve' subcommand
│   └── build.go          # The 'build' subcommand
├── engine/               # Core business logic (name based on your domain)
│   ├── server.go
│   └── compiler.go
├── go.mod
└── go.sum
```

`main.go` is intentionally minimal, and it is the **only** place `os.Exit` appears:

```go
package main

import (
    "context"
    "fmt"
    "os"
    "os/signal"
    "syscall"

    "github.com/spf13/myapp/cmd"
)

func main() {
    os.Exit(run())
}

// run owns every defer. os.Exit skips deferred calls, so main must not exit
// until run has returned and stop() has already fired.
func run() int {
    ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
    defer stop()

    if err := cmd.NewRootCmd().ExecuteContext(ctx); err != nil {
        fmt.Fprintln(os.Stderr, err) // SilenceErrors: the error is printed exactly once, here
        return 1
    }
    return 0
}
```

Two rules follow from this shape:

- **`os.Exit` is an accepted top-level exception, scoped to `main` after `run()` returns.** It is never called from a command, a `RunE`, a library package, or any function that holds a live `defer` — `os.Exit` unwinds nothing, so a deferred `stop()`, `Close()`, or `Flush()` would be silently dropped. The `main` → `run()` split is what makes the single `os.Exit` safe: by the time it runs, every defer in `run` has already executed.
- **There is exactly one execution path.** No `cmd.Execute()` wrapper: a second entry point that calls `NewRootCmd().Execute()` bypasses the signal context above, and two entry points drift. Callers who need the bare command tree (tests, embedding) call `NewRootCmd()` and drive it themselves.

### The Factory Pattern (`cmd/root.go`)

Each command file exports a constructor. The root factory creates its own Viper instance and wires everything together — no package-level state:

```go
func NewRootCmd() *cobra.Command {
    v := viper.New()

    rootCmd := &cobra.Command{
        Use:           "mycli",
        Short:         "A brief description",
        SilenceUsage:  true,
        SilenceErrors: true,
        PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
            return initConfig(v, cmd)
        },
    }

    rootCmd.PersistentFlags().String("config", "", "config file path")
    rootCmd.PersistentFlags().String("log-level", "info", "log level")

    rootCmd.AddCommand(NewServeCmd(v))
    rootCmd.AddCommand(NewBuildCmd(v))
    return rootCmd
}
```

**No package-level escape hatch here.** The boundary is stated once, under *Commands Are Built, Not Declared*: build-time version strings and `main.go` — nothing else. A factory that feels heavy for a small tool is still the factory to write.

### Decouple Commands from Execution

The files in your `cmd/` package should do exactly three things:

1. Define the Cobra command, its aliases, and help text.
2. Bind Viper flags and configuration for that specific command.
3. Call a function in your core logic package (e.g., `engine`), passing in the parsed configuration and the command context.

Your core logic (the `engine` package) should have absolutely **zero** imports from `github.com/spf13/cobra` or `github.com/spf13/viper`.

## Cobra Best Practices

Flag parsing, context plumbing, and the command tree below do not change with the Go release — nothing in Go 1.26 or Go 1.27 alters how Cobra parses flags, how `cmd.Context()` flows into your code, or how commands are constructed and added. The version-gated guidance in this skill lives only in error inspection and config decoding, and is called out where it appears; each of those recommendations is bounded by the module's `go` directive (see the floor-first rule above).

### 1. Use `RunE` for Native Error Handling

Avoid `Run`. If a command fails, use `RunE` to return the error up the execution chain. This allows the root command to handle errors gracefully and consistently, rather than relying on scattered `log.Fatal` calls that bypass `defer` statements.

```go
// Idiomatic: Returning errors to be handled by the executor
func NewServeCmd(v *viper.Viper) *cobra.Command {
    cmd := &cobra.Command{
        Use:   "serve",
        Short: "Starts the primary application server",
        RunE: func(cmd *cobra.Command, args []string) error {
            var cfg engine.Config
            if err := v.Unmarshal(&cfg); err != nil {
                return fmt.Errorf("decoding config: %w", err)
            }
            return engine.Serve(cmd.Context(), cfg)
        },
    }
    cmd.Flags().String("addr", ":8080", "listen address")
    return cmd
}
```

### 2. Silence Usage on Application Errors

By default, Cobra prints the full help text whenever an error is returned. This is confusing if the error was a runtime failure (like a network timeout) rather than a syntax error. Set `SilenceUsage: true` and `SilenceErrors: true` on the root (as shown above) and let `main.go` print the error once.

### 3. Context-Aware Commands

Modern Go relies heavily on `context.Context` for cancellation and timeouts. Pass the Cobra command's context directly to your business logic:

```go
RunE: func(cmd *cobra.Command, args []string) error {
    return engine.Process(cmd.Context(), args)
}
```

To make that context respect `Ctrl+C`, wire signal handling in `main.go`:

```go
ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
defer stop()
err := cmd.NewRootCmd().ExecuteContext(ctx)
```

### 4. Validate Positional Arguments with `Args`

Never validate argument counts inside `RunE` — Cobra does it for you, before `RunE` runs, with correct usage errors:

```go
cmd := &cobra.Command{
    Use:  "build [target]",
    Args: cobra.ExactArgs(1),
    // Also: cobra.NoArgs, cobra.MinimumNArgs(n), cobra.MaximumNArgs(n),
    // cobra.RangeArgs(min, max), cobra.OnlyValidArgs (with ValidArgs set)
    // Combine: cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs)
}
```

### 5. `PersistentPreRunE` for Shared Setup

Use `PersistentPreRunE` on the root command to run setup (config loading, logging) after flags are parsed but before any subcommand runs. Cobra runs it for every subcommand automatically.

**Gotcha:** if a subcommand also defines `PersistentPreRunE`, it *replaces* the parent's — Cobra does not chain them by default. Either call the parent's explicitly, or opt into chaining globally with `cobra.EnableTraverseRunHooks = true` (Cobra 1.8+).

### 6. Flag Design

```go
// Persistent flags — inherited by all subcommands
rootCmd.PersistentFlags().String("config", "", "config file path")
rootCmd.PersistentFlags().BoolP("verbose", "v", false, "enable verbose output")

// Local flags — only for this command
serveCmd.Flags().String("addr", ":8080", "listen address")

// Required flags — Cobra validates before RunE is called
serveCmd.Flags().String("name", "", "required name")
serveCmd.MarkFlagRequired("name")

// Flag relationships (Cobra 1.8+ for OneRequired)
serveCmd.MarkFlagsMutuallyExclusive("json", "yaml")
serveCmd.MarkFlagsRequiredTogether("user", "password")
serveCmd.MarkFlagsOneRequired("file", "url")
```

- Use `PersistentFlags` for cross-cutting concerns (config, verbosity, output format).
- Use `Flags` for command-specific options.
- Always provide short flags (`-v`, `-o`) for common options.


For help grouping, shell completion, and output capture: [references/cobra-advanced.md](references/cobra-advanced.md).

Renaming a flag or folding a command into another is a breaking change for every script, CI job, and completion file in the wild. Cobra and pflag carry the deprecation machinery: the old surface keeps working while announcing its successor.

```go
// Deprecate a flag. It still parses; pflag prints the message on use and help
// annotates the entry "(DEPRECATED: use --addr)".
if err := serveCmd.Flags().MarkDeprecated("port", "use --addr"); err != nil {
    return err
}

// Deprecate only the shorthand, keeping the long form alive.
if err := serveCmd.Flags().MarkShorthandDeprecated("verbose", "use --verbose"); err != nil {
    return err
}

// Hide a flag entirely: still accepted, absent from help and completion.
// Use for debug or internal escape hatches, not for anything users were told to type.
if err := serveCmd.Flags().MarkHidden("legacy-mode"); err != nil {
    return err
}

// Deprecate a whole command. It still runs, prints its message via cmd.Printf
// (so SetOut captures it), and drops out of help and completion listings.
migrateCmd := &cobra.Command{
    Use:        "migrate",
    Deprecated: "use \"myapp db migrate\"",
    RunE:       runMigrate,
}
```

`MarkDeprecated`/`MarkHidden` are `*pflag.FlagSet` methods via `cmd.Flags()` — don't discard the error. A deprecated command announces itself when run; a hidden one doesn't, but neither appears in listings.

**Migration-path rule:** never deprecate without naming the replacement, and never remove on the same release that deprecated. The message is a migration instruction — `MarkDeprecated("port", "use --addr")`, not `MarkDeprecated("port", "")`; a bare "deprecated" is a bug report from your users, not guidance. Keep the old form parsing for at least one minor release after the replacement ships, then delete it in a major version, alongside the deprecation's own removal.

## Viper Configuration Patterns

### 12-Factor Configuration

The environment supplies the values; the binary supplies only publishable defaults ([factor III](https://12factor.net/config)). One `v.SetEnvPrefix("MYAPP")` + `SetEnvKeyReplacer` gives every key a predictable env name (§4); precedence is `Set()` → flags → env → file → defaults (§3); secrets never ship as defaults; `--help` documents the knobs.

### 1. Prefer a Viper Instance over the Global Singleton

`viper.New()` gives you an instance you can inject, isolate in tests, and run concurrently. The package-level `viper.Get*` singleton exists for convenience in small tools — the same trade-off as pflag's default FlagSet. For anything with tests or multiple commands, pass `*viper.Viper` explicitly (as the factory examples above do).

### 2. Unmarshal into Typed Structs

Define a typed config struct, unmarshal at the routing layer, and pass that struct down — not `v.GetString` scattered through business logic.

```go
type Config struct {
    Host string `mapstructure:"host"`
    Port int    `mapstructure:"port"`
}

var cfg Config
if err := v.Unmarshal(&cfg); err != nil {
    return fmt.Errorf("unable to decode config: %w", err)
}
```

A config struct that is also *marshaled* needs JSON tags. Use `omitzero` (Go 1.24) over `omitempty` for struct-typed fields (`time.Time`, nested structs): `omitempty` never omits a struct, so a zero `time.Time` emits `"0001-01-01T00:00:00Z"`; `omitzero` omits it:

```go
type Config struct {
    Host     string        `mapstructure:"host"     json:"host"`
    Started  time.Time     `mapstructure:"started"  json:"started,omitzero"` // omitempty would emit the zero time
    Listen   ListenConfig  `mapstructure:"listen"   json:"listen,omitzero"`  // omitempty never omits a struct
}
```

### 3. The Binding Hierarchy

Viper merges configuration sources in this order (highest → lowest priority):

1. **Explicit `Set()`** calls in code
2. **Flags** (bound via `BindPFlag` / `BindPFlags`)
3. **Environment variables** (`MYCLI_PORT`)
4. **Config file** (`~/.mycli.yaml`, `./.mycli.yaml`)
5. **Defaults** (`v.SetDefault`)

Bind a whole flag set at once in `PersistentPreRunE` (not `init()` — avoids the collision where two subcommands' `init()` bind the same key and the last one silently wins):

```go
if err := v.BindPFlags(cmd.Flags()); err != nil { return err }
```

### 4. Environment Variable Mapping

With `v.SetEnvPrefix("MYAPP")` and `v.AutomaticEnv()`:

| Viper key | Environment variable |
|-----------|---------------------|
| `log-level` | `MYAPP_LOG_LEVEL` |
| `serve.addr` | `MYAPP_SERVE_ADDR` |
| `db.password` | `MYAPP_DB_PASSWORD` |

Nested keys with dots or dashes require a replacer to map correctly to env var names:

```go
v.SetEnvKeyReplacer(strings.NewReplacer("-", "_", ".", "_"))
```

**Critical gotcha — `AutomaticEnv` + `Unmarshal`:** `Unmarshal` only walks keys Viper already knows about (from defaults, config file, or explicit binds). A value set *only* via environment variable is invisible to `Unmarshal` unless the key was registered. The fix: call `v.SetDefault` for every key in your config struct, or `v.BindEnv` each key explicitly. This is the most common "my env var is ignored" bug in Viper applications.

### 5. Config File Setup

```go
func initConfig(v *viper.Viper, cmd *cobra.Command) error {
    if cfgFile, _ := cmd.Flags().GetString("config"); cfgFile != "" {
        v.SetConfigFile(cfgFile)
    } else {
        home, err := os.UserHomeDir()
        if err != nil {
            return err
        }
        v.AddConfigPath(home)
        v.AddConfigPath(".")
        v.SetConfigType("yaml")
        v.SetConfigName(".myapp")
    }

    v.SetEnvPrefix("MYAPP")
    v.SetEnvKeyReplacer(strings.NewReplacer("-", "_", ".", "_"))
    v.AutomaticEnv()

    if err := v.ReadInConfig(); err != nil {
        // Go 1.26+: generic, type-safe, and faster than errors.As.
        if _, ok := errors.AsType[viper.ConfigFileNotFoundError](err); !ok {
            return fmt.Errorf("reading config: %w", err)
        }
        // Not found is fine — defaults, env, and flags still apply
    }

    return v.BindPFlags(cmd.Flags())
}
```

`errors.AsType[T](err)` (Go 1.26 and above) returns the matched value and an `ok` bool, so it needs no predeclared target variable and cannot be misused by passing a non-pointer. Below Go 1.26, use the classic form with `errors.As`, which remains correct and is the version-floored fallback:

```go
var notFound viper.ConfigFileNotFoundError
if !errors.As(err, &notFound) {
    return fmt.Errorf("reading config: %w", err)
}
```

Use YAML as the default format — it's readable and supports nesting:

```yaml
# ~/.myapp.yaml
log-level: debug

serve:
  addr: ":9090"
  port: 9090

db:
  host: localhost
  port: 5432
  name: myapp
```

### 6. JSON Config Files Under Go 1.27's `encoding/json`

Go 1.27 runs `encoding/json` on the `encoding/json/v2` implementation, but the v1 API keeps v1 behavior: duplicate object names still take the last value, invalid UTF-8 still becomes U+FFFD, and only decode error *text* may change — match by type, never by message. Viper decodes `.json` through `encoding/json`, so an existing config still loads. Strict decoding belongs to the new `encoding/json/v2` package (Go 1.27): to make a duplicated key a startup error, validate the raw bytes with it before `ReadConfig`. Validation snippet, `GOEXPERIMENT=nojsonv2`, and the `mapstructure` note: [references/viper-config.md](references/viper-config.md).

### 7. Writing Configuration Back

Viper can serialize its merged state: `WriteConfig`/`SafeWriteConfig`/`WriteConfigAs`/`SafeWriteConfigAs`, with the safe variants using `O_EXCL`. **First-run pitfall:** `WriteConfig`/`SafeWriteConfig` write to a file Viper has found or been named — on a first run with no file, they fail with `ConfigFileNotFoundError`. Resolve the path and use `SafeWriteConfigAs` explicitly. Never write back on a read path; make persistence a command. Full table, code, and limits: [references/viper-config.md](references/viper-config.md).

## Version Handling

Cobra has a built-in version mechanism — prefer it over a subcommand:

```go
// Set at build time via -ldflags: var (version, commit, date string)
rootCmd := &cobra.Command{
    Use:     "myapp",
    Version: fmt.Sprintf("%s (commit:%s %s)", version, commit, date),
}
// `myapp --version` works; customize with `SetVersionTemplate`.
```

```bash
go build -ldflags="-X 'github.com/spf13/myapp/cmd.version=1.2.3' \
                   -X 'github.com/spf13/myapp/cmd.commit=$(git rev-parse --short HEAD)' \
                   -X 'github.com/spf13/myapp/cmd.date=$(date -u +%Y-%m-%dT%H:%M:%SZ)'"
```

Add a `version` *subcommand* only if you need structured output (`version --json`) — and print via `cmd.OutOrStdout()`, never `fmt.Printf`.

## Testing CLI Commands

**Anti-Pattern:** Testing CLI commands by compiling the binary and using `os/exec`. This is extremely slow, brittle, and makes it difficult to measure test coverage.

**Second anti-pattern:** executing a package-level subcommand directly. `cmd.Execute()` always executes from the *root* of the command tree, and mutating a global command leaks flag state between tests. This is exactly why commands are built by factories.

```go
// Idiomatic: In-memory CLI testing via the factory
func TestServeCommand(t *testing.T) {
    tests := []struct {
        name    string
        args    []string
        want    string
        wantErr bool
    }{
        {"default port", []string{"serve"}, "listening on :8080", false},
        {"custom port", []string{"serve", "--addr", ":9090"}, "listening on :9090", false},
        {"bad flag", []string{"serve", "--bogus"}, "", true},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            buf := new(bytes.Buffer)

            root := cmd.NewRootCmd() // fresh tree + fresh Viper every test
            root.SetOut(buf)
            root.SetErr(buf)
            root.SetArgs(tt.args) // full CLI invocation, exactly as a user types it

            err := root.ExecuteContext(t.Context())
            if (err != nil) != tt.wantErr {
                t.Fatalf("Execute() error = %v, wantErr %v", err, tt.wantErr)
            }
            if !strings.Contains(buf.String(), tt.want) {
                t.Errorf("output = %q, want substring %q", buf.String(), tt.want)
            }
        })
    }
}
```

- **Always execute through the root** with `SetArgs` containing the subcommand name — that's the real code path users hit, including flag parsing and `PersistentPreRunE`.
- A fresh `NewRootCmd()` per test case means no shared state, no `viper.Reset()`, and tests can run in parallel.
- If you're stuck with the global Viper singleton, call `viper.Reset()` between tests — its state bleeds across test cases.
- Use `SetOut` / `SetErr` to capture output — which only works if commands print via `cmd.OutOrStdout()`.
- Set config env vars with `t.Setenv` — it restores them automatically and blocks accidental `t.Parallel` misuse.

## Common Mistakes

- **Executing a subcommand variable directly in tests**: `serverCmd.Execute()` runs the whole tree from the root, not just that command. Build fresh trees with a factory and drive them through `SetArgs`.
- **Env-only values missing after `Unmarshal`**: `AutomaticEnv` doesn't register keys. `SetDefault` or `BindEnv` every key in your config struct (see the gotcha above).
- **Assuming Go 1.27 made JSON config decoding strict**: `encoding/json` still accepts duplicate object names (last wins) and invalid UTF-8; only `encoding/json/v2` rejects them. If a duplicated key should stop the command, validate with `encoding/json/v2` before `ReadConfig` (see §6 above).
- **Deprecating without a migration path**: `MarkDeprecated(name, "")` and `Deprecated: "deprecated"` leave users with a warning and no replacement, and deleting in the same release strands every script. Name the successor and keep the old form parsing until the next major (Cobra §10).
- **`WriteConfig` on a first run**: it writes to a path Viper *found*, so with no config file yet it returns the very `ConfigFileNotFoundError` your loader ignored. Resolve the path and use `SafeWriteConfigAs`/`WriteConfigAs` for `init`-style commands (Viper §7).
- **Accessing Viper before config is loaded**: values are empty until your setup has run (`PersistentPreRunE` or `cobra.OnInitialize`). Don't read Viper in `init()` functions or `var` blocks.
- **Forgetting `BindPFlags`**: flags are not automatically visible to Viper. Bind the command's flag set explicitly.
- **Missing `SetEnvKeyReplacer`**: nested keys with dots (e.g., `serve.addr`) won't match `MYAPP_SERVE_ADDR` without a replacer.
- **Cobra/Viper imports in business logic**: the `engine` package must never import Cobra or Viper. Pass typed config structs instead.
- **`fmt.Printf` inside commands**: bypasses `SetOut`, breaking test capture. Use `cmd.OutOrStdout()`.
- **Over-nesting subcommands**: two levels (`app command subcommand`) is usually the right depth. Three or more levels confuse users.
