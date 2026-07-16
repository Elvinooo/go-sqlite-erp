package supplier

import "github.com/shopspring/decimal"

type SupplierRequest struct {
	Code           string          `json:"code" binding:"required,max=64"`
	Name           string          `json:"name" binding:"required,max=128"`
	ContactName    string          `json:"contactName" binding:"max=64"`
	Phone          string          `json:"phone" binding:"max=32"`
	Email          string          `json:"email" binding:"max=128"`
	TaxNo          string          `json:"taxNo" binding:"max=64"`
	Address        string          `json:"address" binding:"max=255"`
	SupplierTypes  string          `json:"supplierTypes"`
	PayableBalance decimal.Decimal `json:"payableBalance"`
	Status         string          `json:"status"`
	Remark         string          `json:"remark"`
}
