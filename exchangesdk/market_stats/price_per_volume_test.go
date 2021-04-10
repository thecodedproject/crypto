package market_stats_test

import (
	"testing"

	tfy_assert "github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/thecodedproject/crypto/exchangesdk"
	"github.com/thecodedproject/crypto/exchangesdk/market_stats"
	"github.com/thecodedproject/gotest/assert"
)

func TestVolumePrice(t *testing.T) {

	testCases := []struct {
		name                string
		orders              []exchangesdk.OrderBookOrder
		volume              float64
		precision           int32
		expectedPrice       float64
		expectedErrorString string
	}{
		{
			name: "single order over volume price",
			orders: []exchangesdk.OrderBookOrder{
				{
					Price:  2.5,
					Volume: 2.0,
				},
			},
			volume:        1.0,
			precision:     int32(2),
			expectedPrice: 2.5,
		},
		{
			name: "multipl orders over volume price with equal volumes takes average",
			orders: []exchangesdk.OrderBookOrder{
				{
					Price:  1.0,
					Volume: 0.5,
				},
				{
					Price:  2.0,
					Volume: 0.5,
				},
				{
					Price:  5.0,
					Volume: 0.5,
				},
				{
					Price:  4.0,
					Volume: 0.5,
				},
			},
			volume:        2.0,
			precision:     int32(2),
			expectedPrice: 3.0,
		},
		{
			name: "multipl orders over volume price with unequal volumes takes weighted average",
			orders: []exchangesdk.OrderBookOrder{
				{
					Price:  1.0,
					Volume: 1.0,
				},
				{
					Price:  2.0,
					Volume: 0.75,
				},
				{
					Price:  5.0,
					Volume: 2.0,
				},
				{
					Price:  4.0,
					Volume: 0.25,
				},
			},
			volume:        4.0,
			precision:     int32(2),
			expectedPrice: 3.375,
		},
		{
			name: "orders which exceeed the volume are ignored",
			orders: []exchangesdk.OrderBookOrder{
				{
					Price:  1.0,
					Volume: 1.0,
				},
				{
					Price:  2.0,
					Volume: 0.75,
				},
				{
					Price:  5.0,
					Volume: 2.0,
				},
				{
					Price:  4.0,
					Volume: 1.25,
				},
				{
					Price:  5.0,
					Volume: 2.0,
				},
				{
					Price:  4.0,
					Volume: 1.25,
				},
			},
			volume:        4.0,
			precision:     int32(3),
			expectedPrice: 3.375,
		},
		{
			name: "orders which don't make up volume returns ErrVolumePriceNotEnoughOrders",
			orders: []exchangesdk.OrderBookOrder{
				{
					Price:  1.0,
					Volume: 1.0,
				},
				{
					Price:  2.0,
					Volume: 0.75,
				},
				{
					Price:  4.0,
					Volume: 1.25,
				},
			},
			volume:              4.0,
			precision:           int32(3),
			expectedErrorString: market_stats.ErrVolumePriceNotEnoughOrders,
		},
	}

	for _, test := range testCases {
		t.Run(test.name, func(t *testing.T) {

			actualPrice, err := market_stats.VolumePrice(
				&test.orders,
				test.volume,
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
