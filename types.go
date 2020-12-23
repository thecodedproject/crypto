package crypto

//go:generate enumer -type=ApiProvider -trimprefix=ApiProvider -json -text -transform=snake
//go:generate enumer -type=Pair -trimprefix=Pair -json -text -transform=snake

//ApiProvider represents the company that provides an API (e.g. Luno or Binance)
type ApiProvider int

const (
	ApiProviderUnknown ApiProvider = 0
	ApiProviderDummyExchange ApiProvider = 1
	ApiProviderLuno ApiProvider = 2
	ApiProviderBinance ApiProvider = 3
	ApiProviderSentinal ApiProvider = 4
)

type AuthConfig struct {
	Provider ApiProvider `json:"provider"`
	Key string `json:"key"`
	Secret string `json:"secret"`
}

type Pair int

const (
	PairUnknown Pair = 0
	PairBTCEUR Pair = 1
	PairBTCUSDT Pair = 2

	PairLTCBTC Pair = 3
	PairSentinal Pair = 4
)

type Exchange struct {
	Provider ApiProvider
	Pair Pair
}
