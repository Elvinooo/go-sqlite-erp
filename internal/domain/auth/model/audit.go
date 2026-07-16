package model

import "time"

type LoginLog struct {
	ID        uint64    `json:"id" gorm:"primaryKey;autoIncrement"`
	TenantID  uint64    `json:"tenantId" gorm:"not null;default:1;index"`
	Username  string    `json:"username" gorm:"type:varchar(64);not null;index"`
	UserID    *uint64   `json:"userId" gorm:"index"`
	IP        string    `json:"ip" gorm:"type:varchar(64);index"`
	UserAgent string    `json:"userAgent" gorm:"type:varchar(255)"`
	Success   bool      `json:"success" gorm:"not null;default:false;index"`
	Message   string    `json:"message" gorm:"type:varchar(255)"`
	CreatedAt time.Time `json:"createdAt" gorm:"not null;index"`
}

func (LoginLog) TableName() string { return "login_logs" }

type OperationLog struct {
	ID          uint64    `json:"id" gorm:"primaryKey;autoIncrement"`
	TenantID    uint64    `json:"tenantId" gorm:"not null;default:1;index"`
	UserID      *uint64   `json:"userId" gorm:"index"`
	Username    string    `json:"username" gorm:"type:varchar(64);index"`
	Module      string    `json:"module" gorm:"type:varchar(64);index"`
	Action      string    `json:"action" gorm:"type:varchar(64);index"`
	Method      string    `json:"method" gorm:"type:varchar(16);not null"`
	Path        string    `json:"path" gorm:"type:varchar(255);not null;index"`
	IP          string    `json:"ip" gorm:"type:varchar(64);index"`
	UserAgent   string    `json:"userAgent" gorm:"type:varchar(255)"`
	RequestBody string    `json:"requestBody" gorm:"type:text"`
	StatusCode  int       `json:"statusCode" gorm:"not null;default:0"`
	CostMS      int64     `json:"costMs" gorm:"not null;default:0"`
	CreatedAt   time.Time `json:"createdAt" gorm:"not null;index"`
}

func (OperationLog) TableName() string { return "operation_logs" }
