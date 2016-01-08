package stream

import (
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/vladimirvivien/automi/api"

	"golang.org/x/net/context"
)

type strSrc struct {
	src    []string
	output chan interface{}
	log    *logrus.Entry
}

func newStrSrc(s []string) *strSrc {
	return &strSrc{
		src:    s,
		output: make(chan interface{}, 1024),
		log:    logrus.WithField("Component", "StrSrc"),
	}
}
func (s *strSrc) GetOutput() <-chan interface{} {
	return s.output
}
func (s *strSrc) Open(ctx context.Context) error {
	s.log.Infoln("Opening stream source")
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
	log   *logrus.Entry
}

func newStrSink() *strSink {
	return &strSink{
		sink: make([]string, 0),
		done: make(chan struct{}),
		log:  logrus.WithField("Component", "StrSink"),
	}
}
func (s *strSink) SetInput(in <-chan interface{}) {
	s.input = in
}

func (s *strSink) Open(ctx context.Context) <-chan error {
	s.log.Infoln("Opening stream sink")
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

func TestStream_Process(t *testing.T) {
	src := NewSliceSource("hello", "world")
	snk := NewDrain()
	strm := New().From(src).Process(func(s string) string {
		return strings.ToUpper(s)
	}).To(snk)

	wait := make(chan struct{})
	go func() {
		defer close(wait)
		for data := range snk.GetOutput() {
			val := data.(string)
			if val != "HELLO" && val != "WORLD" {
				t.Fatalf("Got %v of type %T", val, val)
			}
		}
	}()

	select {
	case err := <-strm.Open():
		if err != nil {
			t.Fatal(err)
		}
		select {
		case <-wait:
		case <-time.After(100 * time.Microsecond):
			t.Fatal("Draining took too long")
		}
	case <-time.After(50 * time.Millisecond):
		t.Fatal("Waited too long ...")
	}
}

func TestStream_Filter(t *testing.T) {
	src := newStrSrc([]string{"HELLO", "WORLD", "HOW", "ARE", "YOU"})
	snk := newStrSink()
	strm := New().From(src).Filter(func(data string) bool {
		return !strings.Contains(data, "O")
	})
	strm.To(snk)

	select {
	case err := <-strm.Open():
		if err != nil {
			t.Fatal(err)
		}
	case <-time.After(50 * time.Millisecond):
		t.Fatal("Waited too long ...")
	}
	if len(snk.sink) != 1 {
		t.Fatal("Filter failed, expected 1 element, got ", len(snk.sink))
	}
}

func TestStream_Map(t *testing.T) {
	src := newStrSrc([]string{"HELLO", "WORLD", "HOW", "ARE", "YOU"})
	snk := NewDrain()
	strm := New().From(src).Map(func(data string) int {
		return len(data)
	}).To(snk)

	var m sync.RWMutex
	count := 0
	wait := make(chan struct{})
	go func() {
		defer close(wait)
		for data := range snk.GetOutput() {
			val := data.(int)
			m.Lock()
			count += val
			m.Unlock()
		}
	}()

	select {
	case err := <-strm.Open():
		if err != nil {
			t.Fatal(err)
		}
		select {
		case <-wait:
		case <-time.After(50 * time.Millisecond):
			t.Fatal("Waited too long for sink output to process")
		}

	case <-time.After(50 * time.Millisecond):
		t.Fatal("Waited too long for stream to Open...")
	}

	m.RLock()
	if count != 19 {
		t.Fatal("Map failed, expected count 19, got ", count)
	}
	m.RUnlock()
}

func TestStream_FlatMap(t *testing.T) {
	src := newStrSrc([]string{"HELLO WORLD", "HOW ARE YOU?"})
	snk := NewDrain()
	strm := New().From(src).FlatMap(func(data string) []string {
		return strings.Split(data, " ")
	}).To(snk)

	var m sync.RWMutex
	count := 0
	expected := 5
	wait := make(chan struct{})
	go func() {
		defer close(wait)
		for data := range snk.GetOutput() {
			vals := data.([]string)
			m.Lock()
			count += len(vals)
			m.Unlock()
		}
	}()

	select {
	case err := <-strm.Open():
		if err != nil {
			t.Fatal(err)
		}
		select {
		case <-wait:
			if count != expected {
				t.Fatalf("Expecting %d words, got %d", expected, count)
			}
		case <-time.After(50 * time.Millisecond):
			t.Fatal("Took too long to process sink output")
		}
	case <-time.After(50 * time.Millisecond):
		t.Fatal("Waited too long ...")
	}
}
func TestStream_Reduce(t *testing.T) {
	src := NewSliceSource(1, 2, 3, 4, 5)
	snk := NewDrain()
	strm := New().From(src).Reduce(func(op1, op2 interface{}) interface{} {
		prev := op1.(int)
		return prev + op2.(int)
	}).SetInitialState(0).To(snk)

	actual := 15
	wait := make(chan struct{})
	go func() {
		defer close(wait)
		select {
		case result := <-snk.GetOutput():
			if result.(int) != actual {
				t.Fatal("Expecting ", actual, " got ", result)
			}
		case <-time.After(5 * time.Millisecond):
			t.Fatal("Sink took too long to get result")
		}
	}()

	select {
	case err := <-strm.Open():
		if err != nil {
			t.Fatal(err)
		}
		select {
		case <-wait:
		case <-time.After(10 * time.Millisecond):
			t.Fatal("Stream took too long to complete")
		}
	case <-time.After(10 * time.Millisecond):
		t.Fatal("Took too long")
	}
}
