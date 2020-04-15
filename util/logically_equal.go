package util

import (
	"fmt"
	"github.com/shopspring/decimal"
  "github.com/stretchr/testify/assert"
	"reflect"
	"testing"
)

// TODO Add support for comparing pointers and arrays
func LogicallyEqual(t *testing.T, a, b interface{}, s ...interface{}) bool {

	if a == nil || b == nil {
		return assert.Equal(t, a, b, s...)
	}

	aType := reflect.TypeOf(a)
	bType := reflect.TypeOf(b)

	if isShopspringDecimal(aType) && isShopspringDecimal(bType) {

		message := fmt.Sprint(s...)
		return assert.Equalf(t, 0, a.(decimal.Decimal).Cmp(b.(decimal.Decimal)),
			"%s: Decimals not equal.\n\tExpected: %s\n\tActual: %s", message, a, b)

	}
	if aType == bType && aType.Kind() == reflect.Struct {

		aValue := reflect.ValueOf(a)
		bValue := reflect.ValueOf(b)
		retVal := true
		publicFields := 0
		for i:=0; i<aType.NumField(); i++ {
			if aValue.Field(i).CanInterface() {
				fieldName := aType.Field(i).Name
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
	}

	return assert.Equal(t, a, b, s...)
}

func isShopspringDecimal(t reflect.Type) bool {

	return t.PkgPath() == "github.com/shopspring/decimal" && t.Name() == "Decimal"
}
