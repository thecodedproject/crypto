package crypto_test

import (
	"testing"
	"encoding/json"
	"github.com/thecodedproject/crypto"
	"github.com/stretchr/testify/require"
)

func TestMarshalMapWithExchangeAsKey(t *testing.T) {

	m := map[crypto.Exchange]int{
		{
			Provider: crypto.ApiProviderBinance,
			Pair: crypto.PairBTCEUR,
		}: 0,
	}

	_, err := json.Marshal(m)
	require.NoError(t, err)
}

