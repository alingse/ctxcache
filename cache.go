package ctxcache

import (
	"context"
	"sync"
)

type cache[K comparable, V any] struct {
	context.Context
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

func WithCache[K comparable, V any](ctx context.Context, f func(K) V) context.Context {
	cache := &cache[K, V]{
		Context: ctx,
		loader:  f,
		data:    make(map[K]V),
	}
	return cache
}

func FromContext[K comparable, V any](ctx context.Context, f func(K) V) func(K) V {
	cache, ok := ctx.(*cache[K, V])
	if !ok {
		return f
	}
	return cache.cacheLoader
}
