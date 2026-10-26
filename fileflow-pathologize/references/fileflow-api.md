# fileflow & pathologize — API reference

Companion to `SKILL.md`. This is the surface of one reference implementation of
the two safe-file rules; nothing here is required reading to apply them.

```sh
go get github.com/spf13/fileflow
go get github.com/spf13/pathologize
```

## fileflow — package functions

```go
func Move(src, dst string) (string, error)   // rename if same FS, else copy+delete
func Rename(src, dst string) (string, error) // same-FS only; fails on cross-device
func Copy(src, dst string) (string, error)   // atomic copy; returns final path
func Exists(path string) bool
func Replace(src, dst string) (string, error) // always returns dst; replaces symlink dst, rejects symlink src
func Equal(file1, file2 string) (bool, error) // byte-for-byte comparison
```

- All three operations return the **final destination path** (see conflict
  handling) and create missing destination directories.
- `Exists` reports only an accessible regular file; it returns false for
  directories.
- `CopyWithPaths` no longer exists — `Copy` creates parents itself.

## Configuring behavior with `Flow`

```go
type Flow struct {
    FindAvailableName func(string) (string, error)
    DirMode           fs.FileMode
    NoCreateDirs      bool
}
```

The zero value is ready to use:

- nil `FindAvailableName` → `FindAvailableNameInc`
- zero `DirMode` → `fileflow.DefaultDirMode` (`0o755`)
- `NoCreateDirs: true` → destinations must already exist; the operation fails
  with an error matching `fs.ErrNotExist` via `errors.Is` instead of creating
  the path. Use it when the destination is a required mount point and creating
  it on the wrong volume would be dangerous.

```go
f := fileflow.Flow{
    FindAvailableName: fileflow.FindAvailableNameAuto,
    DirMode:           0o700,
    NoCreateDirs:      true,
}

final, err := f.Move(src, dst)
```

`Flow` has `Move`, `Rename`, and `Copy` methods with the package-level
signatures. Configure an instance before first use; do not mutate one while its
methods may be running. The older mutable package globals are gone.

## Conflict handling

When `dst` already exists, fileflow never blindly overwrites:

| Existing file | `Copy` | `Move` / `Rename` |
|---|---|---|
| identical to source | leaves `src` untouched, returns `dst`, writes nothing | removes `src`, returns `dst` |
| differs from source | chooses an available name and writes there | chooses an available name and moves there |

Suffixing adds an incrementing counter: `photo.jpg` → `photo-1.jpg` →
`photo-2.jpg`, across at most 100 candidates; the limit is internal and not
configurable. Exhausting it returns `ErrMaxAttemptsReached`. Always use the
returned path — never assume the file landed at `dst`.

## Naming strategies (the matrix)

`FindAvailableName` receives the requested base name and returns an unused name.
It may also be any `func(baseName string) (string, error)`.

| Strategy | Rule | `file-3.txt` (taken) | `report-2024.txt` (taken) |
|---|---|---|---|
| `FindAvailableNameInc` (default) | keep the full requested name, append `-1`, `-2`, … | `file-3-1.txt` | `report-2024-1.txt` |
| `FindAvailableNameCont` | continue a trailing counter of one or more digits | `file-4.txt` | `report-2025.txt` |
| `FindAvailableNameAuto` | continue one- or two-digit counters; three or more digits are meaningful data | `file-4.txt` | `report-2024-1.txt` |
| `FindAvailableNameTS` | append a nanosecond timestamp `20060102-150405.000000000`; fall back to incrementing on collision | timestamped name | timestamped name |

Choose deliberately: `Inc` is safest for opaque names; `Cont` suits sequences
you own; `Auto` protects embedded years and zero-padded IDs; `TS` suits
one-shot captures where ordering by name matters.

## Write safety and durability

- `Rename` uses no-replace primitives: Windows `MoveFileEx` without replace, and
  on Unix atomic link-and-remove. A racing writer cannot normally clobber an
  occupied destination.
- On filesystems without hard-link support (FAT, some network mounts) rename
  falls back to `os.Rename`; the no-overwrite guarantee is then best-effort.
- `Copy` writes to a temporary file in the destination directory, syncs it,
  preserves the source permission mode, and renames it into place with the same
  no-replace logic. A failed copy exposes no partial destination and leaves no
  temporary file behind.

## Errors

```go
fileflow.ErrSameFile               // src and dst are the same file
fileflow.ErrMaxAttemptsReached     // ran out of conflict-suffix attempts
*fileflow.ErrFailedMovingFile      // fields: Src, Dst, Err
*fileflow.ErrFailedRemovingOriginal // fields: File, Err
```

Sentinel values are `errors.Is` targets; structured errors are `errors.As`
targets.

- `ErrSameFile` detects equivalent paths, symlinks, and case-only differences on
  case-insensitive filesystems — not just identical strings. Treat it as a no-op
  success only when that matches the caller's semantics.
- A removal failure returns the final destination together with
  `ErrFailedRemovingOriginal`: the content arrived safely but the source could
  not be deleted.

```go
var moveErr *fileflow.ErrFailedMovingFile
if errors.As(err, &moveErr) {
    log.Printf("move %s -> %s failed: %v", moveErr.Src, moveErr.Dst, moveErr.Err)
}
```

On Go 1.26 or later, `errors.AsType[T]` is type-safe and faster and removes the
declared-target dance; below that floor keep the `errors.As` form. `AsType` is
generic:

```go
if moveErr, ok := errors.AsType[*fileflow.ErrFailedMovingFile](err); ok {
    log.Printf("move %s -> %s failed: %v", moveErr.Src, moveErr.Dst, moveErr.Err)
}
```

## pathologize — functions

```go
func Clean(filename string) string             // sanitize ONE name/segment
func CleanPath(path string) string             // make a full path VALID (volume-aware)
func Join(root string, parts ...string) string // trusted root + untrusted parts, contained
func IsClean(filename string) bool             // true if Clean(name) == name
```

```go
pathologize.Clean("CON..")        // "CON_"
pathologize.Clean("a/b:c*d?.txt") // "abcd.txt"  (/, :, *, ? are illegal)
```

`Clean` strips control characters plus `` \ / : * ? " < > | ``, maps invalid
UTF-8 to U+FFFD, trims leading/trailing whitespace and trailing dots, caps the
name at 255 bytes, and defuses reserved device names (`CON` → `CON_`). It is
idempotent and never returns an empty string (falls back to `"file"`).

### `Join`

- `root` passes through verbatim: separators, drive prefix, and trailing slash
  preserved, never sanitized. It must be trusted (configured, not user input).
- Each part is split on both `/` and `\`; structural segments (empty, `.`, `..`)
  are dropped; every remaining segment runs through `Clean`. A part that
  sanitizes to nothing contributes `"file"`.
- The result stays lexically under `root`: no absolute path injection, no
  escape. This containment is what `CleanPath` does not provide.

```go
pathologize.Join("/data", "../../etc", "pass:wd") // "/data/etc/passwd"
pathologize.Join(`C:\Sorted`, "ig", userName)     // root verbatim, userName cleaned
```

### `CleanPath`

Runs `path.Clean` semantics (collapse `//`, resolve `.` and internal `..`, drop a
trailing separator) and additionally `Clean`s each component. Volume-aware and
separator-agnostic:

- both `/` and `\` are accepted as separators; output always uses `/`
- a Windows drive prefix (`C:`) passes through verbatim
- a UNC prefix (`\\server\share`) is preserved as `//server/share`

```go
pathologize.CleanPath(`C:\Users\me\report:final`) // "C:/Users/me/reportfinal"
pathologize.CleanPath("a//b/../c")                // "a/c"
```

It makes a path **valid, not safe**: like `path.Clean`, it preserves a leading
`..` in a relative path and keeps an absolute path absolute. Use it only on
paths you control; for untrusted input use `Join` or `Clean` per segment.
