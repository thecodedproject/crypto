package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/thecodedproject/crypto"
	"github.com/thecodedproject/crypto/exchangesdk"
	"github.com/thecodedproject/crypto/exchangesdk/factory"
	"github.com/thecodedproject/crypto/exchangesdk/market_stats"
	"github.com/thecodedproject/crypto/io"
	"github.com/thecodedproject/crypto/util"
)

var logPeriod = 10 * time.Second
var volumePrice = 1.0

type Stats struct {
	BestBid util.MovingStats
	BestAsk util.MovingStats

	VolumeBuyPrice  util.MovingStats
	VolumeSellPrice util.MovingStats

	BuySellWeight util.MovingStats
}

func updateStatsWithOrderBook(stats *Stats, ob *exchangesdk.OrderBook) {

	stats.BestBid.Add(ob.Timestamp, ob.Bids[0].Price)
	stats.BestAsk.Add(ob.Timestamp, ob.Asks[0].Price)

	buyPrice, sellPrice, err := market_stats.CalcPricePerVolumeStats(ob, volumePrice)
	if err != nil {
		log.Fatal(err)
	}

	stats.VolumeBuyPrice.Add(ob.Timestamp, buyPrice)
	stats.VolumeSellPrice.Add(ob.Timestamp, sellPrice)
}

func updateStatsWithTrade(stats *Stats, trade *exchangesdk.OrderBookTrade) {

	weight := trade.Volume
	if trade.MakerSide == exchangesdk.OrderBookSideBid {
		weight = -weight
	}

	if stats.BuySellWeight == nil {
		log.Fatal("updateStatesWithTrade nil")
	}

	stats.BuySellWeight.Add(trade.Timestamp, weight)
}

func minutesAgo(i int) time.Time {

	return time.Now().Add(time.Duration(-i) * time.Minute)
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

	bestBid := stats.BestBid.Latest() //Mean(statsTime)
	bestBidGrad, err := stats.BestBid.Gradient(statsTime)
	if err != nil {
		log.Fatal(err)
	}
	bestAsk := stats.BestAsk.Latest() //Mean(statsTime)
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

func logStatsForever(
	ctx context.Context,
	wg *sync.WaitGroup,
	apiAuth crypto.AuthConfig,
) {

	wg.Add(1)
	obf, tradeStream, err := factory.NewMarketFollower(
		ctx,
		wg,
		crypto.Exchange{
			Provider: apiAuth.Provider,
			Pair:     crypto.PairBTCEUR,
		},
		apiAuth,
	)
	if err != nil {
		log.Fatal("failed to create market follower:", err)
	}

	log.Printf("VolSell (var.)\t\tBestBid (var.)\t\tBestAsk (var.)\t\tVolBuy (var.)\t\tBSWeight(1m)\t\tBSWeight(5m)\n")

	var stats Stats
	stats.BuySellWeight = util.NewMovingStats(6 * time.Minute)
	stats.VolumeBuyPrice = util.NewMovingStats(6 * time.Minute)
	stats.VolumeSellPrice = util.NewMovingStats(6 * time.Minute)
	stats.BestBid = util.NewMovingStats(6 * time.Minute)
	stats.BestAsk = util.NewMovingStats(6 * time.Minute)

	for {
		select {
		case ob, more := <-obf:
			if more {
				updateStatsWithOrderBook(&stats, &ob)
			} else {
				log.Println("obf closed")
				wg.Done()
				return
			}
		case trade, more := <-tradeStream:
			if more {
				updateStatsWithTrade(&stats, &trade)
			} else {
				log.Println("trade stream closed")
				wg.Done()
				return
			}
		case <-time.After(nextLogPeriod()):
			logStats(&stats)
		case <-ctx.Done():
			wg.Done()
			return
		}
	}
}

func main() {

	flag.Parse()
	if flag.NArg() != 1 {
		log.Fatal("Usage: market_follower [luno|binance|dummy]")
	}

	var apiCreds crypto.AuthConfig
	switch flag.Arg(0) {
	case "luno":
		var err error
		apiCreds, err = io.GetAuthConfigByName(
			"api_auth.json",
			"luno_jcooper_readonly_test_20201210",
		)
		if err != nil {
			log.Fatal(err)
		}
	case "binance":
		// Binance doesnt require api creds
		apiCreds = crypto.AuthConfig{
			Provider: crypto.ApiProviderBinance,
		}
	case "dummy":
		// Dummy exchange doesnt require api creds
		apiCreds = crypto.AuthConfig{
			Provider: crypto.ApiProviderDummyExchange,
		}
	default:
		log.Fatal("Unknown exchange:", flag.Arg(0))
	}

	ctx, cancel := context.WithCancel(context.Background())
	var wg sync.WaitGroup

	wg.Add(1)
	go logStatsForever(ctx, &wg, apiCreds)

	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGTERM, syscall.SIGINT)

	select {
	case sig := <-ch:
		log.Println("Received OS signal:", sig.String())
		cancel()
		wg.Wait()
		return
	}
}
