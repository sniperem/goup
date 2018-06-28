package gateio

import "encoding/json"

type (
	reply struct {
		Result  string
		Message string
		Code    int
	}

	symbolInfo struct {
		Precision int     `json:"decimal_places"`
		MinAmount float64 `json:"min_amount"`
		Fee       float64
	}

	symbolsInfo struct {
		Result string
		Pairs  []map[string]symbolInfo
	}

	balances struct {
		reply
		Available map[string]string
		Locked    map[string]string
	}

	order struct {
		reply
		OrderNumber  int64
		Rate         string
		LeftAmount   string
		FilledAmount string
		FilledRate   string
	}

	orderDetail struct {
		reply
		Order struct {
			OrderNumber   string
			Status        string
			CurrencyPair  string
			Type          string
			Rate          string
			Amount        string
			InitialRate   string
			InitialAmount string
		}
	}

	openOrders struct {
		reply
		Elapsed string `json:"elapsed"`
		Orders  []struct {
			// WTF!! OrderNumber here is int while it's string else where
			OrderNumber   int64  `json:"orderNumber"`
			Type          string `json:"type"`
			Rate          string `json:"rate"`
			Amount        string `json:"amount"`
			Total         string `json:"total"`
			InitialRate   string `json:"initialRate"`
			InitialAmount string `json:"initialAmount"`
			FilledRate    string `json:"filledRate"`
			FilledAmount  string `json:"filledAmount"`
			CurrencyPair  string `json:"currencyPair"`
			Timestamp     int64  `json:"timestamp"`
			Status        string `json:"status"`
		} `json:"orders"`
	}

	errorMsg struct {
		Code    int
		Message string
	}

	wsMsg struct {
		id     int64
		Method string
		Error  errorMsg
		Params json.RawMessage
		Result json.RawMessage
	}

	wsTrade struct {
		ID     int     `json:"id"`
		Time   float64 `json:"time"`
		Price  string  `json:"price"`
		Amount string  `json:"amount"`
		Type   string  `json:"type"`
	}
)
