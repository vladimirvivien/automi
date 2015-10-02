package context

import (
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

// WithAuxChan sets the auiliary channel for components using the
// specified context so they can export message/event that is not
// part of the main processing flow.  This is useful for pushing rejected
// processed items, system messages, or communicating events, etc.
// The Plan that is executing the flow will make the auxiliary channel
// available for downstream processing.
func WithAuxChan(ctx context.Context, auxChan chan<- interface{}) context.Context {
	return context.WithValue(ctx, auxChKey, auxChan)
}

// GetAuxChan returns the auxiliary channel used for communicating
// non-processing messaging or event to outside
func GetAuxChan(ctx context.Context) (chan<- interface{}, bool) {
	ch, ok := ctx.Value(auxChKey).(chan interface{})
	return ch, ok
}
