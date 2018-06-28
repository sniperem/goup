# goup
A universal API for different cryptocurrency exchanges. Currently supports [cobinhood](https://cobinhood.com/), [gate.io](gate.io), [coinbene](https://www.coinbene.com/).
***
I've gone through many API docs of exchanges, including binance, huobi, bittrex, idex, etc. Some exchanges seem not to care about it, or not capable to design a good API. For example,

* kucoin: no websocket interface which means you have to repeatedly poll the server for updates.
* cryptopia: worst of worst.
* idex: this so-called decentralized exchange really sucks. Its website is ugly and lagging, even worse, it leads high CPU usage and goes offline frequently. In the world of cryptocurrency, timing is more critical than stock! The API is unstable too, no developer fixes a simple bug for long time.

As a top exchange, binance is doing a good job on API. Real-time order book via Websocket, API is well documented and updated in time.  
**But what surprises me is a small but rising exchange--[cobinhood](https://cobinhood.com/), the api is designed delicately, and supports trading via websocket(even binance doesn't support it), and ZERO TRADING FEE!  It's a gem out of exchanges. Currently the volume is small, hope more people know it and happy trading, keep away from the trash exchanges**
***
Here is the high-level API:

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
