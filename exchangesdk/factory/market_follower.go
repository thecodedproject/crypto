package factory

import (
	"context"
	"log"
	"sync"

	"github.com/thecodedproject/crypto"
	"github.com/thecodedproject/crypto/exchangesdk"
	"github.com/thecodedproject/crypto/exchangesdk/binance"
	"github.com/thecodedproject/crypto/exchangesdk/dummyclient"
	"github.com/thecodedproject/crypto/exchangesdk/luno"
)

func NewMarketFollower(
	ctx context.Context,
	wg *sync.WaitGroup,
	exchange crypto.Exchange,
	apiAuth crypto.AuthConfig,
) (<-chan exchangesdk.OrderBook, <-chan exchangesdk.OrderBookTrade, error) {

	switch exchange.Provider {
	case crypto.ApiProviderDummyExchange:
		return dummyclient.NewMarketFollower(
			ctx,
			wg,
			exchange.Pair,
		)
	case crypto.ApiProviderDummyExchangeBinanceMarket:
		return binance.NewMarketFollower(
			ctx,
			wg,
			exchange.Pair,
		)
	case crypto.ApiProviderLuno:
		return luno.NewOrderBookFollowerAndTradeStream(
			ctx,
			wg,
			exchange.Pair,
			apiAuth.Key,
			apiAuth.Secret,
		)
	case crypto.ApiProviderBinance:
		return binance.NewMarketFollower(
			ctx,
			wg,
			exchange.Pair,
		)
	default:
		log.Fatal("NewMarketFollower: Unknown exchange")
		return nil, nil, nil
	}
}
