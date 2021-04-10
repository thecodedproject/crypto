package market_stats

import (
	"errors"

	"github.com/thecodedproject/crypto/exchangesdk"
)

const (
	ErrVolumePriceNotEnoughOrders = "Not enough order to calc VolumePrice"
)

func CalcPricePerVolumeStats(ob *exchangesdk.OrderBook, volume float64) (float64, float64, error) {

	var err error
	volumeBuyPrice, err := VolumePrice(
		&ob.Asks,
		volume,
	)
	if err != nil {
		return 0, 0, err
	}

	volumeSellPrice, err := VolumePrice(
		&ob.Bids,
		volume,
	)
	if err != nil {
		return 0, 0, err
	}

	return volumeBuyPrice, volumeSellPrice, nil
}

func VolumePrice(
	orders *[]exchangesdk.OrderBookOrder,
	volume float64,
) (float64, error) {

	var volumeSum float64
	var weightedPriceSum float64
	for i := range *orders {

		volumeRemaining := volume - volumeSum

		if (*orders)[i].Volume < volumeRemaining {
			weightedPriceSum = weightedPriceSum + ((*orders)[i].Price * (*orders)[i].Volume)
			volumeSum = volumeSum + (*orders)[i].Volume

			if i == len(*orders)-1 {
				return 0.0, errors.New(ErrVolumePriceNotEnoughOrders)
			}

			continue
		}

		weightedPriceSum = weightedPriceSum + ((*orders)[i].Price * volumeRemaining)
		break
	}

	volumePrice := weightedPriceSum / volume

	return volumePrice, nil
}
