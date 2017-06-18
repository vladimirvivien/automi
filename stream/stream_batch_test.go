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
		val := snk.Get()[0].([]map[interface{}][]interface{})
		if len(val[0]) != 3 {
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
		result := snk.Get()[0].([]map[interface{}][]interface{})
		if len(result[0]) != 3 {
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
		result := snk.Get()[0].([]map[interface{}][]interface{})
		if len(result[0]) != 2 {
			t.Fatal("unexpected group size:", len(result))
		}
	case <-time.After(10 * time.Millisecond):
		t.Fatal("Took too long")
	}
}

func TestStream_SortByKey(t *testing.T) {
	src := emitters.Slice([]map[string]string{
		{"Name": "Mercury", "Diameter": "4879"},
		{"Name": "Venus", "Diameter": "12104"},
		{"Name": "Uranus", "Diameter": "50724"},
		{"Name": "Saturn", "Diameter": "116464"},
		{"Name": "Earth", "Diameter": "12742"},
	})

	snk := collectors.Slice()
	strm := New(src).Batch().SortByKey("Name").SinkTo(snk)

	select {
	case err := <-strm.Open():
		if err != nil {
			t.Fatal(err)
		}
		result := snk.Get()[0].([]map[string]string)
		if result[0]["Name"] != "Earth" && result[1]["Name"] != "Mercury" && result[2]["Name"] != "Saturn" {
			t.Fatal("unexpected sort order", result)
		}
	case <-time.After(10 * time.Millisecond):
		t.Fatal("Took too long")
	}
}

func TestStream_SortByName(t *testing.T) {
	src := emitters.Slice([]struct {
		Name     string
		Diameter int
	}{
		{Name: "Mercury", Diameter: 4879},
		{Name: "Venus", Diameter: 12104},
		{Name: "Uranus", Diameter: 50724},
		{Name: "Saturn", Diameter: 116464},
		{Name: "Earth", Diameter: 12742},
	})

	snk := collectors.Slice()
	strm := New(src).Batch().SortByName("Name").SinkTo(snk)

	select {
	case err := <-strm.Open():
		if err != nil {
			t.Fatal(err)
		}
		result := snk.Get()[0].([]struct {
			Name     string
			Diameter int
		})
		if result[0].Name != "Earth" && result[1].Name != "Mercury" && result[2].Name != "Saturn" {
			t.Fatal("unexpected sort order", result)
		}
	case <-time.After(10 * time.Millisecond):
		t.Fatal("Took too long")
	}
}

func TestStream_SortByPos(t *testing.T) {
	src := emitters.Slice([][]string{
		{"Mercury", "4879"},
		{"Venus", "12104"},
		{"Uranus", "50724"},
		{"Saturn", "116464"},
		{"Earth", "12742"},
	})

	snk := collectors.Slice()
	strm := New(src).Batch().SortByPos(0).SinkTo(snk)

	select {
	case err := <-strm.Open():
		if err != nil {
			t.Fatal(err)
		}
		result := snk.Get()[0].([][]string)
		if result[0][0] != "Earth" && result[1][0] != "Mercury" && result[2][0] != "Saturn" {
			t.Fatal("unexpected sort order", result)
		}
	case <-time.After(10 * time.Millisecond):
		t.Fatal("Took too long")
	}
}

func TestStream_SortWith(t *testing.T) {
	src := emitters.Slice([]string{
		"Mercury",
		"Venus",
		"Uranus",
		"Saturn",
		"Earth",
	})

	snk := collectors.Slice()
	strm := New(src).Batch().SortWith(func(batch interface{}, i, j int) bool {
		items := batch.([]string)
		return items[i] < items[j]
	})
	strm = strm.SinkTo(snk)

	select {
	case err := <-strm.Open():
		if err != nil {
			t.Fatal(err)
		}
		result := snk.Get()[0].([]string)
		if result[0] != "Earth" && result[1] != "Mercury" && result[2] != "Saturn" {
			t.Fatal("unexpected sort order", result)
		}
	case <-time.After(10 * time.Millisecond):
		t.Fatal("Took too long")
	}
}

func TestStream_Sum(t *testing.T) {
	src := emitters.Slice([]int{
		4879,
		12104,
		50724,
		116464,
		12742,
	})

	snk := collectors.Slice()
	strm := New(src).Batch().Sum().SinkTo(snk)

	select {
	case err := <-strm.Open():
		if err != nil {
			t.Fatal(err)
		}
		result := snk.Get()[0].(float64)
		if result <= 116464 {
			t.Fatal("unexpected result:", result)
		}
	case <-time.After(10 * time.Millisecond):
		t.Fatal("Took too long")
	}
}

func TestStream_SumByKey(t *testing.T) {
	src := emitters.Slice([]map[string]int{
		{"Diameter": 4879},
		{"Diameter": 12104},
		{"Diameter": 50724},
		{"Diameter": 116464},
		{"Diameter": 12742},
	})

	snk := collectors.Slice()
	strm := New(src).Batch().SumByKey("Diameter").SinkTo(snk)

	select {
	case err := <-strm.Open():
		if err != nil {
			t.Fatal(err)
		}
		result := snk.Get()[0].([]map[interface{}]float64)
		t.Log("Sum calculated:", result[0]["Diameter"])
		if result[0]["Diameter"] <= 116464 {
			t.Fatal("unexpected result:", result[0])
		}
	case <-time.After(10 * time.Millisecond):
		t.Fatal("Took too long")
	}
}

func TestStream_SumByName(t *testing.T) {
	src := emitters.Slice([]struct{ Diam int }{
		{Diam: 4879},
		{Diam: 12104},
		{Diam: 50724},
		{Diam: 116464},
		{Diam: 12742},
	})

	snk := collectors.Slice()
	strm := New(src).Batch().SumByName("Diam").SinkTo(snk)

	select {
	case err := <-strm.Open():
		if err != nil {
			t.Fatal(err)
		}
		result := snk.Get()[0].([]map[string]float64)
		t.Log("Sum calculated:", result[0])
		if result[0]["Diam"] <= 116464 {
			t.Fatal("unexpected result:", result[0])
		}
	case <-time.After(10 * time.Millisecond):
		t.Fatal("Took too long")
	}
}

func TestStream_SumByPos(t *testing.T) {
	src := emitters.Slice([][]int{
		{4879, 12104, 50724, 116464, 12742},
		{1, -1, 0, 1, 0},
	})

	snk := collectors.Slice()
	strm := New(src).Batch().SumByPos(3).SinkTo(snk)

	select {
	case err := <-strm.Open():
		if err != nil {
			t.Fatal(err)
		}
		result := snk.Get()[0].([]map[int]float64)
		if result[0][3] <= 116464 {
			t.Fatal("unexpected result:", result[0])
		}
	case <-time.After(10 * time.Millisecond):
		t.Fatal("Took too long")
	}
}
