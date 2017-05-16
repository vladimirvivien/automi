package stream

import (
	"context"
	"log"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/vladimirvivien/automi/api"
)

type strSrc struct {
	src    []string
	output chan interface{}
	log    *log.Logger
}

func newStrSrc(s []string) *strSrc {
	return &strSrc{
		src:    s,
		output: make(chan interface{}, 1024),
		log:    log.New(os.Stderr, "", log.Flags()),
	}
}
func (s *strSrc) GetOutput() <-chan interface{} {
	return s.output
}
func (s *strSrc) Open(ctx context.Context) error {
	s.log.Print("Opening stream source")
	go func() {
		defer close(s.output)
		for _, str := range s.src {
			s.output <- str
		}
	}()
	return nil
}

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
	st := New()
	if st.ops == nil {
		t.Fatal("Ops slice not initialized")
	}
}

func TestStream_BuilderMethods(t *testing.T) {
	op := api.UnFunc(func(ctx context.Context, data interface{}) interface{} {
		return nil
	})

	st := New()
	st.
		From(newStrSrc([]string{"Hello", "World", "!!"})).
		To(newStrSink()).
		Transform(op)

	if st.source == nil {
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
	src := newStrSrc([]string{"Hello", "World"})
	snk := newStrSink()
	op1 := api.UnFunc(func(ctx context.Context, data interface{}) interface{} {
		return nil
	})
	op2 := api.UnFunc(func(ctx context.Context, data interface{}) interface{} {
		return nil
	})

	strm := New().From(src).To(snk)

	if err := strm.initGraph(); err != nil {
		t.Fatal(err)
	}

	if src.GetOutput() != snk.input {
		t.Fatal("Source not link to sink when no ops are present")
	}

	strm = New().From(src).Transform(op1).Transform(op2).To(snk)
	if err := strm.initGraph(); err != nil {
		t.Fatal(err)
	}

	if len(strm.ops) != 2 {
		t.Fatal("Not adding operations to stream")
	}

	if src.GetOutput() == snk.input {
		t.Fatal("Graph invalid, source skipping ops, linked to sink!")
	}

	if strm.ops[1].GetOutput() != snk.input {
		t.Fatal("Sink not linked to last element in graph")
	}

}

func TestStream_Open_NoOp(t *testing.T) {
	src := newStrSrc([]string{"Hello", "World"})
	snk := newStrSink()
	st := New()
	st.From(src).To(snk)
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
	src := newStrSrc([]string{"HELLO", "WORLD", "HOW", "ARE", "YOU"})
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

	strm := New().From(src).Transform(op1).Transform(op2).To(snk)
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
