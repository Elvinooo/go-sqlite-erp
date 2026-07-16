package model

import shared "erp/internal/shared/model"

type Menu struct {
	shared.BaseModel
	ParentID       *uint64 `json:"parentId" gorm:"index"`
	Name           string  `json:"name" gorm:"type:varchar(64);not null;index"`
	Title          string  `json:"title" gorm:"type:varchar(64);not null"`
	Path           string  `json:"path" gorm:"type:varchar(255);index"`
	Component      string  `json:"component" gorm:"type:varchar(255)"`
	Icon           string  `json:"icon" gorm:"type:varchar(64)"`
	Type           string  `json:"type" gorm:"type:varchar(20);not null;default:menu;index"`
	PermissionCode string  `json:"permissionCode" gorm:"type:varchar(128);index"`
	Sort           int     `json:"sort" gorm:"not null;default:0;index"`
	Visible        bool    `json:"visible" gorm:"not null;default:true"`
	Status         string  `json:"status" gorm:"type:varchar(20);not null;default:active;index"`
	Children       []Menu  `json:"children" gorm:"-"`
}

func (Menu) TableName() string { return "menus" }
