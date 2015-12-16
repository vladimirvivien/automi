package stream

import "github.com/vladimirvivien/automi/api"

type Stream struct {
	source api.StreamSource
	sink   api.StreamSink
	ops    []*api.Operator
}

func New() *Stream {
	s := &Stream{
		ops: make([]*api.Operator, 0),
	}
	return s
}

func (s *Stream) From(src api.StreamSource) *Stream {
	s.source = src
	return s
}

func (s *Stream) To(sink api.StreamSink) *Stream {
	s.sink = sink
	return s
}

func (s *Stream) Open() <-chan struct{} {
	return nil
}

func (s *Stream) linkOps() {

	// if there are no ops, link source to sink
	if len(s.ops) == 0 {
		s.sink.SetInput(s.source.GetOutput())
		return
	}

	// link ops
	for i, op := range s.ops {
		if i == 0 { // link 1st to source
			op.SetInput(s.source.GetOutput())
		} else {
			op.SetInput(s.ops[i-1].GetOutput())
		}
	}

	// link last op to sink
	s.sink.SetInput(s.ops[len(s.ops)-1].GetOutput())
}
