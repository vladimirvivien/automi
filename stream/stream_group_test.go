package stream

import (
	"strings"
	"testing"
	"time"

	"github.com/vladimirvivien/automi/api/tuple"
)

// TestStream_GroupByInt_KV groups incoming tuple.KV data items.
func TestStream_GroupByInt_KV(t *testing.T) {
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
	src := NewSliceSource("Hello World Wide", "Hello Milkyway Away", "Our World Now")
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
