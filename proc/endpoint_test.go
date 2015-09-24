package proc

import (
	"testing"
	"time"

	"golang.org/x/net/context"
)

func TestEndpoint_Init(t *testing.T) {
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
		Function: func(i interface{}) error {
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

func TestEndpoint_Exec(t *testing.T) {
	in := make(chan interface{})
	go func() {
		in <- "A"
		in <- "B"
		in <- "C"
		close(in)
	}()

	i := &Endpoint{
		Name: "proc",
		Function: func(i interface{}) error {
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
