package sinks

import (
	"context"
	"testing"
	"time"
)

func TestNullSink_Open(t *testing.T) {
	in := make(chan interface{})
	go func() {
		in <- []string{"Christophe", "Petion", "Dessaline"}
		in <- []string{"Toussaint", "Guerrier", "Caiman"}
		close(in)
	}()

	sink := Null()
	sink.SetInput(in)

	// process
	select {
	case err := <-sink.Open(context.Background()):
		if err != nil {
			t.Fatal(err)
		}
		if _, ok := <-in; ok {
			t.Fatal("channel should have been closed by nullsink node")
		}
	case <-time.After(50 * time.Millisecond):
		t.Fatal("Sink took too long to open")
	}

}
