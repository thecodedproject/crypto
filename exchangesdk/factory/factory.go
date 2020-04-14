package factory

import (
	"context"
	"fmt"
	"github.com/thecodedproject/crypto/exchangesdk"
	"github.com/thecodedproject/crypto/exchangesdk/luno"
	"github.com/thecodedproject/crypto/exchangesdk/binance"
)

func NewClient(
	ctx context.Context,
	exchangeName string,
	apiKey string,
	apiSecret string,
) (exchangesdk.Client, error) {

	switch exchangeName {
	case "luno":
		return luno.NewClient(
			ctx,
			apiKey,
			apiSecret,
		)
	case "binance":
		return binance.NewClient(
			apiKey,
			apiSecret,
		)
	case "dummyclient":
		return dummyclient.NewClient(
				apiKey,
				apiSecret,
			)
	default:
		return nil, fmt.Errorf("Cannot create client; Unknown exchange %s", exchangeName)
	}

}
