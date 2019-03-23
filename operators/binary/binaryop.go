package binary

import (
	"context"
	"fmt"
	"sync"

	"github.com/vladimirvivien/automi/api"
	autoctx "github.com/vladimirvivien/automi/api/context"
	"github.com/vladimirvivien/automi/util"
)

// BinaryOperator represents an operator that knows how to run a
// binary operations such as aggregation, reduction, etc.
type BinaryOperator struct {
	op          api.BinOperation
	state       interface{}
	concurrency int
	input       <-chan interface{}
	output      chan interface{}
	logf        api.LogFunc
	errf        api.ErrorFunc
	cancelled   bool
	mutex       sync.RWMutex
}

// New creates a new binary operator
func New() *BinaryOperator {
	// extract logger
	o := new(BinaryOperator)
	o.concurrency = 1
	o.output = make(chan interface{}, 1024)
	return o
}

// SetOperation sets the operation to execute
func (o *BinaryOperator) SetOperation(op api.BinOperation) {
	o.op = op
}

// SetInitialState sets an initial value used with the first streamed item
func (o *BinaryOperator) SetInitialState(val interface{}) {
	o.state = val
}

// SetConcurrency sets the concurrency level
func (o *BinaryOperator) SetConcurrency(concurr int) {
	o.concurrency = concurr
	if o.concurrency < 1 {
		o.concurrency = 1
	}
}

// SetInput sets the input channel for the executor node
func (o *BinaryOperator) SetInput(in <-chan interface{}) {
	o.input = in
}

// GetOutput returns the output channel for the executor node
func (o *BinaryOperator) GetOutput() <-chan interface{} {
	return o.output
}

// Exec executes the associated operation
func (o *BinaryOperator) Exec(ctx context.Context) (err error) {
	o.logf = autoctx.GetLogFunc(ctx)
	o.errf = autoctx.GetErrFunc(ctx)
	util.Logfn(o.logf, "Binary operator starting")

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
			o.output <- o.state
			close(o.output)
			util.Logfn(o.logf, "Binary operator done")
		}()

		var barrier sync.WaitGroup
		wgDelta := o.concurrency
		barrier.Add(wgDelta)

		for i := 0; i < o.concurrency; i++ { // workers
			go func(wg *sync.WaitGroup) {
				defer wg.Done()
				o.doProc(ctx)
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
				util.Logfn(o.logf, "Binary operator cancelled")
				return
			}
		case <-ctx.Done():
			return
		}
	}()
	return nil
}

// doProc is a helper function that executes the operation
func (o *BinaryOperator) doProc(ctx context.Context) {
	if o.op == nil {
		util.Logfn(o.logf, "Binary operator has no operation")
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

			o.state = o.op.Apply(exeCtx, o.state, item)

			switch val := o.state.(type) {
			case nil:
				continue
			case api.StreamError:
				util.Logfn(o.logf, val)
				autoctx.Err(o.errf, val)
				continue
			}

		// is cancelling
		case <-ctx.Done():
			util.Logfn(o.logf, "Binary operator cancelling")
			o.mutex.Lock()
			cancel()
			o.cancelled = true
			o.mutex.Unlock()
			return
		}
	}
}
