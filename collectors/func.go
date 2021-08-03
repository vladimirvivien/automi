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

// FuncCollector is a collector that uses a function
// to collect data.  The specified function must be
// of type:
//   CollectorFunc
type FuncCollector struct {
	input <-chan interface{}
	logf  api.LogFunc
	errf  api.ErrorFunc
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
	c.errf = autoctx.GetErrFunc(ctx)

	util.Logfn(c.logf, "Opening func collector")
	result := make(chan error)

	if c.input == nil {
		go func() { result <- errors.New("Func collector missing input") }()
		return result
	}

	if c.f == nil {
		err := errors.New("Func collector missing function")
		util.Logfn(c.logf, err)
		autoctx.Err(c.errf, api.Error(err.Error()))
		go func() { result <- err }()
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
					util.Logfn(c.logf, err)
					autoctx.Err(c.errf, api.Error(err.Error()))
				}
			case <-ctx.Done():
				return
			}
		}
	}()

	return result
}
