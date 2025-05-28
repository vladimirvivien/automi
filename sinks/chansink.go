package sinks

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/vladimirvivien/automi/api"
	"github.com/vladimirvivien/automi/log"
)

// ChanSink sends streamed items to an output channel.
type ChanSink[T any] struct {
	output chan T
	input  <-chan any
	logf   api.StreamLogFunc
}

// Chan is the constructor function which returns a new ChanSink.
func Chan[T any](outputChan chan T) *ChanSink[T] {
	return &ChanSink[T]{
		output: outputChan,
		logf:   log.NoLogFunc,
	}
}

// SetInput sets the source for the sink.
func (s *ChanSink[T]) SetInput(in <-chan any) {
	s.input = in
}

// Get returns the output channel used by the sink.
func (s *ChanSink[T]) Get() <-chan T {
	return s.output
}

// SetLogFunc sets a logging func for the component.
func (s *ChanSink[T]) SetLogFunc(f api.StreamLogFunc) {
	s.logf = f
}

// Open starts the sink and returns and waits on the returned
// channel for the sink to be done or an error to be received.
func (s *ChanSink[T]) Open(ctx context.Context) <-chan error {
	result := make(chan error)

	s.logf(ctx, log.LogInfo(
		"Component starting",
		slog.String("sink", "Chan"),
	))

	go func() {
		defer func() {
			close(result)
			s.logf(ctx, log.LogInfo(
				"Component closing",
				slog.String("sink", "Chan"),
			))
			close(s.output) // Ensure output channel is closed
		}()

		for {
			select {
			case item, opened := <-s.input:
				if !opened {
					return
				}
				data, ok := item.(T)
				if !ok {
					s.logf(ctx, log.LogDebug(
						"Error: unexpected data type",
						slog.String("sink", "Chan"),
						slog.String("type", fmt.Sprintf("%T", item)),
					))
					continue
				}
				s.output <- data
			case <-ctx.Done():
				return
			}
		}
	}()

	return result
}
