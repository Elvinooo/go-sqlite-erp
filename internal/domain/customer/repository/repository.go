package repository

import (
	"context"

	customermodel "erp/internal/domain/customer/model"
	"erp/internal/shared/query"
	"github.com/shopspring/decimal"
)

type CustomerOrderHistory struct {
	ID               uint64          `json:"id"`
	OrderNo          string          `json:"orderNo"`
	OrderDate        string          `json:"orderDate"`
	Status           string          `json:"status"`
	TotalAmount      decimal.Decimal `json:"totalAmount"`
	ReceivedAmount   decimal.Decimal `json:"receivedAmount"`
	ReceivableAmount decimal.Decimal `json:"receivableAmount"`
	ProfitAmount     decimal.Decimal `json:"profitAmount"`
}

type CustomerRepository interface {
	FindByID(ctx context.Context, tenantID uint64, id uint64) (*customermodel.Customer, error)
	FindByCode(ctx context.Context, tenantID uint64, code string) (*customermodel.Customer, error)
	List(ctx context.Context, tenantID uint64, q query.PageQuery) ([]customermodel.Customer, int64, error)
	Create(ctx context.Context, customer *customermodel.Customer) error
	Update(ctx context.Context, customer *customermodel.Customer) error
	Delete(ctx context.Context, tenantID uint64, id uint64) error
	BatchCreate(ctx context.Context, customers []customermodel.Customer) error
	Debt(ctx context.Context, tenantID uint64, id uint64) (decimal.Decimal, error)
	OrderHistory(ctx context.Context, tenantID uint64, id uint64, q query.PageQuery) ([]CustomerOrderHistory, int64, error)
}
