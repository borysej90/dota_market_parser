package monobank

import (
	"encoding/json"
	"github.com/pkg/errors"
	"net/http"
	"time"

	"dota_market_notifier/internal/currency"
)

var _ currency.Parser = &MonoParser{}

const MonoURL = "https://api.monobank.ua/bank/currency"

type currencyData struct {
	CodeA    int     `json:"currencyCodeA,omitempty"`
	CodeB    int     `json:"currencyCodeB,omitempty"`
	RateBuy  float32 `json:"rateBuy,omitempty"`
	RateSell float32 `json:"rateSell,omitempty"`
}

type MonoParser struct {
	rates    map[int]float32
	validTil time.Time
}

func New() *MonoParser {
	return &MonoParser{rates: make(map[int]float32)}
}

func (m *MonoParser) GetCurrencyRate(currencyName string) (float32, error) {
	if m.rates == nil || m.validTil.Before(time.Now()) {
		if err := m.fetchAPI(); err != nil {
			return 0, err
		}
	}
	currencyCode, ok := currency.Currency[currencyName]
	if !ok {
		return 0, errors.Errorf("unsupported currency: %s", currencyName)
	}
	return m.rates[currencyCode], nil
}

func (m *MonoParser) fetchAPI() error {
	client := http.Client{Timeout: 30 * time.Second}
	resp, err := client.Get(MonoURL)
	if err != nil {
		return err
	}
	var res []currencyData
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return err
	}
	for _, curr := range res {
		if curr.CodeB != currency.Currency["UAH"] {
			continue
		}
		switch curr.CodeA {
		case currency.Currency["USD"], currency.Currency["EUR"], currency.Currency["RUB"]:
			m.rates[curr.CodeA] = curr.RateSell
		}
	}
	m.validTil = time.Now().Add(5 * time.Minute)
	return nil
}
