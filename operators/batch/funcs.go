package batch

import (
	"context"
	"reflect"
	"sort"
	"strings"

	"github.com/vladimirvivien/automi/api"
	"github.com/vladimirvivien/automi/util"
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

func SumByPosFunc(pos int) api.UnFunc {
	return api.UnFunc(func(ctx context.Context, param0 interface{}) interface{} {
		dataType := reflect.TypeOf(param0)
		dataVal := reflect.ValueOf(param0)

		// validate expected type
		if dataType.Kind() != reflect.Slice && dataType.Kind() != reflect.Array {
			return param0 // ignores the data
		}

		sum := float64(0)

		// walk [][]interface{}
		for i := 0; i < dataVal.Len(); i++ {
			row := dataVal.Index(i)
			switch row.Type().Kind() {
			case reflect.Slice, reflect.Array:
				posVal := row.Index(pos)
				if posVal.IsValid() {
					if util.IsFloatValue(posVal) {
						sum += posVal.Float()
					}
					if util.IsIntValue(posVal) {
						sum += float64(posVal.Int())
					}
					if posVal.Type().Kind() == reflect.Interface {
						if util.IsFloatValue(posVal.Elem()) {
							sum += posVal.Elem().Float()
						}
						if util.IsIntValue(posVal.Elem()) {
							sum += float64(posVal.Elem().Int())
						}
					}
				}
			default: // TODO handle type mismatch
			}
		}
		return sum
	})

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

// SumByNameFunc generates an api.UnFunc that sums incoming batched items
// by field name.  The batched data is expected to be in the following formats:
// * []struct{} - slice of structs
// The operation walks the slice and sums up the struct field F if of types
// * struct{F Numeric}
// * struct{F []Numeric}
// * return float64
func SumByNameFunc(name string) api.UnFunc {
	return api.UnFunc(func(ctx context.Context, param0 interface{}) interface{} {
		dataType := reflect.TypeOf(param0)
		dataVal := reflect.ValueOf(param0)

		// validate expected type
		if dataType.Kind() != reflect.Slice && dataType.Kind() != reflect.Array {
			return param0 // ignores the data
		}
		name = strings.Title(name) // avoid unexported field panic

		sum := float64(0)

		// walk the slice
		for i := 0; i < dataVal.Len(); i++ {
			item := dataVal.Index(i)
			switch item.Type().Kind() {
			case reflect.Struct:
				val := item.FieldByName(name)
				if val.IsValid() {
					switch {
					case util.IsFloatValue(val):
						sum += val.Float()
					case util.IsIntValue(val):
						sum += float64(val.Int())
					default:
						// TODO sum when field F in struct{F []numeric}
					}
				}
			default: //TODO handle type mismatch
			}
		}
		return sum
	})
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

// SumByKeyFunc generates an api.UnFunc that sums incoming batched items
// by key value.  The batched data is expected to be in the following formats:
// * []map[T]T - slice of maps
// The operation walks the slice and sums up each item when
// * map[key]Numeric
// * map[key][]Numeric
// * returns a float64 value
func SumByKeyFunc(key interface{}) api.UnFunc {
	return api.UnFunc(func(ctx context.Context, param0 interface{}) interface{} {
		dataType := reflect.TypeOf(param0)
		dataVal := reflect.ValueOf(param0)

		// validate expected type
		if dataType.Kind() != reflect.Slice && dataType.Kind() != reflect.Array {
			return param0 // ignores the data
		}

		sum := float64(0)

		// walk the slice
		for i := 0; i < dataVal.Len(); i++ {
			item := dataVal.Index(i)
			if item.IsValid() {
				switch item.Type().Kind() {
				case reflect.Map:
					val := item.MapIndex(reflect.ValueOf(key))
					if val.IsValid() {
						switch {
						case util.IsFloatValue(val):
							sum += val.Float()
						case util.IsIntValue(val):
							sum += float64(val.Int())
						default:
							// TODO sum when map[key] returns []Numeric
						}
					}
				default: //TODO handle type mismatch
				}
			}
		}

		return sum
	})
}

// SumInts generates an api.UnFunc that sums up the batched items from  upstream.
// The data is expected to be in format:
// * []integers
// * []floats
// * [][]integers
// * [][]floats
// The function returs the sum as an float64
func Sum() api.UnFunc {
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
				if util.IsFloatValue(item) {
					*op0 += item.Float()
				}
				if util.IsIntValue(item) {
					*op0 += float64(item.Int())
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

// SoftByPosFunc generates a api.UnFunc that sorts batched data from upstream.
// The batched data is expected to be in the following format(s):
// * [][]T
// with each iteration i for batch v:
// - check v[i][pos] to be of type string, integers, float
// - comapre V[i-1][pos] and v[i][pos]
// The compared value at v[i][pos] must be comparable types
// Returs a sorted [][]T
func SortByPosFunc(pos int) api.UnFunc {
	return api.UnFunc(func(ctx context.Context, param0 interface{}) interface{} {
		dataType := reflect.TypeOf(param0)
		dataVal := reflect.ValueOf(param0)

		// validate expected type
		if dataType.Kind() != reflect.Slice && dataType.Kind() != reflect.Array {
			return param0 // ignores the data
		}

		// use sort.Sort() to sepecify a Less function
		sort.Slice(dataVal.Interface(), func(i, j int) bool {
			rowI := dataVal.Index(i)
			rowJ := dataVal.Index(j)
			// can we compare current and previous rows i, j
			typeIOk := rowI.Type().Kind() == reflect.Slice || rowI.Type().Kind() == reflect.Array
			typeJOk := rowJ.Type().Kind() == reflect.Slice || rowI.Type().Kind() == reflect.Array

			// determine type of value at row_i[pos], prev row_j[pos]
			// then compare them.
			if typeIOk && typeJOk {
				itemI := rowI.Index(pos)
				itemJ := rowJ.Index(pos)

				return util.IsLess(itemI, itemJ)
			}
			return false
		})

		return dataVal.Interface()
	})
}

// SortByNameFunc generates a api.UnFunc operation that sorts batched items from upstream
// using the field name of items in the batch.  The batched data is of the form:
// * []struct
// The field identified by name must be of comparable values.
// The function returns a sorted []struct
func SortByNameFunc(name string) api.UnFunc {
	return api.UnFunc(func(ctx context.Context, param0 interface{}) interface{} {
		dataType := reflect.TypeOf(param0)
		dataVal := reflect.ValueOf(param0)

		// validate expected type
		if dataType.Kind() != reflect.Slice && dataType.Kind() != reflect.Array {
			return param0 // ignores the data
		}

		name = strings.Title(name) // cap name to avoid panic
		sort.Slice(dataVal.Interface(), func(i, j int) bool {
			itemI := dataVal.Index(i)
			itemJ := dataVal.Index(j)

			// are items i, j structs
			typeIOk := itemI.Type().Kind() == reflect.Struct
			typeJOk := itemI.Type().Kind() == reflect.Struct

			if typeIOk && typeJOk {
				valI := itemI.FieldByName(name)
				valJ := itemJ.FieldByName(name)
				return util.IsLess(valI, valJ)
			}

			return false
		})

		return dataVal.Interface()
	})
}

// SortByKeyFunc generates a api.UnFunc operation that sorts batched items from upsteram
// using the key value of maps in the batch.  The batched data is of the form:
// * [] map[K]V
// The key specified for sorting must be of comparable types
// The function returns sorted []map[K]V
func SortByKeyFunc(key interface{}) api.UnFunc {
	return api.UnFunc(func(ctx context.Context, param0 interface{}) interface{} {
		dataType := reflect.TypeOf(param0)
		dataVal := reflect.ValueOf(param0)

		// validate expected type
		if dataType.Kind() != reflect.Slice && dataType.Kind() != reflect.Array {
			return param0 // ignores the data
		}

		sort.Slice(dataVal.Interface(), func(i, j int) bool {
			itemI := dataVal.Index(i)
			itemJ := dataVal.Index(j)

			// are items i, j maps
			typeIOk := itemI.Type().Kind() == reflect.Map
			typeJOk := itemI.Type().Kind() == reflect.Map

			if typeIOk && typeJOk {
				valI := itemI.MapIndex(reflect.ValueOf(key))
				valJ := itemJ.MapIndex(reflect.ValueOf(key))
				return util.IsLess(valI, valJ)
			}

			return false
		})

		return dataVal.Interface()
	})
}

func ForAll(f func(ctx context.Context, batch interface{}) map[interface{}][]interface{}) api.UnFunc {
	return api.UnFunc(func(ctx context.Context, param0 interface{}) interface{} {
		return f(ctx, param0)
	})
}
