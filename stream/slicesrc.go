package stream

import (
	"context"
	"log"

	autoctx "github.com/vladimirvivien/automi/api/context"
)

type SliceSrc struct {
	src    []interface{}
	output chan interface{}
	log    *log.Logger
}

func NewSliceSource(elems ...interface{}) *SliceSrc {
	return &SliceSrc{
		src:    elems,
		output: make(chan interface{}, 1024),
	}
}
func (s *SliceSrc) GetOutput() <-chan interface{} {
	return s.output
}
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
