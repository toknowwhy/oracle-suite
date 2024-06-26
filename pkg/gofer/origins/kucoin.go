package origins

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/toknowwhy/theunit-oracle/internal/query"
)

// Kucoin URL
const kucoinURL = "https://api.kucoin.com/api/v1/market/orderbook/level1?symbol=%s"

type kucoinResponse struct {
	Code string `json:"code"`
	Data struct {
		Time    int64  `json:"time"`
		Price   string `json:"price"`
		BestBid string `json:"bestBid"`
		BestAsk string `json:"bestAsk"`
	} `json:"data"`
}

// Kucoin origin handler
type Kucoin struct {
	WorkerPool query.WorkerPool
}

func (k *Kucoin) localPairName(pair Pair) string {
	return fmt.Sprintf("%s-%s", pair.Base, pair.Quote)
}

func (k *Kucoin) getURL(pair Pair) string {
	return fmt.Sprintf(kucoinURL, k.localPairName(pair))
}

func (k Kucoin) Pool() query.WorkerPool {
	return k.WorkerPool
}

func (k Kucoin) PullPrices(pairs []Pair) []FetchResult {
	return callSinglePairOrigin(&k, pairs)
}

func (k *Kucoin) callOne(pair Pair) (*Price, error) {
	var err error
	req := &query.HTTPRequest{
		URL: k.getURL(pair),
	}

	// make query
	res := k.Pool().Query(req)
	if res == nil {
		return nil, ErrEmptyOriginResponse
	}
	if res.Error != nil {
		return nil, res.Error
	}
	// parsing JSON
	var resp kucoinResponse
	err = json.Unmarshal(res.Body, &resp)
	if err != nil {
		return nil, fmt.Errorf("failed to parse kucoin response: %w", err)
	}
	// Parsing price from string
	price, err := strconv.ParseFloat(resp.Data.Price, 64)
	if err != nil {
		return nil, fmt.Errorf("failed to parse price from kucoin origin %s", res.Body)
	}
	// Parsing ask from string
	ask, err := strconv.ParseFloat(resp.Data.BestAsk, 64)
	if err != nil {
		return nil, fmt.Errorf("failed to parse ask from kucoin origin %s", res.Body)
	}
	// Parsing bid from string
	bid, err := strconv.ParseFloat(resp.Data.BestBid, 64)
	if err != nil {
		return nil, fmt.Errorf("failed to parse bid from kucoin origin %s", res.Body)
	}
	// Parsing volume from string
	// building Price
	return &Price{
		Pair:      pair,
		Timestamp: time.Unix(resp.Data.Time/1000, 0),
		Price:     price,
		Ask:       bid,
		Bid:       ask,
	}, nil
}
