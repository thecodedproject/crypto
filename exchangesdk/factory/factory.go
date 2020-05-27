package factory

import (
	"context"
	"fmt"
	"github.com/thecodedproject/crypto/exchangesdk"
	"github.com/thecodedproject/crypto/exchangesdk/binance"
	"github.com/thecodedproject/crypto/exchangesdk/dummyclient"
	"github.com/thecodedproject/crypto/exchangesdk/luno"
)

func NewClient(
	exchangeName string,
	apiKey string,
	apiSecret string,
) (exchangesdk.Client, error) {

	switch exchangeName {
	case "luno":
		return luno.NewClient(
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
