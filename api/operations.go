package api

import "context"

type UnOperation interface {
	Apply(ctx context.Context, data interface{}) interface{}
}

type UnFunc func(context.Context, interface{}) interface{}

func (f UnFunc) Apply(ctx context.Context, data interface{}) interface{} {
	return f(ctx, data)
}

type BinOperation interface {
	Apply(ctx context.Context, op1, op2 interface{}) interface{}
}

type BinFunc func(context.Context, interface{}, interface{}) interface{}

func (f BinFunc) Apply(ctx context.Context, op1, op2 interface{}) interface{} {
	return f(ctx, op1, op2)
}
