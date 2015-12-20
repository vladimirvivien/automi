package api

import "golang.org/x/net/context"

type UnaryOp interface {
	Apply(ctx context.Context, data interface{}) interface{}
}

type OpFunc func(context.Context, interface{}) interface{}

func (f OpFunc) Apply(ctx context.Context, data interface{}) interface{} {
	return f(ctx, data)
}

type BiOperation interface {
	Apply(ctx context.Context, op1, op2 interface{}) interface{}
}
