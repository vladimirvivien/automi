package emitters

import (
	"bufio"
	"context"
	"errors"
	"io"
	"log"

	autoctx "github.com/vladimirvivien/automi/api/context"
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
	log        *log.Logger
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
	e.log = autoctx.GetLogger(ctx)
	e.log.Print("opening io.Reader emitter")

	// use scanner to tokenize reader stream
	// the text value of token is sent downstream
	go func() {
		defer close(e.output)

		for e.scanner.Scan() {
			//TODO: handle scanner errors

			select {
			case e.output <- e.scanner.Text():
			case <-ctx.Done():
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
