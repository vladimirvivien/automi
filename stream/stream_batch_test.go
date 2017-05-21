package stream

import (
	"testing"
	"time"

	"github.com/vladimirvivien/automi/collectors"
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
	snk := collectors.Slice()
	strm := New(src).Batch().GroupByKey("Device").SinkTo(snk)

	select {
	case err := <-strm.Open():
		if err != nil {
			t.Fatal(err)
		}
		val := snk.Get()[0].(map[interface{}][]interface{})
		if len(val) != 3 {
			t.Fatal("unxpected group size:", len(val))
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
	snk := collectors.Slice()
	strm := New(src).Batch().GroupByName("Device").SinkTo(snk)

	select {
	case err := <-strm.Open():
		if err != nil {
			t.Fatal(err)
		}
		result := snk.Get()[0].(map[interface{}][]interface{})
		if len(result) != 3 {
			t.Fatal("unexpected group size:", len(result))
		}
	case <-time.After(10 * time.Millisecond):
		t.Fatal("Took too long")
	}
}

func TestStream_GroupByPos(t *testing.T) {
	src := emitters.Slice([][]string{
		{"request", "/i/a", "00:11:51:AA", "accepted"},
		{"response", "/i/a/", "00:11:51:AA", "served"},
		{"request", "/i/b", "00:11:22:33", "accepted"},
		{"response", "/i/b", "00:11:22:33", "served"},
		{"request", "/i/c", "00:11:51:AA", "accepted"},
		{"response", "/i/c", "00:11:51:AA", "served"},
		{"request", "/i/d", "00:BB:22:DD", "accepted"},
		{"response", "/i/d", "00:BB:22:DD", "served"},
	})
	snk := collectors.Slice()

	strm := New(src).Batch().GroupByPos(3).SinkTo(snk)

	select {
	case err := <-strm.Open():
		if err != nil {
			t.Fatal(err)
		}
		result := snk.Get()[0].(map[interface{}][]interface{})
		if len(result) != 2 {
			t.Fatal("unexpected group size:", len(result))
		}
	case <-time.After(10 * time.Millisecond):
		t.Fatal("Took too long")
	}
}

func TestStream_SortByKey(t *testing.T) {
	src := emitters.Slice([]map[string]interface{}{
		{"Name": "Mercury", "Diameter": 4879},
		{"Name": "Venus", "Diameter": 12104},
		{"Name": "Uranus", "Diameter": 50724},
		{"Name": "Saturn", "Diameter": 116464},
		{"Name": "Earth", "Diameter": 12742},
	})

	snk := collectors.Slice()
	strm := New(src).Batch().SortByKey("Name").SinkTo(snk)

	select {
	case err := <-strm.Open():
		if err != nil {
			t.Fatal(err)
		}
		result := snk.Get()[0].([]map[string]interface{})
		if result[0]["Name"] != "Earth" && result[1]["Name"] != "Mercury" && result[2]["Name"] != "Saturn" {
			t.Fatal("unexpected sort order", result)
		}
	case <-time.After(10 * time.Millisecond):
		t.Fatal("Took too long")
	}

}
