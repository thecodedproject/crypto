package profitloss

import (
	"github.com/shopspring/decimal"
)

type CalcType int

const (
	CalcTypeAverage = 0
	CalcTypeUnknown = 1
	calcTypeSentinal = 2
)

type OrderType string

const (
	OrderTypeBid = "BID"
	OrderTypeAsk = "ASK"
)

type CompletedOrder struct {
	Price decimal.Decimal
	Volume decimal.Decimal
	BaseFee decimal.Decimal
	CounterFee decimal.Decimal
	Type OrderType
}

type Report struct {
	Type CalcType `json:"type"`
	BaseBought decimal.Decimal `json:"base_bought"`
	BaseSold decimal.Decimal `json:"base_sold"`
	BaseFees decimal.Decimal `json:"base_fees"`
	CounterBought decimal.Decimal `json:"counter_bought"`
	CounterSold decimal.Decimal `json:"counter_sold"`
	CounterFees decimal.Decimal `json:"counter_fees"`
	OrderCount int64 `json:"order_count"`
}

func (r Report) RealisedGain() decimal.Decimal {
	volumeForRealisedGain := decimal.Min(r.BaseBought, r.BaseSold)
	return r.AverageSellPrice().Sub(r.AverageBuyPrice()).Mul(volumeForRealisedGain).Sub(r.CounterFees)
}

func (r Report) UnrealisedGain(marketPrice decimal.Decimal) decimal.Decimal {
	return marketPrice.Sub(r.AverageBuyPrice()).Div(r.BaseBalance())
}

func (r Report) AverageBuyPrice() decimal.Decimal {
	return r.CounterSold.Div(r.BaseBought)
}

func (r Report) AverageSellPrice() decimal.Decimal {
	return r.CounterBought.Div(r.BaseSold)
}

func (r Report) BaseBalance() decimal.Decimal {
	return r.BaseBought.Sub(r.BaseSold).Sub(r.BaseFees)
}

func (r Report) CounterBalance() decimal.Decimal {
	return r.CounterBought.Sub(r.CounterSold).Sub(r.CounterFees)
}

func (r Report) TotalVolume() decimal.Decimal {
	return r.BaseSold.Add(r.BaseBought)
}

func Add(r Report, orders ...CompletedOrder) Report {

	for _, o := range orders {
		orderCost := o.Volume.Mul(o.Price)

		if o.Type == OrderTypeBid {
			r.BaseBought = r.BaseBought.Add(o.Volume)
			r.CounterSold = r.CounterSold.Add(orderCost)
		}	else {
			r.CounterBought = r.CounterBought.Add(orderCost)
			r.BaseSold = r.BaseSold.Add(o.Volume)
		}

		r.CounterFees = r.CounterFees.Add(o.CounterFee)
		r.BaseFees = r.BaseFees.Add(o.BaseFee)
		r.OrderCount++
	}

	return r
}

