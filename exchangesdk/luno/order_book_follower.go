package luno

import (
	"encoding/json"
	"errors"
	"fmt"
	"math"
	//"github.com/shopspring/decimal"
	"github.com/thecodedproject/crypto/exchangesdk"
	"github.com/gorilla/websocket"
	"log"
	"time"
)

const (
	wsUrl = "wss://ws.luno.com/api/1/stream/XBTEUR"

	// TODO: Set these in a more robust way
	MARKET_PRICE_PRECISION = 0.01
	MARKET_VOLUME_PRECISION = 1e-8
)

type InternalOrderBook struct {
	Bids map[string]Order
	Asks map[string]Order

	LastSequenceId int64
	LastUpdateTimestamp int64
}

type Order struct {
	Id string `json:"id"`
	Price float64 `json:"price,string"`
	Volume float64 `json:"volume,string"`
}

type OrderBookSnapshot struct {
	Sequence int64 `json:"sequence,string"`
	Asks []Order `json:"asks"`
	Bids []Order `json:"bids"`
	Timestamp int64 `json:"timestamp"`
}

type TradeUpdate struct {
	Base float64 `json:"base,string"`
	Counter float64 `json:"counter,string"`
	MakerOrderId string `json:"maker_order_id"`
	//TakerOrderId string `json:"taker_order_id"`
}

type CreateUpdate struct {
	OrderId string `json:"order_id"`
	OrderType string `json:"type"`
	Price float64 `json:"price,string"`
	Volume float64 `json:"volume,string"`
}

type DeleteUpdate struct {
	OrderId string `json:"order_id"`
}

type StatusUpdate struct {
	Status string `json:"status"`
}

type OrderBookUpdate struct {
	Sequence int64 `json:"sequence,string"`
	TradeUpdates []*TradeUpdate `json:"trade_updates"`
	CreateUpdate *CreateUpdate `json:"create_update"`
	DeleteUpdate *DeleteUpdate `json:"delete_update"`
	StatusUpdate *StatusUpdate `json:"status_update"`
	Timestamp int64 `json:"timestamp"`
}


func NewOrderBookFollowerAndTradeStream(
	pair exchangesdk.Pair,
	apiKey string,
	apiSecret string,
) (<-chan exchangesdk.OrderBook, <-chan exchangesdk.OrderBookTrade) {

	if pair != exchangesdk.BTCEUR {
		log.Fatal("Only BTCEUR is supported")
	}

	return followForever(
		apiKey,
		apiSecret,
	)
}

func followForever(
	apiKey string,
	apiSecret string,
) (<-chan exchangesdk.OrderBook, <-chan exchangesdk.OrderBookTrade) {

	log.Println("Running obf")

	obf := make(chan exchangesdk.OrderBook, 1)
	tradeStream := make(chan exchangesdk.OrderBookTrade, 1)
	var ob InternalOrderBook

	go func() {

		ws, _, err := websocket.DefaultDialer.Dial(wsUrl, nil)
		if err != nil {
			log.Fatal(err)
		}
		defer ws.Close()

		creds := struct{
			Key string `json:"api_key_id"`
			Secret string `json:"api_key_secret"`
		}{
			Key: apiKey,
			Secret: apiSecret,
		}

		if err := ws.WriteJSON(creds); err != nil {
			log.Fatal(err)
		}

		_, msg, err := ws.ReadMessage()
		if err != nil {
			log.Fatal("ReadMessage error:", err)
		}

		snapshot := OrderBookSnapshot{}
		if err := json.Unmarshal(msg, &snapshot); err != nil {
			log.Fatal(err)
		}
		handleSnapshot(&ob, snapshot)
		obf <- *toSortedOrderBook(&ob)

		for {

			_, msg, err := ws.ReadMessage()
			if err != nil {
				log.Fatal("ReadMessage error:", err)
			}

			if string(msg) == "" {
				continue
			}

			update := OrderBookUpdate{}
			if err := json.Unmarshal(msg, &update); err != nil {
				log.Fatal(err, string(msg))
			}

			obUpdated, err := HandleUpdate(&ob, update)
			if err != nil {
				log.Println("OrderBookFollower error:", err)
				close(obf)
				return
			}
			if obUpdated {
				obf <- *toSortedOrderBook(&ob)
			}

			for _, tradeUpdate := range update.TradeUpdates {
				t, err := convertToSdkTrade(&ob, tradeUpdate, update.Timestamp)
				if err != nil {
					log.Println("TradeStream error:", err)
					close(tradeStream)
					return
				}
				tradeStream <- t
			}
		}
	}()

	return obf, tradeStream
}


func handleSnapshot(ob *InternalOrderBook, s OrderBookSnapshot) {

	ob.Bids = convertSnapshotOrders(s.Bids)
	ob.Asks = convertSnapshotOrders(s.Asks)
	ob.LastSequenceId = s.Sequence
	ob.LastUpdateTimestamp = s.Timestamp
}

func convertSnapshotOrders(ol []Order) map[string]Order {

	r := make(map[string]Order)
	for _, o := range ol {
		r[o.Id] = o
	}
	return r
}

func toSortedOrderBook(ob *InternalOrderBook) *exchangesdk.OrderBook {

	var o exchangesdk.OrderBook
	o.Bids = make([]exchangesdk.OrderBookOrder, len(ob.Bids))
	o.Asks = make([]exchangesdk.OrderBookOrder, len(ob.Asks))

	iBid := 0
	for _, bid := range ob.Bids {
		o.Bids[iBid].Price = bid.Price
		o.Bids[iBid].Volume = bid.Volume
		iBid++
	}

	iAsk := 0
	for _, ask := range ob.Asks {
		o.Asks[iAsk].Price = ask.Price
		o.Asks[iAsk].Volume = ask.Volume
		iAsk++
	}

	exchangesdk.SortOrderBook(&o)

	o.Timestamp = time.Unix(0, ob.LastUpdateTimestamp * int64(time.Millisecond))

	return &o
}

func HandleUpdate(ob *InternalOrderBook, u OrderBookUpdate) (bool, error) {

	var updated bool

	if u.Sequence <= ob.LastSequenceId {
		return updated, nil
	}

	if u.Sequence != ob.LastSequenceId+1 {
		return updated, fmt.Errorf("Out of sequence OrderBookUpdate")
	}

	for _, t := range u.TradeUpdates {
		updated = true
		if err := HandleTrade(ob, t); err != nil {
			return updated, err
		}
	}

	if u.CreateUpdate != nil {
		updated = true
		if err := HandleCreate(ob, u.CreateUpdate); err != nil {
			return updated, err
		}
	}

	if u.DeleteUpdate != nil {
		updated = true
		if err := HandleDelete(ob, u.DeleteUpdate); err != nil {
			return updated, err
		}
	}

	ob.LastSequenceId = u.Sequence
	ob.LastUpdateTimestamp = u.Timestamp

	return updated, nil
}

func HandleTrade(ob *InternalOrderBook, t *TradeUpdate) error {
	if t.Base < 0 {
		return fmt.Errorf("negative trade base")
	}

	orderUpdated, err := updateOrdersWithTrade(ob.Bids, t.MakerOrderId, t.Base)
	if err != nil {
		return err
	}
	if orderUpdated {
		return nil
	}

	orderUpdated, err = updateOrdersWithTrade(ob.Asks, t.MakerOrderId, t.Base)
	if err != nil {
		return err
	}
	if orderUpdated {
		return nil
	}

	return fmt.Errorf("received trade for unknown Order id `%s`", t.MakerOrderId)
}

func updateOrdersWithTrade(
	m map[string]Order,
	id string,
	tradeVolume float64,
) (bool, error) {

	o, ok := m[id]
	if !ok {
		return false, nil
	}

	o.Volume -= tradeVolume

	if o.Volume < 0 {
		return false, fmt.Errorf(
			"recieved trade which would make Order volume negative (%f)",
			o.Volume,
		)
	}

	if hasZeroVolume(o) {
		delete(m, id)
	} else {
		m[id] = o
	}
	return true, nil
}

func HandleCreate(ob *InternalOrderBook, u *CreateUpdate) error {
	o := Order{
		Id:     u.OrderId,
		Price:  u.Price,
		Volume: u.Volume,
	}

	// TODO Remove hardcoding of type strings here
	if u.OrderType == "BID" {
		ob.Bids[o.Id] = o
	} else if u.OrderType == "ASK" {
		ob.Asks[o.Id] = o
	} else {
		return fmt.Errorf("streaming: unknown Order type")
	}

	return nil
}

func HandleDelete(ob *InternalOrderBook, u *DeleteUpdate) error {

	delete(ob.Bids, u.OrderId)
	delete(ob.Asks, u.OrderId)
	return nil
}

func convertToSdkTrade(
	ob *InternalOrderBook,
	t *TradeUpdate,
	timestamp int64,
) (exchangesdk.OrderBookTrade, error) {

	var makerSide exchangesdk.MarketSide

	var isBid bool
	var isAsk bool
	_, isBid = ob.Bids[t.MakerOrderId]
	_, isAsk = ob.Asks[t.MakerOrderId]
	if isBid {
		makerSide = exchangesdk.MarketSideBuy
	} else if isAsk {
		makerSide = exchangesdk.MarketSideBuy
	} else {
		return exchangesdk.OrderBookTrade{}, errors.New("received trade with unknown order id")
	}

	return exchangesdk.OrderBookTrade{
		MakerSide: makerSide,
		Price: t.Counter,
		Volume: t.Base,
		Timestamp: time.Unix(0, timestamp * int64(time.Millisecond)),
	}, nil
}

func hasZeroVolume(o Order) bool {

	return math.Abs(o.Volume) < (MARKET_VOLUME_PRECISION/float64(2))
}
