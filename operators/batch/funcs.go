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
//   [][]T - where []T is a slice or array of data items
// The function returns type
//   []map[interface{}][]interface{}
func GroupByPosFunc(pos int) api.UnFunc {
	return api.UnFunc(func(ctx context.Context, param0 interface{}) interface{} {
		dataType := reflect.TypeOf(param0)
		dataVal := reflect.ValueOf(param0)

		// validate expected type
		if dataType.Kind() != reflect.Slice && dataType.Kind() != reflect.Array {
			return param0 // ignores the data
		}

		group := make(map[interface{}][]interface{})
		groupItems := func(key reflect.Value, row reflect.Value, grp map[interface{}][]interface{}) {
			for j := 0; j < row.Len(); j++ {
				if j != pos {
					grp[key.Interface()] = append(
						grp[key.Interface()],
						row.Index(j).Interface(),
					)
				}
			}
		}

		// walk dimension 1 of [][]interface{}
		for i := 0; i < dataVal.Len(); i++ {
			row := dataVal.Index(i)
			switch row.Type().Kind() {
			case reflect.Slice, reflect.Array:
				key := row.Index(pos)
				if key.IsValid() {
					groupItems(key, row, group)
				}
			case reflect.Interface:
				elem := row.Elem()
				switch elem.Type().Kind() {
				case reflect.Slice, reflect.Array:
					key := elem.Index(pos)
					groupItems(key, elem, group)
				}

			default: // TODO handle type mismatch
			}
		}
		return []map[interface{}][]interface{}{group}
	})
}

// SumByPosFunc generates an api.UnFunc that sums incoming values from upstream batched
// items based on their position.  The stream data is expected to be of types:
//   [][]T - where []T is a slice of integers or floating points
// The function returns a value of type
//   []map[int]float64
// More specifically a value:
//   []map[int]float64{{pos: sum}} where sum is the calculated sum.
func SumByPosFunc(pos int) api.UnFunc {
	return api.UnFunc(func(ctx context.Context, param0 interface{}) interface{} {
		dataType := reflect.TypeOf(param0)
		dataVal := reflect.ValueOf(param0)

		// validate expected type
		if dataType.Kind() != reflect.Slice && dataType.Kind() != reflect.Array {
			return param0 // ignores the data
		}

		result := make(map[int]float64)

		for i := 0; i < dataVal.Len(); i++ {
			row := dataVal.Index(i)
			switch row.Type().Kind() {
			case reflect.Slice, reflect.Array:
				val := row.Index(pos)
				result[pos] += sumAll(val)
			case reflect.Interface:
				elem := row.Elem()
				switch elem.Type().Kind() {
				case reflect.Slice, reflect.Array:
					val := elem.Index(pos)
					result[pos] += sumAll(val)
				}
			default: // TODO handle type mismatch
			}
		}
		return []map[int]float64{result}
	})

}

// GroupByNameFunc generates an api.UnFunc that groups incoming batched items
// by struct field name.  The batched data is expected to be of type:
//   []struct{T} - where T is the type of a struct fields identified by name
// The function returns a type
//   []map[interface{}][]interface{}
// Where the map that uses the field values as key to group the items.
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

		groupItems := func(key, value reflect.Value, grp map[interface{}][]interface{}) {
			if key.IsValid() {
				grp[key.Interface()] = append(
					grp[key.Interface()], value.Interface(),
				)
			}
		}

		// walk the slice
		for i := 0; i < dataVal.Len(); i++ {
			item := dataVal.Index(i)
			switch item.Type().Kind() {
			case reflect.Struct:
				key := item.FieldByName(name)
				if key.IsValid() {
					groupItems(key, item, group)
				}
			case reflect.Interface:
				mapItem := item.Elem()
				if mapItem.Type().Kind() == reflect.Struct {
					itemKey := mapItem.FieldByName(name)
					groupItems(itemKey, mapItem, group)
				}

			default: //TODO handle type mismatch
			}
		}
		return []map[interface{}][]interface{}{group}
	})
}

// SumByNameFunc generates an api.UnFunc that sums incoming batched items
// by sturct field name.  The batched data is expected to be of type:
//   - []struct{F} - where field F is either an integer or floating point
//   - []struct{V} - where field V is a slice of integers or floating points
// The function returns value of type
//   map[string]float64
// For instance
//   []map[string]float64{{name:sum}}
// Where sum is the total calculated sum for fields name.
func SumByNameFunc(name string) api.UnFunc {
	return api.UnFunc(func(ctx context.Context, param0 interface{}) interface{} {
		dataType := reflect.TypeOf(param0)
		dataVal := reflect.ValueOf(param0)

		// validate expected type
		if dataType.Kind() != reflect.Slice && dataType.Kind() != reflect.Array {
			return param0 // ignores the data
		}

		name = strings.Title(name) // avoid unexported field panic
		result := make(map[string]float64)

		// walk the slice
		for i := 0; i < dataVal.Len(); i++ {
			item := dataVal.Index(i)
			switch item.Type().Kind() {
			case reflect.Struct:
				if name != "" {
					val := item.FieldByName(name)
					result[name] += sumAll(val)
				} else {
					// if no field provide, sum all fields
					for i := 0; i < item.Type().NumField(); i++ {
						itemField := item.Type().Field(i)
						itemVal := item.Field(i)
						result[itemField.Name] += sumAll(itemVal)
					}
				}
			case reflect.Interface:
				elem := item.Elem()
				if elem.Type().Kind() == reflect.Struct {
					if name != "" {
						val := elem.FieldByName(name)
						result[name] += sumAll(val)
					} else {
						for i := 0; i < elem.Type().NumField(); i++ {
							itemField := elem.Type().Field(i)
							itemVal := elem.Field(i)
							result[itemField.Name] += sumAll(itemVal)
						}
					}
				}
			default: //TODO handle type mismatch

			}

		}
		return []map[string]float64{result}
	})
}

// GroupByKeyFunc generates an api.UnFunc that groups incoming batched items
// by key value.  The batched data is expected to be in the following type:
//   []map[K]V - slice of map[K]V
// The batched data is grouped in a slice of map of type
//   []map[interface{}][]interface{}
// Where items with simlar K values are assigned the same key in the result map.
func GroupByKeyFunc(key interface{}) api.UnFunc {
	return api.UnFunc(func(ctx context.Context, param0 interface{}) interface{} {
		dataType := reflect.TypeOf(param0)
		dataVal := reflect.ValueOf(param0)

		// validate expected type
		if dataType.Kind() != reflect.Slice && dataType.Kind() != reflect.Array {
			return param0 // ignores the data
		}

		group := make(map[interface{}][]interface{})
		groupItems := func(key, value reflect.Value, grp map[interface{}][]interface{}) {
			if key.IsValid() {
				grp[key.Interface()] = append(
					group[key.Interface()], value.Interface(),
				)
			}
		}

		// walk the slice
		for i := 0; i < dataVal.Len(); i++ {
			item := dataVal.Index(i)
			if item.IsValid() {
				switch item.Type().Kind() {
				case reflect.Map:
					itemKey := item.MapIndex(reflect.ValueOf(key))
					groupItems(itemKey, item, group)
				case reflect.Interface:
					mapItem := item.Elem()
					if mapItem.Type().Kind() == reflect.Map {
						itemKey := mapItem.MapIndex(reflect.ValueOf(key))
						groupItems(itemKey, mapItem, group)
					}

				default: //TODO handle type mismatch
				}
			}
		}

		return []map[interface{}][]interface{}{group}
	})
}

// SumByKeyFunc generates an api.UnFunc that sums incoming batched items
// by key value.  The batched data can be of the following types:
//   []map[K]V - where V is either an integer or a floating point
//   []map[K][]V - where []V is a slice of integers or floating points
// The function returns type
//   []map[interface{}]
// For instance:
//   []map[interface{}]float64{key: sum}
// Where sum is the total calculated sum for a given key.
// If key == nil, it returns sums for all keys.
func SumByKeyFunc(key interface{}) api.UnFunc {
	return api.UnFunc(func(ctx context.Context, param0 interface{}) interface{} {
		dataType := reflect.TypeOf(param0)
		dataVal := reflect.ValueOf(param0)

		// validate expected type
		if dataType.Kind() != reflect.Slice && dataType.Kind() != reflect.Array {
			return param0 // ignores the data
		}

		result := make(map[interface{}]float64)

		// walk the slice
		for i := 0; i < dataVal.Len(); i++ {
			item := dataVal.Index(i)
			if item.IsValid() {
				switch item.Type().Kind() {
				case reflect.Map:
					if key != nil {
						val := item.MapIndex(reflect.ValueOf(key))
						result[key] += sumAll(val)
					} else {
						// if not key provided, sum up all map entries
						for _, k := range item.MapKeys() {
							val := item.MapIndex(k)
							result[k.Interface()] += sumAll(val)
						}
					}
				case reflect.Interface:
					elem := item.Elem()
					if elem.Type().Kind() == reflect.Map {
						if key != nil {
							val := elem.MapIndex(reflect.ValueOf(key))
							result[key] += sumAll(val)
						} else {
							// if key is nil, sum up all map entries
							for _, k := range elem.MapKeys() {
								val := elem.MapIndex(k)
								result[k.Interface()] += sumAll(val)
							}
						}
					}
				default: //TODO handle type mismatch
				}
			}
		}

		return []map[interface{}]float64{result}
	})
}

// SumFunc generates an api.UnFunc that sums batched items from upstream.
// The data is expected to be of the following types:
//  []integers
//  []floats
//  [][]integers
//  [][]floats
// The function returns the sum as a float64
func SumFunc() api.UnFunc {
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
				case reflect.Interface:
					elem := item.Elem()
					switch elem.Type().Kind() {
					case reflect.Slice, reflect.Array:
						for j := 0; j < elem.Len(); j++ {
							updateSum(&sum, elem.Index(j))
						}
					}
				default:
					updateSum(&sum, item)
				}
			}
		}

		return sum

	})
}

// SortFunc generates an api.UnFunc that sorts batched data from upstream.
// The batched items are expected to be in the following type:
//   []T - where T is comparable type (string, numeric, etc)
//
// with each iteration i for batch v:
//   - check v[i] to be of type string, integers, float
//   - Use package sort and a Less function to compare v[i] and v[i+1]
// The function returns the sorted slice
func SortFunc() api.UnFunc {
	return api.UnFunc(func(ctx context.Context, param0 interface{}) interface{} {
		dataType := reflect.TypeOf(param0)
		dataVal := reflect.ValueOf(param0)

		// validate expected type
		if dataType.Kind() != reflect.Slice && dataType.Kind() != reflect.Array {
			return param0 // ignores the data
		}

		// use sort.Sort() to sepecify a Less function
		sort.Slice(dataVal.Interface(), func(i, j int) bool {
			itemI := dataVal.Index(i)
			itemJ := dataVal.Index(j)
			return util.IsLess(itemI, itemJ)
		})

		return dataVal.Interface()
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

// ForAll returns an api.UnFunc that is applied on all items in the batch
func ForAll(f func(ctx context.Context, batch interface{}) map[interface{}][]interface{}) api.UnFunc {
	return api.UnFunc(func(ctx context.Context, param0 interface{}) interface{} {
		return f(ctx, param0)
	})
}

func sumAll(item reflect.Value) float64 {
	if !item.IsValid() {
		return 0.0
	}

	itemVal := item
	if item.Type().Kind() == reflect.Interface {
		itemVal = item.Elem()
	}

	sum := 0.0
	switch itemVal.Type().Kind() {
	case reflect.Array, reflect.Slice:
		for i := 0; i < itemVal.Len(); i++ {
			sum += util.ValueAsFloat(itemVal.Index(i))
		}
	default:
		sum = util.ValueAsFloat(itemVal)
	}
	return sum
}
