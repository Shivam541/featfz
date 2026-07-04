package dao

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"strings"
	"testing"

	_ "github.com/go-sql-driver/mysql"
)

func openIntegrationDB(t *testing.T) *sql.DB {
	t.Helper()

	dsn := strings.TrimSpace(os.Getenv("TEST_DB_DSN"))
	if dsn == "" {
		t.Skip("set TEST_DB_DSN to run dao integration tests")
	}

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		t.Fatalf("open test db: %v", err)
	}

	if err := db.PingContext(context.Background()); err != nil {
		_ = db.Close()
		t.Fatalf("ping test db: %v", err)
	}

	t.Cleanup(func() {
		_ = db.Close()
	})

	resetIntegrationTables(t, db)

	return db
}

func resetIntegrationTables(t *testing.T, db *sql.DB) {
	t.Helper()

	for _, query := range []string{
		"DELETE FROM flag_user_overrides",
		"DELETE FROM flags",
		"DELETE FROM tenants",
	} {
		if _, err := db.ExecContext(context.Background(), query); err != nil {
			t.Fatalf("reset integration tables with %q: %v", query, err)
		}
	}
}

func insertIntegrationTenant(t *testing.T, db *sql.DB, name, appID, secret string) int64 {
	t.Helper()

	result, err := db.ExecContext(context.Background(), `
INSERT INTO tenants (name, app_id, jwt_secret)
VALUES (?, ?, ?)
`, name, appID, secret)
	if err != nil {
		t.Fatalf("insert tenant %s: %v", appID, err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		t.Fatalf("tenant last insert id: %v", err)
	}

	return id
}

func requireDuplicateError(t *testing.T, err error) {
	t.Helper()

	if err == nil {
		t.Fatal("expected duplicate key error, got nil")
	}

	if !strings.Contains(strings.ToLower(err.Error()), "duplicate") {
		t.Fatalf("expected duplicate key error, got %v", err)
	}
}

func assertRowCount(t *testing.T, db *sql.DB, table string, want int) {
	t.Helper()

	var got int
	if err := db.QueryRowContext(context.Background(), fmt.Sprintf("SELECT COUNT(*) FROM %s", table)).Scan(&got); err != nil {
		t.Fatalf("count rows for %s: %v", table, err)
	}

	if got != want {
		t.Fatalf("expected %d rows in %s, got %d", want, table, got)
	}
}
