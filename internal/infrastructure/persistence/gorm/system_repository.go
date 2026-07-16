package gormrepo

import (
	"context"
	"time"

	systemmodel "erp/internal/domain/system/model"
	"erp/internal/shared/query"
	"gorm.io/gorm"
)

type SettingRepository struct {
	db *gorm.DB
}

func NewSettingRepository(db *gorm.DB) *SettingRepository {
	return &SettingRepository{db: db}
}

func (r *SettingRepository) FindByID(ctx context.Context, tenantID uint64, id uint64) (*systemmodel.SystemSetting, error) {
	var setting systemmodel.SystemSetting
	err := r.db.WithContext(ctx).Where("tenant_id = ? AND id = ?", tenantID, id).First(&setting).Error
	return &setting, err
}

func (r *SettingRepository) FindByKey(ctx context.Context, tenantID uint64, key string) (*systemmodel.SystemSetting, error) {
	var setting systemmodel.SystemSetting
	err := r.db.WithContext(ctx).Where("tenant_id = ? AND setting_key = ?", tenantID, key).First(&setting).Error
	return &setting, err
}

func (r *SettingRepository) List(ctx context.Context, tenantID uint64, q query.PageQuery) ([]systemmodel.SystemSetting, int64, error) {
	var settings []systemmodel.SystemSetting
	var total int64
	db := r.db.WithContext(ctx).Model(&systemmodel.SystemSetting{}).Where("tenant_id = ?", tenantID)
	if q.Keyword != "" {
		like := "%" + q.Keyword + "%"
		db = db.Where("group_name LIKE ? OR setting_key LIKE ?", like, like)
	}
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	err := db.Order(safeOrder(q)).Offset(q.Offset()).Limit(q.PageSize).Find(&settings).Error
	return settings, total, err
}

func (r *SettingRepository) Create(ctx context.Context, setting *systemmodel.SystemSetting) error {
	return r.db.WithContext(ctx).Create(setting).Error
}

func (r *SettingRepository) Update(ctx context.Context, setting *systemmodel.SystemSetting) error {
	return r.db.WithContext(ctx).Save(setting).Error
}

func (r *SettingRepository) Delete(ctx context.Context, tenantID uint64, id uint64) error {
	return r.db.WithContext(ctx).Where("tenant_id = ? AND id = ?", tenantID, id).Delete(&systemmodel.SystemSetting{}).Error
}

func (r *SettingRepository) RestoreTestData(ctx context.Context, tenantID uint64, operatorID uint64) (map[string]any, error) {
	result := map[string]any{}
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var loginLogCount int64
		if err := tx.Table("login_logs").Where("tenant_id = ?", tenantID).Count(&loginLogCount).Error; err != nil {
			return err
		}
		var operationLogCount int64
		if err := tx.Table("operation_logs").Where("tenant_id = ?", tenantID).Count(&operationLogCount).Error; err != nil {
			return err
		}
		accountLoginResult := tx.Table("users").
			Where("tenant_id = ? AND (last_login_at IS NOT NULL OR COALESCE(last_login_ip, '') <> '')", tenantID).
			Updates(map[string]any{"last_login_at": nil, "last_login_ip": ""})
		if accountLoginResult.Error != nil {
			return accountLoginResult.Error
		}
		tables := []string{
			"login_logs", "operation_logs",
			"business_photos",
			"document_delete_details", "document_delete_records", "sales_order_deletions",
			"customer_statement_items", "customer_statements",
			"receivables", "payables", "finance_records", "finance_accounts",
			"inventory_movements", "inventory_stocks", "inventory_batches",
			"sales_cost_allocations", "sales_order_items", "sales_orders",
			"purchase_order_items", "purchase_orders",
			"repair_order_items", "repair_orders",
			"engineer_check_ins",
			"project_projects",
			"products",
			"customer_contacts", "customer_addresses", "customers",
			"suppliers",
		}
		if err := deleteTables(tx, tables); err != nil {
			return err
		}
		now := time.Now()
		customers := restoreCustomers(tenantID)
		suppliers := restoreSuppliers(tenantID)
		products := restoreProducts(tenantID)
		accounts := restoreFinanceAccounts(tenantID)
		if err := seedRows(tx, "customers", customers, operatorID, now); err != nil {
			return err
		}
		if err := seedRows(tx, "suppliers", suppliers, operatorID, now); err != nil {
			return err
		}
		if err := seedRows(tx, "products", products, operatorID, now); err != nil {
			return err
		}
		if err := seedRows(tx, "finance_accounts", accounts, operatorID, now); err != nil {
			return err
		}
		result = map[string]any{
			"customers": len(customers), "suppliers": len(suppliers), "products": len(products), "financeAccounts": len(accounts),
			"loginLogs": loginLogCount, "operationLogs": operationLogCount, "accountLoginLogs": accountLoginResult.RowsAffected,
		}
		return nil
	})
	return result, err
}

func seedRows(tx *gorm.DB, table string, rows []map[string]any, operatorID uint64, now time.Time) error {
	for _, row := range rows {
		row["created_by"] = operatorID
		row["updated_by"] = operatorID
		row["created_at"] = now
		row["updated_at"] = now
		if err := tx.Table(table).Create(row).Error; err != nil {
			return err
		}
	}
	return nil
}

func restoreCustomers(tenantID uint64) []map[string]any {
	return []map[string]any{
		{"tenant_id": tenantID, "code": "KH-001", "name": "杭州星河科技有限公司", "type": "company", "level": "A", "phone": "13800010001", "billing_cycle": "monthly", "credit_days": 30, "credit_limit": 200000, "status": "active"},
		{"tenant_id": tenantID, "code": "KH-002", "name": "宁波蓝海工程有限公司", "type": "company", "level": "B", "phone": "13800010002", "billing_cycle": "quarterly", "credit_days": 90, "credit_limit": 300000, "status": "active"},
		{"tenant_id": tenantID, "code": "KH-003", "name": "绍兴智联办公设备商行", "type": "company", "level": "B", "phone": "13800010003", "billing_cycle": "monthly", "credit_days": 30, "credit_limit": 100000, "status": "active"},
		{"tenant_id": tenantID, "code": "KH-004", "name": "温州启明学校", "type": "company", "level": "A", "phone": "13800010004", "billing_cycle": "yearly", "credit_days": 365, "credit_limit": 500000, "status": "active"},
	}
}

func restoreSuppliers(tenantID uint64) []map[string]any {
	return []map[string]any{
		{"tenant_id": tenantID, "code": "GYS-001", "name": "杭州联创电脑科技有限公司", "contact_name": "周经理", "phone": "13900020001", "supplier_types": "商品供应商", "status": "active"},
		{"tenant_id": tenantID, "code": "GYS-002", "name": "刘师傅维修服务部", "contact_name": "刘师傅", "phone": "13900020002", "supplier_types": "外协服务商", "status": "active"},
		{"tenant_id": tenantID, "code": "GYS-003", "name": "张师傅网络调试工作室", "contact_name": "张师傅", "phone": "13900020003", "supplier_types": "外协服务商", "status": "active"},
		{"tenant_id": tenantID, "code": "GYS-004", "name": "浙江安讯弱电工程队", "contact_name": "孙工", "phone": "13900020004", "supplier_types": "工程施工方、综合供应商", "status": "active"},
	}
}

func restoreProducts(tenantID uint64) []map[string]any {
	return []map[string]any{
		{"tenant_id": tenantID, "code": "SP-001", "name": "联想商用台式电脑 M460", "category": "电脑整机", "brand": "联想", "spec": "i5/16G/512G", "unit": "台", "status": "active", "min_stock": 5},
		{"tenant_id": tenantID, "code": "SP-002", "name": "惠普黑白激光打印机 P1108", "category": "办公设备", "brand": "惠普", "spec": "A4 黑白", "unit": "台", "status": "active", "min_stock": 3},
		{"tenant_id": tenantID, "code": "SP-003", "name": "海康威视千兆交换机 8口", "category": "网络设备", "brand": "海康威视", "spec": "8口千兆", "unit": "台", "status": "active", "min_stock": 10},
		{"tenant_id": tenantID, "code": "SP-004", "name": "金士顿 16G DDR4 内存", "category": "维修配件", "brand": "金士顿", "spec": "DDR4 16G", "unit": "条", "status": "active", "min_stock": 20},
		{"tenant_id": tenantID, "code": "SP-005", "name": "西部数据 1TB 固态硬盘", "category": "维修配件", "brand": "西部数据", "spec": "SATA 1TB", "unit": "块", "status": "active", "min_stock": 20},
	}
}

func restoreFinanceAccounts(tenantID uint64) []map[string]any {
	return []map[string]any{
		{"tenant_id": tenantID, "code": "CASH", "name": "现金", "account_type": "cash", "opening_balance": 10000, "balance": 10000, "status": "active"},
		{"tenant_id": tenantID, "code": "WECHAT", "name": "微信", "account_type": "wechat", "opening_balance": 20000, "balance": 20000, "status": "active"},
		{"tenant_id": tenantID, "code": "ALIPAY", "name": "支付宝", "account_type": "alipay", "opening_balance": 15000, "balance": 15000, "status": "active"},
		{"tenant_id": tenantID, "code": "ICBC", "name": "工商银行", "account_type": "bank", "opening_balance": 100000, "balance": 100000, "status": "active"},
	}
}

func deleteTables(tx *gorm.DB, tables []string) error {
	for _, table := range tables {
		if err := tx.Exec("DELETE FROM " + table).Error; err != nil {
			return err
		}
	}
	return resetSQLiteSequences(tx, tables)
}

func resetSQLiteSequences(tx *gorm.DB, tables []string) error {
	var exists int64
	if err := tx.Raw("SELECT COUNT(1) FROM sqlite_master WHERE type = 'table' AND name = 'sqlite_sequence'").Scan(&exists).Error; err != nil {
		return err
	}
	if exists == 0 {
		return nil
	}
	return tx.Exec("DELETE FROM sqlite_sequence WHERE name IN ?", tables).Error
}
