package stream

import (
	"strings"
	"testing"
	"time"

	"github.com/vladimirvivien/automi/collectors"
	"github.com/vladimirvivien/automi/emitters"
)

func TestStream_Process(t *testing.T) {
	src := emitters.Slice([]string{"hello", "world"})
	snk := collectors.Slice()
	strm := New(src).Process(func(s string) string {
		return strings.ToUpper(s)
	}).SinkTo(snk)

	select {
	case err := <-strm.Open():
		if err != nil {
			t.Fatal(err)
		}
		for _, data := range snk.Get() {
			val := data.(string)
			if val != "HELLO" && val != "WORLD" {
				t.Fatalf("got unexpected value %v of type %T", val, val)
			}
		}
	case <-time.After(50 * time.Millisecond):
		t.Fatal("Waited too long ...")
	}
}

func TestStream_Filter(t *testing.T) {
	src := emitters.Slice([]string{"HELLO", "WORLD", "HOW", "ARE", "YOU"})
	snk := collectors.Slice()
	strm := New(src).Filter(func(data string) bool {
		return !strings.Contains(data, "O")
	})
	strm.SinkTo(snk)

	select {
	case err := <-strm.Open():
		if err != nil {
			t.Fatal(err)
		}
	case <-time.After(50 * time.Millisecond):
		t.Fatal("Waited too long ...")
	}
	if len(snk.Get()) != 1 {
		t.Fatal("Filter failed, expected 1 element, got ", len(snk.Get()))
	}
}

func TestStream_Map(t *testing.T) {

	src := emitters.Slice([]string{"HELLO", "WORLD", "HOW", "ARE", "YOU"})
	snk := collectors.Slice()
	strm := New(src).Map(func(data string) int {
		return len(data)
	}).SinkTo(snk)

	count := 0

	select {
	case err := <-strm.Open():
		if err != nil {
			t.Fatal(err)
		}
		for _, data := range snk.Get() {
			val := data.(int)
			count += val
		}
	case <-time.After(50 * time.Millisecond):
		t.Fatal("Waited too long for stream to Open...")
	}

	if count != 19 {
		t.Fatal("Map failed, expected count 19, got ", count)
	}
}

func TestStream_FlatMap(t *testing.T) {
	src := emitters.Slice([]string{"HELLO WORLD", "HOW ARE YOU?"})
	snk := collectors.Slice()
	strm := New(src).FlatMap(func(data string) []string {
		return strings.Split(data, " ")
	}).SinkTo(snk)

	count := 0
	expected := 20

	select {
	case err := <-strm.Open():
		if err != nil {
			t.Fatal(err)
		}

		for _, data := range snk.Get() {
			vals := data.(string)
			count += len(vals)
		}

		if count != expected {
			t.Fatalf("Expecting %d words, got %d", expected, count)
		}
	case <-time.After(50 * time.Millisecond):
		t.Fatal("Waited too long ...")
	}
}
