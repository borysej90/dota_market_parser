package market

import (
	"context"
	dmn "dota_market_notifier"
	"dota_market_notifier/internal/currency"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/pkg/errors"
)

const (
	DotaMarketBaseURL = "https://market.dota2.net/api/v2"
	KeyQueryParamKey  = "key"
	ItemQueryParamKey = "list_hash_name[]"
)

type Market struct {
	currency currency.Parser
	apiKey   string

	// qps defines how many requests per second is allowed on the market.
	qps uint
	// itemsPerRequest defines how many items can be queried in one request to the market.
	itemsPerRequest uint
}

func New(currency currency.Parser, apiKey string) *Market {
	return &Market{
		currency:        currency,
		apiKey:          apiKey,
		qps:             5,
		itemsPerRequest: 50,
	}
}

func (m Market) GetLastPrices(ctx context.Context, itemNames []string) ([]dmn.TradeLot, error) {
	ret := make([]dmn.TradeLot, 0, len(itemNames))
	getNextChunk := divideItemsIntoChunks(itemNames, int(m.itemsPerRequest))
	for chunk, ok := getNextChunk(); ok; chunk, ok = getNextChunk() {
		startTime := time.Now()
		items, err := m.processRequest(ctx, m.deriveRawQueryWithItems(chunk))
		if err != nil {
			fmt.Printf("failed to process market request: %v\n", err)
		} else {
			ret = append(ret, items...)
		}
		timeSpent := time.Since(startTime)
		if delta := time.Second/time.Duration(m.qps) - timeSpent; delta > 0 {
			time.Sleep(delta)
		}
	}
	return ret, nil
}

func (m Market) processRequest(ctx context.Context, rawQuery string) ([]dmn.TradeLot, error) {
	req, err := http.NewRequestWithContext(
		ctx, http.MethodGet, DotaMarketBaseURL+"/search-list-items-by-hash-name-all", nil,
	)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create a request")
	}
	req.URL.RawQuery = rawQuery
	client := http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get last items prices")
	}
	defer resp.Body.Close()
	response := getItemsResponse{}
	if err = json.NewDecoder(resp.Body).Decode(&response); err != nil {
		fmt.Printf("failed to parse response: %v\n", err)
		return nil, errors.Wrap(err, "failed to parse response")
	}
	if !response.Success {
		fmt.Println("request was not successful, skipping")
		return nil, errors.New("request was not successful")
	}
	ret := make([]dmn.TradeLot, 0)
	for k, v := range response.Data {
		price, err := m.parsePrice(response.Currency, v[0].Price)
		if err != nil {
			fmt.Printf("failed to parse price: %v\n", err)
			continue
		}
		ret = append(ret, dmn.TradeLot{
			Name:  k,
			Price: price,
		})
	}
	return ret, nil
}

func (m Market) deriveRawQueryWithItems(itemNames []string) string {
	query := make(url.Values)
	for _, item := range itemNames {
		query.Add(ItemQueryParamKey, item)
	}
	query.Set(KeyQueryParamKey, m.apiKey)
	return query.Encode()
}

func (m Market) parsePrice(currency, priceStr string) (float32, error) {
	priceInt, err := strconv.Atoi(priceStr)
	if err != nil {
		return 0, errors.Wrap(err, "failed to parse price from string")
	}
	var price float32
	if currency == "RUB" {
		price = float32(priceInt) / 100
	} else {
		price = float32(priceInt) / 1000
	}
	rate, err := m.currency.GetCurrencyRate(currency)
	if err != nil {
		return 0, errors.Wrap(err, "failed to get currency rate")
	}
	return price * rate, nil
}

func divideItemsIntoChunks(items []string, chunkSize int) func() ([]string, bool) {
	nextPos := 0
	return func() ([]string, bool) {
		if nextPos >= len(items) {
			return nil, false
		}
		n := min(nextPos+chunkSize, len(items))
		defer func() {
			nextPos = n
		}()
		return items[nextPos:n], true
	}
}

func min(x, y int) int {
	if x < y {
		return x
	}
	return y
}
