package luno

import (
	"context"
	"errors"
	"fmt"
	"github.com/shopspring/decimal"
	"github.com/thecodedproject/crypto/exchangesdk"
	"github.com/thecodedproject/crypto/util"
	"strconv"
	"testing"
	"time"
	luno_sdk "github.com/luno/luno-go"
	lunodecimal "github.com/luno/luno-go/decimal"
)

const (
  TRADINGPAIR = "XBTEUR"
  BASE = "XBT"
  COUNTER = "EUR"
)

type LunoSdk interface {
	GetTicker(ctx context.Context, req *luno_sdk.GetTickerRequest) (*luno_sdk.GetTickerResponse, error)
	PostLimitOrder(ctx context.Context, req *luno_sdk.PostLimitOrderRequest) (*luno_sdk.PostLimitOrderResponse, error)
	StopOrder(ctx context.Context, req *luno_sdk.StopOrderRequest) (*luno_sdk.StopOrderResponse, error)
	GetOrder(ctx context.Context, req *luno_sdk.GetOrderRequest) (*luno_sdk.GetOrderResponse, error)
	ListUserTrades(ctx context.Context, req *luno_sdk.ListUserTradesRequest) (*luno_sdk.ListUserTradesResponse, error)
}

type tradesAndLastSeq struct {
	Trades []exchangesdk.Trade
	SequenceOfLastTrade int64
}

type client struct {
  lunoSdk LunoSdk
  baseAccountId int64
  counterAccountId int64
	tradesByPage map[int64]tradesAndLastSeq
}

func NewClient(ctx context.Context, id, secret string) (*client, error) {
  c := luno_sdk.NewClient()
  c.SetAuth(id, secret)

  accountIds, err := getAccountIds(ctx, c)
  if err != nil {
    return nil, err
  }

  baseAccountId, ok := accountIds[BASE]
  if !ok {
    return nil, errors.New("No base accountId found")
  }
  counterAccountId, ok := accountIds[COUNTER]
  if !ok {
    return nil, errors.New("No counter accountId found")
  }

  return &client{
    lunoSdk: c,
    baseAccountId: baseAccountId,
    counterAccountId: counterAccountId,
		tradesByPage: make(map[int64]tradesAndLastSeq),
  }, nil
}

func NewClientForTesting(_ *testing.T, lunoSdk LunoSdk, baseAccountId, counterAccountId int64) *client {

	return &client{
		lunoSdk: lunoSdk,
		baseAccountId: baseAccountId,
		counterAccountId: counterAccountId,
		tradesByPage: make(map[int64]tradesAndLastSeq),
	}
}

func (l *client) LatestPrice(ctx context.Context) (decimal.Decimal, error) {

  req := luno_sdk.GetTickerRequest{Pair: TRADINGPAIR}
  res, err := l.lunoSdk.GetTicker(ctx, &req)
  if err != nil {
    return decimal.Decimal{}, err
  }

  askPrice := res.Ask
  bidPrice := res.Bid

  midPrice := bidPrice.Add(askPrice.Sub(bidPrice).DivInt64(2))

	return lunoToShopSpringDecimal(midPrice)
}

func (l *client) PostLimitOrder(ctx context.Context, order exchangesdk.Order) (string, error) {

	lunoPrice, err := lunoFromShopSpringDecimal(order.Price)
	if err != nil {
		return "", err
	}
	lunoVolume, err := lunoFromShopSpringDecimal(order.Volume)
	if err != nil {
		return "", err
	}

  req := luno_sdk.PostLimitOrderRequest{
    Pair: TRADINGPAIR,
    Price: lunoPrice,
    Volume: lunoVolume,
    Type: luno_sdk.OrderType(order.Type),
    BaseAccountId: l.baseAccountId,
    CounterAccountId: l.counterAccountId,
    PostOnly: true,
  }

  res, err := l.lunoSdk.PostLimitOrder(ctx, &req)
  if err != nil {
    return "", err
  }

  return res.OrderId, nil
}

// TODO StopOrder used to return (bool, error) (returing false if stop order failed... it was refactored away to get things running quicker... decide if that was a better interface
func (l *client) StopOrder(ctx context.Context, orderId string) error {

  req := luno_sdk.StopOrderRequest{
    OrderId: orderId,
  }

  res, err := l.lunoSdk.StopOrder(ctx, &req)
  if err != nil {
    return nil
  }

	if res.Success == false {
		return errors.New("Failed to stop order")
	}

  return nil
}

func (l *client) GetOrderStatus(ctx context.Context, orderId string) (exchangesdk.OrderStatus, error) {

  req := luno_sdk.GetOrderRequest{
    Id: orderId,
  }

  res, err := l.lunoSdk.GetOrder(ctx, &req)
  if err != nil {
    return exchangesdk.OrderStatus{}, err
  }

  return exchangesdk.OrderStatus{
    State: exchangesdk.OrderState(res.State),
    Type: exchangesdk.OrderType(res.Type),
  }, nil
}

func (l* client) GetTrades(ctx context.Context, page int64) ([]exchangesdk.Trade, error) {

	if page < 1 {
		return nil, errors.New(
			fmt.Sprintf("Cannot get page less than 1; trying to get page %d", page))
	}

	t, ok := l.tradesByPage[page]
	if ok {
		return t.Trades, nil
	}

	req := luno_sdk.ListUserTradesRequest{
		Pair: TRADINGPAIR,
	}

	if page > 1 {

		_, err := l.GetTrades(ctx, page-1)
		if err != nil {
			return nil, err
		}

		previousPageTrades, ok := l.tradesByPage[page-1]
		if !ok {
			return []exchangesdk.Trade{}, nil
		}

		req.AfterSeq = previousPageTrades.SequenceOfLastTrade
	}

	res, err := l.lunoSdk.ListUserTrades(ctx, &req)
	if err != nil {
		return nil, err
	}

	trades, err := convertLunoTrades(res.Trades)
	if err != nil {
		return nil, err
	}

	if len(trades) == 100 {
		l.tradesByPage[page] = tradesAndLastSeq{
			Trades: trades,
			SequenceOfLastTrade: res.Trades[99].Sequence,
		}
	}

	return trades, nil
}

func convertLunoTrades(lunoTrades []luno_sdk.Trade) ([]exchangesdk.Trade, error) {

	trades := make([]exchangesdk.Trade, 0, len(lunoTrades))
	for _, lunoTrade := range lunoTrades {

		price, err := lunoToShopSpringDecimal(lunoTrade.Price)
		if err != nil {
			return nil, err
		}
		volume, err := lunoToShopSpringDecimal(lunoTrade.Volume)
		if err != nil {
			return nil, err
		}
		baseFee, err := lunoToShopSpringDecimal(lunoTrade.FeeBase)
		if err != nil {
			return nil, err
		}
		counterFee, err := lunoToShopSpringDecimal(lunoTrade.FeeCounter)
		if err != nil {
			return nil, err
		}

		trades = append(trades, exchangesdk.Trade{
			OrderId: lunoTrade.OrderId,
			Timestamp: time.Time(lunoTrade.Timestamp),
			Price: price,
			Volume: volume,
			BaseFee: baseFee,
			CounterFee: counterFee,
			Type: exchangesdk.OrderType(lunoTrade.Type),
		})
	}

	return trades, nil
}

func getAccountIds(ctx context.Context, c *luno_sdk.Client) (map[string]int64, error) {

  res, err := c.GetBalances(ctx, &luno_sdk.GetBalancesRequest{})
  if err != nil {
    return nil, err
  }

  ids := make(map[string]int64)
  for _, account := range res.Balance {
    id, err := strconv.ParseInt(account.AccountId, 10, 64)
    if err != nil {
      return nil, err
    }
    ids[account.Asset] = id
  }

  return ids, nil
}

// TODO add tests for this function around edge cases
func lunoToShopSpringDecimal(in lunodecimal.Decimal) (decimal.Decimal, error) {

	var emptyLunoDec lunodecimal.Decimal
	if in == emptyLunoDec {
		return decimal.Decimal{}, nil
	}

	inString := in.String()

	// Redundant check to ensure the value round trips - not sure how else to make sure that this conversion has worked?
	// TODO think of other ways? Could also do a float64 conversion check on the decimal.Decimal val to triple check?
	d, err := decimal.NewFromString(inString)
	if err != nil {
		return decimal.Decimal{}, err
	}

	dF64, _ := d.Float64()

	if !util.Float64Near(in.Float64(), dF64) {
		return decimal.Decimal{}, errors.New(fmt.Sprintf("Error converting from luno decimal; Input: %s Output: %s", in, d))
	}

	return d, nil
}

func lunoFromShopSpringDecimal(in decimal.Decimal) (lunodecimal.Decimal, error) {

	inString := in.String()

	// TODO same here for adding triple check on conversion
	lunoDec, err := lunodecimal.NewFromString(inString)
	if err != nil {
		return lunodecimal.Decimal{}, err
	}

	inF64, _ := in.Float64()

	if !util.Float64Near(inF64, lunoDec.Float64()) {
		return lunodecimal.Decimal{}, errors.New(fmt.Sprintf("Error converting to luno decimal; Input: %s Output: %s", in, lunoDec))
	}

	return lunoDec, nil
}
