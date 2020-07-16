package binance_test

import (
	"github.com/shopspring/decimal"
	"github.com/thecodedproject/crypto/exchangesdk/binance"
	"github.com/thecodedproject/gotest/assert"
	"testing"
)

func D(f float64) decimal.Decimal {

	return decimal.NewFromFloat(f)
}

func TestUpdateOrders(t *testing.T) {

	currentOrders := []binance.Order{
		{
			Price: 1.0,
			Volume: 1.1,
		},
		{
			Price: 2.0,
			Volume: 3.1,
		},
	}

	updates := [][]string{
		{"1.0", "0"},
		{"2.0", "2.1"},
	}

	expectedOrders := []binance.Order{
		{
			Price: 2.0,
			Volume: 2.1,
		},
	}

	binance.UpdateOrders(&currentOrders, updates)

	assert.LogicallyEqual(t, expectedOrders, currentOrders)
}
