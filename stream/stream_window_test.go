package stream

import (
	"cmp"
	"context"
	"slices"
	"sync/atomic"
	"testing"
	"time"

	"github.com/vladimirvivien/automi/operators/exec"
	"github.com/vladimirvivien/automi/operators/window"
	"github.com/vladimirvivien/automi/sinks"
	"github.com/vladimirvivien/automi/sources"
)

func TestStreamWindow_GroupByIndex(t *testing.T) {
	src := sources.Slice([][]string{
		{"request", "/i/aa", "00:11:51:AA", "accepted"},
		{"request", "/i/aa", "00:11:51:AA", "accepted"},
		{"request", "/i/b", "00:BB:22:DD", "accepted"},
		{"request", "/i/aa", "00:11:51:AA", "accepted"},
		{"request", "/i/aa", "00:11:51:AA", "accepted"},
		{"request", "/i/aa", "00:11:51:AA", "accepted"},
		{"request", "/i/aaaa", "00:11:51:AB", "accepted"},

		{"request", "/i/aa", "00:12:51:AA", "accepted"},
		{"request", "/i/aa", "00:12:51:AA", "accepted"},
		{"request", "/i/b", "00:AA:22:DD", "accepted"},
		{"request", "/i/aa", "00:12:51:AA", "accepted"},
		{"request", "/i/aa", "00:12:51:AA", "accepted"},
		{"request", "/i/aa", "00:12:51:AA", "accepted"},
		{"request", "/i/aaaa", "00:12:51:AB", "accepted"},
		{"request", "/i/b", "00:AA:22:DD", "accepted"},
	})

	strm := From(src).Run(
		// Window(with incoming type T), will output []T
		window.Batch[[]string](),
		// GroupByIndex receives []T, output map[T][][]T
		exec.GroupByIndex[[][]string](2),
	)
	// setup sink to receive map[any][]T
	var count atomic.Int32
	strm.Into(sinks.Func(func(data map[string][][]string) error {
		for range data {
			count.Add(1)
		}
		return nil
	}))

	select {
	case err := <-strm.Open(context.Background()):
		if err != nil {
			t.Fatal(err)
		}
	case <-time.After(10 * time.Millisecond):
		t.Fatal("Took too long")
	}

	// count group recevied
	if count.Load() != 6 {
		t.Fatalf("Unexpected group count %d", count.Load())
	}
}

func TestStreamWindow_GroupByStructField(t *testing.T) {
	var count atomic.Int32
	type log struct{ Event, Src, Device, Result string }
	src := sources.Slice([]log{
		{Event: "request", Src: "/i/a", Device: "00:11:51:AA", Result: "accepted"},
		{Event: "response", Src: "/i/a/", Device: "00:11:51:AA", Result: "served"},
		{Event: "request", Src: "/i/b", Device: "00:11:22:33", Result: "accepted"},
		{Event: "response", Src: "/i/b", Device: "00:11:22:33", Result: "served"},
		{Event: "request", Src: "/i/c", Device: "00:11:51:AA", Result: "accepted"},
		{Event: "response", Src: "/i/c", Device: "00:11:51:AA", Result: "served"},
		{Event: "request", Src: "/i/d", Device: "00:BB:22:DD", Result: "accepted"},
		{Event: "response", Src: "/i/d", Device: "00:BB:22:DD", Result: "served"},
	})

	strm := From(src).Run(
		// Window(with incoming type T), will output []T
		window.Batch[log](),
		// GroupByStructField receives []T, output map[any][]T
		exec.GroupByStructField[[]log]("Device"),
	)
	// setup sink to receive map[any][]T
	strm.Into(sinks.Func(func(data map[any][]log) error {
		for range data {
			count.Add(1)
		}
		return nil
	}))

	select {
	case err := <-strm.Open(context.Background()):
		if err != nil {
			t.Fatal(err)
		}
	case <-time.After(10 * time.Millisecond):
		t.Fatal("Took too long")
	}

	// count group recevied
	if count.Load() != 3 {
		t.Fatalf("Unexpected group count %d", count.Load())
	}
}

func TestStreamWindow_GroupByMapKey(t *testing.T) {
	var count atomic.Int32
	src := sources.Slice([]map[string]string{
		{"Event": "request", "Src": "/i/a", "Device": "00:11:51:AA", "Result": "accepted"},
		{"Event": "response", "Src": "/i/a/", "Device": "00:11:51:AA", "Result": "served"},
		{"Event": "request", "Src": "/i/b", "Device": "00:11:22:33", "Result": "accepted"},
		{"Event": "response", "Src": "/i/b", "Device": "00:11:22:33", "Result": "served"},
		{"Event": "request", "Src": "/i/c", "Device": "00:11:51:AA", "Result": "accepted"},
		{"Event": "response", "Src": "/i/c", "Device": "00:11:51:AA", "Result": "served"},
		{"Event": "request", "Src": "/i/d", "Device": "00:BB:22:DD", "Result": "accepted"},
		{"Event": "response", "Src": "/i/d", "Device": "00:BB:22:DD", "Result": "served"},
	})

	strm := From(src).Run(
		// Window(with incoming type T), will output []T
		window.BySize[map[string]string](4),
		// GroupByMap receives []T, output map[any][]T
		exec.GroupByMapKey[[]map[string]string]("Device"),
	)
	// setup sink to receive map[any][]T
	strm.Into(sinks.Func(func(data map[string][]map[string]string) error {
		for range data {
			count.Add(1)
		}
		return nil
	}))

	select {
	case err := <-strm.Open(context.Background()):
		if err != nil {
			t.Fatal(err)
		}
	case <-time.After(10 * time.Millisecond):
		t.Fatal("Took too long")
	}

	// count group recevied
	if count.Load() != 4 {
		t.Fatalf("Unexpected group count %d", count.Load())
	}
}

func TestStreamWindow_SumByIndex(t *testing.T) {
	src := sources.Slice([][]int{
		{4879, 12104, 50724, 116464, 12742},
		{1, -1, 0, 1, 0},
		{79, 104, 724, 4, 2},
	})

	strm := From(src).Run(
		// Window(with incoming type T), will output []T
		window.Batch[[]int](),
		// SumByIndex receives []T, outputs sum of values float64
		exec.SumByIndex[[][]int](2),
	)
	// setup sink to collect sums
	var sum atomic.Int64
	strm.Into(sinks.Func(func(data float64) error {
		sum.Add(int64(data))
		return nil
	}))

	select {
	case err := <-strm.Open(context.Background()):
		if err != nil {
			t.Fatal(err)
		}
	case <-time.After(10 * time.Millisecond):
		t.Fatal("Took too long")
	}
	if sum.Load() != 51448 {
		t.Fatalf("Unexpected sum: %d", sum.Load())
	}
}

func TestStreamWindow_SumByStructField(t *testing.T) {
	type vehicle struct {
		Vehicle, Kind, Engine string
		Size                  int
	}
	src := sources.Slice([]vehicle{
		{"Spirit", "plane", "propeller", 12},
		{"Voyager", "satellite", "gravitational", 8},
		{"BigFoot", "truck", "diesel", 8},
		{"Enola", "plane", "propeller", 12},
		{"Memphis", "plane", "propeller", 48},
	})

	strm := From(src).Run(
		// Window(with incoming type T), will output []T
		window.Batch[vehicle](),
		// SumByStructField receives []T, outputs sum of values float64
		exec.SumByStructField[[]vehicle]("Size"),
	)

	// setup sink to collect sums
	var sum atomic.Int64
	strm.Into(sinks.Func(func(data float64) error {
		sum.Add(int64(data))
		return nil
	}))

	select {
	case err := <-strm.Open(context.Background()):
		if err != nil {
			t.Fatal(err)
		}
	case <-time.After(10 * time.Millisecond):
		t.Fatal("Took too long")
	}
	if sum.Load() != 88 {
		t.Fatalf("Unexpected sum: %d", sum.Load())
	}
}

func TestStreamWindow_SumByMapKey(t *testing.T) {
	src := sources.Slice([]map[string]int{
		{"vehicle": 0, "weight": 2},
		{"vehicle": 2, "weight": 4},
		{"vehicle": 3, "weight": 2},
		{"vehicle": 1, "weight": 5},
		{"vehicle": 5},
	})

	strm := From(src).Run(
		// Window(with incoming type T), will output []T
		window.Batch[map[string]int](),
		// SumByStructField receives []T, outputs sum of values float64
		exec.SumByMapKey[[]map[string]int]("weight"),
	)

	// setup sink to collect sums
	var sum atomic.Int64
	strm.Into(sinks.Func(func(data float64) error {
		sum.Add(int64(data))
		return nil
	}))

	select {
	case err := <-strm.Open(context.Background()):
		if err != nil {
			t.Fatal(err)
		}
	case <-time.After(10 * time.Millisecond):
		t.Fatal("Took too long")
	}
	if sum.Load() != 13 {
		t.Fatalf("Unexpected sum: %d", sum.Load())
	}
}

func TestStreamWindow_SumAll1D(t *testing.T) {
	src := sources.Slice([]float32{10.0, 70.0, 20.0, 40.0, 60, 90, 0, 80, 30})

	strm := From(src).Run(
		// Window(with incoming type T), will output []T
		window.Batch[float32](),
		// SumByStructField receives []T, outputs sum of values float64
		exec.SumAll1D[[]float32](),
	)

	// setup sink to collect sums
	var sum atomic.Int64
	strm.Into(sinks.Func(func(data float64) error {
		sum.Add(int64(data))
		return nil
	}))

	select {
	case err := <-strm.Open(context.Background()):
		if err != nil {
			t.Fatal(err)
		}
	case <-time.After(10 * time.Millisecond):
		t.Fatal("Took too long")
	}
	if sum.Load() != 400 {
		t.Fatalf("Unexpected sum: %d", sum.Load())
	}
}

func TestStreamWindow_SumAll2D(t *testing.T) {
	src := sources.Slice([][]int{
		{10, 70, 20},
		{40, 60, 90},
		{0, 80, 30},
	})

	strm := From(src).Run(
		// Window(with incoming type T), will output []T
		window.Batch[[]int](),
		// SumByStructField receives []T, outputs sum of values float64
		exec.SumAll2D[[][]int](),
	)

	// setup sink to collect sums
	var sum atomic.Int64
	strm.Into(sinks.Func(func(data float64) error {
		sum.Add(int64(data))
		return nil
	}))

	select {
	case err := <-strm.Open(context.Background()):
		if err != nil {
			t.Fatal(err)
		}
	case <-time.After(10 * time.Millisecond):
		t.Fatal("Took too long")
	}
	if sum.Load() != 400 {
		t.Fatalf("Unexpected sum: %d", sum.Load())
	}
}

func TestStreamWindow_SortSlice(t *testing.T) {
	src := sources.Slice([]string{"Spirit", "Voyager", "BigFoot", "Enola", "Memphis"})

	strm := From(src).Run(
		// Window(with incoming type T), will output []T
		window.Batch[string](),
		// SortSlice receives []T, outputs sum of values float64
		exec.SortSlice[[]string](),
	)

	// setup sink to collect data
	strm.Into(sinks.Func(func(sorted []string) error {
		if !slices.IsSorted[[]string](sorted) {
			t.Fatal("Data is not sorted: ", sorted)
		}
		return nil
	}))

	select {
	case err := <-strm.Open(context.Background()):
		if err != nil {
			t.Fatal(err)
		}
	case <-time.After(10 * time.Millisecond):
		t.Fatal("Took too long")
	}
}

func TestStreamWindow_SortSliceByIndex(t *testing.T) {
	src := sources.Slice([][]string{
		{"Spirit", "plane", "propeller"},
		{"Voyager", "satellite", "gravitational"},
		{"BigFoot", "truck", "diesel"},
		{"Enola", "plane", "propeller"},
		{"Memphis", "plane", "propeller"},
	})

	strm := From(src).Run(
		// Window(with incoming type T), will output []T
		window.Batch[[]string](),
		// SortSlice receives []T, outputs sum of values float64
		exec.SortSliceByIndex[[][]string](0),
	)

	// setup sink to collect data
	strm.Into(sinks.Func(func(sorted [][]string) error {
		var col []string
		for _, row := range sorted {
			col = append(col, row[0])
		}
		if !slices.IsSorted[[]string](col) {
			t.Fatal("Data is not sorted: ", col)
		}
		return nil
	}))

	select {
	case err := <-strm.Open(context.Background()):
		if err != nil {
			t.Fatal(err)
		}
	case <-time.After(10 * time.Millisecond):
		t.Fatal("Took too long")
	}
}

func TestStreamWindow_SortSliceByStructField(t *testing.T) {
	type vehicle struct {
		Vehicle, Kind, Engine string
		Size                  int
	}

	src := sources.Slice([]vehicle{
		{"Spirit", "plane", "propeller", 12},
		{"Voyager", "satellite", "gravitational", 8},
		{"BigFoot", "truck", "diesel", 8},
		{"Enola", "plane", "propeller", 12},
		{"Memphis", "plane", "propeller", 48},
	})

	strm := From(src).Run(
		// Window(with incoming type T), will output []T
		window.Batch[vehicle](),
		// SortSlice receives []T, outputs sum of values float64
		exec.SortByStructField[[]vehicle]("Vehicle"),
	)

	// setup sink to collect data
	strm.Into(sinks.Func(func(sorted []vehicle) error {
		var col []string
		for _, row := range sorted {
			col = append(col, row.Vehicle)
		}
		if !slices.IsSorted[[]string](col) {
			t.Fatal("Data is not sorted: ", col)
		}
		return nil
	}))

	select {
	case err := <-strm.Open(context.Background()):
		if err != nil {
			t.Fatal(err)
		}
	case <-time.After(10 * time.Millisecond):
		t.Fatal("Took too long")
	}
}

func TestStreamWindow_SortByMapKey(t *testing.T) {
	src := sources.Slice([]map[string]string{
		{"Vehicle": "Spirit", "Kind": "plane", "Engine": "propeller"},
		{"Vehicle": "Voyager", "Kind": "satellite", "Engine": "gravitational"},
		{"Vehicle": "BigFoot", "Kind": "truck", "Engine": "diesel"},
		{"Vehicle": "Enola", "Kind": "plane", "Engine": "propeller"},
		{"Vehicle": "Memphis", "Kind": "plane", "Engine": "propeller"},
	})

	strm := From(src).Run(
		// Window(with incoming type T), will output []T
		window.Batch[map[string]string](),
		// SortSlice receives []T, outputs sum of values float64
		exec.SortByMapKey[[]map[string]string]("Vehicle"),
	)

	// setup sink to collect data
	strm.Into(sinks.Func(func(sorted []map[string]string) error {
		var col []string
		for _, row := range sorted {
			col = append(col, row["Vehicle"])
		}
		if !slices.IsSorted[[]string](col) {
			t.Fatal("Data is not sorted: ", col)
		}
		return nil
	}))

	select {
	case err := <-strm.Open(context.Background()):
		if err != nil {
			t.Fatal(err)
		}
	case <-time.After(10 * time.Millisecond):
		t.Fatal("Took too long")
	}
}

func TestStreamWindow_SortWithFunc(t *testing.T) {
	src := sources.Slice([]string{
		"Spririt",
		"Voyager",
		"BigFoot",
		"Enola",
		"Memphis",
	})

	strm := From(src).Run(
		// Window(with incoming type T), will output []T
		window.Batch[string](),
		// SortSlice receives []T, outputs sum of values float64
		exec.SortWithFunc[[]string](func(i, j string) int {
			return cmp.Compare(i, j)
		}),
	)

	// setup sink to collect data
	strm.Into(sinks.Func(func(sorted []string) error {
		if !slices.IsSorted[[]string](sorted) {
			t.Fatal("Data is not sorted: ", sorted)
		}
		return nil
	}))

	select {
	case err := <-strm.Open(context.Background()):
		if err != nil {
			t.Fatal(err)
		}
	case <-time.After(10 * time.Millisecond):
		t.Fatal("Took too long")
	}
}
