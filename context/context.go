package context

import (
	"fmt"

	"github.com/Sirupsen/logrus"
	"golang.org/x/net/context"
)

const (
	logEntryKey = iota
	auxChKey
)

// WithLogEntry sets a Logrus Entry struct in the context
// to be used for common logging infrastructure
func WithLogEntry(ctx context.Context, logEntry *logrus.Entry) context.Context {
	return context.WithValue(ctx, logEntryKey, logEntry)
}

// GetLogEntry returns the Logrus Entry to use for logging
func GetLogEntry(ctx context.Context) (*logrus.Entry, bool) {
	l, ok := ctx.Value(logEntryKey).(*logrus.Entry)
	return l, ok
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
	go func() {
		ch <- item
	}()
	return nil
}

// GetAuxChan returns the auxiliary channel used for communicating
// non-processing messaging or event to outside of the flow.
func GetAuxChan(ctx context.Context) (chan<- interface{}, bool) {
	ch, ok := ctx.Value(auxChKey).(chan<- interface{})
	return ch, ok
}
