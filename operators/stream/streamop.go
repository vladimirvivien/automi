package stream

import (
	"context"
	"fmt"
	"reflect"

	"github.com/vladimirvivien/automi/api"
	autoctx "github.com/vladimirvivien/automi/api/context"
	"github.com/vladimirvivien/automi/api/tuple"
	"github.com/vladimirvivien/automi/util"
)

// StreamOperator is an operator takes streamed items of type
// map, array, or slice and unpacks and emits each item individually
// downstream.
type StreamOperator struct {
	input  <-chan interface{}
	output chan interface{}
	logf   api.LogFunc
}

// New creates a *StreamOperator value
func New() *StreamOperator {
	r := new(StreamOperator)
	r.output = make(chan interface{}, 1024)
	return r
}

// SetInput sets the input channel for the executor node
func (r *StreamOperator) SetInput(in <-chan interface{}) {
	r.input = in
}

// GetOutput returns the output channel of the executer node
func (r *StreamOperator) GetOutput() <-chan interface{} {
	return r.output
}

// Exec is the execution starting point for the executor node.
func (r *StreamOperator) Exec(ctx context.Context) (err error) {
	r.logf = autoctx.GetLogFunc(ctx)
	util.Logfn(r.logf, "Stream operator starting")

	if r.input == nil {
		err = fmt.Errorf("No input channel found")
		return
	}

	go func() {
		defer func() {
			util.Logfn(r.logf, "Stream operator done")
			close(r.output)
		}()
		for {
			select {
			case item, opened := <-r.input:
				if !opened {
					return
				}
				itemType := reflect.TypeOf(item)
				itemVal := reflect.ValueOf(item)

				// unpack array, slice, map into individual item stream
				switch itemType.Kind() {
				case reflect.Array, reflect.Slice:
					for i := 0; i < itemVal.Len(); i++ {
						j := itemVal.Index(i)
						r.output <- j.Interface()
					}
				// unpack map as tuple.KV{key, value}
				case reflect.Map:
					for _, key := range itemVal.MapKeys() {
						val := itemVal.MapIndex(key)
						r.output <- tuple.KV{key.Interface(), val.Interface()}
					}
				default:
					r.output <- item
				}
			}
		}
	}()
	return nil
}
