package exchangesdk

import (
	"context"
	"github.com/shopspring/decimal"
	"time"
)

type OrderType string

const (
	OrderTypeBid OrderType = "BID"
	OrderTypeAsk OrderType = "ASK"
)

type OrderState string

const (
	OrderStatePending OrderState = "PENDING"
	OrderStateComplete OrderState = "COMPLETE"
)

type OrderStatus struct {
	State OrderState
	Type OrderType
	FillAmountBase decimal.Decimal
}

// TODO Refactor to create a seperate OrderRequest struct which ocntains only the
// price volume and type, then use that as arg for `PostLimitOrder`
type Order struct {
	Id string `json:"id"`
	Timestamp time.Time `json:"timestamp"`
	Type OrderType `json:"type"`
	Price decimal.Decimal `json:"price"`
	Volume decimal.Decimal `json:"volume"`
}

type Trade struct {
	OrderId string
	Timestamp time.Time
	Price decimal.Decimal
	Volume decimal.Decimal
	BaseFee decimal.Decimal
	CounterFee decimal.Decimal
	Type OrderType
}

type Client interface {
	LatestPrice(ctx context.Context) (decimal.Decimal, error)

	GetOrderStatus(ctx context.Context, orderId string) (OrderStatus, error)
	GetTrades(ctx context.Context, page int64) ([]Trade, error)

	PostLimitOrder(ctx context.Context, order Order) (string, error)
	StopOrder(ctx context.Context, orderId string) error

	// MakerFee returns the fee as a ratio (i.e. 1% returned as 0.01)
	MakerFee() decimal.Decimal

	CounterPrecision() int32
	BasePrecision() int32
}
