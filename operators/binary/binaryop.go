package binary

import (
	"context"
	"fmt"

	"github.com/vladimirvivien/automi/api"
	autoctx "github.com/vladimirvivien/automi/api/context"
	"github.com/vladimirvivien/automi/util"
)

// Operator represents an operator that knows how to run a
// binary operations such as aggregation, reduction, etc.
type Operator struct {
	op          api.BinOperation
	state       interface{}
	concurrency int
	input       <-chan interface{}
	output      chan interface{}
	logf        api.LogFunc
	errf        api.ErrorFunc
}

// New creates a new binary operator
func New() *Operator {
	// extract logger
	o := new(Operator)
	o.concurrency = 1
	o.output = make(chan interface{}, 1024)
	return o
}

// SetOperation sets the operation to execute
func (o *Operator) SetOperation(op api.BinOperation) {
	o.op = op
}

// SetInitialState sets an initial value used with the first streamed item
func (o *Operator) SetInitialState(val interface{}) {
	o.state = val
}

// SetConcurrency sets the concurrency level
func (o *Operator) SetConcurrency(concurr int) {
	o.concurrency = concurr
	if o.concurrency < 1 {
		o.concurrency = 1
	}
}

// SetInput sets the input channel for the executor node
func (o *Operator) SetInput(in <-chan interface{}) {
	o.input = in
}

// GetOutput returns the output channel for the executor node
func (o *Operator) GetOutput() <-chan interface{} {
	return o.output
}

// Exec executes the associated operation
func (o *Operator) Exec(ctx context.Context) (err error) {
	o.logf = autoctx.GetLogFunc(ctx)
	o.errf = autoctx.GetErrFunc(ctx)
	util.Logfn(o.logf, "Binary operator starting")

	if o.input == nil {
		err = fmt.Errorf("No input channel found")
		return
	}

	go func() {
		defer func() {
			o.output <- o.state
			close(o.output)
			util.Logfn(o.logf, "Binary operator done")
		}()
		o.doOp(ctx)
	}()
	return nil
}

// doProc is a helper function that executes the operation
func (o *Operator) doOp(ctx context.Context) {
	if o.op == nil {
		util.Logfn(o.logf, "Binary operator has no operation")
		return
	}
	exeCtx, cancel := context.WithCancel(ctx)

	defer func() {
		util.Logfn(o.logf, "Binary operator cancelling")
		cancel()
	}()

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
		case <-exeCtx.Done():
			return
		}
	}
}
