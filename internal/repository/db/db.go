//go:generate goderive .
package db

import (
	"context"
	"dota_market_notifier/internal/repository"
	"fmt"
	"github.com/jmoiron/sqlx"
	"time"
)

var _ repository.Repo = &Repo{}

const ItemsTableName = "items"

type ItemRecord struct {
	ID        int       `db:"id"`
	Name      string    `db:"name"`
	CreatedAt time.Time `db:"created_at"`
}

func NewDB(host string, port int, user, password, dbName string) *sqlx.DB {
	return sqlx.MustConnect("postgres", fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s",
		user, password, host, port, dbName,
	))
}

type Repo struct {
	db *sqlx.DB
}

func NewRepo(db *sqlx.DB) *Repo {
	return &Repo{db: db}
}

func (r *Repo) GetAllItems(ctx context.Context) ([]string, error) {
	stmt := fmt.Sprintf("SELECT name FROM %s", ItemsTableName)
	records := make([]*ItemRecord, 0)
	if err := r.db.SelectContext(ctx, &records, stmt); err != nil {
		return nil, err
	}
	names := deriveFmapItemsNames(func(item *ItemRecord) string {
		return item.Name
	}, records)
	return names, nil
}
