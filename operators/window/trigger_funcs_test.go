package window

import (
	"context"
	"testing"
	"time"
)

func TestWindowTriggerAllFunc(t *testing.T) {
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
		wctx := WindowContext[any]{
			Item: test.item,
		}
		if TriggerAllFunc[any]()(context.Background(), wctx) {
			t.Fatalf("window.TriggerAllFunc should always return false")
		}
	}
}

func TestWindowTriggerBySizeFunc(t *testing.T) {
	tests := []struct {
		name        string
		windowCount uint64
		size        uint64
		done        bool
	}{
		{name: "not done; small num", windowCount: 0, size: 10, done: false},
		{name: "not done; large num", windowCount: 100, size: 400000, done: false},
		{name: "size=index", windowCount: 2000, size: 2000, done: true},
		{name: "size<index", windowCount: 2000, size: 10, done: true},
	}

	for _, test := range tests {
		t.Logf("running test %s", test.name)
		wctx := WindowContext[any]{
			WindowItemCount: test.windowCount,
		}
		if TriggerBySizeFunc[any](test.size)(context.Background(), wctx) != test.done {
			t.Fatalf("window.TriggerBySizeFunc was triggered inappropriately")
		}
	}
}

func TestWindowTriggerByDurationFunc(t *testing.T) {
	tests := []struct {
		name         string
		duration     time.Duration
		winStartTime time.Time
		done         bool
	}{
		{
			name:         "done; zero time",
			winStartTime: time.Now(),
			duration:     0,
			done:         true,
		},
		{
			name:         "not done; future start time",
			winStartTime: time.Now().Add(100 * time.Millisecond),
			duration:     200 * time.Millisecond,
			done:         false,
		},
		{
			name:         "done; past start tiem; duration reached/exceeded",
			winStartTime: time.Now().Add(-120 * time.Millisecond),
			duration:     110 * time.Millisecond,
			done:         true,
		},
	}

	for _, test := range tests {
		t.Logf("running test %s", test.name)
		wctx := WindowContext[any]{
			WindowStartTime: test.winStartTime,
		}
		if TriggerByDurationFunc[any](test.duration)(context.Background(), wctx) != test.done {
			t.Fatalf("window.TriggerByDurationFunc was triggered inappropriately")
		}
	}
}

func TestWindowTriggeByFunc(t *testing.T) {
	tests := []struct {
		name   string
		fn     TriggerFunction[any]
		winCtx WindowContext[any]
		done   bool
	}{
		{
			name: "fn always returns true",
			fn: func(ctx context.Context, wctx WindowContext[any]) bool {
				return true
			},
			winCtx: WindowContext[any]{
				Item: "hello",
			},
			done: true,
		},
		{
			name: "fn checks item value",
			fn: func(ctx context.Context, wctx WindowContext[any]) bool {
				return wctx.Item == "hello"
			},
			winCtx: WindowContext[any]{
				Item: "hello",
			},
			done: true,
		},
		{
			name: "fn checks item value; not done",
			fn: func(ctx context.Context, wctx WindowContext[any]) bool {
				return wctx.Item == "hello"
			},
			winCtx: WindowContext[any]{
				Item: "world",
			},
			done: false,
		},
	}

	for _, test := range tests {
		t.Logf("running test %s", test.name)
		if TriggerByFunc(test.fn)(context.Background(), test.winCtx) != test.done {
			t.Fatalf("window.TriggerByFunc should have been done")
		}
	}
}
