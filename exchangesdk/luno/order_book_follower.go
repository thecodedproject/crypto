package luno

import (
	"encoding/json"
	"fmt"
	"github.com/shopspring/decimal"
	"github.com/thecodedproject/crypto/exchangesdk"
	"github.com/gorilla/websocket"
	"log"
	"sync"
	//"time"
)

const (
	wsUrl = "wss://ws.luno.com/api/1/stream/XBTEUR"
)

func NewOrderBookFollower(
	apiKey string,
	apiSecret string,
	pair exchangesdk.Pair,

) (*orderBookFollower, error) {

	if pair != exchangesdk.BTCEUR {
		return nil, fmt.Errorf("Only BTCEUR is supported")
	}

	obf := orderBookFollower{
		minAsk: decimal.NewFromFloat(20000.0),
	}

	followForever(
		&obf,
		apiKey,
		apiSecret,
	)

	return &obf, nil
}

type orderBookFollower struct {

	bids map[string]order
	asks map[string]order

	maxBid decimal.Decimal
	minAsk decimal.Decimal

	sequence int64

	mux sync.RWMutex
}

func (obf *orderBookFollower) MidpointPrice() decimal.Decimal {

	return decimal.Decimal{}
}

func (obf *orderBookFollower) MaxBidPrice() decimal.Decimal {

	obf.mux.RLock()
	defer obf.mux.RUnlock()
	return obf.maxBid
}

func (obf *orderBookFollower) MinAskPrice() decimal.Decimal {

	obf.mux.RLock()
	defer obf.mux.RUnlock()
	return obf.minAsk
}

func (obf *orderBookFollower) addSnapshot(s orderBookSnapshot) {

	bids := convertOrders(s.Bids)
	asks := convertOrders(s.Asks)

	obf.mux.Lock()
	defer obf.mux.Unlock()
	obf.bids = bids
	obf.asks = asks
	obf.sequence = s.Sequence
}

func (obf *orderBookFollower) addUpdate(u orderBookUpdate) error {

	if u.Sequence <= obf.sequence {
		return nil
	}

	if u.Sequence != obf.sequence+1 {
		return fmt.Errorf("Out of sequence orderBookUpdate")
	}

	// TODO calc change and only lock mutex when assigning
	obf.mux.Lock()
	defer obf.mux.Unlock()

	for _, t := range u.TradeUpdates {
		if err := obf.processTrade(*t); err != nil {
			return err
		}
	}

	if u.CreateUpdate != nil {
		if err := obf.processCreate(*u.CreateUpdate); err != nil {
			return err
		}
	}

	if u.DeleteUpdate != nil {
		if err := obf.processDelete(*u.DeleteUpdate); err != nil {
			return err
		}
	}

	obf.sequence = u.Sequence

	for _, bid := range obf.bids {
		if bid.Price.GreaterThan(obf.maxBid) {
			obf.maxBid = bid.Price
		}
	}

	for _, ask := range obf.asks {
		if ask.Price.LessThan(obf.minAsk) {
			obf.minAsk = ask.Price
		}
	}

	return nil
}

func (obf *orderBookFollower) processTrade(t TradeUpdate) error {
	if t.Base.Sign() <= 0 {
		return fmt.Errorf("streaming: nonpositive trade")
	}

	ok, err := decTrade(obf.bids, t.OrderID, t.Base)
	if err != nil {
		return err
	}
	if ok {
		return nil
	}

	ok, err = decTrade(obf.asks, t.OrderID, t.Base)
	if err != nil {
		return err
	}
	if ok {
		return nil
	}

	return fmt.Errorf("streaming: trade for unknown order")
}

func decTrade(m map[string]order, id string, base decimal.Decimal) (
	bool, error) {

	o, ok := m[id]
	if !ok {
		return false, nil
	}

	o.Volume = o.Volume.Add(base.Neg())

	if o.Volume.Sign() < 0 {
		return false, fmt.Errorf("streaming: negative volume: %s", o.Volume)
	}

	if o.Volume.Sign() == 0 {
		delete(m, id)
	} else {
		m[id] = o
	}
	return true, nil
}

func (obf *orderBookFollower) processCreate(u CreateUpdate) error {
	o := order{
		ID:     u.OrderID,
		Price:  u.Price,
		Volume: u.Volume,
	}

	if u.Type == string(exchangesdk.OrderTypeBid) {
		obf.bids[o.ID] = o
	} else if u.Type == string(exchangesdk.OrderTypeAsk) {
		obf.asks[o.ID] = o
	} else {
		return fmt.Errorf("streaming: unknown order type")
	}

	return nil
}

func (obf *orderBookFollower) processDelete(u DeleteUpdate) error {
	delete(obf.bids, u.OrderID)
	delete(obf.asks, u.OrderID)
	return nil
}

type order struct {
	ID     string          `json:"id"`
	Price  decimal.Decimal `json:"price,string"`
	Volume decimal.Decimal `json:"volume,string"`
}

type orderBookSnapshot struct {
	Sequence int64       `json:"sequence,string"`
	Asks     []*order    `json:"asks"`
	Bids     []*order    `json:"bids"`
	//Status   luno.Status `json:"status"`
}

type TradeUpdate struct {
	Base    decimal.Decimal `json:"base,string"`
	Counter decimal.Decimal `json:"counter,string"`
	OrderID string          `json:"order_id"`
}

type CreateUpdate struct {
	OrderID string          `json:"order_id"`
	Type    string          `json:"type"`
	Price   decimal.Decimal `json:"price,string"`
	Volume  decimal.Decimal `json:"volume,string"`
}

type DeleteUpdate struct {
	OrderID string `json:"order_id"`
}

type StatusUpdate struct {
	Status string `json:"status"`
}

type orderBookUpdate struct {
	Sequence     int64          `json:"sequence,string"`
	TradeUpdates []*TradeUpdate `json:"trade_updates"`
	CreateUpdate *CreateUpdate  `json:"create_update"`
	DeleteUpdate *DeleteUpdate  `json:"delete_update"`
	StatusUpdate *StatusUpdate  `json:"status_update"`
	Timestamp    int64          `json:"timestamp"`
}

func followForever(
	obf *orderBookFollower,
	apiKey string,
	apiSecret string,
) {


	log.Println("Running obf")

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

		snapshot := orderBookSnapshot{}
		if err := json.Unmarshal(msg, &snapshot); err != nil {
			log.Fatal(err)
		}
		obf.addSnapshot(snapshot)

		for {

			_, msg, err := ws.ReadMessage()
			if err != nil {
				log.Fatal("ReadMessage error:", err)
			}

			update := orderBookUpdate{}
			if err := json.Unmarshal(msg, &update); err != nil {
				log.Fatal(err, string(msg))
			}

			obf.addUpdate(update)

		}

	}()
}

func convertOrders(ol []*order) map[string]order {
	r := make(map[string]order)
	for _, o := range ol {
		r[o.ID] = *o
	}
	return r
}
