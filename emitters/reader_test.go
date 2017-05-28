package emitters

import (
	"bufio"
	"context"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestEmitter_Reader_Words(t *testing.T) {
	data := "1234 Bad Lad sda cat mad back tactful"

	r := Reader(strings.NewReader(data), bufio.ScanWords)
	var m sync.Mutex
	count := 0
	wait := make(chan struct{})
	go func() {
		defer close(wait)
		for _ = range r.GetOutput() {
			m.Lock()
			count++
			m.Unlock()
		}
	}()

	if err := r.Open(context.TODO()); err != nil {
		t.Fatal(err)
	}

	select {
	case <-wait:
	case <-time.After(500 * time.Microsecond):
		t.Fatal("waited too long")
	}
	m.Lock()
	if count != 8 {
		t.Fatal("unexpected item count ", count)
	}
	m.Unlock()

}

func TestEmitter_Reader_Default(t *testing.T) {
	data := "Hello\nWe come\nin peace!"

	r := Reader(strings.NewReader(data), nil)
	var m sync.Mutex
	count := 0
	wait := make(chan struct{})
	go func() {
		defer close(wait)
		for _ = range r.GetOutput() {
			m.Lock()
			count++
			m.Unlock()
		}
	}()

	if err := r.Open(context.TODO()); err != nil {
		t.Fatal(err)
	}

	select {
	case <-wait:
	case <-time.After(500 * time.Microsecond):
		t.Fatal("waited too long")
	}
	m.Lock()
	if count != 3 {
		t.Fatal("unexpected item count ", count)
	}
	m.Unlock()

}
