package business

import (
	"context"
	"fmt"

	businessrepo "erp/internal/domain/business/repository"
	"erp/internal/shared/query"
)

type Service struct {
	repo businessrepo.ModuleRepository
}

func NewService(repo businessrepo.ModuleRepository) *Service {
	return &Service{repo: repo}
}

func (s *Service) Modules() []ModuleMeta {
	return []ModuleMeta{
		{Module: "products", Title: "商品管理", Permission: "product.manage", ListFields: []string{"id", "code", "name", "category", "brand", "spec", "unit", "barcode", "qrCode", "imageUrl", "minStock", "status", "createdAt", "updatedAt"}, ActionCodes: []string{"create", "edit", "delete"}},
		{Module: "sales", Title: "销售管理", Permission: "sales.manage", ListFields: []string{"id", "orderNo", "customerName", "productName", "quantity", "price", "costPrice", "orderDate", "status", "totalAmount", "receivedAmount", "receivableAmount", "costAmount", "profitAmount", "profitRate"}, ActionCodes: []string{"create", "print", "delete"}},
		{Module: "purchase", Title: "采购管理", Permission: "purchase.manage", ListFields: []string{"id", "orderNo", "supplierName", "productName", "quantity", "orderDate", "status", "totalAmount", "payableAmount"}, ActionCodes: []string{"create", "print"}},
		{Module: "inventory", Title: "库存管理", Permission: "inventory.manage", ListFields: []string{"id", "productCode", "productName", "warehouse", "quantity", "availableQuantity", "amount", "minStock", "status", "stockTime"}, ActionCodes: []string{"search", "scan", "scan-in", "scan-out"}},
		{Module: "inventory-movements", Title: "库存流水", Permission: "inventory.manage", ListFields: []string{"id", "movementNo", "productCode", "productName", "spec", "warehouse", "businessType", "businessNo", "quantityChange", "beforeQuantity", "afterQuantity", "purchaseOrderNo", "salesOrderNo", "repairOrderNo", "projectNo", "operatorName", "occurredAt"}, ActionCodes: []string{"search"}},
		{Module: "repair", Title: "维修管理", Permission: "repair.manage", ListFields: []string{"id", "orderNo", "customerName", "deviceName", "repairStatus", "serviceAmount", "partsAmount", "totalAmount", "partsCost", "outsourceCost", "laborCost", "profitAmount", "registeredAt"}, ActionCodes: []string{"create", "quote-print", "repair-print", "settlement-print", "outsource-stats", "parts-stats"}},
		{Module: "project", Title: "工程管理", Permission: "project.manage", ListFields: []string{"id", "projectNo", "name", "customerName", "status", "progress", "budgetAmount", "costAmount", "settleAmount", "contractAmount", "receivedAmount", "startDate", "endDate"}, ActionCodes: []string{"location", "log", "material", "camera", "sign", "print"}},
		{Module: "finance", Title: "资金流水", Permission: "finance.manage", ListFields: []string{"id", "recordNo", "recordType", "accountName", "targetName", "amount", "businessType", "businessNo", "status", "occurredAt"}, ActionCodes: []string{"book"}},
		{Module: "finance-accounts", Title: "资金账户", Permission: "finance.manage", ListFields: []string{"id", "code", "name", "accountType", "openingBalance", "balance", "status"}, ActionCodes: []string{"search"}},
		{Module: "receivables", Title: "应收账款", Permission: "finance.manage", ListFields: []string{"id", "receivableNo", "customerName", "sourceNo", "totalAmount", "receivedAmount", "balanceAmount", "invoiceDate", "dueDate", "status"}, ActionCodes: []string{"receive", "statement", "aging", "reminder", "print"}},
		{Module: "customer-statements", Title: "客户对账单", Permission: "finance.manage", ListFields: []string{"id", "statementNo", "customerName", "startDate", "endDate", "totalAmount", "receivedAmount", "unpaidAmount", "cumulativeDebt", "status"}, ActionCodes: []string{"sales-candidates", "generate", "confirm", "settle", "print"}},
		{Module: "payables", Title: "应付账款", Permission: "finance.manage", ListFields: []string{"id", "payableNo", "supplierName", "sourceNo", "totalAmount", "paidAmount", "balanceAmount", "billDate", "dueDate", "status"}, ActionCodes: []string{"pay", "print"}},
		{Module: "profit-report", Title: "利润报表", Permission: "finance.manage", ListFields: []string{"id", "salesDate", "salesOrderNo", "customerName", "productCode", "productName", "quantity", "salesPrice", "salesAmount", "costAmount", "profitAmount", "profitRate", "purchaseOrderNo"}, ActionCodes: []string{"summary", "ranking", "trend"}},
		{Module: "inventory-asset-report", Title: "库存资产报表", Permission: "finance.manage", ListFields: []string{"id", "productCode", "productName", "brand", "category", "spec", "quantity", "availableQuantity", "purchaseCost", "inventoryAmount", "latestPurchaseDate", "latestSalesDate", "inventoryStatus"}, ActionCodes: []string{"summary", "aging", "slow-moving", "trend"}},
		{Module: "document-delete-records", Title: "单据删除记录", Permission: "system.audit.view", ListFields: []string{"id", "documentType", "documentNo", "deleteUserName", "deleteTime", "deleteReason", "stockProcessed", "financeProcessed", "deleteStatus"}, ActionCodes: []string{"search", "detail"}},
	}
}

func (s *Service) Meta(module string) (ModuleMeta, error) {
	for _, item := range s.Modules() {
		if item.Module == module {
			return item, nil
		}
	}
	return ModuleMeta{}, fmt.Errorf("unsupported business module: %s", module)
}

func (s *Service) List(ctx context.Context, module string, tenantID uint64, q query.PageQuery) ([]map[string]any, int64, error) {
	if _, err := s.Meta(module); err != nil {
		return nil, 0, err
	}
	return s.repo.List(ctx, module, tenantID, q)
}

func (s *Service) Get(ctx context.Context, module string, tenantID uint64, id uint64) (map[string]any, error) {
	if _, err := s.Meta(module); err != nil {
		return nil, err
	}
	return s.repo.Find(ctx, module, tenantID, id)
}

func (s *Service) Create(ctx context.Context, module string, tenantID uint64, operatorID uint64, req ModuleRequest) (map[string]any, error) {
	if _, err := s.Meta(module); err != nil {
		return nil, err
	}
	return s.repo.Create(ctx, module, tenantID, operatorID, map[string]any(req))
}

func (s *Service) Update(ctx context.Context, module string, tenantID uint64, id uint64, operatorID uint64, req ModuleRequest) (map[string]any, error) {
	if _, err := s.Meta(module); err != nil {
		return nil, err
	}
	return s.repo.Update(ctx, module, tenantID, id, operatorID, map[string]any(req))
}

func (s *Service) Delete(ctx context.Context, module string, tenantID uint64, id uint64, operatorID uint64, req DeleteRequest) error {
	if _, err := s.Meta(module); err != nil {
		return err
	}
	return s.repo.Delete(ctx, module, tenantID, id, operatorID, req.Reason)
}

func (s *Service) Action(ctx context.Context, module string, tenantID uint64, operatorID uint64, req ModuleActionRequest) (map[string]any, error) {
	if _, err := s.Meta(module); err != nil {
		return nil, err
	}
	if req.Data == nil {
		req.Data = map[string]any{}
	}
	return s.repo.Action(ctx, module, tenantID, operatorID, req.Action, req.Data)
}

func (s *Service) CreatePhoto(ctx context.Context, module string, tenantID uint64, operatorID uint64, businessID uint64, photo map[string]any) (map[string]any, error) {
	if _, err := s.Meta(module); err != nil {
		return nil, err
	}
	return s.repo.CreatePhoto(ctx, module, tenantID, operatorID, businessID, photo)
}

func (s *Service) ListPhotos(ctx context.Context, module string, tenantID uint64, businessID uint64) ([]map[string]any, error) {
	if _, err := s.Meta(module); err != nil {
		return nil, err
	}
	return s.repo.ListPhotos(ctx, module, tenantID, businessID)
}
