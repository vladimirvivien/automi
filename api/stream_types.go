package api

import (
	"context"
)

// Emitter is a node that has the ability to emit data to an output channel
type Emitter interface {
	GetOutput() <-chan any
}

// Source is a component that has data that can be placed on the stream
type Source interface {
	Emitter
	Reporter
	Open(context.Context) error
}

// Collector has the ability to collect data from an input channel
type Collector interface {
	SetInput(<-chan any)
}

// Sink is a resource that can receive stream data
type Sink interface {
	Collector
	Reporter
	Open(context.Context) <-chan error
}

// Operator is an executor node that applies a function on items in the stream
type Operator interface {
	Collector
	Emitter
	Reporter
	Exec(context.Context) error
}

// StreamItem can be used to provide a rich representation of streaming data.
// Stream data can be wrapped in StreamItem carry additional information downstream
// including context, metadata, and error.
type StreamItem[T any] struct {
	Index    int64             // index of the item in the stream
	Item     T                 // data item being stream
	MetaData map[string]string // user-provided stream metadat
}

// FilterItem represents the result of a Filter operation.
type FilterItem[T any] struct {
	Predicate bool
	Item      T
}

// StreamAction provides hints how stream items should
// treated during processing. Actions will be used by
// opertor executors to determine how to handle items.
type StreamAction uint8

const (
	ActionForwardItem StreamAction = iota // Forward items to down the stream (default)
	ActionSkipItem                        // Drops items from the stream
	ActionRerouteItem                     // Reroutes item to another stream
)

// StreamResult can be used in opertor executors
// to provide hints to the underlying operator
// how to handle the result of an operation. It
// wraps the result along with any error that needs
// to be signaled.
type StreamResult struct {
	Value  any
	Action StreamAction
	Err    error
}
