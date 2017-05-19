package emitters

import (
	"context"
	"sync"
	"testing"
	"time"
)

func TestEmitter_Slice(t *testing.T) {
	s := Slice([]string{"A", "B", "C", "D", "E"})
	var m sync.Mutex
	count := 0
	wait := make(chan struct{})
	go func() {
		defer close(wait)
		for _ = range s.GetOutput() {
			m.Lock()
			count++
			m.Unlock()
		}
	}()

	s.Open(context.TODO())

	select {
	case <-wait:
	case <-time.After(500 * time.Microsecond):
		t.Fatal("waited too long")
	}
	m.Lock()
	if count != 5 {
		t.Fatal("unexpected item count ", count)
	}
	m.Unlock()
}
