package stream

import (
	"bytes"
	"context"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/vladimirvivien/automi/sinks"
	"github.com/vladimirvivien/automi/sources"
	"github.com/vladimirvivien/automi/testutil"
)

func TestStreamToCSVSource(t *testing.T) {
	src := sources.Slice([][]string{
		{"request", "/i/a", "00:11:51:AA", "accepted"},
		{"response", "/i/a/", "00:11:51:AA", "served"},
		{"request", "/i/b", "00:11:22:33", "accepted"},
		{"response", "/i/b", "00:11:22:33", "served"},
		{"request", "/i/c", "00:11:51:AA", "accepted"},
	})

	snk := new(bytes.Buffer)
	strm := From(src)
	strm.WithLogSink(sinks.Func(testutil.LogSinkFunc(t)))
	strm.Into(sinks.CSV(snk))

	select {
	case err := <-strm.Open(context.Background()):
		if err != nil {
			t.Fatal(err)
		}
		lines := strings.Split(strings.TrimSpace(snk.String()), "\n")
		if len(lines) != 5 {
			t.Error("unexpected sink data: want 5 lines, got", len(lines))
		}
	case <-time.After(10 * time.Millisecond):
		t.Fatal("Took too long")
	}
}

func TestStreamToFunc(t *testing.T) {
	src := sources.Slice([][]string{
		{"request", "/i/a", "00:11:51:AA", "accepted"},
		{"response", "/i/a/", "00:11:51:AA", "served"},
		{"request", "/i/b", "00:11:22:33", "accepted"},
		{"response", "/i/b", "00:11:22:33", "served"},
		{"request", "/i/c", "00:11:51:AA", "accepted"},
		{"request", "/i/a", "00:11:51:AA", "accepted"},
		{"response", "/i/a/", "00:11:51:AA", "served"},
		{"request", "/i/b", "00:11:22:33", "accepted"},
		{"response", "/i/b", "00:11:22:33", "served"},
		{"request", "/i/c", "00:11:51:AA", "accepted"},
	})

	var count atomic.Int32
	strm := From(src)
	strm.WithLogSink(sinks.Func(testutil.LogSinkFunc(t)))
	strm.Into(sinks.Func(func(item []string) error {
		count.Add(1)
		return nil
	}))

	select {
	case err := <-strm.Open(context.Background()):
		if err != nil {
			t.Fatal(err)
		}
		if count.Load() != 10 {
			t.Error("unexpected sink data: want 10, got", count.Load())
		}
	case <-time.After(10 * time.Millisecond):
		t.Fatal("Took too long")
	}
}

func TestStreamToDiscard(t *testing.T) {
	src := sources.Slice([][]string{
		{"request", "/i/a", "00:11:51:AA", "accepted"},
		{"response", "/i/a/", "00:11:51:AA", "served"},
		{"request", "/i/b", "00:11:22:33", "accepted"},
		{"response", "/i/b", "00:11:22:33", "served"},
		{"request", "/i/c", "00:11:51:AA", "accepted"},
		{"request", "/i/a", "00:11:51:AA", "accepted"},
		{"response", "/i/a/", "00:11:51:AA", "served"},
		{"request", "/i/b", "00:11:22:33", "accepted"},
		{"response", "/i/b", "00:11:22:33", "served"},
		{"request", "/i/c", "00:11:51:AA", "accepted"},
	})

	strm := From(src)
	strm.WithLogSink(sinks.Func(testutil.LogSinkFunc(t)))
	strm.Into(sinks.Discard())

	select {
	case err := <-strm.Open(context.Background()):
		if err != nil {
			t.Fatal(err)
		}
	case <-time.After(10 * time.Millisecond):
		t.Fatal("Took too long")
	}
}

func TestStreamToSlice(t *testing.T) {
	src := sources.Slice([][]string{
		{"request", "/i/a", "00:11:51:AA", "accepted"},
		{"response", "/i/a/", "00:11:51:AA", "served"},
		{"request", "/i/b", "00:11:22:33", "accepted"},
		{"response", "/i/b", "00:11:22:33", "served"},
		{"request", "/i/c", "00:11:51:AA", "accepted"},
		{"request", "/i/a", "00:11:51:AA", "accepted"},
		{"response", "/i/a/", "00:11:51:AA", "served"},
		{"request", "/i/b", "00:11:22:33", "accepted"},
		{"response", "/i/b", "00:11:22:33", "served"},
		{"request", "/i/c", "00:11:51:AA", "accepted"},
	})

	snk := sinks.Slice[[]string]()
	strm := From(src)
	strm.WithLogSink(sinks.Func(testutil.LogSinkFunc(t)))
	strm.Into(snk)

	select {
	case err := <-strm.Open(context.Background()):
		if err != nil {
			t.Fatal(err)
		}
		data := snk.Get()
		if len(data) != 10 {
			t.Errorf("unexpected sink data: want 10, got %d", len(data))
		}
	case <-time.After(10 * time.Millisecond):
		t.Fatal("Took too long")
	}
}
