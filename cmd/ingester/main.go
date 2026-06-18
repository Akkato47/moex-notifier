package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	ingesterConfig "github.com/Akkato47/moex-notifier/internal/ingester/config"
	candleRepo "github.com/Akkato47/moex-notifier/internal/ingester/features/candle/repository"
	"github.com/Akkato47/moex-notifier/internal/ingester/features/candle/services"
	subscriptionRepo "github.com/Akkato47/moex-notifier/internal/ingester/features/subscription/repository"
	"github.com/Akkato47/moex-notifier/internal/platform/clickhouse"
	"github.com/Akkato47/moex-notifier/internal/platform/config"
	"github.com/Akkato47/moex-notifier/internal/platform/health"
	"github.com/Akkato47/moex-notifier/internal/platform/logger"
	"github.com/Akkato47/moex-notifier/internal/platform/postgres"
	"golang.org/x/sync/errgroup"
)

func main() {
	cfg := ingesterConfig.Config{}
	if err := config.Load(&cfg); err != nil {
		fmt.Fprintf(os.Stderr, "load config: %v\n", err)
		os.Exit(1)
	}

	log, err := logger.NewLogger(logger.Config{
		Level:       cfg.LogLevel,
		ServiceName: "ingester",
		LogPretty:   cfg.LogPretty,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "init logger: %v\n", err)
		os.Exit(1)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	pgPool, err := postgres.New(ctx, postgres.Config{DbUrl: cfg.PostgresDSN})
	if err != nil {
		fmt.Fprintf(os.Stderr, "init postgres: %v\n", err)
		os.Exit(1)
	}
	defer pgPool.Close()

	chConn, err := clickhouse.New(clickhouse.Config{
		Addr:     cfg.ClickHouseAddr,
		Username: cfg.ClickHouseUser,
		Password: cfg.ClickHousePass,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "init clickhouse: %v\n", err)
		os.Exit(1)
	}
	defer chConn.Close()

	tickerRepo := subscriptionRepo.New(pgPool)
	candleRepo := candleRepo.New(chConn)
	moexClient := services.NewMoexClient(&http.Client{}, cfg.MoexBaseURL)

	poller, err := services.NewPoller(moexClient, candleRepo, tickerRepo, log, cfg.PollInterval)
	if err != nil {
		fmt.Fprintf(os.Stderr, "init poller: %v\n", err)
		os.Exit(1)
	}

	healthSrv := health.New(cfg.HealthAddr, []health.NamedChecker{
		{Name: "postgres", Checker: func(ctx context.Context) error { return pgPool.Ping(ctx) }},
		{Name: "clickhouse", Checker: func(ctx context.Context) error { return clickhouse.Ping(ctx, chConn) }},
	}, log)

	eg, ctx := errgroup.WithContext(ctx)
	eg.Go(func() error { return healthSrv.Run(ctx) })
	eg.Go(func() error { return poller.Run(ctx) })

	if err := eg.Wait(); err != nil && !errors.Is(err, context.Canceled) {
		log.Error().Err(err).Msg("ingester exited with error")
	}
}
