package gateio

import (
	"crypto/hmac"
	"crypto/sha512"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"sync"

	"github.com/gorilla/websocket"
	"github.com/jflyup/goup"
	"github.com/jflyup/goup/util"
)

var (
	marketBaseURL  = "http://data.gateio.io/api2/1"
	privateBaseURL = "https://api.gateio.io/api2/1/private"
	wsBaseURL      = "wss://ws.gateio.io/v3/"
)

type Client struct {
	client *http.Client
	accessKey,
	secretKey string
	symbolsInfo  map[goup.CurrencyPair]symbolInfo
	wsConn       *websocket.Conn
	subChannels  []interface{}
	createWsLock sync.Mutex
	writeLock    sync.Mutex
	pubsub       *util.PubSub
	// maintain a local order book
	orderBook map[goup.CurrencyPair]*goup.Depth
}

func NewClient(accesskey, secretkey string) (*Client, error) {
	c := &Client{
		client:      http.DefaultClient,
		accessKey:   accesskey,
		secretKey:   secretkey,
		symbolsInfo: make(map[goup.CurrencyPair]symbolInfo),
		orderBook:   make(map[goup.CurrencyPair]*goup.Depth),
	}

	if err := c.marketInfo(); err != nil {
		return nil, err
	}

	return c, nil
}

func (c *Client) httpDo(method string, url string, param string) ([]byte, error) {
	headers := map[string]string{
		// gateio asks this header
		"Content-Type": "application/x-www-form-urlencoded",
	}

	if method == "POST" {
		headers["key"] = c.accessKey
		headers["sign"] = sign(param, c.secretKey)
	}

	return goup.NewHttpRequest(c.client, method, url, param, headers)
}

// AllSymbols implements the API interface
func (c *Client) AllSymbols() ([]goup.CurrencyPair, error) {
	data, err := c.httpDo("GET", marketBaseURL+"/pairs", "")
	if err != nil {
		return nil, err
	}

	var symbols []string
	if err := json.Unmarshal(data, &symbols); err != nil {
		return nil, err
	}

	var pairs []goup.CurrencyPair
	for _, s := range symbols {
		pair, _ := goup.ParseSymbol(s)
		pairs = append(pairs, pair)
	}

	return pairs, nil
}

func (c *Client) marketInfo() error {
	data, err := c.httpDo("GET", marketBaseURL+"/marketinfo", "")
	if err != nil {
		return err
	}
	var info symbolsInfo
	if err := json.Unmarshal(data, &info); err != nil {
		return err
	}
	for _, i := range info.Pairs {
		for k, v := range i {
			pair, err := goup.ParseSymbol(k)
			if err == nil {
				c.symbolsInfo[pair] = v
			}
		}
	}
	log.Printf("%+v", c.symbolsInfo)
	return nil
}

// LimitBuy implements the API interface
func (c *Client) LimitBuy(amount, price float64, pair goup.CurrencyPair) (*goup.Order, error) {
	return c.placeOrder(amount, price, pair, "buy")
}

// LimitSell implements the API interface
func (c *Client) LimitSell(amount, price float64, pair goup.CurrencyPair) (*goup.Order, error) {
	return c.placeOrder(amount, price, pair, "sell")
}

func (c *Client) MarketBuy(amount, price string, pair goup.CurrencyPair) (*goup.Order, error) {
	// it's a shame gateio doesn't support market buy/sell!
	panic("not implement")
}

func (c *Client) MarketSell(amount, price string, currency goup.CurrencyPair) (*goup.Order, error) {
	panic("not implement")
}

func (c *Client) placeOrder(amount, price float64, pair goup.CurrencyPair, side string) (*goup.Order, error) {
	v, ok := c.symbolsInfo[pair]
	if !ok {
		return nil, errors.New("unsupported symbol")
	}

	amountPrecision := strings.Index(fmt.Sprintf("%f", v.MinAmount), "1") - 1
	if amountPrecision < 0 {
		amountPrecision = 0
	}

	amount = util.Truncate(amount, amountPrecision)
	if amount == 0 {
		return nil, goup.ErrLowAmount
	}

	// TODO round instead of truncate?
	price = util.Truncate(price, v.Precision)

	params := url.Values{}
	params.Set("amount", fmt.Sprint(amount))
	// use %f to prevent scientific notation
	params.Set("rate", fmt.Sprintf("%f", price))
	params.Set("currencyPair", pair.ToSymbol("_"))
	var url string
	if side == "buy" {
		url = privateBaseURL + "/buy"
	} else if side == "sell" {
		url = privateBaseURL + "/sell"
	}

	data, err := c.httpDo("POST", url, params.Encode())
	if err != nil {
		return nil, err
	}

	log.Printf("raw: %s", string(data))
	o := &order{}
	if err := json.Unmarshal(data, o); err != nil {
		return nil, err
	}

	order := &goup.Order{
		OrderID:  fmt.Sprint(o.OrderNumber),
		Currency: pair,
	}

	if side == "buy" {
		order.Side = goup.Buy
	} else if side == "sell" {
		order.Side = goup.Sell
	}

	return order, nil
}

func (c *Client) CancelOrder(orderID string, pair goup.CurrencyPair) (bool, error) {
	params := url.Values{}
	params.Set("orderNumber", orderID)
	params.Set("currencyPair", pair.ToSymbol("_"))
	data, err := c.httpDo("POST", privateBaseURL+"/cancelOrder", params.Encode())
	if err != nil {
		return false, err
	}

	r := struct {
		Result  bool // WTF, it's bool here!!
		Message string
	}{}

	if err = json.Unmarshal(data, &r); err != nil {
		return false, err
	}

	if r.Result {
		return true, nil
	}

	return false, errors.New(r.Message)
}

func (c *Client) GetOrder(orderID string, pair goup.CurrencyPair) (*goup.Order, error) {
	params := url.Values{}

	params.Set("currencyPair", pair.ToSymbol("_"))
	params.Set("orderNumber", orderID)

	data, err := c.httpDo("POST", privateBaseURL+"/getOrder", params.Encode())
	if err != nil {
		return nil, err
	}

	var o orderDetail
	if err := json.Unmarshal(data, &o); err != nil {
		return nil, err
	}

	if o.reply.Result != "true" {
		return nil, errors.New(o.reply.Message)
	}

	order := &goup.Order{
		OrderID:  orderID,
		Currency: pair,
	}

	switch o.Order.Status {
	case "open":
		order.Status = goup.Submitted
		amount := util.ToFloat64(o.Order.Amount)
		if amount > 0 {
			order.Status = goup.PartialFilled
			order.DealAmount = amount
		}
	case "done":
		order.Status = goup.Filled
	case "cancelled":
		order.Status = goup.Canceled
	}

	if o.Order.Type == "buy" {
		order.Side = goup.Buy
	} else {
		order.Side = goup.Sell
	}

	return order, nil
}

func (c *Client) OpenOrders() ([]*goup.Order, error) {
	data, err := c.httpDo("POST", privateBaseURL+"/openOrders", "")
	if err != nil {
		return nil, err
	}

	ords := &openOrders{}
	if err := json.Unmarshal(data, ords); err != nil {
		return nil, err
	}

	var orders []*goup.Order
	for _, order := range ords.Orders {
		o := &goup.Order{
			OrderID: fmt.Sprint(order.OrderNumber),
		}
		o.Currency, _ = goup.ParseSymbol(order.CurrencyPair)

		o.Status = goup.Submitted
		o.DealAmount = util.ToFloat64(order.FilledAmount)
		if o.DealAmount > 0 {
			o.Status = goup.PartialFilled
		}

		if order.Type == "buy" {
			o.Side = goup.Buy
		} else {
			o.Side = goup.Sell
		}

		orders = append(orders, o)
	}

	return orders, nil
}

func (c *Client) GetAccount() (*goup.Account, error) {
	data, err := c.httpDo("POST", privateBaseURL+"/balances", "")
	if err != nil {
		return nil, err
	}

	b := &balances{}
	if err := json.Unmarshal(data, b); err != nil {
		return nil, err
	}

	account := &goup.Account{
		SubAccounts: make(map[goup.Currency]goup.SubAccount),
	}

	for k, v := range b.Available {
		amount := util.ToFloat64(v)
		if amount > 0 {
			account.SubAccounts[goup.NewCurrency(k)] = goup.SubAccount{
				Amount: amount,
			}
		}
	}

	return account, nil
}

func (c *Client) GetTicker(currency goup.CurrencyPair) (*goup.Ticker, error) {
	uri := fmt.Sprintf("%s/ticker/%s", marketBaseURL, strings.ToLower(currency.ToSymbol("_")))

	resp, err := goup.HttpGet(c.client, uri)
	if err != nil {
		return nil, err
	}

	return &goup.Ticker{
		Last: util.ToFloat64(resp["last"]),
		Sell: util.ToFloat64(resp["lowestAsk"]),
		Buy:  util.ToFloat64(resp["highestBid"]),
		High: util.ToFloat64(resp["high24hr"]),
		Low:  util.ToFloat64(resp["low24hr"]),
		Vol:  util.ToFloat64(resp["quoteVolume"]),
	}, nil
}

func (c *Client) GetDepth(pair goup.CurrencyPair, size int) (*goup.Depth, error) {
	resp, err := goup.HttpGet(c.client, fmt.Sprintf("%s/orderBook/%s", marketBaseURL, pair.ToSymbol("_")))
	if err != nil {
		return nil, err
	}

	bids, _ := resp["bids"].([]interface{})
	asks, _ := resp["asks"].([]interface{})

	dep := new(goup.Depth)

	for _, v := range bids {
		r := v.([]interface{})
		dep.BidList = append(dep.BidList, goup.DepthRecord{util.ToFloat64(r[0]), util.ToFloat64(r[1])})
	}

	for _, v := range asks {
		r := v.([]interface{})
		dep.AskList = append(dep.AskList, goup.DepthRecord{util.ToFloat64(r[0]), util.ToFloat64(r[1])})
	}

	sort.Sort(sort.Reverse(dep.AskList))

	return dep, nil
}

func (c *Client) GetKlines(pair goup.CurrencyPair, interval goup.KlineInterval, size, since int) ([]*goup.Kline, error) {
	hour := int(math.Ceil(float64(int(interval)*size) / 60.0))
	url := fmt.Sprintf("%s/candlestick2/%s?group_sec=%d&range_hour=%d",
		marketBaseURL, pair.ToSymbol("_"), int(interval)*60, hour)
	data, err := c.httpDo("GET", url, "")
	if err != nil {
		return nil, err
	}

	//log.Printf("raw: %s", string(data))

	rsp := struct {
		Result string     `json:"result"`
		Data   [][]string `json:"data"`
	}{}

	if err := json.Unmarshal(data, &rsp); err != nil {
		return nil, err
	}

	var klines []*goup.Kline

	for _, k := range rsp.Data {
		kline := &goup.Kline{
			Pair:     pair,
			OpenTime: util.ToInt64(k[0]),
			Open:     util.ToFloat64(k[5]),
			Close:    util.ToFloat64(k[2]),
			High:     util.ToFloat64(k[3]),
			Low:      util.ToFloat64(k[4]),
			Vol:      util.ToFloat64(k[1]),
		}
		klines = append(klines, kline)
	}

	return klines, nil
}

func (c *Client) GetTrades(pair goup.CurrencyPair, since int64) ([]goup.Trade, error) {
	panic("not implement")
}

func (c *Client) ExchangeName() string {
	return goup.Gateio
}

func (c *Client) WsKlines(pair goup.CurrencyPair, interval goup.KlineInterval, handler func(*goup.Kline)) error {
	if err := c.createWsConn(); err != nil {
		return err
	}

	if err := c.subscribe(map[string]interface{}{
		"id":     10,
		"method": "kline.subscribe",
		"params": []interface{}{
			pair.ToSymbol("_"), int(interval) * 60,
		},
	}, true); err != nil {
		return err
	}

	ch := c.pubsub.Sub(strings.Join([]string{"kline.subscribe", pair.ToSymbol("_")}, "."))
	go func() {
		for {
			d := (<-ch).(*goup.Kline)
			//log.Printf("%+v", d)
			handler(d)
		}
	}()

	return nil
}

func (c *Client) WsDepth(pair goup.CurrencyPair, handler func(*goup.Depth)) error {
	if err := c.createWsConn(); err != nil {
		return err
	}

	if err := c.subscribe(map[string]interface{}{
		"id":     1,
		"method": "depth.subscribe",
		"params": []interface{}{
			pair.ToSymbol("_"), 30, "0.00000001",
		},
	}, true); err != nil {
		return err
	}

	ch := c.pubsub.Sub(strings.Join([]string{"depth.subscribe", pair.ToSymbol("_")}, "."))
	go func() {
		for {
			d := (<-ch).(*goup.Depth)
			//log.Printf("%+v", d)
			handler(d)
		}
	}()

	return nil
}

func (c *Client) WsTrades(pair goup.CurrencyPair, handler func([]*goup.Trade)) error {
	if err := c.createWsConn(); err != nil {
		return err
	}

	if err := c.subscribe(map[string]interface{}{
		"id":     2,
		"method": "trades.subscribe",
		"params": []string{
			pair.ToSymbol("_"),
		},
	}, true); err != nil {
		return err
	}

	ch := c.pubsub.Sub(strings.Join([]string{"trades.subscribe", pair.ToSymbol("_")}, "."))
	go func() {
		for {
			d := (<-ch).([]*goup.Trade)
			//log.Printf("%+v", d)
			handler(d)
		}
	}()

	return nil
}

// unexported method
func sign(params, secret string) string {
	key := []byte(secret)
	mac := hmac.New(sha512.New, key)
	mac.Write([]byte(params))
	return fmt.Sprintf("%x", mac.Sum(nil))
}

func updateDepth(data []goup.DepthRecord, el goup.DepthRecord, ask bool) []goup.DepthRecord {
	index := 0
	if ask {
		index = sort.Search(len(data), func(i int) bool { return data[i].Price >= el.Price })
	} else {
		index = sort.Search(len(data), func(i int) bool { return data[i].Price <= el.Price })
	}

	if index < len(data) && data[index].Price == el.Price {
		data[index] = el
		if el.Amount == 0 {
			// indices are in range if 0 <= low <= high <= len(a)
			data = append(data[:index], data[index+1:]...)
		}
	} else {
		data = append(data, goup.DepthRecord{})
		copy(data[index+1:], data[index:])
		data[index] = el
	}

	return data
}

func (c *Client) createWsConn() error {
	c.createWsLock.Lock()
	defer c.createWsLock.Unlock()

	if c.wsConn == nil {
		var err error
		if c.wsConn, _, err = websocket.DefaultDialer.Dial(wsBaseURL, nil); err != nil {
			log.Printf("ERROR\thuobi websocket error: %v", err)
			return err
		}

		c.pubsub = util.NewPubSub(16)
		go c.wsLoop()
	}

	return nil
}

func (c *Client) subscribe(subEvent interface{}, keep bool) error {
	// Applications are responsible for ensuring that
	// no more than one goroutine calls the write methods concurrently
	// and that no more than one goroutine calls the read methods concurrently.
	c.writeLock.Lock()
	err := c.wsConn.WriteJSON(subEvent)
	c.writeLock.Unlock()
	if err != nil {
		log.Printf("ERROR\twebsocket write error: %v", err)
		return err
	}

	if keep {
		// keep this for re-subscribe when reconnecting
		c.subChannels = append(c.subChannels, subEvent)
	}
	return nil
}

func parseTrades(data json.RawMessage) ([]*goup.Trade, error) {
	wsNotify := []interface{}{}
	//log.Printf("raw trades: %s", string(data))
	if err := json.Unmarshal(data, &wsNotify); err != nil {
		log.Printf("json.Unmarshal error: %v, raw msg: %s", err, string(data))
		return nil, err
	}

	pair, _ := goup.ParseSymbol(wsNotify[0].(string))
	fffff := wsNotify[1].([]interface{})

	var trades []*goup.Trade
	for _, f := range fffff {
		t := f.(map[string]interface{})
		tttt := &goup.Trade{
			Pair:   pair,
			Amount: util.ToFloat64(t["amount"].(string)),
			Price:  util.ToFloat64(t["price"].(string)),
			Type:   t["type"].(string),
			Ts:     int64(t["time"].(float64) * 1000),
		}

		trades = append(trades, tttt)
	}

	return trades, nil
}

func (c *Client) maintainDepth(data json.RawMessage) {
	// gateio declare an odd json structure, WTF
	wsNotify := []interface{}{}
	if err := json.Unmarshal(data, &wsNotify); err != nil {
		log.Printf("json.Unmarshal error: %v, raw msg: %s", err, string(data))
	}

	snapshot := wsNotify[0].(bool)
	pair, _ := goup.ParseSymbol(wsNotify[2].(string))
	d := wsNotify[1].(map[string]interface{})

	depth := new(goup.Depth)
	depth.Pair = pair
	bids, ok := d["bids"]
	if ok {
		for _, bid := range bids.([]interface{}) {
			_bid := bid.([]interface{})
			amount := util.ToFloat64(_bid[1])
			price := util.ToFloat64(_bid[0])
			dr := goup.DepthRecord{Amount: amount, Price: price}
			depth.BidList = append(depth.BidList, dr)
		}
	}

	asks, ok := d["asks"]
	if ok {
		for _, ask := range asks.([]interface{}) {
			_ask := ask.([]interface{})
			amount := util.ToFloat64(_ask[1])
			price := util.ToFloat64(_ask[0])
			dr := goup.DepthRecord{Amount: amount, Price: price}
			depth.AskList = append(depth.AskList, dr)
		}
	}

	// maintain a loacl order book
	if snapshot {
		c.orderBook[pair] = depth
	} else {
		log.Printf("update: %+v", depth)
		if localDepth, ok := c.orderBook[pair]; ok {
			for _, ask := range depth.AskList {
				localDepth.AskList = updateDepth(localDepth.AskList, ask, true)
			}

			for _, bid := range depth.BidList {
				localDepth.BidList = updateDepth(localDepth.BidList, bid, false)
			}
		} else {
			log.Printf("illegal depth data")
			return
		}
	}

	c.pubsub.Pub(c.orderBook[pair], strings.Join([]string{"depth.subscribe", depth.Pair.ToSymbol("_")}, "."))
}

func (c *Client) reconnectWs() error {
	c.createWsLock.Lock()
	defer c.createWsLock.Unlock()
	c.wsConn.Close()
	conn, err := util.ReconnectWs(wsBaseURL)
	if err != nil {
		return err
	}
	c.wsConn = conn

	// TODO use a set for channels
	// resubscribe channels
	for _, sub := range c.subChannels {
		c.subscribe(sub, false)
	}

	return nil
}

func (c *Client) wsLoop() {
	m := &wsMsg{}
	for {
		_, msg, err := c.wsConn.ReadMessage()
		if err != nil {
			log.Printf("ERROR\tfailed to read from gateio websocket: %v", err)
			if err := c.reconnectWs(); err != nil {
				log.Printf("ERROR\twebsocket reconnect error")
				return
			}
			continue
		}

		if err := json.Unmarshal(msg, &m); err != nil {
			log.Printf("json.Unmarshal error: %v, raw msg: %s", err, string(msg))
			continue
		}

		switch m.Method {
		case "kline.subscribe":
			//c.pubsub.Pub(trades, strings.Join([]string{"kline.subscribe", trades[0].Pair.ToSymbol("_")}, "."))
		case "depth.update":
			c.maintainDepth(m.Params)
		case "trades.update":
			if trades, err := parseTrades(m.Params); err != nil {
				log.Printf("failed to parse depth: %v", err)
			} else {
				if len(trades) > 0 {
					c.pubsub.Pub(trades, strings.Join([]string{"trades.subscribe", trades[0].Pair.ToSymbol("_")}, "."))
				}
			}
		}
	}
}
