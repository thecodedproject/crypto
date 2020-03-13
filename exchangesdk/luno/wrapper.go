package luno

import (
	"context"
	"errors"
	"fmt"
	"github.com/thecodedproject/crypto/exchangesdk"
	"github.com/shopspring/decimal"
	"strconv"
	luno_sdk "github.com/luno/luno-go"
	lunodecimal "github.com/luno/luno-go/decimal"
)

const (
  TRADINGPAIR = "XBTEUR"
  BASE = "XBT"
  COUNTER = "EUR"
)

type lunoClient struct {
  client *luno_sdk.Client
  baseAccountId int64
  counterAccountId int64
}

func NewClient(ctx context.Context, id, secret string) (*lunoClient, error) {
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

  return &lunoClient{
    client: c,
    baseAccountId: baseAccountId,
    counterAccountId: counterAccountId,
  }, nil
}

func (l *lunoClient) LatestPrice(ctx context.Context) (decimal.Decimal, error) {

  req := luno_sdk.GetTickerRequest{Pair: TRADINGPAIR}
  res, err := l.client.GetTicker(ctx, &req)
  if err != nil {
    return decimal.Decimal{}, err
  }

  askPrice := res.Ask
  bidPrice := res.Bid

  midPrice := bidPrice.Add(askPrice.Sub(bidPrice).DivInt64(2))

	return lunoToShopSpringDecimal(midPrice)
}

func (l *lunoClient) PostLimitOrder(ctx context.Context, order exchangesdk.Order) (string, error) {

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

  res, err := l.client.PostLimitOrder(ctx, &req)
  if err != nil {
    return "", err
  }

  return res.OrderId, nil
}

// TODO StopOrder used to return (bool, error) (returing false if stop order failed... it was refactored away to get things running quicker... decide if that was a better interface
func (l *lunoClient) StopOrder(ctx context.Context, orderId string) error {

  req := luno_sdk.StopOrderRequest{
    OrderId: orderId,
  }

  res, err := l.client.StopOrder(ctx, &req)
  if err != nil {
    return nil
  }

	if res.Success == false {
		return errors.New("Failed to stop order")
	}

  return nil
}

func (l *lunoClient) GetOrderStatus(ctx context.Context, orderId string) (exchangesdk.OrderStatus, error) {

  req := luno_sdk.GetOrderRequest{
    Id: orderId,
  }

  res, err := l.client.GetOrder(ctx, &req)
  if err != nil {
    return exchangesdk.OrderStatus{}, err
  }

  return exchangesdk.OrderStatus{
    State: exchangesdk.OrderState(res.State),
    Type: exchangesdk.OrderType(res.Type),
  }, nil
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

	inString := in.String()

	// Redundant check to ensure the value round trips - not sure how else to make sure that this conversion has worked?
	// TODO think of other ways? Could also do a float64 conversion check on the decimal.Decimal val to triple check?
	d, err := decimal.NewFromString(inString)
	if err != nil {
		return decimal.Decimal{}, err
	}

	if inString != d.String() {
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

	if inString != lunoDec.String() {
		return lunodecimal.Decimal{}, errors.New(fmt.Sprintf("Error converting to luno decimal; Input: %s Output: %s", in, lunoDec))
	}

	return lunoDec, nil
}
