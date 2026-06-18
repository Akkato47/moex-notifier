package repository

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

type TickerRepo struct {
	pool *pgxpool.Pool
}

func New(pool *pgxpool.Pool) *TickerRepo {
	return &TickerRepo{pool: pool}
}

func (r *TickerRepo) GetActiveTickers(ctx context.Context) ([]string, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT DISTINCT ticker FROM alert_rules WHERE is_active = true ORDER BY ticker`,
	)
	if err != nil {
		return nil, fmt.Errorf("query active tickers: %w", err)
	}
	defer rows.Close()

	var tickers []string
	for rows.Next() {
		var ticker string
		if err := rows.Scan(&ticker); err != nil {
			return nil, fmt.Errorf("scan ticker: %w", err)
		}
		tickers = append(tickers, ticker)
	}

	return tickers, rows.Err()
}
