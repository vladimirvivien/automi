package stream

import (
	"context"
	"fmt"
	"log"
	"reflect"

	autoctx "github.com/vladimirvivien/automi/api/context"
	"github.com/vladimirvivien/automi/api/tuple"
)

type StreamOp struct {
	ctx    context.Context
	input  <-chan interface{}
	output chan interface{}
	log    *log.Logger
}

func NewStreamOp(ctx context.Context) *StreamOp {
	log := autoctx.GetLogger(ctx)

	r := new(StreamOp)
	r.ctx = ctx
	r.log = log
	r.output = make(chan interface{}, 1024)

	r.log.Print("component initialized")
	return r
}

func (r *StreamOp) SetInput(in <-chan interface{}) {
	r.input = in
}

func (r *StreamOp) GetOutput() <-chan interface{} {
	return r.output
}

func (r *StreamOp) Exec() (err error) {
	if r.input == nil {
		err = fmt.Errorf("No input channel found")
		return
	}

	go func() {
		defer func() {
			close(r.output)
			r.log.Print("component shutting down")
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
