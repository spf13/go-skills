## Syntax: Use Current Go

LLMs frequently generate outdated syntax. Know the current idioms:

### Range over Integer (Go 1.22)

```go
// Old
for i := 0; i < 10; i++ { ... }

// New
for i := range 10 { ... }
```

### `new(expr)` (Go 1.26)

`new` accepts an expression and returns a pointer to a new variable holding that value, replacing the `v := expr; &v` dance:

```go
// Old
v := yearsSince(born)
p := &v

// New (Go 1.26)
p := new(yearsSince(born))
```

The canonical case is an optional pointer struct field — see Configuration and Struct Design above.

### Generic Methods (Go 1.27)

A method declaration may declare its own type parameters, so a generic helper can be scoped to a type instead of the whole package. `math/rand/v2` gains the method form `(*Rand) N[Int intType](Int) Int` alongside the existing package-level generic `N`:

```go
import "math/rand/v2"

r := rand.New(rand.NewPCG(seed1, seed2))
n := r.N(100) // method form, Go 1.27
```

Two hard limits:

- A method of an interface may not declare type parameters.
- An interface method cannot be implemented by a generic method.

Below Go 1.27, hoist the helper to a package-level generic function.

### Self-Referential Type-Parameter Constraints (Go 1.26)

A generic type may now refer to itself in its own type-parameter list:

```go
type Adder[A Adder[A]] interface {
    Add(A) A
}
```

Before Go 1.26 this was rejected by the compiler.

### Struct Literal Keys: Any Field Selector (Go 1.27)

A key in a struct literal may be any valid field selector for the struct type, not just a top-level field name — so a promoted field can be named directly:

```go
type Base struct{ ID int }

type User struct {
    Base
    Name string
}

u := User{ID: 1, Name: "steve"} // Go 1.27 — Base.ID named directly
```

Below Go 1.27, initialize the embedded value explicitly: `User{Base: Base{ID: 1}, Name: "steve"}`.

### Generalized Function Type Inference (Go 1.27)

Function type inference applies in every context where a generic function is assigned to, or converted to, a value of a matching function type:

```go
type Less func(a, b string) int

// Go 1.26 and earlier: "cannot use generic function cmp.Compare without instantiation"
var less Less = Less(cmp.Compare) // Go 1.27 infers T = string
```

Below Go 1.27, instantiate explicitly: `cmp.Compare[string]`.

### Build Constraints

```go
// Old (deprecated — do not generate)
// +build linux darwin

// Current
//go:build linux || darwin
```

The `//go:build` form is required since Go 1.17. Never emit the old `// +build` syntax.

### `any` Instead of `interface{}`

```go
// Old
func Print(v interface{}) { ... }

// Current — `any` is a built-in alias for interface{} since Go 1.18
func Print(v any) { ... }
```

### Tool Dependencies in `go.mod` (Go 1.24)

For a module whose floor is Go 1.24 or later, never generate a `tools.go` file with blank imports — track dev tools with the `tool` directive:

```bash
go get -tool golang.org/x/tools/cmd/stringer
go tool stringer -type=Color   # run it
```

Below Go 1.24, the `tool` directive does not exist and a `tools.go` file with blank imports remains the required form; keep it behind a `//go:build tools` constraint so it never enters the real build.
