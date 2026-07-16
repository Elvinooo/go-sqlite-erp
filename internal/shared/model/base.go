package model

import (
	"time"

	"gorm.io/gorm"
)

type BaseModel struct {
	ID        uint64         `json:"id" gorm:"primaryKey;autoIncrement"`
	TenantID  uint64         `json:"tenantId" gorm:"not null;default:1;index"`
	Remark    string         `json:"remark" gorm:"type:text"`
	CreatedBy *uint64        `json:"createdBy" gorm:"index"`
	UpdatedBy *uint64        `json:"updatedBy"`
	CreatedAt time.Time      `json:"createdAt" gorm:"not null;index"`
	UpdatedAt time.Time      `json:"updatedAt" gorm:"not null"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
}
