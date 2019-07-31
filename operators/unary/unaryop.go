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
	op          api.UnOperation
	concurrency int
	input       <-chan interface{}
	output      chan interface{}
	logf        api.LogFunc
	errf        api.ErrorFunc
	mutex       sync.RWMutex
}

// NewUnary creates *UnaryOperator value
func New() *UnaryOperator {
	// extract logger
	o := new(UnaryOperator)

	o.concurrency = 1
	o.output = make(chan interface{}, 1024)

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
func (o *UnaryOperator) Exec(ctx context.Context) (err error) {
	o.logf = autoctx.GetLogFunc(ctx)
	o.errf = autoctx.GetErrFunc(ctx)
	util.Logfn(o.logf, "Unary operator started")

	if o.input == nil {
		err = fmt.Errorf("No input channel found")
		return
	}

	go func() {
		defer func() {
			util.Logfn(o.logf, "Unary operator done")
			close(o.output)
		}()

		o.doOp(ctx)
	}()
	return nil
}

func (o *UnaryOperator) doOp(ctx context.Context) {
	if o.op == nil {
		util.Logfn(o.logf, "Unary operator missing operation")
		return
	}
	exeCtx, cancel := context.WithCancel(ctx)

	defer func() {
		util.Logfn(o.logf, "unary operator done, cancelling future items")
		cancel()
	}()

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
			case api.StreamError:
				util.Logfn(o.logf, val)
				autoctx.Err(o.errf, val)
				if item := val.Item(); item != nil {
					select {
					case o.output <- *item:
					case <-exeCtx.Done():
						return
					}
				}
				continue
			case api.PanicStreamError:
				util.Logfn(o.logf, val)
				autoctx.Err(o.errf, api.StreamError(val))
				panic(val)
			case api.CancelStreamError:
				util.Logfn(o.logf, val)
				autoctx.Err(o.errf, api.StreamError(val))
				return
			case error:
				util.Logfn(o.logf, val)
				autoctx.Err(o.errf, api.Error(val.Error()))
				continue

			default:
				select {
				case o.output <- val:
				case <-exeCtx.Done():
					return
				}
			}

		// is cancelling
		case <-exeCtx.Done():
			return
		}
	}
}
