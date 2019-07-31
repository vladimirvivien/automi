package binary

import (
	"context"
	"fmt"
	"reflect"

	"github.com/vladimirvivien/automi/api"
)

// ReduceFunc returns a binary function which takes a user-defined accumulator
// function to apply reduction (fold) logic to incoming streaming items to
// return a single summary value.  The user-provided accumulator function must
// be of type:
//   func(S,T) R
//     where S is the partial result (initially the seed)
//     T is the streamed item from upstream
//     R is the calculated value which becomes partial result for next value
// It is important to understand that applying a reductive operator after an
// open-ended emitter (i.e. a network) may never end.  To force a Reduction function
// to terminate, it is sensible to place it after a batch operator for instance.
func ReduceFunc(f interface{}) (api.BinFunc, error) {
	fntype := reflect.TypeOf(f)
	if err := isBinaryFuncForm(fntype); err != nil {
		return nil, err
	}

	fnval := reflect.ValueOf(f)

	return api.BinFunc(func(ctx context.Context, op0, op1 interface{}) interface{} {
		arg0 := reflect.ValueOf(op0)
		arg1, arg1Type := reflect.ValueOf(op1), reflect.TypeOf(op1)
		if op0 == nil {
			arg0 = reflect.Zero(arg1Type)
		}
		result := fnval.Call([]reflect.Value{arg0, arg1})[0]
		return result.Interface()
	}), nil

}

func isBinaryFuncForm(ftype reflect.Type) error {
	// enforce ftype with sig fn(op1,op2)out
	switch ftype.Kind() {
	case reflect.Func:
		if ftype.NumIn() != 2 {
			return fmt.Errorf("binary function requires two params")
		}
		if ftype.NumOut() != 1 {
			return fmt.Errorf("binary func must return one param")
		}
	default:
		return fmt.Errorf("binary func must be of type func(S,T)R")
	}
	return nil
}
