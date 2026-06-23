package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
	"github.com/Akkato47/moex-notifier/internal/processor/features/candle/domain"
)

type CandleReader struct {
	conn driver.Conn
}

func NewCandleReader(conn driver.Conn) *CandleReader {
	return &CandleReader{conn: conn}
}

func (r *CandleReader) GetCandlesAfter(ctx context.Context, ticker string, after time.Time) ([]domain.Candle, error) {
	rows, err := r.conn.Query(ctx,
		`SELECT ticker, timeframe, ts, open, high, low, close, volume
		 FROM market_candles
		 WHERE ticker = ? AND timeframe = '1' AND ts > ?
		 ORDER BY ts ASC`,
		ticker, after,
	)
	if err != nil {
		return nil, fmt.Errorf("query candles: %w", err)
	}
	defer rows.Close()

	var candles []domain.Candle
	for rows.Next() {
		var c domain.Candle
		if err := rows.Scan(&c.Ticker, &c.Timeframe, &c.Ts, &c.Open, &c.High, &c.Low, &c.Close, &c.Volume); err != nil {
			return nil, fmt.Errorf("scan candle: %w", err)
		}
		candles = append(candles, c)
	}
	return candles, rows.Err()
}
