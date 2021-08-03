package batch

import (
	"context"
	"testing"
)

func TestBatchFuncs_GroupByPos_WithSlice(t *testing.T) {
	op := GroupByPosFunc(0)
	data := [][]string{
		{"aa", "absolute", "resolute"},
		{"ab", "merchant", "errand", "elegant"},
		{"aa", "classic", "magic", "toxic"},
	}
	result := op.Apply(context.TODO(), data)
	mapVal, ok := result.([]map[interface{}][]interface{})
	if !ok {
		t.Fatal("unexpected type from GroupByFunc")
	}

	if len(mapVal[0]) != 2 {
		t.Fatal("expecting map of size 2, but got", len(mapVal))
	}
	aaVal := mapVal[0]["aa"]
	if len(aaVal) != 5 {
		t.Fatal("grouping failed, expected key 'aa' with 5 items, got", len(aaVal))
	}
}

func TestBatchFuncs_GroupByPos_WithArray(t *testing.T) {
	op := GroupByPosFunc(0)
	data := [][4]string{
		{"aa", "absolute", "resolute", "cantolope"},
		{"ab", "merchant", "errand", "elegant"},
		{"aa", "classic", "magic", "toxic"},
	}
	result := op.Apply(context.TODO(), data)
	mapVal, ok := result.([]map[interface{}][]interface{})
	if !ok {
		t.Fatal("unexpected type from GroupByFunc")
	}

	if len(mapVal[0]) != 2 {
		t.Fatal("expecting map of size 2, but got", len(mapVal))
	}
	aaVal := mapVal[0]["aa"]
	if len(aaVal) != 6 {
		t.Fatal("grouping failed, expected key 'aa' with 6 items, got", len(aaVal))
	}
}

func TestBatchFuncs_GroupByName(t *testing.T) {
	op := GroupByNameFunc("Kind")
	data := []struct{ Vehicle, Kind, Engine string }{
		{"Spirit", "plane", "propeller"},
		{"Voyager", "satellite", "gravitational"},
		{"BigFoot", "truck", "diesel"},
		{"Enola", "plane", "propeller"},
		{"Memphis", "plane", "propeller"},
	}
	val := op.Apply(context.TODO(), data)
	group := val.([]map[interface{}][]interface{})
	planes := group[0]["plane"]
	if len(planes) != 3 {
		t.Fatal("expecting group to have 3 planes, got ", len(planes))
	}

	if len(group[0]["truck"]) != 1 {
		t.Fatal("expecting group to have 1 truck, got ", len(group[0]["truck"]))
	}

	// invalid field name
	op = GroupByNameFunc("method")
	val = op.Apply(context.TODO(), data)
	if len(val.([]map[interface{}][]interface{})[0]) != 0 {
		t.Fatal("expecting a group of zero elements, but the result map has elements")
	}

}

func TestBatchFuncs_GroupByKey(t *testing.T) {
	op := GroupByKeyFunc("kind")
	data := []map[string]string{
		{"vehicle": "spirit", "kind": "plane", "engine": "props"},
		{"vehicle": "santa maria", "kind": "boat", "engine": "sail"},
		{"vehicle": "enola", "kind": "plane", "engine": "props"},
		{"vehicle": "voyager1", "kind": "satellite", "engine": "gravity"},
		{"vehicle": "titanic", "kind": "boat", "engine": "diesel"},
	}
	val := op.Apply(context.TODO(), data)
	group := val.([]map[interface{}][]interface{})
	planes := group[0]["plane"]
	if len(planes) != 2 {
		t.Fatal("expecting group to have 2 planes, got ", len(planes))
	}

	if len(group[0]["boat"]) != 2 {
		t.Fatal("expecting group to have 1 truck, got ", len(group[0]["boat"]))
	}

	// invalid field name
	op = GroupByNameFunc("type")
	val = op.Apply(context.TODO(), data)
	if len(val.([]map[interface{}][]interface{})[0]) != 0 {
		t.Fatal("expecting a group of zero elements, but the result map has elements")
	}
}

func TestBatchFuncs_SumInts(t *testing.T) {
	op := SumFunc()
	data := [][]int{
		{10, 70, 20},
		{40, 60, 90},
		{0, 80, 30},
	}
	result := op.Apply(context.TODO(), data)
	if result.(float64) != 400 {
		t.Error("expecting 400, but got ", result)
	}

	data2 := []float32{10.0, 70.0, 20.0, 40.0, 60, 90, 0, 80, 30}

	result = op.Apply(context.TODO(), data2)
	if result.(float64) != 400 {
		t.Error("expecting 400, but got ", result)
	}
}

func TestBatchFuncs_SumByPos(t *testing.T) {
	op := SumByPosFunc(2)
	data := [][]interface{}{
		{"AA", "B", 4},
		{"BB", "A", 2.0},
		{"CA", "D", 4},
	}

	result := op.Apply(context.TODO(), data)
	val := result.([]map[int]float64)

	if val[0][2] != 10 {
		t.Error("expecting 10, got ", result)
	}
}

func TestBatchFuncs_SumByName(t *testing.T) {
	op := SumByNameFunc("Size")
	data := []struct {
		Vehicle, Kind, Engine string
		Size                  int
	}{
		{"Spirit", "plane", "propeller", 12},
		{"Voyager", "satellite", "gravitational", 8},
		{"BigFoot", "truck", "diesel", 8},
		{"Enola", "plane", "propeller", 12},
		{"Memphis", "plane", "propeller", 48},
	}
	val := op.Apply(context.TODO(), data)
	result := val.([]map[string]float64)

	if result[0]["Size"] != 88 {
		t.Error("expecting sum of 88, got ", val)
	}
}

func TestBatchFuncs_SumByName_All(t *testing.T) {
	op := SumByNameFunc("")
	data := []interface{}{
		struct {
			Vehicle string
			Engines int
			Sizes   []int
		}{"Spirit", 2, []int{4, 2, 1}},
		struct {
			Vehicle string
			Engines int
			Sizes   []int
		}{"Voyager", 1, []int{1}},
		struct {
			Vehicle string
			Engines int
			Sizes   []int
		}{"Memphis", 8, []int{2, 4}},
	}
	val := op.Apply(context.TODO(), data)
	result := val.([]map[string]float64)

	if result[0]["Sizes"] != 14 {
		t.Error("expecting sum of 14, got ", result)
	}

	if result[0]["Engines"] != 11 {
		t.Error("expecting sum of 11, got ", result)
	}

}

func TestBatchFuncs_SumByKey(t *testing.T) {
	op := SumByKeyFunc("weight")
	data := []map[string]int{
		{"vehicle": 0, "weight": 2},
		{"vehicle": 2, "weight": 4},
		{"vehicle": 3, "weight": 2},
		{"vehicle": 1, "weight": 5},
		{"vehicle": 5},
	}
	val := op.Apply(context.TODO(), data)
	result := val.([]map[interface{}]float64)

	if result[0]["weight"] != 13 {
		t.Error("expecting sum 13, got ", val)
	}
}

func TestBatchFuncs_SumByKey_All(t *testing.T) {
	op := SumByKeyFunc(nil)
	data := []map[interface{}]interface{}{
		{"vehicle": 0, "weight": 2},
		{"vehicle": 2, "weight": 4},
		{"vehicle": 3, "weight": 2},
		{"vehicle": 1, "weight": 5},
		{"vehicle": []int{5, 5, 2}},
	}
	val := op.Apply(context.TODO(), data)
	result := val.([]map[interface{}]float64)

	if result[0]["vehicle"] != 18 {
		t.Error("expecting sum 18, got ", result[0]["vehicle"])
	}
	if result[0]["weight"] != 13 {
		t.Error("expecting sum 13, got ", result[0]["weight"])
	}
}

func TestBatchFuncs_Sort(t *testing.T) {
	op := SortFunc()
	data := []string{"Spirit", "Voyager", "BigFoot", "Enola", "Memphis"}
	val := op.Apply(context.TODO(), data)

	sorted := val.([]string)

	if sorted[0] != "BigFoot" && sorted[1] != "Enola" && sorted[2] != "Memphis" {
		t.Fatal("unexpected sort order for result: ", sorted)
	}
}

func TestBatchFuncs_SortByPos(t *testing.T) {
	op := SortByPosFunc(0)
	data := [][]string{
		{"Spirit", "plane", "propeller"},
		{"Voyager", "satellite", "gravitational"},
		{"BigFoot", "truck", "diesel"},
		{"Enola", "plane", "propeller"},
		{"Memphis", "plane", "propeller"},
	}
	val := op.Apply(context.TODO(), data)

	sorted := val.([][]string)

	if sorted[0][0] != "BigFoot" && sorted[1][0] != "Enola" && sorted[2][0] != "Memphis" {
		t.Fatal("unexpected sort order for result: ", sorted)
	}
}

func TestBatchFuncs_SortByName(t *testing.T) {
	op := SortByNameFunc("Vehicle")
	type V struct {
		Vehicle, Kind, Engine string
		Size                  int
	}
	data := []V{
		{"Spirit", "plane", "propeller", 12},
		{"Voyager", "satellite", "gravitational", 8},
		{"BigFoot", "truck", "diesel", 8},
		{"Enola", "plane", "propeller", 12},
		{"Memphis", "plane", "propeller", 48},
	}
	val := op.Apply(context.TODO(), data)

	sorted := val.([]V)
	if sorted[0].Vehicle != "BigFoot" && sorted[1].Vehicle != "Enola" && sorted[2].Vehicle != "Memphis" {
		t.Fatal("Unexpected sort order")
	}
}

func TestBatchFuncs_SortByKey(t *testing.T) {
	op := SortByKeyFunc("Vehicle")
	data := []map[string]string{
		{"Vehicle": "Spirit", "Kind": "plane", "Engine": "propeller"},
		{"Vehicle": "Voyager", "Kind": "satellite", "Engine": "gravitational"},
		{"Vehicle": "BigFoot", "Kind": "truck", "Engine": "diesel"},
		{"Vehicle": "Enola", "Kind": "plane", "Engine": "propeller"},
		{"Vehicle": "Memphis", "Kind": "plane", "Engine": "propeller"},
	}
	val := op.Apply(context.TODO(), data)

	sorted := val.([]map[string]string)
	if sorted[0]["Vehicle"] != "BigFoot" && sorted[1]["Vehicle"] != "Enola" && sorted[2]["Vehicle"] != "Memphis" {
		t.Fatal("Unexpected sort order")
	}
}

func TestBatchFuncs_SortWithFunc(t *testing.T) {
	op := SortWithFunc(func(batch interface{}, i, j int) bool {
		items := batch.([]string)
		return items[i] < items[j]
	})
	data := []string{
		"Spririt",
		"Voyager",
		"BigFoot",
		"Enola",
		"Memphis",
	}
	val := op.Apply(context.TODO(), data)
	sorted := val.([]string)
	if sorted[0] != "BigFoot" && sorted[1] != "Enola" && sorted[2] != "Membphis" {
		t.Fatal("Unexpected sort order")
	}
}
