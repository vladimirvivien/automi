package sinks

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/vladimirvivien/automi/api"
	"github.com/vladimirvivien/automi/log"
)

// SliceSink collects streamed items into a slice
type SliceSink[IN any, SLICE []IN] struct {
	slice SLICE
	input <-chan any
	logf  api.StreamLogFunc
}

// Slice is the constructor function which returns a new
// SliceCollector
func Slice[IN any, SLICE []IN]() *SliceSink[IN, SLICE] {
	return &SliceSink[IN, SLICE]{
		logf: log.NoLogFunc,
	}
}

// SetInput sets the source for the collector
func (s *SliceSink[IN, SLICE]) SetInput(in <-chan any) {
	s.input = in
}

// Get returns the slice value used to store collected items
func (s *SliceSink[IN, SLICE]) Get() SLICE {
	return s.slice
}

// SetLogFunc sets a logging func for the component
func (s *SliceSink[IN, SLICE]) SetLogFunc(f api.StreamLogFunc) {
	s.logf = f
}

// Open starts the collector and returns and waits on the returned
// channel for the collector to be done or an error to be received.
func (s *SliceSink[IN, SLICE]) Open(ctx context.Context) <-chan error {
	result := make(chan error)

	s.logf(ctx, log.LogInfo(
		"Component starting",
		slog.String("sink", "Slice"),
	))

	go func() {
		defer func() {
			close(result)
			s.logf(ctx, log.LogInfo(
				"Component closing",
				slog.String("sink", "Slice"),
			))
		}()

		for {
			select {
			case item, opened := <-s.input:
				if !opened {
					return
				}
				data, ok := item.(IN)
				if !ok {
					s.logf(ctx, log.LogDebug(
						"Error: unexpected data type",
						slog.String("sink", "Slice"),
						slog.String("type", fmt.Sprintf("%T", item)),
					))
					continue
				}
				s.slice = append(s.slice, data)
			case <-ctx.Done():
				return
			}
		}
	}()

	return result
}
