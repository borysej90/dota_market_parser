package dota_market_notifier

import "fmt"

type TradeLot struct {
	Name  string
	Price float32
}

func (tl TradeLot) String() string {
	return fmt.Sprintf("Name: %s; Price: %.2f", tl.Name, tl.Price)
}
