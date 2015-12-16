package stream

import (
	"github.com/Sirupsen/logrus"
	"github.com/vladimirvivien/automi/api"
	"golang.org/x/net/context"
)

type Stream struct {
	source api.StreamSource
	sink   api.StreamSink
	ops    []*api.Operator
	ctx    context.Context
	log    *logrus.Entry
}

func New() *Stream {
	s := &Stream{
		ops: make([]*api.Operator, 0),
		log: logrus.WithField("Stream", "Default"),
		ctx: context.Background(),
	}
	return s
}

func (s *Stream) WithContext(ctx context.Context) *Stream {
	s.ctx = ctx
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

func (s *Stream) Do(op api.Operation) *Stream {
	operator := api.NewOperator(s.ctx)
	operator.SetOperation(op)
	s.ops = append(s.ops, operator)
	return s
}

func (s *Stream) Open() <-chan struct{} {
	s.linkOps() // link nodes
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
