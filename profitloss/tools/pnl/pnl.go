package main

import (
	"context"
	"github.com/thecodedproject/crypto/exchangesdk/luno"
	"log"
)

func main() {

	ctx := context.Background()

	c, err := luno.NewClient(
		ctx,
		"c2z6nzka7g2tp",
		"IuDNIAoezXut3iJ8mjvwWp9GB_4nj3TdiBtK1hgGN4Q")
	if err != nil {
		log.Fatal(err)
	}

	trades, err := c.GetTrades(ctx, 0)
	if err != nil {
		log.Fatal(err)
	}

	for _, t := range trades {
		log.Printf("%+v", t)
	}

	log.Printf("Num trades:", len(trades))


	log.Println("success")
}
