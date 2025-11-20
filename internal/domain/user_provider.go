package domain

import "time"

// UserProvider stores external identity provider linkage for a user.
type UserProvider struct {
	ID             string                 `gorm:"type:uuid;default:uuid_generate_v4();primaryKey" json:"id"`
	ProviderType   string                 `gorm:"type:text;not null" json:"provider_type"`
	ProviderUserID string                 `gorm:"type:text;not null" json:"provider_user_id"`
	UserID         string                 `gorm:"type:uuid;not null;index" json:"user_id"`
	Metadata       map[string]interface{} `gorm:"type:jsonb" json:"metadata"`
	CreatedAt      time.Time              `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt      time.Time              `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
}

func (UserProvider) TableName() string {
	return "user_provider"
}
