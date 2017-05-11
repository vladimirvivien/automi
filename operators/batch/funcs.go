package batch

import (
	"context"
	"reflect"
	"sort"
	"strings"

	"github.com/vladimirvivien/automi/api"
	"github.com/vladimirvivien/automi/util"
)

// GroupByPosFunc generates an api.Unary function that groups incoming batched items
// by their position in a slice. The batch is expected to be of type:
//   [][]T
//
// For each item i (in dimension-1):
//  * select value at index [i][pos] as key
//  * group slice [i][] using map[key][]T
//  * return map as result
//
// This method should be called after a Batch operation upstream or it may fail.
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

// SumByPosFunc generates an api.UnFunc that sums incoming values from upstream batched
// items based on their position.  The stream data is expected to be of types:
//   []T - where T is an integer or floating point type
//
// The generated function accumulates the sum from the batch and returns it.
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
// by struct field name.  The batched data is expected to be of type:
//  []T - where T is a struct
//
// The operation walks the slice and for each item t
//   - if item t contains a field with identifier `name`:
//     it puts t into a map of slice of t where t.name is the key
//     map[t.name] = append(slice, t)
// The map is returned  as result.
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
// by sturct field name.  The batched data is expected to be of type:
//   []T - where T is a struct
//
// The operation walks the slice and for each item t
//  - cummulatively sums value t.name
//
// Field t.name value can be of types:
//  - struct{name T} - where T is intergers, floats
//  - struct{name []T} - where T is integers, floats
//
// The generated function returns a value of type float64
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
// by key value.  The batched data is expected to be in the following type:
//  - []map[K]V - slice of map[K]V where K can be a valid map key type
//
// The operation walks the slice and for each item i
//  - it looks for key value that matches the passed key
//  - When a match is found:
//    it stores i in a map of slace []T as map[key] = append(slice, i)
//
// The map is returned as function result.
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
// by key value.  The batched data is expected to be of type:
//   []map[K]V - slice of map[K]V
// Each slice item m can be of type
//   map[K]V - where V can be integers or floats
//   map[K][]V - where []V is a slice of integers or floats
// The function returns a float64 value
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

// Sum generates an api.UnFunc that sums batched items from upstream.
// The data is expected to be of the following types:
//  []integers
//  []floats
//  [][]integers
//  [][]floats
// The function returns the sum as a float64
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

// SortByPosFunc generates a api.UnFunc that sorts batched data from upstream.
// The batched items are expected to be in the following type:
//   [][]T - where T is comparable type
//
// with each iteration i for batch v:
//   - check v[i][pos] to be of type string, integers, float
//   - Use package sort and a Less function to compare v[i][pos] and v[i+1][pos]
// The function returns the sorted slice
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
// using the field name of items in the batch.  The batched data is of type:
//   []T - where T is a struct
// For each struct s, field s.name must be of comparable values.
// The function returns a sorted []T
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
//   []map[K]V - where K is a comparable type
// The function returns sorted []map[K]
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

// SortWithFunc generates an api.UnFunc operation that is intended to sort batched items
// from upstream using the provided Less function to be used with the sort package.
//
// The batched data is expected to be of form:
//  []T - where T is a valid Go type
//
// The specified function should follow the Less function convention of the
// sort package when compairing values from rows i, j.
func SortWithFunc(f func(batch interface{}, i, j int) bool) api.UnFunc {
	return api.UnFunc(func(ctx context.Context, param0 interface{}) interface{} {
		dataType := reflect.TypeOf(param0)
		dataVal := reflect.ValueOf(param0)

		// validate expected type
		if dataType.Kind() != reflect.Slice && dataType.Kind() != reflect.Array {
			return param0 // ignores the data
		}

		sort.Slice(dataVal.Interface(), func(i, j int) bool {
			return f(dataVal.Interface(), i, j)
		})

		return dataVal.Interface()
	})
}

func ForAll(f func(ctx context.Context, batch interface{}) map[interface{}][]interface{}) api.UnFunc {
	return api.UnFunc(func(ctx context.Context, param0 interface{}) interface{} {
		return f(ctx, param0)
	})
}
