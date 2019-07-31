package emitters

import (
	"context"
	"sync"
	"testing"
	"time"
)

func TestEmitter_Chan(t *testing.T) {
	tests := []struct {
		name string
		data []interface{}
	}{
		{"string data", []interface{}{"A", "B", "C"}},
		{"numbers", []interface{}{400, 1200, 12.5, 66.0, 900, 500, 55, 1235543245}},
		{"composite types", []interface{}{[]string{"A", "B"}, []int{6, 4, 3}}},
		{"nothing", nil},
	}

	for _, test := range tests {
		t.Logf("running test %s", test.name)
		ch := make(chan interface{})
		go func(data []interface{}) {
			for _, val := range data {
				ch <- val
			}
			close(ch)
		}(test.data)

		e := Chan(ch)

		var m sync.Mutex
		count := 0
		wait := make(chan struct{})
		go func() {
			defer close(wait)
			for range e.GetOutput() {
				m.Lock()
				count++
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
		if count != len(test.data) {
			t.Fatal("unexpected item count ", count)
		}
		m.Unlock()
	}

}
