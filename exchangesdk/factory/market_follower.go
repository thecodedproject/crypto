package factory

import (
	"context"
	"sync"
	"log"
	"github.com/thecodedproject/crypto"
	"github.com/thecodedproject/crypto/exchangesdk"
	"github.com/thecodedproject/crypto/exchangesdk/binance"
	"github.com/thecodedproject/crypto/exchangesdk/dummyclient"
	"github.com/thecodedproject/crypto/exchangesdk/luno"
)

func NewMarketFollower(
	ctx context.Context,
	wg *sync.WaitGroup,
	pair exchangesdk.Pair,
	apiAuth crypto.AuthConfig,
) (<-chan exchangesdk.OrderBook, <-chan exchangesdk.OrderBookTrade, error) {

	switch apiAuth.ApiExchange {
	case crypto.ExchangeDummyExchange:
		return dummyclient.NewMarketFollower(
			ctx,
			wg,
			pair,
		)
	case crypto.ExchangeLuno:
		return luno.NewOrderBookFollowerAndTradeStream(
			ctx,
			wg,
			pair,
			apiAuth.ApiKey,
			apiAuth.ApiSecret,
		)
	case crypto.ExchangeBinance:
		return binance.NewMarketFollower(
			ctx,
			wg,
			pair,
		)
	default:
		log.Fatal("NewMarketFollower: Unknown exchange")
		return nil, nil, nil
	}
}
