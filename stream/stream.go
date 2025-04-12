package stream

import (
	"context"
	"fmt"
	"sync"

	"github.com/vladimirvivien/automi/api"
	"github.com/vladimirvivien/automi/log"
)

// Stream represents a stream unto  which executor nodes can be
// attached to operate on the streamed data
type Stream struct {
	drain       chan error
	source      api.Source
	nodes       []api.Operator
	sink        api.Sink
	logChan     chan any
	logSink     api.Sink
	logSyncWait sync.WaitGroup
}

// From creates a new *Stream from specified api.Source
func From(src api.Source) *Stream {
	s := &Stream{
		source: src,
		drain:  make(chan error),
	}

	return s
}

func (s *Stream) WithLogSink(sink api.Sink) *Stream {
	if sink == nil {
		return s
	}
	s.logChan = make(chan any, 1024)
	s.logSink = sink
	return s
}

func (s *Stream) Run(nodes ...api.Operator) *Stream {
	s.nodes = nodes
	return s
}

func (s *Stream) Into(sink api.Sink) *Stream {
	s.sink = sink
	return s
}

func (s *Stream) GetSource() api.Source {
	return s.source
}

func (s *Stream) GetSink() api.Sink {
	return s.sink
}

// Open opens the Stream which executes all operators nodes.
func (s *Stream) Open(ctx context.Context) <-chan error {
	s.Log(ctx, log.LogInfo("Starting stream"))

	// setup and open reporters early to report on stream activities
	if s.logSink != nil && s.logChan != nil {
		s.logSyncWait.Add(1)
		s.logSink.SetInput(s.logChan) // bind log chan/sink
		go func() {
			<-s.logSink.Open(ctx)
			s.logSyncWait.Done()
		}()

		s.Log(ctx, log.LogInfo("Initialized stream reporter channel"))
	}

	if err := s.initGraph(ctx); err != nil {
		s.drainErr(err)
		return s.drain
	}

	// open stream
	go func() {
		strmCtx, cancel := context.WithCancel(ctx)
		defer func() {
			cancel()
		}()

		// open source, if err bail
		if err := s.source.Open(strmCtx); err != nil {
			//s.drainErr(err)
			return
		}

		//open all operators in graph, if err bail
		for _, op := range s.nodes {
			if err := op.Exec(strmCtx); err != nil {
				s.drainErr(err)
				return
			}
		}

		// open stream sink and wait for completion
		select {
		case err := <-s.sink.Open(strmCtx):
			s.Log(ctx, log.LogInfo("Closing stream"))
			if s.logChan != nil {
				s.Log(ctx, log.LogInfo("Stopping stream reporter"))
				go close(s.logChan)  // closing reporter chan (no logging after this)
				s.logSyncWait.Wait() // wait for logging to stop
			}
			s.drain <- err
		case <-strmCtx.Done():
			s.Log(ctx, log.LogInfo("Canceling stream"))
			if s.logChan != nil {
				s.Log(ctx, log.LogInfo("Stopping stream reporter"))
				go close(s.logChan)
				s.logSyncWait.Wait()
			}
			s.drain <- strmCtx.Err()
		}
	}()

	return s.drain
}

// initGraph initialize stream graph source + ops +
func (s *Stream) initGraph(ctx context.Context) error {
	s.Log(ctx, log.LogInfo("Initializing stream graph"))
	if s.source == nil {
		s.Log(ctx, log.LogError("No source configured"))
		return api.ErrStreamEmpty
	}
	s.source.SetLogFunc(s.Log)

	if s.sink == nil {
		s.Log(ctx, log.LogError("No sink configured"))
		return api.ErrSinkEmpty
	}
	//s.sink.SetLogFunc(s.Log)

	// if there are no ops, link source to sink
	if len(s.nodes) == 0 && s.sink != nil {
		s.Log(ctx, log.LogWarn("No operator nodes found: binding source to sink directly"))
		s.sink.SetInput(s.source.GetOutput())
		return nil
	}

	// link operators
	s.bindOps(ctx)

	// link last op to sink
	if s.sink != nil {
		idx := len(s.nodes) - 1
		s.sink.SetInput(s.nodes[idx].GetOutput())
		s.Log(ctx, log.LogInfo(fmt.Sprintf("Binding node %d --> sink", idx)))
	}

	return nil
}

// bindOps binds operator channels
func (s *Stream) bindOps(ctx context.Context) {
	if s.nodes == nil {
		s.Log(ctx, log.LogWarn("Node binding: no operator nodes found"))
		return
	}
	for i, op := range s.nodes {
		if i == 0 { // link 1st to source
			op.SetInput(s.source.GetOutput())
			s.Log(ctx, log.LogInfo(fmt.Sprintf("Binding source --> node %d", i)))
		} else {
			op.SetInput(s.nodes[i-1].GetOutput())
			s.Log(ctx, log.LogInfo(fmt.Sprintf("Binding node %d --> node %d", i-1, i)))
		}

		// set operator reporter channels
		op.SetLogFunc(s.Log)
	}
}

func (s *Stream) drainErr(err error) {
	go func() { s.drain <- err }()
}

// Log sends an api.StreamLog to the stream reporter channel
func (s *Stream) Log(ctx context.Context, log api.StreamLog) {
	if s.logSink != nil && s.logChan != nil {
		select {
		case s.logChan <- log:
		case <-ctx.Done():
		default:
		}
	}
}
