package dummyclient

import (
	"context"
	"sync"
	"time"

	"github.com/thecodedproject/crypto"
	"github.com/thecodedproject/crypto/exchangesdk"
)

func NewMarketFollower(
	ctx context.Context,
	wg *sync.WaitGroup,
	_ crypto.Pair,
) (<-chan exchangesdk.OrderBook, <-chan exchangesdk.OrderBookTrade, error) {

	obf := make(chan exchangesdk.OrderBook, 1)
	tradeFollower := make(chan exchangesdk.OrderBookTrade, 1)

	go func() {
		for {
			select {
			case <-ctx.Done():
				wg.Done()
				return
			case <-time.After(time.Second):
				obf <- exchangesdk.OrderBook{
					Timestamp: time.Now(),
					Bids: []exchangesdk.OrderBookOrder{
						{
							Price:  100.0,
							Volume: 1.0,
						},
					},
					Asks: []exchangesdk.OrderBookOrder{
						{
							Price:  200.0,
							Volume: 1.0,
						},
					},
				}
				tradeFollower <- exchangesdk.OrderBookTrade{
					Timestamp: time.Now(),
					MakerSide: exchangesdk.OrderBookSideBid,
					Price:     150.0,
					Volume:    0.1,
				}
			}
		}
	}()

	return obf, tradeFollower, nil
}
