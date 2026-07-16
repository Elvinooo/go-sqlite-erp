package model

import shared "erp/internal/shared/model"

type Role struct {
	shared.BaseModel
	ParentID         *uint64      `json:"parentId" gorm:"index"`
	Code             string       `json:"code" gorm:"type:varchar(64);not null;index"`
	Name             string       `json:"name" gorm:"type:varchar(64);not null;index"`
	DataScope        string       `json:"dataScope" gorm:"type:varchar(32);not null;default:all;index"`
	DataScopeDeptIDs string       `json:"dataScopeDeptIds" gorm:"type:varchar(255)"`
	Sort             int          `json:"sort" gorm:"not null;default:0"`
	Status           string       `json:"status" gorm:"type:varchar(20);not null;default:active;index"`
	Parent           *Role        `json:"parent" gorm:"foreignKey:ParentID"`
	Menus            []Menu       `json:"menus" gorm:"many2many:role_menus;"`
	Permissions      []Permission `json:"permissions" gorm:"many2many:role_permissions;"`
}

func (Role) TableName() string { return "roles" }
