package emitters

import (
	"bufio"
	"context"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestEmitter_Scanner(t *testing.T) {
	tests := []struct {
		test     string
		data     string
		expected []string
		splitter bufio.SplitFunc
	}{
		{
			test:     "WordSplit",
			data:     "the vast universe",
			expected: []string{"the", "vast", "universe"},
			splitter: bufio.ScanWords,
		},
		{
			test:     "LineSplit",
			data:     "hello world\nhello universe",
			expected: []string{"hello world", "hello universe"},
			splitter: bufio.ScanLines,
		},
	}

	var m sync.Mutex
	for _, test := range tests {
		e := Scanner(strings.NewReader(test.data), test.splitter)
		var result []string
		wait := make(chan struct{})

		go func() {
			defer close(wait)
			for item := range e.GetOutput() {
				m.Lock()
				result = append(result, item.(string))
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
		for i, val := range test.expected {
			if val != result[i] {
				t.Fatalf("value mismatch, got %s, expecting %s", result[i], val)
			}
		}
		m.Unlock()
	}
}
