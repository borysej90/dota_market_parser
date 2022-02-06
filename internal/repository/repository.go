package repository

import (
	"context"

	dmn "dota_market_notifier"
)

type Repo interface {
	GetAllItemsNames(ctx context.Context) ([]string, error)
	UpdateItemsHistory(ctx context.Context, items []dmn.TradeLot) error
}
