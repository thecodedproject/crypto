package main

import (
	"github.com/shopspring/decimal"
	"github.com/thecodedproject/crypto/exchangesdk"
	"github.com/thecodedproject/crypto/exchangesdk/binance"
	"log"
	"time"
)

var logPeriod = 1*time.Minute

type AvVals struct {
	Sum decimal.Decimal
	Min decimal.Decimal
	Max decimal.Decimal
}

type Stats struct {

	ObCount int64

	VolumeBuyPrice AvVals
	VolumeSellPrice AvVals

	BestBid AvVals
	BestAsk AvVals
}

func updateStats(stats *Stats, ob *binance.OrderBook) {

	stats.ObCount += 1
	updateAvVals(&stats.VolumeBuyPrice, ob.VolumeBuyPrice)
	updateAvVals(&stats.VolumeSellPrice, ob.VolumeSellPrice)
	updateAvVals(&stats.BestBid, ob.Bids[0].Price)
	updateAvVals(&stats.BestAsk, ob.Asks[0].Price)
}

func updateAvVals(av *AvVals, val decimal.Decimal) {

	av.Sum = av.Sum.Add(val)

	if val.GreaterThan(av.Max) {
		av.Max = val
	}

	if av.Min.IsZero() || val.LessThan(av.Min) {
		av.Min = val
	}
}

func logStats(stats *Stats) {

	log.Printf("%s (%s)\t\t%s (%s)\t\t%s (%s)\t\t%s (%s)\n",
		average(stats.VolumeSellPrice.Sum, stats.ObCount),
		stats.VolumeSellPrice.Max.Sub(stats.VolumeSellPrice.Min),
		average(stats.BestBid.Sum, stats.ObCount),
		stats.BestBid.Max.Sub(stats.BestBid.Min),
		average(stats.BestAsk.Sum, stats.ObCount),
		stats.BestAsk.Max.Sub(stats.BestAsk.Min),
		average(stats.VolumeBuyPrice.Sum, stats.ObCount),
		stats.VolumeBuyPrice.Max.Sub(stats.VolumeBuyPrice.Min),
	)
}

func average(sum decimal.Decimal, count int64) decimal.Decimal {

	if count == 0 {
		return decimal.Decimal{}
	}

	return sum.DivRound(decimal.NewFromInt(count), 2)
}

func nextLogPeriod() time.Duration {

	oneEpoch := logPeriod
	startOfNextEpoch := time.Now().Truncate(oneEpoch).Add(oneEpoch)
	return startOfNextEpoch.Sub(time.Now())
}

func logStatsForever() {

	obf, tradeStream := binance.NewOrderBookFollowerAndTradeStream(
		exchangesdk.BTCEUR,
		decimal.NewFromFloat(1.0),
	)

	log.Printf("VolSell (var.)\t\tBestBid (var.)\t\tBestAsk (var.)\t\tVolBuy (var.)\n")

	var stats Stats

	for {
		select {
			case ob, more := <-obf:
				if more {
					updateStats(&stats, &ob)
				} else {
					log.Println("obf closed")
					return
				}
			case trade, more := <-tradeStream:
				if more {
					log.Printf("Got trade: %+v\n", trade)
				} else {
					log.Println("trade stream closed")
					return
				}
			case <-time.After(nextLogPeriod()):
				logStats(&stats)
				stats = Stats{}
		}
	}
}

func main() {

	logStatsForever()
}
