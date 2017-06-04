package collectors

import (
	"context"
	"errors"
	"log"

	autoctx "github.com/vladimirvivien/automi/api/context"
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
	log   *log.Logger
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
	c.log = autoctx.GetLogger(ctx)
	c.log.Print("opening func collector")
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
			c.log.Println("closing func collector")
			close(result)
		}()

		for val := range c.input {
			if err := c.f(val); err != nil {
				// TODO proper error handling
				c.log.Print(err)
			}
		}
	}()

	return result
}
