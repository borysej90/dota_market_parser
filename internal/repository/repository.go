package repository

import "context"

type Repo interface {
	GetAllItems(ctx context.Context) ([]string, error)
}
