package market_follower

import(
	"time"
)

type OrderBook struct {

	Timestamp time.Time

	Bids []Order
	Asks []Order

	//lastUpdateId int64
	//volumePrice float64
}

type Order struct {
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

type Trade struct {
	MakerSide MarketSide
	Price float64
	Volume float64
	Timestamp time.Time
}

