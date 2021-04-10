package util_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/thecodedproject/crypto/util"
)

func TestFloatNear(t *testing.T) {

	testCases := []struct {
		name     string
		a        float64
		b        float64
		expected bool
	}{
		{
			name:     "Equal floats returns true",
			a:        1.0,
			b:        1.0,
			expected: true,
		},
		{
			name:     "Close floats returns true",
			a:        1.0,
			b:        1.000000000000001,
			expected: true,
		},
		{
			name:     "Not-so-close floats returns false",
			a:        1.0,
			b:        1.000000000000005,
			expected: false,
		},
		{
			name:     "Large close floats returns true",
			a:        9999999999999999,
			b:        9999999999999991,
			expected: true,
		},
		{
			name:     "Large not-so-close floats returns false",
			a:        9999999999999999,
			b:        9999999999999985,
			expected: false,
		},
		{
			name:     "Very large close floats returns true",
			a:        9.999999999999999e300,
			b:        9.999999999999995e300,
			expected: true,
		},
		{
			name:     "Very large not-so-close floats returns false",
			a:        9.999999999999999e300,
			b:        9.999999999999990e300,
			expected: false,
		},
	}

	for _, test := range testCases {

		t.Run(test.name, func(t *testing.T) {
			assert.Equal(t, test.expected, util.Float64Near(test.a, test.b))
		})
	}
}
