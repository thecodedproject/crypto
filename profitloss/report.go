package profitloss

import (
	"github.com/shopspring/decimal"
	"github.com/thecodedproject/crypto/exchangesdk"
)

type CalcType int

const (
	CalcTypeAverage  = 0
	CalcTypeUnknown  = 1
	calcTypeSentinal = 2
)

type Report struct {
	Type                 CalcType        `json:"type"`
	InitalBaseBalance    decimal.Decimal `json:"initial_base_balance"`
	BaseBought           decimal.Decimal `json:"base_bought"`
	BaseSold             decimal.Decimal `json:"base_sold"`
	BaseFees             decimal.Decimal `json:"base_fees"`
	InitalCounterBalance decimal.Decimal `json:"initial_counter_balance"`
	CounterBought        decimal.Decimal `json:"counter_bought"`
	CounterSold          decimal.Decimal `json:"counter_sold"`
	CounterFees          decimal.Decimal `json:"counter_fees"`
	TradeCount           int64           `json:"trade_count"`
}

type Snapshot struct {
	Report
	RealisedGain     decimal.Decimal `json:"realised_gain"`
	UnrealisedGain   decimal.Decimal `json:"unrealised_gain"`
	AverageBuyPrice  decimal.Decimal `json:"averagebuy_price"`
	AverageSellPrice decimal.Decimal `json:"averagesell_price"`
	BaseBalance      decimal.Decimal `json:"base_balance"`
	CounterBalance   decimal.Decimal `json:"counter_balance"`
	TotalVolume      decimal.Decimal `json:"total_volume"`
	TotalGain        decimal.Decimal `json:"total_gain"`
}

func GenerateSnapshot(r Report, marketPrice decimal.Decimal) Snapshot {
	return Snapshot{
		Report:           r,
		RealisedGain:     r.RealisedGain(),
		UnrealisedGain:   r.UnrealisedGain(marketPrice),
		AverageBuyPrice:  r.AverageBuyPrice(),
		AverageSellPrice: r.AverageSellPrice(),
		BaseBalance:      r.BaseBalance(),
		CounterBalance:   r.CounterBalance(),
		TotalVolume:      r.TotalVolume(),
		TotalGain:        r.TotalGain(marketPrice),
	}
}

func (r Report) RealisedGain() decimal.Decimal {
	volumeForRealisedGain := decimal.Min(r.BaseBought, r.BaseSold)
	return r.AverageSellPrice().Sub(r.AverageBuyPrice()).Mul(volumeForRealisedGain).Sub(r.CounterFees)
}

func (r Report) UnrealisedGain(marketPrice decimal.Decimal) decimal.Decimal {
	return marketPrice.Sub(r.AverageBuyPrice()).Mul(r.BaseBalance())
}

func (r Report) AverageBuyPrice() decimal.Decimal {
	if r.BaseBought.IsZero() {
		return decimal.Decimal{}
	}
	return r.CounterSold.Div(r.BaseBought)
}

func (r Report) AverageSellPrice() decimal.Decimal {
	if r.BaseSold.IsZero() {
		return decimal.Decimal{}
	}
	return r.CounterBought.Div(r.BaseSold)
}

func (r Report) BaseBalance() decimal.Decimal {
	return r.InitalBaseBalance.Add(r.BaseBought).Sub(r.BaseSold).Sub(r.BaseFees)
}

func (r Report) CounterBalance() decimal.Decimal {
	return r.InitalCounterBalance.Add(r.CounterBought).Sub(r.CounterSold).Sub(r.CounterFees)
}

func (r Report) TotalVolume() decimal.Decimal {
	return r.BaseSold.Add(r.BaseBought)
}

func (r Report) TotalGain(marketPrice decimal.Decimal) decimal.Decimal {
	return r.RealisedGain().Add(r.UnrealisedGain(marketPrice))
}

func Add(r Report, trades ...exchangesdk.Trade) Report {

	for _, o := range trades {
		orderCost := o.Volume.Mul(o.Price)

		if o.Type == exchangesdk.OrderTypeBid {
			r.BaseBought = r.BaseBought.Add(o.Volume)
			r.CounterSold = r.CounterSold.Add(orderCost)
		} else {
			r.CounterBought = r.CounterBought.Add(orderCost)
			r.BaseSold = r.BaseSold.Add(o.Volume)
		}

		r.CounterFees = r.CounterFees.Add(o.CounterFee)
		r.BaseFees = r.BaseFees.Add(o.BaseFee)
		r.TradeCount++
	}

	return r
}
