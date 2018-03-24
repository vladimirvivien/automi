package batch

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/vladimirvivien/automi/testutil"
)

func TestBatchOp_New(t *testing.T) {
	s := New(context.Background())

	if s.output == nil {
		t.Error("missing output")
	}
	if s.log == nil {
		t.Error("missing logger")
	}
	if s.ctx == nil {
		t.Error("context not set")
	}
}

func TestBatchOp_GettersSetters(t *testing.T) {
	o := New(context.Background())
	in := make(chan interface{})

	o.SetInput(in)
	if o.input == nil {
		t.Error("input not being set")
	}

	if o.GetOutput() == nil {
		t.Fatal("output not set")
	}
}

func TestBatchOp_Exec_OneBatch(t *testing.T) {
	ctx := context.Background()
	o := New(ctx)

	in := make(chan interface{})
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

	batches := 0
	expectedBatches := 1
	batchSize := 0
	var m sync.Mutex

	wait := make(chan struct{})
	go func() {
		defer close(wait)
		for data := range o.GetOutput() {
			batch := data.([]string)
			m.Lock()
			batches++
			batchSize = len(batch)
			m.Unlock()
		}
	}()

	if err := o.Exec(); err != nil {
		t.Fatal(err)
	}

	select {
	case <-wait:
		if batches != expectedBatches {
			t.Fatalf("Expecting %d batch, but got %d", expectedBatches, batches)
		}
		if batchSize != 26 {
			t.Fatal("unexpected batch size ", batchSize)
		}
	case <-time.After(50 * time.Millisecond):
		t.Fatal("Took too long...")
	}
}

func TestBatchOp_Exec_MultipleBatches(t *testing.T) {
	ctx := context.Background()
	o := New(ctx)
	o.SetTrigger(TriggerBySize(4))
	in := make(chan interface{})
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
		close(in)
	}()
	o.SetInput(in)

	batches := 0
	expectedBatches := 3
	var m sync.Mutex

	wait := make(chan struct{})
	go func() {
		defer close(wait)
		for data := range o.GetOutput() {
			batch := data.([]string)
			batchSize := len(batch)
			t.Log("got batch size:", batchSize)
			if batchSize != 4 && batchSize != 2 {
				t.Fatal("unexpected batch size:", batchSize)
			}

			m.Lock()
			batches++
			m.Unlock()
		}
	}()

	if err := o.Exec(); err != nil {
		t.Fatal(err)
	}

	select {
	case <-wait:
		if batches != expectedBatches {
			t.Fatalf("Expecting %d batch, but got %d", expectedBatches, batches)
		}
	case <-time.After(50 * time.Millisecond):
		t.Fatal("Took too long...")
	}
}

func TestBatchOp_BatchSlice(t *testing.T) {
	o := New(context.TODO())

	in := make(chan interface{})
	go func() {
		in <- []string{"AA", "BB"}
		in <- []string{"CC", "DD"}
		in <- []string{"EE", "FF", "GG"}
		close(in)
	}()
	o.SetInput(in)

	wait := make(chan struct{})
	var typeOk bool

	go func() {
		defer close(wait)
		for data := range o.GetOutput() {
			_, typeOk = data.([][]string)
		}
	}()

	if err := o.Exec(); err != nil {
		t.Fatal(err)
	}

	select {
	case <-wait:
		if !typeOk {
			t.Fatal("unexpected batch type")
		}
	case <-time.After(50 * time.Millisecond):
		t.Fatal("Took too long...")
	}

}

func TestBatchOp_BatchMap(t *testing.T) {
	o := New(context.TODO())

	in := make(chan interface{})
	go func() {
		in <- map[string]string{"AA": "AA-AA", "BB": "BB-CC"}
		in <- map[string]string{"CC": "CA", "DD": "BB"}
		in <- map[string]string{"EE": "CAR", "FF": "CEX", "GG": "IEX"}
		close(in)
	}()
	o.SetInput(in)

	wait := make(chan struct{})
	var typeOk bool

	go func() {
		defer close(wait)
		for data := range o.GetOutput() {
			_, typeOk = data.([]map[string]string)
		}
	}()

	if err := o.Exec(); err != nil {
		t.Fatal(err)
	}

	select {
	case <-wait:
		if !typeOk {
			t.Fatal("unexpected batch type")
		}
	case <-time.After(50 * time.Millisecond):
		t.Fatal("Took too long...")
	}
}

func TestBatchOp_BatchStruct(t *testing.T) {
	o := New(context.TODO())
	type log struct{ Event, Req string }
	in := make(chan interface{})
	go func() {
		in <- log{Event: "AA-AA", Req: "BB-CC"}
		in <- log{Event: "CA", Req: "BB"}
		in <- log{Event: "CAR", Req: "CEX"}
		close(in)
	}()
	o.SetInput(in)

	wait := make(chan struct{})
	var typeOk bool

	go func() {
		defer close(wait)
		for data := range o.GetOutput() {
			_, typeOk = data.([]log)
		}
	}()

	if err := o.Exec(); err != nil {
		t.Fatal(err)
	}

	select {
	case <-wait:
		if !typeOk {
			t.Fatal("unexpected batch type")
		}
	case <-time.After(50 * time.Millisecond):
		t.Fatal("Took too long...")
	}
}

func BenchmarkBatchOp_Exec(b *testing.B) {
	ctx := context.Background()
	o := New(ctx)
	N := b.N
	size := func() int {
		if N == 1 {
			return N
		}
		return int(float64(0.5) * float64(N))
	}()

	in := make(chan interface{}, size)
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
		for _ = range o.GetOutput() {
			m.Lock()
			counter++
			m.Unlock()
		}
	}()

	b.Logf("Benchmark N: %d, chan size %d; batch size %d", N, size, expected)
	if err := o.Exec(); err != nil {
		b.Fatal("Error during execution:", err)
	}

	select {
	case <-done:
	case <-time.After(time.Second * 60):
		b.Fatal("Took too long")
	}
	m.RLock()
	if counter != expected {
		b.Fatalf("Expected %d batch,  got %d", expected, counter)
	}
	m.RUnlock()
}
