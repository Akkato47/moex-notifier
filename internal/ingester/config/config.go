package config

import "time"

type Config struct {
	LogLevel      string        `env:"LOG_LEVEL"        envDefault:"info"                        validate:"oneof=debug info warn error"`
	LogPretty     bool          `env:"LOG_PRETTY"       envDefault:"false"`
	HealthAddr    string        `env:"HEALTH_ADDR"      envDefault:":8080"                       validate:"required"`
	ShutdownDelay time.Duration `env:"SHUTDOWN_DELAY"   envDefault:"15s"`
	PostgresDSN   string        `env:"POSTGRES_DSN"     envDefault:"postgres://moex:moex@localhost:5432/moex" validate:"required"`
	ClickHouseAddr string       `env:"CLICKHOUSE_ADDR"  envDefault:"localhost:9000" validate:"required"`
	ClickHouseUser string       `env:"CLICKHOUSE_USER"  envDefault:"moex"          validate:"required"`
	ClickHousePass string       `env:"CLICKHOUSE_PASS"  envDefault:"moex"`
	MoexBaseURL   string        `env:"MOEX_BASE_URL"    envDefault:"https://iss.moex.com/iss"    validate:"required"`
	PollInterval  time.Duration `env:"POLL_INTERVAL"    envDefault:"1m"`
}
