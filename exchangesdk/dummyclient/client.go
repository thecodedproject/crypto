package dummyclient

import (
	"context"
	"encoding/json"
	"github.com/shopspring/decimal"
	"github.com/thecodedproject/crypto/exchangesdk"
	"github.com/thecodedproject/crypto/exchangesdk/requestutil"
	utiltime "github.com/thecodedproject/crypto/util/time"
	"io/ioutil"
	"net/http"
	"net/url"
	"testing"
	"strconv"
	"time"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
)

type marketPair string

const (
	BTCEUR marketPair = "BTCEUR"
)

const (
	baseUrl = "https://api.binance.com"
)

type client struct {
}

func NewClient(apiKey, apiSecret string) (*client, error) {

	return &client{} nil
}

func (c *client) LatestPrice(ctx context.Context) (decimal.Decimal, error) {

	return decimal.NewFromFloat(123.4), nil
}

func (c *client) PostLimitOrder(ctx context.Context, order exchangesdk.Order) (string, error) {

	return "some_order_id", nil
}

func (c *client) StopOrder(ctx context.Context, orderId string) error {

	return nil
}

func (c *client) GetOrderStatus(
	ctx context.Context,
	orderId string,
) (exchangesdk.OrderStatus, error) {

	return exchangesdk.OrderStatus{
		State: exchangesdk.OrderStatePending,
	}, nil
}

func (c *client) GetTrades(ctx context.Context, page int64) ([]exchangesdk.Trade, error) {

	panic("not implemented")
}

func (c *client) MakerFee() decimal.Decimal {

	return decimal.New(75, -5)
}

func (c *client) CounterPrecision() int32 {

	return 2
}

func (c *client) BasePrecision() int32 {

	return 6
}
