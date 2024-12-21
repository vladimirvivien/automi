package funcs

import (
	"cmp"
	"context"
	"reflect"
	"slices"

	"github.com/vladimirvivien/automi/api"
	"github.com/vladimirvivien/automi/reflection"
)

// GroupByIndexFunc generates an api.Unary function that groups incoming batched items
// by their position in a slice. The batch is expected to be of type:
//
//	[][]T - where []T is a slice or array of data items
//
// The function returns type
//
//	[]map[interface{}][]interface{}
func GroupByIndexFunc[IN ~[][]ITEM, OUT map[ITEM][][]ITEM, ITEM comparable](pos int) ExecFunction[IN, OUT] {
	return func(ctx context.Context, param0 IN) OUT {

		group := make(OUT)

		for _, row := range param0 {
			if pos < 0 || pos >= len(row) {
				continue // Skip rows where pos is out of bounds
			}
			key := row[pos]
			group[key] = append(group[key], slices.Delete(row, pos, pos+1)) // Exclude key at pos
		}

		return group
	}
}

// SumByIndexFunc generates an api.UnFunc that sums incoming values from upstream batched
// items based on their position.  The stream data is expected to be of types:
//
//	[][]T - where []T is a slice of integers or floating points
//
// The function returns a value of type
//
//	[]map[int]float64
//
// More specifically a value:
//
//	[]map[int]float64{{pos: sum}} where sum is the calculated sum.
func SumByIndexFunc[IN ~[][]ITEM, ITEM any](index int) ExecFunction[IN, float64] {
	return func(ctx context.Context, param0 IN) float64 {
		var result float64
		for _, row := range param0 {
			if index >= len(row) {
				continue // Skip if pos is out of bounds
			}

			// Check the type of the element at the specified index
			val := reflect.ValueOf(row[index])
			if !val.IsValid() {
				continue
			}
			result += reflection.ValueAsFloat(val)
		}
		return result
	}
}

// GroupByStructFieldFunc generates an api.UnFunc that groups incoming batched items
// by struct field name.  The batched data is expected to be of type:
//
//	[]struct{T} - where T is the type of a struct fields identified by name
//
// The function returns a type
//
//	[]map[interface{}][]interface{}
//
// Where the map that uses the field values as key to group the items.

func GroupByStructFieldFunc[IN ~[]STRUCT, STRUCT any](name string) ExecFunction[IN, map[any]IN] {
	return func(ctx context.Context, param0 IN) map[any]IN {
		group := make(map[any]IN)

		for _, structItem := range param0 {
			item := reflect.ValueOf(structItem)

			if item.Kind() != reflect.Struct {
				// Skip if it's not a struct
				continue
			}

			field := item.FieldByName(name)
			if !field.IsValid() || !field.CanInterface() {
				// Skip if the field doesn't exist or is unexported
				continue
			}

			key := field.Interface()
			group[key] = append(group[key], structItem)
		}

		return group
	}
}

// SumByStructFieldFunc generates an api.UnFunc that sums incoming batched items
// by sturct field name.  The batched data is expected to be of type:
//   - []struct{F} - where field F is either an integer or floating point
//   - []struct{V} - where field V is a slice of integers or floating points
//
// The function returns value of type
//
//	map[string]float64
//
// For instance
//
//	[]map[string]float64{{name:sum}}
//
// Where sum is the total calculated sum for fields name.
func SumByStructFieldFunc[IN ~[]STRUCT, STRUCT any](name string) ExecFunction[IN, float64] {

	return func(ctx context.Context, param0 IN) float64 {
		var result float64

		// Walk the slice
		for _, structItem := range param0 {
			item := reflect.ValueOf(structItem)
			field := item.FieldByName(name)

			if !field.IsValid() || !field.CanInterface() {
				// Skip if the field doesn't exist or is unexported
				continue
			}

			result += reflection.ValueAsFloat(field)
		}

		return result
	}
}

// GroupByMapKeyFunc generates an api.UnFunc that groups incoming batched items
// by key value.  The batched data is expected to be in the following type:
//
//	[]map[K]V - slice of map[K]V
//
// The batched data is grouped in a slice of map of type
//
//	[]map[interface{}][]interface{}
//
// Where items with simlar K values are assigned the same key in the result map.
func GroupByMapKeyFunc[IN ~[]map[K]V, K, V comparable](key K) ExecFunction[IN, map[V]IN] {
	return func(ctx context.Context, param0 IN) map[V]IN {
		group := make(map[V]IN)

		for _, mapItem := range param0 {
			if value, ok := mapItem[key]; ok {
				group[value] = append(group[value], mapItem)
			}
		}

		return group
	}
}

// SumByMapKeyFunc generates an api.UnFunc that sums incoming batched items
// by key value.  The batched data can be of the following types:
//
//	[]map[K]V - where V is either an integer or a floating point
//	[]map[K][]V - where []V is a slice of integers or floating points
//
// The function returns type
//
//	[]map[interface{}]
//
// For instance:
//
//	[]map[interface{}]float64{key: sum}
//
// Where sum is the total calculated sum for a given key.
// If key == nil, it returns sums for all keys.
func SumByMapKeyFunc[IN ~[]map[K]V, K comparable, V any](key K) ExecFunction[IN, float64] {
	return func(ctx context.Context, param0 IN) float64 {
		var result float64

		// walk the slice
		for _, mapItem := range param0 {
			mapVal, ok := mapItem[key]
			if !ok {
				continue
			}
			result += reflection.ValueAsFloat(reflect.ValueOf(mapVal))
		}

		return result
	}
}

// SumFunc generates an api.UnFunc that sums batched items from upstream.
// The data is expected to be of the following types:
//
//	[]integers
//	[]floats
//	[][]integers
//	[][]floats
//
// The function returns the sum as a float64
func SumFunc[IN ~[]ITEM | ~[][]ITEM, ITEM api.NumericConstraint]() ExecFunction[IN, float64] {
	return func(ctx context.Context, param0 IN) float64 {
		var sum float64

		switch v := any(param0).(type) {
		case []ITEM:
			for _, item := range v {
				sum += float64(item)
			}
		case [][]ITEM:
			for _, subSlice := range v {
				for _, item := range subSlice {
					sum += float64(item)
				}
			}
		}

		return sum

	}
}

// SortSliceFunc generates an api.UnFunc that sorts batched data from upstream.
// The batched items are expected to be in the following type:
//
//	[]T - where T is comparable type (string, numeric, etc)
//
// with each iteration i for batch v:
//   - check v[i] to be of type string, integers, float
//   - Use package sort and a Less function to compare v[i] and v[i+1]
//
// The function returns the sorted slice
func SortSliceFunc[SLICE ~[]ITEM, ITEM cmp.Ordered]() ExecFunction[SLICE, SLICE] {
	return func(ctx context.Context, param0 SLICE) SLICE {
		slices.Sort(param0)
		return param0
	}
}

// SortSliceByIndexFunc generates a api.UnFunc that sorts batched data from upstream.
// The batched items are expected to be in the following type:
//
//	[][]T - where T is comparable type
//
// with each iteration i for batch v:
//   - check v[i][pos] to be of type string, integers, float
//   - Use package sort and a Less function to compare v[i][pos] and v[i+1][pos]
//
// The function returns the sorted slice
func SortSliceByIndexFunc[SLICE ~[][]ITEM, ITEM cmp.Ordered](index int) ExecFunction[SLICE, SLICE] {
	return func(ctx context.Context, param0 SLICE) SLICE {
		slices.SortFunc(param0, func(a, b []ITEM) int {
			if index >= len(a) || index >= len(b) {
				return 0 // Consider equal if out-of-bounds
			}
			return cmp.Compare(a[index], b[index])
		})
		return param0
	}
}

// SortByStructFieldFunc generates a api.UnFunc operation that sorts batched items from upstream
// using the field name of items in the batch.  The batched data is of type:
//
//	[]T - where T is a struct
//
// For each struct s, field s.name must be of comparable values.
// The function returns a sorted []T
func SortByStructFieldFunc[SLICE ~[]ITEM, ITEM any](name string) ExecFunction[SLICE, SLICE] {
	return func(ctx context.Context, param0 SLICE) SLICE {
		if reflect.TypeOf(param0).Elem().Kind() != reflect.Struct {
			panic("SortByStructField requires struct")
		}

		slices.SortFunc(param0, func(i, j ITEM) int {
			fieldI := reflect.ValueOf(i).FieldByName(name)
			fieldJ := reflect.ValueOf(j).FieldByName(name)

			if !fieldI.IsValid() || !fieldJ.IsValid() {
				return 0 // equal
			}

			switch {
			case reflection.IsLess(fieldI, fieldJ):
				return -1
			case reflection.IsMore(fieldI, fieldJ):
				return 1
			case reflection.IsEqual(fieldI, fieldJ):
				return 0
			default:
				return 0
			}
		})

		return param0
	}
}

// SortByMapKeyFunc generates a api.UnFunc operation that sorts batched items from upsteram
// using the key value of maps in the batch.  The batched data is of the form:
//
//	[]map[K]V - where K is a comparable type
//
// The function returns sorted []map[K]
func SortByMapKeyFunc[MAPS ~[]map[K]V, K comparable, V cmp.Ordered](key K) ExecFunction[MAPS, MAPS] {
	return func(ctx context.Context, param0 MAPS) MAPS {
		slices.SortFunc(param0, func(i, j map[K]V) int {
			return cmp.Compare(i[key], j[key])

			// // Handle cases where the key might not exist in one or both maps
			// if !okI && !okJ {
			// 	return 0 // Both missing, consider equal
			// }
			// if !okI && okJ {
			// 	return 1 // Only j has the key, j is "greater"
			// }
			// if okI && !okJ {
			// 	return -1 // Only i has the key, i is "less"
			// }

			// if okI && okJ {
			// 	valI := reflect.ValueOf(itemI)
			// 	valJ := reflect.ValueOf(itemJ)
			// 	switch {
			// 	case util.IsEqual(valI, valJ):
			// 		return 0
			// 	case util.IsLess(valI, valJ):
			// 		return -1
			// 	case util.IsMore(valI, valJ):
			// 		return 1
			// 	default:
			// 		return 0
			// 	}
			// }

			// return 0
		})
		return param0
	}
}

// SortWithFuncFunc generates an api.UnFunc operation that is intended to sort batched items
// from upstream using the provided Less function to be used with the sort package.
//
// The batched data is expected to be of form:
//
//	[]T - where T is a valid Go type
//
// The specified function should follow the Less function convention of the
// sort package when compairing values from rows i, j.
func SortWithFuncFunc[SLICE ~[]ITEM, ITEM any](f func(i, j ITEM) int) ExecFunction[SLICE, SLICE] {
	return func(ctx context.Context, param0 SLICE) SLICE {
		slices.SortFunc(param0, func(i, j ITEM) int {
			return f(i, j)
		})

		return param0
	}
}
