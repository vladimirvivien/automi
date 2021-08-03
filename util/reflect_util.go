package util

import "reflect"

// IsNumericValue returns true if val is a number type
func IsNumericValue(val reflect.Value) bool {
	return IsIntValue(val) || IsFloatValue(val)
}

// IsIntValue returns true if val is an integer value
func IsIntValue(val reflect.Value) bool {
	switch val.Type().Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return true
	case reflect.Interface:
		return IsIntValue(val.Elem())
	}
	return false
}

// IsFloatValue returns true if val is a float valuew
func IsFloatValue(val reflect.Value) bool {
	switch val.Type().Kind() {
	case reflect.Float32, reflect.Float64:
		return true
	case reflect.Interface:
		return IsFloatValue(val.Elem())
	}
	return false
}

// ValueAsFloat returns item as a float64
func ValueAsFloat(item reflect.Value) float64 {
	itemVal := item
	if item.Type().Kind() == reflect.Interface {
		itemVal = item.Elem()
	}

	if IsFloatValue(itemVal) {
		return itemVal.Float()
	}
	if IsIntValue(itemVal) {
		return float64(itemVal.Int())
	}
	return 0.0
}

// IsLess does a type-based comparison of itemI and itemJ values
func IsLess(itemI, itemJ reflect.Value) bool {
	switch {
	case IsIntValue(itemI) && IsIntValue(itemJ):
		return itemI.Int() < itemJ.Int()
	case IsFloatValue(itemI) && IsFloatValue(itemJ):
		return itemI.Float() < itemJ.Float()
	case IsIntValue(itemI) && IsFloatValue(itemJ):
		return float64(itemI.Int()) < itemJ.Float()
	case IsFloatValue(itemI) && IsIntValue(itemJ):
		return itemI.Float() < float64(itemJ.Int())
	case itemI.Type().Kind() == reflect.String && itemJ.Type().Kind() == reflect.String:
		return itemI.String() < itemJ.String()
	}
	return false
}
