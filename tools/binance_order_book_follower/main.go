package main

import (
	"github.com/shopspring/decimal"
	"github.com/thecodedproject/crypto/exchangesdk"
	"github.com/thecodedproject/crypto/exchangesdk/binance"
	"github.com/thecodedproject/crypto/util"
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

	BuySellWeight util.MovingStats
}

func updateStatsWithOrderBook(stats *Stats, ob *binance.OrderBook) {

	stats.ObCount += 1
	updateAvVals(&stats.VolumeBuyPrice, ob.VolumeBuyPrice)
	updateAvVals(&stats.VolumeSellPrice, ob.VolumeSellPrice)
	updateAvVals(&stats.BestBid, ob.Bids[0].Price)
	updateAvVals(&stats.BestAsk, ob.Asks[0].Price)
}

func updateStatesWithTrade(stats *Stats, trade *binance.Trade) {

	weight, _ := trade.Volume.Float64()
	if trade.MakerSide == binance.MarketSideBuy {
		weight = -weight
	}

	if stats.BuySellWeight == nil {
		log.Fatal("updateStatesWithTrade nil")
	}

	stats.BuySellWeight.Add(trade.Timestamp, weight)
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

func minutesAgo(i int) time.Time{

	return time.Now().Add(time.Duration(-i)*time.Minute)
}

func logStats(stats *Stats) {

	if stats.BuySellWeight == nil {
		log.Fatal("logStats nil")
	}

	bsWeight1min, err := stats.BuySellWeight.Sum(minutesAgo(1))
	if err != nil {
		log.Fatal(err)
	}

	bsWeight5min, err := stats.BuySellWeight.Sum(minutesAgo(5))
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("%s (%s)\t\t%s (%s)\t\t%s (%s)\t\t%s (%s)\t\t%.3f\t\t%.3f\n",
		average(stats.VolumeSellPrice.Sum, stats.ObCount),
		stats.VolumeSellPrice.Max.Sub(stats.VolumeSellPrice.Min),
		average(stats.BestBid.Sum, stats.ObCount),
		stats.BestBid.Max.Sub(stats.BestBid.Min),
		average(stats.BestAsk.Sum, stats.ObCount),
		stats.BestAsk.Max.Sub(stats.BestAsk.Min),
		average(stats.VolumeBuyPrice.Sum, stats.ObCount),
		stats.VolumeBuyPrice.Max.Sub(stats.VolumeBuyPrice.Min),
		bsWeight1min,
		bsWeight5min,
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

	log.Printf("VolSell (var.)\t\tBestBid (var.)\t\tBestAsk (var.)\t\tVolBuy (var.)\t\tBSWeight(1m)\t\tBSWeight(5m)\n")

	var stats Stats
	stats.BuySellWeight = util.NewMovingStats(6*time.Minute)

	for {
		select {
			case ob, more := <-obf:
				if more {
					updateStatsWithOrderBook(&stats, &ob)
				} else {
					log.Println("obf closed")
					return
				}
			case trade, more := <-tradeStream:
				if more {
					updateStatesWithTrade(&stats, &trade)
				} else {
					log.Println("trade stream closed")
					return
				}
			case <-time.After(nextLogPeriod()):
				logStats(&stats)
				stats.ObCount = 0
				stats.VolumeBuyPrice = AvVals{}
				stats.VolumeSellPrice = AvVals{}
				stats.BestBid = AvVals{}
				stats.BestAsk = AvVals{}
		}
	}
}

func main() {

	logStatsForever()
}
