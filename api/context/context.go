package context

import (
	"context"
	"fmt"
	"log"
	"os"
)

const (
	loggerKey = iota
	auxChKey
)

var (
	logger *log.Logger = log.New(os.Stderr, log.Prefix(), log.Flags())
)

// WithLogger sets a logger value in the cotext.
func WithLogger(ctx context.Context, logger *log.Logger) context.Context {
	return context.WithValue(ctx, loggerKey, logger)
}

// GetLogger returns the logger if one is found, or create one.
func GetLogger(ctx context.Context) *log.Logger {
	l, ok := ctx.Value(loggerKey).(*log.Logger)
	if l == nil || !ok {
		return logger
	}
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
