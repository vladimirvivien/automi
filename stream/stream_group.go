package stream

import (
	"fmt"
	"reflect"

	"github.com/vladimirvivien/automi/api"

	"golang.org/x/net/context"
)

// GroupBy groups elements- based on classification method specified
// by param g which can be one of the followings:
// * int - indicates positional element in a tuple, slice, or array,
//         other types are ignored.
// * string - indicates the name of a field in a struct or a map.
//          other types are ignored.
// * func () int - a function which returns int
// * func () string - a function which returns a string
// GroupBy is a reductive function which will collect upstream elements,
// partition them in a map based on above criteria, and returns the map
// once stream window is closed.
func (s *Stream) GroupBy(g interface{}) *Stream {
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

	operator := NewBinaryOp(s.ctx)
	operator.SetOperation(op)
	operator.SetInitialState(make(map[interface{}][]interface{}))
	s.ops = append(s.ops, operator)
	return s
}

// groupByInt expects incoming data as Pair, []slice, or [n]array.
// It creates the reduction operation and stores incoming data in a map.
func (s *Stream) groupByInt(i int64) api.BinFunc {
	op := api.BinFunc(func(ctx context.Context, op0, op1 interface{}) interface{} {
		stateType := reflect.TypeOf(op0)
		if stateType.Kind() != reflect.Map {
			panic("GroupBy expects a map[keytype][]slice for internal storage")
		}
		stateMap := reflect.ValueOf(op0)

		// save data according to type
		dataType := reflect.TypeOf(op1)
		dataVal := reflect.ValueOf(op1)
		switch dataType.Kind() {
		case reflect.Slice, reflect.Array:
			// build stateMap[key][]slice dynamically. Add item to slice.
			key := dataVal.Index(int(i))
			slice := stateMap.MapIndex(key)
			if !slice.IsValid() {
				slice = reflect.MakeSlice(stateType.Elem(), 0, 0)
				stateMap.SetMapIndex(key, slice)
			}

			// copy value to group in new slice
			for j := 0; j < dataVal.Len(); j++ {
				if j != int(i) {
					slice = reflect.Append(slice, dataVal.Index(j))
				}
			}
			stateMap.SetMapIndex(key, slice)
		default: // ignore anything else
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

	operator := NewBinaryOp(s.ctx)
	operator.SetOperation(op)
	operator.SetInitialState(make(map[interface{}][]interface{}))
	s.ops = append(s.ops, operator)
	return s
}
