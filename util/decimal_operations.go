package util

import (
	"github.com/shopspring/decimal"
)

// DEPRECATED: User decimal.Decimal.Truncate instead
// RoundDown will round a decimal to a precision, always rounding towards zero
// E.g. rounding 1.36 to 1 d.p. gives 1.3
func RoundDown(d decimal.Decimal, precision int32) decimal.Decimal {

	return d.Truncate(precision)
}
