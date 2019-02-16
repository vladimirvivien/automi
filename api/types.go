package api

import (
	"context"
	"fmt"
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
	Exec() error
}

type ProcError struct {
	Err      error
	ProcName string
}

func (e ProcError) Error() string {
	if e.ProcName != "" {
		return fmt.Sprintf("[%s] %v", e.ProcName, e.Err)
	}
	return e.Err.Error()
}

// LogFunc represents a function to handle log events
type LogFunc func(interface{})

// ErrorFunc this type is a user-provided function to handle errors
type ErrorFunc func(StreamError)

// StreamError is used to signal runtime stream error
type StreamError struct {
	Err    error
	OpName string
	ItemID int64
}

func (e StreamError) Error() string {
	return e.Err.Error()
}

// PanicStreamError signals that the stream should panic immediately
type PanicStreamError StreamError

func (e PanicStreamError) Error() string {
	return e.Err.Error()
}

// CancelStreamError signals that all stream activities should stop
// and the streaming should gracefully end
type CancelStreamError StreamError

func (e CancelStreamError) Error() string {
	return e.Err.Error()
}

// StreamItem can be used to provide a rich repressentation of streaming data.
// Stream data can be wrapped in StreamItem carry additional information downstream
// including context, metadata, and error.
type StreamItem struct {
	Index    int64             // index of the item in the stream
	Context  context.Context   // context of the stream
	Data     interface{}       // data being stream
	MetaData map[string]string // user-provided stream metadat
	Error    StreamError       // any error from previous step
}
