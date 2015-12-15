package api

import "golang.org/x/net/context"

type Operation interface {
	Apply(ctx context.Context, data interface{}) interface{}
}

type OpFunc func(context.Context, interface{}) interface{}

func (f OpFunc) Apply(ctx context.Context, data interface{}) interface{} {
	return f(ctx, data)
}
