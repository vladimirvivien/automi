package stream

import (
	"fmt"
	"reflect"

	"golang.org/x/net/context"

	"github.com/Sirupsen/logrus"
	autoctx "github.com/vladimirvivien/automi/api/context"
	"github.com/vladimirvivien/automi/api/tuple"
)

type StreamOp struct {
	ctx    context.Context
	input  <-chan interface{}
	output chan interface{}
	log    *logrus.Entry
}

func NewStreamOp(ctx context.Context) *StreamOp {
	log, ok := autoctx.GetLogEntry(ctx)
	if !ok {
		log = logrus.WithField("Component", "StreamOperator")
		log.Error("Logger not found in context")
	}

	r := new(StreamOp)
	r.ctx = ctx
	r.log = log.WithFields(logrus.Fields{
		"Component": "StreamOperator",
		"Type":      fmt.Sprintf("%T", r),
	})

	r.output = make(chan interface{}, 1024)

	r.log.Infof("Component initialized")
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
			r.log.Info("Component shutting down")
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
