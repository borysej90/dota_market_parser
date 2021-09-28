package dota_market_notifier

import "fmt"

type TradeLot struct {
	Name     string
	Price    float32
	Quantity int
}

func (tl TradeLot) String() string {
	return fmt.Sprintf("Name: %s; Price: %.2f; Quantity: %d", tl.Name, tl.Price, tl.Quantity)
}
