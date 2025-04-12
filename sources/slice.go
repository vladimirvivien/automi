package sources

import (
	"context"
	"log/slog"

	"github.com/vladimirvivien/automi/api"
	"github.com/vladimirvivien/automi/log"
)

// SliceSource is an emitter that gets data from a slice source S []T and
// emits slice items T individually as a stream.
type SliceSource[OUT []ITEM, ITEM any] struct {
	slice  OUT
	output chan any
	logf   api.StreamLogFunc
}

// Slice creates new slice emitter
func Slice[OUT []ITEM, ITEM any](slice OUT) *SliceSource[OUT, ITEM] {
	return &SliceSource[OUT, ITEM]{
		slice:  slice,
		output: make(chan any, 1024),
		logf:   log.NoLogFunc,
	}
}

// GetOutput returns the output channel of this emitter
func (s *SliceSource[OUT, ITEM]) GetOutput() <-chan any {
	return s.output
}

// SetLogFunc sets logging func for component
func (s *SliceSource[OUT, ITEM]) SetLogFunc(f api.StreamLogFunc) {
	s.logf = f
}

// Open reads the source and start streaming data to the emitter
func (s *SliceSource[OUT, ITEM]) Open(ctx context.Context) error {
	s.logf(ctx, log.LogInfo(
		"Component starting",
		slog.String("source", "Slice"),
	))

	go func() {
		exeCtx, cancel := context.WithCancel(ctx)
		defer func() {
			s.logf(ctx, log.LogInfo(
				"Component closing",
				slog.String("source", "Slice"),
			))
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
