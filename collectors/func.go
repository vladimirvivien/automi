package collectors

import (
	"context"
	"errors"

	"github.com/vladimirvivien/automi/api"
	autoctx "github.com/vladimirvivien/automi/api/context"
	"github.com/vladimirvivien/automi/util"
)

// CollectorFunc is a function used to colllect
// incoming stream data. It can be used as a
// stream sink.
type CollectorFunc func(interface{}) error

// FuncCollector is a colletor that uses a function
// to collect data.  The specified function must be
// of type:
//   CollectorFunc
type FuncCollector struct {
	input <-chan interface{}
	logf  api.LogFunc
	f     CollectorFunc
}

// Func creates a new value *FuncCollector that
// will use the specified function parameter to
// collect streaming data.
func Func(f CollectorFunc) *FuncCollector {
	return &FuncCollector{f: f}
}

// SetInput sets the channel input
func (c *FuncCollector) SetInput(in <-chan interface{}) {
	c.input = in
}

// Open is the starting point that starts the collector
func (c *FuncCollector) Open(ctx context.Context) <-chan error {
	c.logf = autoctx.GetLogFunc(ctx)
	util.Logfn(c.logf, "Opening func collector")
	result := make(chan error)

	if c.input == nil {
		go func() { result <- errors.New("func collector missing input") }()
		return result
	}

	if c.f == nil {
		go func() { result <- errors.New("func collector missing function") }()
		return result
	}

	go func() {
		defer func() {
			util.Logfn(c.logf, "Closing func collector")
			close(result)
		}()

		for {
			select {
			case item, opened := <-c.input:
				if !opened {
					return
				}
				if err := c.f(item); err != nil {
					// TODO proper error handling (with StreamError)
					util.Logfn(c.logf, err)
				}
			case <-ctx.Done():
				return
			}
		}
	}()

	return result
}
