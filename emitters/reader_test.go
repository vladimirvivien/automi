package emitters

import (
	"bufio"
	"bytes"
	"context"
	"io"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestEmitter_Reader(t *testing.T) {
	tests := []struct {
		test     string
		r        io.Reader
		expected string
	}{
		{test: "bytes.Buffer", r: bytes.NewBufferString("Hello World"), expected: "Hello World"},
		{test: "bytes.Reader", r: bytes.NewReader([]byte("Hello World")), expected: "Hello World"},
		{test: "strings.Reader", r: strings.NewReader("Hello Universe"), expected: "Hello Universe"},
		{test: "bufio.Reader", r: bufio.NewReader(strings.NewReader("Hello Multiverse")), expected: "Hello Multiverse"},
	}

	var m sync.Mutex

	for _, test := range tests {
		t.Logf("running test %s", test.test)
		e := Reader(test.r)
		var read bytes.Buffer
		wait := make(chan struct{})
		go func() {
			defer close(wait)
			for item := range e.GetOutput() {
				bytes := item.([]byte)
				m.Lock()
				read.Write(bytes)
				m.Unlock()
			}
		}()

		if err := e.Open(context.Background()); err != nil {
			t.Fatal(err)
		}

		select {
		case <-wait:
		case <-time.After(500 * time.Microsecond):
			t.Fatal("waited too long")
		}

		m.Lock()
		if read.String() != test.expected {
			t.Fatalf("expecting [%s] got [%s] ", test.expected, read.String())
		}
		m.Unlock()
	}
}
