package funcs

import "context"

// ExecFunction represents a user-defined function that is executed
// by an Executor operator
type ExecFunction[IN any, OUT any] func(context.Context, IN) OUT
type ExecFuncWithErr[IN any, OUT any] func(context.Context, IN) (OUT, error)
