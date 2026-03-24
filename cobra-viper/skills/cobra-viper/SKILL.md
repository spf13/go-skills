---
name: cobra-viper
description: >
  Expert skill for building CLI applications with Cobra and Viper, authored by spf13 — the
  original creator of both libraries. Covers command-first architecture, decoupled business logic,
  configuration management, environment variable binding, context-aware commands, and in-memory
  CLI testing. Use when building or reviewing any Go CLI application that uses Cobra and/or Viper.
---

# Go CLI Architecture: Cobra & Viper

Idiomatic patterns and best practices for building robust, configuration-driven command-line interfaces using Cobra and Viper.

## When to Activate

- Writing a new CLI application in Go
- Adding commands, subcommands, or flags to an existing Cobra application
- Integrating Viper for configuration file, environment variable, or flag management
- Reviewing or refactoring CLI code that uses Cobra and/or Viper
- Designing the command structure or configuration schema for a CLI tool
- Testing CLI commands

## Core Philosophy

### The Command-First Architecture

Treat your application binary as a router for commands. The CLI framework (Cobra) should solely handle flags, arguments, and routing. Your core business logic should remain completely unaware of the CLI layer, making it highly testable and reusable.

### Unified Configuration

Configuration should be environment-aware and unified. Viper acts as the single source of truth, merging defaults, config files, environment variables, and command-line flags into a cohesive state before passing it to the application logic.

## CLI Package Organization

**Anti-Pattern:** Hiding all your commands and core logic deep inside an `internal/` directory tree, or shoving everything into `main.go`.

### Discoverable, Flat Structures

Command routing and business logic should live in standard, logically named packages. The `cmd/` package handles the CLI surface area, while other top-level packages handle the domain logic.

```
mycli/
├── main.go               # Minimal entry point: strictly calls cmd.Execute()
├── cmd/                  # The Cobra routing layer
│   ├── root.go           # Base command, global flags, and Viper setup
│   ├── serve.go          # The 'serve' subcommand
│   └── build.go          # The 'build' subcommand
├── engine/               # Core business logic (name based on your domain)
│   ├── server.go
│   └── compiler.go
├── go.mod
└── go.sum
```

`main.go` is intentionally minimal:

```go
package main

import "github.com/spf13/myapp/cmd"

func main() {
    cmd.Execute()
}
```

### Decouple Commands from Execution

The files in your `cmd/` package should do exactly three things:

1. Define the Cobra command, its aliases, and help text.
2. Bind Viper flags and configuration for that specific command.
3. Call a function in your core logic package (e.g., `engine`), passing in the parsed configuration and the command context.

Your core logic (the `engine` package) should have absolutely **zero** imports from `github.com/spf13/cobra` or `github.com/spf13/viper`.

## Cobra Best Practices

### 1. Use `RunE` for Native Error Handling

Avoid `Run`. If a command fails, use `RunE` to return the error up the execution chain. This allows the root command to handle errors gracefully and consistently, rather than relying on scattered `log.Fatal` calls that bypass `defer` statements.

```go
// Idiomatic: Returning errors to be handled by the executor
var serverCmd = &cobra.Command{
    Use:   "server",
    Short: "Starts the primary application server",
    RunE: func(cmd *cobra.Command, args []string) error {
        server := engine.NewServer()
        if err := server.Start(); err != nil {
            return fmt.Errorf("server failure: %w", err)
        }
        return nil
    },
}
```

### 2. Silence Usage on Application Errors

By default, Cobra prints the full help text whenever an error is returned. This is confusing if the error was a runtime failure (like a network timeout) rather than a syntax error.

```go
// In cmd/root.go
rootCmd := &cobra.Command{
    Use:           "mycli",
    SilenceUsage:  true, // Don't print help on runtime errors
    SilenceErrors: true, // Allow main.go to handle the error printing
}
```

### 3. Context-Aware Commands

Modern Go relies heavily on `context.Context` for cancellation and timeouts. Pass the Cobra command's context directly to your business logic. This context automatically listens for OS termination signals (like `SIGINT` or `Ctrl+C`).

```go
RunE: func(cmd *cobra.Command, args []string) error {
    // Passes context down for graceful shutdown
    return engine.Process(cmd.Context(), args)
}
```

### 4. `PersistentPreRunE` for Shared Setup

Use `PersistentPreRunE` on the root command to run setup (logging, config validation) after flags are parsed but before any subcommand runs:

```go
rootCmd = &cobra.Command{
    Use:               "myapp",
    PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
        // Safe to read viper values here — flags + config + env are all merged
        return setupLogger(viper.GetString("log-level"))
    },
}
```

Cobra runs `PersistentPreRunE` for every subcommand automatically. If a subcommand also defines `PersistentPreRunE`, you must call the parent's explicitly — Cobra does not chain them automatically.

### 5. Flag Design

```go
// Persistent flags — inherited by all subcommands
rootCmd.PersistentFlags().String("config", "", "config file path")
rootCmd.PersistentFlags().Bool("verbose", false, "enable verbose output")

// Local flags — only for this command
serveCmd.Flags().String("addr", ":8080", "listen address")

// Required flags — Cobra validates before RunE is called
serveCmd.Flags().String("name", "", "required name")
serveCmd.MarkFlagRequired("name")

// Mutually exclusive flags
serveCmd.MarkFlagsMutuallyExclusive("json", "yaml")
```

- Use `PersistentFlags` for cross-cutting concerns (config, verbosity, output format).
- Use `Flags` for command-specific options.
- Always provide short flags (`-v`, `-o`) for common options.

### 6. Shell Completion

Cobra generates shell completion for free:

```go
// Custom completions for a flag
serveCmd.RegisterFlagCompletionFunc("output", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
    return []string{"json", "yaml", "table"}, cobra.ShellCompDirectiveNoFileComp
})
```

```bash
myapp completion bash        > /etc/bash_completion.d/myapp
myapp completion zsh         > "${fpath[1]}/_myapp"
myapp completion fish        > ~/.config/fish/completions/myapp.fish
myapp completion powershell  | Out-File -Encoding utf8 "$PROFILE\myapp.ps1"
```

## Viper Configuration Patterns

### 1. Unmarshal into Typed Structs

**Anti-Pattern:** Calling `viper.GetString("database.host")` deep inside your business logic. This tightly couples your domain to Viper and scatters magic strings throughout your codebase.

Instead, define a strongly-typed configuration struct, unmarshal Viper's state into it at the routing layer (`cmd/`), and pass that struct down.

```go
type Config struct {
    Host string `mapstructure:"host"`
    Port int    `mapstructure:"port"`
}

func initConfig() (*Config, error) {
    var cfg Config
    if err := viper.Unmarshal(&cfg); err != nil {
        return nil, fmt.Errorf("unable to decode config: %w", err)
    }
    return &cfg, nil
}
```

### 2. The Binding Hierarchy

Viper seamlessly merges configuration sources in this order (highest → lowest priority):

1. **Explicit `Set()`** calls in code
2. **Flags** (bound via `BindPFlag`)
3. **Environment variables** (`MYCLI_PORT`)
4. **Config file** (`~/.mycli.yaml`, `./.mycli.yaml`)
5. **Defaults** (`viper.SetDefault`)

You must explicitly bind each source. Binding environment variables is crucial for containerized deployments:

```go
func init() {
    // 1. Define the flag
    rootCmd.PersistentFlags().Int("port", 8080, "Server port")

    // 2. Bind the flag to Viper
    viper.BindPFlag("port", rootCmd.PersistentFlags().Lookup("port"))

    // 3. Enable environment variables (e.g., MYCLI_PORT)
    viper.SetEnvPrefix("mycli")
    viper.AutomaticEnv()

    // 4. Set fallback defaults
    viper.SetDefault("port", 8080)
}
```

### 3. Environment Variable Mapping

With `viper.SetEnvPrefix("MYAPP")` and `viper.AutomaticEnv()`:

| Viper key | Environment variable |
|-----------|---------------------|
| `log-level` | `MYAPP_LOG_LEVEL` |
| `serve.addr` | `MYAPP_SERVE_ADDR` |
| `db.password` | `MYAPP_DB_PASSWORD` |

Nested keys with dots or dashes require a replacer to map correctly to env var names:

```go
viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_", ".", "_"))
```

### 4. Config File Setup (`cmd/root.go`)

```go
func initConfig() {
    if cfgFile != "" {
        viper.SetConfigFile(cfgFile)
    } else {
        home, err := os.UserHomeDir()
        cobra.CheckErr(err)

        viper.AddConfigPath(home)
        viper.AddConfigPath(".")
        viper.SetConfigType("yaml")
        viper.SetConfigName(".myapp")
    }

    viper.SetEnvPrefix("MYAPP")
    viper.AutomaticEnv()

    if err := viper.ReadInConfig(); err != nil {
        if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
            fmt.Fprintf(os.Stderr, "Error reading config: %v\n", err)
            os.Exit(1)
        }
    }
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

## Version Command

```go
// Set at build time via -ldflags
var (
    version = "dev"
    commit  = "none"
    date    = "unknown"
)

var versionCmd = &cobra.Command{
    Use:   "version",
    Short: "Print version information",
    Run: func(cmd *cobra.Command, args []string) {
        fmt.Printf("myapp %s (commit: %s, built: %s)\n", version, commit, date)
    },
}
```

```bash
go build -ldflags="-X 'github.com/spf13/myapp/cmd.version=1.2.3' \
                   -X 'github.com/spf13/myapp/cmd.commit=$(git rev-parse --short HEAD)' \
                   -X 'github.com/spf13/myapp/cmd.date=$(date -u +%Y-%m-%dT%H:%M:%SZ)'"
```

## Testing CLI Commands

**Anti-Pattern:** Testing CLI commands by compiling the binary and using `os/exec`. This is extremely slow, brittle, and makes it difficult to measure test coverage.

Because Cobra commands are just Go structs, you can test them directly in memory by redirecting their inputs, outputs, and arguments.

```go
// Idiomatic: In-memory CLI testing
func TestServerCommand(t *testing.T) {
    // Reset viper state between tests — Viper is a singleton; test pollution is real
    viper.Reset()

    buf := new(bytes.Buffer)
    cmd := serverCmd
    cmd.SetOut(buf)
    cmd.SetErr(buf)

    // Pass arguments exactly as a user would on the command line
    cmd.SetArgs([]string{"--port", "9090"})

    if err := cmd.Execute(); err != nil {
        t.Fatalf("unexpected error: %v", err)
    }

    if !strings.Contains(buf.String(), "Starting server on 9090") {
        t.Errorf("expected output to contain port 9090, got: %s", buf.String())
    }
}
```

- Always `viper.Reset()` between tests — Viper global state bleeds across test cases.
- Use `cmd.SetOut` / `cmd.SetErr` to capture output without monkey-patching `os.Stdout`.
- Never test via compiled binary + `os/exec`; use in-memory execution for speed and coverage.

## Common Mistakes

- **Accessing Viper before `initConfig`**: Viper values are empty until `cobra.OnInitialize` callbacks have run. Don't read Viper in `init()` functions or `var` blocks.
- **Forgetting `BindPFlag`**: Flags are not automatically visible to Viper. You must bind them explicitly.
- **Missing `SetEnvKeyReplacer`**: Nested keys with dots (e.g., `serve.addr`) won't match `MYAPP_SERVE_ADDR` without a replacer.
- **Cobra/Viper imports in business logic**: The `engine` package must never import Cobra or Viper. Pass typed config structs instead.
- **Mutating global command state in tests**: `rootCmd` is a package-level variable. Tests that run in parallel will race. Use factory functions for fully testable CLIs.
- **Over-nesting subcommands**: Two levels (`app command subcommand`) is usually the right depth. Three or more levels confuse users.
