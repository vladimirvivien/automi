package batch

import (
	"context"
	"testing"
)

func TestBatch_GroupByPosFunc_WithSlice(t *testing.T) {
	op := GroupByPosFunc(0)
	data := [][]string{
		[]string{"aa", "absolute", "resolute"},
		[]string{"ab", "merchant", "errand", "elegant"},
		[]string{"aa", "classic", "magic", "toxic"},
	}
	result := op.Apply(context.TODO(), data)
	mapVal, ok := result.(map[interface{}][]interface{})
	if !ok {
		t.Fatal("unexpected type from GroupByFunc")
	}

	if len(mapVal) != 2 {
		t.Fatal("expecting map of size 2, but got", len(mapVal))
	}
	aaVal := mapVal["aa"]
	if len(aaVal) != 5 {
		t.Fatal("grouping failed, expected key 'aa' with 5 items, got", len(aaVal))
	}
}

func TestBatch_GroupByPosFunc_WithArray(t *testing.T) {
	op := GroupByPosFunc(0)
	data := [][4]string{
		[4]string{"aa", "absolute", "resolute", "cantolope"},
		[4]string{"ab", "merchant", "errand", "elegant"},
		[4]string{"aa", "classic", "magic", "toxic"},
	}
	result := op.Apply(context.TODO(), data)
	mapVal, ok := result.(map[interface{}][]interface{})
	if !ok {
		t.Fatal("unexpected type from GroupByFunc")
	}

	if len(mapVal) != 2 {
		t.Fatal("expecting map of size 2, but got", len(mapVal))
	}
	aaVal := mapVal["aa"]
	if len(aaVal) != 6 {
		t.Fatal("grouping failed, expected key 'aa' with 6 items, got", len(aaVal))
	}
}

func TestBatch_GroupByNameFunc(t *testing.T) {
	op := GroupByNameFunc("Kind")
	data := []struct{ Vehicle, Kind, Engine string }{
		{"Spirit", "plane", "propeller"},
		{"Voyager", "satellite", "gravitational"},
		{"BigFoot", "truck", "diesel"},
		{"Enola", "plane", "propeller"},
		{"Memphis", "plane", "propeller"},
	}
	val := op.Apply(context.TODO(), data)
	group := val.(map[interface{}][]interface{})
	planes := group["plane"]
	if len(planes) != 3 {
		t.Fatal("expecting group to have 3 planes, got ", len(planes))
	}

	if len(group["truck"]) != 1 {
		t.Fatal("expecting group to have 1 truck, got ", len(group["truck"]))
	}

	// invalid field name
	op = GroupByNameFunc("method")
	val = op.Apply(context.TODO(), data)
	if len(val.(map[interface{}][]interface{})) != 0 {
		t.Fatal("expecting a group of zero elements, but the result map has elements")
	}

}

func TestBatch_GroupByKeyFunc(t *testing.T) {
	op := GroupByKeyFunc("kind")
	data := []map[string]string{
		{"vehicle": "spirit", "kind": "plane", "engine": "props"},
		{"vehicle": "santa maria", "kind": "boat", "engine": "sail"},
		{"vehicle": "enola", "kind": "plane", "engine": "props"},
		{"vehicle": "voyager1", "kind": "satellite", "engine": "gravity"},
		{"vehicle": "titanic", "kind": "boat", "engine": "diesel"},
	}
	val := op.Apply(context.TODO(), data)
	group := val.(map[interface{}][]interface{})
	planes := group["plane"]
	if len(planes) != 2 {
		t.Fatal("expecting group to have 2 planes, got ", len(planes))
	}

	if len(group["boat"]) != 2 {
		t.Fatal("expecting group to have 1 truck, got ", len(group["boat"]))
	}

	// invalid field name
	op = GroupByNameFunc("type")
	val = op.Apply(context.TODO(), data)
	if len(val.(map[interface{}][]interface{})) != 0 {
		t.Fatal("expecting a group of zero elements, but the result map has elements")
	}
}

func TestBatch_SumInts(t *testing.T) {
	op := SumInts()
	data := [][]int{
		{10, 70, 20},
		{40, 60, 90},
		{0, 80, 30},
	}
	result := op.Apply(context.TODO(), data)
	if result.(int64) != 400 {
		t.Error("expecting 400, but got ", result)
	}

	data2 := []int{10, 70, 20, 40, 60, 90, 0, 80, 30}

	result = op.Apply(context.TODO(), data2)
	if result.(int64) != 400 {
		t.Error("expecting 400, but got ", result)
	}
}

func TestBatch_SumIntsByPos(t *testing.T) {
	op := SumIntsByPosFunc(2)
	data := [][]interface{}{
		{"AA", "B", 4},
		{"BB", "A", 2},
		{"CA", "D", 4},
	}

	result := op.Apply(context.TODO(), data)

	if result.(int64) != 10 {
		t.Error("expecting 10, got ", result)
	}
}

func TestBatch_SumFloats(t *testing.T) {
	op := SumFloats()
	data := [][]float32{
		{10.0, 70.0, 20.0},
		{40.0, 60.0, 90.0},
		{0.0, 80.0, 30.0},
	}
	result := op.Apply(context.TODO(), data)
	if result.(float64) != 400.0 {
		t.Error("expecting 400, but got ", result)
	}

	data2 := []float32{10.0, 70.0, 20.0, 40.0, 60.0, 90.0, 0.0, 80.0, 30.0}

	result = op.Apply(context.TODO(), data2)
	if result.(float64) != 400.0 {
		t.Error("expecting 400, but got ", result)
	}
}
