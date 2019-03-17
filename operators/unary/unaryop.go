package unary

import (
	"context"
	"fmt"
	"sync"

	"github.com/vladimirvivien/automi/api"
	autoctx "github.com/vladimirvivien/automi/api/context"
	"github.com/vladimirvivien/automi/util"
)

type packed struct {
	vals []interface{}
}

func pack(vals ...interface{}) packed {
	return packed{vals}
}

// UnaryOp is an executor node that can execute a unary operation (i.e. transformation, etc)
type UnaryOperator struct {
	ctx         context.Context
	op          api.UnOperation
	concurrency int
	input       <-chan interface{}
	output      chan interface{}
	logf        api.LogFunc
	errf        api.ErrorFunc
	cancelled   bool
	mutex       sync.RWMutex
}

// NewUnary creates *UnaryOperator value
func New(ctx context.Context) *UnaryOperator {
	// extract logger
	o := new(UnaryOperator)
	o.ctx = ctx
	o.logf = autoctx.GetLogFunc(ctx)
	o.errf = autoctx.GetErrFunc(ctx)

	o.concurrency = 1
	o.output = make(chan interface{}, 1024)

	util.Logfn(o.logf, "Unary operator started")
	return o
}

// SetOperation sets the executor operation
func (o *UnaryOperator) SetOperation(op api.UnOperation) {
	o.op = op
}

// SetConcurrency sets the concurrency level for the operation
func (o *UnaryOperator) SetConcurrency(concurr int) {
	o.concurrency = concurr
	if o.concurrency < 1 {
		o.concurrency = 1
	}
}

// SetInput sets the input channel for the executor node
func (o *UnaryOperator) SetInput(in <-chan interface{}) {
	o.input = in
}

// GetOutput returns the output channel for the executor node
func (o *UnaryOperator) GetOutput() <-chan interface{} {
	return o.output
}

// Exec is the entry point for the executor
func (o *UnaryOperator) Exec() (err error) {
	if o.input == nil {
		err = fmt.Errorf("No input channel found")
		return
	}

	// validate p
	if o.concurrency < 1 {
		o.concurrency = 1
	}

	go func() {
		defer func() {
			util.Logfn(o.logf, "Unary operator done")
			close(o.output)
		}()

		var barrier sync.WaitGroup
		wgDelta := o.concurrency
		barrier.Add(wgDelta)

		for i := 0; i < o.concurrency; i++ { // workers
			go func(wg *sync.WaitGroup) {
				defer wg.Done()
				o.doProc(o.ctx)
			}(&barrier)
		}

		wait := make(chan struct{})
		go func() {
			defer close(wait)
			barrier.Wait()
		}()

		select {
		case <-wait:
			if o.cancelled {
				util.Logfn(o.logf, "Unary operator cancelled")
				return
			}
		case <-o.ctx.Done():
			return
		}
	}()
	return nil
}

func (o *UnaryOperator) doProc(ctx context.Context) {
	if o.op == nil {
		util.Logfn(o.logf, "Unary operator missing operation")
		return
	}
	exeCtx, cancel := context.WithCancel(ctx)

	for {
		select {
		// process incoming item
		case item, opened := <-o.input:
			if !opened {
				cancel()
				return
			}

			result := o.op.Apply(exeCtx, item)

			switch val := result.(type) {
			case nil:
				continue
			case api.StreamError:
				util.Logfn(o.logf, val)
				autoctx.Err(o.errf, val)
				if item := val.Item(); item != nil {
					o.output <- *item
				}
				continue
			case api.PanicStreamError:
				util.Logfn(o.logf, val)
				autoctx.Err(o.errf, api.StreamError(val))
				panic(val)
			case api.CancelStreamError:
				util.Logfn(o.logf, val)
				autoctx.Err(o.errf, api.StreamError(val))
				func() {
					o.mutex.Lock()
					defer o.mutex.Unlock()
					cancel()
					o.cancelled = true
				}()
			case error:
				util.Logfn(o.logf, val)
				autoctx.Err(o.errf, api.Error(val.Error()))
				continue

			default:
				o.output <- val
			}

		// is cancelling
		case <-ctx.Done():
			util.Logfn(o.logf, "unary operator cancelling")
			func() {
				o.mutex.Lock()
				defer o.mutex.Unlock()
				cancel()
				o.cancelled = true
			}()
			return
		}
	}
}
