package util

import (
	"errors"
	"log"
	"math"
	"time"

	"github.com/gorilla/websocket"
	"github.com/jflyup/goup"
)

// Retry executes a function until:
// 1. A nil error is returned,
// 2. The max number of attempts has been reached,
// 3. A Stop(...) wrapped error is returned
func Retry(attempts int, sleep time.Duration, fn func() error) error {
	if err := fn(); err != nil {
		if s, ok := err.(stopError); ok {
			// Return the original error for later checking
			return s.error
		}

		if attempts--; attempts > 0 {
			time.Sleep(sleep)
			return Retry(attempts, 2*sleep, fn)
		}
		return err
	}
	return nil
}

type stopError struct {
	error
}

// ReconnectWs re-establish a websocket connection
func ReconnectWs(endpoint string) (c *websocket.Conn, err error) {
	if e := Retry(3, 5*time.Second, func() error {
		c, _, err = websocket.DefaultDialer.Dial(endpoint, nil)
		if err != nil {
			log.Printf("failed to establish a websocket connection: %v, retrying...", err)
			return err
		}
		return nil
	}); e != nil {
		return
	}

	return
}

// CalcBuyPrice returns the price
func CalcBuyPrice(asks goup.DepthRecords, amount float64) float64 {
	cost := amount
	bought := 0.0
	for _, ask := range asks {
		if ask.Price*ask.Amount >= amount {
			bought += amount / ask.Price
			amount = 0
			break
		} else {
			amount -= ask.Price * ask.Amount
			bought += ask.Amount
		}
	}

	// if amount > 0 {
	// 	log.Printf("the market can't fill this sell order, amount: %f", amount)
	// }

	return cost / bought
}

func CalcSellPrice(bids goup.DepthRecords, amount float64) (float64, error) {
	sales := 0.0
	sold := 0.0
	for _, bid := range bids {
		if bid.Amount >= amount {
			sales += amount * bid.Price
			sold += amount
			amount = 0
			break
		} else {
			amount -= bid.Amount
			sold += bid.Amount
			sales += bid.Amount * bid.Price
		}
	}

	var err error
	if amount > 0 {
		//log.Printf("the market can't fill this sell order, amount: %f", amount)
		err = errors.New("the market can't fill this sell order")
	}

	return sales / sold, err
}

func Truncate(num float64, precision int) float64 {
	return math.Floor(num*math.Pow10(precision)) / math.Pow10(precision)
}

// Round return the floating point value number rounded to ndigits digits after the decimal point
