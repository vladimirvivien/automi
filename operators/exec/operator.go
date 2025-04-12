package exec

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/vladimirvivien/automi/api"
	autoctx "github.com/vladimirvivien/automi/api/context"
	"github.com/vladimirvivien/automi/funcs"
	"github.com/vladimirvivien/automi/log"
)

// ExecOperator is an operator node that can execute an arbitrary user-defined Go function
type ExecOperator[IN any, OUT any] struct {
	opFunc      funcs.ExecFunction[IN, OUT]
	concurrency int
	input       <-chan any
	output      chan any
	logf        api.StreamLogFunc
}

// New creates *Operator value
func New[IN, OUT any](f funcs.ExecFunction[IN, OUT]) *ExecOperator[IN, OUT] {
	// extract logger
	o := new(ExecOperator[IN, OUT])
	o.opFunc = f
	o.concurrency = 1
	o.output = make(chan any, 1024)
	o.logf = log.NoLogFunc

	return o
}

// SetConcurrency sets the concurrency level for the operation
func (o *ExecOperator[IN, OUT]) SetConcurrency(concurr int) {
	o.concurrency = concurr
	if o.concurrency < 1 {
		o.concurrency = 1
	}
}

// SetInput sets the input channel for the executor node
func (o *ExecOperator[IN, OUT]) SetInput(in <-chan any) {
	o.input = in
}

// GetOutput returns the output channel for the executor node
func (o *ExecOperator[IN, OUT]) GetOutput() <-chan any {
	return o.output
}

// SetLogFunc sets a function called to capture and log stream events
func (o *ExecOperator[IN, OUT]) SetLogFunc(f api.StreamLogFunc) {
	o.logf = f
}

// Exec is the entry point for the executor
func (o *ExecOperator[IN, OUT]) Exec(ctx context.Context) (err error) {
	if o.opFunc == nil {
		return fmt.Errorf("exec operator missing function")
	}

	if o.input == nil {
		return api.ErrInputChannelUndefined
	}

	o.logf(ctx, log.LogInfo(
		"Component starting",
		slog.String("operator", "Exec"),
	))

	go func() {
		defer func() {
			o.logf(ctx, log.LogInfo(
				"Component closing",
				slog.String("operator", "Exec"),
			))
			close(o.output)
		}()

		o.doOp(ctx)
	}()
	return nil
}

func (o *ExecOperator[IN, OUT]) doOp(ctx context.Context) {
	logCtx := autoctx.WithLogF(ctx, o.logf)
	exeCtx, cancel := context.WithCancel(logCtx)

	defer func() {
		cancel()
	}()

	for {
		select {
		// process incoming item
		case item, opened := <-o.input:
			if !opened {
				o.logf(ctx, log.LogDebug(
					"Component channel closed",
					slog.String("operator", "Exec"),
				))
				return
			}

			param0, ok := any(item).(IN)
			if !ok {
				o.logf(ctx, log.LogError(
					"Unexpected type for Func parameter",
					slog.String("operator", "Exec"),
					slog.String("type", fmt.Sprintf("%T", item)),
				))
				continue
			}
			result := o.opFunc(exeCtx, param0)

			switch val := any(result).(type) {
			case nil:
				continue
			case api.FilterItem[IN]:
				// apply filter predicate
				if val.Predicate {
					select {
					case o.output <- val.Item:
					case <-exeCtx.Done():
						return
					}
					continue
				}

			case api.StreamResult:
				item := val.Value
				err := val.Err
				action := val.Action

				// handle error
				if err != nil {
					o.logf(ctx, log.LogDebug(
						"Error: function execution",
						slog.String("operator", "Exec"),
						slog.String("error", err.Error()),
					))
				}
				switch action {
				case api.ActionSkipItem:
					continue
				case api.ActionRerouteItem:
					// Not implemented yet
				}

				select {
				case o.output <- item:
				case <-exeCtx.Done():
					return
				}

			default:
				select {
				case o.output <- result:
				case <-exeCtx.Done():
					return
				}
			}

		// is cancelling
		case <-exeCtx.Done():
			o.logf(ctx, log.LogDebug(
				"Component context canceled",
				slog.String("operator", "Exec"),
			))
			return
		}
	}
}
