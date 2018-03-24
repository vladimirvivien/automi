package batch

import (
	"context"

	"github.com/vladimirvivien/automi/api"
)

// TriggerAll forces the batch trigger to always return false
// meaning it's never done, causing the batch to exhaust all stream items.
func TriggerAll() api.BatchTriggerFunc {
	return api.BatchTriggerFunc(func(ctx context.Context, item interface{}, i int64) bool {
		return false
	})
}

// TriggerBySize triggers batch based on the specified size.
// The batch remains open as long as the current index of the
// item is less then the specified size.
func TriggerBySize(size int64) api.BatchTriggerFunc {
	return api.BatchTriggerFunc(func(ctx context.Context, item interface{}, i int64) bool {
		return i >= size
	})
}
