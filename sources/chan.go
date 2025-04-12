package sources

import (
	"context"
	"log/slog"

	"github.com/vladimirvivien/automi/api"
	"github.com/vladimirvivien/automi/log"
)

// ChanSource sources its data from a channel C
// and emits each data item T it receives.
type ChanSource[T any] struct {
	channel <-chan T
	output  chan any
	logf    api.StreamLogFunc
}

// Chan constructs a new source from channel C
func Chan[T any](channel <-chan T) *ChanSource[T] {
	return &ChanSource[T]{
		channel: channel,
		output:  make(chan any, 1024),
		logf:    log.NoLogFunc,
	}
}

// GetOutput returns the output channel of this source node
func (c *ChanSource[T]) GetOutput() <-chan any {
	return c.output
}

func (c *ChanSource[T]) SetLogFunc(f api.StreamLogFunc) {
	c.logf = f
}

// Open opens the source node to start streaming data from its source
func (c *ChanSource[T]) Open(ctx context.Context) error {
	if c.channel == nil {
		c.logf(ctx, log.LogError(
			"Source input missing",
			slog.String("source", "Chan"),
		))
		return api.ErrSourceInputUndefined
	}

	c.logf(ctx, log.LogInfo(
		"Component starting",
		slog.String("source", "Chan"),
	))

	go func() {
		exeCtx, cancel := context.WithCancel(ctx)
		defer func() {
			c.logf(ctx, log.LogInfo(
				"Component closing",
				slog.String("source", "Chan"),
			))
			cancel()
			close(c.output)
		}()

		for {
			val, open := <-c.channel
			if !open {
				return
			}
			select {
			case c.output <- val:
			case <-exeCtx.Done():
				return
			}
		}
	}()
	return nil
}
