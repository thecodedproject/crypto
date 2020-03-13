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
		"ke8fjx2pxn56z",
		"BlbkO7-W5Tanzi-NMJaFxkY-0WQITpqy4x8h4huHaLY")
	if err != nil {
		log.Fatal(err)
	}

	log.Println("success")
}
