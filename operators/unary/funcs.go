package unary

import (
	"context"
	"fmt"
	"reflect"

	"github.com/vladimirvivien/automi/api"
)

// ProcessFunc returns a unary function which applies the specified
// user-defined function that processes data items from upstream and
// returns a result value. The provided function must be of type:
//   func(T) R
//   where T is the type of incoming item
//   R the type of returned processed item
func ProcessFunc(f interface{}) (api.UnFunc, error) {
	fntype := reflect.TypeOf(f)
	if err := isUnaryFuncForm(fntype); err != nil {
		return nil, err
	}

	fnval := reflect.ValueOf(f)

	return api.UnFunc(func(ctx context.Context, data interface{}) interface{} {
		arg0 := reflect.ValueOf(data)
		result := fnval.Call([]reflect.Value{arg0})[0]
		return result.Interface()
	}), nil
}

// FilterFunc returns a unary function (api.UnFunc) which applies the user-defined
// filtering to apply predicates that filters out data items from being included
// in the downstream.  The provided user-defined function must be of type:
//   func(T)bool - where T is the type of incoming data item, bool is the value of the predicate
// When the user-defined function returns false, the current processed data item will not
// be placed in the downstream processing.
func FilterFunc(f interface{}) (api.UnFunc, error) {
	fntype := reflect.TypeOf(f)
	if err := isUnaryFuncForm(fntype); err != nil {
		return nil, err
	}
	// ensure bool ret type
	if fntype.Out(0).Kind() != reflect.Bool {
		panic("Filter function must return a bool type")
	}

	fnval := reflect.ValueOf(f)

	return api.UnFunc(func(ctx context.Context, data interface{}) interface{} {
		arg0 := reflect.ValueOf(data)
		result := fnval.Call([]reflect.Value{arg0})[0]
		predicate := result.Bool()
		if !predicate {
			return nil
		}
		return data
	}), nil
}

// MapFunc returns an unary function which applies the user-defined function which
// maps, one-to-one, the incomfing value to a new value.  The user-defined function
// must be of type:
//   func(T) R - where T is the incoming item, R is the type of the returned mapped item
func MapFunc(f interface{}) (api.UnFunc, error) {
	return ProcessFunc(f)
}

// FlatMapFunc returns an unary function which applies a user-defined function which
// takes incoming comsite items and deconstruct them into individual items which can
// then be re-streamed.  The type for the user-defined function is:
//   func (T) R - where R is the original item, R is a slice of decostructed items
// The slice returned should be restreamed by placing each item onto the stream for
// downstream processing.
func FlatMapFunc(f interface{}) (api.UnFunc, error) {
	fntype := reflect.TypeOf(f)
	if err := isUnaryFuncForm(fntype); err != nil {
		return nil, err
	}
	if fntype.Out(0).Kind() != reflect.Slice {
		return nil, fmt.Errorf("FlatMap function must return a slice")
	}

	fnval := reflect.ValueOf(f)

	return api.UnFunc(func(ctx context.Context, data interface{}) interface{} {
		arg0 := reflect.ValueOf(data)
		result := fnval.Call([]reflect.Value{arg0})[0]
		return result.Interface()
	}), nil
}

// isUnaryFuncForm ensures type is a function of form func(in)out.
func isUnaryFuncForm(ftype reflect.Type) error {
	// enforce f with sig fn(in)out
	switch ftype.Kind() {
	case reflect.Func:
		if ftype.NumIn() != 1 {
			return fmt.Errorf("function requires one parameter")
		}
		if ftype.NumOut() != 1 {
			return fmt.Errorf("function must exactly one parameter")
		}
	default:
		return fmt.Errorf("unary operator expects a function argument")
	}
	return nil
}
