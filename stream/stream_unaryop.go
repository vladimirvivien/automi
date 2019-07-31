package stream

import (
	"github.com/vladimirvivien/automi/api"
	"github.com/vladimirvivien/automi/operators/unary"
)

// Transform is the base method used to apply transfomrmative
// unary operations to streamed elements (i.e. filter, map, etc)
// It is exposed here for completeness, use the other more specific methods.
func (s *Stream) Transform(op api.UnOperation) *Stream {
	operator := unary.New()
	operator.SetOperation(op)
	s.ops = append(s.ops, operator)
	return s
}

// Process applies the user-defined function for general processing of incoming
// streamed elements.  The user-defined function must be of type:
//   func(T) R - where T is the incoming item from upstream,
//               R is the type of the processed value
//
// See Also
//
//   "github.com/vladimirvivien/automi/operators/unary"#ProcessFunc
func (s *Stream) Process(f interface{}) *Stream {
	op, err := unary.ProcessFunc(f)
	if err != nil {
		s.drainErr(err)
	}
	return s.Transform(op)
}

// Filter takes a predicate user-defined func that filters the stream.
// The specified function must be of type:
//   func (T) bool
// If the func returns true, current item continues downstream.
func (s *Stream) Filter(f interface{}) *Stream {
	op, err := unary.FilterFunc(f)
	if err != nil {
		s.drainErr(err)
	}
	return s.Transform(op)
}

// Map uses the user-defined function to take the value of an incoming item and
// returns a new value that is said to be mapped to the intial item.  The user-defined
// function must be of type:
//   func(T) R - where T is the type of the incoming item and R the type of the returned item.
func (s *Stream) Map(f interface{}) *Stream {
	op, err := unary.MapFunc(f)
	if err != nil {
		s.drainErr(err)
	}
	return s.Transform(op)
}

// FlatMap similar to Map, however, the user-defined function is expected to return
// a slice of values (instead of just one mapped value) for downstream operators.
// The FlatMap function flatten the slice, returned by the user-defined function,
// into items that are individually streamed. The user-defined function must have
// the the following type:
//   func(T) []R - where T is the incoming item and []R is a slice to be flattened
func (s *Stream) FlatMap(f interface{}) *Stream {
	op, err := unary.FlatMapFunc(f)
	if err != nil {
		s.drainErr(err)
	}
	s.Transform(op) // add flatmap as unary op
	s.ReStream()    // add streamop to unpack flatmap result
	return s
}
