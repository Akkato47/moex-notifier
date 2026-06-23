package config

import "time"

type Config struct {
	LogLevel      string        `env:"LOG_LEVEL"       envDefault:"info"  validate:"oneof=debug info warn error"`
	LogPretty     bool          `env:"LOG_PRETTY"      envDefault:"false"`
	HealthAddr    string        `env:"HEALTH_ADDR"     envDefault:":8081" validate:"required"`
	ShutdownDelay time.Duration `env:"SHUTDOWN_DELAY"  envDefault:"15s"`
	PostgresDSN   string        `env:"POSTGRES_DSN"    envDefault:"postgres://moex:moex@localhost:5433/moex" validate:"required"`

	ClickHouseAddr string `env:"CLICKHOUSE_ADDR" envDefault:"localhost:9000" validate:"required"`
	ClickHouseUser string `env:"CLICKHOUSE_USER" envDefault:"moex"           validate:"required"`
	ClickHousePass string `env:"CLICKHOUSE_PASS" envDefault:"moex"`

	KafkaBrokers []string      `env:"KAFKA_BROKERS" envSeparator:"," envDefault:"localhost:9092" validate:"required"`
	KafkaTopic   string        `env:"KAFKA_TOPIC"   envDefault:"alerts.triggered"               validate:"required"`
	PollInterval time.Duration `env:"POLL_INTERVAL" envDefault:"1m"`
}
