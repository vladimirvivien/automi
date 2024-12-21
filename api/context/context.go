package context

import (
	"context"

	"github.com/vladimirvivien/automi/api"
	"github.com/vladimirvivien/automi/log"
)

type ctxKey int

var (
	logFuncKey ctxKey = 1
)

// WithLogF sets a function to handle logging from data stream components
func WithLogF(ctx context.Context, logFunc api.StreamLogFunc) context.Context {
	return context.WithValue(ctx, logFuncKey, logFunc)
}

// GetLogF returns the log function stored in the context.
func GetLogF(ctx context.Context) api.StreamLogFunc {
	fn, ok := ctx.Value(logFuncKey).(api.StreamLogFunc)
	if !ok {
		return log.NoLogFunc
	}
	return fn
}

// LogF is convenient shortcut that retieves a saved log function
// and uses to log and event.
func LogF(ctx context.Context, log api.StreamLog) {
	GetLogF(ctx)(ctx, log)
}
