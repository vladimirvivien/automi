package sinks

import (
	"context"
	"log/slog"

	"github.com/vladimirvivien/automi/api"
	"github.com/vladimirvivien/automi/log"
)

// DiscardSink is noop sink
type DiscardSink struct {
	input <-chan any
	logf  api.StreamLogFunc
}

// Discard creates a new *DiscardSink with type T
func Discard() *DiscardSink {
	return new(DiscardSink)
}

// SetInput sets the input source for the collector
func (s *DiscardSink) SetInput(in <-chan any) {
	s.input = in
}

// SetLogFunc sets the logging function for the component
func (s *DiscardSink) SetLogFunc(f api.StreamLogFunc) {
	s.logf = f
}

// Open opens the node to start collecting
func (s *DiscardSink) Open(ctx context.Context) <-chan error {
	if s.logf == nil {
		s.logf = log.NoLogFunc
	}
	s.logf(ctx, log.LogInfo(
		"Component starting",
		slog.String("sink", "Discard"),
	))

	result := make(chan error)
	go func() {
		defer func() {
			s.logf(ctx, log.LogInfo(
				"Component closing",
				slog.String("sink", "Discard"),
			))
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
