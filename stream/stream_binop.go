package stream

import (
	"context"
	"fmt"
	"reflect"

	"github.com/vladimirvivien/automi/api"
	"github.com/vladimirvivien/automi/operators"
)

// isBinaryFuncForm ensures type is a function of form func(op1,op2)out.
func (s *Stream) isBinaryFuncForm(ftype reflect.Type) error {
	// enforce ftype with sig fn(op1,op2)out
	switch ftype.Kind() {
	case reflect.Func:
		if ftype.NumIn() != 2 {
			return fmt.Errorf("Function must take 2 parameters")
		}
		if ftype.NumOut() != 1 {
			return fmt.Errorf("Function must return one parameter")
		}
	default:
		return fmt.Errorf("Operation expects a function argument")
	}
	return nil
}

// Accumulate is the raw base  method used to apply  reductive
// operations to stream elements (i.e. reduce, collect, etc).
// Use the other more specific methods instead.
func (s *Stream) Accumulate(op api.BinOperation) *Stream {
	operator := operators.NewBinaryOp(s.ctx)
	operator.SetOperation(op)
	s.ops = append(s.ops, operator)
	return s
}

// SetInitialState sets the initial value to be used as op1 argument
// when applying the binary operator.  This should be called right
// after a reductive operation (i.e. stream.Reduce(fn()).SetInitialState(val)).  It
// finds the last operator and apply the initial state to it.  If last operator is not
// a binary operator operation is ignored.
func (s *Stream) SetInitialState(val interface{}) *Stream {
	lastOp := s.ops[len(s.ops)-1]
	binOp, ok := lastOp.(*operators.BinaryOp)
	if !ok {
		s.log.Print("Unable to SetInitialState on last operator, wrong type. Value not set.")
		return s
	}
	binOp.SetInitialState(val)
	return s
}

// Reduce accumulates and reduce a stream of elements into a single value
func (s *Stream) Reduce(f interface{}) *Stream {
	fntype := reflect.TypeOf(f)
	if err := s.isBinaryFuncForm(fntype); err != nil {
		panic(fmt.Sprintf("Op Reduce() failed: %s", err))
	}

	fnval := reflect.ValueOf(f)

	op := api.BinFunc(func(ctx context.Context, op1, op2 interface{}) interface{} {
		arg0 := reflect.ValueOf(op1)
		arg1, arg1Type := reflect.ValueOf(op2), reflect.TypeOf(op2)
		if op1 == nil {
			arg0 = reflect.Zero(arg1Type)
		}
		result := fnval.Call([]reflect.Value{arg0, arg1})[0]
		return result.Interface()
	})

	return s.Accumulate(op)
}
