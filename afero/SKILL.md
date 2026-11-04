---
name: afero
description: >
  Expert skill for testable filesystem code in Go with spf13/afero, authored by its creator. Covers
  injecting afero.Fs, in-memory tests with MemMapFs, layering with ReadOnlyFs, CopyOnWriteFs and
  BasePathFs, bridging to io/fs, and where MemMapFs differs from a real disk. Use whenever Go code
  reads, writes, or walks files and must be unit-tested without touching disk, whenever afero, an
  injected filesystem, or a --dry-run that must not write is mentioned, and when choosing between
  afero, io/fs with testing/fstest, and os.Root.
---

# Testable Filesystem Code: afero

Business logic that calls `os` directly can only be tested against a real disk. Inject the filesystem instead, so tests run in memory and production runs on the OS. `github.com/spf13/afero` is the interface for code that writes; `io/fs` is the interface for code that only reads.

Every behavior below was verified against afero v1.15.0 on Go 1.27.

## When to Activate

- Go code reads, writes, creates, renames, or walks files and needs unit tests.
- A test writes temp files only to exercise file-handling logic.
- The code takes an `afero.Fs`, or someone asks how to fake the filesystem.
- A command needs a `--dry-run` that shows what would change without writing.
- Code builds paths under a directory from names it does not control.

## Choose the Interface by What the Code Does

| The code… | Accept | Test with | Production |
| --- | --- | --- | --- |
| only reads | `fs.FS` | `fstest.MapFS` | `os.DirFS(dir)` or `root.FS()` |
| reads and writes | `afero.Fs` | `afero.NewMemMapFs()` | `afero.NewOsFs()` |
| must stay inside a directory, whatever the names | `*os.Root` | `os.OpenRoot(t.TempDir())` | `os.OpenRoot(dir)` |

Read-only code needs no dependency: `io/fs` and `testing/fstest` are the standard library. Reach for afero when the code writes.

## The Pattern: Inject the Fs, Choose It in `main`

```go
type Store struct {
    fs  afero.Fs
    dir string
}

func NewStore(fs afero.Fs, dir string) *Store {
    return &Store{fs: fs, dir: dir}
}

func (s *Store) Save(name string, data []byte) error {
    if err := s.fs.MkdirAll(s.dir, 0o755); err != nil {
        return err
    }
    return afero.WriteFile(s.fs, filepath.Join(s.dir, name), data, 0o644)
}
```

```go
func TestSave(t *testing.T) {
    fs := afero.NewMemMapFs()
    s := NewStore(fs, "/data")
    if err := s.Save("a.txt", []byte("hi")); err != nil {
        t.Fatal(err)
    }
    got, err := afero.ReadFile(fs, "/data/a.txt")
    if err != nil || string(got) != "hi" {
        t.Fatalf("ReadFile = %q, %v", got, err)
    }
}
```

`main` passes `afero.NewOsFs()`. Nothing below `main` imports `os` for file access; one stray `os.WriteFile` makes the in-memory test lie.

## MemMapFs Is Not a Disk

A test that passes on `MemMapFs` proves the logic, not the filesystem contract. Four differences decide whether a green test means anything:

1. **It creates missing parent directories.** `afero.WriteFile(mem, "/no/such/dir/f.txt", …)` succeeds and the directories appear; on `OsFs` the same call fails. A forgotten `MkdirAll` passes every in-memory test. To prove the code creates its directories, run that test on disk: `afero.NewBasePathFs(afero.NewOsFs(), t.TempDir())`.
2. **It does not enforce permission bits.** A file written `0o444` opens for writing and accepts the write. Test permission handling on disk, or with `ReadOnlyFs`.
3. **Rooted and relative paths are different files.** `data/x.txt` and `/data/x.txt` are separate entries. Paths are cleaned (`/data//x.txt`, `/data/./x.txt` and `/data/../data/x.txt` all find `/data/x.txt`), but a leading `/` is never added or removed. Build every path from one absolute root with `filepath.Join`.
4. **It is a weak `fs.FS`.** `afero.NewIOFS(mem)` opens only relative-keyed files (`fs.FS` names never start with `/`), yet `ReadDir(".")` still lists the top directory of a rooted key, which then fails to open. Its files also fail `fstest.TestFS`: a `ReadAt` that runs past end of file returns the bytes it read with a `nil` error where `io.EOF` is required. Test `fs.FS` consumers with `fstest.MapFS`, not a bridged `MemMapFs`.

It is safe for concurrent use: parallel goroutines writing different files pass the race detector, so `t.Parallel` subtests may share one `MemMapFs`.

## Layering Filesystems

| Constructor | Behavior |
| --- | --- |
| `afero.NewReadOnlyFs(fs)` | Every write fails with `EPERM`; `errors.Is(err, os.ErrPermission)` is true. |
| `afero.NewCopyOnWriteFs(base, layer)` | Reads fall through to `base`; writes land in `layer`; `base` is never modified. Files that exist only in `base` cannot be removed or renamed. |
| `afero.NewBasePathFs(fs, dir)` | Maps every path under `dir`; `/in.txt` is `dir/in.txt`, and `../` above `dir` does not resolve. |
| `afero.FromIOFS{FS: fsys}` | Adapts a read-only `fs.FS` (`embed.FS`, `fstest.MapFS`) to `afero.Fs`; writes fail. |
| `afero.NewIOFS(fs)` | Adapts an `afero.Fs` to `fs.FS` (see the `MemMapFs` caveat above). |

`CopyOnWriteFs` is a dry run for free when the command only creates or overwrites files. It runs its real code against an overlay, then reports what changed, and the disk is untouched:

```go
func run(fs afero.Fs, dryRun bool) (afero.Fs, error) {
    if dryRun {
        fs = afero.NewCopyOnWriteFs(fs, afero.NewMemMapFs())
    }
    return fs, apply(fs) // apply writes through fs; on a dry run only the overlay changes
}
```

It cannot delete or move a file that exists only in the base: `Remove` fails with "file does not exist", `Rename` fails with `EPERM`, and `RemoveAll` returns `nil` while the file stays visible — a dry run that reports success for a delete it never staged. A command that deletes or moves files needs its own plan-then-apply step for `--dry-run`.

## `BasePathFs` Is Not a Security Boundary

`BasePathFs` confines *lexical* paths only. It follows symlinks: a link inside the base that points outside it reads the outside file. For names from users, archives, or the network, open the directory with `os.Root` (Go 1.24), which rejects both `../` and symlinks that escape with `path escapes from parent`:

```go
root, err := os.OpenRoot(dir)
if err != nil {
    return err
}
defer root.Close()
data, err := root.ReadFile(name) // Root.ReadFile and Root.WriteFile: Go 1.25
```

Below Go 1.25, use `root.Open` and `root.Create`. To make an untrusted name a safe file name before it reaches the filesystem, see the `fileflow-pathologize` skill.

## Helpers

The package-level helpers take an `Fs` first: `afero.ReadFile`, `afero.WriteFile`, `afero.ReadDir`, `afero.Exists`, `afero.DirExists`, `afero.Walk`, `afero.TempDir`, `afero.Glob`. `afero.Afero{Fs: fs}` wraps the same helpers as methods (`a.Exists(path)`), which reads well when a type uses many of them.

## Common Mistakes

- **Mixing `os` and the injected `Fs`.** One `os.Stat` or `os.ReadFile` in the logic means the test checks a disk the code never wrote to.
- **`afero.NewOsFs()` in a unit test.** It touches the disk, so the test is slow, order-dependent, and leaves files. Use `NewMemMapFs`, or `NewBasePathFs(NewOsFs(), t.TempDir())` when disk behavior is the point.
- **Trusting `MemMapFs` to catch a missing `MkdirAll` or a permission bug.** It cannot; see the four differences.
- **Mixing `data/x` and `/data/x` keys.** The test writes one file and the code reads another.
- **Treating `BasePathFs` as a sandbox.** Symlinks escape it. Use `os.Root`.
- **Accepting `afero.Fs` in read-only code.** Accept `fs.FS`; the caller can pass `os.DirFS`, `embed.FS`, `fstest.MapFS`, or `afero.NewIOFS`.

## Installation

```bash
go get github.com/spf13/afero@latest
```
