package stream

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/vladimirvivien/automi/testutil"
)

func TestStreamOp_New(t *testing.T) {
	s := NewStreamOp(context.Background())

	if s.output == nil {
		t.Fatal("Missing output")
	}
}

func TestStreamOp_Params(t *testing.T) {
	o := NewStreamOp(context.Background())
	in := make(chan interface{})

	o.SetInput(in)
	if o.input == nil {
		t.Fatal("Input not being set")
	}

	if o.GetOutput == nil {
		t.Fatal("Output not set")
	}
}

func TestStreamOp_Exec(t *testing.T) {
	ctx, _ := context.WithCancel(context.Background())
	o := NewStreamOp(ctx)

	in := make(chan interface{})
	go func() {
		in <- []string{"A", "B", "C"}
		in <- map[string]int{"D": 2, "E": 5, "F": 6}
		close(in)
	}()
	o.SetInput(in)

	counter := 0
	expected := 6
	var m sync.Mutex

	wait := make(chan struct{})
	go func() {
		defer close(wait)
		for _ = range o.GetOutput() {
			m.Lock()
			counter++
			m.Unlock()
		}
	}()

	if err := o.Exec(); err != nil {
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
}

func BenchmarkStreamOp_Exec(b *testing.B) {
	ctx := context.Background()
	o := NewStreamOp(ctx)
	N := b.N

	chanSize := func() int {
		if N == 1 {
			return N
		}
		return int(float64(0.5) * float64(N))
	}()

	in := make(chan interface{}, chanSize)
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
		for _ = range o.GetOutput() {
			m.Lock()
			counter++
			m.Unlock()
		}
	}()

	if err := o.Exec(); err != nil {
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
