package binance

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/thecodedproject/crypto"
	"github.com/thecodedproject/crypto/exchangesdk"
	"github.com/thecodedproject/crypto/exchangesdk/requestutil"
)

const (
	WEBSOCKET_LIFETIME = 55 * time.Minute
)

type ExchangeConfig struct {
	OrderBookStream string
	TradesStream    string
	PairCode        string
	PricePrecision  float64
	VolPrecision    float64
}

type internalOrderBook struct {
	exchangesdk.OrderBook
	lastUpdateId int64
}

func NewMarketFollower(
	ctx context.Context,
	wg *sync.WaitGroup,
	pair crypto.Pair,
) (<-chan exchangesdk.OrderBook, <-chan exchangesdk.OrderBookTrade, error) {

	exConf, err := getExchangeConfig(pair)
	if err != nil {
		return nil, nil, err
	}

	return followForever(
		ctx,
		wg,
		exConf,
	)
}

func getExchangeConfig(pair crypto.Pair) (ExchangeConfig, error) {

	switch pair {
	case crypto.PairBTCEUR:
		return ExchangeConfig{
			OrderBookStream: "btceur@depth",
			TradesStream:    "btceur@trade",
			PairCode:        "BTCEUR",
			PricePrecision:  1e-2,
			VolPrecision:    1e-8,
		}, nil
	case crypto.PairBTCGBP:
		return ExchangeConfig{
			OrderBookStream: "btcgbp@depth",
			TradesStream:    "btcgbp@trade",
			PairCode:        "BTCGBP",
			PricePrecision:  1e-2,
			VolPrecision:    1e-8,
		}, nil
	case crypto.PairBTCUSDT:
		return ExchangeConfig{
			OrderBookStream: "btcusdt@depth",
			TradesStream:    "btcusdt@trade",
			PairCode:        "BTCUSDT",
			PricePrecision:  1e-2,
			VolPrecision:    1e-8,
		}, nil
	case crypto.PairLTCBTC:
		return ExchangeConfig{
			OrderBookStream: "ltcbtc@depth",
			TradesStream:    "ltcbtc@trade",
			PairCode:        "LTCBTC",
			PricePrecision:  1e-6,
			VolPrecision:    1e-2,
		}, nil
	case crypto.PairETHBTC:
		return ExchangeConfig{
			OrderBookStream: "ethbtc@depth",
			TradesStream:    "ethbtc@trade",
			PairCode:        "ETHBTC",
			PricePrecision:  1e-6,
			VolPrecision:    1e-3,
		}, nil
	case crypto.PairBCHBTC:
		return ExchangeConfig{
			OrderBookStream: "bchbtc@depth",
			TradesStream:    "bchbtc@trade",
			PairCode:        "BCHBTC",
			PricePrecision:  1e-6,
			VolPrecision:    1e-3,
		}, nil
	default:
		return ExchangeConfig{}, fmt.Errorf("%s pair is not support by Binance market follower", pair)
	}
}

func buildWsUrl(exConf ExchangeConfig) string {

	wsUrl := fmt.Sprintf(
		"wss://stream.binance.com:9443/stream?streams=%s/%s",
		exConf.OrderBookStream,
		exConf.TradesStream,
	)
	return wsUrl
}

func followForever(
	ctx context.Context,
	wg *sync.WaitGroup,
	exConf ExchangeConfig,
) (<-chan exchangesdk.OrderBook, <-chan exchangesdk.OrderBookTrade, error) {

	obf := make(chan exchangesdk.OrderBook, 1)
	tradeStream := make(chan exchangesdk.OrderBookTrade, 1)
	var ws *websocket.Conn
	var nextWs *websocket.Conn
	wsAge := time.Time{}
	nextWsAge := time.Time{}
	wsUrl := buildWsUrl(exConf)

	go func() {

		var err error
		ws, wsAge, err = newWebsocket(wsUrl)
		if err != nil {
			log.Println("OrderBookFollower error:", err)
			close(obf)
			wg.Done()
			return
		}
		defer ws.Close()

		ob, err := getLatestSnapshot(exConf.PairCode)
		if err != nil {
			log.Println("OrderBookFollower error:", err)
			close(obf)
			wg.Done()
			return
		}

		for {
			if nextWs == nil && time.Since(wsAge) > WEBSOCKET_LIFETIME {
				nextWs, nextWsAge, err = newWebsocket(wsUrl)
				if err != nil {
					log.Println("OrderBookFollower error:", err)
					close(obf)
					wg.Done()
					return
				}
			}

			_, msg, err := ws.ReadMessage()
			if err != nil {
				log.Println("OrderBookFollower error:", err)
				close(obf)
				wg.Done()
				return
			}

			update := struct {
				Stream string          `json:"stream"`
				Data   json.RawMessage `json:"data"`
			}{}

			err = json.Unmarshal(msg, &update)
			if err != nil {
				log.Println("OrderBookFollower error:", err, string(msg))
				close(obf)
				wg.Done()
				return
			}

			switch update.Stream {
			case exConf.OrderBookStream:
				err := handleOrderBookUpdate(&ob, update.Data, exConf)
				if err != nil {
					log.Println("OrderBookFollower error:", err)
					close(obf)
					wg.Done()
					return
				}

				obf <- ob.OrderBook
			case exConf.TradesStream:
				trade, err := decodeTrade(update.Data)
				if err != nil {
					log.Println("OrderBookFollower error:", err)
					close(tradeStream)
					wg.Done()
					return
				}
				tradeStream <- trade
			}

			if nextWs != nil && time.Since(nextWsAge) > time.Second {
				ws.Close()
				ws = nextWs
				nextWs = nil
				wsAge = nextWsAge
			}

			select {
			case <-ctx.Done():
				wg.Done()
				return
			default:
				continue
			}
		}
	}()

	return obf, tradeStream, nil
}

func getLatestSnapshot(pairCode string) (internalOrderBook, error) {

	path := requestutil.FullPath(baseUrl, "api/v3/depth")
	values := url.Values{}
	values.Add("symbol", pairCode)
	values.Add("limit", "1000")
	path.RawQuery = values.Encode()

	body, err := GetBody(http.DefaultClient.Get(path.String()))
	if err != nil {
		return internalOrderBook{}, err
	}

	snapshot := struct {
		LastUpdateId int64      `json:"lastUpdateId"`
		Bids         [][]string `json:"bids"`
		Asks         [][]string `json:"asks"`
	}{}

	err = json.Unmarshal(body, &snapshot)
	if err != nil {
		return internalOrderBook{}, err
	}

	bids, err := convertOrders(snapshot.Bids)
	if err != nil {
		return internalOrderBook{}, err
	}
	asks, err := convertOrders(snapshot.Asks)
	if err != nil {
		return internalOrderBook{}, err
	}

	ob := internalOrderBook{
		lastUpdateId: snapshot.LastUpdateId,
		OrderBook: exchangesdk.OrderBook{
			Bids: bids,
			Asks: asks,
		},
	}

	err = sortOrderBook(&ob)
	if err != nil {
		return internalOrderBook{}, err
	}

	return ob, nil
}

func handleOrderBookUpdate(
	ob *internalOrderBook,
	updateMsg []byte,
	exConf ExchangeConfig,
) error {

	update := struct {
		FirstUpdateId int64      `json:"U"`
		LastUpdateId  int64      `json:"u"`
		BidUpdates    [][]string `json:"b"`
		AskUpdates    [][]string `json:"a"`
		Timestamp     int64      `json:"E"`
		Temp          string     `json:"e"`
	}{}

	err := json.Unmarshal(updateMsg, &update)
	if err != nil {
		return err
	}

	if update.LastUpdateId <= ob.lastUpdateId {
		return nil
	}

	if update.FirstUpdateId > ob.lastUpdateId+1 {
		return fmt.Errorf(
			"missed some updates; got update %d but last ob update is %d",
			update.FirstUpdateId,
			ob.lastUpdateId,
		)
	}

	err = UpdateOrders(&ob.Bids, update.BidUpdates, exConf)
	if err != nil {
		return err
	}
	err = UpdateOrders(&ob.Asks, update.AskUpdates, exConf)
	if err != nil {
		return err
	}

	err = sortOrderBook(ob)
	if err != nil {
		return err
	}

	ob.lastUpdateId = update.LastUpdateId

	ob.Timestamp = time.Unix(0, update.Timestamp*int64(time.Millisecond))

	return nil
}

func decodeTrade(msgData []byte) (exchangesdk.OrderBookTrade, error) {

	tradeJson := struct {
		Price        float64 `json:"p,string"`
		Volume       float64 `json:"q,string"`
		BuyerIsMaker bool    `json:"m,bool"`
		Timestamp    int64   `json:"E"`
		Temp2        string  `json:"e"`
		Temp         bool    `json:"M,bool"`
	}{}

	err := json.Unmarshal(msgData, &tradeJson)
	if err != nil {
		return exchangesdk.OrderBookTrade{}, err
	}

	makerSide := exchangesdk.OrderBookSideAsk
	if tradeJson.BuyerIsMaker {
		makerSide = exchangesdk.OrderBookSideBid
	}

	return exchangesdk.OrderBookTrade{
		MakerSide: makerSide,
		Price:     tradeJson.Price,
		Volume:    tradeJson.Volume,
		Timestamp: time.Unix(0, tradeJson.Timestamp*int64(time.Millisecond)),
	}, nil
}

func pricesEqual(a, b exchangesdk.OrderBookOrder, pricePrecision float64) bool {

	return math.Abs(a.Price-b.Price) < (pricePrecision / float64(2))
}

func hasZeroVolume(o exchangesdk.OrderBookOrder, volPrecision float64) bool {

	return math.Abs(o.Volume) < (volPrecision / float64(2))
}

func UpdateOrders(
	currentOrders *[]exchangesdk.OrderBookOrder,
	updates [][]string,
	exConf ExchangeConfig,
) error {

	for _, update := range updates {

		orderUpdate, err := convertOrderStrings(update)
		if err != nil {
			return err
		}

		foundOrder := false
		for i := range *currentOrders {
			if pricesEqual((*currentOrders)[i], orderUpdate, exConf.PricePrecision) {
				foundOrder = true

				(*currentOrders)[i].Volume = orderUpdate.Volume

				if hasZeroVolume((*currentOrders)[i], exConf.VolPrecision) {
					(*currentOrders)[i] = (*currentOrders)[len(*currentOrders)-1]
					*currentOrders = (*currentOrders)[:len(*currentOrders)-1]
				}

				break
			}
		}

		if !foundOrder && !hasZeroVolume(orderUpdate, exConf.VolPrecision) {
			*currentOrders = append(*currentOrders, orderUpdate)
		}
	}

	return nil
}

func convertOrders(raw [][]string) ([]exchangesdk.OrderBookOrder, error) {

	orders := make([]exchangesdk.OrderBookOrder, 0, len(raw))
	for _, o := range raw {

		order, err := convertOrderStrings(o)
		if err != nil {
			return nil, err
		}

		orders = append(orders, order)
	}
	return orders, nil
}

func convertOrderStrings(rawOrder []string) (exchangesdk.OrderBookOrder, error) {

	if len(rawOrder) != 2 {
		return exchangesdk.OrderBookOrder{}, fmt.Errorf("Raw order len != 2")
	}

	price, err := strconv.ParseFloat(rawOrder[0], 64)
	if err != nil {
		return exchangesdk.OrderBookOrder{}, err
	}

	volume, err := strconv.ParseFloat(rawOrder[1], 64)
	if err != nil {
		return exchangesdk.OrderBookOrder{}, err
	}

	return exchangesdk.OrderBookOrder{
		Price:  price,
		Volume: volume,
	}, nil
}

// DEPRECATED Use exchangesdk.SortOrderBook instead
func sortOrderBook(ob *internalOrderBook) error {

	err := sortOrders(&ob.Bids, sortOrderingDecending)
	if err != nil {
		return err
	}
	err = sortOrders(&ob.Asks, sortOrderingIncrementing)
	if err != nil {
		return err
	}
	return nil
}

type sortOrdering int

const (
	sortOrderingDecending = iota
	sortOrderingIncrementing
	sortOrderingUnknown
)

func sortOrders(orders *[]exchangesdk.OrderBookOrder, ordering sortOrdering) error {

	switch ordering {
	case sortOrderingDecending:
		sort.Slice(*orders, func(i, j int) bool {

			return (*orders)[i].Price > (*orders)[j].Price
		})
		return nil
	case sortOrderingIncrementing:
		sort.Slice(*orders, func(i, j int) bool {

			return (*orders)[i].Price < (*orders)[j].Price
		})
		return nil
	default:
		return fmt.Errorf("Unknown sort order")
	}
}

func newWebsocket(wsUrl string) (*websocket.Conn, time.Time, error) {

	ws, _, err := websocket.DefaultDialer.Dial(wsUrl, nil)
	if err != nil {
		return nil, time.Time{}, err
	}
	return ws, time.Now(), nil
}
