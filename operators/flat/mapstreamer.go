package flat

import (
	"context"
	"log/slog"
	"reflect"

	"github.com/vladimirvivien/automi/api"
	autoctx "github.com/vladimirvivien/automi/api/context"
	"github.com/vladimirvivien/automi/api/tuple"
	"github.com/vladimirvivien/automi/log"
)

// MapStreamer is an operator that flattens or re-streams bulk mapped items into single items.
// Incoming items are expected to be of types map[K]V -> stream as tuple.Pair{K,V}.
type MapStreamer[IN map[KEY]ITEM, KEY comparable, ITEM any] struct {
	input  <-chan any
	output chan any
	logf   api.StreamLogFunc
}

// New creates a streamer operator
func NewMapStreamer[IN map[KEY]ITEM, KEY comparable, ITEM any]() *MapStreamer[IN, KEY, ITEM] {
	r := new(MapStreamer[IN, KEY, ITEM])
	r.output = make(chan any, 1024)
	r.logf = log.NoLogFunc
	return r
}

// SetInput sets the input channel for the executor node
func (r *MapStreamer[IN, KEY, ITEM]) SetInput(in <-chan any) {
	r.input = in
}

// GetOutput returns the output channel of the executer node
func (r *MapStreamer[IN, KEY, ITEM]) GetOutput() <-chan any {
	return r.output
}

func (r *MapStreamer[IN, KEY, ITEM]) SetLogFunc(f api.StreamLogFunc) {
	r.logf = f
}

// Exec is the execution starting point for the executor node.
func (r *MapStreamer[IN, KEY, ITEM]) Exec(ctx context.Context) (err error) {
	r.logf(ctx, log.LogInfo(
		"Component starting",
		slog.String("operator", "MapStreamer"),
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
				slog.String("operator", "MapStreamer"),
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

				case map[KEY]ITEM:
					items, ok := any(bundle).(map[KEY]ITEM)
					if !ok {
						r.logf(ctx, log.LogError(
							"Failed to convert item to slice",
							slog.String("operator", "MapStreamer"),
							slog.String("type", reflect.TypeOf(bundle).String()),
						))
						continue
					}
					for key, value := range items {
						select {
						case r.output <- tuple.Pair[KEY, ITEM]{Val1: key, Val2: value}:
						case <-exeCtx.Done():
							return
						}
					}

				// If item is not a map, send it as is
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
