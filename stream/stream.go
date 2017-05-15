package stream

import (
	"context"
	"fmt"
	"log"

	"github.com/vladimirvivien/automi/api"
	autoctx "github.com/vladimirvivien/automi/api/context"
	"github.com/vladimirvivien/automi/operators/streamop"
)

// Stream represents a stream unto  which executor nodes can be
// attached to operate on the streamed data
type Stream struct {
	source api.StreamSource
	sink   api.StreamSink
	drain  <-chan interface{}
	ops    []api.Operator
	ctx    context.Context
	log    *log.Logger
}

// New creates a new *Stream value
func New() *Stream {
	s := &Stream{
		ops: make([]api.Operator, 0),
		ctx: context.Background(),
	}
	s.log = autoctx.GetLogger(s.ctx)
	return s
}

// WithContext sets a context.Context to use
func (s *Stream) WithContext(ctx context.Context) *Stream {
	s.ctx = ctx
	return s
}

// From sets the stream source to use
func (s *Stream) From(src api.StreamSource) *Stream {
	s.source = src
	return s
}

// To sets the terminal stream sink to use
func (s *Stream) To(sink api.StreamSink) *Stream {
	s.sink = sink
	return s
}

// ReStream takes upstream items of types []slice []array, map[T]
// and emmits their elements as individual channel items to downstream
// operations.  Items of other types are ignored.
func (s *Stream) ReStream() *Stream {
	sop := streamop.New(s.ctx)
	s.ops = append(s.ops, sop)
	return s
}

// Open opens the Stream which starts all attached operators
func (s *Stream) Open() <-chan error {
	result := make(chan error, 1)
	if err := s.initGraph(); err != nil {
		result <- err
		return result
	}

	// open stream
	go func() {
		// open source, if err bail
		if err := s.source.Open(s.ctx); err != nil {
			result <- err
			return
		}
		//apply operators, if err bail
		for _, op := range s.ops {
			if err := op.Exec(); err != nil {
				result <- err
				return
			}
		}
		// open sink, pipe result out
		err := <-s.sink.Open(s.ctx)
		result <- err
	}()

	return result
}

// bindOps binds operator channels
func (s *Stream) bindOps() {
	s.log.Print("binding operators")
	if s.ops == nil {
		return
	}
	for i, op := range s.ops {
		if i == 0 { // link 1st to source
			op.SetInput(s.source.GetOutput())
		} else {
			op.SetInput(s.ops[i-1].GetOutput())
		}
	}
}

// initGraph initialize stream graph source + ops +
func (s *Stream) initGraph() error {
	s.log.Print("Preparing stream operator graph")
	if s.source == nil {
		return fmt.Errorf("Operator graph failed, missing source")
	}

	// if there are no ops, link source to sink
	if len(s.ops) == 0 && s.sink != nil {
		s.log.Print("No operator nodes found, binding source to sink directly")
		s.sink.SetInput(s.source.GetOutput())
		return nil
	}

	// link ops
	s.bindOps()

	// link last op to sink
	if s.sink != nil {
		s.sink.SetInput(s.ops[len(s.ops)-1].GetOutput())
	}

	return nil
}
