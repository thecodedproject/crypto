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
	"fmt"
	"crypto/hmac"
	"crypto/sha256"
	"strings"
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

	path := requestutil.FullPath(baseUrl, "/api/v3/order")

	values := url.Values{}
	nowMs := utiltime.Now().Round(time.Millisecond).UnixNano() / 1e6
	timestampStr := strconv.FormatInt(nowMs, 10)
	values.Add("timestamp", timestampStr)

	side := "BUY"
	if order.Type == exchangesdk.OrderTypeAsk {
		side = "SELL"
	}

	values.Add("type", "LIMIT")
	values.Add("side", side)
	values.Add("timeInForce", "GTC")
	values.Add("symbol", string(pair))
	values.Add("quantity", order.Volume.String())
	values.Add("price", order.Price.String())

	body, err := postRequestWithHmacAuth(
		c.httpClient,
		c.apiKey,
		c.apiSecret,
		path,
		values,
	)
	if err != nil {
		return "", err
	}

	res := struct{
		Id int64 `json:"orderId"`
	}{}

	err = json.Unmarshal(body, &res)
	if err != nil {
		return "", err
	}

	return strconv.FormatInt(res.Id, 10), nil
}

func postRequestWithHmacAuth(
	c *http.Client,
	key string,
	secret string,
	fullUrl *url.URL,
	values url.Values,
) ([]byte, error) {

	msgToSign := fmt.Sprint(fullUrl.Query().Encode(), values.Encode())

	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(msgToSign))
	signature := hex.EncodeToString(
		mac.Sum(nil),
	)

	values.Add("signature", signature)

	reqBody := strings.NewReader(values.Encode())
	reqMethod := "POST"
	req, err := http.NewRequest(reqMethod, fullUrl.String(), reqBody)
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
