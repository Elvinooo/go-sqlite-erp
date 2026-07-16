package supplier

import (
	"context"
	"strings"

	suppliermodel "erp/internal/domain/supplier/model"
	supplierrepo "erp/internal/domain/supplier/repository"
	"erp/internal/shared/query"
)

type Service struct {
	suppliers supplierrepo.SupplierRepository
}

func NewService(suppliers supplierrepo.SupplierRepository) *Service {
	return &Service{suppliers: suppliers}
}

func (s *Service) List(ctx context.Context, tenantID uint64, q query.PageQuery) ([]suppliermodel.Supplier, int64, error) {
	return s.suppliers.List(ctx, tenantID, q)
}

func (s *Service) Get(ctx context.Context, tenantID uint64, id uint64) (*suppliermodel.Supplier, error) {
	return s.suppliers.FindByID(ctx, tenantID, id)
}

func (s *Service) Create(ctx context.Context, tenantID uint64, req SupplierRequest, operatorID uint64) (*suppliermodel.Supplier, error) {
	item := fromRequest(req)
	item.TenantID = tenantID
	item.CreatedBy = &operatorID
	item.UpdatedBy = &operatorID
	return item, s.suppliers.Create(ctx, item)
}

func (s *Service) Update(ctx context.Context, tenantID uint64, id uint64, req SupplierRequest, operatorID uint64) (*suppliermodel.Supplier, error) {
	item, err := s.suppliers.FindByID(ctx, tenantID, id)
	if err != nil {
		return nil, err
	}
	item.Code = req.Code
	item.Name = req.Name
	item.ContactName = req.ContactName
	item.Phone = req.Phone
	item.Email = req.Email
	item.TaxNo = req.TaxNo
	item.Address = req.Address
	item.SupplierTypes = defaultSupplierTypes(req.SupplierTypes)
	item.PayableBalance = req.PayableBalance
	item.Status = defaultStatus(req.Status)
	item.Remark = req.Remark
	item.UpdatedBy = &operatorID
	return item, s.suppliers.Update(ctx, item)
}

func (s *Service) Delete(ctx context.Context, tenantID uint64, id uint64) error {
	return s.suppliers.Delete(ctx, tenantID, id)
}

func fromRequest(req SupplierRequest) *suppliermodel.Supplier {
	item := &suppliermodel.Supplier{
		Code: req.Code, Name: req.Name, ContactName: req.ContactName, Phone: req.Phone, Email: req.Email,
		TaxNo: req.TaxNo, Address: req.Address, SupplierTypes: defaultSupplierTypes(req.SupplierTypes), PayableBalance: req.PayableBalance, Status: defaultStatus(req.Status),
	}
	item.Remark = req.Remark
	return item
}

func defaultSupplierTypes(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return "商品供应商"
	}
	return value
}

func defaultStatus(value string) string {
	if strings.TrimSpace(value) == "" {
		return "active"
	}
	return value
}
