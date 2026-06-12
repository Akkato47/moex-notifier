package logger

import (
	"fmt"
	"io"
	"os"

	"github.com/rs/zerolog"
)

type Config struct {
	Level       string
	ServiceName string
	LogPretty   bool
}

func NewLogger(cfg Config) (*zerolog.Logger, error) {
	level, err := zerolog.ParseLevel(cfg.Level)
	if err != nil {
		return nil, fmt.Errorf("parse log level: %w", err)
	}

	var w io.Writer = os.Stdout
	if cfg.LogPretty {
		w = newConsoleWriter(os.Stdout)
	}

	logger := zerolog.New(w).
		Level(level).
		With().
		Str("service", cfg.ServiceName).
		Timestamp().
		Caller().
		Logger()

	return &logger, nil
}

func newConsoleWriter(out io.Writer) zerolog.ConsoleWriter {
	return zerolog.ConsoleWriter{
		Out:        out,
		TimeFormat: "2006-01-02T15-04-05.000000",
		NoColor:    false,
	}
}
