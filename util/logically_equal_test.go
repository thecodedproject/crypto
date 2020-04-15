package util_test

import (
	"errors"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/thecodedproject/crypto/util"
	"testing"
)

func TestLogicallyEqual(t *testing.T) {

	testCases := []struct{
		name string
		a interface{}
		b interface{}
		s []interface{}
		pass bool
	}{
		{
			name: "integers",
			a: int(1),
			b: int(1),
			pass: true,
		},
		{
			name: "shopspring decimals equal",
			a: decimal.NewFromFloat(2.0),
			b: decimal.NewFromFloat(20).Div(decimal.NewFromFloat(10)),
			pass: true,
		},
		{
			name: "shopspring decimals not equal",
			a: decimal.NewFromFloat(2.0),
			b: decimal.NewFromFloat(30).Div(decimal.NewFromFloat(10)),
			pass: false,
		},
		{
			name: "nil errors inside structs equal",
			a: struct{
				Err error
			}{},
			b: struct{
				Err error
			}{},
			pass: true,
		},
		{
			name: "nil errors inside structs not equal - one not initalised",
			a: struct{
				Err error
			}{},
			b: struct{
				Err error
			}{
				Err: errors.New("some error"),
			},
			pass: false,
		},
		{
			name: "shopspring decimals inside struct equal",
			a: struct{
				Field decimal.Decimal
			}{decimal.NewFromFloat(2)},
			b: struct{
				Field decimal.Decimal
			}{decimal.NewFromFloat(20).Div(decimal.NewFromFloat(10))},
			pass: true,
		},
		{
			name: "shopspring decimals inside struct not equal",
			a: struct{
				Field decimal.Decimal
			}{decimal.NewFromFloat(2)},
			b: struct{
				Field decimal.Decimal
			}{decimal.NewFromFloat(30).Div(decimal.NewFromFloat(10))},
			pass: false,
		},
		{
			name: "struct with no public fields is compared physically equal when equal",
			a: struct{
				privateOne int
				privateTwo string
			}{1, "a"},
			b: struct{
				privateOne int
				privateTwo string
			}{1, "a"},
			pass: true,
		},
		{
			name: "struct with no public fields is compared physically equal when not equal",
			a: struct{
				privateOne int
				privateTwo string
			}{1, "a"},
			b: struct{
				privateOne int
				privateTwo string
			}{1, "b"},
			pass: false,
		},
	}

	for _, test := range testCases {
		t.Run(test.name, func(t *testing.T) {

			var fakeT testing.T
			res := util.LogicallyEqual(&fakeT, test.a, test.b, test.s...)
			assert.Equal(t, test.pass, res)
		})
	}
}
