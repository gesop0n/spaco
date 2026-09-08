package config

import (
	"fmt"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

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
		return fmt.Errorf("invalid environment: %q", value)
	}
}

type Config struct {
	Database       DatabaseConfig
	App            AppConfig
	Server         ServerConfig
	Authentication AuthenticationConfig
}

type DatabaseConfig struct {
	URL string `env:"DATABASE_URL,required,notEmpty"`
}

type AppConfig struct {
	Environment Environment `env:"ENV,required,notEmpty"`
}

type ServerConfig struct {
	Address         string        `env:"SERVER_ADDRESS"`
	AllowedOrigins  []string      `env:"CORS_ALLOWED_ORIGINS" envDefault:"http://localhost:5173"`
	ShutdownTimeout time.Duration `env:"SHUTDOWN_TIMEOUT" envDefault:"10s"`
}

type AuthenticationConfig struct {
	SupabaseURL string        `env:"SUPABASE_URL,required,notEmpty"`
	Issuer      string        `env:"SUPABASE_JWT_ISSUER"`
	JWKSURL     string        `env:"SUPABASE_JWKS_URL"`
	Audience    string        `env:"SUPABASE_JWT_AUDIENCE" envDefault:"authenticated"`
	ClockSkew   time.Duration `env:"SUPABASE_JWT_CLOCK_SKEW" envDefault:"30s"`
}

// IssuerURLは明示設定を優先し、未設定ならSupabase project URLから組み立てる。
func (c AuthenticationConfig) IssuerURL() string {
	if c.Issuer != "" {
		return strings.TrimRight(c.Issuer, "/")
	}
	return strings.TrimRight(c.SupabaseURL, "/") + "/auth/v1"
}

// JSONWebKeySetURLは、JWT署名検証に使用する公開鍵のURLを返す。
func (c AuthenticationConfig) JSONWebKeySetURL() string {
	if c.JWKSURL != "" {
		return c.JWKSURL
	}
	return strings.TrimRight(c.SupabaseURL, "/") + "/auth/v1/.well-known/jwks.json"
}

func Load() (Config, error) {
	cfg, err := env.ParseAs[Config]()
	if err != nil {
		return Config{}, fmt.Errorf("load config: %w", err)
	}
	cfg.Authentication.SupabaseURL = strings.TrimSpace(cfg.Authentication.SupabaseURL)
	cfg.Authentication.Issuer = strings.TrimSpace(cfg.Authentication.Issuer)
	cfg.Authentication.JWKSURL = strings.TrimSpace(cfg.Authentication.JWKSURL)
	serverAddress, err := resolveServerAddress(cfg.Server.Address, os.Getenv("PORT"))
	if err != nil {
		return Config{}, fmt.Errorf("load config: %w", err)
	}
	cfg.Server.Address = serverAddress
	for index := range cfg.Server.AllowedOrigins {
		cfg.Server.AllowedOrigins[index] = strings.TrimSpace(cfg.Server.AllowedOrigins[index])
	}
	if err := validate(cfg); err != nil {
		return Config{}, fmt.Errorf("load config: %w", err)
	}

	return cfg, nil
}

func resolveServerAddress(address, port string) (string, error) {
	if address != "" {
		return address, nil
	}
	port = strings.TrimSpace(port)
	if port == "" {
		return ":8080", nil
	}
	portNumber, err := strconv.Atoi(port)
	if err != nil || portNumber < 1 || portNumber > 65535 {
		return "", fmt.Errorf("PORT must be a number between 1 and 65535")
	}
	return fmt.Sprintf(":%d", portNumber), nil
}

func validate(cfg Config) error {
	if strings.TrimSpace(cfg.Server.Address) == "" {
		return fmt.Errorf("SERVER_ADDRESS must not be empty")
	}
	if len(cfg.Server.AllowedOrigins) == 0 {
		return fmt.Errorf("CORS_ALLOWED_ORIGINS must contain at least one origin")
	}
	if cfg.Server.ShutdownTimeout <= 0 {
		return fmt.Errorf("SHUTDOWN_TIMEOUT must be positive")
	}
	if cfg.Authentication.ClockSkew < 0 {
		return fmt.Errorf("SUPABASE_JWT_CLOCK_SKEW must not be negative")
	}
	if err := validateHTTPURL("SUPABASE_URL", cfg.Authentication.SupabaseURL); err != nil {
		return err
	}
	if err := validateHTTPURL("Supabase issuer", cfg.Authentication.IssuerURL()); err != nil {
		return err
	}
	if err := validateHTTPURL("Supabase JWKS URL", cfg.Authentication.JSONWebKeySetURL()); err != nil {
		return err
	}
	for _, origin := range cfg.Server.AllowedOrigins {
		if err := validateOrigin(origin); err != nil {
			return err
		}
	}
	return nil
}

func validateHTTPURL(name, value string) error {
	parsed, err := url.ParseRequestURI(value)
	if err != nil || parsed.Host == "" || (parsed.Scheme != "http" && parsed.Scheme != "https") {
		return fmt.Errorf("%s must be an http(s) URL", name)
	}
	return nil
}

func validateOrigin(value string) error {
	parsed, err := url.ParseRequestURI(value)
	if err != nil || parsed.Host == "" || (parsed.Scheme != "http" && parsed.Scheme != "https") {
		return fmt.Errorf("CORS origin %q must be an http(s) origin", value)
	}
	if parsed.Path != "" || parsed.RawQuery != "" || parsed.Fragment != "" {
		return fmt.Errorf("CORS origin %q must not contain a path, query, or fragment", value)
	}
	return nil
}
