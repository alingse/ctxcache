package ctxcache

import (
	"context"
	"sync"
)

type cache[K comparable, V any] struct {
	lock   sync.RWMutex
	data   map[K]V
	loader func(K) V
}

func (c *cache[K, V]) cacheLoader(k K) V {
	c.lock.RLock()
	v, ok := c.data[k]
	if ok {
		c.lock.RUnlock()
		return v
	}
	c.lock.RUnlock()
	// TODO: lock by k
	c.lock.Lock()
	defer c.lock.Unlock()
	v = c.loader(k)
	c.data[k] = v

	return v
}

type FuncID string

type CacheFunc[K comparable, V any] func(K) V

func WithCache[K comparable, V any](ctx context.Context, ctxKey FuncID, f CacheFunc[K, V]) context.Context {
	cache := &cache[K, V]{
		loader: f,
		data:   make(map[K]V),
	}
	ctx = context.WithValue(ctx, ctxKey, cache)
	return ctx
}

func FromContext[K comparable, V any](ctx context.Context, ctxKey FuncID, f CacheFunc[K, V]) (CacheFunc[K, V], bool) {
	cache, ok := ctx.Value(ctxKey).(*cache[K, V])
	if !ok {
		return f, false
	}
	return cache.cacheLoader, true
}
