package goup

type TradeSide int

const (
	Buy = iota
	Sell
)

func (ts TradeSide) String() string {
	switch ts {
	case Buy:
		return "buy"
	case Sell:
		return "sell"
	default:
		return "unknown"
	}
}

// OrderStatus represents status of order
type OrderStatus int

const (
	Submitted OrderStatus = iota
	PartialFilled
	Filled
	Canceled
	Rejected
	Canceling
	Expired
)

func (s OrderStatus) String() string {
	status := [...]string{"Submitted", "Partial Filled", "Filled", "Canceled", "Rejected", "Canceling", "Expired"}
	if s < Submitted || s > Expired {
		return "Unknown"
	}

	return status[s]
}

// KlineInterval is the interval of k line
type KlineInterval int

const (
	KlineInterval1Min  KlineInterval = 1
	KlineInterval5Min  KlineInterval = 5
	KlineInterval15Min KlineInterval = 15
	KlineInterval30Min KlineInterval = 30
	KlineInterval1H    KlineInterval = 60
	KlineInterval4H    KlineInterval = 240
	KlineInterval1Day  KlineInterval = 1440
	KlineInterval1Week
	KlineInterval1Month
)

const (
	Cobinhood = "cobinhood.com"
	Gateio    = "gate.io"
)
