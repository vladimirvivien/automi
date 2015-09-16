package route

import (
	"fmt"
	"testing"
	"time"

	"github.com/vladimirvivien/automi/api"
)

func TestErrCollector_Init(t *testing.T) {
	e := &ErrCollector{}
	if err := e.Init(); err == nil {
		t.Fatal("Expected error for missing attributes")
	}
	e = &ErrCollector{Name: "errors"}
	if err := e.Init(); err == nil {
		t.Fatal("Expected error for missing Input attribute")
	}

	errs := []<-chan api.ProcError{make(chan api.ProcError)}
	e = &ErrCollector{Name: "errors", Input: errs}
	if err := e.Init(); err != nil {
		t.Fatal("Unexpected error after init(): ", err)
	}
	if e.GetOutput() == nil {
		t.Fatal("Output is not set after Init()")
	}
}

func TestErrCollector_Exec(t *testing.T) {
	errs1 := make(chan api.ProcError)
	errs2 := make(chan api.ProcError)
	genErr := func() {
		errs1 <- api.ProcError{ProcName: "err", Err: fmt.Errorf("Error")}
		errs1 <- api.ProcError{ProcName: "err", Err: fmt.Errorf("Error")}
		close(errs1)

		errs2 <- api.ProcError{ProcName: "err", Err: fmt.Errorf("Error")}
		errs2 <- api.ProcError{ProcName: "err", Err: fmt.Errorf("Error")}
		errs2 <- api.ProcError{ProcName: "err", Err: fmt.Errorf("Error")}
		close(errs2)
	}

	errs := []<-chan api.ProcError{errs1, errs2}
	errCol := &ErrCollector{
		Name:  "errors",
		Input: errs,
	}
	if err := errCol.Init(); err != nil {
		t.Fatal(err)
	}

	go genErr()

	count := 0
	wait := make(chan struct{})
	go func(c *int) {
		t.Log("Outputting")
		defer func() {
			t.Log("Done waiting for GetOutput()")
			close(wait)
		}()
		for _ = range errCol.GetOutput() {
			t.Log("Counting")
			*c = *c + 1
		}
	}(&count)

	if err := errCol.Exec(); err != nil {
		t.Fatal(err)
	}
	select {
	case <-wait:
	case <-time.After(500 * time.Millisecond):
		t.Fatal("Waited too long for result")
	}
	if count != 5 {
		t.Fatal("Expecting count to be 5, got", count)
	}
}
