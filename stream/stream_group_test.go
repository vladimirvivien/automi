package stream

import (
	"sync"
	"testing"
	"time"

	"github.com/vladimirvivien/automi/emitters"
)

func TestStream_GroupByKey(t *testing.T) {
	type log map[string]string
	src := emitters.Slice([]map[string]string{
		{"Event": "request", "Src": "/i/a", "Device": "00:11:51:AA", "Result": "accepted"},
		{"Event": "response", "Src": "/i/a/", "Device": "00:11:51:AA", "Result": "served"},
		{"Event": "request", "Src": "/i/b", "Device": "00:11:22:33", "Result": "accepted"},
		{"Event": "response", "Src": "/i/b", "Device": "00:11:22:33", "Result": "served"},
		{"Event": "request", "Src": "/i/c", "Device": "00:11:51:AA", "Result": "accepted"},
		{"Event": "response", "Src": "/i/c", "Device": "00:11:51:AA", "Result": "served"},
		{"Event": "request", "Src": "/i/d", "Device": "00:BB:22:DD", "Result": "accepted"},
		{"Event": "response", "Src": "/i/d", "Device": "00:BB:22:DD", "Result": "served"},
	})
	snk := NewDrain()
	strm := New(src).Batch().GroupByKey("Device").To(snk)

	wait := make(chan struct{})
	var result map[interface{}][]interface{}
	var m sync.Mutex
	go func() {
		defer close(wait)
		r := <-snk.GetOutput()
		m.Lock()
		result = r.(map[interface{}][]interface{})
		m.Unlock()
	}()

	select {
	case err := <-strm.Open():
		if err != nil {
			t.Fatal(err)
		}
		select {
		case <-wait:
			if len(result) != 3 {
				t.Fatal("Unexpected group size ", len(result))
			}
		case <-time.After(10 * time.Millisecond):
			t.Fatal("Stream took too long to complete")
		}
	case <-time.After(10 * time.Millisecond):
		t.Fatal("Took too long")
	}
}

func TestStream_GroupByName(t *testing.T) {
	type log struct{ Event, Src, Device, Result string }
	src := emitters.Slice([]log{
		log{Event: "request", Src: "/i/a", Device: "00:11:51:AA", Result: "accepted"},
		log{Event: "response", Src: "/i/a/", Device: "00:11:51:AA", Result: "served"},
		log{Event: "request", Src: "/i/b", Device: "00:11:22:33", Result: "accepted"},
		log{Event: "response", Src: "/i/b", Device: "00:11:22:33", Result: "served"},
		log{Event: "request", Src: "/i/c", Device: "00:11:51:AA", Result: "accepted"},
		log{Event: "response", Src: "/i/c", Device: "00:11:51:AA", Result: "served"},
		log{Event: "request", Src: "/i/d", Device: "00:BB:22:DD", Result: "accepted"},
		log{Event: "response", Src: "/i/d", Device: "00:BB:22:DD", Result: "served"},
	})
	snk := NewDrain()
	strm := New(src).Batch().GroupByName("Device").To(snk)

	wait := make(chan struct{})
	var result map[interface{}][]interface{}
	var m sync.Mutex
	go func() {
		defer close(wait)
		r := <-snk.GetOutput()
		m.Lock()
		result = r.(map[interface{}][]interface{})
		m.Unlock()
	}()

	select {
	case err := <-strm.Open():
		if err != nil {
			t.Fatal(err)
		}
		select {
		case <-wait:
			if len(result) != 3 {
				t.Fatal("unexpected group size:", len(result))
			}
		case <-time.After(10 * time.Millisecond):
			t.Fatal("Stream took too long to complete")
		}
	case <-time.After(10 * time.Millisecond):
		t.Fatal("Took too long")
	}
}
