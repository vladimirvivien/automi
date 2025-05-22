package sinks

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/vladimirvivien/automi/api"
	"github.com/vladimirvivien/automi/log"
)

// ChanSink is a sink that sends incoming items of type T to a Go channel.
type ChanSink[T any] struct {
	outputChan chan<- T
	inputChan  <-chan any
	logf       api.StreamLogFunc
}

// Channel creates a new *ChanSink that will send items to the provided channel.
// The provided channel `ch` will be closed by this Sink when the input stream is exhausted or the context is cancelled.
func Channel[T any](ch chan<- T) *ChanSink[T] {
	return &ChanSink[T]{
		outputChan: ch,
		logf:       log.NoLogFunc,
	}
}

// SetInput sets the input channel for the sink.
func (s *ChanSink[T]) SetInput(in <-chan any) {
	s.inputChan = in
}

// SetLogFunc sets the log function for the sink.
func (s *ChanSink[T]) SetLogFunc(f api.StreamLogFunc) {
	s.logf = f
}

// Open starts the sink processing. It reads items from the input channel
// and sends them to the configured output channel.
// It returns a channel that will receive an error if one occurs during processing,
// or nil if processing completes successfully. The error channel is closed afterwards.
// The output channel provided during construction is closed by this sink when
// the input channel is exhausted or the context is cancelled.
func (s *ChanSink[T]) Open(ctx context.Context) <-chan error {
	errChan := make(chan error, 1) // Buffered to prevent blocking if sending error

	s.logf(ctx, log.LogInfo(
		"Component starting",
		slog.String("sink", "ChanSink"),
	))

	if s.inputChan == nil {
		s.logf(ctx, log.LogError(
			"Input channel missing",
			slog.String("sink", "ChanSink"),
		))
		errChan <- api.ErrInputChannelUndefined
		close(errChan)
		// Also close the outputChan if it's not nil, as per documented behavior.
		if s.outputChan != nil {
			close(s.outputChan)
		}
		return errChan
	}

	if s.outputChan == nil {
		s.logf(ctx, log.LogError(
			"Output channel missing",
			slog.String("sink", "ChanSink"),
		))
		errChan <- api.ErrSinkDestinationUndefined
		close(errChan)
		return errChan
	}

	go func() {
		defer func() {
			s.logf(ctx, log.LogInfo(
				"Component closing",
				slog.String("sink", "ChanSink"),
			))
			close(s.outputChan) // Close the output channel when done.
			close(errChan)      // Close the error channel.
		}()

		for {
			select {
			case item, open := <-s.inputChan:
				if !open {
					s.logf(ctx, log.LogInfo(
						"Input channel closed",
						slog.String("sink", "ChanSink"),
					))
					return // Input channel closed, normal completion.
				}

				typedItem, ok := item.(T)
				if !ok {
					errMsg := fmt.Sprintf("unexpected data type: expected %T, got %T", *new(T), item)
					s.logf(ctx, log.LogDebug( // Using Debug as it's a per-item error
						errMsg,
						slog.String("sink", "ChanSink"),
					))
					// Decide whether to send an error or continue.
					// For now, let's log and continue, as per other sinks.
					// If this should be a fatal error for the sink, send to errChan and return.
					continue
				}

				// Send item to output channel, respecting context cancellation
				select {
				case s.outputChan <- typedItem:
					// Item sent successfully
				case <-ctx.Done():
					s.logf(ctx, log.LogInfo(
						"Context cancelled, closing sink",
						slog.String("sink", "ChanSink"),
						slog.String("error", ctx.Err().Error()),
					))
					return // Context cancelled
				}

			case <-ctx.Done():
				s.logf(ctx, log.LogInfo(
					"Context cancelled, closing sink",
					slog.String("sink", "ChanSink"),
					slog.String("error", ctx.Err().Error()),
				))
				return // Context cancelled
			}
		}
	}()

	return errChan
}
