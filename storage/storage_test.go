package storage

import (
	"context"
	"testing"
)

func TestService(t *testing.T) {
	ctx, _ := context.WithCancel(context.Background())
	s := NewService()
	if s == nil {
		return
	}
	<-ctx.Done()
}
