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
	go func() {
		errs1 <- api.ProcError{ProcName: "err", Err: fmt.Errorf("Error")}
		errs1 <- api.ProcError{ProcName: "err", Err: fmt.Errorf("Error")}
		close(errs1)
	}()
	go func() {
		errs2 <- api.ProcError{ProcName: "err", Err: fmt.Errorf("Error")}
		errs2 <- api.ProcError{ProcName: "err", Err: fmt.Errorf("Error")}
		errs2 <- api.ProcError{ProcName: "err", Err: fmt.Errorf("Error")}
		close(errs2)
	}()

	errs := []<-chan api.ProcError{errs2, errs1}
	errCol := &ErrCollector{
		Name:  "errors",
		Input: errs,
	}
	if err := errCol.Init(); err != nil {
		t.Fatal(err)
	}

	count := 0
	wait := make(chan struct{})
	go func() {
		defer func() {
			close(wait)
		}()
		for _ = range errCol.GetOutput() {
			count++
		}
	}()

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
