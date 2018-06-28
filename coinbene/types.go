package coinbene

type (
	rsp struct {
		Description string
		Status      string
		Timestamp   int64
	}
	depthRsp struct {
		rsp
		Symbol    string
		Orderbook struct {
			Asks []struct {
				Price    float64
				Quantity float64
			}

			Bids []struct {
				Price    float64
				Quantity float64
			}
		}
	}

	tickerRsp struct {
		rsp
		Ticker []struct {
			Symbol string
			Last   string
			Bid    string
			Ask    string
			High   string `json:"24hrHigh"`
			Low    string `json:"24hrLow"`
			Vol    string `json:"24hrVol"`
			Amount string `json:"24hrAmt"`
		}
	}

	tradesRsp struct {
		rsp
		Symbol string `json:"symbol"`
		Trades struct {
			TradeID  string  `json:"tradeId "`
			Price    float64 `json:"price"`
			Quantity float64 `json:"quantity"`
			Take     string  `json:"take"`
			Time     string  `json:"time"`
		} `json:"trades"`
	}

	orderRsp struct {
		rsp
		Orderid string `json:"orderid"`
	}

	balanceRsp struct {
		rsp
		Orderid string `json:"orderid"`
		Account string `json:"account"`
		Balance []struct {
			Asset     string `json:"asset"`
			Available string `json:"available"`
			Reserved  string `json:"reserved"`
			Total     string `json:"total"`
		} `json:"balance"`
	}

	openOrdersRsp struct {
		Orders struct {
			Page     int `json:"page"`
			Pagesize int `json:"pagesize"`
			Result   []struct {
				Createtime     string `json:"createtime"`
				Filledamount   string `json:"filledamount"`
				Filledquantity string `json:"filledquantity"`
				Lastmodified   string `json:"lastmodified"`
				Orderid        string `json:"orderid,omitempty"`
				Orderquantity  string `json:"orderquantity"`
				Orderstatus    string `json:"orderstatus"`
				Price          string `json:"price"`
				Symbol         string `json:"symbol"`
				Type           string `json:"type"`
				OrderID        string `json:"orderId,omitempty"`
			} `json:"result"`
			Totalcount int `json:"totalcount"`
		} `json:"orders"`
		rsp
	}
)
