# ctxcache

A lightweight Go context-level cache library that provides simple caching for functions with the signature `func(K) V`.

## Design Philosophy

**Keep it simple**: ctxcache only supports single-parameter, single-return-value functions `func(K) V`. For scenarios requiring `(V, error)`, users can wrap the result in a custom struct.

This design keeps the API minimal and predictable while remaining flexible enough for most use cases.

## Installation

```bash
go get github.com/alingse/ctxcache
```

## Quick Start

### Basic Usage

```go
package main

import (
    "context"
    "fmt"
    "strconv"

    "github.com/alingse/ctxcache"
)

func getNumber(n int64) string {
    fmt.Printf("fetching number: %d\n", n)
    return strconv.FormatInt(n, 10)
}

func main() {
    ctx := context.Background()

    // Register cache for getNumber function
    ctx = ctxcache.WithCache[int64, string](ctx, ctxcache.FuncID("getNumber"), getNumber)

    // Get cached function
    cachedGetNumber, found := ctxcache.FromContext[int64, string](ctx, ctxcache.FuncID("getNumber"))
    if !found {
        panic("cache not found")
    }

    // First call - executes the function
    cachedGetNumber(42)  // Output: fetching number: 42

    // Second call - uses cache
    cachedGetNumber(42)  // No output (cached!)

    // Different key - executes the function
    cachedGetNumber(99)  // Output: fetching number: 99
}
```

### Using FromContextLoader (Convenience API)

`FromContextLoader` provides a simpler API that always returns a callable function, automatically falling back to the original loader if cache is not available.

```go
package main

import (
    "context"
    "fmt"
    "strconv"

    "github.com/alingse/ctxcache"
)

func getNumber(n int64) string {
    fmt.Printf("fetching number: %d\n", n)
    return strconv.FormatInt(n, 10)
}

func main() {
    ctx := context.Background()

    // Register cache (optional - can be skipped if you want fallback behavior)
    ctx = ctxcache.WithCache[int64, string](ctx, ctxcache.FuncID("getNumber"), getNumber)

    // Get loader - works whether cache exists or not
    loader := ctxcache.FromContextLoader(ctx, ctxcache.FuncID("getNumber"), getNumber)

    // Just use it - no need to check bool!
    loader(42)  // Output: fetching number: 42
    loader(42)  // No output (cached!)
}
```

**When to use which:**
- Use `FromContext` when you need to know whether caching is enabled
- Use `FromContextLoader` when you just want a callable function with automatic fallback

```go
package main

import (
    "context"
    "fmt"
    "strconv"

    "github.com/alingse/ctxcache"
)

func getNumber(n int64) string {
    fmt.Printf("fetching number: %d\n", n)
    return strconv.FormatInt(n, 10)
}

func main() {
    ctx := context.Background()

    // Register cache for getNumber function
    ctx = ctxcache.WithCache[int64, string](ctx, ctxcache.FuncID("getNumber"), getNumber)

    // Get cached function
    cachedGetNumber, found := ctxcache.FromContext[int64, string](ctx, ctxcache.FuncID("getNumber"))
    if !found {
        panic("cache not found")
    }

    // First call - executes the function
    cachedGetNumber(42)  // Output: fetching number: 42

    // Second call - uses cache
    cachedGetNumber(42)  // No output (cached!)

    // Different key - executes the function
    cachedGetNumber(99)  // Output: fetching number: 99
}
```

## Handling `(V, error)` Return Values

Since ctxcache only supports `func(K) V`, wrap errors in a result struct:

```go
package main

import (
    "context"
    "errors"
    "fmt"

    "github.com/alingse/ctxcache"
)

// Result wraps both value and error
type Result[T any] struct {
    Value T
    Error error
}

// Simulated database function
func getUser(id int64) Result[User] {
    // Simulate DB lookup error
    if id <= 0 {
        return Result[User]{Error: errors.New("invalid user id")}
    }
    return Result[User]{Value: User{ID: id, Name: "Alice"}}
}

type User struct {
    ID   int64
    Name string
}

func main() {
    ctx := context.Background()
    ctx = ctxcache.WithCache(ctx, ctxcache.FuncID("getUser"), getUser)

    cachedGetUser, _ := ctxcache.FromContext[int64, Result[User]](ctx, ctxcache.FuncID("getUser"))

    // Valid user - cached
    r1 := cachedGetUser(1)
    if r1.Error != nil {
        fmt.Println("Error:", r1.Error)
    } else {
        fmt.Printf("User: %+v\n", r1.Value)  // User: {ID:1 Name:Alice}
    }

    // Invalid ID - cached (including error)
    r2 := cachedGetUser(-1)
    fmt.Println("Error:", r2.Error)  // Error: invalid user id
}
```

## API Reference

### `type FuncID string`

A unique identifier for a cached function. Use this to register and retrieve cached functions from context.

### `func WithCache[K comparable, V any](ctx context.Context, ctxKey FuncID, f CacheFunc[K, V]) context.Context`

Registers a cached function in the context and returns the new context.

- `K`: Key type (must be comparable)
- `V`: Value type (any)
- `f`: The function to cache

### `func FromContext[K comparable, V any](ctx context.Context, ctxKey FuncID) (CacheFunc[K, V], bool)`

Retrieves a cached function from the context.

Returns:
- `CacheFunc[K, V]`: The cached function (or `nil` if not found)
- `bool`: `true` if cache was found, `false` otherwise

### `func FromContextLoader[K comparable, V any](ctx context.Context, ctxKey FuncID, f CacheFunc[K, V]) CacheFunc[K, V]`

Returns a cached function if available, otherwise returns the original loader function.

This is a convenience function that always returns a callable function without needing to check a bool return value. Use this when you want to use caching if available but fall back to the original function transparently.

Returns:
- `CacheFunc[K, V]`: The cached function (if found), or the original `f` (if not found)

### `type CacheFunc[K comparable, V any] func(K) V`

The cached function signature.

## Usage Scenarios

- **Request-scoped caching**: Cache expensive computations within a single HTTP request
- **Avoiding redundant database queries**: When the same data might be requested multiple times
- **Shared context across goroutines**: Safe for concurrent reads within the same context

## Example: HTTP Middleware

```go
func CacheMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // Initialize cache for this request
        ctx := ctxcache.WithCache(r.Context(), ctxcache.FuncID("userLoader"), loadUserFromDB)
        next.ServeHTTP(w, r.WithContext(ctx))
    })
}

func handler(w http.ResponseWriter, r *http.Request) {
    // Multiple calls in the same request only hit DB once
    getUser, _ := ctxcache.FromContext[int64, User](r.Context(), ctxcache.FuncID("userLoader"))

    user1 := getUser(1)  // DB call
    user2 := getUser(1)  // From cache
}
```

## License

MIT
