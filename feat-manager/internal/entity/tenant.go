package entity

type Tenant struct {
	ID        int64  `gorm:"column:id;primaryKey;autoIncrement"`
	Name      string `gorm:"column:name"`
	AppID     string `gorm:"column:app_id"`
	JWTSecret string `gorm:"column:jwt_secret"`
}

func (Tenant) TableName() string {
	return "tenants"
}
