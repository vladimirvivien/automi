package sinks

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/vladimirvivien/automi/api"
	"github.com/vladimirvivien/automi/log"
)

// ChanSink represents a sink backed by a Go channel.
type ChanSink[IN any, CHAN <-chan IN] struct {
	chansink chan IN
	input    <-chan any
	logf     api.StreamLogFunc
}

// BufferedChan returns a new ChanSink backed by a buffered channel.
func BufferedChan[IN any, CHAN <-chan IN](bufferSize int) *ChanSink[IN, CHAN] {
	return &ChanSink[IN, CHAN]{
		chansink: make(chan IN, bufferSize),
		logf:     log.NoLogFunc,
	}
}

// Chan returns a new ChanSink backed by an unbuffered channel.
func Chan[IN any, CHAN <-chan IN]() *ChanSink[IN, CHAN] {
	return BufferedChan[IN, CHAN](0)
}

// SetInput sets the input for the sink.
func (s *ChanSink[IN, CHAN]) SetInput(in <-chan any) {
	s.input = in
}

// Get returns a receive-only channel used by the sink.
func (s *ChanSink[IN, CHAN]) Get() CHAN {
	return s.chansink
}

// SetLogFunc sets a logging func for the component.
func (s *ChanSink[IN, CHAN]) SetLogFunc(f api.StreamLogFunc) {
	s.logf = f
}

// Open starts the sink and returns and waits on the returned
func (s *ChanSink[IN, CHAN]) Open(ctx context.Context) <-chan error {
	result := make(chan error)

	if s.input == nil {
		go func() { result <- api.ErrInputChannelUndefined }()
		return result
	}

	if s.chansink == nil {
		go func() { result <- api.ErrSinkDestinationUndefined }()
		return result
	}

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
			close(s.chansink) // Ensure output channel is closed
		}()

		if s.input == nil {
			panic("ChanSink input not set")
		}
		if s.chansink == nil {
			panic("ChanSink channel not set")
		}

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
						slog.String("sink", "Chan"),
						slog.String("type", fmt.Sprintf("%T", item)),
					))
					continue
				}
				s.chansink <- data
			case <-ctx.Done():
				return
			}
		}
	}()

	return result
}
