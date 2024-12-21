package flat

import (
	"context"
	"log/slog"
	"reflect"

	"github.com/vladimirvivien/automi/api"
	autoctx "github.com/vladimirvivien/automi/api/context"
	"github.com/vladimirvivien/automi/log"
)

// SliceStreamer is an operator that flatten or re-streams bulk slice items into single items.
// Incoming items are expected to be of types []T -> restreamed as T.
type SliceStreamer[IN []ITEM, ITEM any] struct {
	input  <-chan any
	output chan any
	logf   api.StreamLogFunc
}

// New creates a *Operator value
func NewSliceStreamer[IN []ITEM, ITEM any]() *SliceStreamer[IN, ITEM] {
	r := new(SliceStreamer[IN, ITEM])
	r.output = make(chan any, 1024)
	r.logf = log.NoLogFunc
	return r
}

// SetInput sets the input channel for the executor node
func (r *SliceStreamer[IN, ITEM]) SetInput(in <-chan any) {
	r.input = in
}

// GetOutput returns the output channel of the executer node
func (r *SliceStreamer[IN, ITEM]) GetOutput() <-chan any {
	return r.output
}

func (r *SliceStreamer[IN, ITEM]) SetLogFunc(f api.StreamLogFunc) {
	r.logf = f
}

// Exec is the execution starting point for the executor node.
func (r *SliceStreamer[IN, ITEM]) Exec(ctx context.Context) (err error) {
	r.logf(ctx, log.LogInfo(
		"Component starting",
		slog.String("operator", "SliceStreamer"),
	))

	if r.input == nil {
		err = api.ErrInputChannelUndefined
		return
	}

	go func() {
		logCtx := autoctx.WithLogF(ctx, r.logf)
		exeCtx, cancel := context.WithCancel(logCtx)
		defer func() {
			r.logf(ctx, log.LogInfo(
				"Component closing",
				slog.String("operator", "SliceStreamer"),
			))
			cancel()
			close(r.output)
		}()

		for {
			select {
			case bundle, opened := <-r.input:
				if !opened {
					return
				}

				// unpack array, slice, map into individual item stream
				switch any(bundle).(type) {

				case []ITEM:
					items, ok := any(bundle).([]ITEM)
					if !ok {
						r.logf(ctx, log.LogError(
							"Failed to convert item to slice",
							slog.String("operator", "SliceStreamer"),
							slog.String("type", reflect.TypeOf(bundle).String()),
						))
						continue
					}
					for _, item := range items {
						select {
						case r.output <- item:
						case <-exeCtx.Done():
							return
						}
					}

				// If item is not a slice, send it as is
				default:
					select {
					case r.output <- bundle:
					case <-exeCtx.Done():
						return
					}
				}
			case <-exeCtx.Done():
				return
			}
		}
	}()
	return nil
}
