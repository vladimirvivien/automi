package operators

import (
	"context"
	"fmt"
	"log"

	autoctx "github.com/vladimirvivien/automi/api/context"
)

// BatchOp is an executor that batches incoming streamed items based
// on provided criteria.  The batched items are streamed on the
// ouptut channel for downstream processing.
type BatchOp struct {
	ctx    context.Context
	input  <-chan interface{}
	output chan interface{}
	log    *log.Logger
	size   int
}

// NewBatchOp returns a new BatchOp operator
func NewBatchOp(ctx context.Context) *BatchOp {
	log := autoctx.GetLogger(ctx)
	op := new(BatchOp)
	op.ctx = ctx
	op.log = log
	op.output = make(chan interface{}, 1024)
	op.size = 1024 * 10
	return op
}

// SetInput sets the input channel for the executor node
func (op *BatchOp) SetInput(in <-chan interface{}) {
	op.input = in
}

// GetOutput returns the output channel of the executer node
func (op *BatchOp) GetOutput() <-chan interface{} {
	return op.output
}

// Exec is the execution starting point for the executor node.
func (op *BatchOp) Exec() (err error) {
	if op.input == nil {
		err = fmt.Errorf("No input channel found")
		return
	}

	go func() {
		defer func() {
			close(op.output)
			op.log.Print("component shutting down")
		}()
		batch := make([]interface{}, 0, op.size)
		counter := 0
		for {
			select {
			case item, opened := <-op.input:
				if !opened {
					op.output <- batch
					return
				}
				batch = append(batch, item)
				switch {
				case counter < op.size:
					counter++
				case counter >= op.size:
					counter = 0
					op.output <- batch
				}
			}
		}
	}()
	return nil
}
