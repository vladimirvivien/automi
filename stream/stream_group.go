package stream

import (
	"context"
	"fmt"
	"reflect"

	"github.com/vladimirvivien/automi/api"
	"github.com/vladimirvivien/automi/operators"
	"github.com/vladimirvivien/automi/operators/batch"
)

// GroupByPos groups incoming batched items by their position in a slice.
// The batch is expected to be of type [][]interface{}, a two dimension slice
// The method walks dimension 1 of slice and groups data in dim 2 into map[key][]value
// where key is the value found at pos for each row.
// This method should be called after a Batch operation or it will fail.
func (s *Stream) GroupByPos(pos int) *Stream {
	s.ops = append(s.ops, batch.GroupByPosFunc(pos))
	return s
}

// groupByName returns a binary function that groups streaming item
// by a struct.name or a map[name].  If the item is not a struct or map,
// it is ignored.
func (s *Stream) groupByName(name string) api.BinFunc {
	op := api.BinFunc(func(ctx context.Context, op0, op1 interface{}) interface{} {
		stateType := reflect.TypeOf(op0)
		if stateType.Kind() != reflect.Map {
			panic("GroupBy expects a map[keytype][]slice for internal storage")
		}
		stateMap := reflect.ValueOf(op0)

		// stream item data type and value
		dataType := reflect.TypeOf(op1)
		dataVal := reflect.ValueOf(op1)
		key := dataVal.FieldByName(name)

		if key.IsValid() {
			// ensure state map[K][]slice is ready
			slice := stateMap.MapIndex(key)
			if !slice.IsValid() {
				slice = reflect.MakeSlice(stateType.Elem(), 0, 0)
				stateMap.SetMapIndex(key, slice)
			}

			switch dataType.Kind() {
			// append struct.name = V to state map[name][]slice{V}
			case reflect.Struct:
				slice = reflect.Append(slice, dataVal)
				stateMap.SetMapIndex(key, slice)
			default:
			}
		}

		return stateMap.Interface()

	})

	return op
}

// SumBy is a convenient aggregation method that will add up
// numeric values stored in stream items.  SumBy expects specific
// type of data, from upstream, for it to work properly as outlined below.
//   * map[K]numeric - a map where the values are numeric elements
//   * tuple.KV - where kv[1] are numeric elements
//   * []slice, [n]array - where slice[i] or array[i] are
//     slices or arrays of numeric values
//   * struct{} - where the specified element is a slice or array of ints
// SumBy parameter can be of the following type:
//   * int - used as index when stream is slice or array, or tuple.KV
//   * "string" - the name of a map key with where the value is a slice or array
// SumBy emmits its result as map[key]int
func (s *Stream) SumBy(g interface{}) *Stream {
	gType := reflect.TypeOf(g)
	gVal := reflect.ValueOf(g)

	var op api.BinFunc
	switch gType.Kind() {
	case reflect.Int, reflect.Int16, reflect.Int32, reflect.Int64:
		idx := gVal.Int()
		op = s.groupByInt(idx)
	case reflect.String:
	case reflect.Func:
	default:
		panic(fmt.Sprintf("GroupBy failed, type %T is not a supported classifier", g))
	}

	operator := operators.NewBinaryOp(s.ctx)
	operator.SetOperation(op)
	operator.SetInitialState(make(map[interface{}][]interface{}))
	s.ops = append(s.ops, operator)
	return s
}
