package window

import "time"

// Batch creaets a window that only closes when the stream is closed
// thus batching all streamed items. This Generally useful for
// testing or small streamed item count.
func Batch[IN any]() *WindowOperator[IN] {
	return New(TriggerAllFunc[IN]())
}

// BySize creates a new fixed-size window using the specified item count.
func BySize[IN any](size uint64) *WindowOperator[IN] {
	trigger := TriggerBySizeFunc[IN](size)
	return New(trigger)
}

// ByDuration creates a new window to collect items for the specified duration.
func ByDuration[IN any](dur time.Duration) *WindowOperator[IN] {
	trigger := TriggerByDurationFunc[IN](dur)
	return New(trigger)
}

// ByFunc creates a new window using the specified trigger function.
func ByFunc[IN any](trigger TriggerFunction[IN]) *WindowOperator[IN] {
	return New(trigger)
}
