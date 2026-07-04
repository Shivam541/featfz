package config

import (
	"strings"
	"testing"
	"time"
)

func TestLoadFromLookup(t *testing.T) {
	base := map[string]string{
		"DB_HOST":     "127.0.0.1",
		"DB_USER":     "feat_manager",
		"DB_PASSWORD": "feat_manager",
		"DB_NAME":     "feat_manager",
	}

	tests := []struct {
		name      string
		env       map[string]string
		wantErr   string
		assertCfg func(t *testing.T, cfg Config)
	}{
		{
			name: "loads config with defaults",
			env:  base,
			assertCfg: func(t *testing.T, cfg Config) {
				t.Helper()

				if cfg.AppEnv != "development" {
					t.Fatalf("expected development app env, got %q", cfg.AppEnv)
				}
				if cfg.HTTPAddr != ":8080" {
					t.Fatalf("expected default http addr, got %q", cfg.HTTPAddr)
				}
				if cfg.DBPort != "3306" {
					t.Fatalf("expected default db port, got %q", cfg.DBPort)
				}
				if cfg.DBMaxOpenConns != 10 {
					t.Fatalf("expected default max open conns, got %d", cfg.DBMaxOpenConns)
				}
				if cfg.DBMaxIdleConns != 5 {
					t.Fatalf("expected default max idle conns, got %d", cfg.DBMaxIdleConns)
				}
				if cfg.DBConnMaxLifetime != 5*time.Minute {
					t.Fatalf("expected default conn lifetime, got %s", cfg.DBConnMaxLifetime)
				}
			},
		},
		{
			name: "loads config with overrides",
			env: map[string]string{
				"APP_ENV":              "test",
				"HTTP_ADDR":            "127.0.0.1:9090",
				"DB_HOST":              "db",
				"DB_PORT":              "4406",
				"DB_USER":              "another",
				"DB_PASSWORD":          "secret",
				"DB_NAME":              "flagdb",
				"DB_MAX_OPEN_CONNS":    "20",
				"DB_MAX_IDLE_CONNS":    "8",
				"DB_CONN_MAX_LIFETIME": "30s",
			},
			assertCfg: func(t *testing.T, cfg Config) {
				t.Helper()

				if cfg.AppEnv != "test" {
					t.Fatalf("expected test app env, got %q", cfg.AppEnv)
				}
				if cfg.HTTPAddr != "127.0.0.1:9090" {
					t.Fatalf("expected override http addr, got %q", cfg.HTTPAddr)
				}
				if cfg.DBPort != "4406" {
					t.Fatalf("expected override db port, got %q", cfg.DBPort)
				}
				if cfg.DBMaxOpenConns != 20 {
					t.Fatalf("expected max open conns 20, got %d", cfg.DBMaxOpenConns)
				}
				if cfg.DBMaxIdleConns != 8 {
					t.Fatalf("expected max idle conns 8, got %d", cfg.DBMaxIdleConns)
				}
				if cfg.DBConnMaxLifetime != 30*time.Second {
					t.Fatalf("expected conn lifetime 30s, got %s", cfg.DBConnMaxLifetime)
				}
			},
		},
		{
			name: "fails when required config missing",
			env: map[string]string{
				"DB_PORT": "3306",
			},
			wantErr: "missing required config: DB_HOST, DB_USER, DB_PASSWORD, DB_NAME",
		},
		{
			name: "fails on invalid int",
			env: map[string]string{
				"DB_HOST":           "127.0.0.1",
				"DB_USER":           "feat_manager",
				"DB_PASSWORD":       "feat_manager",
				"DB_NAME":           "feat_manager",
				"DB_MAX_OPEN_CONNS": "abc",
			},
			wantErr: "DB_MAX_OPEN_CONNS must be a valid integer",
		},
		{
			name: "fails on invalid duration",
			env: map[string]string{
				"DB_HOST":              "127.0.0.1",
				"DB_USER":              "feat_manager",
				"DB_PASSWORD":          "feat_manager",
				"DB_NAME":              "feat_manager",
				"DB_CONN_MAX_LIFETIME": "later",
			},
			wantErr: "DB_CONN_MAX_LIFETIME must be a valid duration",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg, err := LoadFromLookup(mapLookup(tt.env))
			if tt.wantErr != "" {
				if err == nil {
					t.Fatalf("expected error %q, got nil", tt.wantErr)
				}
				if !strings.Contains(err.Error(), tt.wantErr) {
					t.Fatalf("expected error containing %q, got %q", tt.wantErr, err.Error())
				}
				return
			}

			if err != nil {
				t.Fatalf("expected nil error, got %v", err)
			}

			if tt.assertCfg != nil {
				tt.assertCfg(t, cfg)
			}
		})
	}
}

func mapLookup(values map[string]string) func(string) (string, bool) {
	return func(key string) (string, bool) {
		value, ok := values[key]
		return value, ok
	}
}
