package sinks

import (
	"context"
	"fmt"
	"io"
	"log/slog"

	"github.com/vladimirvivien/automi/api"
	"github.com/vladimirvivien/automi/log"
)

// WriterSink collects streamed items and uses an io.Writer to capture them
type WriterSink[T []byte | string] struct {
	writer io.Writer
	input  <-chan any
	logf   api.StreamLogFunc
}

// Writer creates a new WriteCollector
func Writer[T []byte | string](writer io.Writer) *WriterSink[T] {
	return &WriterSink[T]{
		writer: writer,
		logf:   log.NoLogFunc,
	}
}

// SetInput sets the source for the collector
func (c *WriterSink[T]) SetInput(in <-chan any) {
	c.input = in
}

// SetLogFunc sets logging function for component
func (c *WriterSink[T]) SetLogFunc(f api.StreamLogFunc) {
	c.logf = f
}

// Open starts the collector and returns a channel to wait for
// collection to complete or returns an error if one occurred.
func (c *WriterSink[T]) Open(ctx context.Context) <-chan error {
	c.logf(ctx, log.LogInfo(
		"Component starting",
		slog.String("sink", "Writer"),
	))
	result := make(chan error)

	go func() {
		defer func() {
			c.logf(ctx, log.LogInfo(
				"Component closing",
				slog.String("sink", "Writer"),
			))
			close(result)
		}()

		for {
			select {
			case val, opened := <-c.input:
				if !opened {
					return
				}
				switch data := any(val).(type) {
				case string:
					_, err := fmt.Fprint(c.writer, data)
					if err != nil {
						c.logf(ctx, log.LogDebug(
							"Error: writing string",
							slog.String("sink", "Writer"),
							slog.String("error", err.Error()),
						))
						continue
					}
				case []byte:
					if _, err := c.writer.Write(data); err != nil {
						c.logf(ctx, log.LogDebug(
							"Error: writing bytes",
							slog.String("sink", "Writer"),
							slog.String("error", err.Error()),
						))
						continue
					}
				default:
					// other types are serialized using string representation
					// extracted by fmt
					_, err := fmt.Fprintf(c.writer, "%v", data)
					if err != nil {
						c.logf(ctx, log.LogDebug(
							"Error: writing data",
							slog.String("sink", "Writer"),
							slog.String("error", err.Error()),
						))
						continue
					}
				}
			case <-ctx.Done():
				return
			}
		}
	}()

	return result
}
