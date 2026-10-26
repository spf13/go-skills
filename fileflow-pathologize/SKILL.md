---
name: fileflow-pathologize
description: >
  Use when Go code moves, copies, renames, downloads, extracts, or generates files on disk — an
  EXDEV "invalid cross-device link" rename, a "file already exists" conflict, a destination path
  built from untrusted input (user data, scraped names, regex captures) — or when a filename or
  path segment must be safe on every OS. Not for plain `os` file I/O (use go), CLI file commands
  (use cobra-viper), or desktop file dialogs (use wails).
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

## When NOT to Use This Skill

For general file I/O patterns use `go`. For CLI file commands use `cobra-viper`. For desktop file dialogs use `wails`. This skill is for cross-filesystem moves and untrusted filename sanitization.

## Rule 1: Trusted root, sanitized parts

`pathologize.Join` takes the root **verbatim** (separators, drive and UNC
prefixes preserved, never sanitized) and runs every untrusted part through
`Clean`: split on `/` and `\`, drop empty/`.`/`..` segments, cap each at 255
bytes, defuse reserved device names (`CON` → `CON_`), strip control characters
and the characters illegal on any OS (`` \ / : * ? " < > | ``), and keep the
result lexically **under** the root. A part cannot inject an absolute path or
`../` its way out. The root must come from configuration, never from the
request.

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


Prefer `Move` over `Rename`: `Rename` refuses cross-device moves. On Unix,
`Rename` links without replace; on FAT or network mounts without hard links it
falls back to `os.Rename` and no-overwrite becomes best-effort. `Copy` writes a
synced temp file in the destination directory and renames it in, so a failure
leaves no partial destination.

### `os.Root` vs `pathologize.Join`

`os.Root` (Go 1.24+) is the stdlib way to **confine opens inside a trusted
directory**: `root.OpenRoot(dir)` makes every later operation resolve from a
directory handle, so neither `..` nor a symlink can escape. It does **not**
sanitize names — it will happily create `CON` or `a:b` inside its root.
`pathologize.Join` is lexical and cross-OS: it yields a name valid on every OS
before anything is opened, and it keeps the result under the root as a string.
They answer different questions and compose well: when you need both portability
and symlink-proof confinement, `Join` the name then open it under an `os.Root`.

### Testing: these operations are not `afero`-injectable

`fileflow.Move`, `Copy`, and `Rename` call the OS directly — `MoveFileEx` on
Windows, link+remove/`os.Rename` on Unix, plus `fsync`/`chmod`/device checks.
No `afero.Fs` seam exists or should: a memory FS cannot model those primitives,
so it would test the fake instead of the no-replace and cross-device behavior
that is the point of the package. `Flow.FindAvailableName` (a plain
`func(string)(string,error)`) is unit-testable without disk; moves themselves
are integration-tested on a real `t.TempDir()`.

### Symlinks: followed or preserved

| Case | `Rename` / same-FS `Move` | cross-FS `Move` / `Copy` | `Replace` |
|---|---|---|---|
| `src` is a symlink | the **link itself** moves | **dereferenced** into a regular file, then the link is removed | rejected (`fs.ErrInvalid`) |
| `dst` names a symlink | `Exists`/`Equal` follow it → suffix or `ErrSameFile` | same as left | **link is replaced, not followed** |

`pathologize` never sees links — `Clean`/`CleanPath`/`Join` match bytes, not
inodes. A symlink *inside* the trusted root can still point outside it, so
confinement there belongs to `os.Root`.

### Verifying a destructive move (there is no dry run)

`Move`/`Rename`/`Replace` delete `src`; there is no dry-run mode. Rules that
hold up in review:

1. **Never retry blind** — re-inspect first; `fileflow.Equal(src, dst)` tells you
   whether a second try will suffix or re-attempt.
2. **`fileflow.Equal` checks identity** — `Exists` reports a symlink target,
   so `Equal` before delete is the only trustworthy verify.
3. **`Replace` is verified by path identity** — it always returns `dst`, never a
   suffix, so the caller owns verifying that `dst` now holds the expected content.

### Failure modes: ENOSPC, quota, and partial failure

- `Copy` and cross-FS `Move` stage content in the destination directory before
  committing; under ENOSPC the staging temp is removed and `src` is untouched.
- Same-FS `Rename`/`Move` allocate no file content, so they survive ENOSPC at
  the data level but still need a directory-entry slot.
- Under quota exceeded, the write fails like ENOSPC — the source is untouched,
  the partial temp is cleaned up.
- A `Copy`/`Move` interrupted between staging and commit leaves an orphaned temp
  file; `Replace`'s atomic-rename contract still holds on crash, but the caller
  must not treat a returned error as "nothing happened" — the source may have
  been synced and the temp present.

### Long paths on Windows

Windows caps each path segment at 255 bytes and the total legacy path at 260
chars; beyond that requires the `\\?\` prefix. `pathologize.Clean` enforces
the 255-byte segment cap. A `Join` component over `>260` chars needs
`\\?\`-prefix handling by the caller — `fileflow` itself does not special-case
long paths.

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
| `os.MkdirAll` before a move/copy | These create parents by default |

`filepath.Ext` remains the extension parser: `strings.CutLast` (Go 1.27) slices
around a separator, it does not parse extensions.
