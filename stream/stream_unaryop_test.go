package stream

import (
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/vladimirvivien/automi/sources"
)

func TestStream_Process(t *testing.T) {
	src := sources.Slice("hello", "world")
	snk := NewDrain()
	strm := New().From(src).Process(func(s string) string {
		return strings.ToUpper(s)
	}).To(snk)

	wait := make(chan struct{})
	go func() {
		defer close(wait)
		for data := range snk.GetOutput() {
			val := data.(string)
			if val != "HELLO" && val != "WORLD" {
				t.Fatalf("Got %v of type %T", val, val)
			}
		}
	}()

	select {
	case err := <-strm.Open():
		if err != nil {
			t.Fatal(err)
		}
		select {
		case <-wait:
		case <-time.After(100 * time.Microsecond):
			t.Fatal("Draining took too long")
		}
	case <-time.After(50 * time.Millisecond):
		t.Fatal("Waited too long ...")
	}
}

func TestStream_Filter(t *testing.T) {
	src := newStrSrc([]string{"HELLO", "WORLD", "HOW", "ARE", "YOU"})
	snk := newStrSink()
	strm := New().From(src).Filter(func(data string) bool {
		return !strings.Contains(data, "O")
	})
	strm.To(snk)

	select {
	case err := <-strm.Open():
		if err != nil {
			t.Fatal(err)
		}
	case <-time.After(50 * time.Millisecond):
		t.Fatal("Waited too long ...")
	}
	if len(snk.sink) != 1 {
		t.Fatal("Filter failed, expected 1 element, got ", len(snk.sink))
	}
}

func TestStream_Map(t *testing.T) {

	src := newStrSrc([]string{"HELLO", "WORLD", "HOW", "ARE", "YOU"})
	snk := NewDrain()
	strm := New().From(src).Map(func(data string) int {
		return len(data)
	}).To(snk)

	var m sync.RWMutex
	count := 0
	wait := make(chan struct{})
	go func() {
		defer close(wait)
		for data := range snk.GetOutput() {
			val := data.(int)
			m.Lock()
			count += val
			m.Unlock()
		}
	}()

	select {
	case err := <-strm.Open():
		if err != nil {
			t.Fatal(err)
		}
		select {
		case <-wait:
		case <-time.After(50 * time.Millisecond):
			t.Fatal("Waited too long for sink output to process")
		}

	case <-time.After(50 * time.Millisecond):
		t.Fatal("Waited too long for stream to Open...")
	}

	m.RLock()
	if count != 19 {
		t.Fatal("Map failed, expected count 19, got ", count)
	}
	m.RUnlock()
}

func TestStream_FlatMap(t *testing.T) {
	src := newStrSrc([]string{"HELLO WORLD", "HOW ARE YOU?"})
	snk := NewDrain()
	strm := New().From(src).FlatMap(func(data string) []string {
		return strings.Split(data, " ")
	}).To(snk)

	var m sync.RWMutex
	count := 0
	expected := 20
	wait := make(chan struct{})
	go func() {
		defer close(wait)
		for data := range snk.GetOutput() {
			vals := data.(string)
			m.Lock()
			count += len(vals)
			m.Unlock()
		}
	}()

	select {
	case err := <-strm.Open():
		if err != nil {
			t.Fatal(err)
		}
		select {
		case <-wait:
			if count != expected {
				t.Fatalf("Expecting %d words, got %d", expected, count)
			}
		case <-time.After(50 * time.Millisecond):
			t.Fatal("Took too long to process sink output")
		}
	case <-time.After(50 * time.Millisecond):
		t.Fatal("Waited too long ...")
	}
}
