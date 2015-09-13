package sup

import (
	"testing"
)

func TestProbeCreation(t *testing.T) {
	in := make(chan interface{})
	p := &Probe{Name: "P1", Input: in, Output: nil}
	if p.GetName() != p.Name {
		t.Fatal("Error, GetName() not returning Name attribute")
	}
	if p.GetInput() != in {
		t.Fatal("Error, GetInput().Items() not returning expected channel")
	}
	if p.GetOutput() != nil {
		t.Fatal("Error, GetOutput should be nil")
	}
}

func TestProbeInit(t *testing.T) {
	in := make(chan interface{})
	p := &Probe{
		Name:  "P1",
		Input: in,
	}
	err := p.Init()
	if err != nil {
		t.Fatal(err)
	}

	if p.GetInput() != p.Input {
		t.Fatal("Failed to set input properly")
	}
	if p.GetOutput() == nil {
		t.Fatal("Failed to create Output channel after init()")
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
		Name:  "P1",
		Input: in,
		Examine: func(data interface{}) interface{} {
			i, ok := data.(int)
			if !ok {
				t.Fatal("Unexpected type")
			}
			sum += i
			return i
		},
	}

	if err := p.Init(); err != nil {
		t.Fatal(err)
	}

	if p.GetOutput() == nil {
		t.Fatal("Failed to create Output channel after init")
	}
	if p.GetInput() != in {
		t.Fatal("Failed to set Input channel after Init()")
	}

	if err := p.Exec(); err != nil {
		t.Fatal(err)
	}

	actual := 0
	for i := range p.GetOutput() {
		if s, ok := i.(int); ok {
			actual += s
		}
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
		Name:  "P1",
		Input: in,
		Examine: func(data interface{}) interface{} {
			i, ok := data.(int)
			if !ok {
				t.Fatal("Unexpected type")
			}
			submitted += i
			return i
		},
	}
	if err := p.Init(); err != nil {
		t.Fatal(err)
	}

	if err := p.Exec(); err != nil {
		t.Fatal(err)
	}

	calculated := 0
	p2 := &Probe{
		Name:  "P2",
		Input: p.GetOutput(),
		Examine: func(data interface{}) interface{} {
			if s, ok := data.(int); ok {
				calculated += s
			}
			return data
		},
	}

	if err := p2.Init(); err != nil {
		t.Fatal(err)
	}

	if err := p2.Exec(); err != nil {
		t.Fatal(err)
	}

	if submitted != calculated {
		t.Fatalf("Data flow broken, expecting %d, got %d", submitted, calculated)
	}
}
