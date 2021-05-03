package binance

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"testing"
	"time"

	"github.com/shopspring/decimal"
	"github.com/thecodedproject/crypto"
	"github.com/thecodedproject/crypto/exchangesdk"
	"github.com/thecodedproject/crypto/exchangesdk/requestutil"
	utiltime "github.com/thecodedproject/crypto/util/time"
)

const (
	baseUrl = "https://api.binance.com"
)

type client struct {
	apiKey      string
	apiSecret   string
	httpClient  *http.Client
	tradingPair string
	pair        crypto.Pair
}

var _ exchangesdk.Client = (*client)(nil)

func NewClient(
	apiKey string,
	apiSecret string,
	pair crypto.Pair,
) (*client, error) {

	tradingPair, err := getBinanceTradingPair(pair)
	if err != nil {
		return nil, err
	}

	return &client{
		apiKey:      apiKey,
		apiSecret:   apiSecret,
		httpClient:  http.DefaultClient,
		tradingPair: tradingPair,
	}, nil
}

func NewClientForTesting(
	t *testing.T,
	apiKey string,
	apiSecret string,
	tradingPair string,
	handler func(req *http.Request) *http.Response,
) *client {

	return &client{
		apiKey:    apiKey,
		apiSecret: apiSecret,
		httpClient: &http.Client{
			Transport: requestutil.RoundTripFunc(handler),
		},
		tradingPair: tradingPair,
	}
}

func getBinanceTradingPair(pair crypto.Pair) (string, error) {

	switch pair {
	case crypto.PairBTCEUR:
		return "BTCEUR", nil
	case crypto.PairBTCGBP:
		return "BTCGBP", nil
	case crypto.PairBTCUSDT:
		return "BTCUSDT", nil
	case crypto.PairLTCBTC:
		return "LTCBTC", nil
	case crypto.PairETHBTC:
		return "ETHBTC", nil
	case crypto.PairBCHBTC:
		return "BCHBTC", nil
	default:
		return "", fmt.Errorf("Pair %s is not supported by exchagnesdk.Luno", pair)
	}
}

func (c *client) Exchange() crypto.Exchange {

	return crypto.Exchange{
		Provider: crypto.ApiProviderBinance,
		Pair:     c.pair,
	}
}

func (c *client) LatestPrice(ctx context.Context) (decimal.Decimal, error) {

	path := requestutil.FullPath(baseUrl, "/api/v3/ticker/price")
	values := url.Values{}
	values.Add("symbol", c.tradingPair)
	path.RawQuery = values.Encode()

	body, err := GetBody(c.httpClient.Get(path.String()))
	if err != nil {
		return decimal.Decimal{}, err
	}

	latestPrice := struct {
		Price decimal.Decimal `json:"price"`
	}{}

	err = json.Unmarshal(body, &latestPrice)
	if err != nil {
		return decimal.Decimal{}, err
	}

	return latestPrice.Price, nil
}

func (c *client) PostLimitOrder(
	ctx context.Context,
	order exchangesdk.Order,
) (string, error) {

	side := "BUY"
	if order.Type == exchangesdk.OrderTypeAsk {
		side = "SELL"
	}

	values := url.Values{}
	values.Add("type", "LIMIT")
	values.Add("side", side)
	values.Add("timeInForce", "GTC")
	values.Add("quantity", order.Volume.String())
	values.Add("price", order.Price.String())

	body, err := requestToOrderEndpointWithAuth(
		"POST",
		c.httpClient,
		c.apiKey,
		c.apiSecret,
		c.tradingPair,
		values,
	)
	if err != nil {
		return "", err
	}

	res := struct {
		Id string `json:"clientOrderId"`
	}{}

	err = json.Unmarshal(body, &res)
	if err != nil {
		return "", err
	}

	return res.Id, nil
}

func (c *client) PostStopLimitOrder(
	ctx context.Context,
	order exchangesdk.StopLimitOrder,
) (string, error) {

	side, err := sideFromOrderBookSide(order.Side)

	values := url.Values{}
	values.Add("type", "STOP_LOSS_LIMIT")
	values.Add("side", side)
	values.Add("timeInForce", "GTC")
	values.Add("quantity", order.Volume.String())
	values.Add("price", order.LimitPrice.String())
	values.Add("stopPrice", order.StopPrice.String())

	body, err := requestToOrderEndpointWithAuth(
		"POST",
		c.httpClient,
		c.apiKey,
		c.apiSecret,
		c.tradingPair,
		values,
	)
	if err != nil {
		return "", err
	}

	res := struct {
		Id string `json:"clientOrderId"`
	}{}

	err = json.Unmarshal(body, &res)
	if err != nil {
		return "", err
	}

	return res.Id, nil
}

func (c *client) CancelOrder(ctx context.Context, orderId string) error {

	values := url.Values{}
	values.Add("origClientOrderId", orderId)

	_, err := requestToOrderEndpointWithAuth(
		"DELETE",
		c.httpClient,
		c.apiKey,
		c.apiSecret,
		c.tradingPair,
		values,
	)
	return err
}

func (c *client) GetOrderStatus(
	ctx context.Context,
	orderId string,
) (exchangesdk.OrderStatus, error) {

	values := url.Values{}
	values.Add("origClientOrderId", orderId)

	body, err := requestToOrderEndpointWithAuth(
		"GET",
		c.httpClient,
		c.apiKey,
		c.apiSecret,
		c.tradingPair,
		values,
	)
	if err != nil {
		return exchangesdk.OrderStatus{}, err
	}

	res := struct {
		Status              string          `json:"status"`
		Side                string          `json:"side"`
		ExecutedQty         decimal.Decimal `json:"executedQty"`
		CummulativeQuoteQty decimal.Decimal `json:"cummulativeQuoteQty"`
		IsWorking           bool            `json:"isWorking"`
	}{}

	err = json.Unmarshal(body, &res)
	if err != nil {
		return exchangesdk.OrderStatus{}, err
	}

	state := exchangesdk.OrderStateUnknown
	if res.Status == "NEW" {
		if res.IsWorking {
			state = exchangesdk.OrderStateInOrderBook
		} else {
			state = exchangesdk.OrderStateAwaitingTrigger
		}
	} else if res.Status == "PARTIALLY_FILLED" {
		state = exchangesdk.OrderStateInOrderBook
	} else if res.Status == "FILLED" {
		state = exchangesdk.OrderStateFilled
	}

	orderType := exchangesdk.OrderTypeBid
	if res.Side == "SELL" {
		orderType = exchangesdk.OrderTypeAsk
	}

	return exchangesdk.OrderStatus{
		State:             state,
		Type:              orderType,
		FillAmountBase:    res.ExecutedQty,
		FillAmountCounter: res.CummulativeQuoteQty,
	}, nil
}

func (c *client) GetTrades(ctx context.Context, page int64) ([]exchangesdk.Trade, error) {

	panic("not implemented")
}

func (c *client) MakerFee() decimal.Decimal {

	return decimal.New(75, -5)
}

func (c *client) TakerFee() decimal.Decimal {

	return decimal.New(75, -5)
}

func (l *client) CounterPrecision() int32 {

	return 2
}

func (l *client) BasePrecision() int32 {

	return 6
}

func requestToOrderEndpointWithAuth(
	reqMethod string,
	httpClient *http.Client,
	apiKey string,
	apiSecret string,
	pair string,
	values url.Values,
) ([]byte, error) {

	path := requestutil.FullPath(baseUrl, "/api/v3/order")

	nowMs := utiltime.Now().Round(time.Millisecond).UnixNano() / 1e6
	timestampStr := strconv.FormatInt(nowMs, 10)
	values.Add("timestamp", timestampStr)
	values.Add("symbol", pair)

	path.RawQuery = values.Encode()

	return requestWithHmacAuth(
		reqMethod,
		httpClient,
		apiKey,
		apiSecret,
		path,
	)
}

func requestWithHmacAuth(
	reqMethod string,
	c *http.Client,
	key string,
	secret string,
	fullUrl *url.URL,
) ([]byte, error) {

	msgToSign := fullUrl.RawQuery

	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(msgToSign))
	signature := hex.EncodeToString(
		mac.Sum(nil),
	)

	query := fullUrl.Query()
	query.Add("signature", signature)
	fullUrl.RawQuery = query.Encode()

	req, err := http.NewRequest(reqMethod, fullUrl.String(), nil)
	if err != nil {
		return nil, err
	}

	req.Header.Add("X-MBX-APIKEY", key)

	return GetBody(c.Do(req))
}

func GetBody(res *http.Response, err error) ([]byte, error) {

	if err != nil {
		return nil, err
	}

	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	if res.StatusCode != http.StatusOK {

		errStruct := struct {
			ErrCode *int64 `json:"code"`
			ErrMsg  string `json:"msg"`
		}{}

		err := json.Unmarshal(body, &errStruct)
		if err != nil {
			return nil, requestutil.HttpStatusError(
				res,
				"Error decoding errMsg:",
				err,
			)
		}
		if errStruct.ErrCode != nil {
			return nil, requestutil.HttpStatusError(
				res,
				*errStruct.ErrCode,
				errStruct.ErrMsg,
			)
		}
		return nil, requestutil.HttpStatusError(res)
	}

	return body, nil
}

func sideFromOrderBookSide(
	side exchangesdk.OrderBookSide,
) (string, error) {

	switch side {
	case exchangesdk.OrderBookSideBid:
		return "BUY", nil
	case exchangesdk.OrderBookSideAsk:
		return "SELL", nil
	default:
		return "", fmt.Errorf(
			"dont know how to convert OrderBookSide `%s` to binance order side",
			side,
		)
	}

}
