package dummyclient

import (
	"context"
	"github.com/thecodedproject/crypto/exchangesdk"
	"sync"
	"time"

	"fmt"
)

func NewMarketFollower(
	ctx context.Context,
	wg *sync.WaitGroup,
	pair exchangesdk.Pair,
) (<-chan exchangesdk.OrderBook, <-chan exchangesdk.OrderBookTrade, error) {

	obf := make(chan exchangesdk.OrderBook, 1)
	tradeFollower := make(chan exchangesdk.OrderBookTrade, 1)

	go func() {
		for {
			select {
			case <-ctx.Done():
				fmt.Println("dummyclient: context Done")
				wg.Done()
				return
			case <-time.After(time.Second):
				obf <- exchangesdk.OrderBook{
					Timestamp: time.Now(),
					Bids: []exchangesdk.OrderBookOrder{
						{
							Price: 100.0,
							Volume: 1.0,
						},
					},
					Asks: []exchangesdk.OrderBookOrder{
						{
							Price: 200.0,
							Volume: 1.0,
						},
					},
				}
				tradeFollower <- exchangesdk.OrderBookTrade{
					Timestamp: time.Now(),
					MakerSide: exchangesdk.MarketSideBuy,
					Price: 150.0,
					Volume: 0.1,
				}
			}
		}
	}()

	return obf, tradeFollower, nil
}
