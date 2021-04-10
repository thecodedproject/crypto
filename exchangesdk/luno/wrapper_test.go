package luno_test

import (
	"context"
	"strconv"
	"testing"

	luno_sdk "github.com/luno/luno-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/thecodedproject/crypto/exchangesdk"
	"github.com/thecodedproject/crypto/exchangesdk/luno"
)

func makeSomeLunoTrades(n int64, offset int64) []luno_sdk.Trade {

	trades := make([]luno_sdk.Trade, 0, n)
	for i := offset; i < (n + offset); i++ {
		trades = append(trades, luno_sdk.Trade{
			OrderId:  strconv.FormatInt(i, 10),
			Sequence: i,
		})
	}
	return trades
}

func makeSomeTrades(n int64) []exchangesdk.Trade {

	trades := make([]exchangesdk.Trade, 0, n)
	for i := int64(0); i < n; i++ {
		trades = append(trades, exchangesdk.Trade{
			OrderId: strconv.FormatInt(i, 10),
		})
	}
	return trades
}

func TestGetTradesForPageLessThanOneReqturnsError(t *testing.T) {

	m := new(luno.MockLunoSdk)
	ctx := context.Background()
	c := luno.NewClientForTesting(t, m)
	_, err := c.GetTrades(ctx, 0)
	assert.Error(t, err)
}

func TestGetTradesWhenThereAreNoTrades(t *testing.T) {

	m := new(luno.MockLunoSdk)

	req := &luno_sdk.ListUserTradesRequest{
		Pair: "TestPair",
	}
	var res luno_sdk.ListUserTradesResponse
	m.On("ListUserTrades", mock.Anything, req).Return(&res, nil)

	ctx := context.Background()
	c := luno.NewClientForTesting(t, m)
	trades, err := c.GetTrades(ctx, 1)
	require.NoError(t, err)

	assert.Equal(t, 0, len(trades))
}

func TestGetTradesFirstPageWhenThereAreSomeTrades(t *testing.T) {

	m := new(luno.MockLunoSdk)

	req := &luno_sdk.ListUserTradesRequest{
		Pair: "TestPair",
	}
	res := luno_sdk.ListUserTradesResponse{
		Trades: []luno_sdk.Trade{
			{
				OrderId:  "1",
				Sequence: 1,
			},
			{
				OrderId:  "2",
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
	c := luno.NewClientForTesting(t, m)
	trades, err := c.GetTrades(ctx, 1)
	require.NoError(t, err)

	assert.Equal(t, expected, trades)
}

func TestGetTradesFirstPageMultipleTimesWhenFullOfTradesMultipleTimesOnlyCallsListTradesOnce(t *testing.T) {

	m := new(luno.MockLunoSdk)

	req := &luno_sdk.ListUserTradesRequest{
		Pair: "TestPair",
	}
	res := luno_sdk.ListUserTradesResponse{
		Trades: makeSomeLunoTrades(100, 0),
	}
	m.On("ListUserTrades", mock.Anything, req).Return(&res, nil).Once()

	expected := makeSomeTrades(100)

	ctx := context.Background()
	c := luno.NewClientForTesting(t, m)
	trades, err := c.GetTrades(ctx, 1)
	require.NoError(t, err)
	assert.Equal(t, expected, trades)

	trades, err = c.GetTrades(ctx, 1)
	require.NoError(t, err)
	assert.Equal(t, expected, trades)
}

func TestGetTradesSecondPageWhenFirstPageIsFull(t *testing.T) {

	m := new(luno.MockLunoSdk)

	firstReq := luno_sdk.ListUserTradesRequest{
		Pair: "TestPair",
	}
	firstRes := luno_sdk.ListUserTradesResponse{
		Trades: makeSomeLunoTrades(100, 0),
	}
	m.On("ListUserTrades", mock.Anything, &firstReq).Return(&firstRes, nil)

	secondReq := luno_sdk.ListUserTradesRequest{
		Pair:     "TestPair",
		AfterSeq: 99,
	}
	secondRes := luno_sdk.ListUserTradesResponse{
		Trades: []luno_sdk.Trade{
			{
				OrderId:  "201",
				Sequence: 201,
			},
		},
	}
	m.On("ListUserTrades", mock.Anything, &secondReq).Return(&secondRes, nil)

	expected := []exchangesdk.Trade{
		{
			OrderId: "201",
		},
	}

	ctx := context.Background()
	c := luno.NewClientForTesting(t, m)
	trades, err := c.GetTrades(ctx, 2)
	require.NoError(t, err)

	assert.Equal(t, expected, trades)
}

func TestRepeatedlyGetTradesForSecondPageWhenFirstPageIsFullRequestsFirstPageOnce(t *testing.T) {

	m := new(luno.MockLunoSdk)

	firstReq := luno_sdk.ListUserTradesRequest{
		Pair: "TestPair",
	}
	firstRes := luno_sdk.ListUserTradesResponse{
		Trades: makeSomeLunoTrades(100, 0),
	}
	m.On("ListUserTrades", mock.Anything, &firstReq).Return(&firstRes, nil).Once()

	secondReq := luno_sdk.ListUserTradesRequest{
		Pair:     "TestPair",
		AfterSeq: 99,
	}
	secondRes := luno_sdk.ListUserTradesResponse{
		Trades: []luno_sdk.Trade{
			{
				OrderId:  "201",
				Sequence: 201,
			},
		},
	}
	m.On("ListUserTrades", mock.Anything, &secondReq).Return(&secondRes, nil).Times(3)

	expected := []exchangesdk.Trade{
		{
			OrderId: "201",
		},
	}

	ctx := context.Background()
	c := luno.NewClientForTesting(t, m)

	trades, err := c.GetTrades(ctx, 2)
	require.NoError(t, err)
	assert.Equal(t, expected, trades)

	trades, err = c.GetTrades(ctx, 2)
	require.NoError(t, err)
	assert.Equal(t, expected, trades)

	trades, err = c.GetTrades(ctx, 2)
	require.NoError(t, err)
	assert.Equal(t, expected, trades)
}

func TestRepeatedlyGetTradesForThridPageWhenSecondANdFirstPageIsFullRequestsThosePagesOnce(t *testing.T) {

	m := new(luno.MockLunoSdk)

	firstReq := luno_sdk.ListUserTradesRequest{
		Pair: "TestPair",
	}
	firstRes := luno_sdk.ListUserTradesResponse{
		Trades: makeSomeLunoTrades(100, 0),
	}
	m.On("ListUserTrades", mock.Anything, &firstReq).Return(&firstRes, nil).Once()

	secondReq := luno_sdk.ListUserTradesRequest{
		Pair:     "TestPair",
		AfterSeq: 99,
	}
	secondRes := luno_sdk.ListUserTradesResponse{
		Trades: makeSomeLunoTrades(100, 100),
	}
	m.On("ListUserTrades", mock.Anything, &secondReq).Return(&secondRes, nil).Once()

	thirdReq := luno_sdk.ListUserTradesRequest{
		Pair:     "TestPair",
		AfterSeq: 199,
	}
	thirdRes := luno_sdk.ListUserTradesResponse{
		Trades: []luno_sdk.Trade{
			{
				OrderId:  "301",
				Sequence: 301,
			},
		},
	}
	m.On("ListUserTrades", mock.Anything, &thirdReq).Return(&thirdRes, nil).Times(3)

	expected := []exchangesdk.Trade{
		{
			OrderId: "301",
		},
	}

	ctx := context.Background()
	c := luno.NewClientForTesting(t, m)

	trades, err := c.GetTrades(ctx, 3)
	require.NoError(t, err)
	assert.Equal(t, expected, trades)

	trades, err = c.GetTrades(ctx, 3)
	require.NoError(t, err)
	assert.Equal(t, expected, trades)

	trades, err = c.GetTrades(ctx, 3)
	require.NoError(t, err)
	assert.Equal(t, expected, trades)
}

func TestGetTradesThirdPageWhenSecondPageIsNotFullReturnsEmpty(t *testing.T) {

	m := new(luno.MockLunoSdk)

	firstReq := luno_sdk.ListUserTradesRequest{
		Pair: "TestPair",
	}
	firstRes := luno_sdk.ListUserTradesResponse{
		Trades: makeSomeLunoTrades(100, 0),
	}
	m.On("ListUserTrades", mock.Anything, &firstReq).Return(&firstRes, nil).Once()

	secondReq := luno_sdk.ListUserTradesRequest{
		Pair:     "TestPair",
		AfterSeq: 99,
	}
	secondRes := luno_sdk.ListUserTradesResponse{
		Trades: []luno_sdk.Trade{
			{
				OrderId:  "201",
				Sequence: 201,
			},
		},
	}
	m.On("ListUserTrades", mock.Anything, &secondReq).Return(&secondRes, nil).Once()

	ctx := context.Background()
	c := luno.NewClientForTesting(t, m)
	trades, err := c.GetTrades(ctx, 3)
	require.NoError(t, err)

	assert.Equal(t, 0, len(trades))
}

func TestGetTradesForHundredthPageWhenOnlyFirstPageHasTradesOnlyRequestsFirstPageFromAPI(t *testing.T) {

	m := new(luno.MockLunoSdk)

	req := &luno_sdk.ListUserTradesRequest{
		Pair: "TestPair",
	}
	res := luno_sdk.ListUserTradesResponse{
		Trades: []luno_sdk.Trade{
			{
				OrderId:  "1",
				Sequence: 1,
			},
			{
				OrderId:  "2",
				Sequence: 2,
			},
		},
	}
	m.On("ListUserTrades", mock.Anything, req).Return(&res, nil).Once()

	ctx := context.Background()
	c := luno.NewClientForTesting(t, m)
	trades, err := c.GetTrades(ctx, 100)
	require.NoError(t, err)

	assert.Equal(t, 0, len(trades))
}
