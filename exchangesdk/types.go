package exchangesdk

import(
	"time"
)

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

type MarketSide int

const (
	MarketSideUnknown MarketSide = iota
	MarketSideBuy
	MarketSideSell
	tradeSideSentinal
)

func (s MarketSide) String() string {

	switch s {
	case MarketSideBuy:
		return "MarketSideBuy"
	case MarketSideSell:
		return "MarketSideSell"
	default:
		return "MarketSideUnknown"
	}
}

// OrderBookTrade represents a trade as seen in the OrderBook
type OrderBookTrade struct {
	MakerSide MarketSide
	Price float64
	Volume float64
	Timestamp time.Time
}

