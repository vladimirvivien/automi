package batch

import (
	"context"
	"fmt"
	"log"
	"reflect"

	autoctx "github.com/vladimirvivien/automi/api/context"
)

// BatchOperator is an executor that batches incoming streamed items based
// on provided criteria.  The batched items are streamed on the
// ouptut channel for downstream processing.
type BatchOperator struct {
	ctx    context.Context
	input  <-chan interface{}
	output chan interface{}
	log    *log.Logger
	size   int
}

// New returns a new BatchOperator operator
func New(ctx context.Context) *BatchOperator {
	log := autoctx.GetLogger(ctx)
	op := new(BatchOperator)
	op.ctx = ctx
	op.log = log
	op.output = make(chan interface{}, 1024)
	op.size = 1024 * 10
	return op
}

// SetInput sets the input channel for the executor node
func (op *BatchOperator) SetInput(in <-chan interface{}) {
	op.input = in
}

// GetOutput returns the output channel of the executer node
func (op *BatchOperator) GetOutput() <-chan interface{} {
	return op.output
}

// Exec is the execution starting point for the operator node.
// The batch operator batches N size items from upstream into
// a slice []T.  When the slice reaches size N, the slice is sent
// downstream for processing.
func (op *BatchOperator) Exec() (err error) {
	if op.input == nil {
		err = fmt.Errorf("No input channel found")
		return
	}

	// The operator dynamically creates the slice of []T.
	// it does this by creating automatically detecting elem type T
	// from the first item that shows up in the channel

	go func() {
		var batchValue reflect.Value
		defer func() {
			// push any straggler items in batch
			if batchValue.IsValid() && batchValue.Len() > 0 {
				op.output <- batchValue.Interface()
			}
			close(op.output)
			op.log.Print("component shutting down")
		}()

		counter := 0
		for {
			select {
			case item, opened := <-op.input:
				if !opened {
					return
				}
				if !batchValue.IsValid() {
					batchType := op.makeBatchType(item)
					batchValue = reflect.MakeSlice(reflect.SliceOf(batchType), 0, op.size)
				}

				batchValue = reflect.Append(batchValue, reflect.ValueOf(item))
				if counter < op.size-1 {
					counter++
					continue
				}
				if counter >= op.size-1 {
					op.output <- batchValue.Interface()
					counter = 0
					batchType := op.makeBatchType(item)
					batchValue = reflect.MakeSlice(reflect.SliceOf(batchType), 0, op.size)
				}
			}
		}
	}()
	return nil
}

// makeBatchType detects and return type to be used for the batch based
// on items in the
func (op *BatchOperator) makeBatchType(item interface{}) reflect.Type {
	itemType := reflect.TypeOf(item)
	var retType reflect.Type

	switch itemType.Kind() {
	case reflect.Array, reflect.Slice:
		elem := itemType.Elem()
		retType = reflect.SliceOf(elem)
	case reflect.Map:
		elemType := itemType.Elem()
		keyType := itemType.Key()
		retType = reflect.MapOf(keyType, elemType)
	default:
		retType = itemType
	}

	return retType
}
