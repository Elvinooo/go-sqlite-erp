package model

import (
	"time"

	shared "erp/internal/shared/model"
)

type User struct {
	shared.BaseModel
	Username           string     `json:"username" gorm:"type:varchar(64);not null;index"`
	PasswordHash       string     `json:"-" gorm:"type:varchar(255);not null"`
	RealName           string     `json:"realName" gorm:"type:varchar(64);index"`
	Phone              string     `json:"phone" gorm:"type:varchar(32);index"`
	Email              string     `json:"email" gorm:"type:varchar(128)"`
	Avatar             string     `json:"avatar" gorm:"type:varchar(255)"`
	Status             string     `json:"status" gorm:"type:varchar(20);not null;default:active;index"`
	LastLoginAt        *time.Time `json:"lastLoginAt"`
	LastLoginIP        string     `json:"lastLoginIp" gorm:"type:varchar(64)"`
	MustChangePassword bool       `json:"mustChangePassword" gorm:"not null;default:false"`
	PasswordChangedAt  *time.Time `json:"passwordChangedAt"`
	PasswordVersion    int        `json:"-" gorm:"not null;default:0"`
	Roles              []Role     `json:"roles" gorm:"many2many:user_roles;"`
}

func (User) TableName() string { return "users" }

func (u User) IsActive() bool {
	return u.Status == "active"
}
