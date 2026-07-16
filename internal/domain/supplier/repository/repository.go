package repository

import (
	"context"

	suppliermodel "erp/internal/domain/supplier/model"
	"erp/internal/shared/query"
)

type SupplierRepository interface {
	FindByID(ctx context.Context, tenantID uint64, id uint64) (*suppliermodel.Supplier, error)
	List(ctx context.Context, tenantID uint64, q query.PageQuery) ([]suppliermodel.Supplier, int64, error)
	Create(ctx context.Context, supplier *suppliermodel.Supplier) error
	Update(ctx context.Context, supplier *suppliermodel.Supplier) error
	Delete(ctx context.Context, tenantID uint64, id uint64) error
}
