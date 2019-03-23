package api

import (
	"context"
)

type Emitter interface {
	GetOutput() <-chan interface{}
}

type Source interface {
	Emitter
	Open(context.Context) error
}

type Collector interface {
	SetInput(<-chan interface{})
}

type Sink interface {
	Collector
	Open(context.Context) <-chan error
}

//Deprecated Source represents a node that can source data
//type Source interface {
//	GetOutput() <-chan interface{}
//}

// StreamSource Represents a source of data stream
//type StreamSource interface {
//	Source
//	Open(context.Context) error
//}

// Sink represents a node that can absorb/consume data
//type Sink interface {
//	SetInput(<-chan interface{})
//}

// SteamSink  represents a node that can stream data to be absorbed/consumed
//type StreamSink interface {
//	Sink
//	Open(context.Context) <-chan error
//}

// Operator is an executor node that applies a function on items in the stream
type Operator interface {
	Collector
	Emitter
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

func (e CancelStreamError) Error() string {
	return e.err
}

//CancellationError returns a CancelStreamError
func CancellationError(msg string) CancelStreamError {
	return CancelStreamError(Error(msg))
}

// StreamItem can be used to provide a rich repressentation of streaming data.
// Stream data can be wrapped in StreamItem carry additional information downstream
// including context, metadata, and error.
type StreamItem struct {
	Index    int64             // index of the item in the stream
	Item     interface{}       // data item being stream
	MetaData map[string]string // user-provided stream metadat
	Context  context.Context   // stream context
}
