package entity

import "time"

type FlagUserOverride struct {
	ID        int64     `gorm:"column:id;primaryKey;autoIncrement"`
	TenantID  int64     `gorm:"column:tenant_id"`
	FlagID    int64     `gorm:"column:flag_id"`
	UserID    string    `gorm:"column:user_id"`
	Enabled   bool      `gorm:"column:enabled"`
	CreatedAt time.Time `gorm:"column:created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at"`
}

func (FlagUserOverride) TableName() string {
	return "flag_user_overrides"
}
