package emitters

import (
	"context"

	"github.com/vladimirvivien/automi/api"
	"github.com/vladimirvivien/automi/util"
)

// SliceEmitter is an emitter that gets data from a slice source S []T and
// emits slice items T individually as a stream.
type SliceEmitter[T any, S []T] struct {
	slice  S
	output chan T
	logf   api.LogFunc
}

// Slice creates new slice emitter
func Slice[T any, S []T](slice S) *SliceEmitter[T, S] {
	return &SliceEmitter[T, S]{
		slice:  slice,
		output: make(chan T, 1024),
	}
}

// GetOutput returns the output channel of this emitter
func (s *SliceEmitter[T, S]) GetOutput() <-chan T {
	return s.output
}

// Open reads the source and start streaming data to the emitter
func (s *SliceEmitter[T, S]) Open(ctx context.Context) error {
	go func() {
		exeCtx, cancel := context.WithCancel(ctx)
		defer func() {
			util.Logfn(s.logf, "Slice emitter closing")
			cancel()
			close(s.output)
		}()
		for _, val := range s.slice {
			select {
			case s.output <- val:
			case <-exeCtx.Done():
				return
			}
		}
	}()
	return nil
}
