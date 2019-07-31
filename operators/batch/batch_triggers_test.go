package batch

import (
	"context"
	"testing"
)

func TestBatchTriggers_All(t *testing.T) {
	tests := []struct {
		name  string
		item  interface{}
		index int64
	}{
		{name: "nil data; no size", item: nil, index: 0},
		{name: "data; no size", item: []string{"hello"}, index: 0},
		{name: "data; size", item: "hello world", index: 2000},
	}

	for _, test := range tests {
		t.Logf("running test %s", test.name)
		if TriggerAll().Done(context.Background(), test.item, test.index) {
			t.Fatalf("batch.TriggerAll.Done should always return false")
		}
	}
}

func TestBatchTriggers_BySize(t *testing.T) {
	tests := []struct {
		name  string
		index int64
		size  int64
		done  bool
	}{
		{name: "not done; small num", index: 0, size: 10, done: false},
		{name: "not done; large num", index: 100, size: 400000, done: false},
		{name: "size=index", index: 2000, size: 2000, done: true},
		{name: "size<index", index: 2000, size: 10, done: true},
	}

	for _, test := range tests {
		t.Logf("running test %s", test.name)
		if TriggerBySize(test.size).Done(context.Background(), nil, test.index) != test.done {
			t.Fatalf("batch.TriggerAll.Done should always return false")
		}
	}
}
