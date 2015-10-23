package proc

import (
	"math/rand"
	"sync"
	"testing"
	"time"

	"github.com/vladimirvivien/automi/sup"

	"golang.org/x/net/context"
)

func TestItemInit(t *testing.T) {
	i := &Item{}
	if err := i.Init(context.TODO()); err == nil {
		t.Fatal("Error expected for missing attributes")
	}

	i = &Item{Name: "proc"}
	if err := i.Init(context.TODO()); err == nil {
		t.Fatal("Error expected for missing attributes Input and Function")
	}

	in := make(chan interface{})
	i = &Item{Name: "proc"}
	i.SetInput(in)
	if err := i.Init(context.TODO()); err == nil {
		t.Fatal("Error expected for missing attribute Function")
	}

	i = &Item{
		Name: "proc",
		Function: func(ctx context.Context, i interface{}) interface{} {
			return i
		},
	}
	i.SetInput(in)
	if err := i.Init(context.TODO()); err != nil {
		t.Fatal("No error expected after Init(): ", err)
	}

	if i.GetName() == "" {
		t.Fatal("Name attribute not set")
	}
}

func TestItemExec(t *testing.T) {
	in := make(chan interface{})
	go func() {
		in <- "A"
		in <- "B"
		in <- "C"
		close(in)
	}()

	i := &Item{
		Name: "proc",
		Function: func(ctx context.Context, i interface{}) interface{} {
			s := i.(string)
			return rune(s[0])
		},
	}
	i.SetInput(in)
	if err := i.Init(context.TODO()); err != nil {
		t.Fatal("Unable to init()", err)
	}

	// process output
	done := make(chan struct{})
	go func() {
		close(done)
		for item := range i.GetOutput() {
			r, ok := item.(rune)
			if !ok {
				t.Fatalf("Error, expecting rune type, got %T", item)
			}
			if r != 65 && r != 66 && r != 67 {
				t.Fatal("Unexpected data from process output", r)
			}
		}
	}()

	if err := i.Exec(context.TODO()); err != nil {
		t.Fatal("Error during execution:", err)
	}

	select {
	case <-done:
	case <-time.After(time.Millisecond * 50):
		t.Fatal("Took too long")
	}

}

func TestItemExec_Cancel(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	dataCount := 100
	in := make(chan interface{}, 10)
	go func() {
		defer close(in)
		rand.Seed(time.Now().Unix())
		for i := 0; i < dataCount; i++ {
			in <- sup.GenWord()
			if rand.Intn(dataCount) > 70 {
				cancel()
				return
			}
		}
	}()

	counter := 0
	var m sync.RWMutex
	i := &Item{
		Name:        "proc",
		Concurrency: 4,
		Function: func(ctx context.Context, i interface{}) interface{} {
			m.Lock()
			counter++
			m.Unlock()
			return i
		},
	}
	i.SetInput(in)
	if err := i.Init(ctx); err != nil {
		t.Fatal("Unable to init()", err)
	}

	go func() {
		if err := i.Exec(ctx); err != nil {
			t.Fatal("Error during execution:", err)
		}
	}()

	done := make(chan struct{})
	go func() {
		defer close(done)
		for _ = range i.GetOutput() {
		}
	}()

	select {
	case <-done:
	case <-time.After(time.Millisecond * 50):
		t.Fatal("Took too long")
	}

	m.RLock()
	t.Logf("Data input %d, processed %d", dataCount, counter)
	if counter <= dataCount {
		return
	} else {
		t.Fatalf("Cancellation failed. Expecting processed items %d to be less than %d", counter, dataCount)
	}
	m.RUnlock()

}

func BenchmarkItem(b *testing.B) {
	ctx := context.Background()
	N := b.N
	chanSize := func() int {
		if N == 1 {
			return N
		}
		return int(float64(0.5) * float64(N))
	}()

	in := make(chan interface{}, chanSize)
	go func() {
		for i := 0; i < N; i++ {
			in <- sup.GenWord()
		}
		close(in)
	}()

	counter := 0
	var m sync.RWMutex

	i := &Item{
		Name:        "proc",
		Concurrency: 4,
		Function: func(ctx context.Context, i interface{}) interface{} {
			m.Lock()
			counter++
			m.Unlock()
			return i
		},
	}
	i.SetInput(in)
	if err := i.Init(ctx); err != nil {
		b.Fatal("Unable to init()", err)
	}

	// process output
	done := make(chan struct{})
	go func() {
		defer close(done)
		for _ = range i.GetOutput() {
		}
	}()

	go func() {
		if err := i.Exec(ctx); err != nil {
			b.Fatal("Error during execution:", err)
		}
	}()

	select {
	case <-done:
	case <-time.After(time.Second * 60):
		b.Fatal("Took too long")
	}
	m.RLock()
	b.Logf("Input %d, counted %d", N, counter)
	if counter != N {
		b.Fatalf("Expected %d items processed,  got %d", N, counter)
	}
	m.RUnlock()
}
