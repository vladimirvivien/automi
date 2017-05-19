package collectors

import (
	"context"
	"testing"
	"time"
)

func TestCollector_Slice(t *testing.T) {
	sc := Slice()
	in := make(chan interface{})
	go func() {
		in <- "A"
		in <- "B"
		in <- "C"
		in <- "D"
		in <- "E"
		in <- "F"
		close(in)
	}()
	sc.SetInput(in)

	select {
	case err := <-sc.Open(context.TODO()):
		if err != nil {
			t.Fatal(err)
		}
		result := sc.Get()
		if len(result) != 6 {
			t.Fatal("unexpected slice length ", len(result))
		}
	case <-time.After(50 * time.Millisecond):
		t.Fatal("Waited too long ...")
	}
}
