package api

import "context"

// UnOperation interface represents unary operations (i.e. Map, Filter, etc)
type UnOperation interface {
	Apply(ctx context.Context, data interface{}) interface{}
}

// UnFunc implements UnOperation as type func (context.Context, interface{})
type UnFunc func(context.Context, interface{}) interface{}

// Apply implements UnOperation.Apply method
func (f UnFunc) Apply(ctx context.Context, data interface{}) interface{} {
	return f(ctx, data)
}

// BinOperation interface represents binary opeartions (i.e. Reduce, etc)
type BinOperation interface {
	Apply(ctx context.Context, op1, op2 interface{}) interface{}
}

// BinFunc implements BinOperation as type func(context.Context, interface{}, interface{})
type BinFunc func(context.Context, interface{}, interface{}) interface{}

// Apply implements BinOpeartion.Apply
func (f BinFunc) Apply(ctx context.Context, op1, op2 interface{}) interface{} {
	return f(ctx, op1, op2)
}
