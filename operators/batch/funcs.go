package batch

import (
	"context"
	"reflect"

	"github.com/vladimirvivien/automi/api"
)

// GroupByPosFunc generates a Unary Function that groups incoming batched items
// by their position in a slice. The batch is expected to be of type [][]interface{},
// a two-dimensional slice where the second dimension is a alice of data.
//
// The operation does the following:
//  * walks dimension 1 (D-1) of slice
//  * for row in D-1, select item at index pos as key
//  * group slice in D-2 unto map[key][]interface{}
// This method should be called after a Batch operation or it will fail.
func GroupByPosFunc(pos int) api.UnFunc {
	return api.UnFunc(func(ctx context.Context, param0 interface{}) interface{} {
		dataType := reflect.TypeOf(param0)
		dataVal := reflect.ValueOf(param0)

		// validate expected type
		if dataType.Kind() != reflect.Slice && dataType.Kind() != reflect.Array {
			return param0 // ignores the data
		}

		group := make(map[interface{}][]interface{})

		// walk dimension 1 of [][]interface{}
		for i := 0; i < dataVal.Len(); i++ {
			row := dataVal.Index(i)
			switch row.Type().Kind() {
			case reflect.Slice, reflect.Array:
				key := row.Index(pos)
				if key.IsValid() {
					// walk dim 2, grab all items in slice, and map in group
					for j := 0; j < row.Len(); j++ {
						if j != pos {
							group[key.Interface()] = append(
								group[key.Interface()],
								row.Index(j),
							)
						}
					}
				}
			}
		}
		return group
	})
}
