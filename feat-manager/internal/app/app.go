package app

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"

	"github.com/shivam/featfz/feat-manager/internal/config"
	httpapi "github.com/shivam/featfz/feat-manager/internal/http"
	"github.com/shivam/featfz/feat-manager/internal/mysql"
)

type Dependencies struct {
	OpenDB func(context.Context, config.Config) (*sql.DB, error)
}

type Runtime struct {
	DB      *sql.DB
	Handler http.Handler
}

func New(ctx context.Context, cfg config.Config, deps Dependencies) (*Runtime, error) {
	openDB := deps.OpenDB
	if openDB == nil {
		openDB = mysql.Open
	}

	db, err := openDB(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("open db: %w", err)
	}

	return &Runtime{
		DB:      db,
		Handler: httpapi.NewRouter(),
	}, nil
}

func (r *Runtime) Close() error {
	if r == nil || r.DB == nil {
		return nil
	}

	return r.DB.Close()
}
