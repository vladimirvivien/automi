package stream

import (
	"context"
	"log"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/vladimirvivien/automi/api"
	"github.com/vladimirvivien/automi/emitters"
)

// ***** Sink *****
type strSink struct {
	sink  []string
	input <-chan interface{}
	done  chan struct{}
	log   *log.Logger
}

func newStrSink() *strSink {
	return &strSink{
		sink: make([]string, 0),
		done: make(chan struct{}),
		log:  log.New(os.Stderr, "", log.Flags()),
	}
}
func (s *strSink) SetInput(in <-chan interface{}) {
	s.input = in
}

func (s *strSink) Open(ctx context.Context) <-chan error {
	s.log.Print("Opening stream sink")
	result := make(chan error)
	go func() {
		defer close(s.done)
		for str := range s.input {
			s.sink = append(s.sink, str.(string))
		}
		close(result)
	}()
	return result
}

// *** Tests *** //
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
		To(newStrSink()).
		Transform(api.UnFunc(op))

	if st.srcParam == nil {
		t.Fatal("From() not setting source")
	}
	if st.sink == nil {
		t.Fatal("To() not setting sink")
	}
	if len(st.ops) != 1 {
		t.Fatal("Operation not added to ops slice")
	}
}

func TestStream_InitGraph(t *testing.T) {
	src := emitters.Slice([]string{"Hello", "World"})
	snk := newStrSink()
	op1 := api.UnFunc(func(ctx context.Context, data interface{}) interface{} {
		return nil
	})
	op2 := api.UnFunc(func(ctx context.Context, data interface{}) interface{} {
		return nil
	})

	strm := New(src).To(snk)

	if err := strm.initGraph(); err != nil {
		t.Fatal(err)
	}

	if strm.source.GetOutput() != snk.input {
		t.Fatal("Source not link to sink when no ops are present")
	}

	strm = New(src).Transform(op1).Transform(op2).To(snk)
	if err := strm.initGraph(); err != nil {
		t.Fatal(err)
	}

	if len(strm.ops) != 2 {
		t.Fatal("Not adding operations to stream")
	}

	if strm.source.GetOutput() == snk.input {
		t.Fatal("Graph invalid, source skipping ops, linked to sink!")
	}

	if strm.ops[1].GetOutput() != snk.input {
		t.Fatal("Sink not linked to last element in graph")
	}

}

func TestStream_Open_NoOp(t *testing.T) {
	t.Skip()
	src := emitters.Slice([]string{"Hello", "World"})
	snk := newStrSink()
	st := New(src).To(snk)
	select {
	case err := <-st.Open():
		if err != nil {
			t.Fatal(err)
		}
	case <-time.After(50 * time.Millisecond):
		t.Fatal("Waited too long ...")
	}
	if len(snk.sink) != 2 {
		t.Fatal("Data not streaming, expected 2 elements, got ", len(snk.sink))
	}
}

func TestStream_Open_WithOp(t *testing.T) {
	t.Skip()
	src := emitters.Slice([]string{"HELLO", "WORLD", "HOW", "ARE", "YOU"})
	snk := newStrSink()
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

	strm := New(src).Transform(op1).Transform(op2).To(snk)
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
