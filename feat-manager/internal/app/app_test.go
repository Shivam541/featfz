package app

import (
	"context"
	"database/sql"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/shivam/featfz/feat-manager/internal/config"
)

func TestNew(t *testing.T) {
	cfg := config.Config{
		HTTPAddr:   ":8080",
		DBHost:     "127.0.0.1",
		DBPort:     "3306",
		DBUser:     "feat_manager",
		DBPassword: "feat_manager",
		DBName:     "feat_manager",
	}

	t.Run("returns runtime when dependencies initialize", func(t *testing.T) {
		runtime, err := New(context.Background(), cfg, Dependencies{
			OpenDB: func(context.Context, config.Config) (*sql.DB, error) {
				return &sql.DB{}, nil
			},
			Logger: slog.New(slog.NewTextHandler(io.Discard, nil)),
		})
		if err != nil {
			t.Fatalf("expected nil error, got %v", err)
		}

		req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
		rec := httptest.NewRecorder()

		runtime.Handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", rec.Code)
		}
	})

	t.Run("fails fast when db dependency fails", func(t *testing.T) {
		_, err := New(context.Background(), cfg, Dependencies{
			OpenDB: func(context.Context, config.Config) (*sql.DB, error) {
				return nil, errors.New("db down")
			},
			Logger: slog.New(slog.NewTextHandler(io.Discard, nil)),
		})
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if err.Error() != "open db: db down" {
			t.Fatalf("expected wrapped db error, got %q", err.Error())
		}
	})
}
