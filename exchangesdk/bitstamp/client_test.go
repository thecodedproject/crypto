package bitstamp_test

import (
	"bytes"
	"context"
	"fmt"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/thecodedproject/crypto/exchangesdk"
	"github.com/thecodedproject/crypto/exchangesdk/bitstamp"
	"github.com/thecodedproject/crypto/util"
	utiltime "github.com/thecodedproject/crypto/util/time"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"testing"
	"time"
)

func TestLatestPriceWhenBitstampReturns200(t *testing.T) {

	handlerCalled := false
	c := bitstamp.NewClientForTesting(t, "k", "s", func(req *http.Request) *http.Response {

		handlerCalled = true
		assert.Equal(
			t,
			"https://www.bitstamp.net/api/v2/ticker/btceur/",
			req.URL.String(),
		)

		return &http.Response{
			StatusCode: 200,
			Body: ioutil.NopCloser(bytes.NewBufferString(
				"{\"last\": \"123.4\"}",
			)),
		}
	})

	val, err := c.LatestPrice(context.Background())
	require.NoError(t, err)

	assert.True(t, handlerCalled)
	util.LogicallyEqual(t, decimal.New(1234, -1), val)
}

func TestLatestPriceWhenBitstampReturns4XXReturnsError(t *testing.T) {

	handlerCalled := false
	c := bitstamp.NewClientForTesting(t, "k", "s", func(req *http.Request) *http.Response {

		handlerCalled = true
		assert.Equal(
			t,
			"https://www.bitstamp.net/api/v2/ticker/btceur/",
			req.URL.String(),
		)

		return &http.Response{
			Status: "403 - Not Authorised",
			StatusCode: 403,
		}
	})

	_, err := c.LatestPrice(context.Background())
	require.Error(t, err)
	assert.True(t, handlerCalled)
}


func TestSuccessfulPostBuyLimitOrder(t *testing.T) {

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
	c := bitstamp.NewClientForTesting(t, "k", "s", func(req *http.Request) *http.Response {

		handlerCalled = true
		assert.Equal(
			t,
			"https://www.bitstamp.net/api/v2/buy/btceur/",
			req.URL.String(),
		)

		reqValues := getReqValues(t, req)
		assert.Equal(t, order.Price.String(), reqValues.Get("price"))
		assert.Equal(t, order.Volume.String(), reqValues.Get("amount"))

		checkReqHeaders(
			t,
			req,
			"7ec2a7c4b9c29d36a3e4923d2e8e9c29b59b8b4be8aa2f776a78426dd9486eb5",
			nowTime,
		)

		return &http.Response{
			StatusCode: 200,
			Body: resBodyFromJsonf("{\"id\": \"%s\"}", expectedId),
			Header: resHeaders(
				"93af757fa082515e061483b197245709befc2cc25db3df013fe6696b9c164883",
			),
		}
	})

	id, err := c.PostLimitOrder(context.Background(), order)
	require.NoError(t, err)

	assert.True(t, handlerCalled)
	assert.Equal(t, expectedId, id)
}

func TestSuccessfulPostSellLimitOrder(t *testing.T) {

	order := exchangesdk.Order{
		Type: exchangesdk.OrderTypeAsk,
		Price: decimal.New(1234, -1),
		Volume: decimal.New(5678, -2),
	}
	expectedId := "12345"

	nowTime := time.Unix(12345, 0)
	reset := utiltime.SetTimeNowForTesting(t, nowTime)
	defer reset()

	handlerCalled := false
	c := bitstamp.NewClientForTesting(t, "k", "s", func(req *http.Request) *http.Response {

		handlerCalled = true
		assert.Equal(
			t,
			"https://www.bitstamp.net/api/v2/sell/btceur/",
			req.URL.String(),
		)

		reqValues := getReqValues(t, req)
		assert.Equal(t, order.Price.String(), reqValues.Get("price"))
		assert.Equal(t, order.Volume.String(), reqValues.Get("amount"))

		checkReqHeaders(
			t,
			req,
			"75319f25b20e8969ff1ab8af03335fe5d4177ba169366963788712a6d6682ded",
			nowTime,
		)

		return &http.Response{
			StatusCode: 200,
			Body: resBodyFromJsonf("{\"id\": \"%s\"}", expectedId),
			Header: resHeaders(
				"93af757fa082515e061483b197245709befc2cc25db3df013fe6696b9c164883",
			),
		}
	})

	id, err := c.PostLimitOrder(context.Background(), order)
	require.NoError(t, err)

	assert.True(t, handlerCalled)
	assert.Equal(t, expectedId, id)
}

func TestUnsuccessfulPostBuyLimitOrder(t *testing.T) {

	order := exchangesdk.Order{
		Type: exchangesdk.OrderTypeBid,
		Price: decimal.New(1234, -1),
		Volume: decimal.New(5678, -2),
	}

	nowTime := time.Unix(12345, 0)
	reset := utiltime.SetTimeNowForTesting(t, nowTime)
	defer reset()

	handlerCalled := false
	c := bitstamp.NewClientForTesting(t, "k", "s", func(req *http.Request) *http.Response {

		handlerCalled = true
		assert.Equal(
			t,
			"https://www.bitstamp.net/api/v2/buy/btceur/",
			req.URL.String(),
		)

		reqValues := getReqValues(t, req)
		assert.Equal(t, order.Price.String(), reqValues.Get("price"))
		assert.Equal(t, order.Volume.String(), reqValues.Get("amount"))

		checkReqHeaders(
			t,
			req,
			"7ec2a7c4b9c29d36a3e4923d2e8e9c29b59b8b4be8aa2f776a78426dd9486eb5",
			nowTime,
		)

		return &http.Response{
			StatusCode: 200,
			Body: resBodyFromJsonf("{\"status\": \"error\", \"reason\": \"some error\"}"),
			Header: resHeaders(
				"085dd6e4f5d825e009bd2f1fc22faca64717c6591ca26c8cc37b81ebb7e12dbf",
			),
		}
	})

	_, err := c.PostLimitOrder(context.Background(), order)
	require.Error(t, err)
	assert.True(t, handlerCalled)
}

func TestPostLimitOrderWhichReturns4XX(t *testing.T) {

	order := exchangesdk.Order{
		Type: exchangesdk.OrderTypeBid,
		Price: decimal.New(1234, -1),
		Volume: decimal.New(5678, -2),
	}

	nowTime := time.Unix(12345, 0)
	reset := utiltime.SetTimeNowForTesting(t, nowTime)
	defer reset()

	handlerCalled := false
	c := bitstamp.NewClientForTesting(t, "k", "s", func(req *http.Request) *http.Response {

		handlerCalled = true
		assert.Equal(
			t,
			"https://www.bitstamp.net/api/v2/buy/btceur/",
			req.URL.String(),
		)

		reqValues := getReqValues(t, req)
		assert.Equal(t, order.Price.String(), reqValues.Get("price"))
		assert.Equal(t, order.Volume.String(), reqValues.Get("amount"))

		checkReqHeaders(
			t,
			req,
			"7ec2a7c4b9c29d36a3e4923d2e8e9c29b59b8b4be8aa2f776a78426dd9486eb5",
			nowTime,
		)

		return &http.Response{
			StatusCode: 403,
		}
	})

	_, err := c.PostLimitOrder(context.Background(), order)
	require.Error(t, err)
	assert.True(t, handlerCalled)
}

func TestPostLimitOrderWithDefaultOrderTypeReturnsError(t *testing.T) {

	order := exchangesdk.Order{
		Price: decimal.New(1234, -1),
		Volume: decimal.New(5678, -2),
	}

	c := bitstamp.NewClientForTesting(t, "k", "s", func(req *http.Request) *http.Response {

		require.Fail(t, "Must not make http request")
		return &http.Response{
			StatusCode: 500,
		}
	})

	_, err := c.PostLimitOrder(context.Background(), order)
	require.Error(t, err)
}

func TestSuccessfulStopOrder(t *testing.T) {

	orderId := "1234565432"

	nowTime := time.Unix(12345, 0)
	reset := utiltime.SetTimeNowForTesting(t, nowTime)
	defer reset()

	handlerCalled := false
	c := bitstamp.NewClientForTesting(t, "k", "s", func(req *http.Request) *http.Response {

		handlerCalled = true
		assert.Equal(
			t,
			"https://www.bitstamp.net/api/v2/cancel_order/",
			req.URL.String(),
		)

		reqValues := getReqValues(t, req)
		assert.Equal(t, orderId, reqValues.Get("id"))

		checkReqHeaders(
			t,
			req,
			"3d6ac45913e205377ff4b7470ce16dfbb07bf020d09ced8457873cb2eed51118",
			nowTime,
		)

		return &http.Response{
			StatusCode: 200,
			Body: resBodyFromJsonf("{\"id\": \"%s\"}", orderId),
			Header: resHeaders(
				"1937abed95d1402bf187584041b039471b025e8677b6faa75ef1f793e323c69d",
			),
		}
	})

	err := c.StopOrder(context.Background(), orderId)
	require.NoError(t, err)
	assert.True(t, handlerCalled)
}

func TestUnsuccessfulStopOrder(t *testing.T) {

	orderId := "1234565432"

	nowTime := time.Unix(12345, 0)
	reset := utiltime.SetTimeNowForTesting(t, nowTime)
	defer reset()

	handlerCalled := false
	c := bitstamp.NewClientForTesting(t, "k", "s", func(req *http.Request) *http.Response {

		handlerCalled = true
		assert.Equal(
			t,
			"https://www.bitstamp.net/api/v2/cancel_order/",
			req.URL.String(),
		)

		reqValues := getReqValues(t, req)
		assert.Equal(t, orderId, reqValues.Get("id"))

		checkReqHeaders(
			t,
			req,
			"3d6ac45913e205377ff4b7470ce16dfbb07bf020d09ced8457873cb2eed51118",
			nowTime,
		)

		return &http.Response{
			StatusCode: 200,
			Body: resBodyFromJsonf("{\"error\": \"some_error\"}"),
			Header: resHeaders(
				"b721063e9d07f71eccf9f8b4861322fd4c877ef13f64cd9d61926150b21cc37b",
			),
		}
	})

	err := c.StopOrder(context.Background(), orderId)
	require.Error(t, err)
	assert.True(t, handlerCalled)
}

func resBodyFromJsonf(
	jsonStringf string,
	i ...interface{},
) io.ReadCloser {

	jsonString := fmt.Sprintf(jsonStringf, i...)
	jsonBuffer := bytes.NewBufferString(jsonString)
	return ioutil.NopCloser(jsonBuffer)
}

func getReqValues(
	t *testing.T,
	req *http.Request,
) url.Values {

	defer req.Body.Close()
	body, err := ioutil.ReadAll(req.Body)
	require.NoError(t, err)

	reqValues, err := url.ParseQuery(string(body))
	require.NoError(t, err)

	return reqValues
}

func checkReqHeaders(
	t *testing.T,
	req *http.Request,
	authSig string,
	timenow time.Time,
) {

	nonce := fmt.Sprintf("%036x", timenow.UnixNano())
	timestampStr := strconv.FormatInt(
		timenow.Round(time.Millisecond).UnixNano()/1e6,
		10,
	)

	assert.Equal(t, "BITSTAMP k", req.Header.Get("X-Auth"))
	assert.Equal(t, authSig, req.Header.Get("X-Auth-Signature"))
	assert.Equal(t, nonce, req.Header.Get("X-Auth-Nonce"))
	assert.Equal(t, timestampStr, req.Header.Get("X-Auth-Timestamp"))
	assert.Equal(t, "v2", req.Header.Get("X-Auth-Version"))
	assert.Equal(t, "application/x-www-form-urlencoded", req.Header.Get("Content-Type"))
}

func resHeaders(authSig string) http.Header {

	resHeaders := make(http.Header)
	resHeaders.Add("Content-Type", "application/x-www-form-urlencoded")
	resHeaders.Add("X-Server-Auth-Signature", authSig)

	return resHeaders
}
