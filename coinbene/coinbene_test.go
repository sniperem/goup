package coinbene

import (
	"testing"

	"github.com/jflyup/goup"
)

func TestGetDepth(t *testing.T) {
	c := NewClient("", "")
	if _, err := c.GetDepth(0, goup.NewCurrencyPair("ABT", "ETH")); err != nil {
		t.Error(err)
	}
}

func TestGetTicker(t *testing.T) {
	c := NewClient("", "")
	if _, err := c.GetTicker(goup.NewCurrencyPair("ABT", "ETH")); err != nil {
		t.Error(err)
	}
}

func TestLimitBuy(t *testing.T) {
	c := NewClient("0924451f52dd61b02552e245235328db", "86ec0616e7f94164af8abfa734413b3e")
	if order, err := c.LimitBuy(10, 0.0012, goup.NewCurrencyPair("ABT", "ETH")); err != nil {
		t.Error(err)
	} else {
		t.Log(order)
	}
}

func TestGetAccount(t *testing.T) {
	c := NewClient("0924451f52dd61b02552e245235328db", "86ec0616e7f94164af8abfa734413b3e")
	if account, err := c.GetAccount(); err != nil {
		t.Error(err)
	} else {
		t.Log(account)
	}
}

func TestCancelOrder(t *testing.T) {
	c := NewClient("0924451f52dd61b02552e245235328db", "86ec0616e7f94164af8abfa734413b3e")
	if _, err := c.CancelOrder("1234", goup.NewCurrencyPair("ABT", "ETH")); err == nil {
		t.Error("error")
	} else {
		t.Log(err)
	}
}
