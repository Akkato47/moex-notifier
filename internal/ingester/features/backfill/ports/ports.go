package ports

import (
	"context"
	"time"

	"github.com/Akkato47/moex-notifier/internal/ingester/features/backfill/domain"
)

type CandleFetcher interface {
	GetCandlesByPeriod(ctx context.Context, ticker string, from, to time.Time) ([]domain.Candle, error)
}

type CandleSink interface {
	WriteCandles(ctx context.Context, candles []domain.Candle) error
}
