package stream

import (
	"fmt"
	"reflect"
	"sync"

	"golang.org/x/net/context"

	"github.com/Sirupsen/logrus"
	"github.com/vladimirvivien/automi/api/tuple"
)

type RestreamOp struct {
	ctx       contex.Context
	input     <-chan interface{}
	output    chan interface{}
	log       *logrus.Entry
	cancelled bool
	mutex     sync.RWMutex
}

func NewRestreamOp(ctx context.Context) *RestreamOp {
	log, ok := autoctx.GetLogEntry(ctx)
	if !ok {
		log = logrus.WithField("Component", "UnaryOperator")
		log.Error("Logger not found in context")
	}

	r := new(RestreamOp)
	r.ctx = ctx
	r.log = log.WithFields(logrus.Fields{
		"Component": "RestreamOperator",
		"Type":      fmt.Sprintf("%T", o),
	})

	r.concurrency = 1
	r.output = make(chan interface{}, 1024)

	r.log.Infof("Component initialized")
	return r
}

func (r *RestreamOp) SetInput(in <-chan interface{}) {
	r.input = in
}

func (r *RestreamOp) GetOutput() <-chan interface{} {
	return r.output
}

func (r *RestreamOp) Exec() {
	if r.input == nil {
		err = fmt.Errorf("No input channel found")
		return
	}

	go func() {
		defer func() {
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
						r.output <- j.Interface{}
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
}
