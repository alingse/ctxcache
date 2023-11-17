package main

import (
	"context"
	"fmt"
	"strconv"

	"github.com/alingse/ctxcache"
)

var funcID = "get_number"

func getNumber(n int64) string {
	fmt.Printf("got number: %d \n", n)
	return strconv.FormatInt(n, 10)
}

func callGetNumber(ctx context.Context, n int64) {
	getNumberCache, _ := ctxcache.FromContext[int64, string](ctx, funcID)
	getNumberCache(n)
}

func main() {
	var ctx = context.Background()

	// register funcID' cache
	ctx = ctxcache.WithCache[int64, string](ctx, funcID, getNumber)

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	callGetNumber(ctx, 1)
	callGetNumber(ctx, 1)
	callGetNumber(ctx, 2)
}
