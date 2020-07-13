package binance

import (
	"errors"
	"github.com/shopspring/decimal"
)

const (
	pricePrecision = int32(2)

	ErrVolumePriceNotEnoughOrders = "Not enough order to calc VolumePrice"
)

func calcStats(ob *OrderBook) error {

	var err error
	ob.VolumeBuyPrice, err = VolumePrice(
		&ob.Asks,
		ob.volumePrice,
		pricePrecision,
	)
	if err != nil {
		return err
	}

	ob.VolumeSellPrice, err = VolumePrice(
		&ob.Bids,
		ob.volumePrice,
		pricePrecision,
	)
	if err != nil {
		return err
	}

	return nil
}

func VolumePrice(
	orders *[]Order,
	volume decimal.Decimal,
	precision int32,
) (decimal.Decimal, error) {

	var volumeSum decimal.Decimal
	var weightedPriceSum decimal.Decimal
	for i := range *orders {

		volumeRemaining := volume.Sub(volumeSum)

		if (*orders)[i].Volume.LessThan(volumeRemaining) {
			weightedPriceSum = weightedPriceSum.Add(
				(*orders)[i].Price.Mul((*orders)[i].Volume),
			)
			volumeSum = volumeSum.Add((*orders)[i].Volume)

			if i == len(*orders)-1 {
				return decimal.Decimal{}, errors.New(ErrVolumePriceNotEnoughOrders)
			}

			continue
		}

		weightedPriceSum = weightedPriceSum.Add(
			(*orders)[i].Price.Mul(volumeRemaining),
		)
		break
	}

	volumePrice := weightedPriceSum.DivRound(
		volume,
		precision,
	)

	return volumePrice, nil
}
