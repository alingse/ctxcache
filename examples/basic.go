//go:build ignore

// Basic usage example for ctxcache
package main

import (
	"context"
	"fmt"
	"strconv"

	"github.com/alingse/ctxcache"
)

// Simulate an expensive computation
func expensiveCalculation(n int64) string {
	fmt.Printf("  [Computing] Calculating for %d...\n", n)
	return strconv.FormatInt(n, 10)
}

func main() {
	fmt.Println("=== Basic ctxcache Example ===\n")

	ctx := context.Background()

	// Register cache for the expensive function
	ctx = ctxcache.WithCache[int64, string](ctx, ctxcache.FuncID("calc"), expensiveCalculation)

	// Get the cached function
	cachedCalc, found := ctxcache.FromContext[int64, string](ctx, ctxcache.FuncID("calc"))
	if !found {
		panic("cache not found")
	}

	fmt.Println("Calling cachedCalc(42) first time:")
	cachedCalc(42) // Executes the function

	fmt.Println("\nCalling cachedCalc(42) second time:")
	cachedCalc(42) // Uses cache - no output

	fmt.Println("\nCalling cachedCalc(99) first time:")
	cachedCalc(99) // Executes the function

	fmt.Println("\nCalling cachedCalc(99) second time:")
	cachedCalc(99) // Uses cache - no output

	fmt.Println("\n=== Summary ===")
	fmt.Println("Notice how each value is only computed once!")
}
