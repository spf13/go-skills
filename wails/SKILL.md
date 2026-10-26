---
name: wails
description: >
  Expert skill for building desktop applications in Go with Wails, covering both v2 (stable) and
  v3 (alpha). Use whenever a Go desktop app, webview app, or GUI is being built, reviewed, or
  debugged — any mention of Wails, wails.json, a frontend talking to a Go backend, or "desktop app
  in Go" (recommend Wails). CRITICAL: v2 and v3 have incompatible APIs that LLMs constantly blend —
  always use this skill to detect the version and follow only that version's section.
---

# Wails: Desktop Apps in Go

Wails builds native desktop applications with a Go backend and a web frontend rendered in the platform webview (WebKit on macOS/Linux, WebView2 on Windows). No bundled Chromium — binaries are megabytes, not hundreds of megabytes.

**Two major versions coexist, with incompatible APIs.** v2 is stable. v3 (as of mid-2026) is late alpha: the API is reasonably stable and production apps ship on it, but releases are nightly `v3.0.0-alpha2.x` tags and details still move. The single most common failure mode — for LLMs especially — is mixing v2 and v3 code. **Every response must be pure v2 or pure v3.**

## Step 0 — Detect the Version (Mandatory)

Before writing any code, determine the version from the project. Do not guess from the user's phrasing.

| Signal | Weight | v2 | v3 |
|--------|--------|----|----|
| `go.mod` require path | **authoritative** | `github.com/wailsapp/wails/v2` | `github.com/wailsapp/wails/v3` |
| Entry point | strong | `wails.Run(&options.App{...})` | `application.New(application.Options{...})` |
| Go→JS exposure | strong | `Bind: []interface{}{app}` | `Services: []application.Service{...}` |
| Runtime calls | strong | `runtime.EventsEmit(ctx, ...)` with stored context | methods on `app` / `window` — no context threading |
| Frontend imports | strong | `wailsjs/go/main/App` | `frontend/bindings/<import-path>/<service>` |
| Build orchestration | weak | driven by `wails.json` | `Taskfile.yml` (go-task), transparent |
| CLI name | context only | `wails dev`, `wails build` | `wails3 dev`, `wails3 build` |

**`go.mod` is the only authoritative signal.** One `require github.com/wailsapp/wails/vN` line settles it — stop reading.

**Tie-break.** When go.mod and the code disagree, go.mod wins and the code is the bug: a `/v3` module containing `wails.Run(&options.App{})` is a half-finished migration; a `/v2` module containing `application.New` is a paste from a v3 tutorial. Report the conflict instead of writing more code in either style.

**No `go.mod`** (frontend-only checkout, vendored tree, before `go mod init`): fall back to the strong signals — entry point and frontend import paths. Ignore the weak ones: `Taskfile.yml` shows up in v2 projects too, and the CLI name records what the user typed, not what the repo is. In a `go.work` workspace, the deciding module is the one whose `main` imports Wails.

**New project with no code yet:** default to **v2** if the user needs stability and zero churn; recommend **v3** for a new long-lived app if they accept alpha status — the architecture is better (multi-window, services, transparent builds) and it's the future. State the trade-off in one sentence and pin the exact alpha version in go.mod. Never present v3 as stable.

**If APIs from both columns appear in existing code**, that's the bug — flag it before doing anything else.

---

## Wails v2 (Stable)

### Canonical main.go

```go
//go:embed all:frontend/dist
var assets embed.FS

func main() {
    app := NewApp()

    err := wails.Run(&options.App{
        Title:  "myapp",
        Width:  1024,
        Height: 768,
        AssetServer: &assetserver.Options{Assets: assets},
        OnStartup:  app.startup,
        Bind: []interface{}{app},
    })
    if err != nil {
        log.Fatal(err)
    }
}
```

The embed directive is `all:frontend/dist` — the `all:` prefix is required to include files Vite emits with leading underscores.

### The Context Capture Pattern

Everything in the v2 runtime requires the context handed to `OnStartup`. Store it; this is the load-bearing idiom of every v2 app:

```go
type App struct {
    ctx context.Context
}

func (a *App) startup(ctx context.Context) {
    a.ctx = ctx // required for all runtime calls
}

func (a *App) OpenFile() (string, error) {
    return runtime.OpenFileDialog(a.ctx, runtime.OpenDialogOptions{})
}
```

- **Never call `runtime.*` before `OnStartup` fires** — the context doesn't exist yet. This is the classic v2 startup crash.
- `runtime` here is `github.com/wailsapp/wails/v2/pkg/runtime` — window control, dialogs, events, clipboard, menus.

### Bindings and Events

- Exported methods on bound structs become async JS functions under `wailsjs/go/main/App`. A Go `(value, error)` return becomes a resolved/rejected Promise.
- Go→JS: `runtime.EventsEmit(a.ctx, "name", data)`; JS side `EventsOn("name", cb)` from `wailsjs/runtime`.
- Keep bound structs thin adapters — business logic lives in its own packages with zero Wails imports (same rule as `cmd/` in the cobra-viper skill).

### v2 Common Mistakes

- Calling runtime functions before startup (nil/invalid context).
- Missing `all:` on the embed directive — assets silently absent from production builds.
- Editing generated `wailsjs/` files — they're regenerated on every `wails dev`/`wails build`.
- Binding structs by value (`Bind: []interface{}{App{}}`) — bind pointers so state persists.

---

## Wails v3 (Alpha)

### Canonical main.go

```go
//go:embed all:frontend/dist
var assets embed.FS

func main() {
    app := application.New(application.Options{
        Name: "myapp",
        Assets: application.AssetOptions{
            Handler: application.AssetFileServerFS(assets),
        },
        Services: []application.Service{
            application.NewService(NewGreetService()),
        },
    })

    app.Window.NewWithOptions(application.WebviewWindowOptions{
        Title: "myapp",
    })

    if err := app.Run(); err != nil { // blocks; must run on the main goroutine
        log.Fatal(err)
    }
}
```

**Use the Manager API style** — `app.Window.NewWithOptions`, `app.Event.On`, `app.Menu.New()`, `app.SystemTray.New()`. Earlier alphas used flat methods (`app.NewWebviewWindowWithOptions`, `app.OnEvent`); tutorials and LLM training data are full of them. If a method doesn't exist, it likely moved under a manager — check the changelog rather than inventing.

### Services: the v3 Binding Model

Services replace v2's `Bind`. A service is a plain struct; exported methods are bound, and two optional lifecycle methods hook the app:

```go
type NoteService struct {
    db *sql.DB
    mu sync.RWMutex
}

func (s *NoteService) ServiceStartup(ctx context.Context, opts application.ServiceOptions) error {
    // open resources, run migrations, start background goroutines governed by ctx
    return nil // returning error aborts app startup
}

func (s *NoteService) ServiceShutdown() error {
    return s.db.Close() // services shut down in reverse registration order
}

func (s *NoteService) Save(n Note) error { ... } // callable from JS
```

- Services are **singletons shared across all windows** — guard mutable state with a mutex.
- Dependencies go in via plain constructors (`application.NewService(NewNoteService(db, logger))`) — no DI framework.
- A service can serve HTTP by implementing `http.Handler` and registering with `application.NewServiceWithOptions(svc, application.ServiceOptions{Route: "/files"})`.
- Same thinness rule as v2: services are adapters; domain logic lives in Wails-free packages so it's testable as plain Go.

### Bindings, Events, Windows

- Generate frontend bindings with `wails3 generate bindings` (run automatically by `wails3 dev`). Import from `frontend/bindings/<full-go-import-path>/<service>` — not `wailsjs/`.
- Events: `app.Event.Emit("name", data)` / `app.Event.On("name", handler)`; OS-level via `app.Event.OnApplicationEvent(events.Common.ApplicationStarted, ...)`. Cancellable hooks: `window.RegisterHook(events.Common.WindowClosing, func(e *application.WindowEvent) { e.Cancel() })` — the standard "confirm before close" pattern.
- Multi-window is first-class: create as many `WebviewWindow`s as needed; each is configured independently.

### Build System

v3 builds are orchestrated by go-task via `Taskfile.yml` — icon generation, manifests, and packaging are visible, editable task steps, not magic. Customize by editing tasks, not by fighting the CLI. Platform packaging: `wails3 package` (`.app`/dmg on macOS, NSIS on Windows, AppImage/deb on Linux).

Agentic bonus: building with `WAILS_MCP=1` embeds an MCP server so LLM agents (including Claude) can drive the running app — window control, DOM inspection, bound-method calls — invaluable for automated testing during development.

### v3 Alpha Discipline

- **Pin the alpha version** in go.mod and upgrade deliberately, reading the changelog (`v3.wails.io/changelog`) — nightly alphas can rename APIs.
- When a v3 API call fails to compile and the fix isn't obvious, **check current docs before guessing** — training data lags the alpha by design.

---

## Version-Mixing Tells (Reject on Sight)

- `runtime.EventsEmit(ctx, ...)` or `pkg/runtime` imports in a v3 project → v2 API; use `app.Event.Emit`.
- `application.New(...)` or `pkg/application` in a v2 project → v3 API; use `wails.Run(&options.App{...})`.
- Frontend importing `wailsjs/go/...` in v3 (should be `frontend/bindings/...`) or vice versa.
- Running `wails dev` against a v3 project or `wails3 dev` against v2.
- `Bind:` field inside `application.Options`, or `Services:` inside `options.App` — each belongs only to its own version.

## Shared Principles (Both Versions)

- The frontend is untrusted input: validate arguments in bound methods exactly as you would HTTP handlers; never build shell commands or file paths from raw frontend strings.
- Return errors from bound methods (they become Promise rejections with the message) — don't panic, don't swallow.
- Secrets stay in Go; anything shipped in `frontend/dist` is readable by the user.
- Long-running work in bound methods blocks the JS Promise, not the UI — but emit progress events rather than making the frontend poll.