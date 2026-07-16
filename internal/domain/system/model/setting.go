package model

import shared "erp/internal/shared/model"

type SystemSetting struct {
	shared.BaseModel
	GroupName    string `json:"groupName" gorm:"type:varchar(64);not null;default:system;index"`
	SettingKey   string `json:"settingKey" gorm:"type:varchar(128);not null;index"`
	SettingValue string `json:"settingValue" gorm:"type:text"`
	ValueType    string `json:"valueType" gorm:"type:varchar(32);not null;default:string"`
	IsPublic     bool   `json:"isPublic" gorm:"not null;default:false"`
}

func (SystemSetting) TableName() string { return "system_settings" }
