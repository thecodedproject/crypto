package util

import (
	"fmt"
	"github.com/shopspring/decimal"
  "github.com/stretchr/testify/assert"
	"reflect"
	"testing"
	"sort"
)

// TODO Add support for comparing pointers and arrays
func LogicallyEqual(t *testing.T, a, b interface{}, s ...interface{}) bool {

	if a == nil || b == nil {
		return assert.Equal(t, a, b, s...)
	}

	aType := reflect.TypeOf(a)
	bType := reflect.TypeOf(b)

	if aType != bType {
		return assert.Equal(t, a, b, s...)
	}

	if isShopspringDecimal(aType) {

		message := fmt.Sprint(s...)
		return assert.Equalf(t, 0, a.(decimal.Decimal).Cmp(b.(decimal.Decimal)),
			"%s: Decimals not equal.\n\tExpected: %s\n\tActual: %s", message, a, b)

	}

	switch aType.Kind() {
	case reflect.Struct:
		return structsLogicallyEqual(t, a, b, s...)
	case reflect.Map:
		return mapsLogicallyEqual(t, a, b, s...)
	default:
		return assert.Equal(t, a, b, s...)
	}
}

func isShopspringDecimal(t reflect.Type) bool {

	return t.PkgPath() == "github.com/shopspring/decimal" && t.Name() == "Decimal"
}

func structsLogicallyEqual(
	t *testing.T,
	a interface{},
	b interface{},
	s ...interface{},
) bool {

	aValue := reflect.ValueOf(a)
	bValue := reflect.ValueOf(b)
	retVal := true
	publicFields := 0
	for i:=0; i<aValue.Type().NumField(); i++ {
		if aValue.Field(i).CanInterface() {
			fieldName := aValue.Type().Field(i).Name
			aField := aValue.Field(i).Interface()
			bField := bValue.Field(i).Interface()

			messageAndFieldName := append(s, "."+fieldName)
			retVal = retVal && LogicallyEqual(t, aField, bField, messageAndFieldName...)
			publicFields++
		}
	}

	if publicFields > 0 {
		return retVal
	}

	return assert.Equal(t, a, b, s...)
}

func mapsLogicallyEqual(
	t *testing.T,
	a interface{},
	b interface{},
	s ...interface{},
) bool {

	aValue := reflect.ValueOf(a)
	bValue := reflect.ValueOf(b)

	lengthMsg := append([]interface{}{"Length of map"}, s...)
	ok := assert.Equal(t, aValue.Len(), bValue.Len(), lengthMsg...)
	if !ok {
		return false
	}

	keysMsg := append([]interface{}{"Keys of map"}, s...)
	ok = assert.Equal(
		t,
		sortedMapKeys(aValue),
		sortedMapKeys(bValue),
		keysMsg...,
	)
	if !ok {
		return false
	}

	retval := true
	for _, key := range aValue.MapKeys() {
		messageAndFieldName := append(s, ".['"+key.String()+"']")

		retval = retval && LogicallyEqual(
			t,
			aValue.MapIndex(key).Interface(),
			bValue.MapIndex(key).Interface(),
			messageAndFieldName...,
		)
	}

	return retval
}

func sortedMapKeys(value reflect.Value) []string {

	mapKeys := value.MapKeys()
	mapKeysStr := make([]string, 0, len(mapKeys))
	for _, keyVal := range mapKeys {
		mapKeysStr = append(mapKeysStr, keyVal.String())
	}
	sort.Strings(mapKeysStr)
	return mapKeysStr
}
