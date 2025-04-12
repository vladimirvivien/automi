package window

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/vladimirvivien/automi/testutil"
)

func TestWindowNew(t *testing.T) {
	s := New[any](nil)
	if s.output == nil {
		t.Error("missing output")
	}
}

func TestWindowGettersSetters(t *testing.T) {
	o := New[any](nil)
	in := make(chan interface{})

	o.SetInput(in)
	if o.input == nil {
		t.Error("input not being set")
	}

	if o.GetOutput() == nil {
		t.Fatal("output not set")
	}
}

func TestWindowExec_TriggerNone(t *testing.T) {
	o := New[string](nil)

	in := make(chan any)
	go func() {
		in <- "A"
		in <- "B"
		in <- "C"
		in <- "D"
		in <- "E"
		in <- "F"
		in <- "G"
		in <- "H"
		close(in)
	}()
	o.SetInput(in)

	var batches atomic.Int32
	expectedBatches := 1
	var batchSize atomic.Int32

	wait := make(chan struct{})
	go func() {
		defer close(wait)
		for data := range o.GetOutput() {
			batch := data.([]string)
			batches.Add(1)
			batchSize.Add(int32(len(batch)))
		}
	}()

	if err := o.Exec(context.TODO()); err != nil {
		t.Fatal(err)
	}

	select {
	case <-wait:
		if int(batches.Load()) != expectedBatches {
			t.Fatalf("Expecting %d batch, but got %d", expectedBatches, batches.Load())
		}
		if batchSize.Load() != 8 {
			t.Fatal("unexpected batch size ", batchSize.Load())
		}
	case <-time.After(50 * time.Millisecond):
		t.Fatal("Took too long...")
	}
}

func TestWindowExec_TriggerBySize(t *testing.T) {
	o := New(TriggerBySizeFunc[string](6))
	in := make(chan any)
	go func() {
		in <- "A"
		in <- "B"
		in <- "C"
		in <- "D"
		in <- "E"
		in <- "F"

		in <- "G"
		in <- "H"
		in <- "I"
		in <- "J"
		in <- "K"
		in <- "L"

		in <- "M"
		in <- "N"
		in <- "O"
		in <- "P"
		in <- "Q"
		in <- "R"

		in <- "S"
		in <- "T"
		in <- "U"
		in <- "V"
		in <- "W"
		in <- "X"

		in <- "Y"
		in <- "Z"
		close(in)
	}()
	o.SetInput(in)

	var batches atomic.Int32
	expectedBatches := 5

	wait := make(chan struct{})
	go func() {
		defer close(wait)
		for data := range o.GetOutput() {
			batch := data.([]string)
			batchSize := len(batch)
			t.Log("got batch size:", batchSize)
			if batchSize != 6 && batchSize != 2 {
				t.Error("unexpected batch size:", batchSize)
			}
			batches.Add(1)
		}
	}()

	if err := o.Exec(context.TODO()); err != nil {
		t.Fatal(err)
	}

	select {
	case <-wait:
		if int(batches.Load()) != expectedBatches {
			t.Fatalf("Expecting %d batch, but got %d", expectedBatches, batches.Load())
		}
	case <-time.After(50 * time.Millisecond):
		t.Fatal("Took too long...")
	}
}

func TestWindowSlice(t *testing.T) {
	o := New[[]string](nil)

	in := make(chan any)
	go func() {
		in <- []string{"AA", "BB"}
		in <- []string{"CC", "DD"}
		in <- []string{"EE", "FF", "GG"}
		close(in)
	}()
	o.SetInput(in)
	wait := make(chan struct{})
	var count atomic.Int32
	var typeOk atomic.Bool
	go func() {
		defer close(wait)
		for data := range o.GetOutput() {
			count.Add(1)
			_, ok := data.([][]string)
			typeOk.Store(ok)
		}
	}()

	if err := o.Exec(context.TODO()); err != nil {
		t.Fatal(err)
	}

	select {
	case <-wait:
		if !typeOk.Load() {
			t.Error("unexpected batch type")
		}
		if count.Load() != 1 {
			t.Fatalf("unexpected aggregation count: %d", count.Load())
		}
	case <-time.After(50 * time.Millisecond):
		t.Fatal("Took too long...")
	}

}

func TestWindowMap(t *testing.T) {
	o := New[map[string][]string](nil)
	in := make(chan any)
	go func() {
		in <- map[string][]string{"AA": {"AA-AA"}, "BB": {"BB-CC", "DC-EB"}}
		in <- map[string][]string{"CC": {"CA", "DD", "BB"}}
		in <- map[string][]string{"EE": {"CAR", "FF"}, "CEX": {"GG", "IEX"}}
		in <- map[string][]string{"CC": {"CA", "DD"}, "XE": {"BB"}}
		close(in)
	}()
	o.SetInput(in)
	wait := make(chan struct{})
	var typeOk atomic.Bool
	go func() {
		defer close(wait)
		for data := range o.GetOutput() {
			_, ok := data.([]map[string][]string)
			typeOk.Store(ok)
		}
	}()

	if err := o.Exec(context.TODO()); err != nil {
		t.Fatal(err)
	}

	select {
	case <-wait:
		if !typeOk.Load() {
			t.Fatal("Unexpected data type from window")
		}
	case <-time.After(50 * time.Millisecond):
		t.Fatal("Took too long...")
	}
}

func TestWindowStruct(t *testing.T) {
	type log struct{ Event, Req string }

	o := New[log](nil)
	in := make(chan any)
	go func() {
		in <- log{Event: "AA-AA", Req: "BB-CC"}
		in <- log{Event: "CA", Req: "BB"}
		in <- log{Event: "CAR", Req: "CEX"}
		close(in)
	}()
	o.SetInput(in)

	wait := make(chan struct{})
	var typeOk atomic.Bool

	go func() {
		defer close(wait)
		for data := range o.GetOutput() {
			_, ok := data.([]log)
			typeOk.Store(ok)
		}
	}()

	if err := o.Exec(context.TODO()); err != nil {
		t.Fatal(err)
	}

	select {
	case <-wait:
		if !typeOk.Load() {
			t.Fatal("unexpected batch type")
		}
	case <-time.After(50 * time.Millisecond):
		t.Fatal("Took too long...")
	}
}

func BenchmarkWindowExec(b *testing.B) {
	ctx := context.Background()
	N := b.N

	size := func() int {
		if N == 1 {
			return N
		}
		return int(float64(0.5) * float64(N))
	}()

	o := New[string](TriggerBySizeFunc[string](uint64(size)))

	in := make(chan any, size)
	o.SetInput(in)
	go func() {
		for i := 0; i < N; i++ {
			in <- testutil.GenWord()
		}
		close(in)
	}()

	counter := 0
	expected := int(N / size)
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

	b.Logf("Benchmark N: %d, chan size %d; batch size %d", N, size, expected)
	if err := o.Exec(ctx); err != nil {
		b.Fatal("Error during execution:", err)
	}

	select {
	case <-done:
	case <-time.After(time.Second * 60):
		b.Fatal("Took too long")
	}
}
