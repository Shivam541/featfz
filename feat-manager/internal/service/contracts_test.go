package service

import (
	"context"
	"testing"
)

func TestStaticHealthChecker(t *testing.T) {
	status := StaticHealthChecker{}.Check(context.Background())

	if status.Status != "ok" {
		t.Fatalf("expected ok status, got %q", status.Status)
	}
}
