package stream

import (
	"fmt"
	"reflect"

	"golang.org/x/net/context"
)

// GroupBy groups elements based on classification method specified
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

	var op BinFunc
	switch gType.Kind() {
	case reflect.Int, reflect.Int16, reflect.Int32, reflect.Int64:
		idx := gVal.Int()
		op = s.groupByInt(idx)
	case reflect.String:
	case reflect.Func:
	default:
		panic(fmt.Sprintf("GroupBy failed, type %T is not a supported classifier", g))
	}
	s.Accumulate(op)
	return s
}

// groupByInt expects incoming data as Pair, []slice, or [n]array.
// It creates the reduction operation and stores incoming data in a map.
func (s *Stream) groupByInt(i int64) BinFunc {
	op := BinFunc(func(ctx context.Context, op0, op1 interface{}) interface{} {
		stateType := reflect.TypeOf(op0)
		if stateType.Kind() != reflect.Map {
			panic("GroupBy expects a map for internal storage") // should never happen
		}
		stateMap := reflect.ValueOf(op0)

		// save data according to type
		dataType := reflect.TypeOf(op1)
		dataVal := reflect.ValueOf(op1)
		idxVal := dataVal.Index(int(i)) //key
		switch dataType.Kind() {
		case reflect.Slice, reflect.Array:
			if dataType.Name() == "tuple.KV" {
				stateMap.SetMapIndex(dataVal.Index(0), dataVal.Index(1))
			} else {
				stateMap.SetMapIndex(idxVal, dataVal)
			}
		default: // ignore anything else
		}

		return stateMap.Interface()
	})

	return op
}
