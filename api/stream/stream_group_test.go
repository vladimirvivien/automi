package stream

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/vladimirvivien/automi/api/tuple"
)

func TestStream_GroupByInt(t *testing.T) {
	src := NewSliceSource("Hello World", "Hello Milkyway", "Hello Universe")
	snk := NewDrain()
	strm := New().From(src).FlatMap(func(line string) []string {
		return strings.Split(line, " ")
	}).Map(func(data string) tuple.KV {
		return tuple.KV{data, 1}
	}).GroupBy(0).To(snk)

	wait := make(chan struct{})
	go func() {
		defer close(wait)
		for result := range snk.GetOutput() {
			t.Log(fmt.Sprintf("%v", result))
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
