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
