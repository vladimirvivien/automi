package context

import (
	"github.com/Sirupsen/logrus"
	"golang.org/x/net/context"
)

const (
	logEntryKey = iota
)

func WithLogEntry(ctx context.Context, logEntry *logrus.Entry) context.Context {
	return context.WithValue(ctx, logEntryKey, logEntry)
}

func GetLogEntry(ctx context.Context) (*logrus.Entry, bool) {
	l, ok := ctx.Value(logEntryKey).(*logrus.Entry)
	return l, ok
}
