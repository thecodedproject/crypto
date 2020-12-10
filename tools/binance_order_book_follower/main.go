package main

import (
	"github.com/thecodedproject/crypto/exchangesdk"
	"github.com/thecodedproject/crypto/exchangesdk/binance"
	"github.com/thecodedproject/crypto/market_follower"
	market_follower_stats"github.com/thecodedproject/crypto/market_follower/stats"
	"github.com/thecodedproject/crypto/util"
	"log"
	"time"
)

var logPeriod = 10*time.Second
var volumePrice = 1.0

type Stats struct {

	BestBid util.MovingStats
	BestAsk util.MovingStats

	VolumeBuyPrice util.MovingStats
	VolumeSellPrice util.MovingStats

	BuySellWeight util.MovingStats
}

func updateStatsWithOrderBook(stats *Stats, ob *market_follower.OrderBook) {

	stats.BestBid.Add(ob.Timestamp, ob.Bids[0].Price)
	stats.BestAsk.Add(ob.Timestamp, ob.Asks[0].Price)

	buyPrice, sellPrice, err := market_follower_stats.CalcPricePerVolumeStats(ob, volumePrice)
	if err != nil {
		log.Fatal(err)
	}

	stats.VolumeBuyPrice.Add(ob.Timestamp, buyPrice)
	stats.VolumeSellPrice.Add(ob.Timestamp, sellPrice)
}

func updateStatsWithTrade(stats *Stats, trade *market_follower.Trade) {

	weight := trade.Volume
	if trade.MakerSide == market_follower.MarketSideBuy {
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

	statsTime := minutesAgo(1)

	bsWeight1min, err := stats.BuySellWeight.Sum(statsTime)
	if err != nil {
		log.Fatal(err)
	}

	bsWeight5min, err := stats.BuySellWeight.Sum(minutesAgo(5))
	if err != nil {
		log.Fatal(err)
	}

	bestBid, err := stats.BestBid.Mean(statsTime)
	if err != nil {
		log.Fatal(err)
	}
	bestBidGrad, err := stats.BestBid.Gradient(statsTime)
	if err != nil {
		log.Fatal(err)
	}
	bestAsk, err := stats.BestAsk.Mean(statsTime)
	if err != nil {
		log.Fatal(err)
	}
	bestAskGrad, err := stats.BestAsk.Gradient(statsTime)
	if err != nil {
		log.Fatal(err)
	}
	avSellPrice, err := stats.VolumeSellPrice.Mean(statsTime)
	if err != nil {
		log.Fatal(err)
	}
	avSellGrad, err := stats.VolumeSellPrice.Gradient(statsTime)
	if err != nil {
		log.Fatal(err)
	}
	avBuyPrice, err := stats.VolumeBuyPrice.Mean(statsTime)
	if err != nil {
		log.Fatal(err)
	}
	avBuyGrad, err := stats.VolumeBuyPrice.Gradient(statsTime)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("%.2f (%.2f)\t\t%.2f (%.2f)\t\t%.2f (%.2f)\t\t%.2f (%.2f)\t\t%.2f\t\t%.2f\n",
		avSellPrice,
		avSellGrad,
		bestBid,
		bestBidGrad,
		bestAsk,
		bestAskGrad,
		avBuyPrice,
		avBuyGrad,
		bsWeight1min,
		bsWeight5min,
	)
}

func nextLogPeriod() time.Duration {

	oneEpoch := logPeriod
	startOfNextEpoch := time.Now().Truncate(oneEpoch).Add(oneEpoch)
	return startOfNextEpoch.Sub(time.Now())
}

func logStatsForever() {

	obf, tradeStream := binance.NewOrderBookFollowerAndTradeStream(
		exchangesdk.BTCEUR,
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
					updateStatsWithTrade(&stats, &trade)
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
