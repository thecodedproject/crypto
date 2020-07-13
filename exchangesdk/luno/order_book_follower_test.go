package luno_test

import (
	"fmt"
	"log"
	"github.com/stretchr/testify/require"
	"github.com/thecodedproject/crypto/exchangesdk"
	"github.com/thecodedproject/crypto/exchangesdk/luno"
	"testing"
	"time"
)

func TestOrderBookFollower(t *testing.T) {

	t.Skip("do not call luno api in tests")

	ob, err := luno.NewOrderBookFollower(
		"api_key",
		"api_secrect",
		exchangesdk.BTCEUR,
	)
	require.NoError(t, err)

	go func() {
		for {
			select {
				case <-time.After(5*time.Second):
					log.Println("Max bid: ", ob.MaxBidPrice(), "Min Ask: ", ob.MinAskPrice())
			}
		}
	}()

	select {
		case <-time.After(60*time.Second):
			// ...
	}

	fmt.Println(ob.MidpointPrice())

}
