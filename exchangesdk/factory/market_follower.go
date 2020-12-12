package factory

import (
	"log"
	"github.com/thecodedproject/crypto"
	"github.com/thecodedproject/crypto/exchangesdk"
	"github.com/thecodedproject/crypto/exchangesdk/binance"
	"github.com/thecodedproject/crypto/exchangesdk/luno"
)

func NewMarketFollower(
	pair exchangesdk.Pair,
	apiAuth crypto.AuthConfig,
) (<-chan exchangesdk.OrderBook, <-chan exchangesdk.OrderBookTrade) {

	switch apiAuth.ApiExchange {
	case crypto.ExchangeLuno:
		return luno.NewOrderBookFollowerAndTradeStream(
			pair,
			apiAuth.ApiKey,
			apiAuth.ApiSecret,
		)
	case crypto.ExchangeBinance:
		return binance.NewOrderBookFollowerAndTradeStream(
			pair,
		)
	default:
		log.Fatal("NewMarketFollower: Unknown exchange")
		return nil, nil
	}
}
