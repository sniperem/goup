package cobinhood

import "encoding/json"

type Order struct {
	ID          string `json:"id"`
	TradingPair string `json:"trading_pair"`
	State       string
	Side        string
	Type        string
	Price       string
	Size        string
	Filled      string
	Timestamp   int64
}

type balance struct {
	Currency string
	Type     string
	Total    string
	OnOrder  string `json:"on_order"`
	Locked   bool
	UsdValue string `json:"usd_value"`
	BtcValue string `json:"btc_value"`
}

type TradingPair struct {
	ID              string `json:"Id"`
	BaseCurrencyId  string `json:"Base_currency_id"`
	QuoteCurrencyId string `json:"Quote_currency_id"`
	BaseMinSize     string `json:"Base_min_size"`
	BaseMaxSize     string `json:"Base_max_size"`
	QuoteIncrement  string `json:"Quote_increment"`
}

type PlaceOrder struct {
	TradingPairId string `json:"trading_pair_id"`
	Side          string `json:"side"`
	Type          string `json:"type"`
	Price         string `json:"price"`
	Size          string `json:"size"`
}

type Currency struct {
	Currency      string `json:"currency"`
	Name          string `json:"name"`
	MinUnit       string `json:"min_unit"`
	DepositFee    string `json:"deposit_fee"`
	WithdrawalFee string `json:"withdrawal_fee"`
	MinAmount     string `json:"funding_min_size"`
}

type Response struct {
	Success bool   `json:"success"`
	Result  Result `json:"result"`
	Error   errorMsg
}

type Result struct {
	//Info         Info         `json:"info"`
	Currencies      []Currency `json:"currencies"`
	QuoteCurrencies []Currency `json:"quote_currencies"`
	Balances        []balance
	TradingPairs    []TradingPair `json:"trading_pairs"`
	// Orderbook    Orderbook    `json:"orderbook"`
	// Ticker       Ticker       `json:"ticker"`
	// Trades       Trades       `json:"trades"`
	Order  Order   `json:"order"`
	Orders []Order `json:"orders"`
	Error  string  `json:"error"`
}

type errorMsg struct {
	Err string `json:"error_code"`
}

type wsRsp struct {
	Header []string        `json:"h"` // example: ["order-book.COB-ETH.1E-7", "2", "u"],
	Data   json.RawMessage `json:"d"`
}

type wsDepth struct {
	Bids [][]string
	Asks [][]string
}

type wsOrderParams struct {
	Action           string
	TradingPairID    string `json:"trading_pair_id"`
	Side             string `json:"side"`
	Type             string `json:"type"`
	Price            string `json:"price"`
	Size             string `json:"size"`
	StopPrice        string `json:"stop_price"`        // mandatory for stop/stop-limit order
	TrailingDistance string `json:"trailing_distance"` // mandatory for trailing-stop order
	ID               string `json:"id"`
}
