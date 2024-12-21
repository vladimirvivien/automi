package window

import (
	"context"
	"time"
)

type WindowContext[IN any] struct {
	OperatorStartTime time.Time // The time when the window operator started
	OperatorItemCount uint64    // The total number of items operator have received
	WindowStartTime   time.Time // The time the current window started
	WindowItemCount   uint64    // The current item count of the window
	Item              IN        // The last data item admitted to the current window
	ItemWindowTime    time.Time // The time an item got admitted to the current window
}

type TriggerFunction[IN any] func(context.Context, WindowContext[IN]) bool
