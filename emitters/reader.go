package emitters

import (
	"bufio"
	"context"
	"errors"
	"io"
	"log"

	autoctx "github.com/vladimirvivien/automi/api/context"
)

// ReaderEmitter takes an io.Reader as its source and emits
// and wraps it into a bufio.Scanner.  The scanner tokenizes
// the source data using the splitter func of type bufio.SplitFunc
// and emits each token as []byte.
type ReaderEmitter struct {
	rdrParam   io.Reader
	spltrParam bufio.SplitFunc
	scanner    *bufio.Scanner
	output     chan interface{}
	log        *log.Logger
}

// Reader returns a *ReaderEmitter that wraps parameter reader into
// a bufio.Scanner.  It uses the specified SplitFunc to tokenize the
// IO stream.  If splitter set to nil, it users bufio.ScanLines by default.
func Reader(reader io.Reader, splitter bufio.SplitFunc) *ReaderEmitter {
	return &ReaderEmitter{
		rdrParam:   reader,
		spltrParam: splitter,
		output:     make(chan interface{}, 1024),
	}
}

// GetOuptut returns the output channel of this source node
func (e *ReaderEmitter) GetOutput() <-chan interface{} {
	return e.output
}

// Open opens the emitter to start emitting data
func (e *ReaderEmitter) Open(ctx context.Context) error {
	if err := e.setupScanner(); err != nil {
		return err
	}
	e.log = autoctx.GetLogger(ctx)
	e.log.Print("opening io.Reader emitter")

	// use scanner to tokenize reader stream
	go func() {
		defer close(e.output)
		for e.scanner.Scan() {
			bytes := make([]byte, len(e.scanner.Bytes()))
			copy(bytes, e.scanner.Bytes())

			select {
			case e.output <- bytes:
			case <-ctx.Done():
				return
			}
		}
	}()
	return nil
}

func (e *ReaderEmitter) setupScanner() error {
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
