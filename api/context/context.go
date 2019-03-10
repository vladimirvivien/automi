package context

import (
	"context"
	"fmt"

	"github.com/vladimirvivien/automi/api"
)

type ctxKey int

var (
	logFuncKey ctxKey = 1
	errFuncKey ctxKey = 2
)

// WithLogFunc sets the function to handle logging from runtime components
func WithLogFunc(ctx context.Context, logFunc api.LogFunc) context.Context {
	return context.WithValue(ctx, logFuncKey, logFunc)
}

// GetLogFunc returns the log function stored in the context.
func GetLogFunc(ctx context.Context) func(interface{}) {
	fn, ok := ctx.Value(logFuncKey).(func(interface{}))
	if !ok {
		return nil
	}
	return fn
}

// Log retrieves Log Function from context and invokes it with message
func Log(ctx context.Context, message interface{}) error {
	val := ctx.Value(logFuncKey)
	if val == nil {
		return fmt.Errorf("log function nil")
	}
	fn, ok := val.(api.LogFunc)
	if !ok {
		return fmt.Errorf("unexpected log func type %T", fn)
	}
	if fn != nil {
		fn(message)
	}
	return nil
}

// WithErrorFunc sets the function to handle error from runtime components
func WithErrorFunc(ctx context.Context, errFunc api.ErrorFunc) context.Context {
	return context.WithValue(ctx, errFuncKey, errFunc)
}

// GetErrFunc returns the error handling function stored in the context.
func GetErrFunc(ctx context.Context) api.ErrorFunc {
	fn, ok := ctx.Value(errFuncKey).(api.ErrorFunc)
	if !ok {
		return nil
	}
	return fn
}

// HandleErr used to invoke registered error handler to handle error.
func HandleErr(ctx context.Context, err error) error {
	val := ctx.Value(errFuncKey)
	if val == nil {
		return fmt.Errorf("error function nil")
	}
	fn, ok := val.(api.LogFunc)
	if !ok {
		return fmt.Errorf("unexpected error func type %T", fn)
	}
	if fn != nil {
		fn(err)
	}
	return nil
}

func Err(fn api.ErrorFunc, err api.StreamError) {
	if fn != nil {
		fn(err)
	}
}
