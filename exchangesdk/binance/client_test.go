package binance_test

import (
	"context"
	"net/http"
	"strconv"
	"testing"
	"time"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/thecodedproject/crypto/exchangesdk"
	"github.com/thecodedproject/crypto/exchangesdk/binance"
	"github.com/thecodedproject/crypto/exchangesdk/requestutil"
	"github.com/thecodedproject/crypto/util"
	utiltime "github.com/thecodedproject/crypto/util/time"
)

func timeAsMsStr(t time.Time) string {

	return strconv.FormatInt(
		t.Round(time.Millisecond).UnixNano()/1e6,
		10,
	)
}

func TestLatestPriceWhenBinanceReturns200(t *testing.T) {

	pair := "BTCEUR"

	handlerCalled := false
	c := binance.NewClientForTesting(t, "k", "s", pair, func(req *http.Request) *http.Response {

		handlerCalled = true
		assert.Equal(
			t,
			"https://api.binance.com/api/v3/ticker/price?symbol="+string(pair),
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

	pair := "BTCEUR"

	errorMsg := "some error"

	handlerCalled := false
	c := binance.NewClientForTesting(t, "k", "s", pair, func(req *http.Request) *http.Response {

		handlerCalled = true
		assert.Equal(
			t,
			"https://api.binance.com/api/v3/ticker/price?symbol="+string(pair),
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

	pair := "BTCEUR"
	order := exchangesdk.Order{
		Type:   exchangesdk.OrderTypeBid,
		Price:  decimal.New(1234, -1),
		Volume: decimal.New(5678, -2),
	}
	expectedId := "dce12345"

	nowTime := time.Unix(12345, 0)
	reset := utiltime.SetTimeNowForTesting(t, nowTime)
	defer reset()

	handlerCalled := false
	c := binance.NewClientForTesting(t, "k", "s", pair, func(req *http.Request) *http.Response {

		handlerCalled = true
		assert.Contains(
			t,
			req.URL.String(),
			"https://api.binance.com/api/v3/order",
		)
		assert.Equal(t, "POST", req.Method)

		values := req.URL.Query()

		assert.Equal(
			t,
			"9fbe3b9a1e92dc1dbcc7298ba8a54b3e4572db6d046dc41079081a5ff072c863",
			values.Get("signature"),
		)
		assert.Equal(t, timeAsMsStr(nowTime), values.Get("timestamp"))
		assert.Equal(t, "LIMIT", values.Get("type"))
		assert.Equal(t, "BUY", values.Get("side"))
		assert.Equal(t, "GTC", values.Get("timeInForce"))
		assert.Equal(t, string(pair), values.Get("symbol"))
		assert.Equal(t, order.Volume.String(), values.Get("quantity"))
		assert.Equal(t, order.Price.String(), values.Get("price"))

		assert.Equal(t, "k", req.Header.Get("X-MBX-APIKEY"))

		return &http.Response{
			StatusCode: 200,
			Body:       requestutil.ResBodyFromJsonf(t, "{\"clientOrderId\": \"%s\"}", expectedId),
		}
	})

	id, err := c.PostLimitOrder(context.Background(), order)
	require.NoError(t, err)

	assert.True(t, handlerCalled)
	assert.Equal(t, expectedId, id)
}

func TestSuccessfulPostSellLimitOrder(t *testing.T) {

	pair := "BTCEUR"
	order := exchangesdk.Order{
		Type:   exchangesdk.OrderTypeAsk,
		Price:  decimal.New(1232, -1),
		Volume: decimal.New(5671, -2),
	}
	expectedId := "abc12346"

	nowTime := time.Unix(12876, 0)
	reset := utiltime.SetTimeNowForTesting(t, nowTime)
	defer reset()

	handlerCalled := false
	c := binance.NewClientForTesting(t, "k", "s", pair, func(req *http.Request) *http.Response {

		handlerCalled = true
		assert.Contains(
			t,
			req.URL.String(),
			"https://api.binance.com/api/v3/order",
		)
		assert.Equal(t, "POST", req.Method)

		values := req.URL.Query()

		assert.Equal(
			t,
			"60b99103abfdb93655f2fca6cac7d6eb8fa2c2f7654bdbff09cbcae70b5cc3ce",
			values.Get("signature"),
		)
		assert.Equal(t, timeAsMsStr(nowTime), values.Get("timestamp"))
		assert.Equal(t, "LIMIT", values.Get("type"))
		assert.Equal(t, "SELL", values.Get("side"))
		assert.Equal(t, "GTC", values.Get("timeInForce"))
		assert.Equal(t, string(pair), values.Get("symbol"))
		assert.Equal(t, order.Volume.String(), values.Get("quantity"))
		assert.Equal(t, order.Price.String(), values.Get("price"))

		assert.Equal(t, "k", req.Header.Get("X-MBX-APIKEY"))

		return &http.Response{
			StatusCode: 200,
			Body:       requestutil.ResBodyFromJsonf(t, "{\"clientOrderId\": \"%s\"}", expectedId),
		}
	})

	id, err := c.PostLimitOrder(context.Background(), order)
	require.NoError(t, err)

	assert.True(t, handlerCalled)
	assert.Equal(t, expectedId, id)
}

func TestSuccessfulPostSellLimitOrderWhichReturns4XX(t *testing.T) {

	pair := "BTCEUR"
	order := exchangesdk.Order{
		Type:   exchangesdk.OrderTypeAsk,
		Price:  decimal.New(1232, -1),
		Volume: decimal.New(5671, -2),
	}

	nowTime := time.Unix(12876, 0)
	reset := utiltime.SetTimeNowForTesting(t, nowTime)

	errorMsg := "some error msg"
	defer reset()

	handlerCalled := false
	c := binance.NewClientForTesting(t, "k", "s", pair, func(req *http.Request) *http.Response {

		handlerCalled = true
		assert.Contains(
			t,
			req.URL.String(),
			"https://api.binance.com/api/v3/order",
		)
		assert.Equal(t, "POST", req.Method)

		values := req.URL.Query()

		assert.Equal(
			t,
			"60b99103abfdb93655f2fca6cac7d6eb8fa2c2f7654bdbff09cbcae70b5cc3ce",
			values.Get("signature"),
		)
		assert.Equal(t, timeAsMsStr(nowTime), values.Get("timestamp"))
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

func TestPostStopLimitOrder(t *testing.T) {

	pair := "BTCEUR"
	order := exchangesdk.StopLimitOrder{
		Side:       exchangesdk.OrderBookSideAsk,
		StopPrice:  decimal.New(3456, -1),
		LimitPrice: decimal.New(1232, -1),
		Volume:     decimal.New(5671, -2),
	}
	expectedId := "abc12346"

	nowTime := time.Unix(12876, 0)
	reset := utiltime.SetTimeNowForTesting(t, nowTime)
	defer reset()

	handlerCalled := false
	c := binance.NewClientForTesting(t, "k", "s", pair, func(req *http.Request) *http.Response {

		handlerCalled = true
		assert.Contains(
			t,
			req.URL.String(),
			"https://api.binance.com/api/v3/order",
		)
		assert.Equal(t, "POST", req.Method)

		values := req.URL.Query()

		assert.Equal(
			t,
			"b9b662465346539b0c4c818aacb2e569737413b18633e571f11a260fb4c775ab",
			values.Get("signature"),
		)
		assert.Equal(t, timeAsMsStr(nowTime), values.Get("timestamp"))
		assert.Equal(t, "STOP_LOSS_LIMIT", values.Get("type"))
		assert.Equal(t, "SELL", values.Get("side"))
		assert.Equal(t, "GTC", values.Get("timeInForce"))
		assert.Equal(t, string(pair), values.Get("symbol"))
		assert.Equal(t, order.Volume.String(), values.Get("quantity"))
		assert.Equal(t, order.LimitPrice.String(), values.Get("price"))
		assert.Equal(t, order.StopPrice.String(), values.Get("stopPrice"))

		assert.Equal(t, "k", req.Header.Get("X-MBX-APIKEY"))

		return &http.Response{
			StatusCode: 200,
			Body:       requestutil.ResBodyFromJsonf(t, "{\"clientOrderId\": \"%s\"}", expectedId),
		}
	})

	id, err := c.PostStopLimitOrder(context.Background(), order)
	require.NoError(t, err)

	assert.True(t, handlerCalled)
	assert.Equal(t, expectedId, id)
}

func TestSuccessfulCancelLimitOrder(t *testing.T) {

	pair := "BTCEUR"
	orderId := "12346"

	nowTime := time.Unix(13876, 0)
	reset := utiltime.SetTimeNowForTesting(t, nowTime)
	defer reset()

	handlerCalled := false
	c := binance.NewClientForTesting(t, "k", "s", pair, func(req *http.Request) *http.Response {

		handlerCalled = true
		assert.Contains(
			t,
			req.URL.String(),
			"https://api.binance.com/api/v3/order",
		)
		assert.Equal(t, "DELETE", req.Method)

		values := req.URL.Query()

		assert.Equal(
			t,
			"2e59d9fe933e4a7426488fff23ab145c9350d2cb61949ac7eef3cd46c3b164a2",
			values.Get("signature"),
		)
		assert.Equal(t, timeAsMsStr(nowTime), values.Get("timestamp"))
		assert.Equal(t, string(pair), values.Get("symbol"))
		assert.Equal(t, orderId, values.Get("origClientOrderId"))

		assert.Equal(t, "k", req.Header.Get("X-MBX-APIKEY"))

		return &http.Response{
			StatusCode: 200,
			Body: requestutil.ResBodyFromJsonf(
				t,
				"{\"clientOrderId\": \"%s\", \"status\": \"CANCELED\"}",
				orderId,
			),
		}
	})

	err := c.CancelOrder(context.Background(), orderId)
	require.NoError(t, err)
	assert.True(t, handlerCalled)
}

func TestUnsuccessfulCancelLimitOrder(t *testing.T) {

	pair := "BTCEUR"
	orderId := "12346"

	nowTime := time.Unix(13876, 0)
	reset := utiltime.SetTimeNowForTesting(t, nowTime)
	defer reset()

	errorMsg := "some error message"
	handlerCalled := false
	c := binance.NewClientForTesting(t, "k", "s", pair, func(req *http.Request) *http.Response {

		handlerCalled = true
		assert.Contains(
			t,
			req.URL.String(),
			"https://api.binance.com/api/v3/order",
		)
		assert.Equal(t, "DELETE", req.Method)

		values := req.URL.Query()

		assert.Equal(
			t,
			"2e59d9fe933e4a7426488fff23ab145c9350d2cb61949ac7eef3cd46c3b164a2",
			values.Get("signature"),
		)
		assert.Equal(t, timeAsMsStr(nowTime), values.Get("timestamp"))
		assert.Equal(t, string(pair), values.Get("symbol"))
		assert.Equal(t, orderId, values.Get("origClientOrderId"))

		assert.Equal(t, "k", req.Header.Get("X-MBX-APIKEY"))

		return &http.Response{
			StatusCode: 400,
			Body: requestutil.ResBodyFromJsonf(
				t,
				"{\"code\": -1235, \"msg\": \"%s\"}",
				errorMsg,
			),
		}
	})

	err := c.CancelOrder(context.Background(), orderId)
	require.Error(t, err)
	assert.Contains(t, err.Error(), errorMsg)
	assert.True(t, handlerCalled)
}

func TestSuccessfulGetOrderStatus(t *testing.T) {

	testCases := []struct {
		name           string
		resBody        string
		expectedStatus exchangesdk.OrderStatus
	}{
		{
			name:    "Bid stop limit order, not triggered, no fill",
			resBody: "{\"executedQty\": \"0.0\", \"status\": \"NEW\", \"side\": \"BUY\", \"isWorking\": false}",
			expectedStatus: exchangesdk.OrderStatus{
				State:          exchangesdk.OrderStateAwaitingTrigger,
				Type:           exchangesdk.OrderTypeBid,
				FillAmountBase: decimal.Decimal{},
			},
		},
		{
			name:    "Bid order, pending, no fill",
			resBody: "{\"executedQty\": \"0.0\", \"status\": \"NEW\", \"side\": \"BUY\", \"isWorking\": true}",
			expectedStatus: exchangesdk.OrderStatus{
				State:          exchangesdk.OrderStateInOrderBook,
				Type:           exchangesdk.OrderTypeBid,
				FillAmountBase: decimal.Decimal{},
			},
		},
		{
			name:    "Bid order, pending, partial fill",
			resBody: "{\"executedQty\": \"1.23\", \"status\": \"PARTIALLY_FILLED\", \"side\": \"BUY\", \"isWorking\": true, \"cummulativeQuoteQty\": \"45.6\"}",
			expectedStatus: exchangesdk.OrderStatus{
				State:             exchangesdk.OrderStateInOrderBook,
				Type:              exchangesdk.OrderTypeBid,
				FillAmountBase:    decimal.New(123, -2),
				FillAmountCounter: decimal.New(456, -1),
			},
		},
		{
			name:    "Bid order, completed, filled",
			resBody: "{\"executedQty\": \"2.23\", \"status\": \"FILLED\", \"side\": \"BUY\", \"isWorking\": true, \"cummulativeQuoteQty\": \"7.89\"}",
			expectedStatus: exchangesdk.OrderStatus{
				State:             exchangesdk.OrderStateFilled,
				Type:              exchangesdk.OrderTypeBid,
				FillAmountBase:    decimal.New(223, -2),
				FillAmountCounter: decimal.New(789, -2),
			},
		},
		{
			name:    "Ask order, pending, no fill",
			resBody: "{\"executedQty\": \"0.0\", \"status\": \"NEW\", \"side\": \"SELL\", \"isWorking\": true}",
			expectedStatus: exchangesdk.OrderStatus{
				State:          exchangesdk.OrderStateInOrderBook,
				Type:           exchangesdk.OrderTypeAsk,
				FillAmountBase: decimal.Decimal{},
			},
		},
		{
			name:    "Ask order, pending, partial fill",
			resBody: "{\"executedQty\": \"1.23\", \"status\": \"PARTIALLY_FILLED\", \"side\": \"SELL\", \"isWorking\": true, \"cummulativeQuoteQty\": \"4.89\"}",
			expectedStatus: exchangesdk.OrderStatus{
				State:             exchangesdk.OrderStateInOrderBook,
				Type:              exchangesdk.OrderTypeAsk,
				FillAmountBase:    decimal.New(123, -2),
				FillAmountCounter: decimal.New(489, -2),
			},
		},
		{
			name:    "Ask order, completed, filled",
			resBody: "{\"executedQty\": \"4.23\", \"status\": \"FILLED\", \"side\": \"SELL\", \"isWorking\": true, \"cummulativeQuoteQty\": \"3.89\"}",
			expectedStatus: exchangesdk.OrderStatus{
				State:             exchangesdk.OrderStateFilled,
				Type:              exchangesdk.OrderTypeAsk,
				FillAmountBase:    decimal.New(423, -2),
				FillAmountCounter: decimal.New(389, -2),
			},
		},
	}

	for _, test := range testCases {
		t.Run(test.name, func(t *testing.T) {

			pair := "BTCEUR"
			orderId := "sda12346"

			nowTime := time.Unix(13876, 0)
			reset := utiltime.SetTimeNowForTesting(t, nowTime)
			defer reset()

			handlerCalled := false
			c := binance.NewClientForTesting(t, "k", "s", pair, func(req *http.Request) *http.Response {

				handlerCalled = true
				assert.Contains(
					t,
					req.URL.String(),
					"https://api.binance.com/api/v3/order",
				)
				assert.Equal(t, "GET", req.Method)

				values := req.URL.Query()

				assert.Equal(
					t,
					"981ba3a377fd892908878bfcd98b889743037e7e8c8468bded3389a6ca062b9b",
					values.Get("signature"),
				)
				assert.Equal(t, timeAsMsStr(nowTime), values.Get("timestamp"))
				assert.Equal(t, string(pair), values.Get("symbol"))
				assert.Equal(t, orderId, values.Get("origClientOrderId"))

				assert.Equal(t, "k", req.Header.Get("X-MBX-APIKEY"))

				return &http.Response{
					StatusCode: 200,
					Body: requestutil.ResBodyFromJsonf(
						t,
						test.resBody,
					),
				}
			})

			actualStatus, err := c.GetOrderStatus(context.Background(), orderId)
			require.NoError(t, err)
			assert.True(t, handlerCalled)

			util.LogicallyEqual(t, test.expectedStatus, actualStatus)
		})
	}
}
