package crypto

//go:generate enumer -type=Exchange -trimprefix=Exchange -json

type Exchange int

const (
	ExchangeUnknown Exchange = 0
	ExchangeDummyExchange Exchange = 1
	ExchangeLuno Exchange = 2
	ExchangeBinance Exchange = 3
	ExchangeSentinal Exchange = 4
)

type AuthConfig struct {
	ApiExchange Exchange `json:"exchange"`
	ApiKey string `json:"api_key"`
	ApiSecret string `json:"api_secret"`
}
