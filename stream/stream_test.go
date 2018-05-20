package stream

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/vladimirvivien/automi/api"
	"github.com/vladimirvivien/automi/collectors"
	"github.com/vladimirvivien/automi/emitters"
)

func TestStream_New(t *testing.T) {
	st := New([]interface{}{"hello"})
	if st.ops == nil {
		t.Fatal("Ops slice not initialized")
	}
	if st.ctx == nil {
		t.Fatal("unitialized context")
	}
	if st.srcParam == nil {
		t.Fatal("src param not initialized")
	}
}

func TestStream_BuilderMethods(t *testing.T) {
	op := func(ctx context.Context, data interface{}) interface{} {
		return nil
	}

	st := New([]interface{}{"Hello", "World", "!!!"}).
		Into(collectors.Slice()).
		Transform(api.UnFunc(op))

	if st.srcParam == nil {
		t.Fatal("From() not setting source")
	}
	if st.snkParam == nil {
		t.Fatal("To() not setting sink")
	}
	if len(st.ops) != 1 {
		t.Fatal("Operation not added to ops slice")
	}
}

func TestStream_InitGraph(t *testing.T) {
	src := emitters.Slice([]string{"Hello", "World"})
	snk := collectors.Slice()
	op1 := api.UnFunc(func(ctx context.Context, data interface{}) interface{} {
		return nil
	})
	op2 := api.UnFunc(func(ctx context.Context, data interface{}) interface{} {
		return nil
	})

	strm := New(src).Into(snk)

	if err := strm.initGraph(); err != nil {
		t.Fatal(err)
	}

	strm = New(src).Transform(op1).Transform(op2).Into(snk)
	if err := strm.initGraph(); err != nil {
		t.Fatal(err)
	}

	if len(strm.ops) != 2 {
		t.Fatal("Not adding operations to stream")
	}
}

func TestStream_Open_NoOp(t *testing.T) {
	snk := collectors.Slice()
	st := New([]string{"Hello", "World"}).Into(snk)
	select {
	case err := <-st.Open():
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

func TestStream_Open_WithOp(t *testing.T) {
	src := emitters.Slice([]string{"HELLO", "WORLD", "HOW", "ARE", "YOU"})
	snk := collectors.Slice()
	op1 := api.UnFunc(func(ctx context.Context, data interface{}) interface{} {
		str := data.(string)
		return len(str)
	})

	var m sync.RWMutex
	runeCount := 0
	op2 := api.UnFunc(func(ctx context.Context, data interface{}) interface{} {
		length := data.(int)
		m.Lock()
		runeCount += length
		m.Unlock()
		return nil
	})

	strm := New(src).Transform(op1).Transform(op2).Into(snk)
	select {
	case err := <-strm.Open():
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
