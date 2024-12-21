package sources

import (
	"bufio"
	"context"
	"io"
	"log/slog"

	"github.com/vladimirvivien/automi/api"
	"github.com/vladimirvivien/automi/log"
)

// ScannerSource sources an io.Reader and uses bufio.Scanner
// to tokenize and emit the data as []byte.
type ScannerSource[OUT []byte] struct {
	rdrParam   io.Reader
	spltrParam bufio.SplitFunc
	scanner    *bufio.Scanner
	output     chan any
	logf       api.StreamLogFunc
}

// Scanner returns a *ScannerEmitter that takes an io.Reader as its source.
// It uses splitter to tokenize the data and emits its as a chuck of bytes.
func Scanner[OUT []byte](reader io.Reader, splitter bufio.SplitFunc) *ScannerSource[OUT] {
	return &ScannerSource[OUT]{
		rdrParam:   reader,
		spltrParam: splitter,
		output:     make(chan any, 1024),
		logf:       log.NoLogFunc,
	}
}

// GetOutput returns the output channel of this source node
func (e *ScannerSource[OUT]) GetOutput() <-chan any {
	return e.output
}

// SetLogFunc sets a logging function for the component
func (e *ScannerSource[OUT]) SetLogFunc(f api.StreamLogFunc) {
	e.logf = f
}

// Open opens the emitter to start emitting data
func (e *ScannerSource[OUT]) Open(ctx context.Context) error {
	if err := e.setupScanner(); err != nil {
		return err
	}

	e.logf(ctx, log.LogInfo(
		"Component starting",
		slog.String("source", "io.Scanner"),
	))

	// use scanner to tokenize reader stream
	// the text value of token is sent downstream
	go func() {
		exeCtx, cancel := context.WithCancel(ctx)
		defer func() {
			e.logf(ctx, log.LogInfo(
				"Component closing",
				slog.String("source", "io.Scanner"),
			))
			cancel()
			close(e.output)
		}()

		for e.scanner.Scan() {
			if err := e.scanner.Err(); err != nil {
				e.logf(ctx, log.LogDebug(
					"Error: reading source",
					slog.String("source", "io.Scanner"),
					slog.String("error", err.Error()),
				))
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

func (e *ScannerSource[OUT]) setupScanner() error {
	if e.rdrParam == nil {
		return api.ErrSourceInputUndefined
	}

	e.scanner = bufio.NewScanner(e.rdrParam)
	e.scanner.Split(bufio.ScanLines)

	if e.spltrParam != nil {
		e.scanner.Split(e.spltrParam)
	}
	return nil
}
