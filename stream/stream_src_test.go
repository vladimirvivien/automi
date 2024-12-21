package stream

import (
	"bufio"
	"bytes"
	"context"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/vladimirvivien/automi/sinks"
	"github.com/vladimirvivien/automi/sources"
	"github.com/vladimirvivien/automi/testutil"
	"github.com/vladimirvivien/gexe"
)

func TestStreamFromSliceSource(t *testing.T) {
	src := sources.Slice([][]string{
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
		if count.Load() != 5 {
			t.Fatal("expecting count 5, got", count.Load())
		}
	case <-time.After(10 * time.Millisecond):
		t.Fatal("Took too long")
	}
}

func TestStreamFromChannelSource(t *testing.T) {
	data := [][]string{
		{"request", "/i/a", "00:11:51:AA", "accepted"},
		{"response", "/i/a/", "00:11:51:AA", "served"},
		{"request", "/i/b", "00:11:22:33", "accepted"},
		{"response", "/i/b", "00:11:22:33", "served"},
		{"request", "/i/c", "00:11:51:AA", "accepted"},
	}

	in := make(chan []string)
	go func(source [][]string) {
		for _, slice := range source {
			in <- slice
		}
		close(in)
	}(data)

	var count atomic.Int32
	strm := From(sources.Chan(in))
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
		if count.Load() != 5 {
			t.Fatal("expecting count 5, got", count.Load())
		}
	case <-time.After(10 * time.Millisecond):
		t.Fatal("Took too long")
	}
}

func TestStreamFromReaderSource(t *testing.T) {
	data := `
"request", "/i/a", "00:11:51:AA", "accepted"
"response", "/i/a/", "00:11:51:AA", "served"
"request", "/i/b", "00:11:22:33", "accepted"
"response", "/i/b", "00:11:22:33", "served"
"request", "/i/c", "00:11:51:AA", "accepted"
`

	expected := len(data)
	var count atomic.Int32

	reader := strings.NewReader(data)
	strm := From(sources.Reader(reader))
	strm.WithLogSink(sinks.Func(testutil.LogSinkFunc(t)))
	strm.Into(sinks.Func(func(items []byte) error {
		count.Add(int32(len(items)))
		return nil
	}))

	select {
	case err := <-strm.Open(context.Background()):
		if err != nil {
			t.Fatal(err)
		}
		if int(count.Load()) != expected {
			t.Fatalf("expecting count %d, got %d", expected, count.Load())
		}
	case <-time.After(10 * time.Millisecond):
		t.Fatal("Took too long")
	}
}

func TestStreamFromCSVSource(t *testing.T) {
	data := `
"request", "/i/a", "00:11:51:AA", "accepted"
"response", "/i/a/", "00:11:51:AA", "served"
"request", "/i/b", "00:11:22:33", "accepted"
"response", "/i/b", "00:11:22:33", "served"
"request", "/i/c", "00:11:51:AA", "accepted"
`

	// write to a file
	fileName := filepath.Join(t.TempDir(), t.Name())
	if err := gexe.FileWrite(fileName).String(data).Err(); err != nil {
		t.Fatal(err)
	}

	var count atomic.Int32
	csv := sources.CSV(bytes.NewReader(gexe.FileRead(fileName).Bytes()))
	strm := From(csv)
	strm.WithLogSink(sinks.Func(testutil.LogSinkFunc(t)))
	strm.Into(sinks.Func(func(items []string) error {
		count.Add(1)
		return nil
	}))

	select {
	case err := <-strm.Open(context.Background()):
		if err != nil {
			t.Fatal(err)
		}
		if int(count.Load()) != 5 {
			t.Fatalf("expecting %d lines, got %d", 5, count.Load())
		}
	case <-time.After(10 * time.Millisecond):
		t.Fatal("Took too long")
	}
}

func TestStreamFromScanner(t *testing.T) {
	data := `"request", "/i/a", "00:11:51:AA", "accepted"
"response", "/i/a/", "00:11:51:AA", "served"
"request", "/i/b", "00:11:22:33", "accepted"
"response", "/i/b", "00:11:22:33", "served"
"request", "/i/c", "00:11:51:AA", "accepted"
`

	var count atomic.Int32
	csv := sources.Scanner(strings.NewReader(data), bufio.ScanLines)
	strm := From(csv)
	strm.WithLogSink(sinks.Func(testutil.LogSinkFunc(t)))
	strm.Into(sinks.Func(func(items []byte) error {
		count.Add(1)
		return nil
	}))

	select {
	case err := <-strm.Open(context.Background()):
		if err != nil {
			t.Fatal(err)
		}
		if int(count.Load()) != 5 {
			t.Fatalf("expecting %d lines, got %d", 5, count.Load())
		}
	case <-time.After(10 * time.Millisecond):
		t.Fatal("Took too long")
	}
}
