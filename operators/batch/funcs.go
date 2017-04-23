package batch

import (
	"context"
	"reflect"
	"strings"

	"github.com/vladimirvivien/automi/api"
)

// GroupByPosFunc generates a Unary Function that groups incoming batched items
// by their position in a slice. The batch is expected to be of type [][]T,
// a two-dimensional slice where the second dimension is a slice of data.
//
// The operation does the following:
//  * walks dimension 1 (D-1) of slice
//  * for each item, select value at index pos as key
//  * group slice in D-2 unto map[key][]T
//  * return map as result
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
			default: // TODO handle type mismatch
			}
		}
		return group
	})
}

func SumIntsByPosFunc(pos int) api.UnFunc {
	return api.UnFunc(func(ctx context.Context, param0 interface{}) interface{} {
		dataType := reflect.TypeOf(param0)
		dataVal := reflect.ValueOf(param0)

		// validate expected type
		if dataType.Kind() != reflect.Slice && dataType.Kind() != reflect.Array {
			return param0 // ignores the data
		}

		sum := int64(0)

		// walk [][]interface{}
		for i := 0; i < dataVal.Len(); i++ {
			row := dataVal.Index(i)
			switch row.Type().Kind() {
			case reflect.Slice, reflect.Array:
				posVal := row.Index(pos)
				if posVal.IsValid() {
					switch posVal.Type().Kind() {
					case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
						sum += posVal.Int()
					case reflect.Interface:
						sum += posVal.Elem().Int()
					default:
					}
				}
			default: // TODO handle type mismatch
			}
		}
		return sum
	})

}

func SumFloatsByPosFunc(pos int) api.UnFunc {
	return nil
}

// GroupByNameFunc generates an api.UnFunc that groups incoming batched items
// by field name.  The batched data is expected to be in the following formats:
// * []struct{} - slice of structs
// The operation walks the slice and for each item
// * it looks for field that matches name
// * if a match is found, it groups struct into a map[field_value][]T
// * return map as result
func GroupByNameFunc(name string) api.UnFunc {
	return api.UnFunc(func(ctx context.Context, param0 interface{}) interface{} {
		dataType := reflect.TypeOf(param0)
		dataVal := reflect.ValueOf(param0)

		// validate expected type
		if dataType.Kind() != reflect.Slice && dataType.Kind() != reflect.Array {
			return param0 // ignores the data
		}
		name = strings.Title(name) // avoid unexported field panic
		group := make(map[interface{}][]interface{})

		// walk the slice
		for i := 0; i < dataVal.Len(); i++ {
			item := dataVal.Index(i)
			switch item.Type().Kind() {
			case reflect.Struct:
				key := item.FieldByName(name)
				if key.IsValid() {
					group[key.Interface()] = append(
						group[key.Interface()], item.Interface(),
					)
				}
			default: //TODO handle type mismatch
			}
		}
		return group
	})
}

func SumIntsByNameFunc(name string) api.UnFunc {
	return nil
}

func SumFloatsByNameFunc(name string) api.UnFunc {
	return nil
}

// GroupByKeyFunc generates an api.UnFunc that groups incoming batched items
// by key value.  The batched data is expected to be in the following formats:
// * []map[T]T - slice of maps
// The operation walks the slice and for each item
// * it looks for key that matches the passed key
// * if a match is found, it groups the map value into another internal map[key][]T
// * the internal map is returned as result
func GroupByKeyFunc(key interface{}) api.UnFunc {
	return api.UnFunc(func(ctx context.Context, param0 interface{}) interface{} {
		dataType := reflect.TypeOf(param0)
		dataVal := reflect.ValueOf(param0)

		// validate expected type
		if dataType.Kind() != reflect.Slice && dataType.Kind() != reflect.Array {
			return param0 // ignores the data
		}

		group := make(map[interface{}][]interface{})

		// walk the slice
		for i := 0; i < dataVal.Len(); i++ {
			item := dataVal.Index(i)
			if item.IsValid() {
				switch item.Type().Kind() {
				case reflect.Map:
					key := item.MapIndex(reflect.ValueOf(key))
					if key.IsValid() {
						group[key.Interface()] = append(
							group[key.Interface()], item.Interface(),
						)
					}
				default: //TODO handle type mismatch
				}
			}
		}

		return group
	})
}

// SumInts generates an api.UnFunc that sums up the batched items from  upstream.
// The data is expected to be in format:
// * []integers
// * [][]integers
// The function returs the sum as an int64
func SumInts() api.UnFunc {
	return api.UnFunc(func(ctx context.Context, param0 interface{}) interface{} {
		dataType := reflect.TypeOf(param0)
		dataVal := reflect.ValueOf(param0)

		// validate expected type
		if dataType.Kind() != reflect.Slice && dataType.Kind() != reflect.Array {
			return param0 // ignores the data
		}

		var sum int64

		updateSum := func(op0 *int64, item reflect.Value) {
			if item.IsValid() {
				switch item.Type().Kind() {
				case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
					*op0 += item.Int()
				default: //TODO handle type mismatch
				}

			}
		}

		for i := 0; i < dataVal.Len(); i++ {
			item := dataVal.Index(i)
			if item.IsValid() {
				switch item.Type().Kind() {
				case reflect.Slice, reflect.Array:
					for j := 0; j < item.Len(); j++ {
						updateSum(&sum, item.Index(j))
					}
				default:
					updateSum(&sum, item)
				}
			}
		}

		return sum

	})
}

// SumFloats generates an api.UnFunc that sums up the batched items from  upstream.
// The data is expected to be in format:
// * []floats
// * [][]floats
// The function returs the sum as an float64
func SumFloats() api.UnFunc {
	return api.UnFunc(func(ctx context.Context, param0 interface{}) interface{} {
		dataType := reflect.TypeOf(param0)
		dataVal := reflect.ValueOf(param0)

		// validate expected type
		if dataType.Kind() != reflect.Slice && dataType.Kind() != reflect.Array {
			return param0 // ignores the data
		}

		var sum float64

		updateSum := func(op0 *float64, item reflect.Value) {
			if item.IsValid() {
				switch item.Type().Kind() {
				case reflect.Float32, reflect.Float64:
					*op0 += item.Float()
				default: //TODO handle type mismatch
				}
			}
		}

		for i := 0; i < dataVal.Len(); i++ {
			item := dataVal.Index(i)
			if item.IsValid() {
				switch item.Type().Kind() {
				case reflect.Slice, reflect.Array:
					for j := 0; j < item.Len(); j++ {
						updateSum(&sum, item.Index(j))
					}
				default:
					updateSum(&sum, item)
				}
			}
		}

		return sum

	})
}

func Sort() api.UnFunc {
	return nil
}

func ForAll(f func(ctx context.Context, batch interface{}) map[interface{}][]interface{}) api.UnFunc {
	return api.UnFunc(func(ctx context.Context, param0 interface{}) interface{} {
		return f(ctx, param0)
	})
}
