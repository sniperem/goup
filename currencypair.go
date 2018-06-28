package goup

import (
	"errors"
	"strings"
)

// Currency is a token/coin
type Currency string

func NewCurrency(s string) Currency {
	return Currency(strings.ToUpper(s))
}

const (
	ETH = Currency("ETH")
)

// CurrencyPair is a trading pair
type CurrencyPair struct {
	Base  Currency
	Quote Currency
}

// String implements the Stringer interface
func (pair CurrencyPair) String() string {
	// make sure it's upper case
	return strings.ToUpper(pair.ToSymbol(""))
}

// NewCurrencyPair creates a trading pair from string
func NewCurrencyPair(base, quote string) CurrencyPair {
	return CurrencyPair{Currency(strings.ToUpper(base)), Currency(strings.ToUpper(quote))}
}

// ParseSymbol parses s as a tradeing pair, using one of the following formats:
// ethusdt, NEOETH, zil/eth, blz_eth
func ParseSymbol(s string) (CurrencyPair, error) {
	s = strings.ToUpper(s)
	f := func(c rune) bool {
		return c == '/' || c == '_'
	}
	symbols := strings.FieldsFunc(s, f)
	if len(symbols) == 2 {
		return NewCurrencyPair(symbols[0], symbols[1]), nil
	} else if strings.HasSuffix(s, "ETH") ||
		strings.HasSuffix(s, "BTC") || strings.HasSuffix(s, "BNB") {
		return NewCurrencyPair(s[:len(s)-3], s[len(s)-3:]), nil
	} else if strings.HasSuffix(s, "USDT") {
		return NewCurrencyPair(s[:len(s)-4], s[len(s)-4:]), nil
	} else if strings.HasSuffix(s, "HT") {
		return NewCurrencyPair(s[:len(s)-2], s[len(s)-2:]), nil
	}

	return CurrencyPair{}, errors.New("unsupported tradeing pair")
}

// ToSymbol concatenate base assert and quote asset with intervening a sep
func (pair CurrencyPair) ToSymbol(sep string) string {
	return strings.ToUpper(strings.Join([]string{string(pair.Base), string(pair.Quote)}, sep))
}
