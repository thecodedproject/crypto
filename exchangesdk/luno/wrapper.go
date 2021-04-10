package luno

import (
	"context"
	"errors"
	"fmt"
	"github.com/shopspring/decimal"
	"github.com/thecodedproject/crypto"
	"github.com/thecodedproject/crypto/exchangesdk"
	"github.com/thecodedproject/crypto/util"
	"testing"
	"time"
	luno_sdk "github.com/luno/luno-go"
	lunodecimal "github.com/luno/luno-go/decimal"
)

// LunoSdk is the interface presented by the Luno Go SDK.
// It is defined here as a way of mocking the Luno Go SDK.
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
	tradingPair string
	tradesByPage map[int64]tradesAndLastSeq
}

func NewClient(
	id string,
	secret string,
	pair crypto.Pair,
) (*client, error) {

	tradingPair, err := getLunoTradingPair(pair)
	if err != nil {
		return nil, err
	}

  c := luno_sdk.NewClient()
  c.SetAuth(id, secret)

  return &client{
    lunoSdk: c,
		tradingPair: tradingPair,
		tradesByPage: make(map[int64]tradesAndLastSeq),
  }, nil
}

func NewClientForTesting(_ *testing.T, lunoSdk LunoSdk) *client {

	return &client{
		lunoSdk: lunoSdk,
		tradingPair: "TestPair",
		tradesByPage: make(map[int64]tradesAndLastSeq),
	}
}

func getLunoTradingPair(pair crypto.Pair) (string, error) {

	switch pair {
	case crypto.PairBTCEUR:
		return "XBTEUR", nil
	case crypto.PairBTCGBP:
		return "XBTGBP", nil
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

func (l *client) LatestPrice(ctx context.Context) (decimal.Decimal, error) {

  req := luno_sdk.GetTickerRequest{Pair: l.tradingPair}
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
    Pair: l.tradingPair,
    Price: lunoPrice,
    Volume: lunoVolume,
    Type: luno_sdk.OrderType(order.Type),
    PostOnly: true,
  }

  res, err := l.lunoSdk.PostLimitOrder(ctx, &req)
  if err != nil {
    return "", err
  }

  return res.OrderId, nil
}

func (l *client) PostStopLimitOrder(
	ctx context.Context,
	order exchangesdk.StopLimitOrder,
) (string, error) {

	lunoStopPrice, err := lunoFromShopSpringDecimal(order.StopPrice)
	if err != nil {
		return "", err
	}

	lunoLimitPrice, err := lunoFromShopSpringDecimal(order.LimitPrice)
	if err != nil {
		return "", err
	}

	lunoVolume, err := lunoFromShopSpringDecimal(order.Volume)
	if err != nil {
		return "", err
	}

	orderType, err := orderBookSideToLunoOrderType(order.Side)
	if err != nil {
		return "", err
	}

  req := luno_sdk.PostLimitOrderRequest{
    Pair: l.tradingPair,
		StopDirection: "RELATIVE_LAST_TRADE",
		StopPrice: lunoStopPrice,
    Price: lunoLimitPrice,
    Volume: lunoVolume,
    Type: orderType,
  }

  res, err := l.lunoSdk.PostLimitOrder(ctx, &req)
  if err != nil {
    return "", err
  }

  return res.OrderId, nil

}

func (l *client) CancelOrder(ctx context.Context, orderId string) error {

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

func (l *client) MakerFee() decimal.Decimal {

	return decimal.Decimal{}
}

func (l *client) CounterPrecision() int32 {

	return 2
}

func (l *client) BasePrecision() int32 {

	return 4
}

func (l *client) GetOrderStatus(ctx context.Context, orderId string) (exchangesdk.OrderStatus, error) {

  req := luno_sdk.GetOrderRequest{
    Id: orderId,
  }

  res, err := l.lunoSdk.GetOrder(ctx, &req)
  if err != nil {
    return exchangesdk.OrderStatus{}, err
  }

	fillAmountBase, err := lunoToShopSpringDecimal(res.Base)
	if err != nil {
    return exchangesdk.OrderStatus{}, err
	}

	state := exchangesdk.OrderStateUnknown
	if res.State == "PENDING" {
		state = exchangesdk.OrderStateInOrderBook
	} else if res.State == "COMPLETED" {
		state = exchangesdk.OrderStateFilled
	} else if res.State == "CANCELLED" {
		state = exchangesdk.OrderStateCancelled
	}

  return exchangesdk.OrderStatus{
    State: state,
    Type: exchangesdk.OrderType(res.Type),
		FillAmountBase: fillAmountBase,
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
		Pair: l.tradingPair,
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

func orderBookSideToLunoOrderType(
	side exchangesdk.OrderBookSide,
) (luno_sdk.OrderType, error) {

	switch side {
	case exchangesdk.OrderBookSideBid:
		return luno_sdk.OrderTypeBuy, nil
	case exchangesdk.OrderBookSideAsk:
		return luno_sdk.OrderTypeAsk, nil
	default:
		return luno_sdk.OrderTypeBid, fmt.Errorf(
			"dont know how to convert OrderBookSide `%s` to luno_sdk.OrderType",
			side,
		)
	}
}

func lunoOrderTypeToOrderBookSide(
	orderType luno_sdk.OrderType,
) (exchangesdk.OrderBookSide, error) {

	switch orderType {
	case luno_sdk.OrderTypeAsk:
		return exchangesdk.OrderBookSideAsk, nil
	case luno_sdk.OrderTypeBid:
		return exchangesdk.OrderBookSideBid, nil
	default:
		return exchangesdk.OrderBookSideUnknown, fmt.Errorf(
			"dont know how to convert luno_sdk.OrderType `%s` to OrderBookSide",
			orderType,
		)
	}
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
