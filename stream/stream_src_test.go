package stream

import (
	"strings"
	"testing"
	"time"

	"github.com/vladimirvivien/automi/collectors"
	"github.com/vladimirvivien/automi/emitters"
)

func TestStream_SliceSource(t *testing.T) {
	src := emitters.Slice([][]string{
		{"request", "/i/a", "00:11:51:AA", "accepted"},
		{"response", "/i/a/", "00:11:51:AA", "served"},
		{"request", "/i/b", "00:11:22:33", "accepted"},
		{"response", "/i/b", "00:11:22:33", "served"},
		{"request", "/i/c", "00:11:51:AA", "accepted"},
	})

	count := 0
	strm := New(src).Into(collectors.Func(func(val interface{}) error {
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

func TestStream_ChannelSource(t *testing.T) {
	data := [][]string{
		{"request", "/i/a", "00:11:51:AA", "accepted"},
		{"response", "/i/a/", "00:11:51:AA", "served"},
		{"request", "/i/b", "00:11:22:33", "accepted"},
		{"response", "/i/b", "00:11:22:33", "served"},
		{"request", "/i/c", "00:11:51:AA", "accepted"},
	}

	ch := make(chan []string)
	go func(source [][]string) {
		for _, slice := range source {
			ch <- slice
		}
		close(ch)
	}(data)

	count := 0
	strm := New(ch).Into(collectors.Func(func(val interface{}) error {
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

func TestStream_ReaderSource(t *testing.T) {
	data := `"request", "/i/a", "00:11:51:AA", "accepted"
"response", "/i/a/", "00:11:51:AA", "served"
"request", "/i/b", "00:11:22:33", "accepted"
"response", "/i/b", "00:11:22:33", "served"
"request", "/i/c", "00:11:51:AA", "accepted"`

	expected := len(data)
	counted := 0

	reader := strings.NewReader(data)
	strm := New(reader).Into(collectors.Func(func(val interface{}) error {
		counted = len(val.([]byte))
		return nil
	}))
	select {
	case err := <-strm.Open():
		if err != nil {
			t.Fatal(err)
		}
		if counted != expected {
			t.Fatalf("expecting count %d, got %d", expected, counted)
		}
	case <-time.After(10 * time.Millisecond):
		t.Fatal("Took too long")
	}
}
