package model

import (
	"github.com/shopspring/decimal"

	shared "erp/internal/shared/model"
)

type Customer struct {
	shared.BaseModel
	Code              string          `json:"code" gorm:"type:varchar(64);not null;index"`
	Name              string          `json:"name" gorm:"type:varchar(128);not null;index"`
	Type              string          `json:"type" gorm:"type:varchar(32);not null;default:company;index"`
	Level             string          `json:"level" gorm:"type:varchar(32);index"`
	Phone             string          `json:"phone" gorm:"type:varchar(32);index"`
	Email             string          `json:"email" gorm:"type:varchar(128)"`
	TaxNo             string          `json:"taxNo" gorm:"type:varchar(64);index"`
	Address           string          `json:"address" gorm:"type:varchar(255)"`
	PaymentMethod     string          `json:"paymentMethod" gorm:"type:varchar(32);not null;default:immediate;index"`
	CreditLimit       decimal.Decimal `json:"creditLimit" gorm:"type:decimal(18,4);not null;default:0"`
	CreditDays        int             `json:"creditDays" gorm:"not null;default:0"`
	BillingCycle      string          `json:"billingCycle" gorm:"type:varchar(32);not null;default:none;index"`
	PaymentDay        int             `json:"paymentDay" gorm:"not null;default:0"`
	ReceivableBalance decimal.Decimal `json:"receivableBalance" gorm:"type:decimal(18,4);not null;default:0;index"`
	Status            string          `json:"status" gorm:"type:varchar(20);not null;default:active;index"`
}

func (Customer) TableName() string { return "customers" }

type CustomerContact struct {
	shared.BaseModel
	CustomerID uint64 `json:"customerId" gorm:"not null;index"`
	Name       string `json:"name" gorm:"type:varchar(64);not null;index"`
	Phone      string `json:"phone" gorm:"type:varchar(32);index"`
	Email      string `json:"email" gorm:"type:varchar(128)"`
	Position   string `json:"position" gorm:"type:varchar(64)"`
	IsDefault  bool   `json:"isDefault" gorm:"not null;default:false;index"`
}

func (CustomerContact) TableName() string { return "customer_contacts" }

type CustomerAddress struct {
	shared.BaseModel
	CustomerID   uint64 `json:"customerId" gorm:"not null;index"`
	ContactName  string `json:"contactName" gorm:"type:varchar(64)"`
	ContactPhone string `json:"contactPhone" gorm:"type:varchar(32)"`
	Province     string `json:"province" gorm:"type:varchar(64)"`
	City         string `json:"city" gorm:"type:varchar(64)"`
	District     string `json:"district" gorm:"type:varchar(64)"`
	Address      string `json:"address" gorm:"type:varchar(255);not null"`
	IsDefault    bool   `json:"isDefault" gorm:"not null;default:false;index"`
}

func (CustomerAddress) TableName() string { return "customer_addresses" }
