package stream

import (
	"context"
	"errors"
	"io"
	"os"
	"reflect"

	"github.com/vladimirvivien/automi/api"
	autoctx "github.com/vladimirvivien/automi/api/context"
	"github.com/vladimirvivien/automi/collectors"
	"github.com/vladimirvivien/automi/emitters"
	streamop "github.com/vladimirvivien/automi/operators/stream"
	"github.com/vladimirvivien/automi/util"
)

// Stream represents a stream unto  which executor nodes can be
// attached to operate on the streamed data
type Stream struct {
	srcParam interface{}
	snkParam interface{}
	source   api.Source
	sink     api.Sink
	drain    chan error
	ops      []api.Operator
	ctx      context.Context
	logf     api.LogFunc
	errf     api.ErrorFunc
}

// New creates a new *Stream value
func New(src interface{}) *Stream {
	s := &Stream{
		srcParam: src,
		ops:      make([]api.Operator, 0),
		drain:    make(chan error),
	}

	return s
}

// WithContext sets a context.Context to use.
func (s *Stream) WithContext(ctx context.Context) *Stream {
	s.ctx = ctx
	return s
}

// WithLogFunc sets a function that will receive internal log events
// at runtime.  Supported log function type: func(interface{})
func (s *Stream) WithLogFunc(fn api.LogFunc) *Stream {
	s.logf = fn
	return s
}

// WithErrorFunc sets a function of type func(StreamError) that will be
// invoked when an operator indicates it wants to signal an error by
// defining an operator function of the form func(data)error.
func (s *Stream) WithErrorFunc(fn api.ErrorFunc) *Stream {
	s.errf = fn
	return s
}

// From sets the stream source to use
//func (s *Stream) From(src api.StreamSource) *Stream {
//	s.source = src
//	return s
//}

// Into sets the terminal stream sink to use
func (s *Stream) Into(snk interface{}) *Stream {
	s.snkParam = snk
	return s
}

// ReStream takes upstream items of types []slice []array, map[T]
// and emmits their elements as individual channel items to downstream
// operations.  Items of other types are ignored.
func (s *Stream) ReStream() *Stream {
	sop := streamop.New()
	s.ops = append(s.ops, sop)
	return s
}

// Open opens the Stream which executes all operators nodes.
// If there's an issue prior to execution, an error is returned
// in the error channel.
func (s *Stream) Open() <-chan error {
	s.prepareContext() // ensure context is set

	if err := s.initGraph(); err != nil {
		s.drainErr(err)
		return s.drain
	}

	util.Logfn(s.logf, "Opening stream")

	// open stream
	go func() {
		// open source, if err bail
		if err := s.source.Open(s.ctx); err != nil {
			s.drainErr(err)
			return
		}
		//apply operators, if err bail
		for _, op := range s.ops {
			if err := op.Exec(s.ctx); err != nil {
				s.drainErr(err)
				return
			}
		}

		// open stream sink, after log sink is ready.
		err := <-s.sink.Open(s.ctx)
		util.Logfn(s.logf, "Closing stream")
		s.drain <- err

	}()

	return s.drain
}

// prepareContext setups internal context before
// stream starts execution.
func (s *Stream) prepareContext() {
	if s.ctx == nil {
		s.ctx = context.TODO()
	}
	s.ctx = autoctx.WithLogFunc(s.ctx, s.logf)
	s.ctx = autoctx.WithErrorFunc(s.ctx, s.errf)
}

// bindOps binds operator channels
func (s *Stream) bindOps() {
	util.Logfn(s.logf, "Binding operators")
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
	util.Logfn(s.logf, "Initializing operators")

	// setup source type
	if err := s.setupSource(); err != nil {
		return err
	}

	// setup sink type
	if err := s.setupSink(); err != nil {
		return err
	}

	// if there are no ops, link source to sink
	if len(s.ops) == 0 && s.sink != nil {
		util.Logfn(s.logf, "No operators in stream, binding source to sink directly")
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

// setupSource checks the source, setup the proper type or return err if problem
func (s *Stream) setupSource() error {
	if s.srcParam == nil {
		return errors.New("stream missing source parameter")
	}

	// check specific type
	switch src := s.srcParam.(type) {
	case (api.Source):
		s.source = src
	case *os.File:
		// assume csv
		s.source = emitters.CSV(src)
	case string:
		// assume csv file name
		s.source = emitters.CSV(src)
	case io.Reader:
		s.source = emitters.Reader(src)
	}

	// check on type kind
	srcType := reflect.TypeOf(s.srcParam)
	switch srcType.Kind() {
	case reflect.Slice:
		s.source = emitters.Slice(s.srcParam)
	case reflect.Chan:
		s.source = emitters.Chan(s.srcParam)
	}

	if s.source == nil {
		return errors.New("invalid source")
	}

	return nil
}

// setupSink checks the sink param, setup the proper type or return err if problem
func (s *Stream) setupSink() error {
	// if sink param is nil, use null collector
	if s.snkParam == nil {
		s.sink = collectors.Null()
		return nil
	}

	// check specific type
	switch snk := s.snkParam.(type) {
	case api.Sink:
		s.sink = snk
	case string:
		// assume csv file name
		s.sink = collectors.CSV(snk)
	case *os.File:
		// assume csv file
		s.sink = collectors.CSV(snk)
	case io.Writer:
		s.sink = collectors.Writer(snk)

	default:
		// check by type kind
		srcType := reflect.TypeOf(s.snkParam)
		switch srcType.Kind() {
		case reflect.Slice:
			s.sink = collectors.Slice()
		case reflect.Chan:
		}
	}

	if s.sink == nil {
		return errors.New("invalid sink")
	}

	return nil
}

func (s *Stream) drainErr(err error) {
	go func() { s.drain <- err }()
}
