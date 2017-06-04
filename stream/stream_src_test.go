package stream

import (
	"testing"
	"time"

	"github.com/vladimirvivien/automi/collectors"
	"github.com/vladimirvivien/automi/emitters"
)

func TestStream_WithSource(t *testing.T) {
	src := emitters.Slice([][]string{
		{"request", "/i/a", "00:11:51:AA", "accepted"},
		{"response", "/i/a/", "00:11:51:AA", "served"},
		{"request", "/i/b", "00:11:22:33", "accepted"},
		{"response", "/i/b", "00:11:22:33", "served"},
		{"request", "/i/c", "00:11:51:AA", "accepted"},
	})

	count := 0
	strm := New(src).SinkTo(collectors.Func(func(val interface{}) error {
		count++
		return nil
	}))
	select {
	case err := <-strm.Open():
		if err != nil {
			t.Fatal(err)
		}
		if count != 5 {
			t.Fatal("expecting count 5, got", count)
		}
	case <-time.After(10 * time.Millisecond):
		t.Fatal("Took too long")
	}
}
