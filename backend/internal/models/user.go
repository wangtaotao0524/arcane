package models

import (
	"time"
)

type User struct {
	Username              string      `json:"username" gorm:"uniqueIndex;not null"`
	PasswordHash          string      `json:"-" gorm:"column:password_hash"`
	DisplayName           *string     `json:"displayName,omitempty" gorm:"column:display_name"`
	Email                 *string     `json:"email,omitempty"`
	Roles                 StringSlice `json:"roles" gorm:"type:text;not null;default:'[]'"`
	RequirePasswordChange bool        `json:"requirePasswordChange" gorm:"column:require_password_change;default:false"`
	OidcSubjectId         *string     `json:"oidcSubjectId,omitempty" gorm:"column:oidc_subject_id"`
	LastLogin             *time.Time  `json:"lastLogin,omitempty" gorm:"column:last_login"`
	BaseModel
}

func (User) TableName() string {
	return "users_table"
}
