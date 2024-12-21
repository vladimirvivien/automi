package sources

import (
	"context"
	"io"
	"log/slog"

	"github.com/vladimirvivien/automi/api"
	"github.com/vladimirvivien/automi/log"
)

// ReaderSource sources data from an io.Reader and emits chunk of N bytes
type ReaderSource[OUT []byte] struct {
	reader io.Reader
	size   int
	output chan any
	logf   api.StreamLogFunc
}

// Reader creates a *ReaderEmitter source that can emit []bytes
func Reader[OUT []byte](reader io.Reader) *ReaderSource[OUT] {
	return &ReaderSource[OUT]{
		reader: reader,
		size:   1024,
		output: make(chan any, 1024),
		logf:   log.NoLogFunc,
	}
}

// BufferSize sets the slice chunck size used to transfer data
// read from source io.Reader.
func (e *ReaderSource[OUT]) BufferSize(s int) *ReaderSource[OUT] {
	e.size = s
	return e
}

// GetOutput returns the output channel of this source node
func (e *ReaderSource[OUT]) GetOutput() <-chan any {
	return e.output
}

// SetLogFunc sets a log function for the component
func (e *ReaderSource[OUT]) SetLogFunc(f api.StreamLogFunc) {
	e.logf = f
}

// Open opens the emitter to start emitting data
func (e *ReaderSource[OUT]) Open(ctx context.Context) error {
	if err := e.setupReader(); err != nil {
		return err
	}

	e.logf(ctx, log.LogInfo(
		"Component starting",
		slog.String("source", "io.Reader"),
	))

	go func() {
		exeCtx, cancel := context.WithCancel(ctx)
		defer func() {
			e.logf(ctx, log.LogInfo(
				"Component closing",
				slog.String("source", "io.Reader"),
			))

			cancel()
			close(e.output)
		}()

		for {
			buf := make([]byte, e.size)
			bytesRead, err := e.reader.Read(buf)

			if bytesRead > 0 {
				select {
				case e.output <- buf[0:bytesRead]:
				case <-exeCtx.Done():
					return
				}
			}
			if err != nil {
				// Any error closes channel
				e.logf(ctx, log.LogDebug(
					"Error: reading source",
					slog.String("source", "io.Reader"),
					slog.String("error", err.Error()),
				))
				return
			}
		}
	}()
	return nil
}

func (e *ReaderSource[OUT]) setupReader() error {
	if e.reader == nil {
		return api.ErrSourceInputUndefined
	}

	if e.size <= 0 {
		e.size = 10 * 1024 // default 10k buffer
	}
	return nil
}
