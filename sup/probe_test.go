package sup

import (
	"sync"
	"testing"
	"time"

	"golang.org/x/net/context"
)

func TestProbeInit(t *testing.T) {
	in := make(chan interface{})
	p := &Probe{Name: "P1"}
	p.SetInput(in)

	if err := p.Init(context.TODO()); err != nil {
		t.Fatal(err)
	}

	if p.GetName() != p.Name {
		t.Fatal("Error, GetName() not returning Name attribute")
	}
	if p.input != in {
		t.Fatal("Error, GetInput().Items() not returning expected channel")
	}
	if p.GetOutput() == nil {
		t.Fatal("Error, GetOutput() should not be nil")
	}
}

func TestProbeExec(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	in := make(chan interface{}, 10)
	in <- 5
	in <- 5

	sum := 0
	p := &Probe{
		Name: "P1",
		Examine: func(ctx context.Context, data interface{}) interface{} {
			i, ok := data.(int)
			if !ok {
				t.Fatal("Unexpected type")
			}
			sum += i
			return i
		},
	}

	p.SetInput(in)

	if err := p.Init(ctx); err != nil {
		t.Fatal(err)
	}

	if p.GetOutput() == nil {
		t.Fatal("Failed to create Output channel after init")
	}

	if err := p.Exec(ctx); err != nil {
		t.Fatal(err)
	}

	actual := 0
	wait := make(chan struct{})
	go func() {
		defer close(wait)
		for i := range p.GetOutput() {
			if s, ok := i.(int); ok {
				actual += s
			}
		}
	}()

	cancel()

	select {
	case <-wait:
	case <-time.After(5 * time.Millisecond):
		t.Fatal("Took too long")
	}

	if sum != actual {
		t.Fatalf("Data flow broken, expecting %d, got %d", sum, actual)
	}
}

func TestProbeChaining(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	in := make(chan interface{}, 2)
	in <- 5
	in <- 5

	var m sync.RWMutex
	submitted := 0
	p := &Probe{
		Name: "P1",
		Examine: func(ctx context.Context, data interface{}) interface{} {
			i, ok := data.(int)
			if !ok {
				t.Fatal("Unexpected type")
			}
			m.Lock()
			submitted += i
			m.Unlock()
			return i
		},
	}

	p.SetInput(in)
	if err := p.Init(ctx); err != nil {
		t.Fatal(err)
	}

	if err := p.Exec(ctx); err != nil {
		t.Fatal(err)
	}

	calculated := 0
	p2 := &Probe{
		Name: "P2",
		Examine: func(ctx context.Context, data interface{}) interface{} {
			if s, ok := data.(int); ok {
				m.Lock()
				calculated += s
				m.Unlock()
			}
			return data
		},
	}
	p2.SetInput(p.GetOutput())
	if err := p2.Init(ctx); err != nil {
		t.Fatal(err)
	}

	if err := p2.Exec(ctx); err != nil {
		t.Fatal(err)
	}

	wait := make(chan struct{})
	go func() {
		for _ = range p2.output {
		}
		close(wait)
	}()

	cancel()

	select {
	case <-wait:
	case <-time.After(1 * time.Millisecond):
		t.Fatal("Took too long...")
	}

	m.RLock()
	if submitted != calculated {
		t.Fatalf("Data flow broken, expecting %d, got %d", submitted, calculated)
	}
	m.RUnlock()
}

func BenchmarkProbe(b *testing.B) {
	ctx := context.Background()

	in := make(chan interface{}, b.N)
	go func() {
		for i := 0; i < b.N; i++ {
			in <- i
		}
		close(in)
	}()

	p := &Probe{
		Name: "P1",
		Examine: func(ctx context.Context, data interface{}) interface{} {
			return data
		},
	}

	p.SetInput(in)
	if err := p.Init(ctx); err != nil {
		b.Fatal(err)
	}

	if err := p.Exec(ctx); err != nil {
		b.Fatal(err)
	}

	counter := 0
	var mutex sync.RWMutex
	p2 := &Probe{
		Name: "P2",
		Examine: func(ctx context.Context, data interface{}) interface{} {
			mutex.Lock()
			counter++
			mutex.Unlock()
			return data
		},
	}
	p2.SetInput(p.GetOutput())
	if err := p2.Init(ctx); err != nil {
		b.Fatal(err)
	}

	if err := p2.Exec(ctx); err != nil {
		b.Fatal(err)
	}

	wait := make(chan struct{})
	go func() {
		for _ = range p2.output {
		}
		close(wait)
	}()

	select {
	case <-wait:
	case <-time.After(60 * time.Second):
		b.Fatal("Took too long...")
	}

	mutex.RLock()
	if counter != b.N {
		b.Fatal("Not processing all channel items")
	}
	mutex.RUnlock()
}

func BenchmarkProbe_WithCancel(b *testing.B) {
	ctx, cancel := context.WithCancel(context.Background())
	N := b.N

	in := make(chan interface{}, b.N)
	go func() {
		for i := 0; i < N; i++ {
			in <- i
		}
		cancel()
	}()

	p := &Probe{
		Name: "P1",
		Examine: func(ctx context.Context, data interface{}) interface{} {
			return data
		},
	}

	p.SetInput(in)
	if err := p.Init(ctx); err != nil {
		b.Fatal(err)
	}

	if err := p.Exec(ctx); err != nil {
		b.Fatal(err)
	}

	counter := 0
	var mutex sync.RWMutex
	p2 := &Probe{
		Name: "P2",
		Examine: func(ctx context.Context, data interface{}) interface{} {
			mutex.Lock()
			counter++
			mutex.Unlock()
			return data
		},
	}
	p2.SetInput(p.GetOutput())
	if err := p2.Init(ctx); err != nil {
		b.Fatal(err)
	}

	if err := p2.Exec(ctx); err != nil {
		b.Fatal(err)
	}

	wait := make(chan struct{})
	go func() {
		for _ = range p2.output {
		}
		close(wait)
	}()

	select {
	case <-wait:
	case <-time.After(60 * time.Second):
		b.Fatal("Took too long...")
	}

	mutex.RLock()
	if counter == N {
		return
	}
	if counter >= 1 && counter < N {
		return
	} else {
		b.Fatal("Not cancelling, counter = ", counter, " benchmark.N = ", N)
	}
	mutex.RUnlock()
}
