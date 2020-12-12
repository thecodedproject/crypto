package crypto

//go:generate enumer -type=Exchange -trimprefix=Exchange -json

type Exchange int

const (
	ExchangeUnknown Exchange = 0
	ExchangeLuno Exchange = 1
	ExchangeBinance Exchange = 2
	ExchangeSentinal Exchange = 3
)

type AuthConfig struct {
	ApiExchange Exchange `json:"exchange"`
	ApiKey string `json:"api_key"`
	ApiSecret string `json:"api_secret"`
}
