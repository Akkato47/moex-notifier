package repository

import (
	"context"
	"fmt"

	"github.com/Akkato47/moex-notifier/internal/ingester/features/candle/domain"
	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
)

type CandleRepo struct {
	conn driver.Conn
}

func New(conn driver.Conn) *CandleRepo {
	return &CandleRepo{conn: conn}
}

func (r *CandleRepo) WriteCandles(ctx context.Context, candles []domain.Candle) error {
	batch, err := r.conn.PrepareBatch(ctx, "INSERT INTO market_candles (ticker, timeframe, ts, open, high, low, close, volume)")
	if err != nil {
		return fmt.Errorf("prepare batch: %w", err)
	}

	for _, c := range candles {
		if err := batch.Append(c.Ticker, c.Timeframe, c.Ts, c.Open, c.High, c.Low, c.Close, c.Volume); err != nil {
			return fmt.Errorf("append candle: %w", err)
		}
	}

	if err := batch.Send(); err != nil {
		return fmt.Errorf("send batch: %w", err)
	}

	return nil
}