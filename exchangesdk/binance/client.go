package binance

import (
	"context"
	"encoding/json"
	"github.com/shopspring/decimal"
	"github.com/thecodedproject/crypto/exchangesdk"
	"github.com/thecodedproject/crypto/exchangesdk/requestutil"
	utiltime "github.com/thecodedproject/crypto/util/time"
	"io/ioutil"
	"net/http"
	"net/url"
	"testing"
	"strconv"
	"time"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
)

type marketPair string

const (
	BTCEUR marketPair = "BTCEUR"
)

const (
	baseUrl = "https://api.binance.com"
)

type client struct {

	apiKey string
	apiSecret string
	httpClient *http.Client
}

func NewClient(apiKey, apiSecret string) (*client, error) {

	return &client{
		apiKey: apiKey,
		apiSecret: apiSecret,
		httpClient: http.DefaultClient,
	}, nil
}


func NewClientForTesting(
	t *testing.T,
	apiKey string,
	apiSecret string,
	handler func(req *http.Request) *http.Response,
) *client {

	return &client{
		apiKey: apiKey,
		apiSecret: apiSecret,
		httpClient: &http.Client{
			Transport: requestutil.RoundTripFunc(handler),
		},
	}
}

func (c *client) LatestPrice(ctx context.Context) (decimal.Decimal, error) {

	return c.LatestPriceForPair(ctx, BTCEUR)
}

func (c *client) LatestPriceForPair(ctx context.Context, pair marketPair) (decimal.Decimal, error) {

	path := requestutil.FullPath(baseUrl, "/api/v3/ticker/price")
	values := url.Values{}
	values.Add("symbol", string(pair))
	path.RawQuery = values.Encode()

	body, err := GetBody(c.httpClient.Get(path.String()))
	if err != nil {
		return decimal.Decimal{}, err
	}

	latestPrice := struct{
		Price decimal.Decimal `json:"price"`
	}{}

	err = json.Unmarshal(body, &latestPrice)
	if err != nil {
		return decimal.Decimal{}, err
	}

	return latestPrice.Price, nil
}

func (c *client) PostLimitOrder(ctx context.Context, order exchangesdk.Order) (string, error) {

	return c.PostLimitOrderForPair(ctx, BTCEUR, order)
}

func (c *client) PostLimitOrderForPair(
	ctx context.Context,
	pair marketPair,
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
		pair,
		values,
	)
	if err != nil {
		return "", err
	}

	res := struct{
		Id string `json:"clientOrderId"`
	}{}

	err = json.Unmarshal(body, &res)
	if err != nil {
		return "", err
	}

	return res.Id, nil
}

func (c *client) StopOrder(ctx context.Context, orderId string) error {

	return c.StopOrderForPair(ctx, BTCEUR, orderId)
}

func (c *client) StopOrderForPair(
	ctx context.Context,
	pair marketPair,
	orderId string,
) error {

	values := url.Values{}
	values.Add("origClientOrderId", orderId)

	_, err := requestToOrderEndpointWithAuth(
		"DELETE",
		c.httpClient,
		c.apiKey,
		c.apiSecret,
		pair,
		values,
	)
	return err
}

func (c *client) GetOrderStatus(
	ctx context.Context,
	orderId string,
) (exchangesdk.OrderStatus, error) {

	return c.GetOrderStatusForPair(ctx, BTCEUR, orderId)
}

func (c *client) GetOrderStatusForPair(
	ctx context.Context,
	pair marketPair,
	orderId string,
) (exchangesdk.OrderStatus, error) {

	values := url.Values{}
	values.Add("origClientOrderId", orderId)

	body, err := requestToOrderEndpointWithAuth(
		"GET",
		c.httpClient,
		c.apiKey,
		c.apiSecret,
		pair,
		values,
	)
	if err != nil {
		return exchangesdk.OrderStatus{}, err
	}

	res := struct{
		Status string `json:"status"`
		Side string `json:"side"`
		FillAmount decimal.Decimal `json:"executedQty"`
	}{}

	err = json.Unmarshal(body, &res)
	if err != nil {
		return exchangesdk.OrderStatus{}, err
	}

	state := exchangesdk.OrderStatePending
	if res.Status == "FILLED" {
		state = exchangesdk.OrderStateComplete
	}

	orderType := exchangesdk.OrderTypeBid
	if res.Side == "SELL" {
		orderType = exchangesdk.OrderTypeAsk
	}

	return exchangesdk.OrderStatus{
		State: state,
		Type: orderType,
		FillAmountBase: res.FillAmount,
	}, nil
}

func (c *client) GetTrades(ctx context.Context, page int64) ([]exchangesdk.Trade, error) {

	panic("not implemented")
}

func (c *client) MakerFee() decimal.Decimal {

	return decimal.New(75, -5)
}

func requestToOrderEndpointWithAuth(
	reqMethod string,
	httpClient *http.Client,
	apiKey string,
	apiSecret string,
	pair marketPair,
	values url.Values,
) ([]byte, error) {

	path := requestutil.FullPath(baseUrl, "/api/v3/order")

	nowMs := utiltime.Now().Round(time.Millisecond).UnixNano() / 1e6
	timestampStr := strconv.FormatInt(nowMs, 10)
	values.Add("timestamp", timestampStr)
	values.Add("symbol", string(pair))

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
			ErrMsg string `json:"msg"`
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
