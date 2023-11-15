package main

import (
	"context"
	"fmt"
	"strconv"

	"github.com/alingse/ctxcache"
)

func main() {
	var ctx = context.Background()
	var getNumber = func(n int64) string {
		fmt.Printf("got number: %d \n", n)
		return strconv.FormatInt(n, 10)
	}
	ctx = ctxcache.WithCache[int64, string](ctx, getNumber)

	var getNumberCache = ctxcache.FromContext(ctx, getNumber)

	getNumberCache(1)
	getNumberCache(1)
	getNumberCache(1)
}
