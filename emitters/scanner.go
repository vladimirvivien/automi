package emitters

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"

	"github.com/vladimirvivien/automi/api"
	autoctx "github.com/vladimirvivien/automi/api/context"
	"github.com/vladimirvivien/automi/util"
)

// ScannerEmitter takes an io.Reader as its source and emits
// data that is tokenized using a bufio.Scanner.
type ScannerEmitter[T []byte] struct {
	rdrParam   io.Reader
	spltrParam bufio.SplitFunc
	scanner    *bufio.Scanner
	output     chan T
	logf       api.LogFunc
	errf       api.ErrorFunc
}

// Scanner returns a *ScannerEmitter that takes an io.Reader as its source.
// It uses splitter to tokenize the data and emits its as a chuck of bytes.
func Scanner[T []byte](reader io.Reader, splitter bufio.SplitFunc) *ScannerEmitter[T] {
	return &ScannerEmitter[T]{
		rdrParam:   reader,
		spltrParam: splitter,
		output:     make(chan T, 1024),
	}
}

// GetOutput returns the output channel of this source node
func (e *ScannerEmitter[T]) GetOutput() <-chan T {
	return e.output
}

// Open opens the emitter to start emitting data
func (e *ScannerEmitter[T]) Open(ctx context.Context) error {
	if err := e.setupScanner(); err != nil {
		return err
	}
	e.logf = autoctx.GetLogFunc(ctx)
	e.errf = autoctx.GetErrFunc(ctx)

	util.Logfn(e.logf, "Scanner emitter starting")

	// use scanner to tokenize reader stream
	// the text value of token is sent downstream
	go func() {
		exeCtx, cancel := context.WithCancel(ctx)
		defer func() {
			util.Logfn(e.logf, "Scanner emitter closing")
			cancel()
			close(e.output)
		}()

		for e.scanner.Scan() {
			if err := e.scanner.Err(); err != nil {
				util.Logfn(e.logf, fmt.Errorf("scanner emitter error: %w", err))
				autoctx.Err(e.errf, api.Error(err.Error()))
			}
			select {
			case e.output <- e.scanner.Bytes():
			case <-exeCtx.Done():
				return
			}
		}
	}()
	return nil
}

func (e *ScannerEmitter[T]) setupScanner() error {
	if e.rdrParam == nil {
		return errors.New("emitter missing io.Reader source")
	}

	e.scanner = bufio.NewScanner(e.rdrParam)
	e.scanner.Split(bufio.ScanLines)

	if e.spltrParam != nil {
		e.scanner.Split(e.spltrParam)
	}
	return nil
}
