package domain

import "time"

type UserProfile struct {
	ID          string    `gorm:"type:uuid;default:uuid_generate_v4();primaryKey" json:"id"`
	UserID      string    `gorm:"type:uuid;not null;uniqueIndex" json:"user_id"`
	DisplayName *string   `gorm:"column:display_name" json:"display_name"`
	AvatarURL   *string   `gorm:"column:avatar_url" json:"avatar_url"`
	CreatedAt   time.Time `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt   time.Time `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
}

func (UserProfile) TableName() string {
	return "user_profile"
}

func (p *UserProfile) Update(displayName, avatarURL *string) {
	if displayName != nil {
		p.DisplayName = displayName
	}
	if avatarURL != nil {
		p.AvatarURL = avatarURL
	}
}
