package proc

import (
	"math/rand"
	"sync"
	"testing"
	"time"

	"github.com/vladimirvivien/automi/testutil"

	"golang.org/x/net/context"
)

func TestEndpointInit(t *testing.T) {
	i := &Endpoint{}
	if err := i.Init(context.TODO()); err == nil {
		t.Fatal("Error expected for missing attributes")
	}

	i = &Endpoint{Name: "proc"}
	if err := i.Init(context.TODO()); err == nil {
		t.Fatal("Error expected for missing attributes Input and Function")
	}

	in := make(chan interface{})
	i = &Endpoint{Name: "proc"}
	i.SetInput(in)
	if err := i.Init(context.TODO()); err == nil {
		t.Fatal("Error expected for missing attribute Function")
	}

	i = &Endpoint{
		Name: "proc",
		Function: func(ctx context.Context, i interface{}) error {
			return nil
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

func TestEndpointExec(t *testing.T) {
	in := make(chan interface{}, 5)
	go func() {
		in <- "A"
		in <- "B"
		in <- "C"
		close(in)
	}()

	i := &Endpoint{
		Name: "proc",
		Function: func(ctx context.Context, i interface{}) error {
			s := i.(string)
			if s != "A" && s != "B" && s != "C" {
				t.Fatal("Unexpected data from input", s)
			}
			return nil
		},
	}
	i.SetInput(in)
	if err := i.Init(context.TODO()); err != nil {
		t.Fatal("Unable to init()", err)
	}

	if err := i.Exec(context.TODO()); err != nil {
		t.Fatal("Error during execution:", err)
	}

	select {
	case <-i.Done():
	case <-time.After(time.Millisecond * 50):
		t.Fatal("Took too long")
	}

}

func TestEndpointExec_Cancel(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	dataCount := 100
	in := make(chan interface{}, 20)
	go func() {
		defer close(in)
		rand.Seed(time.Now().Unix())
		for i := 0; i < dataCount; i++ {
			in <- i
			if rand.Intn(dataCount) > 50 {
				cancel()
				return
			}
		}
	}()

	counter := 0
	var m sync.RWMutex
	i := &Endpoint{
		Name:        "proc",
		Concurrency: 4,
		Function: func(ctx context.Context, i interface{}) error {
			m.Lock()
			counter++
			m.Unlock()
			return nil
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

	select {
	case <-i.Done():
	case <-time.After(time.Millisecond * 100):
		t.Fatal("Took too long")
	}

	m.RLock()
	t.Logf("Input %d, counted %d", dataCount, counter)
	if counter <= dataCount {
		return
	} else {
		t.Fatalf("Cancellation failed.  Expected counter %d to less than 100", counter)
	}
	m.RUnlock()
}

func BenchmarkEndpoint(b *testing.B) {
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
			in <- testutil.GenWord()
		}
		close(in)
	}()

	counter := 0
	var m sync.RWMutex
	i := &Endpoint{
		Name:        "proc",
		Concurrency: 4,
		Function: func(ctx context.Context, i interface{}) error {
			m.Lock()
			counter++
			m.Unlock()
			return nil
		},
	}
	i.SetInput(in)
	if err := i.Init(ctx); err != nil {
		b.Fatal("Unable to init()", err)
	}

	if err := i.Exec(ctx); err != nil {
		b.Fatal("Error during execution:", err)
	}

	select {
	case <-i.Done():
	case <-time.After(60 * time.Second):
		b.Fatal("Took too long")
	}

	m.RLock()
	b.Logf("Input %d, counted %d", N, counter)
	if counter != N {
		b.Fatalf("Expected %d items processed,  got %d", N, counter)
	}
	m.RUnlock()

}
