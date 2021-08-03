package collectors

import (
	"context"

	"github.com/vladimirvivien/automi/api"
	autoctx "github.com/vladimirvivien/automi/api/context"
	"github.com/vladimirvivien/automi/util"
)

// SliceCollector is a collector that collects streamed items
// into a slice
type SliceCollector struct {
	slice []interface{}
	input <-chan interface{}
	logf  api.LogFunc
}

// Slice is the constructor function which returns a new
// SliceCollector
func Slice() *SliceCollector {
	return new(SliceCollector)
}

// SetInput sets the source for the collector
func (s *SliceCollector) SetInput(in <-chan interface{}) {
	s.input = in
}

// Get returns the slice value used to store collected items
func (s *SliceCollector) Get() []interface{} {
	return s.slice
}

// Open starts the collector and returns and waits on the returned
// channel for the collector to be done or an error to be received.
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
