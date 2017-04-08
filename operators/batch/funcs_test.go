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
