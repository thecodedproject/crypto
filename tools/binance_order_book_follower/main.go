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

type Stats struct {

	BestBid util.MovingStats
	BestAsk util.MovingStats

	VolumeBuyPrice util.MovingStats
	VolumeSellPrice util.MovingStats

	BuySellWeight util.MovingStats
}

func updateStatsWithOrderBook(stats *Stats, ob *binance.OrderBook) {

	bestBid, _ := ob.Bids[0].Price.Float64()
	bestAsk, _ := ob.Asks[0].Price.Float64()

	stats.BestBid.Add(ob.Timestamp, bestBid)
	stats.BestAsk.Add(ob.Timestamp, bestAsk)

	buyPrice, _ := ob.VolumeBuyPrice.Float64()
	sellPrice, _ := ob.VolumeSellPrice.Float64()

	stats.VolumeBuyPrice.Add(ob.Timestamp, buyPrice)
	stats.VolumeSellPrice.Add(ob.Timestamp, sellPrice)
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

func minutesAgo(i int) time.Time{

	return time.Now().Add(time.Duration(-i)*time.Minute)
}

func logStats(stats *Stats) {

	minAgo := minutesAgo(1)

	bsWeight1min, err := stats.BuySellWeight.Sum(minAgo)
	if err != nil {
		log.Fatal(err)
	}

	bsWeight5min, err := stats.BuySellWeight.Sum(minutesAgo(5))
	if err != nil {
		log.Fatal(err)
	}

	bestBid, err := stats.BestBid.Mean(minAgo)
	if err != nil {
		log.Fatal(err)
	}
	bestBidVar, err := stats.BestBid.Variation(minAgo)
	if err != nil {
		log.Fatal(err)
	}
	bestAsk, err := stats.BestAsk.Mean(minAgo)
	if err != nil {
		log.Fatal(err)
	}
	bestAskVar, err := stats.BestAsk.Variation(minAgo)
	if err != nil {
		log.Fatal(err)
	}
	avSellPrice, err := stats.VolumeSellPrice.Mean(minAgo)
	if err != nil {
		log.Fatal(err)
	}
	avSellVariation, err := stats.VolumeSellPrice.Variation(minAgo)
	if err != nil {
		log.Fatal(err)
	}
	avBuyPrice, err := stats.VolumeBuyPrice.Mean(minAgo)
	if err != nil {
		log.Fatal(err)
	}
	avBuyVariation, err := stats.VolumeBuyPrice.Variation(minAgo)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("%.3f (%.3f)\t\t%.3f (%.3f)\t\t%.3f (%.3f)\t\t%.3f (%.3f)\t\t%.3f\t\t%.3f\n",
		avSellPrice,
		avSellVariation,
		bestBid,
		bestBidVar,
		bestAsk,
		bestAskVar,
		avBuyPrice,
		avBuyVariation,
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
	stats.VolumeBuyPrice = util.NewMovingStats(6*time.Minute)
	stats.VolumeSellPrice = util.NewMovingStats(6*time.Minute)
	stats.BestBid = util.NewMovingStats(6*time.Minute)
	stats.BestAsk = util.NewMovingStats(6*time.Minute)

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
		}
	}
}

func main() {

	logStatsForever()
}
