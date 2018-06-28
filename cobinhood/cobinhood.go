package cobinhood

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/jflyup/goup"
	"github.com/jflyup/goup/util"
)

const (
	baseURL = "https://api.cobinhood.com"
)

type Client struct {
	apiKey       string
	wsConn       *websocket.Conn
	createWsLock sync.Mutex
	writeLock    sync.Mutex
	pubsub       *util.PubSub
	currencyInfo map[goup.Currency]Currency
}

func NewClient(apiKey string) (*Client, error) {
	client := &Client{
		apiKey:       apiKey,
		currencyInfo: make(map[goup.Currency]Currency),
	}

	if err := client.currencies(); err != nil {
		return nil, err
	}

	return client, nil
}

func (c *Client) OpenOrders() ([]*goup.Order, error) {
	return nil, errors.New("not implemented")
}

func (c *Client) AllSymbols() ([]goup.CurrencyPair, error) {
	rsp, err := c.get("/v1/market/trading_pairs")

	if err != nil {
		return nil, err
	}

	var symbols []goup.CurrencyPair
	for _, pair := range rsp.Result.TradingPairs {
		s := goup.NewCurrencyPair(pair.BaseCurrencyId, pair.QuoteCurrencyId)
		symbols = append(symbols, s)
	}

	return symbols, nil
}

func (c *Client) GetOrder(orderID string, pair goup.CurrencyPair) (*goup.Order, error) {
	rsp, err := c.get(fmt.Sprintf("/v1/trading/orders/%s", orderID))

	if err != nil {
		return nil, err
	}

	order := rsp.Result.Order
	// [queued, open, partially_filled, filled, cancelled, rejected,
	// pending_cancellation, pending_modifications, triggered]
	ord := &goup.Order{
		Price:      util.ToFloat64(order.Price),
		Amount:     util.ToFloat64(order.Size),
		DealAmount: util.ToFloat64(order.Filled),
		Fee:        0, // zero trading fee!
		OrderID:    orderID,
		// CreateTime int64 // in ms
		// FinishTime int64
		Currency: pair,
	}

	switch order.State {
	case "filled":
		ord.Status = goup.Filled
	case "partially_filled":
		ord.Status = goup.PartialFilled
	case "cancelled":
		ord.Status = goup.Canceled
	case "pending_cancellation":
		ord.Status = goup.Canceling
	case "rejected":
		ord.Status = goup.Rejected
	}

	if order.Side == "ask" {
		ord.Side = goup.Sell
	} else {
		ord.Side = goup.Buy
	}

	return ord, nil
}

func (c *Client) CancelOrder(orderID string, pair goup.CurrencyPair) (bool, error) {
	err := c.delete(fmt.Sprintf("/v1/trading/orders/%s", orderID))

	if err != nil {
		return false, err
	}

	return true, nil
}

func (c *Client) get(path string) (*Response, error) {
	req, err := c.request("GET", path, nil)

	if err != nil {
		return nil, err
	}

	rsp, err := c.client().Do(req)

	if err != nil {
		return nil, err
	}

	defer rsp.Body.Close()

	if rsp.StatusCode > 399 {
		return nil, fmt.Errorf("HTTP Status Code: %d ", rsp.StatusCode)
	}

	data, err := ioutil.ReadAll(rsp.Body)

	if err != nil {
		return nil, err
	}

	jsonRsp := &Response{}
	err = json.Unmarshal(data, jsonRsp)

	if err != nil {
		return nil, err
	}

	if jsonRsp.Success == false {
		return nil, errors.New(jsonRsp.Error.Err)
	}

	return jsonRsp, nil
}

func (c *Client) post(path string, body io.Reader) (*Response, error) {
	req, err := c.request("POST", path, body)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	rsp, err := c.client().Do(req)

	if err != nil {
		return nil, err
	}

	defer rsp.Body.Close()

	if rsp.StatusCode > 399 {
		return nil, fmt.Errorf("response status: %d ", rsp.StatusCode)
	}

	data, err := ioutil.ReadAll(rsp.Body)

	if err != nil {
		return nil, err
	}

	var jsonRsp *Response
	err = json.Unmarshal(data, jsonRsp)

	if err != nil {
		return nil, err
	}

	if jsonRsp.Success == false {
		return nil, errors.New(jsonRsp.Error.Err)
	}

	return jsonRsp, nil
}

func (c *Client) delete(path string) error {
	req, err := c.request("DELETE", path, nil)

	if err != nil {
		return nil
	}

	rsp, err := c.client().Do(req)

	if err != nil {
		return err
	}

	defer rsp.Body.Close()
	data, err := ioutil.ReadAll(rsp.Body)

	if err != nil {
		return err
	}

	var jsonRsp *Response
	err = json.Unmarshal(data, jsonRsp)

	if err != nil {
		return err
	}

	if jsonRsp.Success == false {
		return errors.New(jsonRsp.Error.Err)
	}

	return nil
}

func (c *Client) client() *http.Client {
	client := &http.Client{}
	return client
}

func (c *Client) request(method, path string, body io.Reader) (*http.Request, error) {
	req, err := http.NewRequest(method, baseURL+path, body)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Content-Type", "application/json")
	if c.apiKey != "" {
		req.Header.Add("Authorization", c.apiKey)
	}

	req.Header.Add("nonce", fmt.Sprintf("%v", int32(time.Now().Unix())))

	return req, nil
}

func (c *Client) currencies() error {
	rsp, err := c.get("/v1/market/currencies")
	if err != nil {
		return err
	}

	for _, currency := range rsp.Result.Currencies {
		c.currencyInfo[goup.NewCurrency(currency.Currency)] = currency
	}

	return nil
}

func (c *Client) ExchangeName() string {
	return goup.Cobinhood
}

func (c *Client) LimitBuy(amount, price float64, pair goup.CurrencyPair) (*goup.Order, error) {
	// data := new(bytes.Buffer)
	// err := json.NewEncoder(data).Encode(datajson)
	info, ok := c.currencyInfo[pair.Base]
	if !ok {
		return nil, errors.New("unsupported symbol")
	}

	precision := strings.Index(info.MinUnit, "1") - 1
	if precision < 0 {
		precision = 0
	}
	amount = util.Truncate(amount, precision)
	if amount == 0 {
		// TODO error type
		return nil, errors.New("too little")
	}

	params := PlaceOrder{
		TradingPairId: pair.ToSymbol("-"),
		// TODO
		Side:  "bid",
		Type:  "limit",
		Price: fmt.Sprintf("%.8f", price),
		Size:  fmt.Sprint(amount),
	}

	data, _ := json.Marshal(params)

	rsp, err := c.post("/v1/trading/orders", bytes.NewReader(data))
	if err != nil {
		return nil, err
	}

	order := rsp.Result.Order
	// [queued, open, partially_filled, filled, cancelled, rejected,
	// pending_cancellation, pending_modifications, triggered]
	ord := &goup.Order{
		Price:      util.ToFloat64(order.Price),
		Amount:     util.ToFloat64(order.Size),
		DealAmount: util.ToFloat64(order.Filled),
		//Fee float64
		OrderID: order.ID,
		// CreateTime int64 // in ms
		// FinishTime int64
		Currency: pair,
		Side:     goup.Buy,
	}

	switch order.State {
	case "filled":
		ord.Status = goup.Filled
	case "partially_filled":
		ord.Status = goup.PartialFilled
	case "cancelled":
		ord.Status = goup.Canceled
	case "pending_cancellation":
		ord.Status = goup.Canceling
	case "rejected":
		ord.Status = goup.Rejected
	}

	return ord, nil
}

func (c *Client) GetAccount() (*goup.Account, error) {
	rsp, err := c.get("/v1/wallet/balances")
	if err != nil {
		return nil, err
	}

	acc := &goup.Account{}

	acc.SubAccounts = make(map[goup.Currency]goup.SubAccount)

	for _, balance := range rsp.Result.Balances {
		currency := goup.Currency(strings.ToLower(balance.Currency))
		acc.SubAccounts[currency] = goup.SubAccount{
			Currency:     currency,
			Amount:       util.ToFloat64(balance.Total),
			ForzenAmount: util.ToFloat64(balance.OnOrder),
		}
	}

	return acc, nil
}

func (c *Client) WsDepth(pair goup.CurrencyPair, handler func(*goup.Depth)) error {
	// available precisions could be acquired from REST,
	// endpoint: /v1/market/orderbook/precisions/<trading_pair_id>
	// if rsp, err := c.get("/v1/market/orderbook/precisions/"); err != nil {
	// 	return err
	// }
	// "result": [
	// 	"1E-7",
	// 	"5E-7",
	// ]
	if err := c.createWsConn(); err != nil {
		return err
	}

	if err := c.subscribe(map[string]interface{}{
		"action":          "subscribe",
		"type":            "order-book",
		"trading_pair_id": pair.ToSymbol("-"),
		"precision":       "1E-7",
	}, true); err != nil {
		return err
	}

	chDepth := c.pubsub.Sub(strings.Join([]string{"order-book", pair.ToSymbol("-"), "1E-7"}, "."))
	go func() {
		d := <-chDepth
		depth := d.(*goup.Depth)
		handler(depth)
	}()

	return nil
}

func (c *Client) WsTrades(pair goup.CurrencyPair, handler func([]*goup.Trade)) error {
	if err := c.createWsConn(); err != nil {
		return err
	}

	if err := c.subscribe(map[string]interface{}{
		"action":          "subscribe",
		"type":            "trade",
		"trading_pair_id": pair.ToSymbol("-"),
	}, true); err != nil {
		return err
	}

	chTrade := c.pubsub.Sub(strings.Join([]string{"trade", pair.ToSymbol("-")}, "."))
	go func() {
		t := <-chTrade
		trade := t.(*goup.Trade)
		handler([]*goup.Trade{trade})
	}()

	return nil
}

func (c *Client) WsKlines(pair goup.CurrencyPair, interval goup.KlineInterval, handler func(*goup.Kline)) error {
	return errors.New("not implemented")
}

// 0: limit
// 1: market
// 2: stop
// 3: limit_stop
// 4: trailing_fiat_stop (not valid yet)
// 5: fill_or_kill (not valid yet)
// 6: trailing_percent_stop (not valid yet)

func (c *Client) WsLimitBuy(amount, price float64, pair goup.CurrencyPair) (*goup.Order, error) {
	// data := new(bytes.Buffer)
	// err := json.NewEncoder(data).Encode(datajson)
	params := wsOrderParams{
		TradingPairID: pair.ToSymbol("-"),
		Side:          "bid",
		//Type:          "limit",
		Price: fmt.Sprintf("%.8f", price),
		Size:  fmt.Sprint(amount),
	}

	err := c.subscribe(params, false)
	if err != nil {
		return nil, err
	}

	return nil, nil
	// // [queued, open, partially_filled, filled, cancelled, rejected,
	// // pending_cancellation, pending_modifications, triggered]
	// ord := &goup.Order{
	// 	Price:      util.ToFloat64(order.Price),
	// 	Amount:     util.ToFloat64(order.Size),
	// 	DealAmount: util.ToFloat64(order.Filled),
	// 	//Fee float64
	// 	OrderID: order.ID,
	// 	// CreateTime int64 // in ms
	// 	// FinishTime int64
	// 	Currency: pair,
	// 	Side:     goup.BUY,
	// }

	// switch order.State {
	// case "filled":
	// 	ord.Status = goup.Filled
	// case "partially_filled":
	// 	ord.Status = goup.PartialFilled
	// case "cancelled":
	// 	ord.Status = goup.Canceled
	// case "pending_cancellation":
	// 	ord.Status = goup.Canceling
	// case "rejected":
	// 	ord.Status = goup.Rejected
	// }

	// return ord, nil
}
