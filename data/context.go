package data

import (
	"context"
	"time"
)

func NewContext(parent context.Context, timeoutMillis int64, kv map[any]any) (context.Context, context.CancelFunc) {
	var cancelFunc context.CancelFunc
	ctx := parent
	if ctx == nil {
		ctx = context.Background()
	}

	// Stash the values
	for k, v := range kv {
		ctx = context.WithValue(ctx, k, v)
	}
	if timeoutMillis > 0 {
		ctx, cancelFunc = context.WithTimeout(ctx, time.Duration(timeoutMillis)*time.Millisecond)
	}
	return ctx, cancelFunc
}
