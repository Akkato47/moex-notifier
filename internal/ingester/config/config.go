package config

import "time"

type Config struct {
	LogLevel      string        `env:"LOG_LEVEL"       envDefault:"info"  validate:"oneof=debug info warn error"`
	LogPretty     bool          `env:"LOG_PRETTY"      envDefault:"false"`
	HealthAddr    string        `env:"HEALTH_ADDR"     envDefault:":8080" validate:"required"`
	ShutdownDelay time.Duration `env:"SHUTDOWN_DELAY"  envDefault:"15s"`
}
