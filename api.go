package goup

import "errors"

var (
	ErrAPILimit            = errors.New("api limit")
	ErrSignature           = errors.New("signature error")
	ErrInsufficientBalance = errors.New("insufficient balance")
	ErrInvalidSymbol       = errors.New("invalid symbol")
	ErrLowAmount           = errors.New("amount too low")
)

// API offers an universal API for exchanges
type API interface {
	LimitBuy(amount, price float64, pair CurrencyPair) (*Order, error)
	LimitSell(amount, price float64, pair CurrencyPair) (*Order, error)
	MarketBuy(amount, price float64, pair CurrencyPair) (*Order, error)
	MarketSell(amount, price float64, pair CurrencyPair) (*Order, error)
	CancelOrder(orderID string, pair CurrencyPair) (bool, error)
	// GetOrder get detail of single order
	GetOrder(orderID string, pair CurrencyPair) (*Order, error)
	OpenOrders(pair CurrencyPair) ([]*Order, error)
	GetOrderHistory(pair CurrencyPair, currentPage, pageSize int) ([]Order, error)
	GetAccount() (*Account, error)
	// AllSymbols lists all supported symbols of exchange
	AllSymbols() ([]CurrencyPair, error)
	GetTicker(pair CurrencyPair) (*Ticker, error)
	GetDepth(pair CurrencyPair, size int) (*Depth, error)
	GetKlines(pair CurrencyPair, interval KlineInterval, size, since int) ([]*Kline, error)
	GetTrades(pair CurrencyPair, since int64) ([]*Trade, error)

	// WsDepth gets latest order book of specified symbol via websocket
	WsDepth(pair CurrencyPair, handler func(*Depth)) error
	// WsTrades gets updates of trade info via websocket
	WsTrades(pair CurrencyPair, handler func([]*Trade)) error
	// WsKlines gets updates of kline via websocket
	WsKlines(pair CurrencyPair, interval KlineInterval, handler func(*Kline)) error
	ExchangeName() string
}
