package stream

import (
	"testing"
	"time"

	"github.com/vladimirvivien/automi/sources"
)

func TestStream_GroupByKey(t *testing.T) {
	type log map[string]string
	src := sources.Slice(
		map[string]string{"Event": "request", "Src": "/i/a", "Device": "00:11:51:AA", "Result": "accepted"},
		map[string]string{"Event": "response", "Src": "/i/a/", "Device": "00:11:51:AA", "Result": "served"},
		map[string]string{"Event": "request", "Src": "/i/b", "Device": "00:11:22:33", "Result": "accepted"},
		map[string]string{"Event": "response", "Src": "/i/b", "Device": "00:11:22:33", "Result": "served"},
		map[string]string{"Event": "request", "Src": "/i/c", "Device": "00:11:51:AA", "Result": "accepted"},
		map[string]string{"Event": "response", "Src": "/i/c", "Device": "00:11:51:AA", "Result": "served"},
		map[string]string{"Event": "request", "Src": "/i/d", "Device": "00:BB:22:DD", "Result": "accepted"},
		map[string]string{"Event": "response", "Src": "/i/d", "Device": "00:BB:22:DD", "Result": "served"},
	)
	snk := NewDrain()
	strm := New().From(src).Batch().GroupByKey("Device").To(snk)

	wait := make(chan struct{})
	go func() {
		defer close(wait)
		result := <-snk.GetOutput()
		m := result.(map[interface{}][]interface{})
		t.Logf("%T", m)
		if len(m) != 4 {
			t.Fatal("unexpected item group size ", len(m))
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

func TestStream_GroupByName(t *testing.T) {
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
	strm := New().From(src).Batch().GroupByName("Device").To(snk)

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
