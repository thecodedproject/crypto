package luno_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/thecodedproject/crypto/exchangesdk/luno"
)

func TestHandleUpdate(t *testing.T) {

	testCases := []struct {
		Name              string
		OrderBook         luno.InternalOrderBook
		Update            luno.OrderBookUpdate
		ExpectedOrderBook luno.InternalOrderBook
	}{
		{
			Name: "Empty update gives identical order book",
			OrderBook: luno.InternalOrderBook{
				Bids: map[string]luno.Order{
					"b1": luno.Order{
						Price: 1.0,
					},
				},
				Asks: map[string]luno.Order{
					"a1": luno.Order{
						Price: 1.0,
					},
				},
			},
			ExpectedOrderBook: luno.InternalOrderBook{
				Bids: map[string]luno.Order{
					"b1": luno.Order{
						Price: 1.0,
					},
				},
				Asks: map[string]luno.Order{
					"a1": luno.Order{
						Price: 1.0,
					},
				},
			},
		},
		{
			Name: "With trade update",
			Update: luno.OrderBookUpdate{
				Sequence: 2,
				TradeUpdates: []*luno.TradeUpdate{
					{
						Base:         0.25,
						MakerOrderId: "a1",
					},
					{
						Base:         1.0,
						MakerOrderId: "b1",
					},
				},
			},
			OrderBook: luno.InternalOrderBook{
				LastSequenceId: 1,
				Bids: map[string]luno.Order{
					"b1": luno.Order{
						Volume: 1.0,
					},
				},
				Asks: map[string]luno.Order{
					"a1": luno.Order{
						Volume: 1.0,
					},
				},
			},
			ExpectedOrderBook: luno.InternalOrderBook{
				LastSequenceId: 2,
				Bids:           map[string]luno.Order{},
				Asks: map[string]luno.Order{
					"a1": luno.Order{
						Volume: 0.75,
					},
				},
			},
		},
		{
			Name: "With create update",
			Update: luno.OrderBookUpdate{
				Sequence: 2,
				CreateUpdate: &luno.CreateUpdate{
					OrderId:   "b2",
					OrderType: "BID",
					Volume:    1.4,
				},
			},
			OrderBook: luno.InternalOrderBook{
				LastSequenceId: 1,
				Bids: map[string]luno.Order{
					"b1": luno.Order{
						Volume: 1.0,
					},
				},
				Asks: map[string]luno.Order{
					"a1": luno.Order{
						Volume: 1.0,
					},
				},
			},
			ExpectedOrderBook: luno.InternalOrderBook{
				LastSequenceId: 2,
				Bids: map[string]luno.Order{
					"b1": luno.Order{
						Volume: 1.0,
					},
					"b2": luno.Order{
						Id:     "b2",
						Volume: 1.4,
					},
				},
				Asks: map[string]luno.Order{
					"a1": luno.Order{
						Volume: 1.0,
					},
				},
			},
		},
		{
			Name: "With delete update",
			Update: luno.OrderBookUpdate{
				Sequence: 2,
				DeleteUpdate: &luno.DeleteUpdate{
					OrderId: "a1",
				},
			},
			OrderBook: luno.InternalOrderBook{
				LastSequenceId: 1,
				Bids: map[string]luno.Order{
					"b1": luno.Order{
						Volume: 1.0,
					},
				},
				Asks: map[string]luno.Order{
					"a1": luno.Order{
						Volume: 1.0,
					},
				},
			},
			ExpectedOrderBook: luno.InternalOrderBook{
				LastSequenceId: 2,
				Bids: map[string]luno.Order{
					"b1": luno.Order{
						Volume: 1.0,
					},
				},
				Asks: map[string]luno.Order{},
			},
		},
	}

	for _, test := range testCases {
		t.Run(test.Name, func(t *testing.T) {

			_, err := luno.HandleUpdate(&test.OrderBook, test.Update, 1e-8)
			require.NoError(t, err)

			assert.Equal(t, test.ExpectedOrderBook, test.OrderBook)

		})
	}
}
