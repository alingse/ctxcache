package main

import (
	"context"
	"fmt"
	"strconv"

	"github.com/alingse/ctxcache"
)

// Simulated expensive function - e.g., database query or API call
func fetchUserFromDB(userID int64) string {
	fmt.Printf("  [DB] Fetching user ID: %d\n", userID)
	return fmt.Sprintf("User-%d", userID)
}

// Simulated expensive computation
func computeExpensive(n int) string {
	fmt.Printf("  [Compute] Calculating for: %d\n", n)
	return strconv.Itoa(n * n)
}

func main() {
	fmt.Println("=== FromContextLoader Example ===\n")

	// Example 1: With cache enabled
	fmt.Println("Example 1: With cache enabled")
	fmt.Println("--------------------------------")
	ctx1 := context.Background()
	ctx1 = ctxcache.WithCache[int64, string](ctx1, ctxcache.FuncID("userLoader"), fetchUserFromDB)

	loader1 := ctxcache.FromContextLoader(ctx1, ctxcache.FuncID("userLoader"), fetchUserFromDB)

	fmt.Println("First call:")
	result1 := loader1(42) // Will call the original function
	fmt.Printf("Result: %s\n\n", result1)

	fmt.Println("Second call (cached):")
	result2 := loader1(42) // Will use cache
	fmt.Printf("Result: %s\n\n", result2)

	// Example 2: Without cache (fallback to original function)
	fmt.Println("Example 2: Without cache (fallback behavior)")
	fmt.Println("------------------------------------------------")
	ctx2 := context.Background() // No cache registered!

	loader2 := ctxcache.FromContextLoader(ctx2, ctxcache.FuncID("computeLoader"), computeExpensive)

	fmt.Println("First call:")
	result3 := loader2(5) // Will call the original function
	fmt.Printf("Result: %s\n\n", result3)

	fmt.Println("Second call (no cache, calls function again):")
	result4 := loader2(5) // Will call the original function again (no cache!)
	fmt.Printf("Result: %s\n\n", result4)

	// Example 3: Comparison with FromContext
	fmt.Println("Example 3: Comparing FromContext vs FromContextLoader")
	fmt.Println("------------------------------------------------------")

	// Old way with FromContext - need to check bool
	fmt.Println("Using FromContext (old way):")
	cachedFunc, ok := ctxcache.FromContext[int64, string](ctx1, ctxcache.FuncID("userLoader"))
	if ok {
		fmt.Println("  Cache found, using cached function")
		cachedFunc(99)
	} else {
		fmt.Println("  Cache not found, using original function")
		fetchUserFromDB(99)
	}

	// New way with FromContextLoader - no need to check
	fmt.Println("\nUsing FromContextLoader (new way):")
	loader3 := ctxcache.FromContextLoader(ctx1, ctxcache.FuncID("userLoader"), fetchUserFromDB)
	loader3(99) // Works whether cache exists or not!

	fmt.Println("\n=== Summary ===")
	fmt.Println("FromContextLoader provides a more convenient API:")
	fmt.Println("- No need to check bool return value")
	fmt.Println("- Automatically falls back to original function if cache doesn't exist")
	fmt.Println("- Cleaner, more readable code")
}
