package stream

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/vladimirvivien/automi/operators/exec"
	"github.com/vladimirvivien/automi/sinks"
	"github.com/vladimirvivien/automi/sources"
	"github.com/vladimirvivien/automi/testutil"
)

func TestNewStream(t *testing.T) {
	st := From(sources.Slice([]int{1, 2, 3}))
	if st.source == nil {
		t.Fatal("source not getting set")
	}
}

func TestStreamSetup(t *testing.T) {
	st := From(sources.Slice([]string{"Hello", "World", "!!!"})).
		WithLogSink(sinks.Func(testutil.LogSinkFunc(t))).
		Into(sinks.Slice[[]string]())
	if st.source == nil {
		t.Fatal("From() not setting source")
	}
	if st.sink == nil {
		t.Fatal("Into() not setting sink")
	}
	if len(st.nodes) != 0 {
		t.Fatal("operations should not be set")
	}
}

func TestStreamInitGraph(t *testing.T) {
	src := sources.Slice([]string{"Hello", "World"})
	snk := sinks.Slice[[]string]()
	op1 := func(ctx context.Context, in string) string {
		return in
	}
	op2 := func(ctx context.Context, in string) string {
		return in
	}

	strm := From(src).Run(
		exec.Execute(op1),
		exec.Execute(op2),
	)
	strm.WithLogSink(sinks.Func(testutil.LogSinkFunc(t)))
	strm.Into(snk)

	if err := strm.initGraph(context.Background()); err != nil {
		t.Fatal(err)
	}

	if len(strm.nodes) != 2 {
		t.Fatal("Not adding operations to stream")
	}
}

func TestStreamOpenWithNoOp(t *testing.T) {
	snk := sinks.Slice[string]()
	st := From(sources.Slice([]string{"Hello", "World"}))
	st.WithLogSink(sinks.Func(testutil.LogSinkFunc(t)))
	st.Into(snk)
	select {
	case err := <-st.Open(context.Background()):
		if err != nil {
			t.Fatal(err)
		}
		if len(snk.Get()) != 2 {
			t.Fatal("Data not streaming, expected 2 elements, got ", len(snk.Get()))
		}
	case <-time.After(50 * time.Millisecond):
		t.Fatal("Waited too long ...")
	}
}

func TestStreamOpenWithOp(t *testing.T) {
	src := sources.Slice([]string{"HELLO", "WORLD", "HOW", "ARE", "YOU"})
	snk := sinks.Slice[[]int]()
	op1 := func(ctx context.Context, in string) int {
		return len(in)
	}

	var m sync.RWMutex
	runeCount := 0
	op2 := func(ctx context.Context, in int) int {
		m.Lock()
		runeCount += in
		m.Unlock()
		return runeCount
	}

	strm := From(src).Run(
		exec.Execute(op1),
		exec.Execute(op2),
	).WithLogSink(sinks.Func(testutil.LogSinkFunc(t))).Into(snk)
	select {
	case err := <-strm.Open(context.Background()):
		if err != nil {
			t.Fatal(err)
		}
	case <-time.After(50 * time.Millisecond):
		t.Fatal("Waited too long ...")
	}
	m.RLock()
	if runeCount != 19 {
		t.Fatal("Data not streaming, runeCount 19, got ", runeCount)
	}
	m.RUnlock()
}
