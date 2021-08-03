package batch

import (
	"context"
	"fmt"
	"reflect"

	"github.com/vladimirvivien/automi/api"
	autoctx "github.com/vladimirvivien/automi/api/context"
	"github.com/vladimirvivien/automi/util"
)

// Operator is an executor that batches incoming streamed items based
// on provided criteria.  The batched items are streamed on the
// output channel for downstream processing.
type Operator struct {
	input   <-chan interface{}
	output  chan interface{}
	logf    api.LogFunc
	trigger api.BatchTrigger
}

// New returns a new Operator operator
func New() *Operator {
	op := new(Operator)
	op.output = make(chan interface{}, 1024)
	return op
}

// SetInput sets the input channel for the executor node
func (op *Operator) SetInput(in <-chan interface{}) {
	op.input = in
}

// GetOutput returns the output channel of the executer node
func (op *Operator) GetOutput() <-chan interface{} {
	return op.output
}

// SetTrigger sets the batch operation to apply for this operator
func (op *Operator) SetTrigger(trigger api.BatchTrigger) {
	op.trigger = trigger
}

// Exec is the execution starting point for the operator node.
// The batch operator batches N size items from upstream into
// a slice []T.  When the slice reaches size N, the slice is sent
// downstream for processing.
func (op *Operator) Exec(ctx context.Context) (err error) {
	op.logf = autoctx.GetLogFunc(ctx)
	util.Logfn(op.logf, "Batch operator starting")

	if op.input == nil {
		err = fmt.Errorf("No input channel found")
		return
	}

	// The operator dynamically creates the slice of []T.
	// it does this by creating automatically detecting elem type T
	// from the first item that shows up in the channel

	go func() {
		var batchValue reflect.Value
		exeCtx, cancel := context.WithCancel(ctx)

		defer func() {
			util.Logfn(op.logf, "Closing batch operator")
			// push any straggler items in batch
			if batchValue.IsValid() && batchValue.Len() > 0 {
				op.output <- batchValue.Interface()
			}
			cancel()
			close(op.output)
		}()

		// default to TriggerAll.
		if op.trigger == nil {
			op.trigger = TriggerAll()
		}

		var index int64 = 1
		for {
			select {
			case item, opened := <-op.input:
				if !opened {
					return
				}
				// detect type of first item to create proper
				// Slice type for batch.
				if !batchValue.IsValid() {
					batchType := op.makeBatchType(item)
					batchValue = reflect.MakeSlice(reflect.SliceOf(batchType), 0, 1)
				}

				batchValue = reflect.Append(batchValue, reflect.ValueOf(item))
				done := op.trigger.Done(ctx, item, index)
				if !done {
					index++
					continue
				}

				// done batching, push downstream
				select {
				case op.output <- batchValue.Interface():
					index = 1
					batchType := op.makeBatchType(item)
					batchValue = reflect.MakeSlice(reflect.SliceOf(batchType), 0, 1)
				case <-exeCtx.Done():
					return
				}

			case <-exeCtx.Done():
				return
			}
		}
	}()
	return nil
}

// makeBatchType detects and return type to be used for the batch based
// on items in the
func (op *Operator) makeBatchType(item interface{}) reflect.Type {
	itemType := reflect.TypeOf(item)
	var retType reflect.Type

	switch itemType.Kind() {
	case reflect.Slice:
		elem := itemType.Elem()
		retType = reflect.SliceOf(elem)
	case reflect.Array:
		elemType := itemType.Elem()
		itemCount := itemType.Len()
		retType = reflect.ArrayOf(itemCount, elemType)
	case reflect.Map:
		elemType := itemType.Elem()
		keyType := itemType.Key()
		retType = reflect.MapOf(keyType, elemType)
	default:
		retType = itemType
	}

	return retType
}
