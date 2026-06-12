package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	ingesterConfig "github.com/Akkato47/moex-notifier/internal/ingester/config"
	"github.com/Akkato47/moex-notifier/internal/platform/config"
	"github.com/Akkato47/moex-notifier/internal/platform/health"
	"github.com/Akkato47/moex-notifier/internal/platform/logger"
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

	healthSrv := health.New(cfg.HealthAddr, nil, log)

	eg, ctx := errgroup.WithContext(ctx)
	eg.Go(func() error { return healthSrv.Run(ctx) })
	if err := eg.Wait(); err != nil && !errors.Is(err, context.Canceled) {
		log.Error().Err(err).Msg("ingester exited with error")
	}
}
