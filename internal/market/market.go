package market

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	"net/http"
	"strings"
	"time"

	dmn "dota_market_notifier"
	"dota_market_notifier/internal/currency"
	"dota_market_notifier/internal/repository"
)

const Url = "https://market.dota2.net/api/v2"

type Parser struct {
	repo     repository.Repo
	currency currency.Parser
	apiKey   string
}

func New(repo repository.Repo, currencyParser currency.Parser, apiKey string) *Parser {
	return &Parser{
		repo:     repo,
		currency: currencyParser,
		apiKey:   apiKey,
	}
}

func (p *Parser) GetLastPrices(ctx context.Context) ([]*dmn.TradeLot, error) {
	names, err := p.repo.GetAllItemsNames(ctx)
	if err != nil {
		return nil, err
	}
	tradeLots := make([]*dmn.TradeLot, 0, len(names))
	for _, name := range names {
		tradeLot, err := p.getItemPrice(ctx, name)
		if err != nil {
			return nil, err
		}
		tradeLots = append(tradeLots, tradeLot)
	}
	return tradeLots, nil
}

func (p *Parser) getItemPrice(_ context.Context, name string) (*dmn.TradeLot, error) {
	client := http.Client{Timeout: 30 * time.Second}
	resp, err := client.Get(fmt.Sprintf(
		"%s/search-item-by-hash-name?key=%s&hash_name=%s",
		Url, p.apiKey, strings.ReplaceAll(name, " ", "%20"),
	))
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != 200 {
		return nil, errors.Errorf("failed to get info about '%s': %s", name, resp.Status)
	}
	var res response
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return nil, err
	}
	if !res.Success {
		return nil, errors.Errorf("unsuccessful request: %s", res.Error)
	}
	cheapestItem := res.Data[0]
	var price float32
	price, err = p.currency.GetCurrencyRate(res.Currency)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get currency rate for %s", res.Currency)
	}
	if res.Currency == "RUB" {
		// price variable already has currency rate written, so we can just multiply it
		// by actual price and get converted one
		price *= float32(cheapestItem.Price) / 100
	} else {
		price *= float32(cheapestItem.Price) / 1000
	}
	return &dmn.TradeLot{
		Name:  name,
		Price: price,
	}, nil
}

func (p *Parser) UpdateHistory(ctx context.Context, lots []dmn.TradeLot) error {
	return p.repo.UpdateItemsHistory(ctx, lots)
}

// response represents returned data from Dota 2 Market API endpoint.
type response struct {
	Success  bool       `json:"success,omitempty"`
	Currency string     `json:"currency,omitempty"`
	Data     []itemData `json:"data,omitempty"`
	Error    string     `json:"error"`
}

type itemData struct {
	Name  string `json:"market_hash_name,omitempty"`
	Price int    `json:"price,omitempty"`
	Count int    `json:"count,omitempty"`
}
