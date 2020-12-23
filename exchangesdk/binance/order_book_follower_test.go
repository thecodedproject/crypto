package binance_test

import (
	"github.com/shopspring/decimal"
	"github.com/thecodedproject/crypto/exchangesdk/binance"
	"github.com/thecodedproject/crypto/exchangesdk"
	"github.com/thecodedproject/gotest/assert"
	"testing"
)

func D(f float64) decimal.Decimal {

	return decimal.NewFromFloat(f)
}

func TestUpdateOrders(t *testing.T) {

	currentOrders := []exchangesdk.OrderBookOrder{
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

	expectedOrders := []exchangesdk.OrderBookOrder{
		{
			Price: 2.0,
			Volume: 2.1,
		},
	}

	binance.UpdateOrders(
		&currentOrders,
		updates,
		binance.ExchangeConfig{
			PricePrecision: 1e-2,
			VolPrecision: 1e-8,
		},
	)

	assert.LogicallyEqual(t, expectedOrders, currentOrders)
}
