package collectors

import (
	"context"

	"github.com/vladimirvivien/automi/api"
	autoctx "github.com/vladimirvivien/automi/api/context"
	"github.com/vladimirvivien/automi/util"
)

// NullCollector represents a collector that terminates
// to a noop collector
type NullCollector struct {
	input <-chan interface{}
	logf  api.LogFunc
}

// Null creates the new value of the collector
func Null() *NullCollector {
	return new(NullCollector)
}

// SetInput sets the input source for the collector
func (s *NullCollector) SetInput(in <-chan interface{}) {
	s.input = in
}

// Open opens the node to start collecting
func (s *NullCollector) Open(ctx context.Context) <-chan error {
	result := make(chan error)
	s.logf = autoctx.GetLogFunc(ctx)
	util.Logfn(s.logf, "Opening null collector")

	go func() {
		defer func() {
			util.Logfn(s.logf, "Closing null collector")
			close(result)
		}()

		for {
			select {
			case _, opened := <-s.input:
				if !opened {
					return
				}
			case <-ctx.Done():
				return
			}
		}
	}()
	return result
}
