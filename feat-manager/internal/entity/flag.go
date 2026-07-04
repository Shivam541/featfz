package entity

import "time"

type Flag struct {
	ID             int64      `gorm:"column:id;primaryKey;autoIncrement"`
	TenantID       int64      `gorm:"column:tenant_id"`
	Key            string     `gorm:"column:key"`
	Description    string     `gorm:"column:description"`
	DefaultEnabled bool       `gorm:"column:default_enabled"`
	ArchivedAt     *time.Time `gorm:"column:archived_at"`
	CreatedAt      time.Time  `gorm:"column:created_at"`
	UpdatedAt      time.Time  `gorm:"column:updated_at"`
}

func (Flag) TableName() string {
	return "flags"
}
