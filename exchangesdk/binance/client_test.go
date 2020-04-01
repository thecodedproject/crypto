package binance_test

import (
	"context"
	//"fmt"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/thecodedproject/crypto/exchangesdk"
	"github.com/thecodedproject/crypto/exchangesdk/binance"
	"github.com/thecodedproject/crypto/exchangesdk/requestutil"
	"github.com/thecodedproject/crypto/util"
	utiltime "github.com/thecodedproject/crypto/util/time"
	"net/http"
	"strconv"
	"testing"
	"time"
)

func timeAsMsStr(t time.Time) string {

	return strconv.FormatInt(
		t.Round(time.Millisecond).UnixNano()/1e6,
		10,
	)
}

func TestLatestPriceWhenBinanceReturns200(t *testing.T) {

	pair := binance.BTCEUR

	handlerCalled := false
	c := binance.NewClientForTesting(t, "k", "s", func(req *http.Request) *http.Response {

		handlerCalled = true
		assert.Equal(
			t,
			"https://api.binance.com/api/v3/ticker/price?symbol=" + string(pair),
			req.URL.String(),
		)

		return &http.Response{
			StatusCode: 200,
			Body: requestutil.ResBodyFromJsonf(
				t,
				"{\"price\": \"123.4\"}",
			),
		}
	})

	val, err := c.LatestPrice(context.Background())
	require.NoError(t, err)

	assert.True(t, handlerCalled)
	util.LogicallyEqual(t, decimal.New(1234, -1), val)
}

func TestLatestPriceWhenBinanceReturns400WithError(t *testing.T) {

	pair := binance.BTCEUR

	errorMsg := "some error"

	handlerCalled := false
	c := binance.NewClientForTesting(t, "k", "s", func(req *http.Request) *http.Response {

		handlerCalled = true
		assert.Equal(
			t,
			"https://api.binance.com/api/v3/ticker/price?symbol=" + string(pair),
			req.URL.String(),
		)

		return &http.Response{
			StatusCode: 400,
			Body: requestutil.ResBodyFromJsonf(
				t,
				"{\"code\": 123, \"msg\": \"%s\"}",
				errorMsg,
			),
		}
	})

	_, err := c.LatestPrice(context.Background())
	require.Error(t, err)
	assert.Contains(t, err.Error(), errorMsg)
	assert.True(t, handlerCalled)
}

func TestSuccessfulPostBuyLimitOrder(t *testing.T) {

	pair := binance.BTCEUR
	order := exchangesdk.Order{
		Type: exchangesdk.OrderTypeBid,
		Price: decimal.New(1234, -1),
		Volume: decimal.New(5678, -2),
	}
	expectedId := "12345"

	nowTime := time.Unix(12345, 0)
	reset := utiltime.SetTimeNowForTesting(t, nowTime)
	defer reset()

	handlerCalled := false
	c := binance.NewClientForTesting(t, "k", "s", func(req *http.Request) *http.Response {

		handlerCalled = true
		assert.Equal(
			t,
			"https://api.binance.com/api/v3/order",
			req.URL.String(),
		)

		values := requestutil.GetReqValues(t, req)

		assert.Equal(
			t,
			"9fbe3b9a1e92dc1dbcc7298ba8a54b3e4572db6d046dc41079081a5ff072c863",
			values.Get("signature"),
		)
		assert.Equal(t, timeAsMsStr(nowTime),	values.Get("timestamp"))
		assert.Equal(t, "LIMIT", values.Get("type"))
		assert.Equal(t, "BUY", values.Get("side"))
		assert.Equal(t, "GTC", values.Get("timeInForce"))
		assert.Equal(t, string(pair), values.Get("symbol"))
		assert.Equal(t, order.Volume.String(), values.Get("quantity"))
		assert.Equal(t, order.Price.String(), values.Get("price"))

		assert.Equal(t, "k", req.Header.Get("X-MBX-APIKEY"))

		return &http.Response{
			StatusCode: 200,
			Body: requestutil.ResBodyFromJsonf(t, "{\"orderId\": %s}", expectedId),
		}
	})

	id, err := c.PostLimitOrder(context.Background(), order)
	require.NoError(t, err)

	assert.True(t, handlerCalled)
	assert.Equal(t, expectedId, id)
}

func TestSuccessfulPostSellLimitOrder(t *testing.T) {

	pair := binance.BTCEUR
	order := exchangesdk.Order{
		Type: exchangesdk.OrderTypeAsk,
		Price: decimal.New(1232, -1),
		Volume: decimal.New(5671, -2),
	}
	expectedId := "12346"

	nowTime := time.Unix(12876, 0)
	reset := utiltime.SetTimeNowForTesting(t, nowTime)
	defer reset()

	handlerCalled := false
	c := binance.NewClientForTesting(t, "k", "s", func(req *http.Request) *http.Response {

		handlerCalled = true
		assert.Equal(
			t,
			"https://api.binance.com/api/v3/order",
			req.URL.String(),
		)

		values := requestutil.GetReqValues(t, req)

		assert.Equal(
			t,
			"60b99103abfdb93655f2fca6cac7d6eb8fa2c2f7654bdbff09cbcae70b5cc3ce",
			values.Get("signature"),
		)
		assert.Equal(t, timeAsMsStr(nowTime),	values.Get("timestamp"))
		assert.Equal(t, "LIMIT", values.Get("type"))
		assert.Equal(t, "SELL", values.Get("side"))
		assert.Equal(t, "GTC", values.Get("timeInForce"))
		assert.Equal(t, string(pair), values.Get("symbol"))
		assert.Equal(t, order.Volume.String(), values.Get("quantity"))
		assert.Equal(t, order.Price.String(), values.Get("price"))

		assert.Equal(t, "k", req.Header.Get("X-MBX-APIKEY"))

		return &http.Response{
			StatusCode: 200,
			Body: requestutil.ResBodyFromJsonf(t, "{\"orderId\": %s}", expectedId),
		}
	})

	id, err := c.PostLimitOrder(context.Background(), order)
	require.NoError(t, err)

	assert.True(t, handlerCalled)
	assert.Equal(t, expectedId, id)
}

func TestSuccessfulPostSellLimitOrderWhichReturns4XX(t *testing.T) {

	pair := binance.BTCEUR
	order := exchangesdk.Order{
		Type: exchangesdk.OrderTypeAsk,
		Price: decimal.New(1232, -1),
		Volume: decimal.New(5671, -2),
	}

	nowTime := time.Unix(12876, 0)
	reset := utiltime.SetTimeNowForTesting(t, nowTime)

	errorMsg := "some error msg"
	defer reset()

	handlerCalled := false
	c := binance.NewClientForTesting(t, "k", "s", func(req *http.Request) *http.Response {

		handlerCalled = true
		assert.Equal(
			t,
			"https://api.binance.com/api/v3/order",
			req.URL.String(),
		)

		values := requestutil.GetReqValues(t, req)

		assert.Equal(
			t,
			"60b99103abfdb93655f2fca6cac7d6eb8fa2c2f7654bdbff09cbcae70b5cc3ce",
			values.Get("signature"),
		)
		assert.Equal(t, timeAsMsStr(nowTime),	values.Get("timestamp"))
		assert.Equal(t, "LIMIT", values.Get("type"))
		assert.Equal(t, "SELL", values.Get("side"))
		assert.Equal(t, "GTC", values.Get("timeInForce"))
		assert.Equal(t, string(pair), values.Get("symbol"))
		assert.Equal(t, order.Volume.String(), values.Get("quantity"))
		assert.Equal(t, order.Price.String(), values.Get("price"))

		assert.Equal(t, "k", req.Header.Get("X-MBX-APIKEY"))

		return &http.Response{
			StatusCode: 403,
			Body: requestutil.ResBodyFromJsonf(t,
				"{\"code\": -1234, \"msg\": \"%s\"}",
				errorMsg,
			),
		}
	})

	_, err := c.PostLimitOrder(context.Background(), order)
	require.Error(t, err)
	assert.Contains(t, err.Error(), errorMsg)
	assert.True(t, handlerCalled)
}
