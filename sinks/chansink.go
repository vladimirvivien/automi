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
// The ChanSink will NOT close the provided channel `ch`.
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

// Open starts the sink processing.
// It returns a channel that signals errors or completion.
// If a panic occurs in the processing goroutine, an error detailing the panic is sent on this channel.
// This channel is always closed when the sink stops processing.
func (s *ChanSink[T]) Open(ctx context.Context) <-chan error {
	errChan := make(chan error, 1) // Buffered to allow sending panic error without blocking

	s.logf(ctx, log.LogInfo("ChanSink: Component starting", slog.String("sink", "ChanSink")))

	if s.inputChan == nil {
		s.logf(ctx, log.LogError("ChanSink: Input channel missing", slog.String("sink", "ChanSink")))
		select {
		case errChan <- api.ErrInputChannelUndefined:
		default:
		}
		close(errChan)
		// Do NOT close s.outputChan here
		return errChan
	}

	if s.outputChan == nil {
		s.logf(ctx, log.LogError("ChanSink: Output channel missing", slog.String("sink", "ChanSink")))
		select {
		case errChan <- api.ErrSinkDestinationUndefined:
		default:
		}
		close(errChan)
		return errChan
	}

	go func() {
		// This defer ensures errChan is always closed and panics are caught.
		defer func() {
			if r := recover(); r != nil {
				panicErr := fmt.Errorf("panic in ChanSink worker: %v", r)
				// Attempt non-blocking send of panicErr.
				select {
				case errChan <- panicErr:
				default: // Avoids blocking if errChan is already full or somehow closed
				}
			}
			close(errChan) // Always close errChan when goroutine exits.
		}()

		// Main processing loop
		for {
			select {
			case item, open := <-s.inputChan:
				if !open {
					s.logf(ctx, log.LogInfo(
						"ChanSink: Input channel closed, stopping",
						slog.String("sink", "ChanSink"),
					))
					return // Normal completion, defer will close errChan.
				}

				typedItem, ok := item.(T)
				if !ok {
					s.logf(ctx, log.LogDebug(
						fmt.Sprintf("ChanSink: Unexpected data type: expected %T, got %T", *new(T), item),
						slog.String("sink", "ChanSink"),
					))
					continue 
				}
				
				select {
				case s.outputChan <- typedItem: // This can panic if s.outputChan is closed by user/test.
					// Item sent successfully.
				case <-ctx.Done():
					s.logf(ctx, log.LogInfo(
						"ChanSink: Context cancelled while attempting to send to outputChan, stopping",
						slog.String("sink", "ChanSink"),
						slog.String("error", ctx.Err().Error()),
					))
					return // Context cancelled, defer will close errChan.
				}

			case <-ctx.Done():
				s.logf(ctx, log.LogInfo(
					"ChanSink: Context cancelled while waiting for input, stopping",
					slog.String("sink", "ChanSink"),
					slog.String("error", ctx.Err().Error()),
				))
				return // Context cancelled, defer will close errChan.
			}
		}
	}()

	return errChan
}
