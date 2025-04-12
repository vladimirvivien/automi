package sinks

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/vladimirvivien/automi/api"
	"github.com/vladimirvivien/automi/log"
)

// FuncSink uses a function to collect streamed items of type T.
type FuncSink[T any] struct {
	input <-chan any
	logf  api.StreamLogFunc
	f     func(T) error
}

// Func creates a new *FuncSink using the specified function
// as its parameter.
func Func[T any](f func(T) error) *FuncSink[T] {
	return &FuncSink[T]{
		f:    f,
		logf: log.NoLogFunc,
	}
}

// SetInput sets the channel input
func (c *FuncSink[T]) SetInput(in <-chan any) {
	c.input = in
}

func (c *FuncSink[T]) SetLogFunc(f api.StreamLogFunc) {
	c.logf = f
}

// Open is the starting point that starts the collector
func (c *FuncSink[T]) Open(ctx context.Context) <-chan error {
	c.logf(ctx, log.LogInfo(
		"Component starting",
		slog.String("sink", "Func"),
	))

	result := make(chan error)

	if c.input == nil {
		go func() { result <- api.ErrInputChannelUndefined }()
		return result
	}

	if c.f == nil {
		c.logf(ctx, log.LogError(
			"No destination resource defined",
			slog.String("sink", "Func"),
		))
		go func() { result <- api.ErrSinkDestinationUndefined }()
		return result
	}

	go func() {
		defer func() {
			c.logf(ctx, log.LogInfo(
				"Component closing",
				slog.String("sink", "Func"),
			))
			close(result)
		}()

		for {
			select {
			case item, opened := <-c.input:
				if !opened {
					return
				}
				itemVal, ok := item.(T)
				if !ok {
					c.logf(ctx, log.LogDebug(
						"Error: unexpected data type",
						slog.String("sink", "Func"),
						slog.String("type", fmt.Sprintf("%T", item)),
					))
					continue
				}
				if err := c.f(itemVal); err != nil {
					c.logf(ctx, log.LogDebug(
						"Error: User function returned error",
						slog.String("sink", "Func"),
						slog.String("error", err.Error()),
					))
				}
			case <-ctx.Done():
				return
			}
		}
	}()

	return result
}
