package stream

import (
	"testing"
	"time"

	"github.com/vladimirvivien/automi/sources"
)

func TestStream_Reduce(t *testing.T) {
	src := sources.Slice(1, 2, 3, 4, 5)
	snk := NewDrain()
	strm := New().From(src).Reduce(0, func(op1, op2 int) int {
		return op1 + op2
	}).To(snk)

	actual := 15
	wait := make(chan struct{})
	go func() {
		defer close(wait)
		select {
		case result := <-snk.GetOutput():
			if result.(int) != actual {
				t.Fatal("Expecting ", actual, " got ", result)
			}
		case <-time.After(5 * time.Millisecond):
			t.Fatal("Sink took too long to get result")
		}
	}()

	select {
	case err := <-strm.Open():
		if err != nil {
			t.Fatal(err)
		}
		select {
		case <-wait:
		case <-time.After(10 * time.Millisecond):
			t.Fatal("Stream took too long to complete")
		}
	case <-time.After(10 * time.Millisecond):
		t.Fatal("Took too long")
	}
}
