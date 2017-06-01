package emitters

import (
	"context"
	"errors"
	"log"
	"reflect"

	autoctx "github.com/vladimirvivien/automi/api/context"
)

// SliceEmitter is an emitter that takes in a slice and
// emits slice items individually as a stream.
type SliceEmitter struct {
	slice  interface{}
	output chan interface{}
	log    *log.Logger
}

// SliceSrc creates new slice source
func Slice(slice interface{}) *SliceEmitter {
	return &SliceEmitter{
		slice:  slice,
		output: make(chan interface{}, 1024),
	}
}

// GetOuptut returns the output channel of this source node
func (s *SliceEmitter) GetOutput() <-chan interface{} {
	return s.output
}

// Open opens the source node to start streaming data on its channel
func (s *SliceEmitter) Open(ctx context.Context) error {
	// ensure slice param is a slice
	sliceType := reflect.TypeOf(s.slice)
	if sliceType.Kind() != reflect.Slice {
		return errors.New("SliceEmitter requires slice")
	}
	s.log = autoctx.GetLogger(ctx)
	s.log.Print("opening slice emitter")
	sliceVal := reflect.ValueOf(s.slice)

	if !sliceVal.IsValid() {
		return errors.New("invalid slice for SliceEmitter")
	}

	go func() {
		defer func() {
			close(s.output)
			s.log.Println("closing slice emitter")
		}()
		for i := 0; i < sliceVal.Len(); i++ {
			val := sliceVal.Index(i)
			s.output <- val.Interface()
		}
	}()
	return nil
}
