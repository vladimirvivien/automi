package stream

import (
	"context"

	"github.com/vladimirvivien/automi/api"
	autoctx "github.com/vladimirvivien/automi/api/context"
	"github.com/vladimirvivien/automi/util"
)

// Drain is a generic sink that terminates streamed data
type Drain struct {
	output chan interface{}
	input  <-chan interface{}
	logFn  api.LogFunc
}

// NewDrain creates a new Drain
func NewDrain() *Drain {
	return &Drain{
		output: make(chan interface{}, 1024),
	}
}

// SetInput sets input channel for executor node
func (s *Drain) SetInput(in <-chan interface{}) {
	s.input = in
}

// GetOutput returns output channel for stream node
func (s *Drain) GetOutput() <-chan interface{} {
	return s.output
}

// Open opens the sink node to start consuming streaming data
func (s *Drain) Open(ctx context.Context) <-chan error {
	s.logFn = autoctx.GetLogFunc(ctx)
	util.Logfn(s.logFn, "Opening drain")
	result := make(chan error)
	go func() {
		defer func() {
			util.Logfn(s.logFn, "Closing drain")
			close(s.output)
			close(result)
		}()
		for data := range s.input {
			s.output <- data
		}
	}()
	return result
}
