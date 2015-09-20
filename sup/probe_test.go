package sup

import (
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
	in := make(chan interface{})
	go func() {
		in <- 5
		in <- 5
		close(in)
	}()

	sum := 0
	p := &Probe{
		Name: "P1",
		Examine: func(data interface{}) interface{} {
			i, ok := data.(int)
			if !ok {
				t.Fatal("Unexpected type")
			}
			sum += i
			return i
		},
	}

	p.SetInput(in)

	if err := p.Init(context.TODO()); err != nil {
		t.Fatal(err)
	}

	if p.GetOutput() == nil {
		t.Fatal("Failed to create Output channel after init")
	}

	if err := p.Exec(context.TODO()); err != nil {
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
	in := make(chan interface{})
	go func() {
		in <- 5
		in <- 5
		close(in)
	}()

	submitted := 0
	p := &Probe{
		Name: "P1",
		Examine: func(data interface{}) interface{} {
			i, ok := data.(int)
			if !ok {
				t.Fatal("Unexpected type")
			}
			submitted += i
			return i
		},
	}

	p.SetInput(in)
	if err := p.Init(context.TODO()); err != nil {
		t.Fatal(err)
	}

	if err := p.Exec(context.TODO()); err != nil {
		t.Fatal(err)
	}

	calculated := 0
	p2 := &Probe{
		Name: "P2",
		Examine: func(data interface{}) interface{} {
			if s, ok := data.(int); ok {
				calculated += s
			}
			return data
		},
	}
	p2.SetInput(p.GetOutput())
	if err := p2.Init(context.TODO()); err != nil {
		t.Fatal(err)
	}

	if err := p2.Exec(context.TODO()); err != nil {
		t.Fatal(err)
	}

	if submitted != calculated {
		t.Fatalf("Data flow broken, expecting %d, got %d", submitted, calculated)
	}
}
