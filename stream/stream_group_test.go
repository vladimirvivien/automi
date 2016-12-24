package stream

import (
	"strings"
	"testing"
	"time"

	"github.com/vladimirvivien/automi/api/tuple"
	"github.com/vladimirvivien/automi/sources"
)

// TestStream_GroupByInt_KV groups incoming tuple.KV data items.
func TestStream_GroupByInt_KV(t *testing.T) {
	src := sources.Slice("Hello World", "Hello Milkyway", "Hello Universe")
	snk := NewDrain()
	strm := New().From(src).FlatMap(func(line string) []string {
		return strings.Split(line, " ")
	}).Map(func(data string) tuple.KV {
		return tuple.KV{data, 1}
	}).GroupBy(0).To(snk)

	wait := make(chan struct{})
	go func() {
		defer close(wait)
		result := <-snk.GetOutput()
		m := result.(map[interface{}][]interface{})
		if len(m["Hello"]) != 3 {
			t.Fatal("Items not grouped properly")
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

// TestStream_GroupByInt_Slice tests  grouping data items received as slice.
func TestStream_GroupByInt_Slice(t *testing.T) {
	src := sources.Slice("Hello World Wide", "Hello Milkyway Away", "Our World Now")
	snk := NewDrain()
	strm := New().From(src).Map(func(item string) []string {
		return strings.Split(item, " ")
	}).GroupBy(1).To(snk)

	wait := make(chan struct{})
	go func() {
		defer close(wait)
		result := <-snk.GetOutput()
		m := result.(map[interface{}][]interface{})
		t.Log(m)
		if len(m["World"]) != 4 {
			t.Fatal("Items not grouped properly")
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

func TestStream_GroupByString_Struct(t *testing.T) {
	type log struct{ Event, Src, Device, Result string }
	src := sources.Slice(
		log{Event: "request", Src: "/i/a", Device: "00:11:51:AA", Result: "accepted"},
		log{Event: "response", Src: "/i/a/", Device: "00:11:51:AA", Result: "served"},
		log{Event: "request", Src: "/i/b", Device: "00:11:22:33", Result: "accepted"},
		log{Event: "response", Src: "/i/b", Device: "00:11:22:33", Result: "served"},
		log{Event: "request", Src: "/i/c", Device: "00:11:51:AA", Result: "accepted"},
		log{Event: "response", Src: "/i/c", Device: "00:11:51:AA", Result: "served"},
		log{Event: "request", Src: "/i/d", Device: "00:BB:22:DD", Result: "accepted"},
		log{Event: "response", Src: "/i/d", Device: "00:BB:22:DD", Result: "served"},
	)
	snk := NewDrain()
	strm := New().From(src).GroupBy("Device").To(snk)

	wait := make(chan struct{})
	go func() {
		defer close(wait)
		result := <-snk.GetOutput()
		m := result.(map[interface{}][]interface{})
		t.Log(m)
		if len(m) != 3 {
			t.Fatal("Items not grouped properly")
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
