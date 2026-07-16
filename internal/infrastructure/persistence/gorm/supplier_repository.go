package gormrepo

import (
	"context"

	suppliermodel "erp/internal/domain/supplier/model"
	"erp/internal/shared/query"
	"gorm.io/gorm"
)

type SupplierRepository struct {
	db *gorm.DB
}

func NewSupplierRepository(db *gorm.DB) *SupplierRepository {
	return &SupplierRepository{db: db}
}

func (r *SupplierRepository) FindByID(ctx context.Context, tenantID uint64, id uint64) (*suppliermodel.Supplier, error) {
	var supplier suppliermodel.Supplier
	err := r.db.WithContext(ctx).Where("tenant_id = ? AND id = ?", tenantID, id).First(&supplier).Error
	return &supplier, err
}

func (r *SupplierRepository) List(ctx context.Context, tenantID uint64, q query.PageQuery) ([]suppliermodel.Supplier, int64, error) {
	var suppliers []suppliermodel.Supplier
	var total int64
	db := r.db.WithContext(ctx).Model(&suppliermodel.Supplier{}).Where("tenant_id = ?", tenantID)
	if q.Keyword != "" {
		like := "%" + q.Keyword + "%"
		db = db.Where("code LIKE ? OR name LIKE ? OR contact_name LIKE ? OR phone LIKE ? OR tax_no LIKE ? OR address LIKE ? OR supplier_types LIKE ?", like, like, like, like, like, like, like)
	}
	if q.SourceType != "" {
		db = db.Where("supplier_types LIKE ?", "%"+q.SourceType+"%")
	}
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	err := db.Order(safeOrder(q)).Offset(q.Offset()).Limit(q.PageSize).Find(&suppliers).Error
	return suppliers, total, err
}

func (r *SupplierRepository) Create(ctx context.Context, supplier *suppliermodel.Supplier) error {
	return r.db.WithContext(ctx).Create(supplier).Error
}

func (r *SupplierRepository) Update(ctx context.Context, supplier *suppliermodel.Supplier) error {
	return r.db.WithContext(ctx).Save(supplier).Error
}

func (r *SupplierRepository) Delete(ctx context.Context, tenantID uint64, id uint64) error {
	return r.db.WithContext(ctx).Where("tenant_id = ? AND id = ?", tenantID, id).Delete(&suppliermodel.Supplier{}).Error
}
