package entity

import (
	"context"
	"log"
	"sync"
)

const (
	wgKey = "wg"
)

func CreateWg(ctx context.Context) context.Context {
	wg := &sync.WaitGroup{}
	return context.WithValue(ctx, wgKey, wg)
}

func GetWg(ctx context.Context) *sync.WaitGroup {
	wg, ok := ctx.Value(wgKey).(*sync.WaitGroup)
	if !ok {
		log.Fatal("context does not contain WaitGroup")
	}
	return wg
}
