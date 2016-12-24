package stream

import (
	"context"
	"log"
	"os"
)

// Drain is a generic sink that terminates streamed data
type Drain struct {
	output chan interface{}
	input  <-chan interface{}
	log    *log.Logger
}

// NewDrain creates a new Drain
func NewDrain() *Drain {
	return &Drain{
		output: make(chan interface{}, 1024),
		log:    log.New(os.Stderr, "", log.Flags()),
	}
}

// SetInput sets input channel for executor node
func (s *Drain) SetInput(in <-chan interface{}) {
	s.input = in
}

// GetOuput returns output channel for stream node
func (s *Drain) GetOutput() <-chan interface{} {
	return s.output
}

// Open opens the sink node to start consuming streaming data
func (s *Drain) Open(ctx context.Context) <-chan error {
	s.log.Print("cpening component")
	result := make(chan error)
	go func() {
		defer func() {
			s.log.Print("closing component")
			close(s.output)
			close(result)
		}()
		for data := range s.input {
			s.output <- data
		}
	}()
	return result
}
