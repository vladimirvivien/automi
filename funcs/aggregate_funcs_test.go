package funcs

import (
	"cmp"
	"context"
	"fmt"
	"slices"
	"testing"
)

func TestGroupByIndexFunc(t *testing.T) {
	t.Run("test1", func(t *testing.T) {
		op := GroupByIndexFunc[[][]string](0)
		data := [][]string{
			{"aa", "absolute", "resolute"},
			{"ab", "merchant", "errand", "elegant"},
			{"aa", "classic", "magic", "toxic"},
		}
		result := op(context.TODO(), data)
		if len(result["aa"]) != 2 {
			t.Fatalf("unexpected group-by length: %d; %#v", len(result["aa"]), result)
		}
	})

	t.Run("test2", func(t *testing.T) {
		op := GroupByIndexFunc[[][]string](2)
		data := [][]string{
			{"ab", "absolute", "magic"},
			{"ab", "merchant", "errand", "elegant"},
			{"aa", "classic", "magic", "toxic"},
			{"da", "absolute", "resolute"},
			{"ab", "merchant", "errand", "elegant"},
			{"ca", "classic", "magic", "toxic"},
		}
		result := op(context.TODO(), data)
		if len(result["magic"]) != 3 {
			t.Fatal("unexpected group-by length:", len(result["magic"]))
		}
	})
}

func TestBatchFuncs_GroupByName(t *testing.T) {
	data := []struct{ Vehicle, Kind, Engine string }{
		{"Spirit", "plane", "propeller"},
		{"Voyager", "satellite", "gravitational"},
		{"BigFoot", "truck", "diesel"},
		{"Enola", "plane", "propeller"},
		{"Memphis", "plane", "propeller"},
	}

	op := GroupByStructFieldFunc[[]struct{ Vehicle, Kind, Engine string }]("Kind")

	val := op(context.TODO(), data)
	planes := val["plane"]
	fmt.Println(planes)
	if len(planes) != 3 {
		t.Fatal("expecting group to have 3 planes, got ", len(planes))
	}

	if len(val["truck"]) != 1 {
		t.Fatal("expecting group to have 1 truck, got ", len(val["truck"]))
	}

	// invalid field name
	op = GroupByStructFieldFunc[[]struct{ Vehicle, Kind, Engine string }]("method")
	val = op(context.TODO(), data)
	if len(val) != 0 {
		t.Fatal("expecting a group of zero elements, but the result map has elements")
	}

}

func TestGroupByMapKeyFunc(t *testing.T) {
	op := GroupByMapKeyFunc[[]map[string]string]("kind")
	data := []map[string]string{
		{"vehicle": "spirit", "kind": "plane", "engine": "props"},
		{"vehicle": "santa maria", "kind": "boat", "engine": "sail"},
		{"vehicle": "enola", "kind": "plane", "engine": "props"},
		{"vehicle": "voyager1", "kind": "satellite", "engine": "gravity"},
		{"vehicle": "titanic", "kind": "boat", "engine": "diesel"},
	}
	val := op(context.TODO(), data)
	planes := val["plane"]
	if len(planes) != 2 {
		t.Fatal("expecting group to have 2 planes, got ", len(planes))
	}

	boats := val["boat"]
	if len(boats) != 2 {
		t.Fatal("expecting group to have 1 truck, got ", len(boats))
	}

	// invalid field name
	op = GroupByMapKeyFunc[[]map[string]string]("type")
	val = op(context.TODO(), data)
	if len(val) != 0 {
		t.Fatal("expecting a group of zero elements, but the result map has elements")
	}
}

func TestBatchFuncs_SumInts(t *testing.T) {
	op := SumFunc[[][]int, int]()
	data := [][]int{
		{10, 70, 20},
		{40, 60, 90},
		{0, 80, 30},
	}
	result := op(context.TODO(), data)
	if result != 400 {
		t.Error("expecting 400, but got ", result)
	}

	data2 := []float32{10.0, 70.0, 20.0, 40.0, 60, 90, 0, 80, 30}

	op2 := SumFunc[[]float32, float32]()
	result = op2(context.TODO(), data2)
	if result != 400 {
		t.Error("expecting 400, but got ", result)
	}
}

func TestSumByIndexFunc(t *testing.T) {
	op := SumByIndexFunc[[][]any](2)
	data := [][]any{
		{"AA", "B", 4},
		{"BB", "A", 2.0},
		{"CA", "D", 4},
	}

	result := op(context.TODO(), data)
	if result != 10 {
		t.Error("expecting 10, got ", result)
	}
}

func TestSumByStructField(t *testing.T) {
	type vehicle struct {
		Vehicle, Kind, Engine string
		Size                  int
	}
	op := SumByStructFieldFunc[[]vehicle]("Size")
	data := []vehicle{
		{"Spirit", "plane", "propeller", 12},
		{"Voyager", "satellite", "gravitational", 8},
		{"BigFoot", "truck", "diesel", 8},
		{"Enola", "plane", "propeller", 12},
		{"Memphis", "plane", "propeller", 48},
	}
	sum := op(context.TODO(), data)
	if sum != 88 {
		t.Error("expecting sum of 88, got ", sum)
	}
}

func TestBatchFuncs_SumByMapKey(t *testing.T) {
	op := SumByMapKeyFunc[[]map[string]int]("weight")
	data := []map[string]int{
		{"vehicle": 0, "weight": 2},
		{"vehicle": 2, "weight": 4},
		{"vehicle": 3, "weight": 2},
		{"vehicle": 1, "weight": 5},
		{"vehicle": 5},
	}
	sum := op(context.TODO(), data)
	if sum != 13 {
		t.Error("expecting sum 13, got ", sum)
	}
}

func TestSortSlice(t *testing.T) {
	op := SortSliceFunc[[]string]()
	data := []string{"Spirit", "Voyager", "BigFoot", "Enola", "Memphis"}
	sorted := op(context.TODO(), data)
	if !slices.IsSorted[[]string](sorted) {
		t.Fatal("unexpected sort order for result: ", sorted)
	}
}

func TestSortSliceByIndex(t *testing.T) {
	op := SortSliceByIndexFunc[[][]string](0)
	data := [][]string{
		{"Spirit", "plane", "propeller"},
		{"Voyager", "satellite", "gravitational"},
		{"BigFoot", "truck", "diesel"},
		{"Enola", "plane", "propeller"},
		{"Memphis", "plane", "propeller"},
	}
	sorted := op(context.TODO(), data)
	var col []string
	for _, row := range sorted {
		col = append(col, row[0])
	}
	if !slices.IsSorted[[]string](col) {
		t.Fatal("unexpected sort order for result: ", sorted)
	}
}

func TestSortByStructField(t *testing.T) {
	type V struct {
		Vehicle, Kind, Engine string
		Size                  int
	}

	op := SortByStructFieldFunc[[]V]("Vehicle")

	data := []V{
		{"Spirit", "plane", "propeller", 12},
		{"Voyager", "satellite", "gravitational", 8},
		{"BigFoot", "truck", "diesel", 8},
		{"Enola", "plane", "propeller", 12},
		{"Memphis", "plane", "propeller", 48},
	}
	sorted := op(context.TODO(), data)
	fmt.Printf("sorted: %v\n", sorted)
	var col []string
	for _, row := range sorted {
		col = append(col, row.Vehicle)
	}
	if !slices.IsSorted[[]string](col) {
		t.Fatal("unexpected sort order for result: ", sorted)
	}
}

func TestSortByMapKey(t *testing.T) {
	op := SortByMapKeyFunc[[]map[string]string]("Vehicle")
	data := []map[string]string{
		{"Vehicle": "Spirit", "Kind": "plane", "Engine": "propeller"},
		{"Vehicle": "Voyager", "Kind": "satellite", "Engine": "gravitational"},
		{"Vehicle": "BigFoot", "Kind": "truck", "Engine": "diesel"},
		{"Vehicle": "Enola", "Kind": "plane", "Engine": "propeller"},
		{"Vehicle": "Memphis", "Kind": "plane", "Engine": "propeller"},
	}
	sorted := op(context.TODO(), data)
	var col []string
	for _, row := range sorted {
		col = append(col, row["Vehicle"])
	}
	if !slices.IsSorted[[]string](col) {
		t.Fatal("unexpected sort order for result: ", sorted)
	}
}

func TestSortWithFunc(t *testing.T) {
	op := SortWithFuncFunc[[]string](func(i, j string) int {
		return cmp.Compare[string](i, j)
	})
	data := []string{
		"Spririt",
		"Voyager",
		"BigFoot",
		"Enola",
		"Memphis",
	}
	sorted := op(context.TODO(), data)
	if sorted[0] != "BigFoot" && sorted[1] != "Enola" && sorted[2] != "Membphis" {
		t.Fatal("Unexpected sort order")
	}
}
