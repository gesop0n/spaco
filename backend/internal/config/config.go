package config

import (
	"fmt"

	"github.com/caarlos0/env/v11"
)

type Environment string

const (
	EnvironmentDevelopment Environment = "development"
	EnvironmentProduction  Environment = "production"
)

func (e *Environment) UnmarshalText(text []byte) error {
	value := Environment(text)

	switch value {
	case EnvironmentDevelopment, EnvironmentProduction:
		*e = value
		return nil
	default:
		return fmt.Errorf("Invalid environment: %q", value)
	}
}

type Config struct {
	Database DatabaseConfig
	App      AppConfig
}

type DatabaseConfig struct {
	URL string `env:"DATABASE_URL,required,notEmpty"`
}

type AppConfig struct {
	Environment Environment `env:"ENV,required,notEmpty"`
}

func Load() (Config, error) {
	cfg, err := env.ParseAs[Config]()
	if err != nil {
		return Config{}, fmt.Errorf("load config: %w", err)
	}

	return cfg, nil
}
