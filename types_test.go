package crypto_test

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/thecodedproject/crypto"
)

func TestMarshalMapWithExchangeAsKey(t *testing.T) {

	m := map[crypto.Exchange]int{
		{
			Provider: crypto.ApiProviderBinance,
			Pair:     crypto.PairBTCEUR,
		}: 0,
	}

	_, err := json.Marshal(m)
	require.NoError(t, err)
}

func TestUnmarshalExchangeFromJson(t *testing.T) {

	jsonString := `
		{
			"provider": "binance",
			"pair": "ltcbtc"
		}
	`

	var e crypto.Exchange
	err := json.Unmarshal([]byte(jsonString), &e)
	require.NoError(t, err)

	require.Equal(t, crypto.ApiProviderBinance, e.Provider)
	require.Equal(t, crypto.PairLTCBTC, e.Pair)
}
