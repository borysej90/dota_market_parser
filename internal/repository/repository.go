package repository

import "context"

type Repo interface {
	GetAllItemsNames(ctx context.Context) ([]string, error)
}
