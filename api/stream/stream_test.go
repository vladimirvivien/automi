package stream

import "testing"

type strSrc struct {
	src    []string
	output chan interface{}
}

func newStrSrc(s []string) *strSrc {
	return &strSrc{src: s, output: make(chan interface{}, 1024)}
}
func (s *strSrc) GetOutput() <-chan interface{} {
	return s.output
}
func (s *strSrc) Open() error {
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
}

func newStrSink() *strSink {
	return &strSink{
		sink: make([]string, 0),
		done: make(chan struct{}),
	}
}
func (s *strSink) SetInput(in <-chan interface{}) {
	s.input = in
}

func (s *strSink) Open() error {
	go func() {
		defer close(s.done)
		for str := range s.input {
			s.sink = append(s.sink, str.(string))
		}
	}()
	return nil
}

func (s *strSink) Done() <-chan struct{} {
	return s.done
}

// *** Tests *** //
func TestStream_New(t *testing.T) {
	st := New()
	if st.ops == nil {
		t.Fatal("Ops slice not initialized")
	}
}

func TestStream_BuilderMethods(t *testing.T) {
	st := New()
	st.
		From(newStrSrc([]string{"Hello", "World", "!!"})).
		To(newStrSink())

	if st.source == nil {
		t.Fatal("From() not setting source")
	}
	if st.sink == nil {
		t.Fatal("To() not setting sink")
	}
}

func TestStream_Linkops(t *testing.T) {
	src := newStrSrc([]string{"Hello", "World"})
	snk := newStrSink()
	st := New().From(src).To(snk)

	st.linkOps() // should link source to sink
	if src.GetOutput() != snk.input {
		t.Fatal("Source not link to sink when no ops are present")
	}
}
