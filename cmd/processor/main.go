package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	candleRepo "github.com/Akkato47/moex-notifier/internal/processor/features/candle/repository"
	"github.com/Akkato47/moex-notifier/internal/processor/features/matching/services"
	ruleRepo "github.com/Akkato47/moex-notifier/internal/processor/features/rule/repository"
	"github.com/Akkato47/moex-notifier/internal/platform/clickhouse"
	"github.com/Akkato47/moex-notifier/internal/platform/config"
	"github.com/Akkato47/moex-notifier/internal/platform/health"
	"github.com/Akkato47/moex-notifier/internal/platform/kafka"
	"github.com/Akkato47/moex-notifier/internal/platform/logger"
	"github.com/Akkato47/moex-notifier/internal/platform/postgres"
	processorConfig "github.com/Akkato47/moex-notifier/internal/processor/config"
	"golang.org/x/sync/errgroup"
)

func main() {
	cfg := processorConfig.Config{}
	if err := config.Load(&cfg); err != nil {
		fmt.Fprintf(os.Stderr, "load config: %v\n", err)
		os.Exit(1)
	}

	log, err := logger.NewLogger(logger.Config{
		Level:       cfg.LogLevel,
		ServiceName: "processor",
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

	producer, err := kafka.NewProducer(cfg.KafkaBrokers, cfg.KafkaTopic)
	if err != nil {
		fmt.Fprintf(os.Stderr, "init kafka producer: %v\n", err)
		os.Exit(1)
	}
	defer producer.Close()

	reader := candleRepo.NewCandleReader(chConn)
	rules := ruleRepo.NewRuleRepo(pgPool)
	matcher := services.NewMatcher(reader, rules, producer, log, cfg.PollInterval)

	healthSrv := health.New(cfg.HealthAddr, []health.NamedChecker{
		{Name: "postgres", Checker: func(ctx context.Context) error { return pgPool.Ping(ctx) }},
		{Name: "clickhouse", Checker: func(ctx context.Context) error { return clickhouse.Ping(ctx, chConn) }},
	}, log)

	eg, ctx := errgroup.WithContext(ctx)
	eg.Go(func() error { return healthSrv.Run(ctx) })
	eg.Go(func() error { return matcher.Run(ctx) })

	if err := eg.Wait(); err != nil && !errors.Is(err, context.Canceled) {
		log.Error().Err(err).Msg("processor exited with error")
	}
}
