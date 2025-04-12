package flat

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/vladimirvivien/automi/testutil"
)

func TestNewSliceRetreamer(t *testing.T) {
	s := NewSliceStreamer[[]int]()
	if s.output == nil {
		t.Fatal("Missing output")
	}
}

func TestSliceStreamerParams(t *testing.T) {
	o := NewSliceStreamer[[]int]()
	in := make(chan interface{})

	o.SetInput(in)
	if o.input == nil {
		t.Fatal("Input not being set")
	}

	if o.GetOutput() == nil {
		t.Fatal("Output not set")
	}
}

func TestSliceStreamer(t *testing.T) {
	t.Run("With string slices", func(t *testing.T) {
		o := NewSliceStreamer[[]string]()

		in := make(chan any)
		go func() {
			in <- []string{"A", "B", "C"}
			in <- []string{"E", "F", "G", "H"}
			in <- []string{"I", "J"}
			close(in)
		}()
		o.SetInput(in)

		counter := 0
		expected := 9
		var m sync.Mutex

		wait := make(chan struct{})
		go func(t *testing.T) {
			defer close(wait)
			for item := range o.GetOutput() {
				m.Lock()
				_, ok := item.(string)
				if !ok {
					t.Errorf("Expecting type string, got %T", item)
				}
				counter++
				m.Unlock()
			}
		}(t)

		if err := o.Exec(context.TODO()); err != nil {
			t.Fatal(err)
		}

		select {
		case <-wait:
			if counter != expected {
				t.Fatalf("Expecting %d items, but got %d", expected, counter)
			}
		case <-time.After(50 * time.Millisecond):
			t.Fatal("Took too long...")
		}
	})

	t.Run("with int slices", func(t *testing.T) {
		o := NewSliceStreamer[[]int]()

		in := make(chan any)
		go func() {
			in <- []int{1, 2, 3}
			in <- []int{1, 2, 3}
			in <- []int{1, 2, 3}
			in <- []int{1, 2, 3}
			close(in)
		}()
		o.SetInput(in)

		counter := 0
		expected := 12
		var m sync.Mutex

		wait := make(chan struct{})
		go func() {
			defer close(wait)
			for item := range o.GetOutput() {
				m.Lock()
				_, ok := item.(int)
				if !ok {
					t.Errorf("Expecting type int, got %T", item)
				}
				counter++
				m.Unlock()
			}
		}()

		if err := o.Exec(context.TODO()); err != nil {
			t.Fatal(err)
		}

		select {
		case <-wait:
			if counter != expected {
				t.Fatalf("Expecting %d items, but got %d", expected, counter)
			}
		case <-time.After(50 * time.Millisecond):
			t.Fatal("Took too long...")
		}
	})

	t.Run("With muti types", func(t *testing.T) {
		o := NewSliceStreamer[[]any, any]()

		in := make(chan any)
		go func() {
			in <- []any{1, "hello", 3.14}
			in <- []any{"World", "pi", 3.14}
			close(in)
		}()
		o.SetInput(in)

		counter := 0
		expected := 6
		var m sync.Mutex

		wait := make(chan struct{})
		go func() {
			defer close(wait)
			for item := range o.GetOutput() {
				m.Lock()
				_ = item
				counter++
				m.Unlock()
			}
		}()

		if err := o.Exec(context.TODO()); err != nil {
			t.Fatal(err)
		}

		select {
		case <-wait:
			if counter != expected {
				t.Fatalf("Expecting %d items, but got %d", expected, counter)
			}
		case <-time.After(50 * time.Millisecond):
			t.Fatal("Took too long...")
		}
	})

	t.Run("With empty slice", func(t *testing.T) {
		o := NewSliceStreamer[[]int]()

		in := make(chan any)
		go func() {
			in <- []int{}
			close(in)
		}()
		o.SetInput(in)

		counter := 0

		wait := make(chan struct{})
		go func() {
			defer close(wait)
			for range o.GetOutput() {
				counter++
			}
		}()

		if err := o.Exec(context.TODO()); err != nil {
			t.Fatal(err)
		}

		select {
		case <-wait:
			if counter != 0 {
				t.Fatalf("Expecting %d items, but got %d", 0, counter)
			}
		case <-time.After(50 * time.Millisecond):
			t.Fatal("Took too long...")
		}
	})
}

func BenchmarkStreamOp_Exec(b *testing.B) {
	o := NewSliceStreamer[[]string]()
	N := b.N

	chanSize := func() int {
		if N == 1 {
			return N
		}
		return int(float64(0.5) * float64(N))
	}()

	in := make(chan any, chanSize)
	o.SetInput(in)
	go func() {
		for i := 0; i < N; i++ {
			in <- []string{
				testutil.GenWord(),
				testutil.GenWord(),
				testutil.GenWord(),
			}
		}
		close(in)
	}()

	counter := 0
	expected := N * 3
	var m sync.RWMutex

	// process output
	done := make(chan struct{})
	go func() {
		defer close(done)
		for range o.GetOutput() {
			m.Lock()
			counter++
			m.Unlock()
		}
	}()

	if err := o.Exec(context.TODO()); err != nil {
		b.Fatal("Error during execution:", err)
	}

	select {
	case <-done:
	case <-time.After(time.Second * 60):
		b.Fatal("Took too long")
	}
	m.RLock()
	b.Logf("Input %d, counted %d", N, counter)
	if counter != expected {
		b.Fatalf("Expected %d items processed,  got %d", expected, counter)
	}
	m.RUnlock()
}
