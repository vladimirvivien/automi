package stream

import (
	"fmt"
	"reflect"

	"github.com/vladimirvivien/automi/api"

	"golang.org/x/net/context"
)

// isUnaryFuncForm ensures type is a function of form func(in)out.
func (s *Stream) isUnaryFuncForm(ftype reflect.Type) error {
	// enforce f with sig fn(in)out
	switch ftype.Kind() {
	case reflect.Func:
		if ftype.NumIn() != 1 {
			return fmt.Errorf("Function must take one parameter")
		}
		if ftype.NumOut() != 1 {
			return fmt.Errorf("Function must return one parameter")
		}
	default:
		return fmt.Errorf("Operation expects a function argument")
	}
	return nil
}

// Transform is the raw  method used to apply transfomrmative
// unary operations to stream elements (i.e. filter, map, etc)
// It is exposed for completeness, use the other more specific methods.
func (s *Stream) Transform(op api.UnOperation) *Stream {
	operator := NewUnaryOp(s.ctx)
	operator.SetOperation(op)
	s.ops = append(s.ops, operator)
	return s
}

// Process is used to express general processing operation of incoming stream
// elements.  Function must be of the form "func(input)output", anything else
// will cause painic.
func (s *Stream) Process(f interface{}) *Stream {
	fntype := reflect.TypeOf(f)
	if err := s.isUnaryFuncForm(fntype); err != nil {
		panic(fmt.Sprintf("Op Process() failed: %s", err))
	}

	fnval := reflect.ValueOf(f)

	op := api.UnFunc(func(ctx context.Context, data interface{}) interface{} {
		arg0 := reflect.ValueOf(data)
		result := fnval.Call([]reflect.Value{arg0})[0]
		return result.Interface()
	})

	return s.Transform(op)
}

// Filter takes a predicate func that filters the stream.
// Function must be of form "func(input)bool", anything else causes panic.
// If func returns true, current item continues downstream.
func (s *Stream) Filter(f interface{}) *Stream {
	fntype := reflect.TypeOf(f)
	if err := s.isUnaryFuncForm(fntype); err != nil {
		panic(fmt.Sprintf("Op Filter() failed :%s", err))
	}
	// ensure bool ret type
	if fntype.Out(0).Kind() != reflect.Bool {
		panic("Op Filter() must return a bool")
	}

	fnval := reflect.ValueOf(f)

	op := api.UnFunc(func(ctx context.Context, data interface{}) interface{} {
		arg0 := reflect.ValueOf(data)
		result := fnval.Call([]reflect.Value{arg0})[0]
		predicate := result.Bool()
		if !predicate {
			return nil
		}
		return data
	})
	return s.Transform(op)
}

// Map takes one value and maps it to another value.
// The map function must be of form "func(input)output",
// anything else causes panic.
func (s *Stream) Map(f interface{}) *Stream {
	return s.Process(f)
}

// FlatMap similar to Map, however, expected to return a slice of values
// to downstream operators.  The operator will flatten the slice and emit
// individually onto the stream. The FlatMap function must have the form
// "func(intput)[]output", anything else will be rejected
func (s *Stream) FlatMap(f interface{}) *Stream {
	fntype := reflect.TypeOf(f)
	if err := s.isUnaryFuncForm(fntype); err != nil {
		panic(fmt.Sprintf("Op FlatMap() failed: %s", err))
	}
	if fntype.Out(0).Kind() != reflect.Slice {
		panic(fmt.Sprintf("Op FlatMap() expects to return a slice of values"))
	}

	fnval := reflect.ValueOf(f)

	op := api.UnFunc(func(ctx context.Context, data interface{}) interface{} {
		arg0 := reflect.ValueOf(data)
		result := fnval.Call([]reflect.Value{arg0})[0]
		return result.Interface()
	})

	s.Transform(op) // add flatmap as unary op
	s.ReStream()    // add streamop to unpack flatmap result
	return s
}
