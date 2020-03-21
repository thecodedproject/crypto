package util_test

import (
	"github.com/shopspring/decimal"
	"github.com/thecodedproject/crypto/util"
	"testing"
)

func TestRoundDown(t *testing.T) {

	testCases := []struct{
		name string
		input decimal.Decimal
		precision int32
		expected decimal.Decimal
	}{
		{
			name: "Zero returns zero",
			input: decimal.Decimal{},
			precision: 3,
			expected: decimal.Decimal{},
		},
		{
			name: "Already rounded number",
			input: decimal.New(2, 0),
			precision: 0,
			expected: decimal.New(2, 0),
		},
		{
			name: "Above .5 gets rounded toward zero - 2 d.p.",
			input: decimal.New(2466, -3),
			precision: 2,
			expected: decimal.New(246, -2),
		},
		{
			name: "Exactly .5 gets rounded toward zero - 2 d.p.",
			input: decimal.New(2465, -3),
			precision: 2,
			expected: decimal.New(246, -2),
		},
		{
			name: "Below .5 gets rounded toward zero - 2 d.p.",
			input: decimal.New(2462, -3),
			precision: 2,
			expected: decimal.New(246, -2),
		},
		{
			name: "Exactly .0 remains the same - 2 d.p.",
			input: decimal.New(2460, -3),
			precision: 2,
			expected: decimal.New(246, -2),
		},
		{
			name: "Above .5 gets rounded toward zero - 6 d.p.",
			input: decimal.New(24668369, -7),
			precision: 6,
			expected: decimal.New(2466836, -6),
		},
		{
			name: "Exactly .5 gets rounded toward zero - 6 d.p.",
			input: decimal.New(6437955, -7),
			precision: 6,
			expected: decimal.New(643795, -6),
		},
		{
			name: "Below .5 gets rounded toward zero - 6 d.p.",
			input: decimal.New(97439605, -7),
			precision: 6,
			expected: decimal.New(9743960, -6),
		},
		{
			name: "Exactly .0 remains the same - 6 d.p.",
			input: decimal.New(1804560, -7),
			precision: 6,
			expected: decimal.New(180456, -6),
		},
		{
			name: "Negative above .5 gets rounded toward zero - 3 d.p.",
			input: decimal.New(-90046, -4),
			precision: 3,
			expected: decimal.New(-9004, -3),
		},
		{
			name: "Negative exactly .5 gets rounded toward zero - 3 d.p.",
			input: decimal.New(-245605, -4),
			precision: 3,
			expected: decimal.New(-24560, -3),
		},
		{
			name: "Negative below .5 gets rounded toward zero - 3 d.p.",
			input: decimal.New(-13794, -4),
			precision: 3,
			expected: decimal.New(-1379, -3),
		},
		{
			name: "Negative exactly .0 remains the same - 3 d.p.",
			input: decimal.New(-96740, -4),
			precision: 3,
			expected: decimal.New(-9674, -3),
		},
		{
			name: "Negative above .5 gets rounded toward zero - 9 d.p.",
			input: decimal.New(-9467043268, -10),
			precision: 9,
			expected: decimal.New(-946704326, -9),
		},
		{
			name: "Negative exactly .5 gets rounded toward zero - 9 d.p.",
			input: decimal.New(-8628514015, -10),
			precision: 9,
			expected: decimal.New(-862851401, -9),
		},
		{
			name: "Negative below .5 gets rounded toward zero - 9 d.p.",
			input: decimal.New(-1855200152, -10),
			precision: 9,
			expected: decimal.New(-185520015, -9),
		},
		{
			name: "Negative exactly .0 remains the same - 9 d.p.",
			input: decimal.New(-1759004720, -10),
			precision: 9,
			expected: decimal.New(-175900472, -9),
		},
	}

	for _, test := range testCases {
		t.Run(test.name, func(t *testing.T) {

			actual := util.RoundDown(test.input, test.precision)
			util.LogicallyEqual(t, test.expected, actual)
		})
	}
}
