package emitters

import (
	"context"
	"errors"
	"fmt"
	"io"

	"github.com/vladimirvivien/automi/api"
	autoctx "github.com/vladimirvivien/automi/api/context"
	"github.com/vladimirvivien/automi/util"
)

// ReaderEmitter takes an io.Reader source and emits chunk of bytes of
// N length, with each iteration.
type ReaderEmitter[T []byte] struct {
	reader io.Reader
	size   int
	output chan T
	logf   api.LogFunc
	errf   api.ErrorFunc
}

// Reader creates a *ReaderEmitter source that can emit []bytes
func Reader[T []byte](reader io.Reader) *ReaderEmitter[T] {
	return &ReaderEmitter[T]{
		reader: reader,
		size:   1024,
		output: make(chan T, 1024),
	}
}

// BufferSize sets the slice chunck size used to transfer data
// read from source io.Reader.
func (e *ReaderEmitter[T]) BufferSize(s int) *ReaderEmitter[T] {
	e.size = s
	return e
}

// GetOutput returns the output channel of this source node
func (e *ReaderEmitter[T]) GetOutput() <-chan T {
	return e.output
}

// Open opens the emitter to start emitting data
func (e *ReaderEmitter[T]) Open(ctx context.Context) error {
	if err := e.setupReader(); err != nil {
		return err
	}

	e.logf = autoctx.GetLogFunc(ctx)
	e.errf = autoctx.GetErrFunc(ctx)

	util.Logfn(e.logf, "Opening io.Reader emitter")

	go func() {
		exeCtx, cancel := context.WithCancel(ctx)
		defer func() {
			util.Logfn(e.logf, "Closing io.Reader emitter")
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
				util.Logfn(e.logf, fmt.Errorf("error reading: %w", err))
				autoctx.Err(e.errf, api.Error(err.Error()))
				return
			}
		}
	}()
	return nil
}

func (e *ReaderEmitter[T]) setupReader() error {
	if e.reader == nil {
		return errors.New("emitter missing io.Reader source")
	}
	if e.size <= 0 {
		e.size = 10 * 1024 // default 10k buffer
	}
	return nil
}
