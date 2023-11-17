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
	ctx, cacheF := ctxcache.WithCache[int64, string](ctx, getNumber)
	cacheF(1)
	cacheF(1)
	cacheF(2)

	ctx = context.WithValue(ctx, struct{}{}, 1)

	getNumberCache, _ := ctxcache.FromContext[int64, string](ctx)
	getNumberCache(1)
	getNumberCache(1)
	getNumberCache(2)
}
