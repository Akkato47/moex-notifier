package services

import (
	"context"
	"fmt"
	"time"

	"github.com/Akkato47/moex-notifier/internal/ingester/features/candle/domain"
	candleRepo "github.com/Akkato47/moex-notifier/internal/ingester/features/candle/repository"
	subscriptionRepo "github.com/Akkato47/moex-notifier/internal/ingester/features/subscription/repository"
	lru "github.com/hashicorp/golang-lru/v2"
	"github.com/rs/zerolog"
	"golang.org/x/sync/errgroup"
)

const (
	lruSize       = 1000
	maxBatch      = 10
	flushInterval = 10 * time.Second
)

type Poller struct {
	fetcher      *MoexClient
	candleRepo   *candleRepo.CandleRepo
	tickerRepo   *subscriptionRepo.TickerRepo
	cache        *lru.Cache[string, struct{}]
	ch           chan domain.Candle
	log          *zerolog.Logger
	pollInterval time.Duration
}

func NewPoller(fetcher *MoexClient, candleRepo *candleRepo.CandleRepo, tickerRepo *subscriptionRepo.TickerRepo, log *zerolog.Logger, pollInterval time.Duration) (*Poller, error) {
	cache, err := lru.New[string, struct{}](lruSize)
	if err != nil {
		return nil, fmt.Errorf("create lru cache: %w", err)
	}

	return &Poller{
		fetcher:      fetcher,
		candleRepo:   candleRepo,
		tickerRepo:   tickerRepo,
		cache:        cache,
		ch:           make(chan domain.Candle, 100),
		log:          log,
		pollInterval: pollInterval,
	}, nil
}

func (p *Poller) Run(ctx context.Context) error {
	g, ctx := errgroup.WithContext(ctx)
	g.Go(func() error { return p.poll(ctx) })
	g.Go(func() error { return p.batch(ctx) })
	return g.Wait()
}

func (p *Poller) poll(ctx context.Context) error {
	ticker := time.NewTicker(p.pollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
			p.log.Debug().Msg("poll tick")
			tickers, err := p.tickerRepo.GetActiveTickers(ctx)
			if err != nil {
				p.log.Error().Err(err).Msg("get active tickers")
				continue
			}
			p.log.Debug().Strs("tickers", tickers).Msg("active tickers fetched")
			for _, t := range tickers {
				p.fetchAndSend(ctx, t)
			}
		}
	}
}

func (p *Poller) fetchAndSend(ctx context.Context, ticker string) {
	from := time.Now().Add(-5 * time.Minute)
	to := time.Now()
	candles, err := p.fetcher.GetCandlesByPeriod(ctx, ticker, from, to)
	if err != nil {
		p.log.Error().Err(err).Str("ticker", ticker).Msg("fetch candles")
		return
	}

	p.log.Debug().
		Str("ticker", ticker).
		Int("total", len(candles)).
		Time("from", from).
		Time("to", to).
		Msg("candles fetched")

	sent := 0
	for _, c := range candles {
		key := fmt.Sprintf("%s:%d", c.Ticker, c.Ts.Unix())
		if p.cache.Contains(key) {
			continue
		}
		p.cache.Add(key, struct{}{})
		select {
		case p.ch <- c:
			sent++
		case <-ctx.Done():
			return
		}
	}

	if sent > 0 {
		p.log.Info().Str("ticker", ticker).Int("new_candles", sent).Msg("candles queued")
	}
}

func (p *Poller) batch(ctx context.Context) error {
	timer := time.NewTimer(flushInterval)
	defer timer.Stop()

	var buf []domain.Candle

	flush := func(ctx context.Context) {
		if len(buf) == 0 {
			return
		}
		if err := p.candleRepo.WriteCandles(ctx, buf); err != nil {
			p.log.Error().Err(err).Int("count", len(buf)).Msg("write candles failed")
		} else {
			p.log.Info().Int("count", len(buf)).Msg("candles written to clickhouse")
		}
		buf = buf[:0]
	}

	for {
		select {
		case c := <-p.ch:
			buf = append(buf, c)
			if len(buf) >= maxBatch {
				flush(ctx)
			}
		case <-timer.C:
			flush(ctx)
			timer.Reset(flushInterval)
		case <-ctx.Done():
			flush(context.Background())
			return nil
		}
	}
}
