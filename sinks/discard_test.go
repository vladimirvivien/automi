package sinks

import (
	"context"
	"testing"
	"time"
)

func TestCollector_Null(t *testing.T) {
	nc := Discard()
	in := make(chan any)
	go func() {
		in <- "A"
		in <- "B"
		in <- "C"
		in <- "D"
		in <- "E"
		in <- "F"
		close(in)
	}()
	nc.SetInput(in)

	select {
	case err := <-nc.Open(context.TODO()):
		if err != nil {
			t.Fatal(err)
		}
		_, opened := <-in
		if opened {
			t.Fatal("expected closed channel")
		}
	case <-time.After(50 * time.Millisecond):
		t.Fatal("Waited too long ...")
	}
}
