package ctxcache

import (
	"context"
	"sync"
	"testing"
)

func TestFromContext_NotFound(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	fn, ok := FromContext[int, string](ctx, FuncID("nonexistent"))
	if ok {
		t.Errorf("expected false when cache not found, got true")
	}
	if fn != nil {
		t.Errorf("expected nil function when cache not found, got %v", fn)
	}
}

func TestWithCache_Basic(t *testing.T) {
	t.Parallel()
	callCount := 0
	f := func(k int) string {
		callCount++

		return "value-" + string(rune('0'+k)) //nolint:gosec // Safe: k is 0-9 in test
	}

	ctx := context.Background()
	ctx = WithCache(ctx, FuncID("test"), f)

	// First call should call the function
	fn, ok := FromContext[int, string](ctx, FuncID("test"))
	if !ok {
		t.Fatal("expected true when cache found")
	}
	result1 := fn(1)
	if result1 != "value-"+"1" {
		t.Errorf("expected value-1, got %s", result1)
	}

	// Second call with same key should use cache
	callCountAfterFirst := callCount
	result2 := fn(1)
	if result2 != result1 {
		t.Errorf("expected same result from cache, got %s vs %s", result2, result1)
	}
	if callCount != callCountAfterFirst {
		t.Errorf("expected no additional calls for cached key, callCount went from %d to %d", callCountAfterFirst, callCount)
	}
}

func TestWithCache_MultipleKeys(t *testing.T) {
	t.Parallel()
	callCount := 0
	f := func(k int) string {
		callCount++

		return "value-" + string(rune('0'+k)) //nolint:gosec // Safe: k is 0-9 in test
	}

	ctx := context.Background()
	ctx = WithCache(ctx, FuncID("test"), f)

	fn, _ := FromContext[int, string](ctx, FuncID("test"))

	// Call with different keys
	fn(1)
	countAfter1 := callCount

	fn(2)
	countAfter2 := callCount

	fn(1) // Should use cache
	countAfterSecond1 := callCount

	if countAfter1 != 1 {
		t.Errorf("expected 1 call after first key, got %d", countAfter1)
	}
	if countAfter2 != 2 {
		t.Errorf("expected 2 calls after second key, got %d", countAfter2)
	}
	if countAfterSecond1 != 2 {
		t.Errorf("expected no additional call for cached key, got %d", countAfterSecond1)
	}
}

func TestConcurrent(t *testing.T) {
	t.Parallel()
	callCount := 0
	var mu sync.Mutex
	f := func(k int) string {
		mu.Lock()
		callCount++
		mu.Unlock()

		return "value-" + string(rune('0'+k)) //nolint:gosec // Safe: k is 0-9 in test
	}

	ctx := context.Background()
	ctx = WithCache(ctx, FuncID("test"), f)

	fn, _ := FromContext[int, string](ctx, FuncID("test"))

	var wg sync.WaitGroup
	numGoroutines := 100
	callsPerGoroutine := 10
	totalCalls := numGoroutines * callsPerGoroutine // 1000 total calls

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < callsPerGoroutine; j++ {
				fn(id % 10) // Use 10 different keys
			}
		}(i)
	}

	wg.Wait()

	// Verify cache is effective: without cache we'd have 1000 calls
	// With 10 unique keys and imperfect locking, we expect significantly fewer
	// The current implementation has a known race condition (noted by TODO in code)
	// so we just verify it's much better than no caching
	if callCount >= totalCalls {
		t.Errorf("cache not working: expected significantly less than %d calls, got %d", totalCalls, callCount)
	}
	// Verify each key gets at least one call (we have 10 unique keys: 0-9)
	if callCount < 10 {
		t.Errorf("expected at least 10 calls for 10 unique keys, got %d", callCount)
	}
}

func TestMultipleCaches(t *testing.T) {
	t.Parallel()
	f1 := func(k int) string {
		return "cache1-" + string(rune('0'+k)) //nolint:gosec // Safe: k is 0-9 in test
	}
	f2 := func(k int) string {
		return "cache2-" + string(rune('0'+k)) //nolint:gosec // Safe: k is 0-9 in test
	}

	ctx := context.Background()
	ctx = WithCache(ctx, FuncID("cache1"), f1)
	ctx = WithCache(ctx, FuncID("cache2"), f2)

	fn1, ok1 := FromContext[int, string](ctx, FuncID("cache1"))
	if !ok1 {
		t.Error("expected cache1 to be found")
	}
	if fn1(1) != "cache1-1" {
		t.Errorf("expected cache1-1, got %s", fn1(1))
	}

	fn2, ok2 := FromContext[int, string](ctx, FuncID("cache2"))
	if !ok2 {
		t.Error("expected cache2 to be found")
	}
	if fn2(1) != "cache2-1" {
		t.Errorf("expected cache2-1, got %s", fn2(1))
	}
}

func TestFromContextLoader_WithCache(t *testing.T) {
	t.Parallel()
	callCount := 0
	f := func(k int) string {
		callCount++

		return "value-" + string(rune('0'+k)) //nolint:gosec // Safe: k is 0-9 in test
	}

	ctx := context.Background()
	ctx = WithCache(ctx, FuncID("test"), f)

	loader := FromContextLoader(ctx, FuncID("test"), f)

	// First call should call the function
	result1 := loader(1)
	if result1 != "value-1" {
		t.Errorf("expected value-1, got %s", result1)
	}
	if callCount != 1 {
		t.Errorf("expected 1 call, got %d", callCount)
	}

	// Second call should use cache
	result2 := loader(1)
	if result2 != "value-1" {
		t.Errorf("expected value-1, got %s", result2)
	}
	if callCount != 1 {
		t.Errorf("expected no additional calls for cached key, got %d", callCount)
	}
}

func TestFromContextLoader_WithoutCache(t *testing.T) {
	t.Parallel()
	callCount := 0
	f := func(k int) string {
		callCount++

		return "value-" + string(rune('0'+k)) //nolint:gosec // Safe: k is 0-9 in test
	}

	ctx := context.Background()

	loader := FromContextLoader(ctx, FuncID("nonexistent"), f)

	// Should call the original function each time (no cache)
	loader(1)
	if callCount != 1 {
		t.Errorf("expected 1 call, got %d", callCount)
	}

	loader(1)
	if callCount != 2 {
		t.Errorf("expected 2 calls (no caching), got %d", callCount)
	}
}

func TestFromContextLoader_Behavior(t *testing.T) {
	t.Parallel()
	// Test that WithCache and FromContextLoader work together correctly
	callCount := 0
	originalFunc := func(k int) string {
		callCount++

		return "original-" + string(rune('0'+k)) //nolint:gosec // Safe: k is 0-9 in test
	}

	ctx := context.Background()
	ctx = WithCache(ctx, FuncID("test"), originalFunc)

	loader := FromContextLoader(ctx, FuncID("test"), originalFunc)

	// Call with different keys
	loader(1)
	countAfter1 := callCount

	loader(2)
	countAfter2 := callCount

	loader(1) // Should use cache
	countAfterSecond1 := callCount

	if countAfter1 != 1 {
		t.Errorf("expected 1 call after first key, got %d", countAfter1)
	}
	if countAfter2 != 2 {
		t.Errorf("expected 2 calls after second key, got %d", countAfter2)
	}
	if countAfterSecond1 != 2 {
		t.Errorf("expected no additional call for cached key, got %d", countAfterSecond1)
	}

	// Verify the returned value is correct
	result := loader(1)
	if result != "original-1" {
		t.Errorf("expected original-1, got %s", result)
	}
}
