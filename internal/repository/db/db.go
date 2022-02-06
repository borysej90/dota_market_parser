//go:generate goderive .
package db

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"time"

	dmn "dota_market_notifier"
	"dota_market_notifier/internal/repository"
)

var _ repository.Repo = &Repo{}

const (
	ItemsTableName   = "items"
	HistoryTableName = "history"
)

type ItemRecord struct {
	ID   int    `db:"id"`
	Name string `db:"name"`

	// history related fields
	Price     float32   `db:"price"`
	CreatedAt time.Time `db:"created_at"`
}

func NewDB(host string, port int, user, password, dbName string) *sqlx.DB {
	return sqlx.MustConnect("postgres", fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s?sslmode=disable",
		user, password, host, port, dbName,
	))
}

type Repo struct {
	db *sqlx.DB
}

func NewRepo(db *sqlx.DB) *Repo {
	return &Repo{db: db}
}

func (r *Repo) GetAllItemsNames(ctx context.Context) ([]string, error) {
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

func (r *Repo) UpdateItemsHistory(ctx context.Context, lots []*dmn.TradeLot) error {
	tx, err := r.db.BeginTxx(ctx, &sql.TxOptions{})
	if err != nil {
		return err
	}
	defer tx.Rollback()
	records := deriveFmapTradeLotToRecord(func(lot *dmn.TradeLot) *ItemRecord {
		return &ItemRecord{
			Name:      lot.Name,
			Price:     lot.Price,
			CreatedAt: time.Now(),
		}
	}, lots)
	stmt := fmt.Sprintf(`
INSERT INTO %s
(
    item_id,
    price,
    quantity,
    created_at
) VALUES (
    (SELECT id FROM %s WHERE name = :name),
    :price,
    :quantity,
    :created_at
)`, HistoryTableName, ItemsTableName)
	prepared, err := tx.PrepareNamedContext(ctx, stmt)
	if err != nil {
		return err
	}
	prepared = tx.NamedStmtContext(ctx, prepared)
	for _, record := range records {
		if _, err := prepared.ExecContext(ctx, record); err != nil {
			return err
		}
	}
	return tx.Commit()
}
