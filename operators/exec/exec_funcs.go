package exec

import (
	"context"

	"github.com/vladimirvivien/automi/api"
	"github.com/vladimirvivien/automi/funcs"
)

// Execute sets up an executor operator with the user-defined ExecFunc.
func Execute[IN, OUT any](f funcs.ExecFunction[IN, OUT]) *ExecOperator[IN, OUT] {
	return New(f)
}

// Filter applies the user-defined ExecFunc to filter out passed data.
// The provided function must be of type:
//
//	func(T)bool - where T is the type of incoming data item, bool is the value of the predicate
//
// If the function returns false, data item will not continue downstream.
func Filter[IN any](f funcs.ExecFunction[IN, bool]) *ExecOperator[IN, api.FilterItem[IN]] {
	wrapper := func(ctx context.Context, data IN) api.FilterItem[IN] {
		predicate := f(ctx, data)
		return api.FilterItem[IN]{Predicate: predicate, Item: data}
	}
	return New(wrapper)
}

// Map applies a user-defined ExecFunc to map incoming item to new value.
// The user-defined function must be of type:
//
//	func(T) R - where T is the incoming item, R is the type of the mapped item
func Map[IN any, OUT any](f func(context.Context, IN) OUT) *ExecOperator[IN, OUT] {
	return Execute(f)
}
