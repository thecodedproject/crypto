package dummyclient

import (
	"context"
	"math/rand"
	"github.com/shopspring/decimal"
	"github.com/thecodedproject/crypto"
	"github.com/thecodedproject/crypto/exchangesdk"
)

type marketPair string

const (
	BTCEUR marketPair = "BTCEUR"
)

const (
	baseUrl = "https://api.binance.com"
)

type client struct {
	lastOrderVolume decimal.Decimal
}

func NewClient(
	apiKey string,
	apiSecret string,
	pair crypto.Pair,
) (*client, error) {

	return &client{}, nil
}

func (c *client) LatestPrice(ctx context.Context) (decimal.Decimal, error) {

	return decimal.NewFromFloat(123.4), nil
}

func (c *client) PostLimitOrder(ctx context.Context, order exchangesdk.Order) (string, error) {

	c.lastOrderVolume = order.Volume
	return "some_order_id", nil
}

func (c *client) PostStopLimitOrder(ctx context.Context, order exchangesdk.StopLimitOrder) (string, error) {

	c.lastOrderVolume = order.Volume
	return "some_order_id", nil
}

func (c *client) CancelOrder(ctx context.Context, orderId string) error {

	return nil
}

func (c *client) GetOrderStatus(
	ctx context.Context,
	orderId string,
) (exchangesdk.OrderStatus, error) {

	if rand.Float64() < 0.5 {
		return exchangesdk.OrderStatus{
			State: exchangesdk.OrderStatePending,
		}, nil
	} else {
		return exchangesdk.OrderStatus{
			State: exchangesdk.OrderStateComplete,
			FillAmountBase: c.lastOrderVolume,
		}, nil
	}
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
