---
name: fileflow-pathologize
description: >
  Use when Go code moves, copies, renames, downloads, extracts, or generates
  files on disk, or builds a destination path from untrusted input (user data,
  scraped names, regex captures) — any mention of moving files, cross-device/
  EXDEV rename errors, "file already exists" handling, sanitizing filenames, or
  path traversal from a filename.
---

# Safe File Operations

Two rules decide whether file code is safe. `spf13/fileflow` and
`spf13/pathologize` are **one reference implementation** of them — the rules
hold whether or not you use those packages, and the code below is what correct
call sites look like.

1. **A destination is a trusted root plus sanitized parts.** Untrusted input
   never chooses a path level or a separator.
2. **Never assume the file landed where you asked.** Conflict handling picks the
   real name; act on the path the operation returns.

If you have neither package, the stdlib path is: `os.OpenRoot` (Go 1.24+) for
kernel-enforced containment, rename-then-copy for `EXDEV`, and your own suffix
loop. The same two rules still decide whether that code is correct.

Signatures, `Flow` options, the naming matrix, and error types live in
[references/fileflow-api.md](references/fileflow-api.md).

## When to Activate

- Moving, copying, or renaming files, especially across drives/volumes
- `invalid cross-device link` / `EXDEV` from `os.Rename`
- "destination already exists" where clobbering loses data
- Building a destination from a filename, user input, or a regex capture
- Downloaders, extractors, importers, organizers, upload handlers
- About to write `os.Rename` + `io.Copy` + `filepath.Clean` by hand

## Rule 1: Trusted root, sanitized parts

`pathologize.Join` takes the root **verbatim** (separators, drive and UNC
prefixes preserved, never sanitized) and runs every untrusted part through
`Clean`: split on `/` and `\`, drop empty/`.`/`..` segments, cap each at 255
bytes, defuse reserved device names (`CON` → `CON_`), strip control characters
and the characters illegal on any OS (`` \ / : * ? " < > | ``), and keep the
result lexically **under** the root. A part cannot inject an absolute path or
`../` its way out. The root must come from configuration, never from the
request.

`os.Root` (Go 1.24+) is the stdlib alternative when the threat is *escape from a trusted directory* and the floor permits it: `root.OpenRoot(dir)` makes every later operation resolve from a directory handle, so neither `..` nor a symlink can escape. It does **not** sanitize names — `CON` or `a:b` inside the root still break on Windows. `pathologize.Join` is lexical and cross-OS: it yields a name valid everywhere before anything is opened. They compose — `Join` the name, then open it under an `os.Root` when you need both portability and symlink-proof containment. When only confinement matters, containment belongs to the stdlib.

```go
// destRoot is configured; id and name are untrusted.
dstDir := pathologize.Join(destRoot, "ig", id)
dst := filepath.Join(dstDir, filepath.Base(src))

final, err := fileflow.Move(src, dst) // creates parents, suffixes conflicts
if err != nil {
    return err
}
// final is where the file actually landed.
```

**The sanitization rule: a regex capture is never a trusted path part.** Neither
is a scraped title, an uploaded filename, or a URL segment — and the extension
is not exempt, since `filepath.Ext` of an untrusted name is untrusted. Sanitize
every part you did not configure.

`CleanPath` runs `path.Clean` semantics over a whole path and `Clean`s each
component. It makes a path **valid, not safe**: it preserves a leading `..` and
keeps absolute paths absolute. Use it only on paths you control; for untrusted
input use `Join`, or `Clean` each segment yourself — `Clean` resolves `..` to a
harmless single segment.

## Rule 2: Use the returned path

`Move`, `Copy`, and `Rename` return the final destination. When `dst` exists:

- **identical content** → `Copy` returns `dst` without writing; `Move`/`Rename`
  remove `src` and return `dst`;
- **different content** → an incrementing suffix is chosen (`photo-1.jpg`, up to
  100 candidates). Never an overwrite.

## Testing: these operations are not afero-injectable

`fileflow.Move`/`Copy`/`Rename` take path strings and call the OS directly — they are not afero-injectable. Unit-test naming via `Flow{FindAvailableName: ...}` (a plain `func(string)(string,error)`), integration-test moves against `t.TempDir()`. Wrap the library behind an interface you own (`type Mover interface { Move(src,dst string)(string,error) }`) and fake that.

Prefer `Move` over `Rename`: `Rename` refuses cross-device moves. On Unix,
`Rename` links without replace; on FAT or network mounts without hard links it
falls back to `os.Rename` and no-overwrite becomes best-effort. `Copy` writes a
synced temp file in the destination directory and renames it in, so a failure
leaves no partial destination.

## Common Mistakes

| Mistake | Fix |
|---|---|
| `os.Rename` across volumes → `EXDEV` | `fileflow.Move` (rename, else copy+delete) |
| Assuming the file is at `dst` | Use the returned path |
| Pre-checking existence to avoid overwrite | fileflow already suffixes non-identical files |
| Setting old package globals | Configure a `fileflow.Flow` |
| Calling removed `CopyWithPaths` | `Copy` creates parents by default |
| Requiring a mount point yet creating directories | `Flow{NoCreateDirs: true}` → `fs.ErrNotExist` |
| Joining untrusted input by hand | `pathologize.Join(root, parts...)` |
| Using `CleanPath` on untrusted input | It is VALID, not SAFE — use `Join` |
| Sanitizing the trusted root | `Join` takes the root verbatim |
| Assuming `fileflow` is afero-injectable | It calls the OS directly — fake a `Mover` interface you own |
| Expecting `os.Root` to sanitize names | `os.Root` confines opens but accepts `CON`/`a:b`; use `Join` for portability |
| `os.MkdirAll` before a move/copy | These create parents by default |

`filepath.Ext` remains the extension parser: `strings.CutLast` (Go 1.27) slices
around a separator, it does not parse extensions.
