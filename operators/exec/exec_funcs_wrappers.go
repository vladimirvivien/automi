package exec

import (
	"cmp"

	"github.com/vladimirvivien/automi/api"
	"github.com/vladimirvivien/automi/funcs"
)

func GroupByIndex[IN ~[][]ITEM, ITEM comparable](pos int) *ExecOperator[IN, map[ITEM][][]ITEM] {
	return Execute(funcs.GroupByIndexFunc[IN](pos))
}

func GroupByStructField[IN ~[]STRUCT, STRUCT any](name string) *ExecOperator[IN, map[any]IN] {
	return Execute(funcs.GroupByStructFieldFunc[IN](name))
}

func GroupByMapKey[IN ~[]map[K]V, K, V comparable](key K) *ExecOperator[IN, map[V]IN] {
	return Execute(funcs.GroupByMapKeyFunc[IN](key))
}

func SumByIndex[IN ~[][]ITEM, ITEM any](pos int) *ExecOperator[IN, float64] {
	return Execute(funcs.SumByIndexFunc[IN](pos))
}

func SumByStructField[IN ~[]STRUCT, STRUCT any](name string) *ExecOperator[IN, float64] {
	return Execute(funcs.SumByStructFieldFunc[IN](name))
}

func SumByMapKey[IN ~[]map[K]V, K comparable, V any](key K) *ExecOperator[IN, float64] {
	return Execute(funcs.SumByMapKeyFunc[IN](key))
}

func Sum[IN ~[]ITEM | ~[][]ITEM, ITEM api.NumericConstraint]() *ExecOperator[IN, float64] {
	return Execute(funcs.SumFunc[IN, ITEM]())
}

func SumAll1D[IN ~[]ITEM, ITEM api.NumericConstraint]() *ExecOperator[IN, float64] {
	return Execute(funcs.SumFunc[IN, ITEM]())
}

func SumAll2D[IN ~[][]ITEM, ITEM api.NumericConstraint]() *ExecOperator[IN, float64] {
	return Execute(funcs.SumFunc[IN, ITEM]())
}

func SortSlice[SLICE ~[]ITEM, ITEM cmp.Ordered]() *ExecOperator[SLICE, SLICE] {
	return Execute(funcs.SortSliceFunc[SLICE]())
}

func SortSliceByIndex[SLICE ~[][]ITEM, ITEM cmp.Ordered](index int) *ExecOperator[SLICE, SLICE] {
	return Execute(funcs.SortSliceByIndexFunc[SLICE](index))
}

func SortByStructField[SLICE ~[]ITEM, ITEM any](name string) *ExecOperator[SLICE, SLICE] {
	return Execute(funcs.SortByStructFieldFunc[SLICE](name))
}

func SortByMapKey[MAPS ~[]map[K]V, K comparable, V cmp.Ordered](key K) *ExecOperator[MAPS, MAPS] {
	return Execute(funcs.SortByMapKeyFunc[MAPS](key))
}

func SortWithFunc[SLICE ~[]ITEM, ITEM any](f func(i, j ITEM) int) *ExecOperator[SLICE, SLICE] {
	return Execute(funcs.SortWithFuncFunc[SLICE](f))
}
