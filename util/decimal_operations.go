package util

import (
	"github.com/shopspring/decimal"
)

// RoundDown will round a decimal to a precision, always rounding towards zero
// E.g. rounding 1.36 to 1 d.p. gives 1.3
func RoundDown(d decimal.Decimal, precision int32) decimal.Decimal {

	if d.Sign() == 0 {
		return d
	} else if d.Sign() == 1 {
		d = d.Sub(decimal.New(5, -precision-1))
	} else {
		d = d.Add(decimal.New(5, -precision-1))
	}
	d = d.Round(precision)
	return d
}
