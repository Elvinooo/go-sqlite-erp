package model

import (
	"github.com/shopspring/decimal"

	shared "erp/internal/shared/model"
)

type Supplier struct {
	shared.BaseModel
	Code           string          `json:"code" gorm:"type:varchar(64);not null;index"`
	Name           string          `json:"name" gorm:"type:varchar(128);not null;index"`
	ContactName    string          `json:"contactName" gorm:"type:varchar(64)"`
	Phone          string          `json:"phone" gorm:"type:varchar(32);index"`
	Email          string          `json:"email" gorm:"type:varchar(128)"`
	TaxNo          string          `json:"taxNo" gorm:"type:varchar(64);index"`
	Address        string          `json:"address" gorm:"type:varchar(255)"`
	SupplierTypes  string          `json:"supplierTypes" gorm:"type:varchar(255);not null;default:商品供应商;index"`
	PayableBalance decimal.Decimal `json:"payableBalance" gorm:"type:decimal(18,4);not null;default:0;index"`
	Status         string          `json:"status" gorm:"type:varchar(20);not null;default:active;index"`
}

func (Supplier) TableName() string { return "suppliers" }
