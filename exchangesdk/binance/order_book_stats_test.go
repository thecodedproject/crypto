package binance_test

import (
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/require"
	"github.com/thecodedproject/crypto/exchangesdk/binance"
	"github.com/thecodedproject/gotest/assert"
	"testing"
	tfy_assert "github.com/stretchr/testify/assert"
)

func TestVolumePrice(t *testing.T) {

	testCases := []struct {
		name string
		orders []binance.Order
		volume decimal.Decimal
		precision int32
		expectedPrice decimal.Decimal
		expectedErrorString string
	}{
		{
			name: "single order over volume price",
			orders: []binance.Order{
				{
					Price: D(2.5),
					Volume: D(2.0),
				},
			},
			volume: D(1.0),
			precision: int32(2),
			expectedPrice: D(2.5),
		},
		{
			name: "multipl orders over volume price with equal volumes takes average",
			orders: []binance.Order{
				{
					Price: D(1.0),
					Volume: D(0.5),
				},
				{
					Price: D(2.0),
					Volume: D(0.5),
				},
				{
					Price: D(5.0),
					Volume: D(0.5),
				},
				{
					Price: D(4.0),
					Volume: D(0.5),
				},
			},
			volume: D(2.0),
			precision: int32(2),
			expectedPrice: D(3.0),
		},
		{
			name: "multipl orders over volume price with unequal volumes takes weighted average",
			orders: []binance.Order{
				{
					Price: D(1.0),
					Volume: D(1.0),
				},
				{
					Price: D(2.0),
					Volume: D(0.75),
				},
				{
					Price: D(5.0),
					Volume: D(2.0),
				},
				{
					Price: D(4.0),
					Volume: D(0.25),
				},
			},
			volume: D(4.0),
			precision: int32(2),
			expectedPrice: D(3.38),
		},
		{
			name: "orders which exceeed the volume are ignored",
			orders: []binance.Order{
				{
					Price: D(1.0),
					Volume: D(1.0),
				},
				{
					Price: D(2.0),
					Volume: D(0.75),
				},
				{
					Price: D(5.0),
					Volume: D(2.0),
				},
				{
					Price: D(4.0),
					Volume: D(1.25),
				},
				{
					Price: D(5.0),
					Volume: D(2.0),
				},
				{
					Price: D(4.0),
					Volume: D(1.25),
				},
			},
			volume: D(4.0),
			precision: int32(3),
			expectedPrice: D(3.375),
		},
		{
			name: "orders which don't make up volume returns ErrVolumePriceNotEnoughOrders",
			orders: []binance.Order{
				{
					Price: D(1.0),
					Volume: D(1.0),
				},
				{
					Price: D(2.0),
					Volume: D(0.75),
				},
				{
					Price: D(4.0),
					Volume: D(1.25),
				},
			},
			volume: D(4.0),
			precision: int32(3),
			expectedErrorString: binance.ErrVolumePriceNotEnoughOrders,
		},
	}

	for _, test := range testCases {
		t.Run(test.name, func(t *testing.T) {

			actualPrice, err := binance.VolumePrice(
				&test.orders,
				test.volume,
				test.precision,
			)

			if test.expectedErrorString != "" {
				require.Error(t, err)
				tfy_assert.Equal(t, test.expectedErrorString, err.Error())
				return
			}

			require.NoError(t, err)

			assert.LogicallyEqual(t, test.expectedPrice, actualPrice)
		})
	}
}
