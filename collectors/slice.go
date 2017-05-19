package collectors

import (
	"context"
	"log"

	autoctx "github.com/vladimirvivien/automi/api/context"
)

type SliceCollector struct {
	slice []interface{}
	input <-chan interface{}
	log   *log.Logger
}

func Slice() *SliceCollector {
	return new(SliceCollector)
}

func (s *SliceCollector) SetInput(in <-chan interface{}) {
	s.input = in
}

func (s *SliceCollector) Get() []interface{} {
	return s.slice
}

func (s *SliceCollector) Open(ctx context.Context) <-chan error {
	s.log = autoctx.GetLogger(ctx)
	s.log.Print("opening slice collector")
	result := make(chan error)

	go func() {
		defer func() {
			close(result)
			s.log.Print("closing slice collector")
		}()
		for val := range s.input {
			s.slice = append(s.slice, val)
		}
	}()

	return result
}
