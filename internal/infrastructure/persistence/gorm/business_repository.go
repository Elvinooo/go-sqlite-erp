package gormrepo

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	apperrors "erp/internal/shared/errors"
	"erp/internal/shared/query"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ModuleRepository struct {
	db *gorm.DB
}

type costAllocation struct {
	BatchID             uint64
	ProductID           uint64
	ProductCode         string
	ProductName         string
	Quantity            float64
	CostPrice           float64
	CostAmount          float64
	PurchaseOrderID     uint64
	PurchaseOrderItemID uint64
	PurchaseOrderNo     string
	SupplierID          uint64
	SupplierName        string
	PurchaseDate        time.Time
	PurchasePrice       float64
}

func NewModuleRepository(db *gorm.DB) *ModuleRepository {
	return &ModuleRepository{db: db}
}

type moduleSpec struct {
	Table       string
	NoField     string
	DateField   string
	StatusField string
	SearchCols  []string
	Defaults    map[string]any
}

var moduleSpecs = map[string]moduleSpec{
	"products": {
		Table: "products", NoField: "", DateField: "", StatusField: "status",
		SearchCols: []string{"code", "name", "barcode", "brand", "category", "status"},
		Defaults:   map[string]any{"unit": "台", "min_stock": 0, "status": "active"},
	},
	"sales": {
		Table: "sales_orders", NoField: "order_no", DateField: "order_date", StatusField: "status",
		SearchCols: []string{"order_no", "customer_name", "status"},
		Defaults:   map[string]any{"status": "已完成", "cost_method": "purchase_source", "total_amount": 0, "received_amount": 0, "receivable_amount": 0, "cost_amount": 0, "profit_amount": 0, "profit_rate": 0},
	},
	"purchase": {
		Table: "purchase_orders", NoField: "order_no", DateField: "order_date", StatusField: "status",
		SearchCols: []string{"order_no", "supplier_name", "status"},
		Defaults:   map[string]any{"status": "已完成", "total_amount": 0, "paid_amount": 0, "payable_amount": 0},
	},
	"inventory": {
		Table: "inventory_stocks", NoField: "", DateField: "stock_time", StatusField: "status",
		SearchCols: []string{"product_code", "product_name", "warehouse", "status"},
		Defaults:   map[string]any{"warehouse": "主仓库", "quantity": 0, "avg_cost": 0, "amount": 0, "min_stock": 0, "status": "正常"},
	},
	"inventory-movements": {
		Table: "inventory_movements", NoField: "movement_no", DateField: "occurred_at", StatusField: "",
		SearchCols: []string{"movement_no", "product_name", "source_type", "direction"},
		Defaults:   map[string]any{"quantity": 0, "unit_cost": 0, "amount": 0},
	},
	"repair": {
		Table: "repair_orders", NoField: "order_no", DateField: "registered_at", StatusField: "repair_status",
		SearchCols: []string{"order_no", "customer_name", "device_name", "repair_status"},
		Defaults:   map[string]any{"repair_status": "已完成", "quantity": 0, "price": 0, "cost_price": 0, "service_amount": 0, "onsite_fee": 0, "detection_fee": 0, "installation_fee": 0, "parts_amount": 0, "parts_cost": 0, "outsource_cost": 0, "labor_cost": 0, "transport_cost": 0, "other_cost": 0, "cost_amount": 0, "profit_amount": 0, "total_amount": 0},
	},
	"project": {
		Table: "project_projects", NoField: "project_no", DateField: "start_date", StatusField: "status",
		SearchCols: []string{"project_no", "name", "customer_name", "status", "contract_no"},
		Defaults:   map[string]any{"status": "planning", "progress": 0, "budget_amount": 0, "cost_amount": 0, "settle_amount": 0, "contract_amount": 0, "received_amount": 0},
	},
	"finance": {
		Table: "finance_records", NoField: "record_no", DateField: "occurred_at", StatusField: "status",
		SearchCols: []string{"record_no", "record_type", "account_name", "target_name", "business_no"},
		Defaults:   map[string]any{"record_type": "日常记账", "business_type": "日常记账", "source_type": "daily", "status": "已确认", "amount": 0},
	},
	"finance-accounts": {
		Table: "finance_accounts", NoField: "", DateField: "", StatusField: "status",
		SearchCols: []string{"code", "name", "account_type", "status"},
		Defaults:   map[string]any{"status": "active", "account_type": "other", "opening_balance": 0, "balance": 0},
	},
	"receivables": {
		Table: "receivables", NoField: "receivable_no", DateField: "invoice_date", StatusField: "status",
		SearchCols: []string{"receivable_no", "customer_name", "source_no", "status"},
		Defaults:   map[string]any{"source_type": "sales", "total_amount": 0, "received_amount": 0, "balance_amount": 0, "settlement_mode": "credit", "status": "unpaid"},
	},
	"customer-statements": {
		Table: "customer_statements", NoField: "statement_no", DateField: "created_at", StatusField: "status",
		SearchCols: []string{"statement_no", "customer_name", "status"},
		Defaults:   map[string]any{"status": "unconfirmed", "total_amount": 0, "received_amount": 0, "unpaid_amount": 0, "cumulative_debt": 0},
	},
	"payables": {
		Table: "payables", NoField: "payable_no", DateField: "bill_date", StatusField: "status",
		SearchCols: []string{"payable_no", "supplier_name", "source_no", "status"},
		Defaults:   map[string]any{"source_type": "purchase", "total_amount": 0, "paid_amount": 0, "balance_amount": 0, "status": "unpaid"},
	},
	"profit-report": {
		Table: "sales_cost_allocations", NoField: "", DateField: "", StatusField: "",
		SearchCols: []string{"product_code", "product_name", "purchase_order_no"},
		Defaults:   map[string]any{},
	},
	"inventory-asset-report": {
		Table: "inventory_batches", NoField: "", DateField: "", StatusField: "",
		SearchCols: []string{"product_code", "product_name", "purchase_order_no", "supplier_name"},
		Defaults:   map[string]any{},
	},
	"document-delete-records": {
		Table: "document_delete_records", NoField: "document_no", DateField: "delete_time", StatusField: "delete_status",
		SearchCols: []string{"document_type", "document_no", "delete_user_name", "delete_reason", "delete_status"},
		Defaults:   map[string]any{"delete_status": "WAITING", "stock_processed": false, "finance_processed": false},
	},
}

func (r *ModuleRepository) List(ctx context.Context, module string, tenantID uint64, q query.PageQuery) ([]map[string]any, int64, error) {
	spec, err := specOf(module)
	if err != nil {
		return nil, 0, err
	}
	if module == "inventory" {
		return r.listInventory(ctx, tenantID, q)
	}
	if module == "inventory-movements" {
		return r.listInventoryMovements(ctx, tenantID, q)
	}
	if module == "profit-report" {
		return r.listProfitReport(ctx, tenantID, q)
	}
	if module == "inventory-asset-report" {
		return r.listInventoryAssetReport(ctx, tenantID, q)
	}
	db := r.db.WithContext(ctx).Table(spec.Table).Where("tenant_id = ?", tenantID)
	if module == "sales" && q.DeletedOnly {
		db = db.Where("deleted_at IS NOT NULL")
	} else {
		db = db.Where("deleted_at IS NULL")
	}
	if strings.TrimSpace(q.Keyword) != "" {
		db = db.Where(searchWhere(spec.SearchCols), searchArgs(spec.SearchCols, q.Keyword)...)
	}
	if module == "document-delete-records" {
		if strings.TrimSpace(q.SourceType) != "" {
			db = db.Where("document_type = ?", q.SourceType)
		}
		if strings.TrimSpace(q.OperatorName) != "" {
			db = db.Where("delete_user_name LIKE ?", "%"+q.OperatorName+"%")
		}
		if strings.TrimSpace(q.StartDate) != "" {
			db = db.Where("delete_time >= ?", q.StartDate)
		}
		if strings.TrimSpace(q.EndDate) != "" {
			db = db.Where("delete_time < ?", exclusiveEndDateArg(q.EndDate))
		}
	}
	if module == "customer-statements" {
		if q.CustomerID > 0 {
			db = db.Where("customer_id = ?", q.CustomerID)
		}
		if strings.TrimSpace(q.SourceType) != "" {
			db = db.Where("status = ?", q.SourceType)
		}
		if strings.TrimSpace(q.StartDate) != "" {
			db = db.Where("end_date >= ?", q.StartDate)
		}
		if strings.TrimSpace(q.EndDate) != "" {
			db = db.Where("start_date < ?", exclusiveEndDateArg(q.EndDate))
		}
	}
	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	rows := make([]map[string]any, 0)
	err = db.Order(moduleOrder(q)).Offset(q.Offset()).Limit(q.PageSize).Find(&rows).Error
	return normalizeRows(rows), total, err
}

func (r *ModuleRepository) enrichDetail(ctx context.Context, module string, tenantID uint64, row map[string]any) error {
	id := mapID(row)
	if id == 0 {
		return nil
	}
	switch module {
	case "sales":
		items := make([]map[string]any, 0)
		if err := r.db.WithContext(ctx).Table("sales_order_items AS soi").
			Joins("LEFT JOIN sales_cost_allocations AS sca ON sca.tenant_id = soi.tenant_id AND sca.sales_order_item_id = soi.id AND sca.deleted_at IS NULL").
			Joins("LEFT JOIN inventory_batches AS ib ON ib.tenant_id = soi.tenant_id AND ib.id = COALESCE(soi.inventory_batch_id, sca.inventory_batch_id) AND ib.deleted_at IS NULL").
			Where("soi.tenant_id = ? AND soi.sales_order_id = ? AND soi.deleted_at IS NULL", tenantID, id).
			Select(`soi.*,
				COALESCE(soi.inventory_batch_id, sca.inventory_batch_id) AS inventory_batch_id,
				COALESCE(soi.purchase_order_id, sca.purchase_order_id) AS purchase_order_id,
				COALESCE(soi.purchase_order_item_id, ib.purchase_order_item_id) AS purchase_order_item_id,
				COALESCE(NULLIF(soi.purchase_order_no, ''), sca.purchase_order_no) AS purchase_order_no,
				COALESCE(NULLIF(soi.supplier_name, ''), sca.supplier_name) AS supplier_name,
				COALESCE(soi.purchase_date, sca.purchase_date) AS purchase_date,
				COALESCE(NULLIF(soi.purchase_price, 0), sca.purchase_price) AS purchase_price`).
			Order("soi.id ASC").
			Find(&items).Error; err != nil {
			return err
		}
		allocations := make([]map[string]any, 0)
		if err := r.db.WithContext(ctx).Table("sales_cost_allocations").
			Where("tenant_id = ? AND sales_order_id = ? AND deleted_at IS NULL", tenantID, id).
			Order("id ASC").
			Find(&allocations).Error; err != nil {
			return err
		}
		row["items"] = normalizeRows(items)
		row["costAllocations"] = normalizeRows(allocations)
		if err := r.enrichMerchantInfo(ctx, tenantID, row); err != nil {
			return err
		}
	case "purchase":
		items := make([]map[string]any, 0)
		if err := r.db.WithContext(ctx).Table("purchase_order_items AS poi").
			Joins("LEFT JOIN products AS p ON p.tenant_id = poi.tenant_id AND p.id = poi.product_id AND p.deleted_at IS NULL").
			Where("poi.tenant_id = ? AND poi.purchase_order_id = ? AND poi.deleted_at IS NULL", tenantID, id).
			Select(`poi.*, p.spec,
				COALESCE((SELECT SUM(ib.inbound_quantity - ib.remaining_quantity) FROM inventory_batches ib WHERE ib.tenant_id = poi.tenant_id AND ib.purchase_order_item_id = poi.id AND ib.deleted_at IS NULL), 0) AS sold_quantity,
				COALESCE((SELECT SUM(ib.remaining_quantity) FROM inventory_batches ib WHERE ib.tenant_id = poi.tenant_id AND ib.purchase_order_item_id = poi.id AND ib.deleted_at IS NULL), 0) AS remaining_inventory,
				COALESCE((SELECT ib.status FROM inventory_batches ib WHERE ib.tenant_id = poi.tenant_id AND ib.purchase_order_item_id = poi.id AND ib.deleted_at IS NULL ORDER BY ib.id LIMIT 1), '') AS inventory_status`).
			Order("poi.id ASC").
			Find(&items).Error; err != nil {
			return err
		}
		batches := make([]map[string]any, 0)
		if err := r.db.WithContext(ctx).Table("inventory_batches").
			Where("tenant_id = ? AND purchase_order_id = ? AND deleted_at IS NULL", tenantID, id).
			Order("id ASC").
			Find(&batches).Error; err != nil {
			return err
		}
		row["items"] = normalizeRows(items)
		row["inventoryBatches"] = normalizeRows(batches)
		if err := r.enrichMerchantInfo(ctx, tenantID, row); err != nil {
			return err
		}
		if createdBy := uintValue(row["created_by"]); createdBy > 0 {
			user := map[string]any{}
			if err := r.db.WithContext(ctx).Table("users").Where("id = ?", createdBy).Take(&user).Error; err == nil {
				row["buyerName"] = firstNotBlank(stringValue(user["real_name"]), stringValue(user["username"]))
			} else if err != gorm.ErrRecordNotFound {
				return err
			}
		}
	case "inventory":
		if err := r.enrichInventoryDetail(ctx, tenantID, row); err != nil {
			return err
		}
	case "finance":
		if err := r.enrichFinanceDetail(ctx, tenantID, row); err != nil {
			return err
		}
	case "customer-statements":
		items := make([]map[string]any, 0)
		if err := r.db.WithContext(ctx).Table("customer_statement_items").
			Where("tenant_id = ? AND statement_id = ? AND deleted_at IS NULL", tenantID, id).
			Order("sale_date ASC, id ASC").Find(&items).Error; err != nil {
			return err
		}
		row["items"] = normalizeRows(items)
	case "repair":
		items := make([]map[string]any, 0)
		if err := r.db.WithContext(ctx).Table("repair_order_items").
			Where("tenant_id = ? AND repair_order_id = ? AND deleted_at IS NULL", tenantID, id).
			Order("id ASC").
			Find(&items).Error; err != nil {
			return err
		}
		row["items"] = normalizeRows(items)
	case "document-delete-records":
		details := make([]map[string]any, 0)
		if err := r.db.WithContext(ctx).Table("document_delete_details").
			Where("tenant_id = ? AND record_id = ? AND deleted_at IS NULL", tenantID, id).
			Order("id ASC").
			Find(&details).Error; err != nil {
			return err
		}
		row["details"] = normalizeRows(details)

		documentType := stringValue(row["document_type"])
		documentID := uintValue(row["document_id"])
		documentNo := stringValue(row["document_no"])
		movements := make([]map[string]any, 0)
		if documentID > 0 {
			if err := r.db.WithContext(ctx).Table("inventory_movements").
				Where("tenant_id = ? AND source_id = ? AND deleted_at IS NULL", tenantID, documentID).
				Where("source_type IN ?", documentDeleteMovementTypes(documentType)).
				Order("occurred_at DESC, id DESC").
				Find(&movements).Error; err != nil {
				return err
			}
		}
		row["stockMovements"] = normalizeRows(movements)

		financeRecords := make([]map[string]any, 0)
		if documentNo != "" {
			if err := r.db.WithContext(ctx).Table("finance_records").
				Where("tenant_id = ? AND business_no = ? AND deleted_at IS NULL", tenantID, documentNo).
				Order("occurred_at DESC, id DESC").
				Find(&financeRecords).Error; err != nil {
				return err
			}
		}
		row["financeRecords"] = normalizeRows(financeRecords)
	}
	return nil
}

func (r *ModuleRepository) Find(ctx context.Context, module string, tenantID uint64, id uint64) (map[string]any, error) {
	spec, err := specOf(module)
	if err != nil {
		return nil, err
	}
	row := map[string]any{}
	err = r.db.WithContext(ctx).Table(spec.Table).Where("tenant_id = ? AND id = ? AND deleted_at IS NULL", tenantID, id).Take(&row).Error
	if err != nil {
		return nil, err
	}
	if err := r.enrichDetail(ctx, module, tenantID, row); err != nil {
		return nil, err
	}
	return normalizeRow(row), nil
}

func (r *ModuleRepository) enrichFinanceDetail(ctx context.Context, tenantID uint64, row map[string]any) error {
	sourceType := stringValue(row["source_type"])
	businessType := stringValue(row["business_type"])
	businessNo := stringValue(row["business_no"])
	sourceID := uintValue(row["source_id"])

	if sourceType == "sales" || strings.Contains(businessType, "销售") || strings.Contains(businessType, "应收") {
		if sales, ok, err := r.findSalesOrder(ctx, tenantID, sourceID, businessNo); err != nil {
			return err
		} else if ok {
			row["salesOrder"] = normalizeRow(sales)
			return nil
		}
		if sales, ok, err := r.findSalesOrderFromReceivable(ctx, tenantID, businessNo); err != nil {
			return err
		} else if ok {
			row["salesOrder"] = normalizeRow(sales)
			return nil
		}
	}

	if sourceType == "purchase" || strings.Contains(businessType, "采购") || strings.Contains(businessType, "应付") {
		if purchase, ok, err := r.findPurchaseOrder(ctx, tenantID, sourceID, businessNo); err != nil {
			return err
		} else if ok {
			row["purchaseOrder"] = normalizeRow(purchase)
			return nil
		}
		if purchase, ok, err := r.findPurchaseOrderFromPayable(ctx, tenantID, businessNo); err != nil {
			return err
		} else if ok {
			row["purchaseOrder"] = normalizeRow(purchase)
			return nil
		}
	}
	return nil
}

func (r *ModuleRepository) enrichMerchantInfo(ctx context.Context, tenantID uint64, row map[string]any) error {
	settings := []map[string]any{}
	if err := r.db.WithContext(ctx).Table("system_settings").
		Where("tenant_id = ? AND setting_key IN ? AND deleted_at IS NULL", tenantID, []string{"merchant.company_name", "merchant.contact_name", "merchant.contact_phone"}).
		Find(&settings).Error; err != nil {
		return err
	}
	for _, setting := range settings {
		switch stringValue(setting["setting_key"]) {
		case "merchant.company_name":
			row["merchantCompanyName"] = stringValue(setting["setting_value"])
		case "merchant.contact_name":
			row["merchantContactName"] = stringValue(setting["setting_value"])
		case "merchant.contact_phone":
			row["merchantContactPhone"] = stringValue(setting["setting_value"])
		}
	}
	return nil
}

func (r *ModuleRepository) listProfitReport(ctx context.Context, tenantID uint64, q query.PageQuery) ([]map[string]any, int64, error) {
	rows := make([]map[string]any, 0)
	db := r.profitSalesBase(ctx, tenantID, q)
	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	order := reportOrder(q, map[string]string{
		"id":              "sca.id",
		"salesDate":       "so.order_date",
		"salesAmount":     "sales_amount",
		"costAmount":      "sca.cost_amount",
		"profitAmount":    "profit_amount",
		"profitRate":      "profit_rate",
		"quantity":        "sca.quantity",
		"salesOrderNo":    "so.order_no",
		"purchaseOrderNo": "sca.purchase_order_no",
	}, "so.order_date")
	if err := db.Select(`sca.id,
		so.id AS sales_order_id,
		so.order_no AS sales_order_no,
		so.order_date AS sales_date,
		so.customer_id,
		so.customer_name,
		sca.product_id,
		sca.product_code,
		sca.product_name,
		COALESCE(p.category, '') AS category,
		COALESCE(p.brand, '') AS brand,
		sca.quantity,
		COALESCE(NULLIF(soi.price, 0), so.price, 0) AS sales_price,
		sca.quantity * COALESCE(NULLIF(soi.price, 0), so.price, 0) AS sales_amount,
		sca.cost_amount,
		sca.quantity * COALESCE(NULLIF(soi.price, 0), so.price, 0) - sca.cost_amount AS profit_amount,
		CASE WHEN sca.quantity * COALESCE(NULLIF(soi.price, 0), so.price, 0) > 0 THEN
			(sca.quantity * COALESCE(NULLIF(soi.price, 0), so.price, 0) - sca.cost_amount) * 100.0 / (sca.quantity * COALESCE(NULLIF(soi.price, 0), so.price, 0))
		ELSE 0 END AS profit_rate,
		sca.purchase_order_id,
		sca.purchase_order_no,
		sca.supplier_id,
		sca.supplier_name`).
		Order(order).Offset(q.Offset()).Limit(q.PageSize).Find(&rows).Error; err != nil {
		return nil, 0, err
	}
	return normalizeRows(rows), total, nil
}

func (r *ModuleRepository) listInventoryAssetReport(ctx context.Context, tenantID uint64, q query.PageQuery) ([]map[string]any, int64, error) {
	rows := make([]map[string]any, 0)
	db := r.inventoryAssetBase(ctx, tenantID, q)
	countRows := make([]map[string]any, 0)
	if err := db.Session(&gorm.Session{}).
		Select("COALESCE(NULLIF(ib.product_code, ''), 'ID:' || COALESCE(CAST(ib.product_id AS TEXT), CAST(ib.id AS TEXT))) AS group_key").
		Group("COALESCE(NULLIF(ib.product_code, ''), 'ID:' || COALESCE(CAST(ib.product_id AS TEXT), CAST(ib.id AS TEXT)))").
		Find(&countRows).Error; err != nil {
		return nil, 0, err
	}
	order := reportOrder(q, map[string]string{
		"id":                 "id",
		"productCode":        "product_code",
		"productName":        "product_name",
		"quantity":           "quantity",
		"inventoryAmount":    "inventory_amount",
		"latestPurchaseDate": "latest_purchase_date",
		"latestSalesDate":    "latest_sales_date",
	}, "latest_purchase_date")
	if err := db.Select(`MIN(ib.id) AS id,
		MIN(ib.product_id) AS product_id,
		MAX(ib.product_code) AS product_code,
		MAX(ib.product_name) AS product_name,
		COALESCE(MAX(p.brand), '') AS brand,
		COALESCE(MAX(p.category), '') AS category,
		COALESCE(MAX(p.spec), '') AS spec,
		SUM(ib.remaining_quantity) AS quantity,
		SUM(ib.remaining_quantity) AS available_quantity,
		CASE WHEN SUM(ib.remaining_quantity) > 0 THEN SUM(ib.remaining_quantity * ib.purchase_price) / SUM(ib.remaining_quantity) ELSE 0 END AS purchase_cost,
		SUM(ib.remaining_quantity * ib.purchase_price) AS inventory_amount,
		MAX(ib.purchase_date) AS latest_purchase_date,
		MAX(last_sales.last_sales_date) AS latest_sales_date,
		CASE
			WHEN SUM(ib.remaining_quantity) <= 0 THEN '缺货'
			WHEN SUM(ib.remaining_quantity) <= COALESCE(MAX(p.min_stock), 0) THEN '预警'
			ELSE '正常'
		END AS inventory_status`).
		Group("COALESCE(NULLIF(ib.product_code, ''), 'ID:' || COALESCE(CAST(ib.product_id AS TEXT), CAST(ib.id AS TEXT)))").
		Order(order).Offset(q.Offset()).Limit(q.PageSize).Find(&rows).Error; err != nil {
		return nil, 0, err
	}
	return normalizeRows(rows), int64(len(countRows)), nil
}

func (r *ModuleRepository) findSalesOrder(ctx context.Context, tenantID uint64, id uint64, orderNo string) (map[string]any, bool, error) {
	row := map[string]any{}
	db := r.db.WithContext(ctx).Table("sales_orders").Where("tenant_id = ? AND deleted_at IS NULL", tenantID)
	if id > 0 {
		db = db.Where("id = ?", id)
	} else if strings.TrimSpace(orderNo) != "" {
		db = db.Where("order_no = ?", orderNo)
	} else {
		return nil, false, nil
	}
	err := db.Take(&row).Error
	if err == gorm.ErrRecordNotFound {
		return nil, false, nil
	}
	return row, err == nil, err
}

func (r *ModuleRepository) findPurchaseOrder(ctx context.Context, tenantID uint64, id uint64, orderNo string) (map[string]any, bool, error) {
	row := map[string]any{}
	db := r.db.WithContext(ctx).Table("purchase_orders").Where("tenant_id = ? AND deleted_at IS NULL", tenantID)
	if id > 0 {
		db = db.Where("id = ?", id)
	} else if strings.TrimSpace(orderNo) != "" {
		db = db.Where("order_no = ?", orderNo)
	} else {
		return nil, false, nil
	}
	err := db.Take(&row).Error
	if err == gorm.ErrRecordNotFound {
		return nil, false, nil
	}
	return row, err == nil, err
}

func (r *ModuleRepository) findSalesOrderFromReceivable(ctx context.Context, tenantID uint64, receivableNo string) (map[string]any, bool, error) {
	if strings.TrimSpace(receivableNo) == "" {
		return nil, false, nil
	}
	receivable := map[string]any{}
	err := r.db.WithContext(ctx).Table("receivables").
		Where("tenant_id = ? AND receivable_no = ? AND deleted_at IS NULL", tenantID, receivableNo).
		Take(&receivable).Error
	if err == gorm.ErrRecordNotFound {
		return nil, false, nil
	}
	if err != nil {
		return nil, false, err
	}
	return r.findSalesOrder(ctx, tenantID, uintValue(receivable["source_id"]), stringValue(receivable["source_no"]))
}

func (r *ModuleRepository) findPurchaseOrderFromPayable(ctx context.Context, tenantID uint64, payableNo string) (map[string]any, bool, error) {
	if strings.TrimSpace(payableNo) == "" {
		return nil, false, nil
	}
	payable := map[string]any{}
	err := r.db.WithContext(ctx).Table("payables").
		Where("tenant_id = ? AND payable_no = ? AND deleted_at IS NULL", tenantID, payableNo).
		Take(&payable).Error
	if err == gorm.ErrRecordNotFound {
		return nil, false, nil
	}
	if err != nil {
		return nil, false, err
	}
	return r.findPurchaseOrder(ctx, tenantID, uintValue(payable["source_id"]), stringValue(payable["source_no"]))
}

func (r *ModuleRepository) Create(ctx context.Context, module string, tenantID uint64, operatorID uint64, data map[string]any) (map[string]any, error) {
	spec, err := specOf(module)
	if err != nil {
		return nil, err
	}
	if module == "inventory" {
		return nil, apperrors.New(40018, "库存数量必须由采购入库、销售出库、维修领料、工程领料或库存盘点自动产生，禁止直接新建库存", 400)
	}
	if module == "finance" || module == "receivables" || module == "payables" {
		return nil, apperrors.New(40031, "应收应付必须由销售、采购、维修、工程等业务自动生成", 400)
	}
	var created map[string]any
	err = r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		repo := &ModuleRepository{db: tx}
		now := time.Now()
		row := normalizeInput(data)
		for key, value := range spec.Defaults {
			if _, ok := row[key]; !ok {
				row[key] = value
			}
		}
		if (module == "sales" || module == "purchase") && empty(row["status"]) {
			row["status"] = "已完成"
		}
		if module == "finance-accounts" {
			if err := repo.prepareFinanceAccount(row, true); err != nil {
				return err
			}
		}
		if err := repo.prepareBusinessCustomer(ctx, module, tenantID, row); err != nil {
			return err
		}
		if err := repo.prepareProduct(ctx, module, tenantID, operatorID, row); err != nil {
			return err
		}
		if module == "repair" {
			if err := repo.prepareRepairParts(ctx, tenantID, row); err != nil {
				return err
			}
		}
		if module == "sales" {
			if err := repo.prepareSalesCost(ctx, tenantID, row); err != nil {
				return err
			}
		}
		if spec.NoField != "" && empty(row[spec.NoField]) {
			row[spec.NoField] = nextNo(module)
		}
		applyBusinessCalculations(module, row)
		if spec.DateField != "" && empty(row[spec.DateField]) {
			row[spec.DateField] = now
		}
		row["tenant_id"] = tenantID
		row["created_by"] = operatorID
		row["updated_by"] = operatorID
		row["created_at"] = now
		row["updated_at"] = now
		dbRow, err := repo.tableRow(spec.Table, row)
		if err != nil {
			return err
		}
		if err := repo.db.Table(spec.Table).Create(dbRow).Error; err != nil {
			return err
		}
		copyGeneratedID(row, dbRow)
		if err := repo.ensureGeneratedID(ctx, spec, tenantID, row); err != nil {
			return err
		}
		if err := repo.createSideEffects(ctx, module, tenantID, operatorID, row); err != nil {
			return err
		}
		created = normalizeRow(row)
		return nil
	})
	return created, err
}

func (r *ModuleRepository) ensureGeneratedID(ctx context.Context, spec moduleSpec, tenantID uint64, row map[string]any) error {
	if mapID(row) > 0 || uintValue(row["@id"]) > 0 {
		return nil
	}
	if spec.NoField == "" || empty(row[spec.NoField]) {
		return nil
	}
	found := map[string]any{}
	if err := r.db.WithContext(ctx).Table(spec.Table).
		Select("id").
		Where("tenant_id = ? AND "+spec.NoField+" = ?", tenantID, row[spec.NoField]).
		Take(&found).Error; err != nil {
		return err
	}
	copyGeneratedID(row, found)
	return nil
}

func (r *ModuleRepository) Update(ctx context.Context, module string, tenantID uint64, id uint64, operatorID uint64, data map[string]any) (map[string]any, error) {
	spec, err := specOf(module)
	if err != nil {
		return nil, err
	}
	if module == "inventory" {
		return nil, apperrors.New(40019, "库存数量必须由业务单据自动产生，禁止直接编辑库存", 400)
	}
	if module == "finance" || module == "receivables" || module == "payables" {
		return nil, apperrors.New(40032, "财务流水、应收账款、应付账款由业务自动维护，禁止手工编辑", 400)
	}
	row := normalizeInput(data)
	delete(row, "id")
	delete(row, "tenant_id")
	delete(row, "created_at")
	delete(row, "created_by")
	row["updated_by"] = operatorID
	row["updated_at"] = time.Now()
	if err := r.prepareProduct(ctx, module, tenantID, operatorID, row); err != nil {
		return nil, err
	}
	if module == "repair" {
		if err := r.prepareRepairParts(ctx, tenantID, row); err != nil {
			return nil, err
		}
	}
	if module == "sales" {
		if err := r.prepareSalesCost(ctx, tenantID, row); err != nil {
			return nil, err
		}
	}
	if module == "finance-accounts" {
		delete(row, "opening_balance")
		delete(row, "balance")
		if err := r.prepareFinanceAccount(row, false); err != nil {
			return nil, err
		}
	}
	if err := r.prepareBusinessCustomer(ctx, module, tenantID, row); err != nil {
		return nil, err
	}
	applyBusinessCalculations(module, row)
	dbRow, err := r.tableRow(spec.Table, row)
	if err != nil {
		return nil, err
	}
	if err := r.db.WithContext(ctx).Table(spec.Table).Where("tenant_id = ? AND id = ? AND deleted_at IS NULL", tenantID, id).Updates(dbRow).Error; err != nil {
		return nil, err
	}
	if module == "repair" {
		if err := r.processRepairCompletion(ctx, tenantID, operatorID, id); err != nil {
			return nil, err
		}
	}
	return r.Find(ctx, module, tenantID, id)
}

func (r *ModuleRepository) createSideEffects(ctx context.Context, module string, tenantID uint64, operatorID uint64, row map[string]any) error {
	switch module {
	case "sales":
		orderID := uintValue(row["id"])
		if orderID == 0 {
			orderID = uintValue(row["@id"])
		}
		if orderID == 0 || empty(row["product_name"]) {
			return nil
		}
		allocations, err := r.consumeSalesBatches(ctx, tenantID, orderID, row)
		if err != nil {
			return err
		}
		if err := r.createSalesItems(ctx, tenantID, operatorID, orderID, row, allocations); err != nil {
			return err
		}
		if err := r.applyInventoryMovement(ctx, tenantID, operatorID, "sales", orderID, row, "out"); err != nil {
			return err
		}
		return r.applySalesSettlement(ctx, tenantID, operatorID, orderID, row)
	case "purchase":
		orderID := uintValue(row["id"])
		if orderID == 0 {
			orderID = uintValue(row["@id"])
		}
		if orderID == 0 || empty(row["product_name"]) {
			return nil
		}
		itemID, err := r.createPurchaseItem(ctx, tenantID, operatorID, orderID, row)
		if err != nil {
			return err
		}
		if err := r.createInventoryBatch(ctx, tenantID, operatorID, orderID, itemID, row); err != nil {
			return err
		}
		if err := r.applyInventoryMovement(ctx, tenantID, operatorID, "purchase", orderID, row, "in"); err != nil {
			return err
		}
		return r.applyPurchaseSettlement(ctx, tenantID, operatorID, orderID, row)
	case "repair":
		orderID := uintValue(row["id"])
		if orderID == 0 {
			orderID = uintValue(row["@id"])
		}
		if orderID == 0 {
			return nil
		}
		if err := r.createRepairItems(ctx, tenantID, operatorID, orderID, row); err != nil {
			return err
		}
		return r.processRepairCompletion(ctx, tenantID, operatorID, orderID)
	case "project":
		projectID := uintValue(row["id"])
		if projectID == 0 {
			projectID = uintValue(row["@id"])
		}
		return r.applyProjectSettlement(ctx, tenantID, operatorID, projectID, row)
	case "finance":
		return r.applyManualDailyFinance(ctx, tenantID, operatorID, row)
	}
	return nil
}

func applyBusinessCalculations(module string, row map[string]any) {
	qty := numberValue(row["quantity"])
	price := numberValue(row["price"])
	costPrice := numberValue(row["cost_price"])
	if module == "products" && empty(row["code"]) {
		row["code"] = nextProductCode()
	}
	if module == "sales" && qty > 0 && price > 0 {
		total := qty * price
		cost := qty * costPrice
		profit := total - cost
		row["total_amount"] = total
		row["cost_amount"] = cost
		row["profit_amount"] = profit
		row["profit_rate"] = 0
		if total > 0 {
			row["profit_rate"] = profit / total * 100
		}
		row["received_amount"] = 0
		row["receivable_amount"] = total
	}
	if module == "purchase" && qty > 0 && price > 0 {
		total := qty * price
		row["total_amount"] = total
		row["paid_amount"] = 0
		row["payable_amount"] = total
	}
	if module == "inventory" {
		qty := numberValue(row["quantity"])
		cost := numberValue(row["avg_cost"])
		if qty > 0 && cost > 0 {
			row["amount"] = qty * cost
		}
		if qty <= numberValue(row["min_stock"]) {
			row["status"] = "预警"
		} else {
			row["status"] = "正常"
		}
	}
	if module == "repair" {
		qty := numberValue(row["quantity"])
		price := numberValue(row["price"])
		costPrice := numberValue(row["cost_price"])
		partsAmount, partsCost := sumRepairParts(repairInputRows(row, "repair_parts"))
		if partsAmount == 0 {
			partsAmount = numberValue(row["parts_amount"])
		}
		if partsAmount == 0 && qty > 0 && price > 0 {
			partsAmount = qty * price
			row["parts_amount"] = partsAmount
		}
		if partsCost == 0 {
			partsCost = numberValue(row["parts_cost"])
		}
		if partsCost == 0 && qty > 0 && costPrice > 0 {
			partsCost = qty * costPrice
			row["parts_cost"] = partsCost
		}
		chargeIncome := sumRepairInputAmount(repairInputRows(row, "charge_items"))
		serviceIncome := numberValue(row["service_amount"]) + numberValue(row["onsite_fee"]) + numberValue(row["detection_fee"]) + numberValue(row["installation_fee"])
		total := serviceIncome + partsAmount
		if chargeIncome > 0 {
			total += chargeIncome
		}
		if total == 0 {
			total = numberValue(row["total_amount"])
		}
		outsourceCost := sumRepairInputAmount(repairInputRows(row, "outsource_items"))
		if outsourceCost == 0 {
			outsourceCost = numberValue(row["outsource_cost"])
		}
		otherCosts := repairInputRows(row, "cost_items")
		laborCost := numberValue(row["labor_cost"]) + sumRepairInputAmountByType(otherCosts, "人工成本")
		transportCost := numberValue(row["transport_cost"]) + sumRepairInputAmountByType(otherCosts, "运输费用")
		otherCost := numberValue(row["other_cost"]) + sumRepairInputAmountByType(otherCosts, "其他费用")
		row["outsource_cost"] = outsourceCost
		row["parts_amount"] = partsAmount
		row["parts_cost"] = partsCost
		row["labor_cost"] = laborCost
		row["transport_cost"] = transportCost
		row["other_cost"] = otherCost
		cost := partsCost + outsourceCost + laborCost + transportCost + otherCost
		row["total_amount"] = total
		row["cost_amount"] = cost
		row["profit_amount"] = total - cost
	}
}

func (r *ModuleRepository) prepareProduct(ctx context.Context, module string, tenantID uint64, operatorID uint64, row map[string]any) error {
	if module == "products" {
		if empty(row["code"]) {
			row["code"] = nextProductCode()
		}
		return nil
	}
	if module == "repair" && empty(row["product_id"]) && empty(row["product_code"]) && empty(row["product_name"]) {
		return nil
	}
	if module != "sales" && module != "purchase" && module != "inventory" && module != "repair" {
		return nil
	}
	product, err := r.findOrCreateProduct(ctx, tenantID, operatorID, row)
	if err != nil || product == nil {
		return err
	}
	row["product_id"] = mapID(product)
	row["product_code"] = stringValue(product["code"])
	row["product_name"] = stringValue(product["name"])
	if module == "sales" {
		if empty(row["cost_price"]) || numberValue(row["cost_price"]) == 0 {
			cost := r.currentStockCost(ctx, tenantID, uintValue(row["product_id"]), stringValue(row["product_code"]), firstNotBlank(stringValue(row["warehouse"]), "主仓库"))
			row["cost_price"] = cost
		}
	}
	if module == "repair" {
		if empty(row["cost_price"]) || numberValue(row["cost_price"]) == 0 {
			cost := r.currentStockCost(ctx, tenantID, uintValue(row["product_id"]), stringValue(row["product_code"]), firstNotBlank(stringValue(row["warehouse"]), "主仓库"))
			row["cost_price"] = cost
		}
	}
	return nil
}

func (r *ModuleRepository) prepareRepairParts(ctx context.Context, tenantID uint64, row map[string]any) error {
	parts := repairInputRows(row, "repair_parts")
	if len(parts) == 0 {
		return nil
	}
	for _, part := range parts {
		qty := repairInputQuantity(part)
		batchID := uintValue(firstPresent(part, "inventory_batch_id", "inventoryBatchId", "batch_id", "batchId"))
		if batchID == 0 {
			return apperrors.New(40040, "维修配件必须选择有库存的采购批次", 400)
		}
		batch := map[string]any{}
		if err := r.db.WithContext(ctx).Table("inventory_batches").
			Where("tenant_id = ? AND id = ? AND remaining_quantity > 0 AND deleted_at IS NULL", tenantID, batchID).
			Take(&batch).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				return apperrors.New(40041, "选择的维修配件库存不可用", 400)
			}
			return err
		}
		if numberValue(batch["remaining_quantity"])+0.000001 < qty {
			return apperrors.New(40042, "维修配件库存不足", 400)
		}
		purchasePrice := numberValue(batch["purchase_price"])
		part["inventory_batch_id"] = batchID
		part["inventoryBatchId"] = batchID
		part["product_id"] = batch["product_id"]
		part["productId"] = batch["product_id"]
		part["product_code"] = batch["product_code"]
		part["productCode"] = batch["product_code"]
		part["product_name"] = batch["product_name"]
		part["productName"] = batch["product_name"]
		if numberValue(firstPresent(part, "cost_price", "costPrice", "purchase_price", "purchasePrice")) <= 0 {
			part["cost_price"] = purchasePrice
			part["costPrice"] = purchasePrice
		}
		if numberValue(firstPresent(part, "purchase_price", "purchasePrice")) <= 0 {
			part["purchase_price"] = purchasePrice
			part["purchasePrice"] = purchasePrice
		}
	}
	row["repair_parts"] = parts
	return nil
}

func (r *ModuleRepository) prepareBusinessCustomer(ctx context.Context, module string, tenantID uint64, row map[string]any) error {
	if module != "sales" && module != "repair" && module != "project" {
		return nil
	}
	customerID := uintValue(row["customer_id"])
	customerName := strings.TrimSpace(stringValue(row["customer_name"]))
	if customerID == 0 && customerName == "" {
		return nil
	}
	customer := map[string]any{}
	db := r.db.WithContext(ctx).Table("customers").Where("tenant_id = ? AND deleted_at IS NULL", tenantID)
	var err error
	if customerID > 0 {
		err = db.Where("id = ?", customerID).Take(&customer).Error
	} else {
		err = db.Where("name = ?", customerName).Take(&customer).Error
	}
	if err == gorm.ErrRecordNotFound {
		return nil
	}
	if err != nil {
		return err
	}
	if id := mapID(customer); id > 0 {
		row["customer_id"] = id
	}
	if name := stringValue(customer["name"]); name != "" {
		row["customer_name"] = name
	}
	return nil
}

func (r *ModuleRepository) prepareSalesCost(ctx context.Context, tenantID uint64, row map[string]any) error {
	qty := numberValue(row["quantity"])
	if qty <= 0 {
		return nil
	}
	productID := uintValue(row["product_id"])
	productCode := stringValue(row["product_code"])
	warehouse := firstNotBlank(stringValue(row["warehouse"]), "主仓库")
	batchIDs := selectedPurchaseBatchIDs(row)
	if len(batchIDs) == 0 {
		return apperrors.New(40015, "销售商品必须选择采购来源", 400)
	}
	allocations, err := r.matchSelectedPurchaseBatches(ctx, tenantID, productID, productCode, warehouse, qty, batchIDs, false, row)
	if err != nil {
		return err
	}
	if len(allocations) == 0 {
		return apperrors.New(40016, "选择的采购来源库存不足", 400)
	}
	totalCost := 0.0
	for _, allocation := range allocations {
		totalCost += allocation.CostAmount
	}
	costPrice := totalCost / qty
	first := allocations[0]
	row["cost_method"] = "purchase_source"
	row["cost_price"] = costPrice
	row["cost_amount"] = totalCost
	row["purchase_price"] = first.PurchasePrice
	row["purchase_order_id"] = nilIfZero(first.PurchaseOrderID)
	row["purchase_order_no"] = first.PurchaseOrderNo
	row["supplier_id"] = nilIfZero(first.SupplierID)
	row["supplier_name"] = first.SupplierName
	if !first.PurchaseDate.IsZero() {
		row["purchase_date"] = first.PurchaseDate
	}
	if first.BatchID > 0 {
		row["inventory_batch_id"] = first.BatchID
	}
	return nil
}

func (r *ModuleRepository) createPurchaseItem(ctx context.Context, tenantID uint64, operatorID uint64, orderID uint64, row map[string]any) (uint64, error) {
	now := time.Now()
	item := map[string]any{
		"tenant_id":         tenantID,
		"purchase_order_id": orderID,
		"product_id":        nilIfZero(uintValue(row["product_id"])),
		"product_code":      row["product_code"],
		"product_name":      row["product_name"],
		"quantity":          numberValue(row["quantity"]),
		"price":             numberValue(row["price"]),
		"amount":            numberValue(row["total_amount"]),
		"warehouse":         firstNotBlank(stringValue(row["warehouse"]), "主仓库"),
		"created_by":        operatorID,
		"updated_by":        operatorID,
		"created_at":        now,
		"updated_at":        now,
	}
	if err := r.db.WithContext(ctx).Table("purchase_order_items").Create(item).Error; err != nil {
		return 0, err
	}
	return mapID(item), nil
}

func (r *ModuleRepository) createInventoryBatch(ctx context.Context, tenantID uint64, operatorID uint64, orderID uint64, itemID uint64, row map[string]any) error {
	qty := numberValue(row["quantity"])
	if qty <= 0 {
		return nil
	}
	now := time.Now()
	batch := map[string]any{
		"tenant_id":              tenantID,
		"batch_no":               nextNo("batch"),
		"product_id":             nilIfZero(uintValue(row["product_id"])),
		"product_code":           row["product_code"],
		"product_name":           row["product_name"],
		"warehouse":              firstNotBlank(stringValue(row["warehouse"]), "主仓库"),
		"purchase_order_id":      nilIfZero(orderID),
		"purchase_order_item_id": nilIfZero(itemID),
		"purchase_order_no":      row["order_no"],
		"supplier_id":            nilIfZero(uintValue(row["supplier_id"])),
		"supplier_name":          row["supplier_name"],
		"purchase_date":          timeValue(row["order_date"], now),
		"purchase_price":         numberValue(row["price"]),
		"inbound_quantity":       qty,
		"remaining_quantity":     qty,
		"status":                 batchStatus(qty, qty),
		"created_by":             operatorID,
		"updated_by":             operatorID,
		"created_at":             now,
		"updated_at":             now,
	}
	return r.db.WithContext(ctx).Table("inventory_batches").Create(batch).Error
}

func (r *ModuleRepository) enrichInventoryDetail(ctx context.Context, tenantID uint64, row map[string]any) error {
	if err := r.mergeInventoryRow(ctx, tenantID, row); err != nil {
		return err
	}
	if err := r.enrichInventoryProducts(ctx, tenantID, []map[string]any{row}); err != nil {
		return err
	}
	productID := uintValue(row["product_id"])
	productCode := stringValue(row["product_code"])
	warehouse := firstNotBlank(stringValue(row["warehouse"]), "主仓库")

	_ = warehouse
	batches := make([]map[string]any, 0)
	db := r.db.WithContext(ctx).Table("inventory_batches").
		Where("tenant_id = ? AND deleted_at IS NULL", tenantID)
	if productID > 0 {
		db = db.Where("product_id = ?", productID)
	} else {
		db = db.Where("product_code = ?", productCode)
	}
	if err := db.Order("purchase_date DESC, id DESC").Find(&batches).Error; err != nil {
		return err
	}

	batchIDs := make([]uint64, 0, len(batches))
	for _, batch := range batches {
		inbound := numberValue(batch["inbound_quantity"])
		remain := numberValue(batch["remaining_quantity"])
		batch["sold_quantity"] = inbound - remain
		batch["status"] = batchStatus(inbound, remain)
		if id := mapID(batch); id > 0 {
			batchIDs = append(batchIDs, id)
		}
	}

	movements := make([]map[string]any, 0)
	moveDB := r.db.WithContext(ctx).Table("inventory_movements").
		Where("tenant_id = ? AND deleted_at IS NULL", tenantID)
	if productID > 0 {
		moveDB = moveDB.Where("product_id = ?", productID)
	} else {
		moveDB = moveDB.Where("product_name = ?", stringValue(row["product_name"]))
	}
	if err := moveDB.Order("occurred_at DESC, id DESC").Limit(100).Find(&movements).Error; err != nil {
		return err
	}

	salesLinks := make([]map[string]any, 0)
	if len(batchIDs) > 0 {
		if err := r.db.WithContext(ctx).Table("sales_cost_allocations AS sca").
			Select(`sca.id, sca.inventory_batch_id, sca.sales_order_id, so.order_no AS sales_order_no,
				so.customer_name, so.order_date, sca.product_code, sca.product_name, sca.quantity,
				sca.cost_price, sca.cost_amount, so.price AS sales_price, so.total_amount AS sales_amount,
				sca.purchase_order_id, sca.purchase_order_no, sca.supplier_name, sca.purchase_date, sca.purchase_price`).
			Joins("LEFT JOIN sales_orders AS so ON so.tenant_id = sca.tenant_id AND so.id = sca.sales_order_id AND so.deleted_at IS NULL").
			Where("sca.tenant_id = ? AND sca.inventory_batch_id IN ? AND sca.deleted_at IS NULL", tenantID, batchIDs).
			Order("so.order_date DESC, sca.id DESC").
			Find(&salesLinks).Error; err != nil {
			return err
		}
	}

	row["purchaseSources"] = normalizeRows(batches)
	row["inventoryMovements"] = normalizeRows(movements)
	row["salesTrace"] = normalizeRows(salesLinks)
	return nil
}

func (r *ModuleRepository) listInventory(ctx context.Context, tenantID uint64, q query.PageQuery) ([]map[string]any, int64, error) {
	base := r.db.WithContext(ctx).Table("inventory_stocks").
		Where("tenant_id = ? AND deleted_at IS NULL", tenantID)
	if strings.TrimSpace(q.Keyword) != "" {
		base = base.Where(searchWhere(moduleSpecs["inventory"].SearchCols), searchArgs(moduleSpecs["inventory"].SearchCols, q.Keyword)...)
	}

	groupKey := "COALESCE(NULLIF(product_code, ''), 'ID:' || COALESCE(CAST(product_id AS TEXT), CAST(id AS TEXT)))"
	countRows := make([]map[string]any, 0)
	if err := base.Session(&gorm.Session{}).
		Select(groupKey + " AS group_key").
		Group(groupKey).
		Find(&countRows).Error; err != nil {
		return nil, 0, err
	}

	rows := make([]map[string]any, 0)
	order := inventoryOrder(q)
	err := base.Session(&gorm.Session{}).
		Select(`MIN(id) AS id,
			MIN(tenant_id) AS tenant_id,
			MIN(product_id) AS product_id,
			MAX(product_code) AS product_code,
			MAX(product_name) AS product_name,
			MAX(warehouse) AS warehouse,
			SUM(quantity) AS quantity,
			SUM(quantity) AS available_quantity,
			0 AS occupied_quantity,
			CASE WHEN SUM(quantity) > 0 THEN SUM(amount) / SUM(quantity) ELSE 0 END AS avg_cost,
			SUM(amount) AS amount,
			MAX(min_stock) AS min_stock,
			CASE WHEN SUM(quantity) <= MAX(min_stock) THEN ? ELSE ? END AS status,
			MAX(stock_time) AS stock_time,
			MIN(created_at) AS created_at,
			MAX(updated_at) AS updated_at`, "预警", "正常").
		Group(groupKey).
		Order(order).
		Offset(q.Offset()).
		Limit(q.PageSize).
		Find(&rows).Error
	if err == nil {
		err = r.enrichInventoryProducts(ctx, tenantID, rows)
	}
	return normalizeRows(rows), int64(len(countRows)), err
}

func (r *ModuleRepository) enrichInventoryProducts(ctx context.Context, tenantID uint64, rows []map[string]any) error {
	if len(rows) == 0 {
		return nil
	}
	codes := make([]string, 0, len(rows))
	seen := map[string]bool{}
	for _, row := range rows {
		code := stringValue(row["product_code"])
		if code != "" && !seen[code] {
			seen[code] = true
			codes = append(codes, code)
		}
	}
	if len(codes) == 0 {
		return nil
	}
	products := make([]map[string]any, 0)
	if err := r.db.WithContext(ctx).Table("products").
		Select("code, category, brand, spec, unit, barcode, qr_code, image_url, min_stock").
		Where("tenant_id = ? AND code IN ? AND deleted_at IS NULL", tenantID, codes).
		Find(&products).Error; err != nil {
		return err
	}
	byCode := map[string]map[string]any{}
	for _, product := range products {
		byCode[stringValue(product["code"])] = product
	}
	for _, row := range rows {
		product := byCode[stringValue(row["product_code"])]
		if product == nil {
			continue
		}
		for _, key := range []string{"category", "brand", "spec", "unit", "barcode", "qr_code", "image_url", "min_stock"} {
			row[key] = product[key]
		}
	}
	return nil
}

func (r *ModuleRepository) listInventoryMovements(ctx context.Context, tenantID uint64, q query.PageQuery) ([]map[string]any, int64, error) {
	rows := make([]map[string]any, 0)
	db := r.db.WithContext(ctx).Table("inventory_movements AS im").
		Joins("LEFT JOIN products AS p ON p.tenant_id = im.tenant_id AND p.deleted_at IS NULL AND (p.id = im.product_id OR (im.product_id IS NULL AND p.code = im.product_code))").
		Joins("LEFT JOIN purchase_orders AS po ON po.tenant_id = im.tenant_id AND po.id = im.source_id AND im.source_type = 'purchase' AND po.deleted_at IS NULL").
		Joins("LEFT JOIN sales_orders AS so ON so.tenant_id = im.tenant_id AND so.id = im.source_id AND im.source_type IN ('sales', 'ORDER_DELETE')").
		Joins("LEFT JOIN repair_orders AS ro ON ro.tenant_id = im.tenant_id AND ro.id = im.source_id AND im.source_type = 'repair' AND ro.deleted_at IS NULL").
		Joins("LEFT JOIN project_projects AS pp ON pp.tenant_id = im.tenant_id AND pp.id = im.source_id AND im.source_type = 'project' AND pp.deleted_at IS NULL").
		Where("im.tenant_id = ? AND im.deleted_at IS NULL", tenantID)

	if keyword := strings.TrimSpace(q.Keyword); keyword != "" {
		db = db.Where("im.product_code LIKE ?", "%"+keyword+"%")
	}
	if productCode := strings.TrimSpace(q.ProductCode); productCode != "" {
		db = db.Where("im.product_code LIKE ?", "%"+productCode+"%")
	}
	if productName := strings.TrimSpace(q.ProductName); productName != "" {
		db = db.Where("im.product_name LIKE ?", "%"+productName+"%")
	}
	if sourceType := inventorySourceType(q.SourceType); sourceType != "" {
		db = db.Where("im.source_type = ?", sourceType)
	}
	if operatorName := strings.TrimSpace(q.OperatorName); operatorName != "" {
		db = db.Where("im.operator_name LIKE ?", "%"+operatorName+"%")
	}
	if q.StartDate != "" {
		db = db.Where("im.occurred_at >= ?", q.StartDate)
	}
	if q.EndDate != "" {
		db = db.Where("im.occurred_at <= ?", q.EndDate)
	}

	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	selectSQL := `im.id, im.movement_no, im.product_id, im.product_code, im.product_name, im.warehouse,
		COALESCE(p.spec, '') AS spec,
		im.source_type, im.source_type AS business_type, im.source_id AS business_id,
		CASE im.source_type
			WHEN 'purchase' THEN po.order_no
			WHEN 'sales' THEN so.order_no
			WHEN 'ORDER_DELETE' THEN so.order_no
			WHEN 'repair' THEN ro.order_no
			WHEN 'project' THEN pp.project_no
			ELSE im.movement_no
		END AS business_no,
		CASE WHEN im.direction = 'out' THEN -im.quantity ELSE im.quantity END AS quantity_change,
		im.before_quantity, im.after_quantity,
		po.id AS purchase_id, po.order_no AS purchase_order_no,
		so.id AS sale_id, so.order_no AS sales_order_no,
		ro.id AS repair_id, ro.order_no AS repair_order_no,
		pp.id AS project_id, pp.project_no AS project_no,
		im.operator_name, im.occurred_at, im.remark`
	err := db.Select(selectSQL).
		Order(inventoryMovementOrder(q)).
		Offset(q.Offset()).
		Limit(q.PageSize).
		Find(&rows).Error
	return normalizeRows(rows), total, err
}

func (r *ModuleRepository) mergeInventoryRow(ctx context.Context, tenantID uint64, row map[string]any) error {
	productID := uintValue(row["product_id"])
	productCode := stringValue(row["product_code"])
	warehouse := firstNotBlank(stringValue(row["warehouse"]), "主仓库")
	merged := map[string]any{}
	_ = warehouse
	db := r.db.WithContext(ctx).Table("inventory_stocks").
		Where("tenant_id = ? AND deleted_at IS NULL", tenantID)
	if productCode != "" {
		db = db.Where("product_code = ?", productCode)
	} else if productID > 0 {
		db = db.Where("product_id = ?", productID)
	} else {
		return nil
	}
	if err := db.Select(`MIN(id) AS id,
		MIN(tenant_id) AS tenant_id,
		MIN(product_id) AS product_id,
		MAX(product_code) AS product_code,
		MAX(product_name) AS product_name,
		MAX(warehouse) AS warehouse,
		SUM(quantity) AS quantity,
		SUM(quantity) AS available_quantity,
		0 AS occupied_quantity,
		CASE WHEN SUM(quantity) > 0 THEN SUM(amount) / SUM(quantity) ELSE 0 END AS avg_cost,
		SUM(amount) AS amount,
		MAX(min_stock) AS min_stock,
		CASE WHEN SUM(quantity) <= MAX(min_stock) THEN ? ELSE ? END AS status,
		MAX(stock_time) AS stock_time,
		MIN(created_at) AS created_at,
		MAX(updated_at) AS updated_at`, "预警", "正常").
		Group("tenant_id").
		Take(&merged).Error; err != nil {
		return err
	}
	for key, value := range merged {
		row[key] = value
	}
	return nil
}

func (r *ModuleRepository) createRepairItems(ctx context.Context, tenantID uint64, operatorID uint64, orderID uint64, row map[string]any) error {
	now := time.Now()
	items := make([]map[string]any, 0)
	for _, charge := range repairInputRows(row, "charge_items") {
		amount := repairInputAmount(charge)
		if amount <= 0 {
			continue
		}
		qty := repairInputQuantity(charge)
		items = append(items, map[string]any{
			"tenant_id":       tenantID,
			"repair_order_id": orderID,
			"item_type":       "charge",
			"item_name":       firstNotBlank(repairInputString(charge, "item_name", "itemName", "charge_item", "chargeItem", "name"), "维修收费"),
			"quantity":        qty,
			"price":           repairInputPrice(charge, amount, qty),
			"amount":          amount,
			"cost_price":      0,
			"cost_amount":     0,
			"remark":          repairInputString(charge, "remark"),
			"created_by":      operatorID,
			"updated_by":      operatorID,
			"created_at":      now,
			"updated_at":      now,
		})
	}
	serviceItems := []struct {
		name  string
		field string
	}{
		{"上门费用", "onsite_fee"},
		{"检测费用", "detection_fee"},
		{"安装费用", "installation_fee"},
	}
	for _, item := range serviceItems {
		amount := numberValue(row[item.field])
		if amount <= 0 {
			continue
		}
		items = append(items, map[string]any{
			"tenant_id":       tenantID,
			"repair_order_id": orderID,
			"item_type":       "service",
			"item_name":       item.name,
			"quantity":        1,
			"price":           amount,
			"amount":          amount,
			"cost_price":      0,
			"cost_amount":     0,
			"created_by":      operatorID,
			"updated_by":      operatorID,
			"created_at":      now,
			"updated_at":      now,
		})
	}
	partRows := repairInputRows(row, "repair_parts")
	for _, part := range partRows {
		qty := repairInputQuantity(part)
		amount := repairInputAmount(part)
		if qty <= 0 || amount <= 0 {
			continue
		}
		costPrice := numberValue(firstPresent(part, "cost_price", "costPrice", "purchase_price", "purchasePrice"))
		costAmount := numberValue(firstPresent(part, "cost_amount", "costAmount"))
		if costAmount == 0 {
			costAmount = qty * costPrice
		}
		items = append(items, map[string]any{
			"tenant_id":          tenantID,
			"repair_order_id":    orderID,
			"item_type":          "part",
			"item_name":          firstNotBlank(repairInputString(part, "product_name", "productName", "item_name", "itemName"), "维修配件"),
			"product_id":         nilIfZero(uintValue(firstPresent(part, "product_id", "productId"))),
			"inventory_batch_id": nilIfZero(uintValue(firstPresent(part, "inventory_batch_id", "inventoryBatchId", "batch_id", "batchId"))),
			"product_code":       repairInputString(part, "product_code", "productCode"),
			"product_name":       repairInputString(part, "product_name", "productName"),
			"quantity":           qty,
			"price":              repairInputPrice(part, amount, qty),
			"amount":             amount,
			"cost_price":         costPrice,
			"cost_amount":        costAmount,
			"remark":             repairInputString(part, "remark"),
			"created_by":         operatorID,
			"updated_by":         operatorID,
			"created_at":         now,
			"updated_at":         now,
		})
	}
	qty := numberValue(row["quantity"])
	if len(partRows) == 0 && qty > 0 && !empty(row["product_name"]) {
		items = append(items, map[string]any{
			"tenant_id":       tenantID,
			"repair_order_id": orderID,
			"item_type":       "part",
			"item_name":       firstNotBlank(stringValue(row["product_name"]), "维修配件"),
			"product_id":      nilIfZero(uintValue(row["product_id"])),
			"product_code":    row["product_code"],
			"product_name":    row["product_name"],
			"quantity":        qty,
			"price":           numberValue(row["price"]),
			"amount":          numberValue(row["parts_amount"]),
			"cost_price":      numberValue(row["cost_price"]),
			"cost_amount":     numberValue(row["parts_cost"]),
			"created_by":      operatorID,
			"updated_by":      operatorID,
			"created_at":      now,
			"updated_at":      now,
		})
	}
	for _, outsource := range repairInputRows(row, "outsource_items") {
		amount := repairInputAmount(outsource)
		if amount <= 0 {
			continue
		}
		qty := repairInputQuantity(outsource)
		items = append(items, map[string]any{
			"tenant_id":       tenantID,
			"repair_order_id": orderID,
			"item_type":       "outsource",
			"item_name":       firstNotBlank(repairInputString(outsource, "service_project", "serviceProject", "item_name", "itemName"), "外协费用"),
			"supplier_id":     nilIfZero(uintValue(firstPresent(outsource, "supplier_id", "supplierId"))),
			"supplier_name":   repairInputString(outsource, "supplier_name", "supplierName"),
			"service_project": repairInputString(outsource, "service_project", "serviceProject", "item_name", "itemName"),
			"quantity":        qty,
			"price":           repairInputPrice(outsource, amount, qty),
			"amount":          0,
			"cost_price":      repairInputPrice(outsource, amount, qty),
			"cost_amount":     amount,
			"remark":          repairInputString(outsource, "remark"),
			"created_by":      operatorID,
			"updated_by":      operatorID,
			"created_at":      now,
			"updated_at":      now,
		})
	}
	for _, costItem := range repairInputRows(row, "cost_items") {
		amount := repairInputAmount(costItem)
		if amount <= 0 {
			continue
		}
		costType := firstNotBlank(repairInputString(costItem, "cost_type", "costType", "item_type", "itemType"), "其他费用")
		if costType == "配件成本" || strings.EqualFold(costType, "part") || strings.EqualFold(costType, "parts") {
			continue
		}
		qty := repairInputQuantity(costItem)
		items = append(items, map[string]any{
			"tenant_id":       tenantID,
			"repair_order_id": orderID,
			"item_type":       repairCostItemType(costType),
			"item_name":       firstNotBlank(repairInputString(costItem, "item_name", "itemName", "name"), costType),
			"quantity":        qty,
			"price":           0,
			"amount":          0,
			"cost_price":      repairInputPrice(costItem, amount, qty),
			"cost_amount":     amount,
			"remark":          repairInputString(costItem, "remark"),
			"created_by":      operatorID,
			"updated_by":      operatorID,
			"created_at":      now,
			"updated_at":      now,
		})
	}
	for _, item := range items {
		if err := r.db.WithContext(ctx).Table("repair_order_items").Create(item).Error; err != nil {
			return err
		}
	}
	return nil
}

func (r *ModuleRepository) consumeRepairPartInventory(ctx context.Context, tenantID uint64, operatorID uint64, orderID uint64, order map[string]any) (bool, error) {
	items := make([]map[string]any, 0)
	if err := r.db.WithContext(ctx).Table("repair_order_items").
		Where("tenant_id = ? AND repair_order_id = ? AND item_type = ? AND quantity > 0 AND deleted_at IS NULL", tenantID, orderID, "part").
		Order("id ASC").Find(&items).Error; err != nil {
		return false, err
	}
	processed := false
	for _, item := range items {
		qty := numberValue(item["quantity"])
		if qty <= 0 {
			continue
		}
		batchID := uintValue(item["inventory_batch_id"])
		if batchID == 0 {
			batchID = uintValue(item["batch_id"])
		}
		if batchID == 0 {
			return false, apperrors.New(40040, "维修配件必须选择有库存的采购批次", 400)
		}
		allocations, err := r.matchSelectedPurchaseBatches(ctx, tenantID, uintValue(item["product_id"]), stringValue(item["product_code"]), firstNotBlank(stringValue(item["warehouse"]), "主仓库"), qty, []uint64{batchID}, true, item)
		if err != nil {
			return false, err
		}
		costAmount := 0.0
		costQty := 0.0
		for _, allocation := range allocations {
			costAmount += allocation.CostAmount
			costQty += allocation.Quantity
		}
		costPrice := numberValue(item["cost_price"])
		if costQty > 0 {
			costPrice = costAmount / costQty
		}
		if err := r.db.WithContext(ctx).Table("repair_order_items").Where("tenant_id = ? AND id = ?", tenantID, mapID(item)).Updates(map[string]any{
			"cost_price":  costPrice,
			"cost_amount": costAmount,
			"updated_by":  operatorID,
			"updated_at":  time.Now(),
		}).Error; err != nil {
			return false, err
		}
		movementRow := map[string]any{
			"product_id":   item["product_id"],
			"product_code": item["product_code"],
			"product_name": item["product_name"],
			"warehouse":    firstNotBlank(stringValue(item["warehouse"]), "主仓库"),
			"quantity":     qty,
			"cost_price":   costPrice,
			"price":        costPrice,
			"order_no":     order["order_no"],
		}
		if err := r.applyInventoryMovement(ctx, tenantID, operatorID, "repair", orderID, movementRow, "out"); err != nil {
			return false, err
		}
		processed = true
	}
	if processed {
		if err := r.refreshRepairOrderCosts(ctx, tenantID, operatorID, orderID); err != nil {
			return false, err
		}
	}
	return processed, nil
}

func (r *ModuleRepository) refreshRepairOrderCosts(ctx context.Context, tenantID uint64, operatorID uint64, orderID uint64) error {
	summary := map[string]any{}
	if err := r.db.WithContext(ctx).Table("repair_order_items").
		Select(`COALESCE(SUM(CASE WHEN item_type = 'part' THEN amount ELSE 0 END), 0) AS parts_amount,
			COALESCE(SUM(CASE WHEN item_type = 'part' THEN cost_amount ELSE 0 END), 0) AS parts_cost,
			COALESCE(SUM(CASE WHEN item_type = 'outsource' THEN cost_amount ELSE 0 END), 0) AS outsource_cost,
			COALESCE(SUM(CASE WHEN item_type IN ('labor', 'transport', 'other') THEN cost_amount ELSE 0 END), 0) AS other_cost`).
		Where("tenant_id = ? AND repair_order_id = ? AND deleted_at IS NULL", tenantID, orderID).
		Take(&summary).Error; err != nil {
		return err
	}
	order := map[string]any{}
	if err := r.db.WithContext(ctx).Table("repair_orders").Select("total_amount").Where("tenant_id = ? AND id = ?", tenantID, orderID).Take(&order).Error; err != nil {
		return err
	}
	partsCost := numberValue(summary["parts_cost"])
	outsourceCost := numberValue(summary["outsource_cost"])
	otherCost := numberValue(summary["other_cost"])
	costAmount := partsCost + outsourceCost + otherCost
	return r.db.WithContext(ctx).Table("repair_orders").Where("tenant_id = ? AND id = ?", tenantID, orderID).Updates(map[string]any{
		"parts_amount":   numberValue(summary["parts_amount"]),
		"parts_cost":     partsCost,
		"outsource_cost": outsourceCost,
		"cost_amount":    costAmount,
		"profit_amount":  numberValue(order["total_amount"]) - costAmount,
		"updated_by":     operatorID,
		"updated_at":     time.Now(),
	}).Error
}

func (r *ModuleRepository) processRepairCompletion(ctx context.Context, tenantID uint64, operatorID uint64, orderID uint64) error {
	row := map[string]any{}
	if err := r.db.WithContext(ctx).Table("repair_orders").Where("tenant_id = ? AND id = ? AND deleted_at IS NULL", tenantID, orderID).Take(&row).Error; err != nil {
		return err
	}
	if !isRepairCompleted(stringValue(row["repair_status"])) {
		return nil
	}
	now := time.Now()
	updates := map[string]any{"updated_by": operatorID, "updated_at": now}
	if !boolValue(row["inventory_done"]) {
		processed, err := r.consumeRepairPartInventory(ctx, tenantID, operatorID, orderID, row)
		if err != nil {
			return err
		}
		if processed {
			updates["inventory_done"] = true
		}
	}
	if !boolValue(row["settlement_done"]) {
		financeProcessed := false
		if numberValue(row["total_amount"]) > 0 {
			customer := r.findCustomer(ctx, tenantID, uintValue(row["customer_id"]), stringValue(row["customer_name"]))
			row["source_type"] = "repair"
			if err := r.createReceivable(ctx, tenantID, operatorID, customer, orderID, row, numberValue(row["total_amount"]), 0, numberValue(row["total_amount"])); err != nil {
				return err
			}
			if err := r.refreshCustomerReceivable(ctx, tenantID, uintValue(customer["id"])); err != nil {
				return err
			}
			financeProcessed = true
		}
		outsourceProcessed, err := r.createRepairOutsourcePayables(ctx, tenantID, operatorID, orderID, row)
		if err != nil {
			return err
		}
		if financeProcessed || outsourceProcessed {
			updates["settlement_done"] = true
		}
	}
	if len(updates) > 2 {
		return r.db.WithContext(ctx).Table("repair_orders").Where("tenant_id = ? AND id = ?", tenantID, orderID).Updates(updates).Error
	}
	return nil
}

func (r *ModuleRepository) createRepairOutsourcePayables(ctx context.Context, tenantID uint64, operatorID uint64, orderID uint64, order map[string]any) (bool, error) {
	items := make([]map[string]any, 0)
	if err := r.db.WithContext(ctx).Table("repair_order_items").
		Where("tenant_id = ? AND repair_order_id = ? AND item_type = ? AND cost_amount > 0 AND deleted_at IS NULL", tenantID, orderID, "outsource").
		Order("id ASC").Find(&items).Error; err != nil {
		return false, err
	}
	if len(items) == 0 {
		return false, nil
	}
	type supplierPayable struct {
		supplierID   uint64
		supplierName string
		amount       float64
	}
	grouped := map[string]*supplierPayable{}
	for _, item := range items {
		supplierID := uintValue(item["supplier_id"])
		supplierName := stringValue(item["supplier_name"])
		if supplierID == 0 && supplierName == "" {
			return false, apperrors.New(40039, "维修外协费用必须选择供应商", 400)
		}
		key := fmt.Sprintf("%d:%s", supplierID, supplierName)
		if grouped[key] == nil {
			grouped[key] = &supplierPayable{supplierID: supplierID, supplierName: supplierName}
		}
		grouped[key].amount += numberValue(item["cost_amount"])
	}
	now := time.Now()
	billDate := timeValue(order["registered_at"], now)
	processed := false
	for _, item := range grouped {
		if item.amount <= 0 {
			continue
		}
		var duplicate int64
		dupDB := r.db.WithContext(ctx).Table("payables").
			Where("tenant_id = ? AND source_type = ? AND source_id = ? AND source_no = ? AND deleted_at IS NULL", tenantID, "repair_outsource", orderID, stringValue(order["order_no"]))
		if item.supplierID > 0 {
			dupDB = dupDB.Where("supplier_id = ?", item.supplierID)
		} else {
			dupDB = dupDB.Where("supplier_name = ?", item.supplierName)
		}
		if err := dupDB.Count(&duplicate).Error; err != nil {
			return false, err
		}
		if duplicate > 0 {
			continue
		}
		payable := map[string]any{
			"tenant_id":      tenantID,
			"payable_no":     nextNo("payables"),
			"supplier_id":    nilIfZero(item.supplierID),
			"supplier_name":  item.supplierName,
			"source_type":    "repair_outsource",
			"source_id":      orderID,
			"source_no":      order["order_no"],
			"total_amount":   item.amount,
			"paid_amount":    0,
			"balance_amount": item.amount,
			"bill_date":      billDate,
			"due_date":       billDate.AddDate(0, 0, 30),
			"status":         payableStatus(item.amount, 0),
			"created_by":     operatorID,
			"updated_by":     operatorID,
			"created_at":     now,
			"updated_at":     now,
		}
		if err := r.db.WithContext(ctx).Table("payables").Create(payable).Error; err != nil {
			return false, err
		}
		if err := r.refreshSupplierPayable(ctx, tenantID, item.supplierID); err != nil {
			return false, err
		}
		processed = true
	}
	return processed, nil
}

func isRepairCompleted(status string) bool {
	status = strings.TrimSpace(strings.ToLower(status))
	return status == "completed" || status == "已完成" || status == "结算完成" || status == "维修完成"
}

func (r *ModuleRepository) createSalesItems(ctx context.Context, tenantID uint64, operatorID uint64, orderID uint64, row map[string]any, allocations []costAllocation) error {
	now := time.Now()
	price := numberValue(row["price"])
	for _, allocation := range allocations {
		amount := allocation.Quantity * price
		item := map[string]any{
			"tenant_id":              tenantID,
			"sales_order_id":         orderID,
			"product_id":             nilIfZero(allocation.ProductID),
			"product_code":           allocation.ProductCode,
			"product_name":           allocation.ProductName,
			"quantity":               allocation.Quantity,
			"price":                  price,
			"amount":                 amount,
			"purchase_source":        "purchase_order",
			"inventory_batch_id":     nilIfZero(allocation.BatchID),
			"purchase_order_id":      nilIfZero(allocation.PurchaseOrderID),
			"purchase_order_item_id": nilIfZero(allocation.PurchaseOrderItemID),
			"purchase_order_no":      allocation.PurchaseOrderNo,
			"supplier_id":            nilIfZero(allocation.SupplierID),
			"supplier_name":          allocation.SupplierName,
			"purchase_date":          nilIfZeroTime(allocation.PurchaseDate),
			"purchase_price":         allocation.PurchasePrice,
			"cost_price":             allocation.CostPrice,
			"cost_amount":            allocation.CostAmount,
			"profit_amount":          amount - allocation.CostAmount,
			"created_by":             operatorID,
			"updated_by":             operatorID,
			"created_at":             now,
			"updated_at":             now,
		}
		if err := r.db.WithContext(ctx).Table("sales_order_items").Create(item).Error; err != nil {
			return err
		}
		itemID := mapID(item)
		if err := r.createSalesCostAllocation(ctx, tenantID, operatorID, orderID, itemID, allocation, now); err != nil {
			return err
		}
	}
	return nil
}

func (r *ModuleRepository) createSalesCostAllocation(ctx context.Context, tenantID uint64, operatorID uint64, orderID uint64, itemID uint64, allocation costAllocation, now time.Time) error {
	record := map[string]any{
		"tenant_id":           tenantID,
		"sales_order_id":      orderID,
		"sales_order_item_id": nilIfZero(itemID),
		"inventory_batch_id":  nilIfZero(allocation.BatchID),
		"product_id":          nilIfZero(allocation.ProductID),
		"product_code":        allocation.ProductCode,
		"product_name":        allocation.ProductName,
		"quantity":            allocation.Quantity,
		"cost_price":          allocation.CostPrice,
		"cost_amount":         allocation.CostAmount,
		"purchase_order_id":   nilIfZero(allocation.PurchaseOrderID),
		"purchase_order_no":   allocation.PurchaseOrderNo,
		"supplier_id":         nilIfZero(allocation.SupplierID),
		"supplier_name":       allocation.SupplierName,
		"purchase_date":       nilIfZeroTime(allocation.PurchaseDate),
		"purchase_price":      allocation.PurchasePrice,
		"created_by":          operatorID,
		"updated_by":          operatorID,
		"created_at":          now,
		"updated_at":          now,
	}
	return r.db.WithContext(ctx).Table("sales_cost_allocations").Create(record).Error
}

func (r *ModuleRepository) consumeSalesBatches(ctx context.Context, tenantID uint64, orderID uint64, row map[string]any) ([]costAllocation, error) {
	qty := numberValue(row["quantity"])
	if qty <= 0 {
		return nil, nil
	}
	return r.matchSelectedPurchaseBatches(ctx, tenantID, uintValue(row["product_id"]), stringValue(row["product_code"]), firstNotBlank(stringValue(row["warehouse"]), "主仓库"), qty, selectedPurchaseBatchIDs(row), true, row)
}

func (r *ModuleRepository) matchInventoryBatches(ctx context.Context, tenantID uint64, productID uint64, productCode string, warehouse string, qty float64, method string, batchID uint64, consume bool, orderID uint64, itemID uint64, row map[string]any) ([]costAllocation, error) {
	return r.matchSelectedPurchaseBatches(ctx, tenantID, productID, productCode, warehouse, qty, []uint64{batchID}, consume, row)
}

func (r *ModuleRepository) matchSelectedPurchaseBatches(ctx context.Context, tenantID uint64, productID uint64, productCode string, warehouse string, qty float64, batchIDs []uint64, consume bool, row map[string]any) ([]costAllocation, error) {
	if qty <= 0 {
		return nil, nil
	}
	batchIDs = compactUintIDs(batchIDs)
	if len(batchIDs) == 0 {
		return nil, apperrors.New(40015, "销售商品必须选择采购来源", 400)
	}
	rows := make([]map[string]any, 0)
	selectedDB := r.db.WithContext(ctx).Table("inventory_batches").
		Where("tenant_id = ? AND id IN ? AND remaining_quantity > 0 AND deleted_at IS NULL", tenantID, batchIDs)
	if err := selectedDB.Find(&rows).Error; err != nil {
		return nil, err
	}
	if len(rows) != len(batchIDs) {
		return nil, apperrors.New(40016, "选择的采购库存不可用或已销售完成", 400)
	}
	rowsByID := make(map[uint64]map[string]any, len(rows))
	for _, selected := range rows {
		if productID > 0 && uintValue(selected["product_id"]) != productID {
			return nil, apperrors.New(40021, "选择的采购来源与销售商品不一致", 400)
		}
		if productID == 0 && productCode != "" && stringValue(selected["product_code"]) != productCode {
			return nil, apperrors.New(40021, "选择的采购来源与销售商品不一致", 400)
		}
		if warehouse != "" && stringValue(selected["warehouse"]) != warehouse {
			return nil, apperrors.New(40021, "选择的采购来源与销售仓库不一致", 400)
		}
		rowsByID[mapID(selected)] = selected
	}
	orderedRows := make([]map[string]any, 0, len(batchIDs))
	for _, id := range batchIDs {
		if selected, ok := rowsByID[id]; ok {
			orderedRows = append(orderedRows, selected)
		}
	}
	return r.consumeBatchRows(ctx, orderedRows, qty, 0, consume)
}
func (r *ModuleRepository) consumeBatchRows(ctx context.Context, rows []map[string]any, qty float64, overrideCost float64, consume bool) ([]costAllocation, error) {
	remaining := qty
	allocations := make([]costAllocation, 0)
	for _, batch := range rows {
		if remaining <= 0 {
			break
		}
		batchRemain := numberValue(batch["remaining_quantity"])
		take := remaining
		if take > batchRemain {
			take = batchRemain
		}
		cost := numberValue(batch["purchase_price"])
		if overrideCost > 0 {
			cost = overrideCost
		}
		allocation := costAllocation{
			BatchID:             mapID(batch),
			ProductID:           uintValue(batch["product_id"]),
			ProductCode:         stringValue(batch["product_code"]),
			ProductName:         stringValue(batch["product_name"]),
			Quantity:            take,
			CostPrice:           cost,
			CostAmount:          take * cost,
			PurchaseOrderID:     uintValue(batch["purchase_order_id"]),
			PurchaseOrderItemID: uintValue(batch["purchase_order_item_id"]),
			PurchaseOrderNo:     stringValue(batch["purchase_order_no"]),
			SupplierID:          uintValue(batch["supplier_id"]),
			SupplierName:        stringValue(batch["supplier_name"]),
			PurchaseDate:        timeValue(batch["purchase_date"], time.Time{}),
			PurchasePrice:       numberValue(batch["purchase_price"]),
		}
		allocations = append(allocations, allocation)
		if consume {
			newRemain := batchRemain - take
			status := batchStatus(numberValue(batch["inbound_quantity"]), newRemain)
			if err := r.db.WithContext(ctx).Table("inventory_batches").Where("id = ?", allocation.BatchID).Updates(map[string]any{
				"remaining_quantity": newRemain,
				"status":             status,
				"updated_at":         time.Now(),
			}).Error; err != nil {
				return nil, err
			}
		}
		remaining -= take
	}
	if remaining > 0 {
		return nil, apperrors.New(40011, fmt.Sprintf("库存批次不足，缺少 %.2f", remaining), 400)
	}
	return allocations, nil
}

func batchStatus(inbound float64, remaining float64) string {
	if remaining <= 0 {
		return "已销售完成"
	}
	if inbound > 0 && remaining < inbound {
		return "销售中"
	}
	return "未销售"
}

func inventoryMovementRemark(sourceType string, sourceID uint64, row map[string]any) string {
	parts := []string{fmt.Sprintf("%s:%d", sourceType, sourceID)}
	if orderNo := stringValue(row["order_no"]); orderNo != "" {
		parts = append(parts, "销售单号:"+orderNo)
	}
	if purchaseNo := stringValue(row["purchase_order_no"]); purchaseNo != "" {
		parts = append(parts, "采购单号:"+purchaseNo)
	}
	return strings.Join(parts, " ")
}

func (r *ModuleRepository) findOrCreateProduct(ctx context.Context, tenantID uint64, operatorID uint64, row map[string]any) (map[string]any, error) {
	productID := uintValue(row["product_id"])
	productCode := strings.TrimSpace(stringValue(row["product_code"]))
	productName := strings.TrimSpace(firstNotBlank(stringValue(row["product_name"]), stringValue(row["name"])))
	product := map[string]any{}
	db := r.db.WithContext(ctx).Table("products").Where("tenant_id = ? AND deleted_at IS NULL", tenantID)
	var err error
	switch {
	case productID > 0:
		err = db.Where("id = ?", productID).Take(&product).Error
	case productCode != "":
		err = db.Where("code = ?", productCode).Take(&product).Error
	case productName != "":
		err = db.Where("name = ?", productName).Take(&product).Error
	default:
		return nil, nil
	}
	if err == nil {
		return product, nil
	}
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, err
	}
	if productName == "" {
		return nil, apperrors.New(40010, "商品不存在，请先选择或输入商品名称", 400)
	}
	now := time.Now()
	if productCode == "" {
		productCode = nextProductCode()
	}
	product = map[string]any{
		"tenant_id":  tenantID,
		"code":       productCode,
		"name":       productName,
		"unit":       firstNotBlank(stringValue(row["unit"]), "台"),
		"min_stock":  0,
		"status":     "active",
		"created_by": operatorID,
		"updated_by": operatorID,
		"created_at": now,
		"updated_at": now,
	}
	if err := r.db.WithContext(ctx).Table("products").Create(product).Error; err != nil {
		return nil, err
	}
	created := map[string]any{}
	if err := r.db.WithContext(ctx).Table("products").Where("tenant_id = ? AND code = ?", tenantID, productCode).Take(&created).Error; err != nil {
		return nil, err
	}
	product = created
	return product, nil
}

func (r *ModuleRepository) applyInventoryMovement(ctx context.Context, tenantID uint64, operatorID uint64, sourceType string, sourceID uint64, row map[string]any, direction string) error {
	qty := numberValue(row["quantity"])
	if qty <= 0 {
		return nil
	}
	productID := uintValue(row["product_id"])
	productCode := stringValue(row["product_code"])
	productName := stringValue(row["product_name"])
	warehouse := firstNotBlank(stringValue(row["warehouse"]), "主仓库")
	unitCost := numberValue(row["cost_price"])
	if sourceType == "purchase" || unitCost == 0 {
		unitCost = numberValue(row["price"])
	}
	stock, err := r.findStock(ctx, tenantID, productID, productCode, warehouse)
	if err != nil {
		return err
	}
	oldQty := numberValue(stock["quantity"])
	oldAvg := numberValue(stock["avg_cost"])
	minStock := numberValue(stock["min_stock"])
	newQty := oldQty
	newAvg := oldAvg
	if direction == "in" {
		newQty = oldQty + qty
		if newQty > 0 {
			newAvg = ((oldQty * oldAvg) + (qty * unitCost)) / newQty
		}
	} else {
		if oldQty < qty {
			if sourceType != "sales" {
				return apperrors.New(40011, fmt.Sprintf("库存不足：%s 当前库存 %.2f，需要 %.2f", productName, oldQty, qty), 400)
			}
			oldQty = qty
		}
		newQty = oldQty - qty
		if unitCost == 0 {
			unitCost = oldAvg
		}
	}
	status := "正常"
	if newQty <= minStock {
		status = "预警"
	}
	now := time.Now()
	operatorName := ""
	if operatorID > 0 {
		user := map[string]any{}
		if err := r.db.WithContext(ctx).Table("users").Select("real_name, username").Where("id = ?", operatorID).Take(&user).Error; err == nil {
			operatorName = firstNotBlank(stringValue(user["real_name"]), stringValue(user["username"]))
		}
	}
	updates := map[string]any{
		"product_id":   nilIfZero(productID),
		"product_code": productCode,
		"product_name": productName,
		"warehouse":    warehouse,
		"quantity":     newQty,
		"avg_cost":     newAvg,
		"amount":       newQty * newAvg,
		"status":       status,
		"stock_time":   now,
		"updated_by":   operatorID,
		"updated_at":   now,
	}
	if mapID(stock) == 0 {
		updates["tenant_id"] = tenantID
		updates["min_stock"] = 0
		updates["created_by"] = operatorID
		updates["created_at"] = now
		if err := r.db.WithContext(ctx).Table("inventory_stocks").Create(updates).Error; err != nil {
			return err
		}
	} else if err := r.db.WithContext(ctx).Table("inventory_stocks").Where("id = ?", mapID(stock)).Updates(updates).Error; err != nil {
		return err
	}
	movement := map[string]any{
		"tenant_id":       tenantID,
		"movement_no":     nextNo("inventory"),
		"product_id":      nilIfZero(productID),
		"product_code":    productCode,
		"product_name":    productName,
		"warehouse":       warehouse,
		"source_type":     sourceType,
		"source_id":       nilIfZero(sourceID),
		"direction":       direction,
		"quantity":        qty,
		"before_quantity": oldQty,
		"after_quantity":  newQty,
		"unit_cost":       unitCost,
		"amount":          qty * unitCost,
		"operator_name":   operatorName,
		"occurred_at":     now,
		"remark":          inventoryMovementRemark(sourceType, sourceID, row),
		"created_by":      operatorID,
		"updated_by":      operatorID,
		"created_at":      now,
		"updated_at":      now,
	}
	if err := r.db.WithContext(ctx).Table("inventory_movements").Create(movement).Error; err != nil {
		return err
	}
	return r.syncProductPrice(ctx, tenantID, productID, sourceType, row, unitCost)
}

func (r *ModuleRepository) findStock(ctx context.Context, tenantID uint64, productID uint64, productCode string, warehouse string) (map[string]any, error) {
	stock := map[string]any{}
	db := r.db.WithContext(ctx).Table("inventory_stocks").Where("tenant_id = ? AND warehouse = ? AND deleted_at IS NULL", tenantID, warehouse)
	if strings.TrimSpace(productCode) != "" {
		db = db.Where("product_code = ?", productCode)
	} else if productID > 0 {
		db = db.Where("product_id = ?", productID)
	} else {
		return map[string]any{}, nil
	}
	err := db.Order("id ASC").Take(&stock).Error
	if err == gorm.ErrRecordNotFound {
		return map[string]any{}, nil
	}
	return stock, err
}

func (r *ModuleRepository) currentStockCost(ctx context.Context, tenantID uint64, productID uint64, productCode string, warehouse string) float64 {
	stock, err := r.findStock(ctx, tenantID, productID, productCode, warehouse)
	if err != nil {
		return 0
	}
	return numberValue(stock["avg_cost"])
}

func (r *ModuleRepository) syncProductPrice(ctx context.Context, tenantID uint64, productID uint64, sourceType string, row map[string]any, unitCost float64) error {
	return nil
}

func (r *ModuleRepository) applySalesSettlement(ctx context.Context, tenantID uint64, operatorID uint64, orderID uint64, row map[string]any) error {
	total := numberValue(row["total_amount"])
	if total <= 0 {
		return nil
	}
	customer := r.findCustomer(ctx, tenantID, uintValue(row["customer_id"]), stringValue(row["customer_name"]))
	row["received_amount"] = 0
	row["receivable_amount"] = total
	if err := r.updateSalesCollection(ctx, tenantID, orderID, 0, total); err != nil {
		return err
	}
	if err := r.checkCreditLimit(ctx, tenantID, customer, total); err != nil {
		return err
	}
	if err := r.createReceivable(ctx, tenantID, operatorID, customer, orderID, row, total, 0, total); err != nil {
		return err
	}
	return r.refreshCustomerReceivable(ctx, tenantID, uintValue(customer["id"]))
}

func (r *ModuleRepository) applyPurchaseSettlement(ctx context.Context, tenantID uint64, operatorID uint64, orderID uint64, row map[string]any) error {
	total := numberValue(row["total_amount"])
	if total <= 0 {
		return nil
	}
	row["paid_amount"] = 0
	row["payable_amount"] = total
	if err := r.createPayable(ctx, tenantID, operatorID, orderID, row, total, 0, total); err != nil {
		return err
	}
	return r.refreshSupplierPayable(ctx, tenantID, uintValue(row["supplier_id"]))
}

func (r *ModuleRepository) createPayable(ctx context.Context, tenantID uint64, operatorID uint64, orderID uint64, row map[string]any, total float64, paid float64, balance float64) error {
	now := time.Now()
	billDate := timeValue(row["order_date"], now)
	payable := map[string]any{
		"tenant_id":      tenantID,
		"payable_no":     nextNo("payables"),
		"supplier_id":    nilIfZero(uintValue(row["supplier_id"])),
		"supplier_name":  stringValue(row["supplier_name"]),
		"source_type":    "purchase",
		"source_id":      orderID,
		"source_no":      row["order_no"],
		"total_amount":   total,
		"paid_amount":    paid,
		"balance_amount": balance,
		"bill_date":      billDate,
		"due_date":       billDate.AddDate(0, 0, 30),
		"status":         payableStatus(balance, paid),
		"created_by":     operatorID,
		"updated_by":     operatorID,
		"created_at":     now,
		"updated_at":     now,
	}
	return r.db.WithContext(ctx).Table("payables").Create(payable).Error
}

func payableStatus(balance float64, paid ...float64) string {
	if balance <= 0 {
		return "paid"
	}
	if len(paid) > 0 && paid[0] > 0 {
		return "partial"
	}
	return "unpaid"
}

func (r *ModuleRepository) refreshSupplierPayable(ctx context.Context, tenantID uint64, supplierID uint64) error {
	if supplierID == 0 {
		return nil
	}
	var balance float64
	if err := r.db.WithContext(ctx).Table("payables").
		Select("COALESCE(SUM(balance_amount), 0)").
		Where("tenant_id = ? AND supplier_id = ? AND status <> ? AND deleted_at IS NULL", tenantID, supplierID, "paid").
		Scan(&balance).Error; err != nil {
		return err
	}
	return r.db.WithContext(ctx).Table("suppliers").Where("tenant_id = ? AND id = ?", tenantID, supplierID).Update("payable_balance", balance).Error
}

func (r *ModuleRepository) applyProjectSettlement(ctx context.Context, tenantID uint64, operatorID uint64, projectID uint64, row map[string]any) error {
	received := numberValue(row["received_amount"])
	settle := numberValue(row["settle_amount"])
	amount := received
	if amount <= 0 {
		amount = settle
	}
	if projectID == 0 || amount <= 0 {
		return nil
	}
	return r.createFinanceReceipt(ctx, tenantID, operatorID, stringValue(row["customer_name"]), amount, stringValue(row["project_no"]), "工程收款", stringValue(row["account_name"]))
}

func (r *ModuleRepository) prepareFinanceAccount(row map[string]any, creating bool) error {
	if empty(row["name"]) {
		return apperrors.New(40037, "资金账户名称不能为空", 400)
	}
	if empty(row["code"]) {
		row["code"] = financeAccountCode(stringValue(row["name"]))
	}
	if empty(row["account_type"]) {
		row["account_type"] = financeAccountType(stringValue(row["name"]))
	}
	if empty(row["status"]) {
		row["status"] = "active"
	}
	if _, ok := row["opening_balance"]; ok {
		opening := numberValue(row["opening_balance"])
		row["opening_balance"] = opening
	}
	if creating {
		row["balance"] = numberValue(row["opening_balance"])
	} else if _, ok := row["balance"]; ok {
		row["balance"] = numberValue(row["balance"])
	}
	return nil
}

func (r *ModuleRepository) validateManualFinanceRecord(row map[string]any) error {
	recordType := strings.TrimSpace(stringValue(row["record_type"]))
	businessType := strings.TrimSpace(stringValue(row["business_type"]))
	if strings.Contains(businessType, "销售") || strings.Contains(businessType, "采购") || strings.Contains(businessType, "维修") || strings.Contains(businessType, "工程") {
		return apperrors.New(40033, "销售、采购、维修、工程财务流水必须由业务单据自动生成", 400)
	}
	if recordType == "收款" || recordType == "付款" {
		return apperrors.New(40034, "业务收付款请在应收账款或应付账款中登记，禁止在资金流水里直接录入", 400)
	}
	row["business_type"] = "日常记账"
	row["source_type"] = "daily"
	if empty(row["account_name"]) {
		row["account_name"] = "现金"
	}
	if recordType == "" {
		row["record_type"] = "日常收入"
	}
	return nil
}

func (r *ModuleRepository) applyManualDailyFinance(ctx context.Context, tenantID uint64, operatorID uint64, row map[string]any) error {
	amount := numberValue(row["amount"])
	if amount <= 0 {
		return nil
	}
	recordType := stringValue(row["record_type"])
	delta := amount
	if strings.Contains(recordType, "支出") {
		delta = -amount
	}
	account, err := r.ensureFinanceAccount(ctx, tenantID, stringValue(row["account_name"]))
	if err != nil {
		return err
	}
	return r.updateFinanceAccountBalance(ctx, mapID(account), delta)
}
func (r *ModuleRepository) findCustomer(ctx context.Context, tenantID uint64, customerID uint64, customerName string) map[string]any {
	customer := map[string]any{}
	db := r.db.WithContext(ctx).Table("customers").Where("tenant_id = ? AND deleted_at IS NULL", tenantID)
	var err error
	if customerID > 0 {
		err = db.Where("id = ?", customerID).Take(&customer).Error
	} else if customerName != "" {
		err = db.Where("name = ?", customerName).Take(&customer).Error
	}
	if err != nil {
		return map[string]any{"name": customerName, "payment_method": "immediate"}
	}
	return customer
}

func isImmediatePayment(method string) bool {
	method = strings.ToLower(strings.TrimSpace(method))
	return method == "" || method == "immediate" || method == "cash" || method == "现结" || method == "现金"
}

func (r *ModuleRepository) checkCreditLimit(ctx context.Context, tenantID uint64, customer map[string]any, amount float64) error {
	customerID := uintValue(customer["id"])
	limit := numberValue(customer["credit_limit"])
	if customerID == 0 || limit <= 0 {
		return nil
	}
	var current float64
	if err := r.db.WithContext(ctx).Table("receivables").
		Select("COALESCE(SUM(balance_amount), 0)").
		Where("tenant_id = ? AND customer_id = ? AND status <> ? AND deleted_at IS NULL", tenantID, customerID, "paid").
		Scan(&current).Error; err != nil {
		return err
	}
	if current+amount > limit {
		return apperrors.New(40012, fmt.Sprintf("客户信用额度不足，信用额度 %.2f，当前欠款 %.2f，本次金额 %.2f", limit, current, amount), 400)
	}
	return nil
}

func (r *ModuleRepository) createReceivable(ctx context.Context, tenantID uint64, operatorID uint64, customer map[string]any, orderID uint64, row map[string]any, total float64, received float64, balance float64) error {
	now := time.Now()
	invoiceDate := timeValue(row["order_date"], now)
	if empty(row["order_date"]) {
		invoiceDate = timeValue(row["registered_at"], invoiceDate)
	}
	customerID := uintValue(customer["id"])
	dueDate := calculateDueDate(invoiceDate, customer)
	status := receivableStatus(balance, dueDate)
	if received > 0 && balance > 0 {
		status = "partial"
	}
	sourceType := firstNotBlank(stringValue(row["source_type"]), "sales")
	sourceNo := firstNotBlank(stringValue(row["order_no"]), stringValue(row["project_no"]))
	receivable := map[string]any{
		"tenant_id":       tenantID,
		"receivable_no":   nextNo("receivables"),
		"customer_id":     nilIfZero(customerID),
		"customer_name":   firstNotBlank(stringValue(customer["name"]), stringValue(row["customer_name"])),
		"source_type":     sourceType,
		"source_id":       orderID,
		"source_no":       sourceNo,
		"total_amount":    total,
		"received_amount": received,
		"balance_amount":  balance,
		"invoice_date":    invoiceDate,
		"due_date":        dueDate,
		"settlement_mode": firstNotBlank(stringValue(customer["payment_method"]), "credit"),
		"status":          status,
		"created_by":      operatorID,
		"updated_by":      operatorID,
		"created_at":      now,
		"updated_at":      now,
	}
	return r.db.WithContext(ctx).Table("receivables").Create(receivable).Error
}

func calculateDueDate(invoiceDate time.Time, customer map[string]any) time.Time {
	days := int(numberValue(customer["credit_days"]))
	cycle := strings.ToLower(strings.TrimSpace(stringValue(customer["billing_cycle"])))
	if days <= 0 {
		switch cycle {
		case "monthly", "month", "月结":
			days = 30
		case "quarterly", "quarter", "季结":
			days = 90
		case "half_year", "half-year", "半年结":
			days = 180
		case "project_acceptance", "项目验收结算":
			days = 30
		default:
			days = 0
		}
	}
	due := invoiceDate.AddDate(0, 0, days)
	paymentDay := int(numberValue(customer["payment_day"]))
	if paymentDay > 0 && paymentDay <= 28 {
		due = time.Date(due.Year(), due.Month(), paymentDay, 23, 59, 59, 0, due.Location())
		if due.Before(invoiceDate) {
			due = due.AddDate(0, 1, 0)
		}
	}
	return due
}

func receivableStatus(balance float64, dueDate time.Time) string {
	if balance <= 0 {
		return "paid"
	}
	if time.Now().After(dueDate) {
		return "overdue"
	}
	return "unpaid"
}

func (r *ModuleRepository) createFinanceReceipt(ctx context.Context, tenantID uint64, operatorID uint64, targetName string, amount float64, businessNo string, businessType string, accountName ...string) error {
	return r.createFinanceFlow(ctx, tenantID, operatorID, "收款", targetName, amount, businessNo, businessType, "income", accountName...)
}

func (r *ModuleRepository) createFinancePayment(ctx context.Context, tenantID uint64, operatorID uint64, targetName string, amount float64, businessNo string, businessType string, accountName ...string) error {
	return r.createFinanceFlow(ctx, tenantID, operatorID, "付款", targetName, amount, businessNo, businessType, "expense", accountName...)
}

func (r *ModuleRepository) createFinanceFlow(ctx context.Context, tenantID uint64, operatorID uint64, recordType string, targetName string, amount float64, businessNo string, businessType string, sourceType string, accountName ...string) error {
	if amount <= 0 {
		return nil
	}
	now := time.Now()
	account, err := r.ensureFinanceAccount(ctx, tenantID, firstNotBlank(append(accountName, "现金")...))
	if err != nil {
		return err
	}
	record := map[string]any{
		"tenant_id":     tenantID,
		"record_no":     nextNo("finance"),
		"record_type":   recordType,
		"account_id":    nilIfZero(mapID(account)),
		"account_name":  stringValue(account["name"]),
		"target_name":   targetName,
		"amount":        amount,
		"status":        "已确认",
		"business_type": businessType,
		"source_type":   sourceType,
		"business_no":   businessNo,
		"occurred_at":   now,
		"created_by":    operatorID,
		"updated_by":    operatorID,
		"created_at":    now,
		"updated_at":    now,
	}
	if err := r.db.WithContext(ctx).Table("finance_records").Create(record).Error; err != nil {
		return err
	}
	delta := amount
	if recordType == "付款" || sourceType == "expense" {
		delta = -amount
	}
	return r.updateFinanceAccountBalance(ctx, mapID(account), delta)
}

func (r *ModuleRepository) ensureFinanceAccount(ctx context.Context, tenantID uint64, name string) (map[string]any, error) {
	name = firstNotBlank(name, "现金")
	account := map[string]any{}
	if err := r.db.WithContext(ctx).Table("finance_accounts").Where("tenant_id = ? AND name = ? AND deleted_at IS NULL", tenantID, name).Take(&account).Error; err == nil {
		return account, nil
	} else if err != gorm.ErrRecordNotFound {
		return nil, err
	}
	now := time.Now()
	account = map[string]any{
		"tenant_id":       tenantID,
		"code":            financeAccountCode(name),
		"name":            name,
		"account_type":    financeAccountType(name),
		"opening_balance": 0,
		"balance":         0,
		"status":          "active",
		"created_at":      now,
		"updated_at":      now,
	}
	if err := r.db.WithContext(ctx).Table("finance_accounts").Create(account).Error; err != nil {
		return nil, err
	}
	return account, nil
}

func (r *ModuleRepository) updateFinanceAccountBalance(ctx context.Context, accountID uint64, delta float64) error {
	if accountID == 0 {
		return nil
	}
	return r.db.WithContext(ctx).Table("finance_accounts").
		Where("id = ?", accountID).
		Updates(map[string]any{
			"balance":    gorm.Expr("balance + ?", delta),
			"updated_at": time.Now(),
		}).Error
}

func (r *ModuleRepository) updateSalesCollection(ctx context.Context, tenantID uint64, orderID uint64, received float64, balance float64) error {
	return r.db.WithContext(ctx).Table("sales_orders").
		Where("tenant_id = ? AND id = ?", tenantID, orderID).
		Updates(map[string]any{"received_amount": received, "receivable_amount": balance, "updated_at": time.Now()}).Error
}

func (r *ModuleRepository) receiveReceivable(ctx context.Context, tenantID uint64, operatorID uint64, data map[string]any) (map[string]any, error) {
	amount := numberValue(data["amount"])
	if amount <= 0 {
		return nil, apperrors.New(40013, "收款金额必须大于0", 400)
	}
	customerID := uintValue(data["customer_id"])
	receivableID := uintValue(data["receivable_id"])
	if receivableID == 0 {
		receivableID = uintValue(data["id"])
	}
	accountName := firstNotBlank(stringValue(data["account_name"]), "现金")
	receivedTotal := amount
	rows := []map[string]any{}
	db := r.db.WithContext(ctx).Table("receivables").Where("tenant_id = ? AND status <> ? AND deleted_at IS NULL", tenantID, "paid")
	if receivableID > 0 {
		db = db.Where("id = ?", receivableID)
	} else if customerID > 0 {
		db = db.Where("customer_id = ?", customerID).Order("due_date ASC, id ASC")
	} else {
		return nil, apperrors.New(40014, "请选择应收账款或客户", 400)
	}
	if err := db.Find(&rows).Error; err != nil {
		return nil, err
	}
	now := time.Now()
	applied := 0.0
	for _, row := range rows {
		if amount <= 0 {
			break
		}
		balance := numberValue(row["balance_amount"])
		apply := amount
		if apply > balance {
			apply = balance
		}
		newReceived := numberValue(row["received_amount"]) + apply
		newBalance := balance - apply
		status := receivableStatus(newBalance, timeValue(row["due_date"], now))
		if newBalance > 0 && newReceived > 0 {
			status = "partial"
		}
		if err := r.db.WithContext(ctx).Table("receivables").Where("id = ?", mapID(row)).Updates(map[string]any{
			"received_amount": newReceived,
			"balance_amount":  newBalance,
			"status":          status,
			"updated_by":      operatorID,
			"updated_at":      now,
		}).Error; err != nil {
			return nil, err
		}
		if stringValue(row["source_type"]) == "sales" {
			_ = r.updateSalesCollection(ctx, tenantID, uintValue(row["source_id"]), newReceived, newBalance)
		}
		customerID = uintValue(row["customer_id"])
		amount -= apply
		applied += apply
	}
	if applied > 0 {
		if err := r.createFinanceReceipt(ctx, tenantID, operatorID, stringValue(data["customer_name"]), applied, firstNotBlank(stringValue(data["business_no"]), "AR"), firstNotBlank(stringValue(data["business_type"]), "应收收款"), accountName); err != nil {
			return nil, err
		}
		_ = r.refreshCustomerReceivable(ctx, tenantID, customerID)
	}
	return map[string]any{"receivedAmount": receivedTotal, "appliedAmount": applied, "remainingAmount": amount}, nil
}

func (r *ModuleRepository) payPayable(ctx context.Context, tenantID uint64, operatorID uint64, data map[string]any) (map[string]any, error) {
	amount := numberValue(data["amount"])
	if amount <= 0 {
		return nil, apperrors.New(40035, "付款金额必须大于0", 400)
	}
	supplierID := uintValue(data["supplier_id"])
	payableID := uintValue(data["payable_id"])
	if payableID == 0 {
		payableID = uintValue(data["id"])
	}
	accountName := firstNotBlank(stringValue(data["account_name"]), "现金")
	paidTotal := amount
	rows := []map[string]any{}
	db := r.db.WithContext(ctx).Table("payables").Where("tenant_id = ? AND status <> ? AND deleted_at IS NULL", tenantID, "paid")
	if payableID > 0 {
		db = db.Where("id = ?", payableID)
	} else if supplierID > 0 {
		db = db.Where("supplier_id = ?", supplierID).Order("due_date ASC, id ASC")
	} else {
		return nil, apperrors.New(40036, "请选择应付账款或供应商", 400)
	}
	if err := db.Find(&rows).Error; err != nil {
		return nil, err
	}
	now := time.Now()
	applied := 0.0
	var lastSupplierID uint64
	for _, row := range rows {
		if amount <= 0 {
			break
		}
		balance := numberValue(row["balance_amount"])
		apply := amount
		if apply > balance {
			apply = balance
		}
		newPaid := numberValue(row["paid_amount"]) + apply
		newBalance := balance - apply
		if err := r.db.WithContext(ctx).Table("payables").Where("id = ?", mapID(row)).Updates(map[string]any{
			"paid_amount":    newPaid,
			"balance_amount": newBalance,
			"status":         payableStatus(newBalance, newPaid),
			"updated_by":     operatorID,
			"updated_at":     now,
		}).Error; err != nil {
			return nil, err
		}
		if stringValue(row["source_type"]) == "purchase" {
			_ = r.db.WithContext(ctx).Table("purchase_orders").Where("tenant_id = ? AND id = ?", tenantID, uintValue(row["source_id"])).Updates(map[string]any{
				"paid_amount":    newPaid,
				"payable_amount": newBalance,
				"updated_at":     now,
			}).Error
		}
		lastSupplierID = uintValue(row["supplier_id"])
		amount -= apply
		applied += apply
	}
	if applied > 0 {
		if err := r.createFinancePayment(ctx, tenantID, operatorID, stringValue(data["supplier_name"]), applied, firstNotBlank(stringValue(data["business_no"]), "AP"), firstNotBlank(stringValue(data["business_type"]), "应付付款"), accountName); err != nil {
			return nil, err
		}
		_ = r.refreshSupplierPayable(ctx, tenantID, lastSupplierID)
	}
	return map[string]any{"paidAmount": paidTotal, "appliedAmount": applied, "remainingAmount": amount}, nil
}

func (r *ModuleRepository) refreshCustomerReceivable(ctx context.Context, tenantID uint64, customerID uint64) error {
	if customerID == 0 {
		return nil
	}
	var balance float64
	if err := r.db.WithContext(ctx).Table("receivables").
		Select("COALESCE(SUM(balance_amount), 0)").
		Where("tenant_id = ? AND customer_id = ? AND status <> ? AND deleted_at IS NULL", tenantID, customerID, "paid").
		Scan(&balance).Error; err != nil {
		return err
	}
	return r.db.WithContext(ctx).Table("customers").Where("tenant_id = ? AND id = ?", tenantID, customerID).Update("receivable_balance", balance).Error
}

func (r *ModuleRepository) Delete(ctx context.Context, module string, tenantID uint64, id uint64, operatorID uint64, reason string) error {
	spec, err := specOf(module)
	if err != nil {
		return err
	}
	if module == "inventory" {
		return apperrors.New(40020, "库存数据由采购、销售、删除反处理等业务自动维护，不能直接删除库存记录", 400)
	}
	if module == "sales" {
		return r.deleteSalesOrder(ctx, tenantID, id, operatorID, reason)
	}
	if module == "purchase" {
		return r.deletePurchaseOrder(ctx, tenantID, id, operatorID, reason)
	}
	if module == "repair" {
		return r.deleteRepairOrder(ctx, tenantID, id, operatorID, reason)
	}
	return r.db.WithContext(ctx).Table(spec.Table).Where("tenant_id = ? AND id = ?", tenantID, id).Update("deleted_at", time.Now()).Error
}

func (r *ModuleRepository) deleteSalesOrder(ctx context.Context, tenantID uint64, id uint64, operatorID uint64, reason string) error {
	reason = strings.TrimSpace(reason)
	if reason == "" {
		return apperrors.New(40033, "请输入删除原因", 400)
	}
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		repo := &ModuleRepository{db: tx}
		order := map[string]any{}
		if err := tx.Table("sales_orders").Where("tenant_id = ? AND id = ?", tenantID, id).Take(&order).Error; err != nil {
			return err
		}
		if !empty(order["deleted_at"]) {
			return apperrors.New(40034, "订单已经删除", 409)
		}
		if salesOrderIsOutbound(stringValue(order["status"])) {
			return apperrors.New(40035, "订单已经出库，无法删除", 400)
		}

		items := make([]map[string]any, 0)
		if err := tx.Table("sales_order_items").
			Where("tenant_id = ? AND sales_order_id = ? AND deleted_at IS NULL", tenantID, id).
			Order("id ASC").Find(&items).Error; err != nil {
			return err
		}
		allocations := make([]map[string]any, 0)
		if err := tx.Table("sales_cost_allocations").
			Where("tenant_id = ? AND sales_order_id = ? AND deleted_at IS NULL", tenantID, id).
			Order("id ASC").Find(&allocations).Error; err != nil {
			return err
		}

		snapshot := make(map[string]any, len(order)+2)
		for key, value := range order {
			snapshot[key] = value
		}
		snapshot["items"] = items
		snapshot["cost_allocations"] = allocations
		itemsJSON, err := json.Marshal(items)
		if err != nil {
			return err
		}
		snapshotJSON, err := json.Marshal(snapshot)
		if err != nil {
			return err
		}
		now := time.Now()
		deletion := map[string]any{
			"tenant_id":         tenantID,
			"original_order_id": id,
			"order_no":          stringValue(order["order_no"]),
			"customer_id":       nilIfZero(uintValue(order["customer_id"])),
			"customer_name":     stringValue(order["customer_name"]),
			"items_json":        string(itemsJSON),
			"deleted_by":        operatorID,
			"deleted_at":        now,
			"delete_reason":     reason,
			"snapshot_json":     string(snapshotJSON),
			"created_at":        now,
		}
		if err := tx.Table("sales_order_deletions").Create(deletion).Error; err != nil {
			var count int64
			_ = tx.Table("sales_order_deletions").Where("tenant_id = ? AND original_order_id = ?", tenantID, id).Count(&count).Error
			if count > 0 {
				return apperrors.New(40034, "订单已经删除", 409)
			}
			return err
		}
		documentRecordID, err := repo.createDocumentDeleteRecord(ctx, tenantID, "SALES_ORDER", id, stringValue(order["order_no"]), reason, operatorID, snapshotJSON, now)
		if err != nil {
			return err
		}
		if err := repo.restoreSalesOrderInventory(ctx, tenantID, operatorID, order, allocations); err != nil {
			return err
		}
		if err := repo.createSalesDeleteDetails(ctx, tenantID, documentRecordID, order, allocations); err != nil {
			return err
		}
		stockProcessed := len(allocations) > 0
		financeProcessed, err := repo.reverseSalesOrderFinance(ctx, tenantID, operatorID, documentRecordID, order, now)
		if err != nil {
			return err
		}
		if err := repo.refreshCustomerReceivable(ctx, tenantID, uintValue(order["customer_id"])); err != nil {
			return err
		}
		result := tx.Table("sales_orders").Where("tenant_id = ? AND id = ? AND deleted_at IS NULL", tenantID, id).Updates(map[string]any{
			"status":            "已删除",
			"received_amount":   0,
			"receivable_amount": 0,
			"deleted_by":        operatorID,
			"delete_reason":     reason,
			"deleted_at":        now,
			"updated_by":        operatorID,
			"updated_at":        now,
		})
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected != 1 {
			return apperrors.New(40034, "订单已经删除", 409)
		}
		if err := repo.completeDocumentDeleteRecord(ctx, tenantID, documentRecordID, stockProcessed, financeProcessed, now); err != nil {
			return err
		}
		return repo.writeSalesOrderDeleteLog(ctx, tenantID, operatorID, id, reason, now)
	})
}

func (r *ModuleRepository) deletePurchaseOrder(ctx context.Context, tenantID uint64, id uint64, operatorID uint64, reason string) error {
	reason = strings.TrimSpace(reason)
	if reason == "" {
		return apperrors.New(40033, "请输入删除原因", 400)
	}
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		repo := &ModuleRepository{db: tx}
		order := map[string]any{}
		if err := tx.Table("purchase_orders").Where("tenant_id = ? AND id = ?", tenantID, id).Take(&order).Error; err != nil {
			return err
		}
		if !empty(order["deleted_at"]) {
			return apperrors.New(40034, "订单已经删除", 409)
		}
		items := make([]map[string]any, 0)
		if err := tx.Table("purchase_order_items").
			Where("tenant_id = ? AND purchase_order_id = ? AND deleted_at IS NULL", tenantID, id).
			Order("id ASC").Find(&items).Error; err != nil {
			return err
		}
		batches := make([]map[string]any, 0)
		if err := tx.Table("inventory_batches").
			Where("tenant_id = ? AND purchase_order_id = ? AND deleted_at IS NULL", tenantID, id).
			Order("id ASC").Find(&batches).Error; err != nil {
			return err
		}
		for _, batch := range batches {
			if numberValue(batch["remaining_quantity"])+0.000001 < numberValue(batch["inbound_quantity"]) {
				return apperrors.New(40038, "采购订单库存已被销售或领用，无法删除", 409)
			}
		}
		snapshot := make(map[string]any, len(order)+2)
		for key, value := range order {
			snapshot[key] = value
		}
		snapshot["items"] = items
		snapshot["inventory_batches"] = batches
		snapshotJSON, err := json.Marshal(snapshot)
		if err != nil {
			return err
		}
		now := time.Now()
		documentRecordID, err := repo.createDocumentDeleteRecord(ctx, tenantID, "PURCHASE_ORDER", id, stringValue(order["order_no"]), reason, operatorID, snapshotJSON, now)
		if err != nil {
			return err
		}
		stockProcessed := false
		for _, item := range items {
			qty := numberValue(item["quantity"])
			if qty <= 0 {
				continue
			}
			movementRow := map[string]any{
				"product_id":   item["product_id"],
				"product_code": item["product_code"],
				"product_name": item["product_name"],
				"warehouse":    item["warehouse"],
				"quantity":     qty,
				"price":        item["price"],
				"order_no":     order["order_no"],
			}
			if err := repo.applyInventoryMovement(ctx, tenantID, operatorID, "PURCHASE_DELETE", id, movementRow, "out"); err != nil {
				return err
			}
			stockProcessed = true
			if err := repo.createDocumentDeleteDetail(ctx, tenantID, documentRecordID, map[string]any{
				"detail_type":   "STOCK",
				"sku_id":        nilIfZero(uintValue(item["product_id"])),
				"sku_code":      item["product_code"],
				"sku_name":      item["product_name"],
				"warehouse":     item["warehouse"],
				"quantity":      qty,
				"stock_change":  -qty,
				"business_type": "PURCHASE_DELETE",
				"business_id":   nilIfZero(id),
				"business_no":   order["order_no"],
			}); err != nil {
				return err
			}
		}
		for _, batch := range batches {
			if err := tx.Table("inventory_batches").Where("tenant_id = ? AND id = ?", tenantID, mapID(batch)).Updates(map[string]any{
				"remaining_quantity": 0,
				"status":             batchStatus(numberValue(batch["inbound_quantity"]), 0),
				"deleted_at":         now,
				"updated_by":         operatorID,
				"updated_at":         now,
			}).Error; err != nil {
				return err
			}
		}
		financeProcessed, err := repo.reverseDocumentFinance(ctx, tenantID, operatorID, documentRecordID, "PURCHASE_ORDER", stringValue(order["order_no"]), stringValue(order["supplier_name"]), now)
		if err != nil {
			return err
		}
		payableResult := tx.Table("payables").Where("tenant_id = ? AND source_type = ? AND source_id = ? AND deleted_at IS NULL", tenantID, "purchase", id).Updates(map[string]any{
			"status":         "已冲销",
			"balance_amount": 0,
			"updated_by":     operatorID,
			"updated_at":     now,
		})
		if payableResult.Error != nil {
			return payableResult.Error
		}
		if payableResult.RowsAffected > 0 {
			financeProcessed = true
		}
		if err := repo.refreshSupplierPayable(ctx, tenantID, uintValue(order["supplier_id"])); err != nil {
			return err
		}
		result := tx.Table("purchase_orders").Where("tenant_id = ? AND id = ? AND deleted_at IS NULL", tenantID, id).Updates(map[string]any{
			"status":        "已删除",
			"deleted_by":    operatorID,
			"delete_reason": reason,
			"deleted_at":    now,
			"updated_by":    operatorID,
			"updated_at":    now,
		})
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected != 1 {
			return apperrors.New(40034, "订单已经删除", 409)
		}
		if err := repo.completeDocumentDeleteRecord(ctx, tenantID, documentRecordID, stockProcessed, financeProcessed, now); err != nil {
			return err
		}
		return repo.writeDocumentDeleteLog(ctx, tenantID, operatorID, "purchase", id, reason, now)
	})
}

func (r *ModuleRepository) deleteRepairOrder(ctx context.Context, tenantID uint64, id uint64, operatorID uint64, reason string) error {
	reason = strings.TrimSpace(reason)
	if reason == "" {
		return apperrors.New(40033, "请输入删除原因", 400)
	}
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		repo := &ModuleRepository{db: tx}
		order := map[string]any{}
		if err := tx.Table("repair_orders").Where("tenant_id = ? AND id = ?", tenantID, id).Take(&order).Error; err != nil {
			return err
		}
		if !empty(order["deleted_at"]) {
			return apperrors.New(40034, "维修单已经删除", 409)
		}
		items := make([]map[string]any, 0)
		if err := tx.Table("repair_order_items").Where("tenant_id = ? AND repair_order_id = ? AND deleted_at IS NULL", tenantID, id).Order("id ASC").Find(&items).Error; err != nil {
			return err
		}
		snapshot := make(map[string]any, len(order)+1)
		for key, value := range order {
			snapshot[key] = value
		}
		snapshot["items"] = items
		snapshotJSON, err := json.Marshal(snapshot)
		if err != nil {
			return err
		}
		now := time.Now()
		recordID, err := repo.createDocumentDeleteRecord(ctx, tenantID, "REPAIR_ORDER", id, stringValue(order["order_no"]), reason, operatorID, snapshotJSON, now)
		if err != nil {
			return err
		}
		financeProcessed, err := repo.reverseRepairFinance(ctx, tenantID, operatorID, recordID, id, order, now)
		if err != nil {
			return err
		}
		if err := tx.Table("repair_order_items").Where("tenant_id = ? AND repair_order_id = ? AND deleted_at IS NULL", tenantID, id).Updates(map[string]any{
			"deleted_at": now,
			"updated_by": operatorID,
			"updated_at": now,
		}).Error; err != nil {
			return err
		}
		result := tx.Table("repair_orders").Where("tenant_id = ? AND id = ? AND deleted_at IS NULL", tenantID, id).Updates(map[string]any{
			"repair_status": "已删除",
			"deleted_by":    operatorID,
			"delete_reason": reason,
			"deleted_at":    now,
			"updated_by":    operatorID,
			"updated_at":    now,
		})
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected != 1 {
			return apperrors.New(40034, "维修单已经删除", 409)
		}
		return repo.completeDocumentDeleteRecord(ctx, tenantID, recordID, false, financeProcessed, now)
	})
}

func (r *ModuleRepository) reverseRepairFinance(ctx context.Context, tenantID uint64, operatorID uint64, recordID uint64, orderID uint64, order map[string]any, now time.Time) (bool, error) {
	processed := false
	customerID := uintValue(order["customer_id"])
	customerName := stringValue(order["customer_name"])
	receivables := make([]map[string]any, 0)
	if err := r.db.WithContext(ctx).Table("receivables").Where("tenant_id = ? AND source_type = ? AND source_id = ? AND deleted_at IS NULL", tenantID, "repair", orderID).Find(&receivables).Error; err != nil {
		return false, err
	}
	for _, receivable := range receivables {
		received := numberValue(receivable["received_amount"])
		if received > 0 {
			if err := r.reverseFinanceReceiptsByBusinessNo(ctx, tenantID, operatorID, recordID, stringValue(receivable["receivable_no"]), firstNotBlank(stringValue(receivable["customer_name"]), customerName), received, "维修删除冲销"); err != nil {
				return false, err
			}
		}
		result := r.db.WithContext(ctx).Table("receivables").Where("tenant_id = ? AND id = ?", tenantID, mapID(receivable)).Updates(map[string]any{
			"status":         "已冲销",
			"balance_amount": 0,
			"updated_by":     operatorID,
			"updated_at":     now,
		})
		if result.Error != nil {
			return false, result.Error
		}
		processed = processed || result.RowsAffected > 0
	}
	payables := make([]map[string]any, 0)
	if err := r.db.WithContext(ctx).Table("payables").Where("tenant_id = ? AND source_type = ? AND source_id = ? AND deleted_at IS NULL", tenantID, "repair_outsource", orderID).Find(&payables).Error; err != nil {
		return false, err
	}
	supplierIDs := map[uint64]struct{}{}
	for _, payable := range payables {
		paid := numberValue(payable["paid_amount"])
		if paid > 0 {
			if err := r.reverseFinancePaymentsByBusinessNo(ctx, tenantID, operatorID, recordID, stringValue(payable["payable_no"]), stringValue(payable["supplier_name"]), paid, "维修删除冲销"); err != nil {
				return false, err
			}
		}
		result := r.db.WithContext(ctx).Table("payables").Where("tenant_id = ? AND id = ?", tenantID, mapID(payable)).Updates(map[string]any{
			"status":         "已冲销",
			"balance_amount": 0,
			"updated_by":     operatorID,
			"updated_at":     now,
		})
		if result.Error != nil {
			return false, result.Error
		}
		processed = processed || result.RowsAffected > 0
		if supplierID := uintValue(payable["supplier_id"]); supplierID > 0 {
			supplierIDs[supplierID] = struct{}{}
		}
	}
	if err := r.refreshCustomerReceivable(ctx, tenantID, customerID); err != nil {
		return false, err
	}
	for supplierID := range supplierIDs {
		if err := r.refreshSupplierPayable(ctx, tenantID, supplierID); err != nil {
			return false, err
		}
	}
	return processed, nil
}

func (r *ModuleRepository) createDocumentDeleteRecord(ctx context.Context, tenantID uint64, documentType string, documentID uint64, documentNo string, reason string, operatorID uint64, beforeData []byte, now time.Time) (uint64, error) {
	username := ""
	if operatorID > 0 {
		user := map[string]any{}
		err := r.db.WithContext(ctx).Table("users").Select("real_name, username").Where("tenant_id = ? AND id = ?", tenantID, operatorID).Take(&user).Error
		if err != nil && err != gorm.ErrRecordNotFound {
			return 0, err
		}
		username = firstNotBlank(stringValue(user["real_name"]), stringValue(user["username"]))
	}
	row := map[string]any{
		"tenant_id":         tenantID,
		"document_type":     documentType,
		"document_id":       documentID,
		"document_no":       documentNo,
		"delete_reason":     reason,
		"delete_user_id":    operatorID,
		"delete_user_name":  username,
		"delete_time":       now,
		"delete_status":     "PROCESSING",
		"before_data":       string(beforeData),
		"stock_processed":   false,
		"finance_processed": false,
		"created_by":        operatorID,
		"updated_by":        operatorID,
		"created_at":        now,
		"updated_at":        now,
	}
	if err := r.db.WithContext(ctx).Table("document_delete_records").Create(row).Error; err != nil {
		var count int64
		_ = r.db.WithContext(ctx).Table("document_delete_records").
			Where("tenant_id = ? AND document_type = ? AND document_id = ?", tenantID, documentType, documentID).
			Count(&count).Error
		if count > 0 {
			return 0, apperrors.New(40034, "订单已经删除", 409)
		}
		return 0, err
	}
	if id := mapID(row); id > 0 {
		return id, nil
	}
	created := map[string]any{}
	if err := r.db.WithContext(ctx).Table("document_delete_records").
		Select("id").
		Where("tenant_id = ? AND document_type = ? AND document_id = ?", tenantID, documentType, documentID).
		Take(&created).Error; err != nil {
		return 0, err
	}
	return mapID(created), nil
}

func (r *ModuleRepository) completeDocumentDeleteRecord(ctx context.Context, tenantID uint64, recordID uint64, stockProcessed bool, financeProcessed bool, now time.Time) error {
	return r.db.WithContext(ctx).Table("document_delete_records").Where("tenant_id = ? AND id = ?", tenantID, recordID).Updates(map[string]any{
		"delete_status":     "SUCCESS",
		"stock_processed":   stockProcessed,
		"finance_processed": financeProcessed,
		"completed_at":      now,
		"updated_at":        now,
	}).Error
}

func (r *ModuleRepository) createDocumentDeleteDetail(ctx context.Context, tenantID uint64, recordID uint64, row map[string]any) error {
	now := time.Now()
	row["tenant_id"] = tenantID
	row["record_id"] = recordID
	row["created_at"] = now
	row["updated_at"] = now
	return r.db.WithContext(ctx).Table("document_delete_details").Create(row).Error
}

func (r *ModuleRepository) createSalesDeleteDetails(ctx context.Context, tenantID uint64, recordID uint64, order map[string]any, allocations []map[string]any) error {
	for _, allocation := range allocations {
		qty := numberValue(allocation["quantity"])
		if err := r.createDocumentDeleteDetail(ctx, tenantID, recordID, map[string]any{
			"detail_type":   "STOCK",
			"sku_id":        nilIfZero(uintValue(allocation["product_id"])),
			"sku_code":      allocation["product_code"],
			"sku_name":      allocation["product_name"],
			"quantity":      qty,
			"stock_change":  qty,
			"business_type": "ORDER_DELETE",
			"business_id":   nilIfZero(mapID(order)),
			"business_no":   order["order_no"],
		}); err != nil {
			return err
		}
	}
	return nil
}

func (r *ModuleRepository) reverseSalesOrderFinance(ctx context.Context, tenantID uint64, operatorID uint64, recordID uint64, order map[string]any, now time.Time) (bool, error) {
	orderID := mapID(order)
	orderNo := stringValue(order["order_no"])
	customerName := stringValue(order["customer_name"])
	processed := false

	receivables := make([]map[string]any, 0)
	if err := r.db.WithContext(ctx).Table("receivables").
		Where("tenant_id = ? AND source_type = ? AND deleted_at IS NULL", tenantID, "sales").
		Where("(source_id = ? OR source_no = ?)", orderID, orderNo).
		Order("id ASC").Find(&receivables).Error; err != nil {
		return false, err
	}
	totalReceived := 0.0
	for _, receivable := range receivables {
		totalReceived += numberValue(receivable["received_amount"])
		if err := r.db.WithContext(ctx).Table("receivables").Where("id = ?", mapID(receivable)).Updates(map[string]any{
			"status":         "已冲销",
			"balance_amount": 0,
			"updated_by":     operatorID,
			"updated_at":     now,
		}).Error; err != nil {
			return false, err
		}
		processed = true
		if err := r.createDocumentDeleteDetail(ctx, tenantID, recordID, map[string]any{
			"detail_type":   "FINANCE",
			"finance_type":  "RECEIVABLE_REVERSAL",
			"amount":        numberValue(receivable["balance_amount"]),
			"business_type": "SALES_ORDER",
			"business_id":   nilIfZero(orderID),
			"business_no":   orderNo,
		}); err != nil {
			return false, err
		}
	}

	statementReceived, err := r.reverseSalesOrderStatementItems(ctx, tenantID, operatorID, recordID, orderID, orderNo, customerName, now)
	if err != nil {
		return false, err
	}
	if statementReceived > 0 {
		processed = true
	}

	directReversed, err := r.reverseSalesOrderDirectFinance(ctx, tenantID, operatorID, recordID, orderNo, customerName)
	if err != nil {
		return false, err
	}
	if directReversed > 0 {
		processed = true
	}

	unmatchedReceived := totalReceived - statementReceived - directReversed
	if unmatchedReceived > 0.000001 {
		if err := r.createFinancePayment(ctx, tenantID, operatorID, customerName, unmatchedReceived, orderNo, "销售删除冲销", "现金"); err != nil {
			return false, err
		}
		processed = true
		if err := r.createDocumentDeleteDetail(ctx, tenantID, recordID, map[string]any{
			"detail_type":   "FINANCE",
			"finance_type":  "RECEIPT_REVERSAL",
			"amount":        unmatchedReceived,
			"business_type": "SALES_ORDER",
			"business_id":   nilIfZero(orderID),
			"business_no":   orderNo,
		}); err != nil {
			return false, err
		}
	}

	return processed, nil
}

func (r *ModuleRepository) reverseSalesOrderStatementItems(ctx context.Context, tenantID uint64, operatorID uint64, recordID uint64, orderID uint64, orderNo string, customerName string, now time.Time) (float64, error) {
	items := make([]map[string]any, 0)
	if err := r.db.WithContext(ctx).Table("customer_statement_items").
		Where("tenant_id = ? AND deleted_at IS NULL", tenantID).
		Where("(sale_id = ? OR sale_no = ? OR source_no = ?)", orderID, orderNo, orderNo).
		Order("id ASC").Find(&items).Error; err != nil {
		return 0, err
	}
	statementIDs := map[uint64]struct{}{}
	totalReversed := 0.0
	for _, item := range items {
		itemReceived := numberValue(item["received_amount"])
		if itemReceived > 0 {
			statement := map[string]any{}
			if err := r.db.WithContext(ctx).Table("customer_statements").
				Where("tenant_id = ? AND id = ? AND deleted_at IS NULL", tenantID, uintValue(item["statement_id"])).
				Take(&statement).Error; err != nil {
				return 0, err
			}
			if err := r.reverseFinanceReceiptsByBusinessNo(ctx, tenantID, operatorID, recordID, stringValue(statement["statement_no"]), firstNotBlank(customerName, stringValue(statement["customer_name"])), itemReceived, "客户对账单删除冲销"); err != nil {
				return 0, err
			}
			totalReversed += itemReceived
		}
		if err := r.db.WithContext(ctx).Table("customer_statement_items").Where("id = ?", mapID(item)).Updates(map[string]any{
			"total_amount":      0,
			"received_amount":   0,
			"unpaid_amount":     0,
			"settlement_status": "已冲销",
			"updated_by":        operatorID,
			"updated_at":        now,
		}).Error; err != nil {
			return 0, err
		}
		statementIDs[uintValue(item["statement_id"])] = struct{}{}
		if err := r.createDocumentDeleteDetail(ctx, tenantID, recordID, map[string]any{
			"detail_type":   "FINANCE",
			"finance_type":  "STATEMENT_REVERSAL",
			"amount":        itemReceived,
			"business_type": "SALES_ORDER",
			"business_id":   nilIfZero(orderID),
			"business_no":   orderNo,
		}); err != nil {
			return 0, err
		}
	}
	for statementID := range statementIDs {
		statement := map[string]any{}
		if err := r.db.WithContext(ctx).Table("customer_statements").Where("tenant_id = ? AND id = ?", tenantID, statementID).Take(&statement).Error; err != nil {
			return 0, err
		}
		if err := r.recalculateCustomerStatement(ctx, tenantID, operatorID, statementID, firstNotBlank(stringValue(statement["status"]), "unconfirmed")); err != nil {
			return 0, err
		}
	}
	return totalReversed, nil
}

func (r *ModuleRepository) reverseSalesOrderDirectFinance(ctx context.Context, tenantID uint64, operatorID uint64, recordID uint64, orderNo string, customerName string) (float64, error) {
	records := make([]map[string]any, 0)
	if err := r.db.WithContext(ctx).Table("finance_records").
		Where("tenant_id = ? AND business_no = ? AND deleted_at IS NULL", tenantID, orderNo).
		Where("business_type <> ?", "销售删除冲销").
		Order("id ASC").Find(&records).Error; err != nil {
		return 0, err
	}
	total := 0.0
	for _, record := range records {
		amount := numberValue(record["amount"])
		if amount <= 0 {
			continue
		}
		if stringValue(record["source_type"]) == "expense" {
			if err := r.createFinanceReceipt(ctx, tenantID, operatorID, firstNotBlank(stringValue(record["target_name"]), customerName), amount, orderNo, "销售删除冲销", stringValue(record["account_name"])); err != nil {
				return 0, err
			}
		} else {
			if err := r.createFinancePayment(ctx, tenantID, operatorID, firstNotBlank(stringValue(record["target_name"]), customerName), amount, orderNo, "销售删除冲销", stringValue(record["account_name"])); err != nil {
				return 0, err
			}
		}
		total += amount
		if err := r.createDocumentDeleteDetail(ctx, tenantID, recordID, map[string]any{
			"detail_type":   "FINANCE",
			"finance_type":  "RECEIPT_REVERSAL",
			"finance_no":    record["record_no"],
			"amount":        amount,
			"business_type": "SALES_ORDER",
			"business_no":   orderNo,
		}); err != nil {
			return 0, err
		}
	}
	return total, nil
}

func (r *ModuleRepository) reverseFinanceReceiptsByBusinessNo(ctx context.Context, tenantID uint64, operatorID uint64, recordID uint64, businessNo string, targetName string, amount float64, reversalType string) error {
	if amount <= 0 {
		return nil
	}
	records := make([]map[string]any, 0)
	if err := r.db.WithContext(ctx).Table("finance_records").
		Where("tenant_id = ? AND business_no = ? AND source_type = ? AND deleted_at IS NULL", tenantID, businessNo, "income").
		Order("id ASC").Find(&records).Error; err != nil {
		return err
	}
	left := amount
	for _, record := range records {
		if left <= 0 {
			break
		}
		reverseAmount := numberValue(record["amount"])
		if reverseAmount > left {
			reverseAmount = left
		}
		if reverseAmount <= 0 {
			continue
		}
		if err := r.createFinancePayment(ctx, tenantID, operatorID, firstNotBlank(stringValue(record["target_name"]), targetName), reverseAmount, businessNo, reversalType, stringValue(record["account_name"])); err != nil {
			return err
		}
		left -= reverseAmount
		if err := r.createDocumentDeleteDetail(ctx, tenantID, recordID, map[string]any{
			"detail_type":   "FINANCE",
			"finance_type":  "STATEMENT_RECEIPT_REVERSAL",
			"finance_no":    record["record_no"],
			"amount":        reverseAmount,
			"business_type": "SALES_ORDER",
			"business_no":   businessNo,
		}); err != nil {
			return err
		}
	}
	if left > 0.000001 {
		if err := r.createFinancePayment(ctx, tenantID, operatorID, targetName, left, businessNo, reversalType, "现金"); err != nil {
			return err
		}
		if err := r.createDocumentDeleteDetail(ctx, tenantID, recordID, map[string]any{
			"detail_type":   "FINANCE",
			"finance_type":  "STATEMENT_RECEIPT_REVERSAL",
			"amount":        left,
			"business_type": "SALES_ORDER",
			"business_no":   businessNo,
		}); err != nil {
			return err
		}
	}
	return nil
}

func (r *ModuleRepository) reverseFinancePaymentsByBusinessNo(ctx context.Context, tenantID uint64, operatorID uint64, recordID uint64, businessNo string, targetName string, amount float64, reversalType string) error {
	if amount <= 0 {
		return nil
	}
	records := make([]map[string]any, 0)
	if err := r.db.WithContext(ctx).Table("finance_records").
		Where("tenant_id = ? AND business_no = ? AND source_type = ? AND deleted_at IS NULL", tenantID, businessNo, "expense").
		Order("id ASC").Find(&records).Error; err != nil {
		return err
	}
	left := amount
	for _, record := range records {
		if left <= 0 {
			break
		}
		reverseAmount := numberValue(record["amount"])
		if reverseAmount > left {
			reverseAmount = left
		}
		if reverseAmount <= 0 {
			continue
		}
		if err := r.createFinanceReceipt(ctx, tenantID, operatorID, firstNotBlank(stringValue(record["target_name"]), targetName), reverseAmount, businessNo, reversalType, stringValue(record["account_name"])); err != nil {
			return err
		}
		left -= reverseAmount
		if err := r.createDocumentDeleteDetail(ctx, tenantID, recordID, map[string]any{
			"detail_type":   "FINANCE",
			"finance_type":  "PAYMENT_REVERSAL",
			"finance_no":    record["record_no"],
			"amount":        reverseAmount,
			"business_type": "REPAIR_ORDER",
			"business_no":   businessNo,
		}); err != nil {
			return err
		}
	}
	if left > 0.000001 {
		if err := r.createFinanceReceipt(ctx, tenantID, operatorID, targetName, left, businessNo, reversalType, "现金"); err != nil {
			return err
		}
		if err := r.createDocumentDeleteDetail(ctx, tenantID, recordID, map[string]any{
			"detail_type":   "FINANCE",
			"finance_type":  "PAYMENT_REVERSAL",
			"amount":        left,
			"business_type": "REPAIR_ORDER",
			"business_no":   businessNo,
		}); err != nil {
			return err
		}
	}
	return nil
}

func (r *ModuleRepository) reverseDocumentFinance(ctx context.Context, tenantID uint64, operatorID uint64, recordID uint64, documentType string, documentNo string, targetName string, now time.Time) (bool, error) {
	records := make([]map[string]any, 0)
	if err := r.db.WithContext(ctx).Table("finance_records").
		Where("tenant_id = ? AND business_no = ? AND deleted_at IS NULL", tenantID, documentNo).
		Order("id ASC").Find(&records).Error; err != nil {
		return false, err
	}
	processed := false
	for _, record := range records {
		amount := numberValue(record["amount"])
		if amount <= 0 {
			continue
		}
		sourceType := stringValue(record["source_type"])
		if sourceType == "expense" {
			if err := r.createFinanceReceipt(ctx, tenantID, operatorID, firstNotBlank(stringValue(record["target_name"]), targetName), amount, documentNo, "删除冲销", stringValue(record["account_name"])); err != nil {
				return false, err
			}
		} else {
			if err := r.createFinancePayment(ctx, tenantID, operatorID, firstNotBlank(stringValue(record["target_name"]), targetName), amount, documentNo, "删除冲销", stringValue(record["account_name"])); err != nil {
				return false, err
			}
		}
		processed = true
		if err := r.createDocumentDeleteDetail(ctx, tenantID, recordID, map[string]any{
			"detail_type":   "FINANCE",
			"finance_type":  "REVERSAL",
			"finance_no":    record["record_no"],
			"amount":        amount,
			"business_type": documentType,
			"business_no":   documentNo,
			"remark":        "删除冲销",
		}); err != nil {
			return false, err
		}
	}
	if documentType == "SALES_ORDER" {
		result := r.db.WithContext(ctx).Table("receivables").Where("tenant_id = ? AND source_type = ? AND source_no = ? AND deleted_at IS NULL", tenantID, "sales", documentNo).Updates(map[string]any{
			"status":         "已冲销",
			"balance_amount": 0,
			"updated_by":     operatorID,
			"updated_at":     now,
		})
		if result.Error != nil {
			return false, result.Error
		}
		if result.RowsAffected > 0 {
			processed = true
		}
	}
	return processed, nil
}

func (r *ModuleRepository) writeDocumentDeleteLog(ctx context.Context, tenantID uint64, operatorID uint64, module string, documentID uint64, reason string, now time.Time) error {
	username := ""
	if operatorID > 0 {
		user := map[string]any{}
		err := r.db.WithContext(ctx).Table("users").Select("username").Where("tenant_id = ? AND id = ?", tenantID, operatorID).Take(&user).Error
		if err != nil && err != gorm.ErrRecordNotFound {
			return err
		}
		username = stringValue(user["username"])
	}
	requestBody, err := json.Marshal(map[string]any{"reason": reason})
	if err != nil {
		return err
	}
	return r.db.WithContext(ctx).Table("operation_logs").Create(map[string]any{
		"tenant_id":    tenantID,
		"user_id":      nilIfZero(operatorID),
		"username":     username,
		"module":       module,
		"action":       "delete",
		"method":       "DELETE",
		"path":         fmt.Sprintf("/api/v1/%s/%d", module, documentID),
		"request_body": string(requestBody),
		"status_code":  200,
		"cost_ms":      0,
		"created_at":   now,
	}).Error
}

func documentDeleteMovementTypes(documentType string) []string {
	switch documentType {
	case "SALES_ORDER":
		return []string{"ORDER_DELETE"}
	case "PURCHASE_ORDER":
		return []string{"PURCHASE_DELETE"}
	default:
		return []string{"ORDER_DELETE", "PURCHASE_DELETE", "STOCK_IN_DELETE", "STOCK_OUT_DELETE"}
	}
}

func (r *ModuleRepository) restoreSalesOrderInventory(ctx context.Context, tenantID uint64, operatorID uint64, order map[string]any, allocations []map[string]any) error {
	if len(allocations) == 0 && numberValue(order["quantity"]) > 0 {
		return apperrors.New(40036, "订单缺少库存成本分摊记录，无法安全恢复库存", 409)
	}
	for _, allocation := range allocations {
		batchID := uintValue(allocation["inventory_batch_id"])
		batch := map[string]any{}
		if batchID == 0 {
			return apperrors.New(40036, "订单库存批次信息不完整，无法安全恢复库存", 409)
		}
		if err := r.db.WithContext(ctx).Table("inventory_batches").
			Where("tenant_id = ? AND id = ? AND deleted_at IS NULL", tenantID, batchID).
			Take(&batch).Error; err != nil {
			return err
		}
		quantity := numberValue(allocation["quantity"])
		newRemaining := numberValue(batch["remaining_quantity"]) + quantity
		inboundQuantity := numberValue(batch["inbound_quantity"])
		if quantity <= 0 || newRemaining > inboundQuantity+0.000001 {
			return apperrors.New(40037, "库存恢复数量异常", 409)
		}
		if err := r.db.WithContext(ctx).Table("inventory_batches").Where("tenant_id = ? AND id = ?", tenantID, batchID).Updates(map[string]any{
			"remaining_quantity": newRemaining,
			"status":             batchStatus(inboundQuantity, newRemaining),
			"updated_by":         operatorID,
			"updated_at":         time.Now(),
		}).Error; err != nil {
			return err
		}
		movementRow := map[string]any{
			"product_id":   allocation["product_id"],
			"product_code": allocation["product_code"],
			"product_name": allocation["product_name"],
			"warehouse":    firstNotBlank(stringValue(batch["warehouse"]), "主仓库"),
			"quantity":     quantity,
			"cost_price":   allocation["cost_price"],
			"order_no":     order["order_no"],
		}
		if err := r.applyInventoryMovement(ctx, tenantID, operatorID, "ORDER_DELETE", mapID(order), movementRow, "in"); err != nil {
			return err
		}
	}
	return nil
}

func (r *ModuleRepository) writeSalesOrderDeleteLog(ctx context.Context, tenantID uint64, operatorID uint64, orderID uint64, reason string, now time.Time) error {
	username := ""
	if operatorID > 0 {
		user := map[string]any{}
		err := r.db.WithContext(ctx).Table("users").Select("username").Where("tenant_id = ? AND id = ?", tenantID, operatorID).Take(&user).Error
		if err != nil && err != gorm.ErrRecordNotFound {
			return err
		}
		username = stringValue(user["username"])
	}
	requestBody, err := json.Marshal(map[string]any{"reason": reason})
	if err != nil {
		return err
	}
	return r.db.WithContext(ctx).Table("operation_logs").Create(map[string]any{
		"tenant_id":    tenantID,
		"user_id":      nilIfZero(operatorID),
		"username":     username,
		"module":       "sales",
		"action":       "delete",
		"method":       "DELETE",
		"path":         fmt.Sprintf("/api/v1/sales/%d", orderID),
		"request_body": string(requestBody),
		"status_code":  200,
		"cost_ms":      0,
		"created_at":   now,
	}).Error
}

func salesOrderIsOutbound(status string) bool {
	switch strings.ToLower(strings.TrimSpace(status)) {
	case "已出库", "已发货", "outbound", "shipped", "delivered":
		return true
	default:
		return false
	}
}

func (r *ModuleRepository) customerStatementAction(ctx context.Context, tenantID uint64, operatorID uint64, action string, data map[string]any) (map[string]any, error) {
	switch action {
	case "sales-candidates":
		return r.customerStatementSalesCandidates(ctx, tenantID, data)
	case "generate":
		return r.generateCustomerStatement(ctx, tenantID, operatorID, data)
	case "confirm":
		return r.confirmCustomerStatement(ctx, tenantID, operatorID, uintValue(data["id"]))
	case "settle":
		return r.settleCustomerStatement(ctx, tenantID, operatorID, data)
	case "print":
		id := uintValue(data["id"])
		return map[string]any{"id": id, "printUrl": fmt.Sprintf("/print/customer-statement/%d?type=statement", id)}, nil
	default:
		return map[string]any{"accepted": true, "action": action}, nil
	}
}

func (r *ModuleRepository) customerStatementSalesCandidates(ctx context.Context, tenantID uint64, data map[string]any) (map[string]any, error) {
	customerID := uintValue(data["customer_id"])
	customerName := strings.TrimSpace(stringValue(data["customer_name"]))
	if customerID > 0 && customerName == "" {
		customer := map[string]any{}
		err := r.db.WithContext(ctx).Table("customers").Select("name").Where("tenant_id = ? AND id = ? AND deleted_at IS NULL", tenantID, customerID).Take(&customer).Error
		if err != nil && err != gorm.ErrRecordNotFound {
			return nil, err
		}
		customerName = strings.TrimSpace(stringValue(customer["name"]))
	}
	page, pageSize := pageFromData(data)
	rows := make([]map[string]any, 0)
	db := r.db.WithContext(ctx).Table("receivables AS ar").
		Joins("LEFT JOIN customer_statement_items AS csi ON csi.tenant_id = ar.tenant_id AND csi.deleted_at IS NULL AND (csi.receivable_id = ar.id OR (csi.source_type = ar.source_type AND csi.source_id = ar.source_id))").
		Joins("LEFT JOIN sales_orders AS so ON so.tenant_id = ar.tenant_id AND ar.source_type = 'sales' AND so.id = ar.source_id AND so.deleted_at IS NULL").
		Where("ar.tenant_id = ? AND ar.deleted_at IS NULL", tenantID)
	hasRepairOrders := r.db.Migrator().HasTable("repair_orders")
	if hasRepairOrders {
		db = db.Joins("LEFT JOIN repair_orders AS ro ON ro.tenant_id = ar.tenant_id AND ar.source_type = 'repair' AND ro.id = ar.source_id AND ro.deleted_at IS NULL")
	}
	if customerID > 0 {
		if customerName != "" {
			db = db.Where("ar.customer_id = ? OR ar.customer_name = ?", customerID, customerName)
		} else {
			db = db.Where("ar.customer_id = ?", customerID)
		}
	}
	if keyword := strings.TrimSpace(stringValue(data["keyword"])); keyword != "" {
		like := "%" + keyword + "%"
		if hasRepairOrders {
			db = db.Where("ar.source_no LIKE ? OR ar.customer_name LIKE ? OR so.product_name LIKE ? OR ro.device_name LIKE ?", like, like, like, like)
		} else {
			db = db.Where("ar.source_no LIKE ? OR ar.customer_name LIKE ? OR so.product_name LIKE ?", like, like, like)
		}
	}
	if start := strings.TrimSpace(stringValue(data["start_date"])); start != "" {
		db = db.Where("ar.invoice_date >= ?", start)
	}
	if end := strings.TrimSpace(stringValue(data["end_date"])); end != "" {
		db = db.Where("ar.invoice_date < ?", exclusiveEndDateArg(end))
	}
	if payStatus := strings.TrimSpace(stringValue(data["payment_status"])); payStatus != "" {
		switch payStatus {
		case "unpaid":
			db = db.Where("ar.received_amount = 0 AND ar.balance_amount > 0")
		case "partial":
			db = db.Where("ar.received_amount > 0 AND ar.balance_amount > 0")
		case "paid":
			db = db.Where("ar.balance_amount <= 0")
		}
	} else {
		db = db.Where("ar.balance_amount > 0")
	}
	if reconciled := strings.TrimSpace(stringValue(data["reconciled"])); reconciled == "" || reconciled == "unreconciled" {
		db = db.Where("csi.id IS NULL")
	} else if reconciled == "reconciled" {
		db = db.Where("csi.id IS NOT NULL")
	}
	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, err
	}
	if total == 0 {
		return r.customerStatementLegacySalesCandidates(ctx, tenantID, data, customerID, customerName, page, pageSize)
	}
	order := safeDetailOrder(data, map[string]string{"orderDate": "ar.invoice_date", "totalAmount": "ar.total_amount"}, "ar.invoice_date")
	productNameExpr := "so.product_name"
	quantityExpr := "so.quantity"
	if hasRepairOrders {
		productNameExpr = "CASE WHEN ar.source_type = 'sales' THEN so.product_name ELSE ro.device_name END"
		quantityExpr = "CASE WHEN ar.source_type = 'sales' THEN so.quantity ELSE 1 END"
	}
	if err := db.Select(fmt.Sprintf(`ar.id AS receivable_id,
		CASE WHEN ar.source_type = 'sales' THEN ar.source_id ELSE 0 END AS sale_id,
		ar.source_type, ar.source_id, ar.source_no, ar.source_no AS sale_no, ar.invoice_date AS sale_date,
		ar.customer_id, ar.customer_name,
		%s AS product_name,
		%s AS quantity,
		ar.total_amount, ar.received_amount, ar.balance_amount AS unpaid_amount,
		ar.status AS settlement_status,
		CASE WHEN csi.id IS NULL THEN 0 ELSE 1 END AS reconciled`, productNameExpr, quantityExpr)).
		Order(order).Offset((page - 1) * pageSize).Limit(pageSize).Find(&rows).Error; err != nil {
		return nil, err
	}
	return map[string]any{"list": normalizeRows(rows), "total": total}, nil
}

func (r *ModuleRepository) customerStatementLegacySalesCandidates(ctx context.Context, tenantID uint64, data map[string]any, customerID uint64, customerName string, page int, pageSize int) (map[string]any, error) {
	rows := make([]map[string]any, 0)
	db := r.db.WithContext(ctx).Table("sales_orders AS so").
		Joins("LEFT JOIN customer_statement_items AS csi ON csi.tenant_id = so.tenant_id AND csi.sale_id = so.id AND csi.deleted_at IS NULL").
		Where("so.tenant_id = ? AND so.deleted_at IS NULL", tenantID)
	if customerID > 0 {
		if customerName != "" {
			db = db.Where("so.customer_id = ? OR so.customer_name = ?", customerID, customerName)
		} else {
			db = db.Where("so.customer_id = ?", customerID)
		}
	}
	if keyword := strings.TrimSpace(stringValue(data["keyword"])); keyword != "" {
		like := "%" + keyword + "%"
		db = db.Where("so.order_no LIKE ? OR so.customer_name LIKE ?", like, like)
	}
	if start := strings.TrimSpace(stringValue(data["start_date"])); start != "" {
		db = db.Where("so.order_date >= ?", start)
	}
	if end := strings.TrimSpace(stringValue(data["end_date"])); end != "" {
		db = db.Where("so.order_date < ?", exclusiveEndDateArg(end))
	}
	if payStatus := strings.TrimSpace(stringValue(data["payment_status"])); payStatus != "" {
		unpaidExpr := "CASE WHEN so.receivable_amount > 0 THEN so.receivable_amount WHEN so.total_amount > so.received_amount THEN so.total_amount - so.received_amount ELSE 0 END"
		switch payStatus {
		case "unpaid":
			db = db.Where("so.received_amount = 0 AND " + unpaidExpr + " > 0")
		case "partial":
			db = db.Where("so.received_amount > 0 AND " + unpaidExpr + " > 0")
		case "paid":
			db = db.Where(unpaidExpr + " <= 0")
		}
	}
	if reconciled := strings.TrimSpace(stringValue(data["reconciled"])); reconciled == "" || reconciled == "unreconciled" {
		db = db.Where("csi.id IS NULL")
	} else if reconciled == "reconciled" {
		db = db.Where("csi.id IS NOT NULL")
	}
	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, err
	}
	order := safeDetailOrder(data, map[string]string{"orderDate": "so.order_date", "totalAmount": "so.total_amount"}, "so.order_date")
	if err := db.Select(`so.id AS sale_id, so.order_no AS sale_no, so.order_date AS sale_date,
		so.customer_id, so.customer_name, so.product_name, so.quantity,
		CASE WHEN so.total_amount > 0 THEN so.total_amount ELSE so.quantity * so.price END AS total_amount,
		so.received_amount,
		CASE
			WHEN so.receivable_amount > 0 THEN so.receivable_amount
			WHEN so.total_amount > so.received_amount THEN so.total_amount - so.received_amount
			WHEN (so.quantity * so.price) > so.received_amount THEN (so.quantity * so.price) - so.received_amount
			ELSE 0
		END AS unpaid_amount,
		CASE
			WHEN (CASE WHEN so.receivable_amount > 0 THEN so.receivable_amount WHEN so.total_amount > so.received_amount THEN so.total_amount - so.received_amount WHEN (so.quantity * so.price) > so.received_amount THEN (so.quantity * so.price) - so.received_amount ELSE 0 END) <= 0 THEN 'paid'
			WHEN so.received_amount > 0 THEN 'partial'
			ELSE 'unpaid'
		END AS settlement_status,
		CASE WHEN csi.id IS NULL THEN 0 ELSE 1 END AS reconciled`).
		Order(order).Offset((page - 1) * pageSize).Limit(pageSize).Find(&rows).Error; err != nil {
		return nil, err
	}
	return map[string]any{"list": normalizeRows(rows), "total": total}, nil
}

func (r *ModuleRepository) generateCustomerStatement(ctx context.Context, tenantID uint64, operatorID uint64, data map[string]any) (map[string]any, error) {
	customerID := uintValue(data["customer_id"])
	receivableIDs := uintSliceValue(data["receivable_ids"])
	saleIDs := uintSliceValue(data["sale_ids"])
	if customerID == 0 || (len(receivableIDs) == 0 && len(saleIDs) == 0) {
		return nil, apperrors.New(40042, "请选择客户和业务单据", 400)
	}
	var created map[string]any
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		repo := &ModuleRepository{db: tx}
		customer := map[string]any{}
		if err := tx.Table("customers").Where("tenant_id = ? AND id = ? AND deleted_at IS NULL", tenantID, customerID).Take(&customer).Error; err != nil {
			return err
		}
		customerName := strings.TrimSpace(stringValue(customer["name"]))
		if len(receivableIDs) == 0 && len(saleIDs) > 0 {
			var duplicate int64
			if err := tx.Table("customer_statement_items").Where("tenant_id = ? AND sale_id IN ? AND deleted_at IS NULL", tenantID, saleIDs).Count(&duplicate).Error; err != nil {
				return err
			}
			if duplicate > 0 {
				return apperrors.New(40043, "所选业务单据已生成对账单，不能重复加入未完成对账单", 400)
			}
			sales := make([]map[string]any, 0)
			salesDB := tx.Table("sales_orders").Where("tenant_id = ? AND id IN ? AND deleted_at IS NULL", tenantID, saleIDs)
			if customerName != "" {
				salesDB = salesDB.Where("customer_id = ? OR customer_name = ?", customerID, customerName)
			} else {
				salesDB = salesDB.Where("customer_id = ?", customerID)
			}
			if err := salesDB.Order("order_date ASC, id ASC").Find(&sales).Error; err != nil {
				return err
			}
			if len(sales) != len(saleIDs) {
				return apperrors.New(40044, "业务单据与所选客户不一致", 400)
			}
			for _, sale := range sales {
				saleTotal, saleReceived, saleUnpaid := statementSaleAmounts(sale)
				if err := repo.syncStatementSaleReceivable(ctx, tenantID, operatorID, customer, sale, saleTotal, saleReceived, saleUnpaid); err != nil {
					return err
				}
			}
		}
		receivables := make([]map[string]any, 0)
		receivableDB := tx.Table("receivables").Where("tenant_id = ? AND deleted_at IS NULL", tenantID)
		if len(receivableIDs) > 0 {
			receivableDB = receivableDB.Where("id IN ?", receivableIDs)
		} else {
			receivableDB = receivableDB.Where("source_type = ? AND source_id IN ?", "sales", saleIDs)
		}
		if customerName != "" {
			receivableDB = receivableDB.Where("customer_id = ? OR customer_name = ?", customerID, customerName)
		} else {
			receivableDB = receivableDB.Where("customer_id = ?", customerID)
		}
		if err := receivableDB.Order("invoice_date ASC, id ASC").Find(&receivables).Error; err != nil {
			return err
		}
		expected := len(receivableIDs)
		if expected == 0 {
			expected = len(saleIDs)
		}
		if len(receivables) != expected {
			return apperrors.New(40044, "业务单据与所选客户不一致", 400)
		}
		var duplicate int64
		dupDB := tx.Table("customer_statement_items").Where("tenant_id = ? AND deleted_at IS NULL", tenantID)
		if len(receivableIDs) > 0 {
			dupDB = dupDB.Where("receivable_id IN ?", receivableIDs)
		} else {
			dupDB = dupDB.Where("sale_id IN ?", saleIDs)
		}
		if err := dupDB.Count(&duplicate).Error; err != nil {
			return err
		}
		if duplicate > 0 {
			return apperrors.New(40043, "所选业务单据已生成对账单，不能重复加入未完成对账单", 400)
		}
		var total, received, unpaid float64
		for _, receivable := range receivables {
			total += numberValue(receivable["total_amount"])
			received += numberValue(receivable["received_amount"])
			unpaid += numberValue(receivable["balance_amount"])
		}
		if err := repo.refreshCustomerReceivable(ctx, tenantID, customerID); err != nil {
			return err
		}
		var cumulative float64
		_ = tx.Table("receivables").Where("tenant_id = ? AND customer_id = ? AND deleted_at IS NULL", tenantID, customerID).Select("COALESCE(SUM(balance_amount), 0)").Scan(&cumulative).Error
		now := time.Now()
		statementNo := nextNo("customer-statements")
		statement := map[string]any{
			"tenant_id": tenantID, "statement_no": statementNo,
			"customer_id": customerID, "customer_name": customer["name"],
			"contact_name": customer["contact_name"], "contact_phone": customer["phone"],
			"start_date":   timeValue(data["start_date"], timeValue(receivables[0]["invoice_date"], now)),
			"end_date":     timeValue(data["end_date"], timeValue(receivables[len(receivables)-1]["invoice_date"], now)),
			"total_amount": total, "received_amount": received, "unpaid_amount": unpaid, "cumulative_debt": cumulative,
			"status": "unconfirmed", "remark": data["remark"],
			"created_by": operatorID, "updated_by": operatorID, "created_at": now, "updated_at": now,
		}
		if err := tx.Table("customer_statements").Create(statement).Error; err != nil {
			return err
		}
		statementID := uint64(0)
		if err := tx.Table("customer_statements").Where("tenant_id = ? AND statement_no = ?", tenantID, statementNo).Select("id").Scan(&statementID).Error; err != nil {
			return err
		}
		if statementID == 0 {
			return apperrors.New(40046, "客户对账单生成失败", 500)
		}
		for _, receivable := range receivables {
			sourceType := firstNotBlank(stringValue(receivable["source_type"]), "sales")
			sourceID := uintValue(receivable["source_id"])
			sourceNo := stringValue(receivable["source_no"])
			saleID := uint64(0)
			if sourceType == "sales" {
				saleID = sourceID
			}
			item := map[string]any{
				"tenant_id": tenantID, "statement_id": statementID, "sale_id": saleID,
				"sale_no": sourceNo, "sale_date": receivable["invoice_date"],
				"source_type": sourceType, "source_id": nilIfZero(sourceID), "source_no": sourceNo,
				"receivable_id":   nilIfZero(mapID(receivable)),
				"product_name":    repo.statementReceivableProductName(ctx, tenantID, receivable),
				"quantity":        repo.statementReceivableQuantity(ctx, tenantID, receivable),
				"total_amount":    numberValue(receivable["total_amount"]),
				"received_amount": numberValue(receivable["received_amount"]), "unpaid_amount": numberValue(receivable["balance_amount"]),
				"settlement_status": statementSettlementStatus(numberValue(receivable["received_amount"]), numberValue(receivable["balance_amount"])),
				"created_by":        operatorID, "updated_by": operatorID, "created_at": now, "updated_at": now,
			}
			if err := tx.Table("customer_statement_items").Create(item).Error; err != nil {
				return err
			}
		}
		result, err := repo.Find(ctx, "customer-statements", tenantID, statementID)
		if err != nil {
			return err
		}
		created = result
		return nil
	})
	return created, err
}

func (r *ModuleRepository) updateCustomerStatementStatus(ctx context.Context, tenantID uint64, operatorID uint64, id uint64, status string) (map[string]any, error) {
	if id == 0 {
		return nil, apperrors.New(40045, "请选择对账单", 400)
	}
	now := time.Now()
	updates := map[string]any{"status": status, "updated_by": operatorID, "updated_at": now}
	if status == "confirmed" {
		updates["confirmed_at"] = now
	}
	if status == "settled" {
		updates["settled_at"] = now
	}
	if err := r.db.WithContext(ctx).Table("customer_statements").Where("tenant_id = ? AND id = ? AND deleted_at IS NULL", tenantID, id).Updates(updates).Error; err != nil {
		return nil, err
	}
	return r.Find(ctx, "customer-statements", tenantID, id)
}

func (r *ModuleRepository) confirmCustomerStatement(ctx context.Context, tenantID uint64, operatorID uint64, id uint64) (map[string]any, error) {
	if id == 0 {
		return nil, apperrors.New(40045, "请选择对账单", 400)
	}
	var result map[string]any
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		repo := &ModuleRepository{db: tx}
		statement := map[string]any{}
		if err := tx.Table("customer_statements").Where("tenant_id = ? AND id = ? AND deleted_at IS NULL", tenantID, id).Take(&statement).Error; err != nil {
			return err
		}
		if stringValue(statement["status"]) != "settled" {
			if err := repo.syncCustomerStatementReceivables(ctx, tenantID, operatorID, statement); err != nil {
				return err
			}
			if err := repo.recalculateCustomerStatement(ctx, tenantID, operatorID, id, "confirmed"); err != nil {
				return err
			}
		}
		row, err := repo.Find(ctx, "customer-statements", tenantID, id)
		if err != nil {
			return err
		}
		result = row
		return nil
	})
	return result, err
}

func (r *ModuleRepository) settleCustomerStatement(ctx context.Context, tenantID uint64, operatorID uint64, data map[string]any) (map[string]any, error) {
	id := uintValue(data["id"])
	if id == 0 {
		return nil, apperrors.New(40045, "请选择对账单", 400)
	}
	accountName := firstNotBlank(stringValue(data["account_name"]), "现金")
	var result map[string]any
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		repo := &ModuleRepository{db: tx}
		statement := map[string]any{}
		if err := tx.Table("customer_statements").Where("tenant_id = ? AND id = ? AND deleted_at IS NULL", tenantID, id).Take(&statement).Error; err != nil {
			return err
		}
		if stringValue(statement["status"]) == "settled" {
			result = normalizeRow(statement)
			return nil
		}
		if err := repo.syncCustomerStatementReceivables(ctx, tenantID, operatorID, statement); err != nil {
			return err
		}
		amountLeft := numberValue(data["amount"])
		if amountLeft <= 0 {
			amountLeft = numberValue(statement["unpaid_amount"])
		}
		items := make([]map[string]any, 0)
		if err := tx.Table("customer_statement_items").Where("tenant_id = ? AND statement_id = ? AND deleted_at IS NULL", tenantID, id).Order("sale_date ASC, id ASC").Find(&items).Error; err != nil {
			return err
		}
		now := time.Now()
		applied := 0.0
		for _, item := range items {
			if amountLeft <= 0 {
				break
			}
			receivable := map[string]any{}
			receivableDB := tx.Table("receivables").Where("tenant_id = ? AND deleted_at IS NULL", tenantID)
			if receivableID := uintValue(item["receivable_id"]); receivableID > 0 {
				receivableDB = receivableDB.Where("id = ?", receivableID)
			} else {
				sourceType := firstNotBlank(stringValue(item["source_type"]), "sales")
				sourceID := uintValue(item["source_id"])
				if sourceID == 0 {
					sourceID = uintValue(item["sale_id"])
				}
				receivableDB = receivableDB.Where("source_type = ?", sourceType).Where("source_id = ? OR source_no = ?", sourceID, stringValue(item["sale_no"]))
			}
			err := receivableDB.Order("id ASC").Take(&receivable).Error
			if err == gorm.ErrRecordNotFound {
				continue
			}
			if err != nil {
				return err
			}
			balance := numberValue(receivable["balance_amount"])
			if balance <= 0 {
				continue
			}
			apply := balance
			if apply > amountLeft {
				apply = amountLeft
			}
			newReceived := numberValue(receivable["received_amount"]) + apply
			newBalance := balance - apply
			status := receivableStatus(newBalance, timeValue(receivable["due_date"], now))
			if newBalance > 0 && newReceived > 0 {
				status = "partial"
			}
			if err := tx.Table("receivables").Where("id = ?", mapID(receivable)).Updates(map[string]any{
				"received_amount": newReceived,
				"balance_amount":  newBalance,
				"status":          status,
				"updated_by":      operatorID,
				"updated_at":      now,
			}).Error; err != nil {
				return err
			}
			if stringValue(receivable["source_type"]) == "sales" {
				if err := repo.updateSalesCollection(ctx, tenantID, uintValue(receivable["source_id"]), newReceived, newBalance); err != nil {
					return err
				}
			}
			itemReceived := numberValue(item["received_amount"]) + apply
			itemUnpaid := numberValue(item["unpaid_amount"]) - apply
			if itemUnpaid < 0 {
				itemUnpaid = 0
			}
			if err := tx.Table("customer_statement_items").Where("id = ?", mapID(item)).Updates(map[string]any{
				"received_amount":   itemReceived,
				"unpaid_amount":     itemUnpaid,
				"settlement_status": statementSettlementStatus(itemReceived, itemUnpaid),
				"updated_by":        operatorID,
				"updated_at":        now,
			}).Error; err != nil {
				return err
			}
			applied += apply
			amountLeft -= apply
		}
		if applied > 0 {
			if err := repo.createFinanceReceipt(ctx, tenantID, operatorID, stringValue(statement["customer_name"]), applied, stringValue(statement["statement_no"]), "客户对账单收款", accountName); err != nil {
				return err
			}
		}
		if err := repo.refreshCustomerReceivable(ctx, tenantID, uintValue(statement["customer_id"])); err != nil {
			return err
		}
		if err := repo.recalculateCustomerStatement(ctx, tenantID, operatorID, id, "settled"); err != nil {
			return err
		}
		row, err := repo.Find(ctx, "customer-statements", tenantID, id)
		if err != nil {
			return err
		}
		result = row
		return nil
	})
	return result, err
}

func (r *ModuleRepository) syncCustomerStatementReceivables(ctx context.Context, tenantID uint64, operatorID uint64, statement map[string]any) error {
	customerID := uintValue(statement["customer_id"])
	customer := map[string]any{}
	if err := r.db.WithContext(ctx).Table("customers").Where("tenant_id = ? AND id = ? AND deleted_at IS NULL", tenantID, customerID).Take(&customer).Error; err != nil {
		return err
	}
	items := make([]map[string]any, 0)
	if err := r.db.WithContext(ctx).Table("customer_statement_items").Where("tenant_id = ? AND statement_id = ? AND deleted_at IS NULL", tenantID, mapID(statement)).Find(&items).Error; err != nil {
		return err
	}
	for _, item := range items {
		receivable := map[string]any{}
		db := r.db.WithContext(ctx).Table("receivables").Where("tenant_id = ? AND deleted_at IS NULL", tenantID)
		if receivableID := uintValue(item["receivable_id"]); receivableID > 0 {
			db = db.Where("id = ?", receivableID)
		} else {
			sourceType := firstNotBlank(stringValue(item["source_type"]), "sales")
			sourceID := uintValue(item["source_id"])
			if sourceID == 0 {
				sourceID = uintValue(item["sale_id"])
			}
			db = db.Where("source_type = ?", sourceType).Where("source_id = ? OR source_no = ?", sourceID, stringValue(item["sale_no"]))
		}
		if err := db.Order("id ASC").Take(&receivable).Error; err != nil {
			return err
		}
		total := numberValue(receivable["total_amount"])
		received := numberValue(receivable["received_amount"])
		unpaid := numberValue(receivable["balance_amount"])
		if err := r.db.WithContext(ctx).Table("customer_statement_items").Where("id = ?", mapID(item)).Updates(map[string]any{
			"total_amount":      total,
			"received_amount":   received,
			"unpaid_amount":     unpaid,
			"settlement_status": statementSettlementStatus(received, unpaid),
			"receivable_id":     nilIfZero(mapID(receivable)),
			"source_type":       receivable["source_type"],
			"source_id":         receivable["source_id"],
			"source_no":         receivable["source_no"],
			"updated_by":        operatorID,
			"updated_at":        time.Now(),
		}).Error; err != nil {
			return err
		}
	}
	return r.refreshCustomerReceivable(ctx, tenantID, customerID)
}

func (r *ModuleRepository) recalculateCustomerStatement(ctx context.Context, tenantID uint64, operatorID uint64, id uint64, preferredStatus string) error {
	sums := map[string]any{}
	if err := r.db.WithContext(ctx).Table("customer_statement_items").
		Where("tenant_id = ? AND statement_id = ? AND deleted_at IS NULL", tenantID, id).
		Select("COALESCE(SUM(total_amount), 0) AS total_amount, COALESCE(SUM(received_amount), 0) AS received_amount, COALESCE(SUM(unpaid_amount), 0) AS unpaid_amount").
		Scan(&sums).Error; err != nil {
		return err
	}
	statement := map[string]any{}
	if err := r.db.WithContext(ctx).Table("customer_statements").Where("tenant_id = ? AND id = ?", tenantID, id).Take(&statement).Error; err != nil {
		return err
	}
	var cumulative float64
	_ = r.db.WithContext(ctx).Table("receivables").Where("tenant_id = ? AND customer_id = ? AND deleted_at IS NULL", tenantID, uintValue(statement["customer_id"])).Select("COALESCE(SUM(balance_amount), 0)").Scan(&cumulative).Error
	now := time.Now()
	unpaid := numberValue(sums["unpaid_amount"])
	status := preferredStatus
	if preferredStatus == "settled" && unpaid > 0 {
		status = "confirmed"
	}
	updates := map[string]any{
		"total_amount":    numberValue(sums["total_amount"]),
		"received_amount": numberValue(sums["received_amount"]),
		"unpaid_amount":   unpaid,
		"cumulative_debt": cumulative,
		"status":          status,
		"updated_by":      operatorID,
		"updated_at":      now,
	}
	if status == "confirmed" {
		updates["confirmed_at"] = now
	}
	if status == "settled" {
		updates["confirmed_at"] = now
		updates["settled_at"] = now
	}
	return r.db.WithContext(ctx).Table("customer_statements").Where("tenant_id = ? AND id = ?", tenantID, id).Updates(updates).Error
}

func (r *ModuleRepository) Action(ctx context.Context, module string, tenantID uint64, operatorID uint64, action string, data map[string]any) (map[string]any, error) {
	if module == "profit-report" {
		return r.profitReportAction(ctx, tenantID, action, normalizeInput(data))
	}
	if module == "inventory-asset-report" {
		return r.inventoryAssetReportAction(ctx, tenantID, action, normalizeInput(data))
	}
	if module == "customer-statements" {
		return r.customerStatementAction(ctx, tenantID, operatorID, action, normalizeInput(data))
	}
	if module == "receivables" && action == "receive" {
		return r.receiveReceivable(ctx, tenantID, operatorID, normalizeInput(data))
	}
	if module == "payables" && action == "pay" {
		return r.payPayable(ctx, tenantID, operatorID, normalizeInput(data))
	}
	if module == "repair" && action == "outsource-stats" {
		return r.repairOutsourceStats(ctx, tenantID, normalizeInput(data))
	}
	if module == "repair" && action == "parts-stats" {
		return r.repairPartsStats(ctx, tenantID, normalizeInput(data))
	}
	if module == "inventory" && action == "detail-tab" {
		return r.inventoryDetailTab(ctx, tenantID, normalizeInput(data))
	}
	if module == "sales" && action == "purchase-sources" {
		return r.salesPurchaseSources(ctx, tenantID, normalizeInput(data))
	}
	if module == "sales" {
		return r.salesAction(ctx, tenantID, operatorID, action, normalizeInput(data))
	}
	if strings.Contains(action, "print") || action == "statement" || action == "reminder" || action == "receipt" {
		id := uintValue(data["id"])
		if id == 0 {
			id = uintValue(data[module+"_id"])
		}
		if id > 0 {
			printModule := module
			printType := action
			if module == "receivables" {
				printModule = "statement"
			}
			if module == "finance" || action == "receipt" {
				printModule = "receipt"
			}
			return map[string]any{"id": id, "printUrl": fmt.Sprintf("/print/%s/%d?type=%s", printModule, id, printType)}, nil
		}
	}
	result := map[string]any{
		"module":     module,
		"action":     action,
		"accepted":   true,
		"operatorId": operatorID,
		"tenantId":   tenantID,
		"requestId":  uuid.NewString(),
		"data":       data,
	}
	return result, nil
}

func (r *ModuleRepository) repairOutsourceStats(ctx context.Context, tenantID uint64, data map[string]any) (map[string]any, error) {
	page, pageSize := pageFromData(data)
	rows := make([]map[string]any, 0)
	db := r.db.WithContext(ctx).Table("repair_order_items AS roi").
		Joins("JOIN repair_orders AS ro ON ro.tenant_id = roi.tenant_id AND ro.id = roi.repair_order_id AND ro.deleted_at IS NULL").
		Joins("LEFT JOIN payables AS p ON p.tenant_id = roi.tenant_id AND p.source_type = ? AND p.source_id = ro.id AND p.supplier_id = roi.supplier_id AND p.deleted_at IS NULL", "repair_outsource").
		Where("roi.tenant_id = ? AND roi.item_type = ? AND roi.deleted_at IS NULL", tenantID, "outsource")
	if supplier := stringValue(data["supplier_name"]); supplier != "" {
		db = db.Where("roi.supplier_name LIKE ?", "%"+supplier+"%")
	}
	db = applyDateRange(db, "ro.registered_at", data)
	var total int64
	countDB := r.db.WithContext(ctx).Table("(?) AS grouped", db.Select("roi.supplier_id, roi.supplier_name").Group("roi.supplier_id, roi.supplier_name"))
	if err := countDB.Count(&total).Error; err != nil {
		return nil, err
	}
	if err := db.Select(`roi.supplier_id, roi.supplier_name,
		COUNT(roi.id) AS outsource_count,
		COUNT(DISTINCT ro.id) AS repair_order_count,
		COALESCE(SUM(roi.cost_amount), 0) AS outsource_amount,
		COALESCE(SUM(p.paid_amount), 0) AS paid_amount,
		COALESCE(SUM(p.balance_amount), 0) AS unpaid_amount`).
		Group("roi.supplier_id, roi.supplier_name").
		Order("outsource_amount DESC").
		Offset((page - 1) * pageSize).Limit(pageSize).Find(&rows).Error; err != nil {
		return nil, err
	}
	return map[string]any{"list": normalizeRows(rows), "total": total, "page": page, "pageSize": pageSize}, nil
}

func (r *ModuleRepository) repairPartsStats(ctx context.Context, tenantID uint64, data map[string]any) (map[string]any, error) {
	page, pageSize := pageFromData(data)
	rows := make([]map[string]any, 0)
	db := r.db.WithContext(ctx).Table("repair_order_items AS roi").
		Joins("JOIN repair_orders AS ro ON ro.tenant_id = roi.tenant_id AND ro.id = roi.repair_order_id AND ro.deleted_at IS NULL").
		Where("roi.tenant_id = ? AND roi.item_type = ? AND roi.deleted_at IS NULL", tenantID, "part")
	if product := stringValue(data["product_name"]); product != "" {
		db = db.Where("roi.product_name LIKE ? OR roi.product_code LIKE ?", "%"+product+"%", "%"+product+"%")
	}
	db = applyDateRange(db, "ro.registered_at", data)
	var total int64
	countDB := r.db.WithContext(ctx).Table("(?) AS grouped", db.Select("roi.product_id, roi.product_code, roi.product_name").Group("roi.product_id, roi.product_code, roi.product_name"))
	if err := countDB.Count(&total).Error; err != nil {
		return nil, err
	}
	if err := db.Select(`roi.product_id, roi.product_code, roi.product_name,
		COUNT(roi.id) AS use_count,
		COUNT(DISTINCT ro.id) AS repair_order_count,
		COALESCE(SUM(roi.quantity), 0) AS quantity,
		COALESCE(SUM(roi.cost_amount), 0) AS cost_amount`).
		Group("roi.product_id, roi.product_code, roi.product_name").
		Order("cost_amount DESC").
		Offset((page - 1) * pageSize).Limit(pageSize).Find(&rows).Error; err != nil {
		return nil, err
	}
	return map[string]any{"list": normalizeRows(rows), "total": total, "page": page, "pageSize": pageSize}, nil
}

func (r *ModuleRepository) profitReportAction(ctx context.Context, tenantID uint64, action string, data map[string]any) (map[string]any, error) {
	q := queryFromData(data)
	switch action {
	case "summary", "":
		summary, err := r.profitReportSummary(ctx, tenantID, q)
		if err != nil {
			return nil, err
		}
		return map[string]any{"summary": summary}, nil
	case "ranking":
		ranking, err := r.profitReportRanking(ctx, tenantID, q)
		if err != nil {
			return nil, err
		}
		return map[string]any{"ranking": ranking}, nil
	case "trend":
		trend, err := r.profitReportTrend(ctx, tenantID, q)
		if err != nil {
			return nil, err
		}
		return map[string]any{"trend": trend}, nil
	default:
		return nil, apperrors.New(40024, "不支持的利润报表操作", 400)
	}
}

func (r *ModuleRepository) inventoryAssetReportAction(ctx context.Context, tenantID uint64, action string, data map[string]any) (map[string]any, error) {
	q := queryFromData(data)
	switch action {
	case "summary", "":
		summary, err := r.inventoryAssetSummary(ctx, tenantID, q)
		if err != nil {
			return nil, err
		}
		return map[string]any{"summary": summary}, nil
	case "aging":
		rows, err := r.inventoryAssetAging(ctx, tenantID, q)
		if err != nil {
			return nil, err
		}
		return map[string]any{"aging": rows}, nil
	case "slow-moving":
		rows, err := r.inventorySlowMoving(ctx, tenantID, q, int(numberValue(data["days"])))
		if err != nil {
			return nil, err
		}
		return map[string]any{"list": rows}, nil
	case "trend":
		trend, err := r.inventoryAssetTrend(ctx, tenantID, q)
		if err != nil {
			return nil, err
		}
		return map[string]any{"trend": trend}, nil
	default:
		return nil, apperrors.New(40025, "不支持的库存资产报表操作", 400)
	}
}

func (r *ModuleRepository) profitReportSummary(ctx context.Context, tenantID uint64, q query.PageQuery) (map[string]any, error) {
	summary := map[string]any{}
	sales := map[string]any{}
	if err := r.profitSalesBase(ctx, tenantID, q).
		Select(`COALESCE(SUM(sca.quantity * COALESCE(NULLIF(soi.price, 0), so.price, 0)), 0) AS sales_income,
			COALESCE(SUM(sca.cost_amount), 0) AS sales_cost,
			COALESCE(SUM(sca.quantity * COALESCE(NULLIF(soi.price, 0), so.price, 0) - sca.cost_amount), 0) AS sales_profit,
			COALESCE(SUM(sca.quantity), 0) AS sales_quantity`).
		Scan(&sales).Error; err != nil {
		return nil, err
	}
	repair := map[string]any{}
	if r.db.Migrator().HasTable("repair_orders") {
		db := r.db.WithContext(ctx).Table("repair_orders").
			Where("tenant_id = ? AND deleted_at IS NULL", tenantID)
		db = applyQueryDateRange(db, "registered_at", q)
		_ = db.Select(`COALESCE(SUM(total_amount), 0) AS repair_income,
			COALESCE(SUM(parts_cost), 0) AS repair_cost,
			COALESCE(SUM(total_amount - parts_cost), 0) AS repair_profit`).Scan(&repair).Error
	}
	project := map[string]any{}
	if r.db.Migrator().HasTable("project_projects") {
		db := r.db.WithContext(ctx).Table("project_projects").
			Where("tenant_id = ? AND deleted_at IS NULL", tenantID)
		db = applyQueryDateRange(db, "COALESCE(end_date, start_date)", q)
		_ = db.Select(`COALESCE(SUM(CASE WHEN settle_amount > 0 THEN settle_amount ELSE contract_amount END), 0) AS project_income,
			COALESCE(SUM(cost_amount), 0) AS project_cost,
			COALESCE(SUM((CASE WHEN settle_amount > 0 THEN settle_amount ELSE contract_amount END) - cost_amount), 0) AS project_profit`).Scan(&project).Error
	}
	finance := map[string]any{}
	if r.db.Migrator().HasTable("finance_records") {
		db := r.db.WithContext(ctx).Table("finance_records").
			Where("tenant_id = ? AND deleted_at IS NULL", tenantID)
		db = applyQueryDateRange(db, "occurred_at", q)
		_ = db.Select(`COALESCE(SUM(CASE WHEN source_type = 'daily' AND (record_type LIKE '%收入%' OR business_type LIKE '%收入%' OR business_type LIKE '%其他收入%') THEN amount ELSE 0 END), 0) AS other_income,
			COALESCE(SUM(CASE WHEN source_type = 'daily' AND (record_type LIKE '%支出%' OR business_type LIKE '%支出%' OR business_type LIKE '%费用%') THEN amount ELSE 0 END), 0) AS expense_amount`).Scan(&finance).Error
	}
	salesIncome := numberValue(sales["sales_income"])
	salesCost := numberValue(sales["sales_cost"])
	salesProfit := numberValue(sales["sales_profit"])
	repairIncome := numberValue(repair["repair_income"])
	repairProfit := numberValue(repair["repair_profit"])
	projectIncome := numberValue(project["project_income"])
	projectProfit := numberValue(project["project_profit"])
	otherIncome := numberValue(finance["other_income"])
	expenseAmount := numberValue(finance["expense_amount"])
	totalIncome := salesIncome + repairIncome + projectIncome + otherIncome
	netProfit := salesProfit + repairProfit + projectProfit + otherIncome - expenseAmount
	summary["salesIncome"] = salesIncome
	summary["salesCost"] = salesCost
	summary["salesProfit"] = salesProfit
	summary["repairIncome"] = repairIncome
	summary["repairProfit"] = repairProfit
	summary["projectIncome"] = projectIncome
	summary["projectProfit"] = projectProfit
	summary["otherIncome"] = otherIncome
	summary["expenseAmount"] = expenseAmount
	summary["netProfit"] = netProfit
	summary["grossProfitRate"] = percentValue(salesProfit, salesIncome)
	summary["netProfitRate"] = percentValue(netProfit, totalIncome)
	return normalizeRow(summary), nil
}

func (r *ModuleRepository) profitReportRanking(ctx context.Context, tenantID uint64, q query.PageQuery) (map[string]any, error) {
	result := map[string]any{}
	for key, group := range map[string]string{
		"products":     "sca.product_name",
		"customers":    "so.customer_name",
		"brands":       "COALESCE(p.brand, '')",
		"salespersons": "COALESCE(u.real_name, u.username, '未分配')",
	} {
		rows := make([]map[string]any, 0)
		err := r.profitSalesBase(ctx, tenantID, q).
			Joins("LEFT JOIN users AS u ON u.id = so.created_by").
			Select(group + ` AS name,
				COALESCE(SUM(sca.quantity), 0) AS quantity,
				COALESCE(SUM(sca.quantity * COALESCE(NULLIF(soi.price, 0), so.price, 0)), 0) AS sales_amount,
				COALESCE(SUM(sca.cost_amount), 0) AS cost_amount,
				COALESCE(SUM(sca.quantity * COALESCE(NULLIF(soi.price, 0), so.price, 0) - sca.cost_amount), 0) AS profit_amount`).
			Group(group).Order("profit_amount DESC").Limit(10).Find(&rows).Error
		if err != nil {
			return nil, err
		}
		result[key] = normalizeRows(rows)
	}
	return result, nil
}

func (r *ModuleRepository) profitReportTrend(ctx context.Context, tenantID uint64, q query.PageQuery) ([]map[string]any, error) {
	rows := make([]map[string]any, 0)
	err := r.profitSalesBase(ctx, tenantID, q).
		Select(`DATE(so.order_date) AS date,
			COALESCE(SUM(sca.quantity * COALESCE(NULLIF(soi.price, 0), so.price, 0)), 0) AS sales_amount,
			COALESCE(SUM(sca.cost_amount), 0) AS cost_amount,
			COALESCE(SUM(sca.quantity * COALESCE(NULLIF(soi.price, 0), so.price, 0) - sca.cost_amount), 0) AS profit_amount`).
		Group("DATE(so.order_date)").Order("DATE(so.order_date) ASC").Find(&rows).Error
	return normalizeRows(rows), err
}

func (r *ModuleRepository) inventoryAssetSummary(ctx context.Context, tenantID uint64, q query.PageQuery) (map[string]any, error) {
	summary := map[string]any{}
	if err := r.inventoryAssetBase(ctx, tenantID, q).
		Select(`COUNT(DISTINCT COALESCE(NULLIF(ib.product_code, ''), 'ID:' || COALESCE(CAST(ib.product_id AS TEXT), CAST(ib.id AS TEXT)))) AS product_count,
			COALESCE(SUM(ib.remaining_quantity), 0) AS total_quantity,
			COALESCE(SUM(ib.remaining_quantity * ib.purchase_price), 0) AS total_value,
			COALESCE(SUM(CASE WHEN ib.remaining_quantity <= COALESCE(p.min_stock, 0) THEN 1 ELSE 0 END), 0) AS warning_count,
			COALESCE(SUM(CASE WHEN ib.remaining_quantity <= 0 THEN 1 ELSE 0 END), 0) AS out_of_stock_count`).
		Scan(&summary).Error; err != nil {
		return nil, err
	}
	slow, err := r.inventorySlowMoving(ctx, tenantID, q, 180)
	if err != nil {
		return nil, err
	}
	summary["slowMovingCount"] = len(slow)
	summary["avgTurnoverDays"] = 0
	return normalizeRow(summary), nil
}

func (r *ModuleRepository) inventoryAssetAging(ctx context.Context, tenantID uint64, q query.PageQuery) ([]map[string]any, error) {
	rows := make([]map[string]any, 0)
	ageDaysExpr := inventoryAgeDaysExpr("ib.purchase_date")
	err := r.inventoryAssetBase(ctx, tenantID, q).
		Select(fmt.Sprintf(`CASE
			WHEN %[1]s <= 30 THEN '30天以内'
			WHEN %[1]s <= 60 THEN '31-60天'
			WHEN %[1]s <= 90 THEN '61-90天'
			WHEN %[1]s <= 180 THEN '91-180天'
			WHEN %[1]s <= 365 THEN '181-365天'
			ELSE '365天以上'
		END AS age_bucket,
		COALESCE(SUM(ib.remaining_quantity), 0) AS quantity,
		COALESCE(SUM(ib.remaining_quantity * ib.purchase_price), 0) AS amount`, ageDaysExpr)).
		Group("age_bucket").Order("MIN(ib.purchase_date) DESC").Find(&rows).Error
	if err != nil {
		return nil, err
	}
	total := 0.0
	for _, row := range rows {
		total += numberValue(row["amount"])
	}
	for _, row := range rows {
		row["ratio"] = percentValue(numberValue(row["amount"]), total)
	}
	return normalizeRows(rows), nil
}

func (r *ModuleRepository) inventorySlowMoving(ctx context.Context, tenantID uint64, q query.PageQuery, days int) ([]map[string]any, error) {
	if days <= 0 {
		days = 180
	}
	rows := make([]map[string]any, 0)
	ageDaysExpr := inventoryAgeDaysExpr("ib.purchase_date")
	lastSalesAgeExpr := inventoryAgeDaysExpr("last_sales.last_sales_date")
	cutoffDate := time.Now().AddDate(0, 0, -days).Format("2006-01-02")
	err := r.inventoryAssetBase(ctx, tenantID, q).
		Where("(last_sales.last_sales_date IS NULL OR date(last_sales.last_sales_date) <= ?)", cutoffDate).
		Select(fmt.Sprintf(`ib.id,
			ib.product_id,
			ib.product_code,
			ib.product_name,
			COALESCE(p.brand, '') AS brand,
			COALESCE(p.category, '') AS category,
			ib.purchase_date,
			last_sales.last_sales_date,
			ib.remaining_quantity AS quantity,
			ib.remaining_quantity * ib.purchase_price AS inventory_amount,
			%[1]s AS age_days,
			CASE
				WHEN last_sales.last_sales_date IS NULL THEN '促销'
				WHEN %[2]s > 365 THEN '退货'
				ELSE '继续销售'
			END AS suggestion`, ageDaysExpr, lastSalesAgeExpr)).
		Order("age_days DESC").Limit(100).Find(&rows).Error
	return normalizeRows(rows), err
}

func (r *ModuleRepository) inventoryAssetTrend(ctx context.Context, tenantID uint64, q query.PageQuery) ([]map[string]any, error) {
	rows := make([]map[string]any, 0)
	err := r.inventoryAssetBase(ctx, tenantID, q).
		Select(`DATE(ib.purchase_date) AS date,
			COALESCE(SUM(ib.remaining_quantity), 0) AS quantity,
			COALESCE(SUM(ib.remaining_quantity * ib.purchase_price), 0) AS inventory_amount`).
		Group("DATE(ib.purchase_date)").Order("DATE(ib.purchase_date) ASC").Find(&rows).Error
	return normalizeRows(rows), err
}

func (r *ModuleRepository) inventoryDetailTab(ctx context.Context, tenantID uint64, data map[string]any) (map[string]any, error) {
	productCode, productID, err := r.inventoryProductIdentity(ctx, tenantID, data)
	if err != nil {
		return nil, err
	}
	page, pageSize := pageFromData(data)
	tab := firstNotBlank(stringValue(data["tab"]), "purchaseSources")
	var list []map[string]any
	var total int64
	switch tab {
	case "purchaseSources", "purchase_sources", "purchase":
		list, total, err = r.inventoryPurchaseSources(ctx, tenantID, productID, productCode, data, page, pageSize)
	case "salesTrace", "sales_trace", "sales":
		list, total, err = r.inventorySalesTrace(ctx, tenantID, productID, productCode, data, page, pageSize)
	case "inventoryMovements", "inventory_movements", "movements":
		list, total, err = r.inventoryMovements(ctx, tenantID, productID, productCode, data, page, pageSize)
	default:
		return nil, apperrors.New(40022, "不支持的库存详情类型", 400)
	}
	if err != nil {
		return nil, err
	}
	return map[string]any{"list": normalizeRows(list), "total": total, "page": page, "pageSize": pageSize}, nil
}

func (r *ModuleRepository) inventoryProductIdentity(ctx context.Context, tenantID uint64, data map[string]any) (string, uint64, error) {
	productCode := stringValue(data["product_code"])
	productID := uintValue(data["product_id"])
	if productCode != "" || productID > 0 {
		return productCode, productID, nil
	}
	id := uintValue(data["id"])
	if id == 0 {
		return "", 0, apperrors.New(40023, "请选择库存商品", 400)
	}
	row := map[string]any{}
	if err := r.db.WithContext(ctx).Table("inventory_stocks").Where("tenant_id = ? AND id = ? AND deleted_at IS NULL", tenantID, id).Take(&row).Error; err != nil {
		return "", 0, err
	}
	return stringValue(row["product_code"]), uintValue(row["product_id"]), nil
}

func (r *ModuleRepository) inventoryPurchaseSources(ctx context.Context, tenantID uint64, productID uint64, productCode string, data map[string]any, page int, pageSize int) ([]map[string]any, int64, error) {
	rows := make([]map[string]any, 0)
	db := r.db.WithContext(ctx).Table("inventory_batches").Where("tenant_id = ? AND deleted_at IS NULL", tenantID)
	db = productFilter(db, productID, productCode)
	if keyword := stringValue(data["keyword"]); keyword != "" {
		like := "%" + keyword + "%"
		db = db.Where("purchase_order_no LIKE ? OR supplier_name LIKE ? OR product_name LIKE ?", like, like, like)
	}
	if supplier := stringValue(data["supplier_name"]); supplier != "" {
		db = db.Where("supplier_name = ?", supplier)
	}
	if status := stringValue(data["status"]); status != "" {
		db = db.Where("status = ?", status)
	}
	db = applyDateRange(db, "purchase_date", data)
	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	order := safeDetailOrder(data, map[string]string{
		"purchaseDate": "purchase_date", "quantity": "inbound_quantity", "amount": "inbound_quantity * purchase_price",
		"inboundQuantity": "inbound_quantity", "remainingQuantity": "remaining_quantity", "purchasePrice": "purchase_price",
	}, "purchase_date")
	if err := db.Select(`id, batch_no, product_id, product_code, product_name, purchase_order_id, purchase_order_no,
		supplier_id, supplier_name, purchase_date, purchase_price, inbound_quantity,
		(inbound_quantity - remaining_quantity) AS sold_quantity, remaining_quantity, status`).
		Order(order).Offset((page - 1) * pageSize).Limit(pageSize).Find(&rows).Error; err != nil {
		return nil, 0, err
	}
	for _, row := range rows {
		row["status"] = batchStatus(numberValue(row["inbound_quantity"]), numberValue(row["remaining_quantity"]))
	}
	return rows, total, nil
}

func (r *ModuleRepository) inventorySalesTrace(ctx context.Context, tenantID uint64, productID uint64, productCode string, data map[string]any, page int, pageSize int) ([]map[string]any, int64, error) {
	rows := make([]map[string]any, 0)
	db := r.db.WithContext(ctx).Table("sales_cost_allocations AS sca").
		Joins("LEFT JOIN sales_orders AS so ON so.tenant_id = sca.tenant_id AND so.id = sca.sales_order_id").
		Joins("LEFT JOIN sales_order_items AS soi ON soi.tenant_id = sca.tenant_id AND soi.id = sca.sales_order_item_id AND soi.deleted_at IS NULL").
		Where("sca.tenant_id = ? AND sca.deleted_at IS NULL", tenantID)
	if productID > 0 {
		db = db.Where("sca.product_id = ?", productID)
	} else {
		db = db.Where("sca.product_code = ?", productCode)
	}
	if keyword := stringValue(data["keyword"]); keyword != "" {
		like := "%" + keyword + "%"
		db = db.Where("so.order_no LIKE ? OR so.customer_name LIKE ? OR sca.purchase_order_no LIKE ?", like, like, like)
	}
	if customer := stringValue(data["customer_name"]); customer != "" {
		db = db.Where("so.customer_name = ?", customer)
	}
	if purchaseNo := stringValue(data["purchase_order_no"]); purchaseNo != "" {
		db = db.Where("sca.purchase_order_no LIKE ?", "%"+purchaseNo+"%")
	}
	if salesNo := stringValue(data["sales_order_no"]); salesNo != "" {
		db = db.Where("so.order_no LIKE ?", "%"+salesNo+"%")
	}
	db = applyDateRange(db, "so.order_date", data)
	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	order := safeDetailOrder(data, map[string]string{
		"salesDate": "so.order_date", "orderDate": "so.order_date", "salesAmount": "(COALESCE(NULLIF(soi.price, 0), so.price, 0) * sca.quantity)", "profitAmount": "(COALESCE(NULLIF(soi.price, 0), so.price, 0) * sca.quantity - sca.cost_amount)", "quantity": "sca.quantity",
	}, "so.order_date")
	if err := db.Select(`sca.id, sca.inventory_batch_id, sca.sales_order_id, so.order_no AS sales_order_no,
		so.customer_name, so.order_date, sca.product_code, sca.product_name, sca.quantity,
		COALESCE(NULLIF(soi.price, 0), so.price, 0) AS sales_price,
		COALESCE(NULLIF(soi.price, 0), so.price, 0) * sca.quantity AS sales_amount,
		sca.cost_price, sca.cost_amount, COALESCE(NULLIF(soi.price, 0), so.price, 0) * sca.quantity - sca.cost_amount AS profit_amount,
		sca.purchase_order_id, sca.purchase_order_no, sca.supplier_name, sca.purchase_date, sca.purchase_price`).
		Order(order).Offset((page - 1) * pageSize).Limit(pageSize).Find(&rows).Error; err != nil {
		return nil, 0, err
	}
	return rows, total, nil
}

func (r *ModuleRepository) inventoryMovements(ctx context.Context, tenantID uint64, productID uint64, productCode string, data map[string]any, page int, pageSize int) ([]map[string]any, int64, error) {
	rows := make([]map[string]any, 0)
	db := r.db.WithContext(ctx).Table("inventory_movements AS im").
		Joins("LEFT JOIN users AS u ON u.id = im.created_by").
		Joins("LEFT JOIN purchase_orders AS po ON po.tenant_id = im.tenant_id AND po.id = im.source_id AND im.source_type IN ('purchase', 'PURCHASE_DELETE') AND po.deleted_at IS NULL").
		Joins("LEFT JOIN sales_orders AS so ON so.tenant_id = im.tenant_id AND so.id = im.source_id AND im.source_type IN ('sales', 'sales_return', 'ORDER_DELETE') AND so.deleted_at IS NULL").
		Joins("LEFT JOIN repair_orders AS ro ON ro.tenant_id = im.tenant_id AND ro.id = im.source_id AND im.source_type = 'repair' AND ro.deleted_at IS NULL").
		Joins("LEFT JOIN project_projects AS pp ON pp.tenant_id = im.tenant_id AND pp.id = im.source_id AND im.source_type = 'project' AND pp.deleted_at IS NULL").
		Where("im.tenant_id = ? AND im.deleted_at IS NULL", tenantID)
	if productID > 0 {
		db = db.Where("im.product_id = ?", productID)
	} else {
		db = db.Where("im.product_code = ?", productCode)
	}
	if keyword := stringValue(data["keyword"]); keyword != "" {
		like := "%" + keyword + "%"
		db = db.Where("im.movement_no LIKE ? OR im.product_name LIKE ? OR im.remark LIKE ?", like, like, like)
	}
	if sourceType := stringValue(data["source_type"]); sourceType != "" {
		db = db.Where("im.source_type = ?", sourceType)
	}
	if operatorName := stringValue(data["operator_name"]); operatorName != "" {
		db = db.Where("COALESCE(im.operator_name, u.real_name, u.username) LIKE ?", "%"+operatorName+"%")
	}
	db = applyDateRange(db, "im.occurred_at", data)
	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	order := safeDetailOrder(data, map[string]string{
		"occurredAt": "im.occurred_at", "quantity": "im.quantity", "amount": "im.amount",
	}, "im.occurred_at")
	if err := db.Select(`im.id, im.movement_no, im.product_id, im.product_code, im.product_name, im.source_type, im.source_id,
		im.direction, CASE WHEN im.direction = 'out' THEN -im.quantity ELSE im.quantity END AS quantity_change,
		im.quantity, im.before_quantity, im.after_quantity, im.unit_cost, im.amount, im.occurred_at,
		COALESCE(po.order_no, so.order_no, ro.order_no, pp.project_no) AS business_no,
		COALESCE(im.operator_name, u.real_name, u.username) AS operator_name, im.remark`).
		Order(order).Offset((page - 1) * pageSize).Limit(pageSize).Find(&rows).Error; err != nil {
		return nil, 0, err
	}
	return rows, total, nil
}

func productFilter(db *gorm.DB, productID uint64, productCode string) *gorm.DB {
	if productID > 0 {
		return db.Where("product_id = ?", productID)
	}
	return db.Where("product_code = ?", productCode)
}

func applyDateRange(db *gorm.DB, column string, data map[string]any) *gorm.DB {
	if start := stringValue(data["start_date"]); start != "" {
		db = db.Where(column+" >= ?", start)
	}
	if end := stringValue(data["end_date"]); end != "" {
		db = db.Where(column+" <= ?", end)
	}
	return db
}

func (r *ModuleRepository) profitSalesBase(ctx context.Context, tenantID uint64, q query.PageQuery) *gorm.DB {
	db := r.db.WithContext(ctx).Table("sales_cost_allocations AS sca").
		Joins("JOIN sales_orders AS so ON so.tenant_id = sca.tenant_id AND so.id = sca.sales_order_id AND so.deleted_at IS NULL").
		Joins("LEFT JOIN sales_order_items AS soi ON soi.tenant_id = sca.tenant_id AND soi.id = sca.sales_order_item_id AND soi.deleted_at IS NULL").
		Joins("LEFT JOIN products AS p ON p.tenant_id = sca.tenant_id AND p.deleted_at IS NULL AND (p.id = sca.product_id OR (sca.product_id IS NULL AND p.code = sca.product_code))").
		Where("sca.tenant_id = ? AND sca.deleted_at IS NULL", tenantID)
	db = applyQueryDateRange(db, "so.order_date", q)
	if keyword := strings.TrimSpace(q.Keyword); keyword != "" {
		like := "%" + keyword + "%"
		db = db.Where("so.order_no LIKE ? OR so.customer_name LIKE ? OR sca.product_code LIKE ? OR sca.product_name LIKE ? OR sca.purchase_order_no LIKE ?", like, like, like, like, like)
	}
	if q.CustomerID > 0 {
		db = db.Where("so.customer_id = ?", q.CustomerID)
	}
	if productCode := strings.TrimSpace(q.ProductCode); productCode != "" {
		db = db.Where("sca.product_code LIKE ?", "%"+productCode+"%")
	}
	if productName := strings.TrimSpace(q.ProductName); productName != "" {
		db = db.Where("sca.product_name LIKE ?", "%"+productName+"%")
	}
	return db
}

func (r *ModuleRepository) inventoryAssetBase(ctx context.Context, tenantID uint64, q query.PageQuery) *gorm.DB {
	db := r.db.WithContext(ctx).Table("inventory_batches AS ib").
		Joins("LEFT JOIN products AS p ON p.tenant_id = ib.tenant_id AND p.deleted_at IS NULL AND (p.id = ib.product_id OR (ib.product_id IS NULL AND p.code = ib.product_code))").
		Joins(`LEFT JOIN (
			SELECT sca.tenant_id, sca.inventory_batch_id, MAX(so.order_date) AS last_sales_date
			FROM sales_cost_allocations sca
			JOIN sales_orders so ON so.tenant_id = sca.tenant_id AND so.id = sca.sales_order_id AND so.deleted_at IS NULL
			WHERE sca.deleted_at IS NULL
			GROUP BY sca.tenant_id, sca.inventory_batch_id
		) AS last_sales ON last_sales.tenant_id = ib.tenant_id AND last_sales.inventory_batch_id = ib.id`).
		Where("ib.tenant_id = ? AND ib.deleted_at IS NULL AND ib.remaining_quantity > 0", tenantID)
	db = applyQueryDateRange(db, "ib.purchase_date", q)
	if keyword := strings.TrimSpace(q.Keyword); keyword != "" {
		like := "%" + keyword + "%"
		db = db.Where("ib.product_code LIKE ? OR ib.product_name LIKE ? OR ib.purchase_order_no LIKE ? OR ib.supplier_name LIKE ? OR p.brand LIKE ? OR p.category LIKE ?", like, like, like, like, like, like)
	}
	if productCode := strings.TrimSpace(q.ProductCode); productCode != "" {
		db = db.Where("ib.product_code LIKE ?", "%"+productCode+"%")
	}
	if productName := strings.TrimSpace(q.ProductName); productName != "" {
		db = db.Where("ib.product_name LIKE ?", "%"+productName+"%")
	}
	return db
}

func applyQueryDateRange(db *gorm.DB, column string, q query.PageQuery) *gorm.DB {
	if strings.TrimSpace(q.StartDate) != "" {
		db = db.Where(column+" >= ?", q.StartDate)
	}
	if strings.TrimSpace(q.EndDate) != "" {
		db = db.Where(column+" < ?", exclusiveEndDateArg(q.EndDate))
	}
	return db
}

func exclusiveEndDateArg(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return value
	}
	if t, err := time.Parse("2006-01-02", value); err == nil {
		return t.AddDate(0, 0, 1).Format("2006-01-02")
	}
	if t, err := time.Parse(time.RFC3339, value); err == nil {
		return t.AddDate(0, 0, 1).Format(time.RFC3339)
	}
	return value
}

func inventoryAgeDaysExpr(column string) string {
	return "CAST(julianday('now') - julianday(date(" + column + ")) AS INTEGER)"
}

func queryFromData(data map[string]any) query.PageQuery {
	page := int(numberValue(data["page"]))
	pageSize := int(numberValue(data["page_size"]))
	if page <= 0 {
		page = int(numberValue(data["page"]))
	}
	if pageSize <= 0 {
		pageSize = int(numberValue(data["pageSize"]))
	}
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 || pageSize > 200 {
		pageSize = 20
	}
	return query.PageQuery{
		Page:        page,
		PageSize:    pageSize,
		Keyword:     stringValue(data["keyword"]),
		SortBy:      stringValue(data["sort_by"]),
		Order:       stringValue(data["order"]),
		StartDate:   stringValue(data["start_date"]),
		EndDate:     stringValue(data["end_date"]),
		ProductCode: stringValue(data["product_code"]),
		ProductName: stringValue(data["product_name"]),
		CustomerID:  uintValue(data["customer_id"]),
	}
}

func reportOrder(q query.PageQuery, allowed map[string]string, fallback string) string {
	column := allowed[q.SortBy]
	if column == "" {
		column = allowed[toSnake(q.SortBy)]
	}
	if column == "" {
		column = fallback
	}
	if q.Order == "asc" {
		return column + " ASC"
	}
	return column + " DESC"
}

func percentValue(numerator float64, denominator float64) float64 {
	if denominator == 0 {
		return 0
	}
	return numerator * 100 / denominator
}

func pageFromData(data map[string]any) (int, int) {
	page := int(numberValue(data["page"]))
	pageSize := int(numberValue(data["page_size"]))
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}
	if pageSize > 100 {
		pageSize = 100
	}
	return page, pageSize
}

func safeDetailOrder(data map[string]any, allowed map[string]string, fallback string) string {
	sortBy := stringValue(data["sort_by"])
	column := allowed[sortBy]
	if column == "" {
		column = fallback
	}
	order := strings.ToLower(stringValue(data["order"]))
	if order != "asc" {
		order = "desc"
	}
	return column + " " + order
}

func (r *ModuleRepository) salesPurchaseSources(ctx context.Context, tenantID uint64, data map[string]any) (map[string]any, error) {
	productID := uintValue(data["product_id"])
	productCode := stringValue(data["product_code"])
	warehouse := firstNotBlank(stringValue(data["warehouse"]), "主仓库")
	if productID == 0 && productCode == "" {
		return map[string]any{"list": []map[string]any{}}, nil
	}
	rows := make([]map[string]any, 0)
	db := r.db.WithContext(ctx).Table("inventory_batches").
		Where("tenant_id = ? AND warehouse = ? AND remaining_quantity > 0 AND deleted_at IS NULL", tenantID, warehouse)
	if productID > 0 {
		db = db.Where("product_id = ?", productID)
	} else {
		db = db.Where("product_code = ?", productCode)
	}
	if err := db.Order("purchase_date ASC, id ASC").Find(&rows).Error; err != nil {
		return nil, err
	}
	for _, row := range rows {
		inbound := numberValue(row["inbound_quantity"])
		remain := numberValue(row["remaining_quantity"])
		row["sold_quantity"] = inbound - remain
		row["status"] = batchStatus(inbound, remain)
	}
	return map[string]any{"list": normalizeRows(rows)}, nil
}

func (r *ModuleRepository) salesAction(ctx context.Context, tenantID uint64, operatorID uint64, action string, data map[string]any) (map[string]any, error) {
	id := uintValue(data["id"])
	if id == 0 {
		id = uintValue(data["sales_order_id"])
	}
	if id == 0 {
		return nil, apperrors.New(40017, "请选择销售单", 400)
	}
	switch action {
	case "approve":
		if err := r.db.WithContext(ctx).Table("sales_orders").Where("tenant_id = ? AND id = ?", tenantID, id).Updates(map[string]any{
			"status": "已审核", "updated_by": operatorID, "updated_at": time.Now(),
		}).Error; err != nil {
			return nil, err
		}
	case "void":
		if err := r.db.WithContext(ctx).Table("sales_orders").Where("tenant_id = ? AND id = ?", tenantID, id).Updates(map[string]any{
			"status": "已作废", "updated_by": operatorID, "updated_at": time.Now(),
		}).Error; err != nil {
			return nil, err
		}
	case "return":
		if err := r.db.WithContext(ctx).Table("sales_orders").Where("tenant_id = ? AND id = ?", tenantID, id).Updates(map[string]any{
			"status": "已退货", "updated_by": operatorID, "updated_at": time.Now(),
		}).Error; err != nil {
			return nil, err
		}
	case "print":
		return map[string]any{"id": id, "printUrl": fmt.Sprintf("/print/sales/%d", id)}, nil
	default:
		return map[string]any{"id": id, "action": action, "accepted": true}, nil
	}
	item, err := r.Find(ctx, "sales", tenantID, id)
	if err != nil {
		return nil, err
	}
	return item, nil
}

func (r *ModuleRepository) CreatePhoto(ctx context.Context, module string, tenantID uint64, operatorID uint64, businessID uint64, photo map[string]any) (map[string]any, error) {
	spec, err := specOf(module)
	if err != nil {
		return nil, err
	}
	business, err := r.Find(ctx, module, tenantID, businessID)
	if err != nil {
		return nil, err
	}
	now := time.Now()
	row := normalizeInput(photo)
	row["tenant_id"] = tenantID
	row["module"] = module
	row["business_id"] = businessID
	row["business_no"] = businessNo(spec, business)
	row["created_by"] = operatorID
	row["updated_by"] = operatorID
	row["created_at"] = now
	row["updated_at"] = now
	row["taken_at"] = now
	if empty(row["scene"]) {
		row["scene"] = "general"
	}
	if err := r.db.WithContext(ctx).Table("business_photos").Create(row).Error; err != nil {
		return nil, err
	}
	return normalizeRow(row), nil
}

func (r *ModuleRepository) ListPhotos(ctx context.Context, module string, tenantID uint64, businessID uint64) ([]map[string]any, error) {
	if _, err := specOf(module); err != nil {
		return nil, err
	}
	rows := make([]map[string]any, 0)
	err := r.db.WithContext(ctx).Table("business_photos").
		Where("tenant_id = ? AND module = ? AND business_id = ? AND deleted_at IS NULL", tenantID, module, businessID).
		Order("created_at DESC").
		Find(&rows).Error
	return normalizeRows(rows), err
}

func businessNo(spec moduleSpec, row map[string]any) string {
	switch spec.NoField {
	case "order_no":
		return fmt.Sprint(row["orderNo"])
	case "project_no":
		return fmt.Sprint(row["projectNo"])
	case "record_no":
		return fmt.Sprint(row["recordNo"])
	case "receivable_no":
		return fmt.Sprint(row["receivableNo"])
	default:
		return ""
	}
}

func specOf(module string) (moduleSpec, error) {
	spec, ok := moduleSpecs[module]
	if !ok {
		return moduleSpec{}, fmt.Errorf("unsupported business module: %s", module)
	}
	return spec, nil
}

func (r *ModuleRepository) tableRow(table string, row map[string]any) (map[string]any, error) {
	columns, err := r.db.Migrator().ColumnTypes(table)
	if err != nil {
		return nil, err
	}
	allowed := make(map[string]bool, len(columns))
	for _, column := range columns {
		allowed[column.Name()] = true
	}
	filtered := make(map[string]any, len(row))
	for key, value := range row {
		if allowed[key] {
			filtered[key] = value
		}
	}
	return filtered, nil
}

func copyGeneratedID(target map[string]any, source map[string]any) {
	if id, ok := source["id"]; ok {
		target["id"] = id
	}
	if id, ok := source["@id"]; ok {
		target["@id"] = id
	}
}

func searchWhere(cols []string) string {
	parts := make([]string, 0, len(cols))
	for range cols {
		parts = append(parts, "%s LIKE ?")
	}
	for i, col := range cols {
		parts[i] = fmt.Sprintf(parts[i], col)
	}
	return "(" + strings.Join(parts, " OR ") + ")"
}

func searchArgs(cols []string, keyword string) []any {
	args := make([]any, 0, len(cols))
	like := "%" + keyword + "%"
	for range cols {
		args = append(args, like)
	}
	return args
}

func moduleOrder(q query.PageQuery) string {
	allowed := map[string]bool{"id": true, "created_at": true, "updated_at": true, "order_date": true, "registered_at": true, "occurred_at": true, "stock_time": true, "start_date": true, "end_date": true, "invoice_date": true, "due_date": true}
	sortBy := toSnake(q.SortBy)
	if !allowed[sortBy] {
		sortBy = "id"
	}
	if q.Order != "asc" {
		return sortBy + " DESC"
	}
	return sortBy + " ASC"
}

func inventoryOrder(q query.PageQuery) string {
	allowed := map[string]bool{"id": true, "product_code": true, "product_name": true, "warehouse": true, "quantity": true, "available_quantity": true, "occupied_quantity": true, "amount": true, "avg_cost": true, "stock_time": true, "updated_at": true}
	sortBy := toSnake(q.SortBy)
	if !allowed[sortBy] {
		sortBy = "id"
	}
	if q.Order != "asc" {
		return sortBy + " DESC"
	}
	return sortBy + " ASC"
}

func inventoryMovementOrder(q query.PageQuery) string {
	allowed := map[string]string{
		"id": "im.id", "occurred_at": "im.occurred_at", "occurredAt": "im.occurred_at",
		"quantity": "im.quantity", "quantity_change": "quantity_change", "quantityChange": "quantity_change",
	}
	sortBy := q.SortBy
	column := allowed[sortBy]
	if column == "" {
		column = allowed[toSnake(sortBy)]
	}
	if column == "" {
		column = "im.occurred_at"
	}
	if q.Order != "asc" {
		return column + " DESC, im.id DESC"
	}
	return column + " ASC, im.id ASC"
}

func inventorySourceType(value string) string {
	switch strings.TrimSpace(value) {
	case "订单删除", "ORDER_DELETE", "order_delete":
		return "ORDER_DELETE"
	case "采购入库", "purchase", "in_purchase":
		return "purchase"
	case "销售出库", "sales", "sale", "out_sales":
		return "sales"
	case "维修领料", "repair":
		return "repair"
	case "工程领料", "project":
		return "project"
	case "销售退货", "sales_return":
		return "sales_return"
	case "采购退货", "purchase_return":
		return "purchase_return"
	case "库存盘点", "stocktake":
		return "stocktake"
	case "库存调整", "adjust":
		return "adjust"
	default:
		return strings.TrimSpace(value)
	}
}

func normalizeInput(data map[string]any) map[string]any {
	row := make(map[string]any, len(data))
	for key, value := range data {
		row[toSnake(key)] = value
	}
	return row
}

func normalizeRows(rows []map[string]any) []map[string]any {
	for i := range rows {
		rows[i] = normalizeRow(rows[i])
	}
	return rows
}

func normalizeRow(row map[string]any) map[string]any {
	out := make(map[string]any, len(row))
	for key, value := range row {
		if key == "@id" {
			key = "id"
		}
		out[toCamel(key)] = value
	}
	return out
}

func toSnake(value string) string {
	var builder strings.Builder
	for i, r := range value {
		if r >= 'A' && r <= 'Z' {
			if i > 0 {
				builder.WriteByte('_')
			}
			builder.WriteRune(r + ('a' - 'A'))
			continue
		}
		builder.WriteRune(r)
	}
	return builder.String()
}

func toCamel(value string) string {
	parts := strings.Split(value, "_")
	for i := 1; i < len(parts); i++ {
		if parts[i] == "" {
			continue
		}
		parts[i] = strings.ToUpper(parts[i][:1]) + parts[i][1:]
	}
	return strings.Join(parts, "")
}

func empty(value any) bool {
	if value == nil {
		return true
	}
	if text, ok := value.(string); ok {
		return strings.TrimSpace(text) == ""
	}
	return false
}

func numberValue(value any) float64 {
	switch item := value.(type) {
	case nil:
		return 0
	case int:
		return float64(item)
	case int64:
		return float64(item)
	case uint64:
		return float64(item)
	case float64:
		return item
	case float32:
		return float64(item)
	case string:
		parsed, _ := strconv.ParseFloat(strings.TrimSpace(item), 64)
		return parsed
	default:
		parsed, _ := strconv.ParseFloat(fmt.Sprint(item), 64)
		return parsed
	}
}

func stringValue(value any) string {
	if value == nil {
		return ""
	}
	return strings.TrimSpace(fmt.Sprint(value))
}

func uintSliceValue(value any) []uint64 {
	result := make([]uint64, 0)
	switch items := value.(type) {
	case []uint64:
		return items
	case []int:
		for _, item := range items {
			if item > 0 {
				result = append(result, uint64(item))
			}
		}
	case []any:
		for _, item := range items {
			if id := uintValue(item); id > 0 {
				result = append(result, id)
			}
		}
	}
	return result
}

func mapRowsValue(value any) []map[string]any {
	rows := make([]map[string]any, 0)
	switch items := value.(type) {
	case []map[string]any:
		return items
	case []any:
		for _, item := range items {
			if row, ok := item.(map[string]any); ok {
				rows = append(rows, row)
			}
		}
	}
	return rows
}

func repairInputRows(row map[string]any, key string) []map[string]any {
	rows := mapRowsValue(row[key])
	if len(rows) == 0 {
		rows = mapRowsValue(row[toCamel(key)])
	}
	return rows
}

func firstPresent(row map[string]any, keys ...string) any {
	for _, key := range keys {
		if value, ok := row[key]; ok && !empty(value) {
			return value
		}
	}
	return nil
}

func repairInputString(row map[string]any, keys ...string) string {
	return stringValue(firstPresent(row, keys...))
}

func repairInputQuantity(row map[string]any) float64 {
	qty := numberValue(firstPresent(row, "quantity", "qty"))
	if qty <= 0 {
		return 1
	}
	return qty
}

func repairInputAmount(row map[string]any) float64 {
	amount := numberValue(firstPresent(row, "amount", "cost_amount", "costAmount", "total_amount", "totalAmount"))
	if amount > 0 {
		return amount
	}
	return repairInputQuantity(row) * numberValue(firstPresent(row, "price", "unit_price", "unitPrice", "cost_price", "costPrice"))
}

func repairInputPrice(row map[string]any, amount float64, qty float64) float64 {
	price := numberValue(firstPresent(row, "price", "unit_price", "unitPrice", "cost_price", "costPrice"))
	if price > 0 {
		return price
	}
	if qty > 0 {
		return amount / qty
	}
	return amount
}

func sumRepairInputAmount(rows []map[string]any) float64 {
	total := 0.0
	for _, row := range rows {
		total += repairInputAmount(row)
	}
	return total
}

func sumRepairParts(rows []map[string]any) (float64, float64) {
	amount := 0.0
	cost := 0.0
	for _, row := range rows {
		qty := repairInputQuantity(row)
		partAmount := repairInputAmount(row)
		amount += partAmount
		partCost := numberValue(firstPresent(row, "cost_amount", "costAmount"))
		if partCost == 0 {
			partCost = qty * numberValue(firstPresent(row, "cost_price", "costPrice", "purchase_price", "purchasePrice"))
		}
		cost += partCost
	}
	return amount, cost
}

func sumRepairInputAmountByType(rows []map[string]any, costType string) float64 {
	total := 0.0
	for _, row := range rows {
		rowType := firstNotBlank(repairInputString(row, "cost_type", "costType", "item_type", "itemType"), "其他费用")
		if rowType == costType {
			total += repairInputAmount(row)
		}
	}
	return total
}

func repairCostItemType(costType string) string {
	switch costType {
	case "人工成本":
		return "labor"
	case "运输费用":
		return "transport"
	case "外协费用":
		return "outsource"
	default:
		return "other_cost"
	}
}

func selectedPurchaseBatchIDs(row map[string]any) []uint64 {
	ids := uintSliceValue(row["inventory_batch_ids"])
	if len(ids) == 0 {
		if id := uintValue(row["inventory_batch_id"]); id > 0 {
			ids = append(ids, id)
		}
	}
	return compactUintIDs(ids)
}

func compactUintIDs(ids []uint64) []uint64 {
	result := make([]uint64, 0, len(ids))
	seen := map[uint64]bool{}
	for _, id := range ids {
		if id == 0 || seen[id] {
			continue
		}
		seen[id] = true
		result = append(result, id)
	}
	return result
}

func statementSettlementStatus(received float64, unpaid float64) string {
	if unpaid <= 0 {
		return "paid"
	}
	if received > 0 {
		return "partial"
	}
	return "unpaid"
}

func statementSaleAmounts(sale map[string]any) (float64, float64, float64) {
	total := numberValue(sale["total_amount"])
	if total <= 0 {
		total = numberValue(sale["quantity"]) * numberValue(sale["price"])
	}
	received := numberValue(sale["received_amount"])
	if received < 0 {
		received = 0
	}
	unpaid := numberValue(sale["receivable_amount"])
	computedUnpaid := total - received
	if computedUnpaid < 0 {
		computedUnpaid = 0
	}
	if unpaid <= 0 && computedUnpaid > 0 {
		unpaid = computedUnpaid
	}
	if unpaid < 0 {
		unpaid = 0
	}
	return total, received, unpaid
}

func (r *ModuleRepository) statementReceivableProductName(ctx context.Context, tenantID uint64, receivable map[string]any) string {
	sourceType := stringValue(receivable["source_type"])
	sourceID := uintValue(receivable["source_id"])
	if sourceID == 0 {
		return stringValue(receivable["source_no"])
	}
	row := map[string]any{}
	switch sourceType {
	case "sales":
		_ = r.db.WithContext(ctx).Table("sales_orders").Select("product_name").Where("tenant_id = ? AND id = ?", tenantID, sourceID).Take(&row).Error
		return firstNotBlank(stringValue(row["product_name"]), stringValue(receivable["source_no"]))
	case "repair":
		_ = r.db.WithContext(ctx).Table("repair_orders").Select("device_name").Where("tenant_id = ? AND id = ?", tenantID, sourceID).Take(&row).Error
		return firstNotBlank(stringValue(row["device_name"]), stringValue(receivable["source_no"]))
	default:
		return stringValue(receivable["source_no"])
	}
}

func (r *ModuleRepository) statementReceivableQuantity(ctx context.Context, tenantID uint64, receivable map[string]any) float64 {
	if stringValue(receivable["source_type"]) != "sales" {
		return 1
	}
	row := map[string]any{}
	if err := r.db.WithContext(ctx).Table("sales_orders").Select("quantity").Where("tenant_id = ? AND id = ?", tenantID, uintValue(receivable["source_id"])).Take(&row).Error; err != nil {
		return 1
	}
	qty := numberValue(row["quantity"])
	if qty <= 0 {
		return 1
	}
	return qty
}

func (r *ModuleRepository) syncStatementSaleReceivable(ctx context.Context, tenantID uint64, operatorID uint64, customer map[string]any, sale map[string]any, total float64, received float64, unpaid float64) error {
	saleID := mapID(sale)
	if saleID == 0 {
		return nil
	}
	now := time.Now()
	customerID := uintValue(customer["id"])
	customerName := firstNotBlank(stringValue(customer["name"]), stringValue(sale["customer_name"]))
	updates := map[string]any{
		"total_amount":      total,
		"received_amount":   received,
		"receivable_amount": unpaid,
		"updated_by":        operatorID,
		"updated_at":        now,
	}
	if customerID > 0 && uintValue(sale["customer_id"]) == 0 {
		updates["customer_id"] = customerID
	}
	if err := r.db.WithContext(ctx).Table("sales_orders").Where("tenant_id = ? AND id = ?", tenantID, saleID).Updates(updates).Error; err != nil {
		return err
	}
	if unpaid <= 0 {
		return nil
	}
	receivable := map[string]any{}
	err := r.db.WithContext(ctx).Table("receivables").
		Where("tenant_id = ? AND source_type = ? AND deleted_at IS NULL", tenantID, "sales").
		Where("source_id = ? OR source_no = ?", saleID, stringValue(sale["order_no"])).
		Order("id ASC").Take(&receivable).Error
	if err == nil {
		status := receivableStatus(unpaid, timeValue(receivable["due_date"], now))
		if unpaid > 0 && received > 0 {
			status = "partial"
		}
		return r.db.WithContext(ctx).Table("receivables").Where("id = ?", mapID(receivable)).Updates(map[string]any{
			"customer_id":     nilIfZero(customerID),
			"customer_name":   customerName,
			"source_id":       saleID,
			"source_no":       sale["order_no"],
			"total_amount":    total,
			"received_amount": received,
			"balance_amount":  unpaid,
			"settlement_mode": firstNotBlank(stringValue(customer["payment_method"]), "credit"),
			"status":          status,
			"updated_by":      operatorID,
			"updated_at":      now,
		}).Error
	}
	if err != gorm.ErrRecordNotFound {
		return err
	}
	row := map[string]any{
		"order_date":    sale["order_date"],
		"order_no":      sale["order_no"],
		"customer_name": customerName,
	}
	return r.createReceivable(ctx, tenantID, operatorID, customer, saleID, row, total, received, unpaid)
}

func boolValue(value any) bool {
	switch item := value.(type) {
	case bool:
		return item
	case int:
		return item != 0
	case int64:
		return item != 0
	case uint64:
		return item != 0
	case float64:
		return item != 0
	case string:
		text := strings.ToLower(strings.TrimSpace(item))
		return text == "true" || text == "1" || text == "yes"
	default:
		return false
	}
}

func timeValue(value any, fallback time.Time) time.Time {
	switch item := value.(type) {
	case time.Time:
		return item
	case string:
		text := strings.TrimSpace(item)
		for _, layout := range []string{time.RFC3339Nano, time.RFC3339, "2006-01-02 15:04:05", "2006-01-02"} {
			if parsed, err := time.Parse(layout, text); err == nil {
				return parsed
			}
		}
	}
	return fallback
}

func firstNotBlank(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}

func nilIfZero(value uint64) any {
	if value == 0 {
		return nil
	}
	return value
}

func nilIfZeroTime(value time.Time) any {
	if value.IsZero() {
		return nil
	}
	return value
}

func mapID(row map[string]any) uint64 {
	if id := uintValue(row["id"]); id > 0 {
		return id
	}
	return uintValue(row["@id"])
}

func uintValue(value any) uint64 {
	switch item := value.(type) {
	case uint64:
		return item
	case int64:
		return uint64(item)
	case int:
		return uint64(item)
	case float64:
		return uint64(item)
	case string:
		parsed, _ := strconv.ParseUint(item, 10, 64)
		return parsed
	default:
		parsed, _ := strconv.ParseUint(fmt.Sprint(item), 10, 64)
		return parsed
	}
}

func nextNo(module string) string {
	prefixes := map[string]string{
		"sales": "SO", "purchase": "PO", "inventory": "ST", "batch": "BAT", "repair": "RO", "project": "PRJ", "finance": "FIN", "receivables": "AR", "customer-statements": "CS", "payables": "AP",
	}
	prefix := prefixes[module]
	if prefix == "" {
		prefix = "BIZ"
	}
	return fmt.Sprintf("%s%s%s", prefix, time.Now().Format("20060102150405"), strings.ToUpper(uuid.NewString()[:6]))
}

func nextProductCode() string {
	return fmt.Sprintf("PRD%s%s", time.Now().Format("20060102150405"), strings.ToUpper(uuid.NewString()[:4]))
}

func financeAccountCode(name string) string {
	switch strings.TrimSpace(name) {
	case "现金":
		return "CASH"
	case "微信":
		return "WECHAT"
	case "支付宝":
		return "ALIPAY"
	case "银行卡":
		return "BANK"
	default:
		return "ACC" + time.Now().Format("20060102150405")
	}
}

func financeAccountType(name string) string {
	switch strings.TrimSpace(name) {
	case "现金":
		return "cash"
	case "微信":
		return "wechat"
	case "支付宝":
		return "alipay"
	case "银行卡":
		return "bank"
	default:
		return "other"
	}
}
