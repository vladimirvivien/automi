package window

import (
	"context"
	"time"
)

// TriggerAllFunc is a function that triggers the window after all
// items streamed have been collected. This is useful for testing or
// with small item count.
func TriggerAllFunc[T any]() TriggerFunction[T] {
	return func(ctx context.Context, wctx WindowContext[T]) bool {
		return false
	}
}

// TriggerBySizeFunc is a function that can trigger the window to
// close base on the specified size (number of items) currently in the window.
func TriggerBySizeFunc[T any](size uint64) TriggerFunction[T] {
	return func(ctx context.Context, wctx WindowContext[T]) bool {
		return wctx.WindowItemCount >= size
	}
}

// TriggerByDurationFunc is a function that can trigger a window based
// on specified duration of the current window runtime.
func TriggerByDurationFunc[T any](duration time.Duration) TriggerFunction[T] {
	return func(ctx context.Context, wctx WindowContext[T]) bool {
		elapsed := time.Since(wctx.WindowStartTime)
		return elapsed >= duration
	}
}

// TriggerByFunc is a function that can trigger a window based on a custom
// trigger function.
func TriggerByFunc[T any](trigger TriggerFunction[T]) TriggerFunction[T] {
	return trigger
}
