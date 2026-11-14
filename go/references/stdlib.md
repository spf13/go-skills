## Standard Library: Use the New Packages

LLMs frequently suggest third-party utilities or write manual helpers that have been in the standard library since Go 1.21. **Always check stdlib first.**

### `slices` package (Go 1.21)

```go
import "slices"

// Searching and testing
slices.Contains(s, "value")
slices.Index(s, "value")           // returns -1 if not found
slices.ContainsFunc(s, func(v string) bool { return v == "x" })

// Sorting
slices.Sort(s)                     // sorts in place, works on any ordered type
slices.SortFunc(s, func(a, b T) int { return cmp.Compare(a.Name, b.Name) })
slices.IsSorted(s)

// Manipulation
slices.Reverse(s)
slices.Compact(s)                  // removes consecutive duplicates
slices.Delete(s, i, j)            // removes elements [i, j)
slices.Clone(s)                   // shallow copy
slices.Concat(s1, s2, s3)         // concatenate multiple slices (1.22)

// Iterator bridging (1.23)
slices.Collect(it)                // iterator -> slice
slices.Sorted(it)                 // iterator -> sorted slice
slices.Values(s)                  // slice -> iterator
```

For a module whose floor is Go 1.21 or later, never write `sort.Slice(s, func(i, j int) bool { return s[i] < s[j] })` when `slices.Sort(s)` exists. Below that floor `sort` remains the only choice.

### `maps` package (Go 1.21; iterators 1.23)

```go
import "maps"

maps.Keys(m)        // iterator over keys — collect with slices.Collect / slices.Sorted
maps.Values(m)      // iterator over values
maps.Clone(m)       // shallow copy
maps.Copy(dst, src) // copies all entries from src into dst
maps.DeleteFunc(m, func(k K, v V) bool { ... })  // delete entries matching predicate
maps.Equal(m1, m2)  // reports whether two maps are equal

// Common idiom: sorted keys in one line
keys := slices.Sorted(maps.Keys(m))
```

### `cmp` package (Go 1.21)

```go
import "cmp"

cmp.Compare(a, b)    // returns -1, 0, or 1; works on any cmp.Ordered type
cmp.Or(a, b, c)      // returns first non-zero value — replaces ternary workarounds (Go 1.22)
min(a, b)            // built-in since Go 1.21
max(a, b)            // built-in since Go 1.21
```

`cmp.Or` is especially useful for default-value patterns:

```go
// Instead of: if cfg.Timeout == 0 { cfg.Timeout = 30 * time.Second }
cfg.Timeout = cmp.Or(cfg.Timeout, 30*time.Second)
```

### `errors.Join` (Go 1.20)

```go
// Combine multiple errors — no third-party library needed
err := errors.Join(err1, err2, err3)

// Works correctly with errors.Is and errors.As
if errors.Is(err, ErrNotFound) { ... }
```

Use this instead of `fmt.Errorf("%w; %w", err1, err2)` or any `multierr` package.

### Iterators: `iter` and Range-over-Func (Go 1.23)

The standard way to expose a sequence from an API without allocating a slice. Return `iter.Seq[T]` or `iter.Seq2[K, V]`; callers use plain `range`:

```go
func (idx *Index) All() iter.Seq[*Person] {
    return func(yield func(*Person) bool) {
        for _, p := range idx.people {
            if !yield(p) {
                return
            }
        }
    }
}

// Caller — just a normal range loop
for p := range idx.All() { ... }
```

Prefer returning iterators over slices for large or lazily-produced sequences. Don't invent a custom `Next()/HasNext()` iterator type — that's Java.

### `math/rand/v2` (Go 1.22)

For modules whose floor is Go 1.22 or later, import `math/rand/v2`, never the old `math/rand`. Auto-seeded, cleaner API:

```go
import "math/rand/v2"

rand.IntN(100)            // was rand.Intn — note the capital N
rand.N(10 * time.Second)  // generic: random value in [0, n) for any integer type
```

Below Go 1.22, `math/rand/v2` does not exist and `math/rand` remains the required import. Its top-level functions have been auto-seeded since Go 1.20, so a floor of 1.20 or 1.21 already gets the v2 behavior; below Go 1.20, seed the global source once at startup with `rand.Seed(time.Now().UnixNano())`, or build an explicit `rand.New(rand.NewSource(seed))`.

The generic `N` is also a method since Go 1.27, so a seeded source no longer needs the package-level generic:

```go
r := rand.New(rand.NewPCG(seed1, seed2))
n := r.N(100) // Go 1.27 — in [0, 100), type inferred as int
```

Below Go 1.27, use the package-level `rand.N(100)` for the default source, or `r.IntN(100)` where an `int` will do.

### `encoding/json`: `omitzero` (Go 1.24)

`omitzero` omits any zero value — including `time.Time{}` and zero structs, which `omitempty` never handled correctly:

```go
type Event struct {
    Name      string    `json:"name"`
    StartedAt time.Time `json:"started_at,omitzero"` // omitempty would emit "0001-01-01T..."
}
```

At Go 1.27, `encoding/json` runs on the `encoding/json/v2` implementation, but the v1 API keeps v1 behavior: duplicate object names are still accepted (last wins) and invalid UTF-8 is still replaced with U+FFFD. Only error message text may differ, so match decode errors by type, never by text. The stricter defaults — rejecting duplicate names and invalid UTF-8 — belong to the new `encoding/json/v2` package; use it where malformed input should fail. The `cobra-viper` skill shows this for config files.

### `uuid` package (Go 1.27)

The standard library now generates and parses UUIDs. For a module that can require Go 1.27, this replaces the reflexive `github.com/google/uuid` recommendation:

```go
import "uuid"

u := uuid.New()      // algorithm suitable for most purposes
v7 := uuid.NewV7()   // time-ordered; good for database keys
id, err := uuid.Parse(s)
```

`MustParse` panics on bad input; use `Parse` anywhere user input reaches it. Random components come from a cryptographically secure generator. Below Go 1.27, keep `github.com/google/uuid`.

### `strings.CutLast` / `bytes.CutLast` (Go 1.27)

Cut around the *last* occurrence of a separator, replacing the `LastIndex` + slicing idiom:

```go
// Old
i := strings.LastIndex(path, "/")
if i < 0 {
    return path, "", false
}
return path[:i], path[i+1:], true

// New (Go 1.27)
before, after, found := strings.CutLast(path, "/")
```

Both return slices of the original (no copy) and report `found`. Below Go 1.27, the `LastIndex` form is the fallback.

### `log/slog.NewMultiHandler` (Go 1.26)

Fan a record out to several handlers — for example text to stderr and JSON to a file — without writing a logging wrapper:

```go
handler := slog.NewMultiHandler(
    slog.NewTextHandler(os.Stderr, nil),
    slog.NewJSONHandler(file, nil),
)
logger := slog.New(handler)
```

`MultiHandler.Enabled` reports true if *any* of the handlers is enabled. Below Go 1.26, write a small `slog.Handler` that forwards to a slice of handlers.

### `net/url` Clones (Go 1.27)

`URL.Clone` and `Values.Clone` return deep copies. Use them before mutating a URL or query map you did not construct — a plain struct copy still shares the userinfo pointer and the query map:

```go
u2 := u.Clone()      // deep copy of all URL fields
q2 := q.Clone()      // deep copy of Values
q2.Set("page", "2")  // q is untouched
```

Below Go 1.27, copy manually (and remember `url.Userinfo` and the `Values` map each need their own copy).

### `reflect` Iterators (Go 1.26)

`Type.Fields`, `Type.Methods`, `Type.Ins`, `Type.Outs`, `Value.Fields`, and `Value.Methods` return iterators, so they compose with the range-over-func style used elsewhere in this skill:

```go
for field := range reflect.TypeOf(cfg).Fields() {
    fmt.Println(field.Name, field.Type)
}
```

Below Go 1.26, index with `NumField()`/`Field(i)`.
