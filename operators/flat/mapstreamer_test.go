package flat

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/vladimirvivien/automi/api/tuple"
	"github.com/vladimirvivien/automi/testutil"
)

func TestMapStreamer_Exec(t *testing.T) {
	t.Run("with map with multiple items", func(t *testing.T) {
		o := NewMapStreamer[map[string]int]()
		input := make(chan any)
		output := o.GetOutput()

		go func() {
			input <- map[string]int{"a": 1, "b": 2, "c": 3}
			input <- map[string]int{"d": 4, "e": 5, "f": 6}
			input <- map[string]int{"g": 7, "h": 8, "i": 9}
			close(input)
		}()

		o.SetInput(input)

		var results []tuple.Pair[string, int]
		var wg sync.WaitGroup
		wg.Add(1)

		go func() {
			defer wg.Done()
			for item := range output {
				pair, ok := item.(tuple.Pair[string, int])
				if !ok {
					t.Errorf("Expected type tuple.Pair[string, int], got %T", item)
					return
				}
				results = append(results, pair)
			}
		}()

		if err := o.Exec(context.Background()); err != nil {
			t.Fatal(err)
		}

		wg.Wait()

		expected := 9
		if len(results) != expected {
			t.Fatalf("Expected %d results, got %d", expected, len(results))
		}
	})

	t.Run("With non map data item", func(t *testing.T) {
		o := NewMapStreamer[map[string]int]()
		input := make(chan any)
		output := o.GetOutput()

		go func() {
			input <- "hello"
			close(input)
		}()

		o.SetInput(input)

		var result any
		var wg sync.WaitGroup
		wg.Add(1)

		go func() {
			defer wg.Done()
			for item := range output {
				result = item
			}
		}()

		if err := o.Exec(context.Background()); err != nil {
			t.Fatal(err)
		}

		wg.Wait()

		expected := "hello"
		if result != expected {
			t.Fatalf("Expected %v, got %v", expected, result)
		}
	})

	t.Run("with empty map", func(t *testing.T) {
		o := NewMapStreamer[map[string]int]()
		input := make(chan any)
		output := o.GetOutput()

		go func() {
			input <- map[string]int{}
			close(input)
		}()

		o.SetInput(input)

		var count int
		var wg sync.WaitGroup
		wg.Add(1)

		go func() {
			defer wg.Done()
			for range output {
				count++
			}
		}()

		if err := o.Exec(context.Background()); err != nil {
			t.Fatal(err)
		}

		wg.Wait()

		expected := 0
		if count != expected {
			t.Fatalf("Expected %d items, got %d", expected, count)
		}
	})

	t.Run("with context cancellation", func(t *testing.T) {
		o := NewMapStreamer[map[string]int, string, int]()
		input := make(chan any)
		output := o.GetOutput()

		ctx, cancel := context.WithCancel(context.Background())

		o.SetInput(input)

		var count int
		var wg sync.WaitGroup
		wg.Add(1)

		go func() {
			defer wg.Done()
			for range output {
				count++
			}
		}()

		go func() {
			time.Sleep(10 * time.Millisecond)
			cancel()
			close(input)
		}()

		if err := o.Exec(ctx); err != nil {
			t.Fatal(err)
		}

		wg.Wait()

		if count > 0 {
			t.Fatalf("Expected 0 items or a few, got %d", count)
		}
	})
}

func BenchmarkMapStreamer_Exec(b *testing.B) {
	o := NewMapStreamer[map[string]string]()
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
			in <- map[string]string{
				testutil.GenWord(): testutil.GenWord(),
			}
		}
		close(in)
	}()

	var counter atomic.Int64
	expected := N

	// process output
	done := make(chan struct{})
	go func() {
		for range o.GetOutput() {
			counter.Add(1)
		}
		close(done)
	}()

	if err := o.Exec(context.TODO()); err != nil {
		b.Fatal("Error during execution:", err)
	}

	select {
	case <-done:
	case <-time.After(time.Second * 60):
		b.Fatal("Took too long")
	}

	b.Logf("Input %d, counted %d", N, counter.Load())
	if int(counter.Load()) != expected {
		b.Fatalf("Expected %d items processed,  got %d", expected, counter.Load())
	}
}
