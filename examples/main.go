package main

import (
	"context"
	"fmt"
	"strconv"

	"github.com/alingse/ctxcache"
)

func getNumber(n int64) string {
	fmt.Printf("got number: %d \n", n)
	return strconv.FormatInt(n, 10)
}

func callGetNumber(ctx context.Context, n int64) {
	getNumber2, _ := ctxcache.FromContext[int64, string](ctx, ctxcache.FuncID("getNumber"), getNumber)
	getNumber2(n)
}

func main() {
	var ctx = context.Background()

	// register funcID' cache
	ctx = ctxcache.WithCache[int64, string](ctx, ctxcache.FuncID("getNumber"), getNumber)

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	callGetNumber(ctx, 1)
	callGetNumber(ctx, 1)
	callGetNumber(ctx, 2)
}
