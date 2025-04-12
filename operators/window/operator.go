package window

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/vladimirvivien/automi/api"
	autoctx "github.com/vladimirvivien/automi/api/context"
	"github.com/vladimirvivien/automi/log"
)

// WindowOperator is an operator node that batches incoming streamed items based
// on provided criteria.
type WindowOperator[IN any] struct {
	trigger TriggerFunction[IN]
	input   <-chan any
	output  chan any
	logf    api.StreamLogFunc
}

// New returns a new *WindowOperator
func New[IN any](trigger TriggerFunction[IN]) *WindowOperator[IN] {
	return &WindowOperator[IN]{
		trigger: trigger,
		output:  make(chan interface{}, 1024),
		logf:    log.NoLogFunc,
	}
}

// SetInput sets the input channel for the operator node
func (op *WindowOperator[IN]) SetInput(in <-chan any) {
	op.input = in
}

// GetOutput returns the output channel of the operator node
func (op *WindowOperator[IN]) GetOutput() <-chan any {
	return op.output
}

// SetLogFunc sets a function called to capture and log stream events
func (o *WindowOperator[IN]) SetLogFunc(f api.StreamLogFunc) {
	o.logf = f
}

// Exec is the exstarting point of the operator node.
func (op *WindowOperator[IN]) Exec(ctx context.Context) (err error) {
	op.logf(ctx, log.LogDebug(
		"Component starting",
		slog.String("operator", "Window"),
	))

	if op.input == nil {
		err = api.ErrInputChannelUndefined
		return
	}

	go func() {
		var itemWindow []IN
		logCtx := autoctx.WithLogF(ctx, op.logf)
		exeCtx, cancel := context.WithCancel(logCtx)
		operatorStartTime := time.Now()
		operatorItemCount := uint64(0)

		defer func() {
			op.logf(ctx, log.LogDebug(
				"Component closing",
				slog.String("operator", "Window"),
			))

			if len(itemWindow) > 0 {
				select {
				case op.output <- itemWindow:
				case <-exeCtx.Done():
				}
			}

			cancel()
			close(op.output)
		}()

		// default to TriggerAll.
		if op.trigger == nil {
			op.trigger = TriggerAllFunc[IN]()
		}
		windowStartTime := time.Now()
		windowItemCount := uint64(1)

		for {
			select {
			case item, opened := <-op.input:
				if !opened {
					return
				}

				operatorItemCount++

				itemVal, ok := item.(IN)
				if !ok {
					op.logf(ctx, log.LogDebug(
						"Error: unexpected data type",
						slog.String("operator", "Window"),
						slog.String("type", fmt.Sprintf("%T", item)),
					))
					continue
				}

				itemWindow = append(itemWindow, itemVal)

				// apply batch trigger function
				done := false
				if op.trigger != nil {
					done = op.trigger(ctx, WindowContext[IN]{
						OperatorStartTime: operatorStartTime,
						OperatorItemCount: operatorItemCount,
						WindowStartTime:   windowStartTime,
						WindowItemCount:   windowItemCount,
						Item:              itemVal,
						ItemWindowTime:    time.Now(),
					})
				}
				if !done {
					windowItemCount++
					continue
				}

				// done batching, output downstream
				select {
				case op.output <- itemWindow:
					// reset window
					windowStartTime = time.Now()
					windowItemCount = 1
					itemWindow = make([]IN, 0)
				case <-exeCtx.Done():
					return
				}

			case <-exeCtx.Done():
				return
			}
		}
	}()
	return nil
}
