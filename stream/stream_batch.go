package stream

import (
	"github.com/vladimirvivien/automi/api"
	"github.com/vladimirvivien/automi/operators/batch"
	"github.com/vladimirvivien/automi/operators/unary"
)

// Batch returns a *batch.Operator with batch.TriggerAll set
func (s *Stream) Batch() *Stream {
	operator := batch.New()
	operator.SetTrigger(batch.TriggerAll())
	return s.appendOp(operator)
}

// BatchBySize returns a *batch.Operator with batch.TriggerBySize set
func (s *Stream) BatchBySize(size int64) *Stream {
	operator := batch.New()
	operator.SetTrigger(batch.TriggerBySize(size))
	return s.appendOp(operator)
}

// GroupByKey groups incoming items that are batched as
// type []map[K]V where parameter key is used to group
// the items when K=key.  Items with same key values are
// grouped in a new map and returned as []map[G]V.
//
// See Also
//
// See batch operator function GroupByKey in
//   "github.com/vladimirvivien/automi/operators/batch/"#GroupByKeyFunc
func (s *Stream) GroupByKey(key interface{}) *Stream {
	operator := unary.New()
	operator.SetOperation(batch.GroupByKeyFunc(key))
	return s.appendOp(operator)
}

// GroupByName groups incoming items that are batched as
// type []T where T is a struct. Parameter name is used to select
// T.name as key to group items with the same value into a map map[key][]T
// that is sent downstream.
//
// See Also
//
// See batch operator function GroupByName in
//    "github.com/vladimirvivien/automi/operators/batch"
func (s *Stream) GroupByName(name string) *Stream {
	operator := unary.New()
	operator.SetOperation(batch.GroupByNameFunc(name))
	return s.appendOp(operator)
}

// GroupByPos groups incoming items that are batched as
// [][]T. For each i in dimension 1, [i][pos] is selected as key
// and grouped in a map, map[key][]T, that is returned downstream.
//
// See Also
//
// See the batch operator function GroupByPosFunc in
//   "github.com/vladimirvivien/automi/operators/batch"
func (s *Stream) GroupByPos(pos int) *Stream {
	operator := unary.New()
	operator.SetOperation(batch.GroupByPosFunc(pos))
	return s.appendOp(operator)
}

// Sort sorts incoming items that are batched as []T where
// value T is comparable.  The operator returns sorted slice []T.
// See also the operator function SortFunc in
//   "github.com/vladimirvivien/automi/operators/batch"
func (s *Stream) Sort() *Stream {
	operator := unary.New()
	operator.SetOperation(batch.SortFunc())
	return s.appendOp(operator)
}

// SortByKey sorts incoming items that are batched as type []map[K]V
// where K is a comparable type specified by param key and used to
// sort the slice.  The opertor returns a sorted []map[K]V.
//
// See also the operator function SortByKeyFunc in
//   "github.com/vladimirvivien/automi/operators/batch"
func (s *Stream) SortByKey(key interface{}) *Stream {
	operator := unary.New()
	operator.SetOperation(batch.SortByKeyFunc(key))
	return s.appendOp(operator)
}

// SortByName sorts incoming items that are batched as []T where
// T struct with fields identified by param name.  Value struct.<name>
// is used to sort the slice.  The operator returns stored slice []T.
//
// See also the operator function SortByNameFunc in
//   "github.com/vladimirvivien/automi/operators/batch"
func (s *Stream) SortByName(name string) *Stream {
	operator := unary.New()
	operator.SetOperation(batch.SortByNameFunc(name))
	return s.appendOp(operator)
}

// SortByPos sorts incoming items that are batched as [][]T where
// value at [][[pos]T is used to sort the slice.  The operator
// returns sorted slice [][]T.
//
// See also the operator function SortByPosFunc in
//   "github.com/vladimirvivien/automi/operators/batch"
func (s *Stream) SortByPos(pos int) *Stream {
	operator := unary.New()
	operator.SetOperation(batch.SortByPosFunc(pos))
	return s.appendOp(operator)
}

// SortWith sorts incoming items that are batched as []T using the
// provided Less function for applicaiton with the sort package.
//
// See also the operator function SortWithFunc in
//   "github.com/vladimirvivien/automi/operators/batch"
func (s *Stream) SortWith(f func(batch interface{}, i, j int) bool) *Stream {
	operator := unary.New()
	operator.SetOperation(batch.SortWithFunc(f))
	return s.appendOp(operator)
}

// Sum sums up numeric items that are batched as []T or [][]T where
// T is an integer or a floating point value. The operator returns a
// single value of type float64.
//
// See also the operator function SumFunc in
//   "github.com/vladimirvivien/automi/operators/batch"
func (s *Stream) Sum() *Stream {
	operator := unary.New()
	operator.SetOperation(batch.SumFunc())
	return s.appendOp(operator)
}

// SumByKey sums up numeric items that are batched as []map[K]V or
// []map[K][]V where key specifies a K value that returns a V or a []V that
// is a numeric (or a slice of) value of type integer or floating
// point. If key == nil, the grand total sum of all values for all keys
// will be calculated.
//
// This operator returns map[interface{}]float64{key:sum} where
// sum is the calculated sum.
//
// See also the operator function SumByKeyFunc in
//   "github.com/vladimirvivien/automi/operators/batch"
func (s *Stream) SumByKey(key interface{}) *Stream {
	operator := unary.New()
	operator.SetOperation(batch.SumByKeyFunc(key))
	return s.appendOp(operator)
}

// SumAllKeys returns a grand total of all keys by calling
//  SumByKey(nil)
func (s *Stream) SumAllKeys() *Stream {
	return s.SumByKey(nil)
}

// SumByName sums up items that are batched as []T where
// T is a struct.  The name parameter sums up fields with
// name identifier and are of integer of floating point
// types.  The operator returns a float64 value.
//
// See also the operator function SumByNameFunc in
//   "github.com/vladimirvivien/automi/operator/batch"
func (s *Stream) SumByName(name string) *Stream {
	operator := unary.New()
	operator.SetOperation(batch.SumByNameFunc(name))
	return s.appendOp(operator)
}

// SumByPos sums up items that are batched as []T or
// [][]T where T is an integer or floating point. Values
// [pos]T or [][pos]T are added and returned as a float64
// value.
//
// See also the operator function SumByPosFunc in
//   "github.com/vladimirvivien/automi/operator/batch"
func (s *Stream) SumByPos(pos int) *Stream {
	operator := unary.New()
	operator.SetOperation(batch.SumByPosFunc(pos))
	return s.appendOp(operator)
}

func (s *Stream) appendOp(operator api.Operator) *Stream {
	s.ops = append(s.ops, operator)
	return s
}
