package proc

import (
	"testing"
	"time"
)

func TestItemProc_Init(t *testing.T) {
	i := &ItemProc{}
	if err := i.Init(); err == nil {
		t.Fatal("Error expected for missing attributes")
	}

	i = &ItemProc{Name: "proc"}
	if err := i.Init(); err == nil {
		t.Fatal("Error expected for missing attributes Input and Function")
	}

	in := make(chan interface{})
	i = &ItemProc{Name: "proc", Input: in}
	if err := i.Init(); err == nil {
		t.Fatal("Error expected for missing attribute Function")
	}

	i = &ItemProc{
		Name:  "proc",
		Input: in,
		Function: func(i interface{}) interface{} {
			return i
		},
	}

	if err := i.Init(); err != nil {
		t.Fatal("No error expected after Init(): ", err)
	}

	if i.GetName() == "" {
		t.Fatal("Name attribute not set")
	}
	if i.GetInput() == nil {
		t.Fatal("Input should not be nil after Init()")
	}
	if i.GetLogs() == nil {
		t.Fatal("Logs channel not available after Init()")
	}
}

func TestItemProc_Exec(t *testing.T) {
	in := make(chan interface{})
	go func() {
		in <- "A"
		in <- "B"
		in <- "C"
		close(in)
	}()

	i := &ItemProc{
		Name:  "proc",
		Input: in,
		Function: func(i interface{}) interface{} {
			s := i.(string)
			return rune(s[0])
		},
	}

	if err := i.Init(); err != nil {
		t.Fatal("Unable to init()", err)
	}

	//grab errors
	go func() {
		for e := range i.GetLogs() {
			t.Log(e)
		}
	}()

	// process output
	go func() {
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

	if err := i.Exec(); err != nil {
		t.Fatal("Error during execution:", err)
	}

	select {
	case <-i.Done():
	case <-time.After(time.Millisecond * 50):
		t.Fatal("Took too long")
	}

}
