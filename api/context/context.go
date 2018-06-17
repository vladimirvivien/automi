package context

import (
	"context"
	"fmt"

	"github.com/go-faces/logger"
)

type ctxKey int

var (
	logKey   ctxKey = 1
	auxChKey ctxKey = 2
)

// WithLogger sets an interface.logger value in context
func WithLogger(ctx context.Context, log logger.Interface) context.Context {
	return context.WithValue(ctx, logKey, log)
}

// GetLogger returns a logger.Interface from provided context.
func GetLogger(ctx context.Context) logger.Interface {
	l, _ := ctx.Value(logKey).(logger.Interface)
	return l
}

// WithAuxChan sets the auiliary channel to be used by components using the
// so they can export message/event that is not
// part of the main processing flow.
func WithAuxChan(ctx context.Context, auxChan chan<- interface{}) context.Context {
	return context.WithValue(ctx, auxChKey, auxChan)
}

// SendAuxMsg submits an item to be sent to the auxiliary channel.
// The item can be any arbitrary value that can be used for non-processing
// messaging such as an event, rejected data, etc.  These messages can be
// processed using the Plan that is running the flow.
func SendAuxMsg(ctx context.Context, item interface{}) error {
	ch, ok := ctx.Value(auxChKey).(chan<- interface{})
	if !ok {
		return fmt.Errorf("Unable to find the auxiliary channel in context")
	}
	ch <- item
	return nil
}

// GetAuxChan returns the auxiliary channel used for communicating
// non-processing messaging or event to outside of the flow.
func GetAuxChan(ctx context.Context) (chan<- interface{}, bool) {
	ch, ok := ctx.Value(auxChKey).(chan<- interface{})
	return ch, ok
}
