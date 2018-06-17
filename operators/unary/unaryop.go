package unary

import (
	"context"
	"fmt"
	"sync"

	"github.com/go-faces/logger"
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
	log         logger.Interface
	cancelled   bool
	mutex       sync.RWMutex
}

// NewUnary creates *UnaryOperator value
func New(ctx context.Context) *UnaryOperator {
	// extract logger
	log := autoctx.GetLogger(ctx)

	o := new(UnaryOperator)
	o.ctx = ctx
	o.log = log

	o.concurrency = 1
	o.output = make(chan interface{}, 1024)

	util.Log(o.log, "unary operator initialized")
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
			util.Log(o.log, "unary operator closing")
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
				util.Log(o.log, "unary operator cancelled")
				return
			}
		case <-o.ctx.Done():
			util.Log(o.log, "unary operator done")
			return
		}
	}()
	return nil
}

func (o *UnaryOperator) doProc(ctx context.Context) {
	if o.op == nil {
		util.Log(o.log, "unary operator missing operation")
		return
	}
	exeCtx, cancel := context.WithCancel(ctx)

	for {
		select {
		// process incoming item
		case item, opened := <-o.input:
			if !opened {
				return
			}

			result := o.op.Apply(exeCtx, item)

			switch val := result.(type) {
			case nil:
				continue
			case error, api.ProcError:
				util.Log(o.log, val)
				continue
			default:
				o.output <- val
			}

		// is cancelling
		case <-ctx.Done():
			util.Log(o.log, "unary operator cancelling")
			o.mutex.Lock()
			cancel()
			o.cancelled = true
			o.mutex.Unlock()
			return
		}
	}
}
