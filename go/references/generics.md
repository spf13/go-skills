## Generics (Go 1.18+)

Generics exist to eliminate duplicated algorithms, not to create type hierarchies. If you are thinking about generics in terms of inheritance or polymorphism, stop — you are writing Java.

### When to Use Generics

Use generics when you have the **same algorithm** that needs to operate on **multiple concrete types**:

```go
// Good: generic algorithm, concrete types as inputs
func Map[S, T any](slice []S, f func(S) T) []T {
    result := make([]T, len(slice))
    for i, v := range slice {
        result[i] = f(v)
    }
    return result
}

// Good: constraint expresses a meaningful requirement
func Min[T cmp.Ordered](a, b T) T {
    if a < b {
        return a
    }
    return b
}
```

### When NOT to Use Generics

```go
// Bad: generic interface for polymorphism — this is Java
type Repository[T any] interface {
    Find(id string) (T, error)
    Save(entity T) error
}

// Good: a concrete interface for what you actually need
type UserStore interface {
    FindUser(id string) (*User, error)
    SaveUser(u *User) error
}
```

- **Do not** create generic base types, generic services, or generic repositories.
- **Do not** use `any` as a constraint to mean "I don't know the type yet." That's a design smell.
- **Do** use `comparable` when you need map keys or equality checks.
- **Do** use `cmp.Ordered` when you need `<`, `>`, `<=`, `>=`.
- Start with a concrete implementation. Generify only when you have the same logic repeated across 3+ types.
- Generic type aliases are fully supported since Go 1.24.
