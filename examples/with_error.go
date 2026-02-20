//go:build ignore

// Example demonstrating how to handle (V, error) return values with ctxcache
package main

import (
	"context"
	"errors"
	"fmt"

	"github.com/alingse/ctxcache"
)

// Result wraps both value and error for functions that may fail
type Result[T any] struct {
	Value T
	Error error
}

// User represents a user record
type User struct {
	ID   int64
	Name string
}

// Simulated database lookup - may return an error
func fetchUserFromDB(id int64) Result[User] {
	fmt.Printf("  [DB Query] Fetching user %d...\n", id)

	// Simulate error cases
	if id <= 0 {
		return Result[User]{Error: errors.New("invalid user id: must be positive")}
	}
	if id > 100 {
		return Result[User]{Error: errors.New("user not found")}
	}

	// Simulate successful lookup
	users := map[int64]User{
		1:  {ID: 1, Name: "Alice"},
		2:  {ID: 2, Name: "Bob"},
		42: {ID: 42, Name: "Douglas"},
	}

	if user, ok := users[id]; ok {
		return Result[User]{Value: user}
	}
	return Result[User]{Error: fmt.Errorf("user %d not found", id)}
}

func main() {
	fmt.Println("=== ctxcache with (V, error) Pattern ===\n")

	ctx := context.Background()

	// Register cache for the user fetch function
	ctx = ctxcache.WithCache(ctx, ctxcache.FuncID("fetchUser"), fetchUserFromDB)

	// Get the cached function
	cachedFetch, _ := ctxcache.FromContext[int64, Result[User]](ctx, ctxcache.FuncID("fetchUser"))

	fmt.Println("Test 1: Fetch valid user (ID=1)")
	r1 := cachedFetch(1)
	if r1.Error != nil {
		fmt.Printf("  Error: %v\n", r1.Error)
	} else {
		fmt.Printf("  Found: %+v\n", r1.Value)
	}

	fmt.Println("\nTest 2: Fetch same user again (ID=1) - should use cache")
	r2 := cachedFetch(1)
	if r2.Error != nil {
		fmt.Printf("  Error: %v\n", r2.Error)
	} else {
		fmt.Printf("  Found: %+v (from cache)\n", r2.Value)
	}

	fmt.Println("\nTest 3: Fetch another valid user (ID=42)")
	r3 := cachedFetch(42)
	if r3.Error != nil {
		fmt.Printf("  Error: %v\n", r3.Error)
	} else {
		fmt.Printf("  Found: %+v\n", r3.Value)
	}

	fmt.Println("\nTest 4: Fetch with invalid ID (-1) - caches error too")
	r4 := cachedFetch(-1)
	fmt.Printf("  Result: %v\n", r4.Error)

	fmt.Println("\nTest 5: Fetch same invalid ID again (-1)")
	r5 := cachedFetch(-1)
	fmt.Printf("  Result: %v (from cache)\n", r5.Error)

	fmt.Println("\n=== Key Takeaways ===")
	fmt.Println("1. Wrap (V, error) returns in a Result struct")
	fmt.Println("2. Both successful results AND errors are cached")
	fmt.Println("3. The cached function is only called once per unique key")
}
