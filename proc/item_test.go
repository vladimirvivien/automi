package proc

import (
	"testing"
	"time"

	"golang.org/x/net/context"
)

func TestItem_Init(t *testing.T) {
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

func TestItem_Exec(t *testing.T) {
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
