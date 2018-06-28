package gateio

import (
	"testing"
	"time"

	"github.com/jflyup/goup"
)

var gate, _ = NewClient("", "")

func TestMarketInfo(t *testing.T) {
	if err := gate.marketInfo(); err != nil {
		t.Errorf("market info error: %v", err)
	}
}

func TestGetAccount(t *testing.T) {
	if account, err := gate.GetAccount(); err != nil {
		t.Errorf("account info error: %v", err)
	} else {
		t.Logf("%+v", account)
	}
}

func TestGetKlines(t *testing.T) {
	if klines, err := gate.GetKlines(goup.NewCurrencyPair("DOCK", "ETH"), goup.KlineInterval1Min, 300, 0); err != nil {
		t.Errorf("klines error: %v", err)
	} else {
		t.Log(klines)
	}
}

func TestAllSymbols(t *testing.T) {
	if symbols, err := gate.AllSymbols(); err != nil {
		t.Errorf("AllSymbols error: %v", err)
	} else {
		if len(symbols) < 100 {
			t.Errorf("error")
		}
	}
}

func TestOpenOrders(t *testing.T) {
	if orders, err := gate.OpenOrders(); err != nil {
		t.Errorf("error: %v", err)
	} else {
		if len(orders) < 2 {
			t.Error("no info")
		}

		t.Logf("%+v", orders)
	}
}
func TestGetOrder(t *testing.T) {
	if order, err := gate.GetOrder("890774002", goup.NewCurrencyPair("DOCK", "ETH")); err != nil {
		t.Errorf("error: %v", err)
	} else {
		t.Log(order)
	}
}

func TestCancelOrder(t *testing.T) {
	if order, err := gate.CancelOrder("899330751", goup.NewCurrencyPair("DOCK", "ETH")); err != nil {
		t.Errorf("error: %v", err)
	} else {
		t.Log(order)
	}
}

func TestLimitSell(t *testing.T) {
	account, err := gate.GetAccount()
	if err != nil {
		t.Errorf("error: %v", err)
	} else {
		amount := account.SubAccounts[goup.NewCurrency("dock")].Amount
		if order, err := gate.LimitSell(amount, 0.000140, goup.NewCurrencyPair("DOCK", "ETH")); err != nil {
			t.Errorf("error: %v", err)
		} else {
			t.Log(order)
		}
	}
}

func TestWsDepth(t *testing.T) {
	if err := gate.WsDepth(goup.NewCurrencyPair("LYM", "ETH"), func(depth *goup.Depth) {
		t.Logf("got depth: %+v", depth)
	}); err != nil {
		t.Errorf("error: %v", err)
	}

	time.Sleep(time.Second * 30)
}

func TestWsTrades(t *testing.T) {
	if err := gate.WsTrades(goup.NewCurrencyPair("LYM", "ETH"), func(trades []*goup.Trade) {
		t.Logf("got depth: %+v", trades)
	}); err != nil {
		t.Errorf("error: %v", err)
	}

	time.Sleep(time.Second * 20)
}

func TestUpdateDepth(t *testing.T) {
	asks := []goup.DepthRecord{
		goup.DepthRecord{Price: 0.00015956, Amount: 11.06957197},
		goup.DepthRecord{Price: 0.00015957, Amount: 6069.4644},
		goup.DepthRecord{Price: 0.00015959, Amount: 38.80574195},
		goup.DepthRecord{Price: 0.00015979, Amount: 31374.8668},
		goup.DepthRecord{Price: 0.0001598, Amount: 20000},
		goup.DepthRecord{Price: 0.0001606, Amount: 5000},
		goup.DepthRecord{Price: 0.00016199, Amount: 2136.71},
	}

	el := goup.DepthRecord{Price: 0.00018955, Amount: 100}
	updated := updateDepth(asks, el, true)
	if updated[7] != el {
		t.Errorf("failed")
	}

	el = goup.DepthRecord{Price: 0.0001598, Amount: 100}
	updated = updateDepth(asks, el, true)
	if updated[4] != el {
		t.Errorf("failed")
	}

	var bids []goup.DepthRecord
	for i := len(asks) - 1; i >= 0; i-- {
		bids = append(bids, asks[i])
	}

	el = goup.DepthRecord{Price: 0.0001607, Amount: 100}
	updated = updateDepth(bids, el, false)
	if updated[1] != el {
		t.Errorf("failed")
	}

	el = goup.DepthRecord{Price: 0.00015956, Amount: 0}
	updated = updateDepth(updated, el, false)
	if len(updated) != 7 {
		t.Errorf("failed")
	}
}
