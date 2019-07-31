package stream

import "github.com/vladimirvivien/automi/operators/binary"

// Reduce accumulates and reduces items from upstream into a
// single value using the initial seed value and the reduction
// binary function.  The provided function must be of type:
//   func(S, T) R
//     where S is the type of the partial result
//     T is the incoming item from the stream
//     R is the type of the result, to be used in the next call
// If reductive operations are called after open-ended emitters
// (i.e. network service), they may never end.
func (s *Stream) Reduce(seed, f interface{}) *Stream {
	operator := binary.New()
	op, err := binary.ReduceFunc(f)
	if err != nil {
		s.drainErr(err)
	}
	operator.SetOperation(op)
	operator.SetInitialState(seed)
	s.ops = append(s.ops, operator)
	return s
}
