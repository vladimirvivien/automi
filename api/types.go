package api

import (
	"fmt"

	"golang.org/x/net/context"
)

type Process interface {
	GetName() string
	Exec(context.Context) error
	Init(context.Context) error
	Uninit(context.Context) error
}

type Source interface {
	GetOutput() <-chan interface{}
}

// StreamSource Represents a source of data stream
type StreamSource interface {
	Source
	Open(context.Context) error
}

type Sink interface {
	SetInput(<-chan interface{})
}

// SteamSink  represents a final sink for a stream
type StreamSink interface {
	Sink
	Open(context.Context) <-chan error
}

type Processor interface {
	Process
	Source
	Sink
}

type Endpoint interface {
	Done() <-chan struct{}
}

type Collector interface {
	SetInputs([]<-chan interface{})
}

type Emitter interface {
	GetOutputs() []<-chan interface{}
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
