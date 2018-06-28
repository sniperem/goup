package coinbene

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/jflyup/goup"
	"github.com/jflyup/goup/util"
)

const (
	baseURL = "https://api.coinbene.com/v1"
)

type Client struct {
	key    string
	secret string
}

func NewClient(apiKey, secretKey string) *Client {
	client := &Client{
		key:    apiKey,
		secret: secretKey,
	}

	return client
}

func (c *Client) httpDo(method, url string, params map[string]interface{}) ([]byte, error) {
	if params != nil {
		params["timestamp"] = time.Now().UnixNano() / (int64(time.Millisecond) / int64(time.Nanosecond))
		params["apiid"] = c.key
		params["secret"] = c.secret

		var keys []string
		for k := range params {
			keys = append(keys, k)
		}
		sort.Strings(keys)

		var urlParams []string
		for _, key := range keys {
			urlParams = append(urlParams, key+"="+fmt.Sprint(params[key]))
		}

		msg := strings.ToUpper(strings.Join(urlParams, "&"))

		params["sign"] = sign(msg)
		delete(params, "secret")
	}

	client := &http.Client{}

	var req *http.Request
	var err error
	if method == "POST" {
		body, _ := json.Marshal(params)
		req, err = http.NewRequest(method, url, bytes.NewReader(body))
		req.Header.Add("Content-Type", "application/json")
	} else {
		req, err = http.NewRequest(method, url, nil)
		req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	}

	if err != nil {
		return nil, err
	}

	rsp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	defer rsp.Body.Close()

	data, err := ioutil.ReadAll(rsp.Body)

	if err != nil {
		return nil, err
	}

	return data, nil
}

func (c *Client) GetAccount() (*goup.Account, error) {
	params := map[string]interface{}{"account": "exchange"}
	data, err := c.httpDo("POST", baseURL+"/trade/balance", params)
	if err != nil {
		return nil, err
	}

	log.Printf("raw: %s", string(data))
	b := &balanceRsp{}
	if err := json.Unmarshal(data, b); err != nil {
		return nil, err
	}

	account := &goup.Account{
		SubAccounts: make(map[goup.Currency]goup.SubAccount),
	}

	for _, v := range b.Balance {
		amount := util.ToFloat64(v.Available)
		if amount > 0 {
			account.SubAccounts[goup.NewCurrency(v.Asset)] = goup.SubAccount{
				Amount: amount,
			}
		}
	}

	return account, nil
}

func (c *Client) LimitBuy(amount, price float64, pair goup.CurrencyPair) (*goup.Order, error) {
	return c.placeOrder(amount, price, pair, "buy")
}

func (c *Client) LimitSell(amount, price float64, pair goup.CurrencyPair) (*goup.Order, error) {
	return c.placeOrder(amount, price, pair, "sell")
}

func (c *Client) MarketBuy(amount, price float64, pair goup.CurrencyPair) (*goup.Order, error) {
	return nil, errors.New("unsupported")
}

func (c *Client) MarketSell(amount, price float64, pair goup.CurrencyPair) (*goup.Order, error) {
	return nil, errors.New("unsupported")
}

func (c *Client) CancelOrder(orderID string, pair goup.CurrencyPair) (bool, error) {
	params := map[string]interface{}{"orderid": orderID}
	data, err := c.httpDo("POST", baseURL+"/trade/order/cancel", params)
	if err != nil {
		return false, err
	}

	o := &orderRsp{}

	if err = json.Unmarshal(data, &o); err != nil {
		return false, err
	}

	if o.Status == "ok" {
		return true, nil
	}

	return false, errors.New(o.Description)
}

func (c *Client) GetTicker(pair goup.CurrencyPair) (*goup.Ticker, error) {
	data, err := c.httpDo("GET",
		fmt.Sprintf("%s/market/ticker?symbol=%s", baseURL, strings.ToLower(pair.String())), nil)
	if err != nil {
		return nil, err
	}

	rsp := &tickerRsp{}
	//log.Printf("raw: %s", string(data))
	if err := json.Unmarshal(data, rsp); err != nil {
		return nil, err
	}

	log.Println(rsp)
	return nil, nil
}

// GetDepth implements the API interface
func (c *Client) GetDepth(pair goup.CurrencyPair, size int) (*goup.Depth, error) {
	data, err := c.httpDo("GET",
		fmt.Sprintf("%s/market/orderbook?symbol=%s&size=%d", baseURL, strings.ToLower(pair.String()), size), nil)
	if err != nil {
		return nil, err
	}

	rsp := &depthRsp{}
	if err := json.Unmarshal(data, rsp); err != nil {
		return nil, err
	}

	log.Println(string(data))
	d := &goup.Depth{
		Pair: pair,
	}
	for _, bid := range rsp.Orderbook.Bids {
		d.BidList = append(d.BidList, goup.DepthRecord{
			Price:  bid.Price,
			Amount: bid.Quantity,
		})
	}

	for _, ask := range rsp.Orderbook.Asks {
		d.AskList = append(d.AskList, goup.DepthRecord{
			Price:  ask.Price,
			Amount: ask.Quantity,
		})
	}

	return d, nil
}

func (c *Client) OpenOrders(pair goup.CurrencyPair) ([]*goup.Order, error) {
	params := map[string]interface{}{"symbol": pair.String()}
	data, err := c.httpDo("POST", baseURL+"/trade/order/open-orders", params)
	if err != nil {
		return nil, err
	}

	rsp := &openOrdersRsp{}
	if err := json.Unmarshal(data, rsp); err != nil {
		return nil, err
	}

	var orders []*goup.Order
	for _, order := range rsp.Orders.Result {
		o := &goup.Order{
			OrderID: order.OrderID,
		}
		o.Currency = pair

		o.Status = goup.Submitted

		o.DealAmount = util.ToFloat64(order.Filledquantity)
		if o.DealAmount > 0 {
			o.Status = goup.PartialFilled
		}

		if order.Type == "buy" {
			o.Side = goup.BUY
		} else {
			o.Side = goup.SELL
		}

		orders = append(orders, o)
	}

	return orders, nil
}

func (c *Client) placeOrder(amount, price float64, pair goup.CurrencyPair, side string) (*goup.Order, error) {
	params := make(map[string]interface{})

	if side == "buy" {
		params["type"] = "buy-limit"
	} else if side == "sell" {
		params["type"] = "sell-limit"
	}

	params["symbol"] = pair.String()
	params["quantity"] = amount
	params["price"] = price

	data, err := c.httpDo("POST", baseURL+"/trade/order/place", params)
	if err != nil {
		return nil, err
	}

	log.Printf("raw: %s", string(data))
	o := &orderRsp{}
	if err := json.Unmarshal(data, o); err != nil {
		return nil, err
	}

	order := &goup.Order{
		OrderID:  fmt.Sprint(o.Orderid),
		Currency: pair,
	}

	if side == "buy" {
		order.Side = goup.BUY
	} else if side == "sell" {
		order.Side = goup.SELL
	}

	return order, nil
}

func sign(data string) string {
	signature := md5.Sum([]byte(data))
	return hex.EncodeToString(signature[:])
}
