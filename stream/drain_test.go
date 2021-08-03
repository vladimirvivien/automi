package stream

import (
	"context"
	"sync"
	"testing"
	"time"
)

func TestDrain_Open(t *testing.T) {
	in := make(chan interface{})
	go func() {
		in <- []string{"A", "B", "C"}
		in <- []string{"D", "E"}
		in <- []string{"G"}
		close(in)
	}()
	d := NewDrain()
	d.SetInput(in)

	var m sync.RWMutex
	count := 0
	wait := make(chan struct{})
	go func() {
		defer close(wait)
		for range d.GetOutput() {
			m.Lock()
			count++
			m.Unlock()
		}
	}()

	select {
	case err := <-d.Open(context.Background()):
		if err != nil {
			t.Fatal(err)
		}
		select {
		case <-wait:
		case <-time.After(50 * time.Millisecond):
			t.Fatal("Drain output processing took too long")
		}
	case <-time.After(50 * time.Millisecond):
		t.Fatal("Drain took too long to open")
	}

	m.RLock()
	if count != 3 {
		t.Fatal("Expected 3 elements, got ", count)
	}
	m.RUnlock()
}
