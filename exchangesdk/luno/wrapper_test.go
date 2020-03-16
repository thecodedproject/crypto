package luno_test

import (
	"context"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/thecodedproject/crypto/exchangesdk"
	"github.com/thecodedproject/crypto/exchangesdk/luno"
	"strconv"
	"testing"
	luno_sdk "github.com/luno/luno-go"
	lunodecimal "github.com/luno/luno-go/decimal"
)

func makeSomeLunoTrades(n int64) []luno_sdk.Trade {

	trades := make([]luno_sdk.Trade, 0, n)
	for i:=int64(0); i<n; i++ {
		trades = append(trades, luno_sdk.Trade{
			OrderId: strconv.FormatInt(i, 10),
			Sequence: i,
		})
	}
	return trades
}

func makeSomeTrades(n int64) []exchangesdk.Trade {

	trades := make([]exchangesdk.Trade, 0, n)
	for i:=int64(0); i<n; i++ {
		trades = append(trades, exchangesdk.Trade{
			OrderId: strconv.FormatInt(i, 10),
		})
	}
	return trades
}

func TestGetTradesForPageLessThanOneReqturnsError(t *testing.T) {

	m := new(luno.MockLunoSdk)
	ctx := context.Background()
	c := luno.NewClientForTesting(t, m, 12, 34)
	_, err := c.GetTrades(ctx, 0)
	assert.Error(t, err)
}

func TestGetTradesWhenThereAreNoTrades(t *testing.T) {

	m := new(luno.MockLunoSdk)

	req := &luno_sdk.ListUserTradesRequest{
		Pair: luno.TRADINGPAIR,
	}
	var res luno_sdk.ListUserTradesResponse
	m.On("ListUserTrades", mock.Anything, req).Return(&res, nil)

	ctx := context.Background()
	c := luno.NewClientForTesting(t, m, 12, 34)
	trades, err := c.GetTrades(ctx, 1)
	require.NoError(t, err)

	assert.Equal(t, 0, len(trades))
}

func TestGetTradesFirstPageWhenThereAreSomeTrades(t *testing.T) {

	m := new(luno.MockLunoSdk)

	req := &luno_sdk.ListUserTradesRequest{
		Pair: luno.TRADINGPAIR,
	}
	res := luno_sdk.ListUserTradesResponse{
		Trades: []luno_sdk.Trade{
			{
				OrderId: "1",
				Sequence: 1,
			},
			{
				OrderId: "2",
				Sequence: 2,
			},
		},
	}
	m.On("ListUserTrades", mock.Anything, req).Return(&res, nil)

	expected := []exchangesdk.Trade{
		{
			OrderId: "1",
		},
		{
			OrderId: "2",
		},
	}

	ctx := context.Background()
	c := luno.NewClientForTesting(t, m, 12, 34)
	trades, err := c.GetTrades(ctx, 1)
	require.NoError(t, err)

	assert.Equal(t, expected, trades)
}

func TestGetTradesFirstPageMultipleTimesWhenFullOfTradesMultipleTimesOnlyCallsListTradesOnce(t *testing.T) {

	m := new(luno.MockLunoSdk)

	req := &luno_sdk.ListUserTradesRequest{
		Pair: luno.TRADINGPAIR,
	}
	res := luno_sdk.ListUserTradesResponse{
		Trades: makeSomeLunoTrades(100),
	}
	m.On("ListUserTrades", mock.Anything, req).Return(&res, nil).Once()

	expected := makeSomeTrades(100)

	ctx := context.Background()
	c := luno.NewClientForTesting(t, m, 12, 34)
	trades, err := c.GetTrades(ctx, 1)
	require.NoError(t, err)
	assert.Equal(t, expected, trades)

	trades, err = c.GetTrades(ctx, 1)
	require.NoError(t, err)
	assert.Equal(t, expected, trades)
}

func TestGetTradesSecondPageWhenFirstPageIsFull(t *testing.T) {

	m := new(luno.MockLunoSdk)

	req := luno_sdk.ListUserTradesRequest{
		Pair: luno.TRADINGPAIR,
	}
	res := luno_sdk.ListUserTradesResponse{
		Trades: makeSomeLunoTrades(100),
	}
	m.On("ListUserTrades", mock.Anything, &req).Return(&res, nil)

	req = luno_sdk.ListUserTradesRequest{
		Pair: luno.TRADINGPAIR,
		AfterSeq: 99,
	}
	res = luno_sdk.ListUserTradesResponse{
		Trades: []luno_sdk.Trade{
			{
				OrderId: "201",
				Sequence: 201,
			},
		},
	}
	m.On("ListUserTrades", mock.Anything, &req).Return(&res, nil)

	expected := []exchangesdk.Trade{
		{
			OrderId: "201",
		},
	}

	ctx := context.Background()
	c := luno.NewClientForTesting(t, m, 12, 34)
	trades, err := c.GetTrades(ctx, 2)
	require.NoError(t, err)

	assert.Equal(t, expected, trades)
}

// TODO remove below tests...

func lunoD(f float64) lunodecimal.Decimal {
	return lunodecimal.NewFromFloat64(f, 4)
}

func D(f float64) decimal.Decimal {
	return decimal.NewFromFloat(f)
}


func TestGetTradesSome(t *testing.T) {

	m := new(luno.MockLunoSdk)

	firstReq := &luno_sdk.ListUserTradesRequest{
		Pair: luno.TRADINGPAIR,
	}
	first := luno_sdk.ListUserTradesResponse{
		Trades: []luno_sdk.Trade{
			{
				Price: lunoD(1),
			},
		},
	}
	m.On("ListUserTrades", mock.Anything, firstReq).Return(&first, nil)
	secondReq := &luno_sdk.ListUserTradesRequest{
		Pair: luno.TRADINGPAIR,
		AfterSeq: 1,
	}
	var second luno_sdk.ListUserTradesResponse
	m.On("ListUserTrades", mock.Anything, secondReq).Return(&second, nil)

	ctx := context.Background()
	res, err := m.ListUserTrades(ctx, firstReq)
	require.NoError(t, err)
	assert.Equal(t, 1, len(res.Trades))

	res, err = m.ListUserTrades(ctx, secondReq)
	require.NoError(t, err)
	assert.Equal(t, 0, len(res.Trades))

/*
	firstExpected := []exchangesdk.Trade{
		{
			Price: D(1),
		},
	}

	ctx := context.Background()
	c := luno.NewClientForTesting(t, m, 12, 34)
	first, err := c.GetTrades(ctx, 1)
	require.NoError(t, err)
	second, err := c.GetTrades(ctx, 1)
	require.NoError(t, err)

	assert.Equal(t, 0, len(trades))
*/
}
