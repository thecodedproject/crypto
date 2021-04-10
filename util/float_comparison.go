package util

import (
	"math"
)

const MAX_FLOAT64_DISTANCE = 5

// Float64Near returns true if values are nearly equal, regardless fo magnitude
// Based on the excellent post (and blog): https://randomascii.wordpress.com/2012/02/25/comparing-floating-point-numbers-2012-edition/
func Float64Near(a, b float64) bool {

	aBits := math.Float64bits(a)
	bBits := math.Float64bits(b)

	aInt := int64(aBits)
	bInt := int64(bBits)

	if int64(math.Abs(float64(aInt-bInt))) > MAX_FLOAT64_DISTANCE {
		return false
	}
	return true
}
