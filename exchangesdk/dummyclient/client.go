package dummyclient

import (
	"context"
	"math/rand"

	"github.com/shopspring/decimal"
	"github.com/thecodedproject/crypto"
	"github.com/thecodedproject/crypto/exchangesdk"
)

type client struct {
	lastOrderLimitPrice decimal.Decimal
	lastOrderVolume decimal.Decimal
	lastOrderSide exchangesdk.OrderBookSide
	exchange        crypto.Exchange
}

func NewClient(
	apiKey string,
	apiSecret string,
	exchange crypto.Exchange,
) (*client, error) {

	return &client{
		exchange: exchange,
	}, nil
}

func (c *client) Exchange() crypto.Exchange {

	return c.exchange
}

func (c *client) LatestPrice(ctx context.Context) (decimal.Decimal, error) {

	return decimal.NewFromFloat(123.4), nil
}

func (c *client) PostLimitOrder(ctx context.Context, order exchangesdk.Order) (string, error) {

	c.lastOrderLimitPrice = order.Price
	c.lastOrderVolume = order.Volume
	return "some_order_id", nil
}

func (c *client) PostStopLimitOrder(ctx context.Context, order exchangesdk.StopLimitOrder) (string, error) {

	c.lastOrderLimitPrice = order.LimitPrice
	c.lastOrderVolume = order.Volume
	c.lastOrderSide = order.Side
	return "some_order_id", nil
}

func (c *client) CancelOrder(ctx context.Context, orderId string) error {

	return nil
}

func (c *client) GetOrderStatus(
	ctx context.Context,
	orderId string,
) (exchangesdk.OrderStatus, error) {

	if r := rand.Float64(); r < 0.3 {
		return exchangesdk.OrderStatus{
			State: exchangesdk.OrderStateAwaitingTrigger,
		}, nil
	} else if r < 0.6 {
		return exchangesdk.OrderStatus{
			State: exchangesdk.OrderStateInOrderBook,
		}, nil
	} else {
		orderType := exchangesdk.OrderTypeBid
		if c.lastOrderSide == exchangesdk.OrderBookSideAsk {
			orderType = exchangesdk.OrderTypeAsk
		}

		return exchangesdk.OrderStatus{
			State:          exchangesdk.OrderStateFilled,
			FillAmountBase: c.lastOrderVolume,
			FillAmountCounter: c.lastOrderLimitPrice.Mul(c.lastOrderVolume),
			Type: orderType,
		}, nil
	}
}

func (c *client) GetTrades(ctx context.Context, page int64) ([]exchangesdk.Trade, error) {

	panic("not implemented")
}

func (c *client) MakerFee() decimal.Decimal {

	return decimal.New(75, -5)
}

func (c *client) TakerFee() decimal.Decimal {

	return decimal.New(75, -5)
}

func (c *client) CounterPrecision() int32 {

	return 2
}

func (c *client) BasePrecision() int32 {

	return 6
}
