package route

import (
	"fmt"
	"testing"
	"time"

	"golang.org/x/net/context"

	"github.com/vladimirvivien/automi/api"
)

func TestItemCollector_Init(t *testing.T) {
	e := &ItemCollector{}
	if err := e.Init(context.TODO()); err == nil {
		t.Fatal("Expected error for missing attributes")
	}
	e = &ItemCollector{Name: "errors"}
	if err := e.Init(context.TODO()); err == nil {
		t.Fatal("Expected error for missing Input attribute")
	}

	errs := []<-chan interface{}{make(chan interface{})}
	e = &ItemCollector{Name: "errors"}
	e.SetInputs(errs)
	if err := e.Init(context.TODO()); err != nil {
		t.Fatal("Unexpected error after init(): ", err)
	}
	if e.GetOutput() == nil {
		t.Fatal("Output is not set after Init()")
	}
}

func TestItemCollector_Exec(t *testing.T) {
	errs1 := make(chan interface{})
	errs2 := make(chan interface{})
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

	errs := []<-chan interface{}{errs2, errs1}
	errCol := &ItemCollector{
		Name: "errors",
	}
	errCol.SetInputs(errs)
	if err := errCol.Init(context.TODO()); err != nil {
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

	if err := errCol.Exec(context.TODO()); err != nil {
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
