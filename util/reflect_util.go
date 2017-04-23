package util

import "reflect"

func IsNumericValue(val reflect.Value) bool {
	return (IsIntValue(val) || IsFloatValue(val))
}

func IsIntValue(val reflect.Value) bool {
	switch val.Type().Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return true
	}
	return false
}

func IsFloatValue(val reflect.Value) bool {
	switch val.Type().Kind() {
	case reflect.Float32, reflect.Float64:
		return true
	}
	return false
}
