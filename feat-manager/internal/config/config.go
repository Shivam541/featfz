package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	AppEnv            string
	HTTPAddr          string
	DBHost            string
	DBPort            string
	DBUser            string
	DBPassword        string
	DBName            string
	DBMaxOpenConns    int
	DBMaxIdleConns    int
	DBConnMaxLifetime time.Duration
}

func Load() (Config, error) {
	return LoadFromLookup(os.LookupEnv)
}

func LoadFromLookup(lookup func(string) (string, bool)) (Config, error) {
	cfg := Config{
		AppEnv:            getEnv(lookup, "APP_ENV", "development"),
		HTTPAddr:          getEnv(lookup, "HTTP_ADDR", ":8080"),
		DBHost:            strings.TrimSpace(requiredEnv(lookup, "DB_HOST")),
		DBPort:            getEnv(lookup, "DB_PORT", "3306"),
		DBUser:            strings.TrimSpace(requiredEnv(lookup, "DB_USER")),
		DBPassword:        requiredEnv(lookup, "DB_PASSWORD"),
		DBName:            strings.TrimSpace(requiredEnv(lookup, "DB_NAME")),
		DBMaxOpenConns:    10,
		DBMaxIdleConns:    5,
		DBConnMaxLifetime: 5 * time.Minute,
	}

	if err := requireFields(cfg); err != nil {
		return Config{}, err
	}

	maxOpen, err := intFromEnv(lookup, "DB_MAX_OPEN_CONNS", cfg.DBMaxOpenConns)
	if err != nil {
		return Config{}, err
	}
	maxIdle, err := intFromEnv(lookup, "DB_MAX_IDLE_CONNS", cfg.DBMaxIdleConns)
	if err != nil {
		return Config{}, err
	}
	connLifetime, err := durationFromEnv(lookup, "DB_CONN_MAX_LIFETIME", cfg.DBConnMaxLifetime)
	if err != nil {
		return Config{}, err
	}

	cfg.DBMaxOpenConns = maxOpen
	cfg.DBMaxIdleConns = maxIdle
	cfg.DBConnMaxLifetime = connLifetime

	return cfg, nil
}

func (c Config) DSN() string {
	return fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true&multiStatements=true",
		c.DBUser,
		c.DBPassword,
		c.DBHost,
		c.DBPort,
		c.DBName,
	)
}

func requireFields(cfg Config) error {
	missing := make([]string, 0, 4)

	if cfg.DBHost == "" {
		missing = append(missing, "DB_HOST")
	}
	if cfg.DBUser == "" {
		missing = append(missing, "DB_USER")
	}
	if cfg.DBPassword == "" {
		missing = append(missing, "DB_PASSWORD")
	}
	if cfg.DBName == "" {
		missing = append(missing, "DB_NAME")
	}

	if len(missing) > 0 {
		return fmt.Errorf("missing required config: %s", strings.Join(missing, ", "))
	}

	return nil
}

func requiredEnv(lookup func(string) (string, bool), key string) string {
	value, _ := lookup(key)
	return value
}

func getEnv(lookup func(string) (string, bool), key, fallback string) string {
	value, ok := lookup(key)
	if !ok || strings.TrimSpace(value) == "" {
		return fallback
	}

	return value
}

func intFromEnv(lookup func(string) (string, bool), key string, fallback int) (int, error) {
	value, ok := lookup(key)
	if !ok || strings.TrimSpace(value) == "" {
		return fallback, nil
	}

	parsed, err := strconv.Atoi(strings.TrimSpace(value))
	if err != nil {
		return 0, fmt.Errorf("%s must be a valid integer", key)
	}

	if parsed < 1 {
		return 0, fmt.Errorf("%s must be greater than zero", key)
	}

	return parsed, nil
}

func durationFromEnv(lookup func(string) (string, bool), key string, fallback time.Duration) (time.Duration, error) {
	value, ok := lookup(key)
	if !ok || strings.TrimSpace(value) == "" {
		return fallback, nil
	}

	parsed, err := time.ParseDuration(strings.TrimSpace(value))
	if err != nil {
		return 0, fmt.Errorf("%s must be a valid duration", key)
	}

	if parsed <= 0 {
		return 0, fmt.Errorf("%s must be greater than zero", key)
	}

	return parsed, nil
}
