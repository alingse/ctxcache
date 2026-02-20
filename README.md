# ctxcache

A lightweight Go context-level cache library for `func(K) V` functions.

## When to Use

Use ctxcache when you need **request-scoped caching** - avoiding redundant computations or database queries within a single request context:

- **HTTP handlers**: Cache user lookups, permissions, or config that may be accessed multiple times during one request
- **GraphQL resolvers**: Prevent N+1 queries by caching loaders within the request
- **Concurrent processing**: Safe for goroutines sharing the same context

## Usage

```bash
go get github.com/alingse/ctxcache
```

### WithCache + FromContext

```go
ctx := context.Background()

// Register cache
ctx = ctxcache.WithCache[int64, string](ctx, ctxcache.FuncID("getNumber"), getNumber)

// Get cached function
cached, _ := ctxcache.FromContext[int64, string](ctx, ctxcache.FuncID("getNumber"))

cached(42) // executes getNumber(42)
cached(42) // cache hit
```

### FromContextLoader (Auto Fallback)

```go
ctx := ctxcache.WithCache[int64, string](ctx, ctxcache.FuncID("getNumber"), getNumber)

// Returns cached func or falls back to original if cache not in context
loader := ctxcache.FromContextLoader(ctx, ctxcache.FuncID("getNumber"), getNumber)

loader(42) // executes
loader(42) // cache hit
```

### HTTP Middleware

```go
func CacheMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        ctx := ctxcache.WithCache(r.Context(), ctxcache.FuncID("userLoader"), loadUserFromDB)
        next.ServeHTTP(w, r.WithContext(ctx))
    })
}

func handler(w http.ResponseWriter, r *http.Request) {
    getUser, _ := ctxcache.FromContext[int64, User](r.Context(), ctxcache.FuncID("userLoader"))
    
    user1 := getUser(1) // DB call
    user2 := getUser(1) // cache hit within same request
}
```

## API

```go
type FuncID string
type CacheFunc[K comparable, V any] func(K) V

func WithCache[K comparable, V any](ctx context.Context, id FuncID, f CacheFunc[K, V]) context.Context
func FromContext[K comparable, V any](ctx context.Context, id FuncID) (CacheFunc[K, V], bool)
func FromContextLoader[K comparable, V any](ctx context.Context, id FuncID, f CacheFunc[K, V]) CacheFunc[K, V]
```

**Note**: For `(V, error)` functions, wrap the result in a struct.

## License

MIT
