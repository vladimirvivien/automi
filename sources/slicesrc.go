package sources

import (
	"context"
	"log"

	autoctx "github.com/vladimirvivien/automi/api/context"
)

// SliceSrc is source node that takes in a slice and
// emits slice items individually as a stream.
type SliceSrc struct {
	src    []interface{}
	output chan interface{}
	log    *log.Logger
}

// Slice creates new slice source
func Slice(elems ...interface{}) *SliceSrc {
	return &SliceSrc{
		src:    elems,
		output: make(chan interface{}, 1024),
	}
}

// GetOutput returns the output channel of this source node
func (s *SliceSrc) GetOutput() <-chan interface{} {
	return s.output
}

// Open opens the source node to start streaming data on its channel
func (s *SliceSrc) Open(ctx context.Context) error {
	s.log = autoctx.GetLogger(ctx)
	s.log.Print("opening slice source")
	go func() {
		defer close(s.output)
		for _, str := range s.src {
			s.output <- str
		}
	}()
	return nil
}
