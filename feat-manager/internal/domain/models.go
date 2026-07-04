package domain

import "time"

type TenantApp struct {
	TenantID   int64
	TenantName string
	AppID      string
	JWTSecret  string
}

type Flag struct {
	ID             int64
	TenantID       int64
	Key            string
	Description    string
	DefaultEnabled bool
	ArchivedAt     *time.Time
}

type FlagUserOverride struct {
	ID        int64
	TenantID  int64
	FlagID    int64
	UserID    string
	Enabled   bool
	CreatedAt time.Time
	UpdatedAt time.Time
}
