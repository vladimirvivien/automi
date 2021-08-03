package emitters

import (
	"context"
	"errors"
	"reflect"

	"github.com/vladimirvivien/automi/api"
	autoctx "github.com/vladimirvivien/automi/api/context"
	"github.com/vladimirvivien/automi/util"
)

// SliceEmitter is an emitter that takes in a slice and
// emits slice items individually as a stream.
type SliceEmitter struct {
	slice  interface{}
	output chan interface{}
	logf   api.LogFunc
}

// Slice creates new slice source
func Slice(slice interface{}) *SliceEmitter {
	return &SliceEmitter{
		slice:  slice,
		output: make(chan interface{}, 1024),
	}
}

// GetOutput returns the output channel of this source node
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
	s.logf = autoctx.GetLogFunc(ctx)
	util.Logfn(s.logf, "Opening slice emitter")
	sliceVal := reflect.ValueOf(s.slice)

	if !sliceVal.IsValid() {
		return errors.New("Invalid slice for SliceEmitter")
	}

	go func() {
		exeCtx, cancel := context.WithCancel(ctx)
		defer func() {
			util.Logfn(s.logf, "Slice emitter closing")
			cancel()
			close(s.output)
		}()
		for i := 0; i < sliceVal.Len(); i++ {
			val := sliceVal.Index(i)
			select {
			case s.output <- val.Interface():
			case <-exeCtx.Done():
				return
			}
		}
	}()
	return nil
}
