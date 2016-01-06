package elem

import "fmt"

// Elem represents a stream data element
type Elem struct {
	Val interface{} // the actual value, use for direct type assertion.
}

// New creates a new data element
func New(val interface{}) *Elem {
	return &Elem{val}
}

// Slice contruct a slice of elem type
func Slice(vals ...interface{}) []*Elem {
	result := make([]*Elem, len(vals))
	for i, val := range vals {
		result[i] = &Elem{val}
	}
	return result
}

// String returns element as a string value
// If the type implements fmt.Stringer, it returns val.String()
// Otherwise it attempts an assertion to that type.
// If assertion fails, it returns the zero value for the type.
func (e *Elem) String() string {
	if val, ok := e.Val.(fmt.Stringer); ok {
		return val.String()
	}
	if val, ok := e.Val.(string); ok {
		return val
	}
	return ""
}

// TODO - Add all Numeric types

// Int returns the element as an int value.
// If assertion fails, it returns the zero value for type.
func (e *Elem) Int() int {
	if val, ok := e.Val.(int); ok {
		return val
	}
	return 0
}

// Int32 returns the element as an int32.
// If assertions fails, it returns the zero value for type.
func (e *Elem) Int32() int32 {
	if val, ok := e.Val.(int32); ok {
		return val
	}
	return 0
}

// Int64 returns the element as an int64.
// If assertion fails, it returns the zero value for type.
func (e *Elem) Int64() int64 {
	if val, ok := e.Val.(int64); ok {
		return val
	}
	return 0
}

// Float32 returns the element as a float32 value.
// If assertion fails, it returns the zero value for type.
func (e *Elem) Float32() float32 {
	if val, ok := e.Val.(float32); ok {
		return val
	}
	return 0
}

// Flat64 returns the element as a float64.
// If assertion fails, it returns the zero value for type.
func (e *Elem) Float64() float64 {
	if val, ok := e.Val.(float64); ok {
		return val
	}
	return 0
}
