package config

import (
	"fmt"
	"os"

	"github.com/caarlos0/env/v11"
	"github.com/go-playground/validator/v10"
	"github.com/joho/godotenv"
)

func Load[T any](cfg *T) error {
	if os.Getenv("APP_ENV") == "local" {
		if err := godotenv.Load(); err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("load .env: %w", err)
		}
	}

	if err := env.Parse(cfg); err != nil {
		return fmt.Errorf("parse env: %w", err)
	}
	cfgValidator := validator.New()

	if err := cfgValidator.Struct(cfg); err != nil {
		return fmt.Errorf("validate config: %w", err)
	}
	return nil
}
