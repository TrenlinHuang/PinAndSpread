package main

import (
	"context"
	netstore "github.com/trenlinhuang/pin-spread/store/net-store"
)

func main() {
	ctx, _ := context.WithCancel(context.Background())
	netstore.New(3000)
	<-ctx.Done()
}
