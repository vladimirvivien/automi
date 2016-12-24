package stream

import (
	"context"
	"log"

	autoctx "github.com/vladimirvivien/automi/api/context"
)

// SliceSrc is source node that streams out each item of a provided sice
type SliceSrc struct {
	src    []interface{}
	output chan interface{}
	log    *log.Logger
}

// SliceSrc creates new slice source
func NewSliceSource(elems ...interface{}) *SliceSrc {
	return &SliceSrc{
		src:    elems,
		output: make(chan interface{}, 1024),
	}
}

// GetOuptut returns the output channel of this source node
func (s *SliceSrc) GetOutput() <-chan interface{} {
	return s.output
}

// Open opens the source node to start streaming data on its channel
func (s *SliceSrc) Open(ctx context.Context) error {
	s.log = autoctx.GetLogger(ctx)
	s.log.Print("opening source")
	go func() {
		defer close(s.output)
		for _, str := range s.src {
			s.output <- str
		}
	}()
	return nil
}
