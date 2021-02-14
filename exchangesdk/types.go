package exchangesdk

import(
	"time"

	"github.com/shopspring/decimal"
)

//go:generate enumer -type=OrderBookSide -trimprefix=OrderBookSide -json -text -transform=snake

type OrderBook struct {

	Timestamp time.Time

	Bids []OrderBookOrder
	Asks []OrderBookOrder
}

// OrderBookOrder represents an order in the OrderBook
type OrderBookOrder struct {
	Price  float64 `json:"price,string"`
	Volume float64 `json:"volume,string"`
}

type OrderBookSide int

const (
	OrderBookSideUnknown OrderBookSide = iota
	OrderBookSideBid
	OrderBookSideAsk
	OrderBookSideSentinal
)

// OrderBookTrade represents a trade as seen in the OrderBook
type OrderBookTrade struct {
	MakerSide OrderBookSide
	Price float64
	Volume float64
	Timestamp time.Time
}

type StopLimitOrder struct {
	Side OrderBookSide
	StopPrice decimal.Decimal
	LimitPrice decimal.Decimal
	Volume decimal.Decimal
}
