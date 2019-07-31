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
// and wraps it into a bufio.Scanner.  The scanner tokenizes
// the source data using the splitter func of type bufio.SplitFunc
// and emits each token as []byte.
type ScannerEmitter struct {
	rdrParam   io.Reader
	spltrParam bufio.SplitFunc
	scanner    *bufio.Scanner
	output     chan interface{}
	logf       api.LogFunc
	errf       api.ErrorFunc
}

// Scanner returns a *ScannerEmitter that wraps io.Reader into
// a bufio.Scanner.  The SplitFunc is used to tokenize the IO stream.
// The text value of the token is sent downstream.
// bufio.ScanLines will be used by default if none is provided.
func Scanner(reader io.Reader, splitter bufio.SplitFunc) *ScannerEmitter {
	return &ScannerEmitter{
		rdrParam:   reader,
		spltrParam: splitter,
		output:     make(chan interface{}, 1024),
	}
}

// GetOutput returns the output channel of this source node
func (e *ScannerEmitter) GetOutput() <-chan interface{} {
	return e.output
}

// Open opens the emitter to start emitting data
func (e *ScannerEmitter) Open(ctx context.Context) error {
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
				util.Logfn(e.logf, fmt.Errorf("Scanner emitter error: %s", err))
				autoctx.Err(e.errf, api.Error(err.Error()))
			}
			select {
			case e.output <- e.scanner.Text():
			case <-exeCtx.Done():
				return
			}
		}
	}()
	return nil
}

func (e *ScannerEmitter) setupScanner() error {
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
