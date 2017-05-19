package stream

import (
	"testing"
	"time"

	"github.com/vladimirvivien/automi/collectors"
	"github.com/vladimirvivien/automi/emitters"
)

func TestStream_Reduce(t *testing.T) {
	snk := collectors.Slice()
	strm := New(emitters.Slice([]int{1, 2, 3, 4, 5})).Reduce(0, func(op1, op2 int) int {
		return op1 + op2
	}).SinkTo(snk)

	actual := 15

	select {
	case err := <-strm.Open():
		if err != nil {
			t.Fatal(err)
		}
		val := snk.Get()[0].(int)
		if val != actual {
			t.Fatal("expecting ", actual, "got", val)
		}
	case <-time.After(10 * time.Millisecond):
		t.Fatal("Took too long")
	}
}
