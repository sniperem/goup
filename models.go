package goup

import (
	"fmt"
)

// Order represents a buy/sell order
type Order struct {
	Price,
	Amount,
	AvgPrice,
	DealAmount,
	Fee float64
	OrderID    string
	CreateTime int64 // in ms
	FinishTime int64
	Status     OrderStatus
	Currency   CurrencyPair
	Side       TradeSide
}

func (o Order) String() string {
	return fmt.Sprintf("%s %s order %s, status: %s", o.Side, o.Currency, o.OrderID, o.Status)
}

// Trade represents a history trade of market
type Trade struct {
	Pair   CurrencyPair
	Tid    int64
	Type   string
	Amount float64
	Price  float64
	Ts     int64
}

type SubAccount struct {
	Currency Currency
	Amount,
	ForzenAmount float64
}

type Account struct {
	Exchange    string
	Asset       float64
	NetAsset    float64
	SubAccounts map[Currency]SubAccount
}

type Ticker struct {
	Last float64 `json:"last"`
	Buy  float64 `json:"buy"`
	Sell float64 `json:"sell"`
	High float64 `json:"high"`
	Low  float64 `json:"low"`
	Vol  float64 `json:"vol"`
	Date uint64  `json:"date"`
}

type DepthRecord struct {
	Price,
	Amount float64
}

type DepthRecords []DepthRecord

func (dr DepthRecords) Len() int {
	return len(dr)
}

func (dr DepthRecords) Swap(i, j int) {
	dr[i], dr[j] = dr[j], dr[i]
}

func (dr DepthRecords) Less(i, j int) bool {
	return dr[i].Price < dr[j].Price
}

// Depth is the order book of a trading pair
type Depth struct {
	Pair CurrencyPair
	AskList,
	BidList DepthRecords
}

// Kline is k-line
type Kline struct {
	Pair     CurrencyPair
	Ts       int64 // in ms for consistency
	OpenTime int64
	Open,
	Close,
	High,
	Low,
	Vol float64
}
