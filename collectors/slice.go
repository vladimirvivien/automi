package collectors

import (
	"context"

	"github.com/vladimirvivien/automi/api"
	autoctx "github.com/vladimirvivien/automi/api/context"
	"github.com/vladimirvivien/automi/util"
)

type SliceCollector struct {
	slice []interface{}
	input <-chan interface{}
	logf  api.LogFunc
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
	s.logf = autoctx.GetLogFunc(ctx)
	util.Logfn(s.logf, "Opening slice collector")
	result := make(chan error)

	go func() {
		defer func() {
			close(result)
			util.Logfn(s.logf, "Closing slice collector")
		}()

		for {
			select {
			case item, opened := <-s.input:
				if !opened {
					return
				}
				s.slice = append(s.slice, item)
			case <-ctx.Done():
				return
			}
		}
	}()

	return result
}
