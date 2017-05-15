package unary

import (
	"context"
	"fmt"
	"log"
	"sync"

	"github.com/vladimirvivien/automi/api"
	autoctx "github.com/vladimirvivien/automi/api/context"
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
	log         *log.Logger
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

	o.log.Printf("Component initialized")
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

	o.log.Print("Execution started")

	go func() {
		defer func() {
			close(o.output)
			o.log.Print("Shuttingdown component")
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
				o.log.Printf("Component cancelling...")
				return
			}
		case <-o.ctx.Done():
			o.log.Print("UnaryOp done.")
			return
		}
	}()
	return nil
}

func (o *UnaryOperator) doProc(ctx context.Context) {
	if o.op == nil {
		o.log.Print("No operation defined for UnaryOp")
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
				o.log.Print(val)
				continue
			default:
				o.output <- val
			}

		// is cancelling
		case <-ctx.Done():
			o.log.Println("Cancelling....")
			o.mutex.Lock()
			cancel()
			o.cancelled = true
			o.mutex.Unlock()
			return
		}
	}
}
