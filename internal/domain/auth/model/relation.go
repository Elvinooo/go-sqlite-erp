package model

import "time"

type UserRole struct {
	UserID    uint64    `gorm:"primaryKey;index"`
	RoleID    uint64    `gorm:"primaryKey;index"`
	CreatedAt time.Time `gorm:"not null"`
}

func (UserRole) TableName() string { return "user_roles" }

type RoleMenu struct {
	RoleID    uint64    `gorm:"primaryKey;index"`
	MenuID    uint64    `gorm:"primaryKey;index"`
	CreatedAt time.Time `gorm:"not null"`
}

func (RoleMenu) TableName() string { return "role_menus" }

type RolePermission struct {
	RoleID       uint64    `gorm:"primaryKey;index"`
	PermissionID uint64    `gorm:"primaryKey;index"`
	CreatedAt    time.Time `gorm:"not null"`
}

func (RolePermission) TableName() string { return "role_permissions" }
