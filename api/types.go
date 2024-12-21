package api

import (
	"context"
)

// Emitter has the ability to emit data to an output channel
type Emitter[T any] interface {
	GetOutput() <-chan T
}

// Source is a component that has data that can be placed on the stream
type Source[T any] interface {
	Emitter[T]
	Open(context.Context) error
}

// Collector has the ability to collect data from an input channel
type Collector[T any] interface {
	SetInput(<-chan T)
}

// Sink is a resource that can receive stream data
type Sink[T any] interface {
	Collector[T]
	Open(context.Context) <-chan error
}

// Operator is an executor node that applies a function on items in the stream
type Operator[T any] interface {
	Collector[T]
	Emitter[T]
	Exec(context.Context) error
}

// LogFunc represents a function to handle log events
type LogFunc func(interface{})

// ErrorFunc this type is a user-provided function to handle errors
type ErrorFunc func(StreamError)

// StreamError is used to signal runtime stream error
type StreamError struct {
	err  string      // Error message
	item *StreamItem // Item that caused error
}

// Error returns a string value for StreamError
func (e StreamError) Error() string {
	return e.err
}

// Item returns the StreamItem associated with the error
func (e StreamError) Item() *StreamItem {
	return e.item
}

// Error returns a StreamError
func Error(msg string) StreamError {
	return StreamError{err: msg}
}

// ErrorWithItem returns a StreamError with provided StreamItem
func ErrorWithItem(msg string, item *StreamItem) StreamError {
	return StreamError{err: msg, item: item}
}

// PanicStreamError signals that the stream should panic immediately
type PanicStreamError StreamError

// Error returns a string value for PanicStreamError
func (e PanicStreamError) Error() string {
	return e.err
}

// PanickingError returns a PanicStreamError
func PanickingError(msg string) PanicStreamError {
	return PanicStreamError(Error(msg))
}

// CancelStreamError signals that all stream activities should stop
// and the streaming should gracefully end
type CancelStreamError StreamError

// Error returns a string value for CancelStreamError
func (e CancelStreamError) Error() string {
	return e.err
}

// CancellationError returns a CancelStreamError
func CancellationError(msg string) CancelStreamError {
	return CancelStreamError(Error(msg))
}

// StreamItem can be used to provide a rich representation of streaming data.
// Stream data can be wrapped in StreamItem carry additional information downstream
// including context, metadata, and error.
type StreamItem struct {
	Index    int64             // index of the item in the stream
	Item     interface{}       // data item being stream
	MetaData map[string]string // user-provided stream metadat
	Context  context.Context   // stream context
}
