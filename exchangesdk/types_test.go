package exchangesdk_test

import (
	"testing"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/thecodedproject/crypto/exchangesdk"
)

func TestOrderStatusAverageFillPrice(t *testing.T) {

	testCases := []struct {
		Name     string
		Status   exchangesdk.OrderStatus
		Expected decimal.Decimal
	}{
		{
			Name: "Zero fill amounts returns zero",
		},
		{
			Name: "Zero base fill amounts returns zero",
			Status: exchangesdk.OrderStatus{
				FillAmountCounter: decimal.New(456, -1),
			},
		},
		{
			Name: "Non-zero fill amounts returns correct price",
			Status: exchangesdk.OrderStatus{
				FillAmountBase:    decimal.New(125, -1),
				FillAmountCounter: decimal.New(456, -1),
			},
			Expected: decimal.New(3648, -3),
		},
	}

	for _, test := range testCases {
		t.Run(test.Name, func(t *testing.T) {
			assert.True(
				t,
				test.Expected.Equal(test.Status.AverageFillPrice()),
				"Not equal",
			)
		})
	}
}
