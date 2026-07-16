package model

import shared "erp/internal/shared/model"

type Permission struct {
	shared.BaseModel
	Code   string `json:"code" gorm:"type:varchar(128);not null;index"`
	Name   string `json:"name" gorm:"type:varchar(128);not null;index"`
	Module string `json:"module" gorm:"type:varchar(64);not null;index"`
	Type   string `json:"type" gorm:"type:varchar(32);not null;default:api;index"`
	Method string `json:"method" gorm:"type:varchar(16)"`
	Path   string `json:"path" gorm:"type:varchar(255);index"`
	Status string `json:"status" gorm:"type:varchar(20);not null;default:active;index"`
}

func (Permission) TableName() string { return "permissions" }
