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
