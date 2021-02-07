package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/thecodedproject/crypto"
	"github.com/thecodedproject/crypto/exchangesdk"
	"github.com/thecodedproject/crypto/exchangesdk/luno"
	"github.com/thecodedproject/crypto/io"
	"github.com/thecodedproject/crypto/profitloss"
	"log"
)

var (
	authPath = flag.String("auth", "api_auth.json", "Path to auth config json file")
)

const (
	MAX_TRADE_PAGES = 100
)

func getAllTrades(ctx context.Context, c exchangesdk.Client) []exchangesdk.Trade {

	trades := make([]exchangesdk.Trade, 0)
	for page:=int64(1); page<=MAX_TRADE_PAGES; page++ {

		tradesForPage, err := c.GetTrades(ctx, page)
		if err != nil {
			log.Fatal(err)
		}

		trades = append(trades, tradesForPage...)

		if len(tradesForPage) < 100 {
			break
		}

		if page == MAX_TRADE_PAGES {
			log.Fatal("Max pages of trades reached!!")
		}
	}

	return trades
}

func main() {

	flag.Parse()

	auth, err := io.GetAuthConfigByName(*authPath, "luno_api_key")
	if err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()
	c, err := luno.NewClient(
		auth.Key,
		auth.Secret,
		crypto.PairBTCEUR,
	)
	if err != nil {
		log.Fatal(err)
	}

	trades := getAllTrades(ctx, c)

	var report profitloss.Report
	report = profitloss.Add(report, trades...)

	marketPrice, err := c.LatestPrice(ctx)
	if err != nil {
		log.Fatal(err)
	}

	snapshot := profitloss.GenerateSnapshot(report, marketPrice)

	reportJson, err := json.Marshal(&snapshot)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(string(reportJson))
}
