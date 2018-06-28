package cobinhood

import (
	"encoding/json"
	"log"
	"strings"

	"github.com/jflyup/goup"
	"github.com/jflyup/goup/util"

	"github.com/gorilla/websocket"
)

var (
	wsBaseURL = "wss://ws.cobinhood.com/v2/ws"
)

func (c *Client) createWsConn() error {
	c.createWsLock.Lock()
	defer c.createWsLock.Unlock()

	if c.wsConn == nil {
		var err error
		if c.wsConn, _, err = websocket.DefaultDialer.Dial(wsBaseURL, nil); err != nil {
			log.Printf("ERROR\thuobi websocket error: %v", err)
			return err
		}

		c.pubsub = util.NewPubSub(16)
		go c.wsLoop()
	}
	return nil
}

func transformDepth(ch string, d *wsDepth) *goup.Depth {
	// ignore the returned error, this should be ok
	pair, _ := goup.ParseSymbol(strings.Split(ch, ".")[1])
	depth := &goup.Depth{
		Pair:    pair,
		AskList: make([]goup.DepthRecord, 0, len(d.Asks)),
		BidList: make([]goup.DepthRecord, 0, len(d.Bids)),
	}

	for _, ask := range d.Asks {
		record := goup.DepthRecord{
			Price:  util.ToFloat64(ask[0]),
			Amount: util.ToFloat64(ask[2]),
		}
		depth.AskList = append(depth.AskList, record)
	}

	for _, bid := range d.Bids {
		record := goup.DepthRecord{
			Price:  util.ToFloat64(bid[0]),
			Amount: util.ToFloat64(bid[2]),
		}

		depth.BidList = append(depth.BidList, record)
	}

	return depth
}

func (c *Client) subscribe(subEvent interface{}, keep bool) error {
	// Applications are responsible for ensuring that
	// no more than one goroutine calls the write methods concurrently
	// and that no more than one goroutine calls the read methods concurrently.
	c.writeLock.Lock()
	err := c.wsConn.WriteJSON(subEvent)
	c.writeLock.Unlock()
	if err != nil {
		log.Printf("ERROR\twebsocket write error: %v", err)
		return err
	}

	if keep {
		// keep this for re-subscribe when reconnecting
	}
	return nil
}

func (c *Client) wsLoop() {
	var rsp wsRsp
	depth := &wsDepth{}
	for {
		_, msg, err := c.wsConn.ReadMessage()
		if err != nil {
			log.Printf("ERROR\tfailed to read from huobi websocket: %v", err)
			// if err := c.reconnectWs(); err != nil {
			// 	log.Printf("ERROR\twebsocket reconnect error")
			// 	return
			// }
			continue
		}

		if err := json.Unmarshal(msg, &rsp); err != nil {
			log.Printf("json.Unmarshal error: %v, raw msg: %s", err, string(msg))
			continue
		}

		if strings.Contains(rsp.Header[0], "order-book") && rsp.Header[2] == "s" {
			if err := json.Unmarshal(rsp.Data, depth); err != nil {
				log.Printf("json.Unmarshal error: %v, raw msg: %s", err, string(msg))
				continue
			}

			c.pubsub.Pub(transformDepth(rsp.Header[0], depth), rsp.Header[0])
		}
	}
}
