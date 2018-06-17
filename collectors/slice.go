package collectors

import (
	"context"

	"github.com/go-faces/logger"
	autoctx "github.com/vladimirvivien/automi/api/context"
	"github.com/vladimirvivien/automi/util"
)

type SliceCollector struct {
	slice []interface{}
	input <-chan interface{}
	log   logger.Interface
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
	util.Log(s.log, "opening slice collector")
	result := make(chan error)

	go func() {
		defer func() {
			close(result)
			util.Log(s.log, "closing slice collector")
		}()
		for val := range s.input {
			s.slice = append(s.slice, val)
		}
	}()

	return result
}
