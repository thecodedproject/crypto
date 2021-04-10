package exchangesdk

import (
	"context"
	"time"

	"github.com/shopspring/decimal"
)

// DEPRECATED: Use the MarketSide type instead
// TODO Replace this type with MarketSide
type OrderType string

const (
	OrderTypeBid OrderType = "BID"
	OrderTypeAsk OrderType = "ASK"
)

// TODO Rename to LimitOrder (this type represents a placed limit order only - and not a generic 'order')
// In general, we are moving to a place where there is no single `Order` type, but specialisations of Order
// i.e. LimitOrder, OrderBookOrder, StopLimitOrder.
type Order struct {
	Id        string          `json:"id"`
	Timestamp time.Time       `json:"timestamp"`
	Type      OrderType       `json:"type"`
	Price     decimal.Decimal `json:"price"`
	Volume    decimal.Decimal `json:"volume"`
}

type Trade struct {
	OrderId    string
	Timestamp  time.Time
	Price      decimal.Decimal
	Volume     decimal.Decimal
	BaseFee    decimal.Decimal
	CounterFee decimal.Decimal
	Type       OrderType
}

type Client interface {
	LatestPrice(ctx context.Context) (decimal.Decimal, error)

	GetOrderStatus(ctx context.Context, orderId string) (OrderStatus, error)
	GetTrades(ctx context.Context, page int64) ([]Trade, error)

	PostLimitOrder(ctx context.Context, order Order) (string, error)

	PostStopLimitOrder(ctx context.Context, o StopLimitOrder) (string, error)

	CancelOrder(ctx context.Context, orderId string) error

	// MakerFee returns the fee as a ratio (i.e. 1% returned as 0.01)
	MakerFee() decimal.Decimal

	CounterPrecision() int32
	BasePrecision() int32
}
