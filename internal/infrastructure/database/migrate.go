package database

import (
	"context"
	stderrors "errors"
	"time"

	"erp/internal/config"
	authmodel "erp/internal/domain/auth/model"
	businessmodel "erp/internal/domain/business/model"
	customermodel "erp/internal/domain/customer/model"
	dashboardmodel "erp/internal/domain/dashboard/model"
	suppliermodel "erp/internal/domain/supplier/model"
	systemmodel "erp/internal/domain/system/model"
	"erp/internal/shared/utils"
	"gorm.io/gorm"
)

func AutoMigrate(db *gorm.DB, cfg config.AdminConfig) error {
	if err := db.AutoMigrate(
		&authmodel.User{}, &authmodel.Role{}, &authmodel.Permission{}, &authmodel.Menu{},
		&authmodel.UserRole{}, &authmodel.RoleMenu{}, &authmodel.RolePermission{},
		&authmodel.LoginLog{}, &authmodel.OperationLog{},
		&customermodel.Customer{}, &customermodel.CustomerContact{}, &customermodel.CustomerAddress{},
		&suppliermodel.Supplier{},
		&businessmodel.Product{},
		&businessmodel.SalesOrder{}, &businessmodel.SalesOrderItem{}, &businessmodel.SalesCostAllocation{}, &businessmodel.SalesOrderDeletion{},
		&businessmodel.DocumentDeleteRecord{}, &businessmodel.DocumentDeleteDetail{},
		&businessmodel.PurchaseOrder{}, &businessmodel.PurchaseOrderItem{}, &businessmodel.InventoryBatch{},
		&businessmodel.InventoryStock{}, &businessmodel.InventoryMovement{}, &businessmodel.RepairOrder{}, &businessmodel.RepairOrderItem{},
		&businessmodel.ProjectProject{}, &businessmodel.FinanceAccount{}, &businessmodel.FinanceRecord{}, &businessmodel.Receivable{}, &businessmodel.CustomerStatement{}, &businessmodel.CustomerStatementItem{}, &businessmodel.Payable{}, &businessmodel.BusinessPhoto{},
		&dashboardmodel.EngineerCheckIn{},
		&systemmodel.SystemSetting{},
	); err != nil {
		return err
	}
	if err := applyIndexes(db); err != nil {
		return err
	}
	if err := applyMissingColumns(db); err != nil {
		return err
	}
	if err := applyDataFixes(db); err != nil {
		return err
	}
	return seedAdmin(db, cfg)
}

func applyIndexes(db *gorm.DB) error {
	statements := []string{
		"CREATE UNIQUE INDEX IF NOT EXISTS ux_users_tenant_username ON users(tenant_id, username)",
		"CREATE UNIQUE INDEX IF NOT EXISTS ux_roles_tenant_code ON roles(tenant_id, code)",
		"CREATE UNIQUE INDEX IF NOT EXISTS ux_permissions_tenant_code ON permissions(tenant_id, code)",
		"CREATE UNIQUE INDEX IF NOT EXISTS ux_settings_tenant_key ON system_settings(tenant_id, setting_key)",
		"CREATE UNIQUE INDEX IF NOT EXISTS ux_customers_tenant_code ON customers(tenant_id, code)",
		"CREATE UNIQUE INDEX IF NOT EXISTS ux_suppliers_tenant_code ON suppliers(tenant_id, code)",
		"CREATE UNIQUE INDEX IF NOT EXISTS ux_products_tenant_code ON products(tenant_id, code)",
		"CREATE INDEX IF NOT EXISTS idx_products_search ON products(tenant_id, name, code, barcode, brand, category)",
		"CREATE INDEX IF NOT EXISTS idx_suppliers_search ON suppliers(tenant_id, name, contact_name, phone, tax_no)",
		"CREATE UNIQUE INDEX IF NOT EXISTS ux_sales_orders_tenant_no ON sales_orders(tenant_id, order_no)",
		"CREATE UNIQUE INDEX IF NOT EXISTS ux_purchase_orders_tenant_no ON purchase_orders(tenant_id, order_no)",
		"CREATE UNIQUE INDEX IF NOT EXISTS ux_repair_orders_tenant_no ON repair_orders(tenant_id, order_no)",
		"CREATE UNIQUE INDEX IF NOT EXISTS ux_project_projects_tenant_no ON project_projects(tenant_id, project_no)",
		"CREATE UNIQUE INDEX IF NOT EXISTS ux_finance_records_tenant_no ON finance_records(tenant_id, record_no)",
		"CREATE UNIQUE INDEX IF NOT EXISTS ux_receivables_tenant_no ON receivables(tenant_id, receivable_no)",
		"CREATE UNIQUE INDEX IF NOT EXISTS ux_customer_statements_tenant_no ON customer_statements(tenant_id, statement_no)",
		"CREATE UNIQUE INDEX IF NOT EXISTS ux_payables_tenant_no ON payables(tenant_id, payable_no)",
		"CREATE UNIQUE INDEX IF NOT EXISTS ux_finance_accounts_tenant_code ON finance_accounts(tenant_id, code)",
		"CREATE INDEX IF NOT EXISTS idx_sales_orders_customer_date ON sales_orders(tenant_id, customer_id, order_date)",
		"CREATE INDEX IF NOT EXISTS idx_sales_orders_profit_rate ON sales_orders(tenant_id, order_date, profit_rate)",
		"CREATE INDEX IF NOT EXISTS idx_purchase_orders_supplier_date ON purchase_orders(tenant_id, supplier_id, order_date)",
		"CREATE INDEX IF NOT EXISTS idx_purchase_items_order ON purchase_order_items(tenant_id, purchase_order_id)",
		"CREATE INDEX IF NOT EXISTS idx_inventory_batches_product_remaining ON inventory_batches(tenant_id, product_id, warehouse, remaining_quantity, purchase_date)",
		"CREATE INDEX IF NOT EXISTS idx_inventory_batches_purchase ON inventory_batches(tenant_id, purchase_order_id, purchase_order_item_id)",
		"CREATE INDEX IF NOT EXISTS idx_inventory_batches_report ON inventory_batches(tenant_id, product_code, remaining_quantity, purchase_date)",
		"CREATE INDEX IF NOT EXISTS idx_sales_items_order ON sales_order_items(tenant_id, sales_order_id)",
		"CREATE INDEX IF NOT EXISTS idx_sales_items_purchase ON sales_order_items(tenant_id, purchase_order_id, purchase_order_item_id, inventory_batch_id)",
		"CREATE INDEX IF NOT EXISTS idx_sales_cost_allocations_order ON sales_cost_allocations(tenant_id, sales_order_id, sales_order_item_id)",
		"CREATE INDEX IF NOT EXISTS idx_sales_cost_allocations_purchase ON sales_cost_allocations(tenant_id, purchase_order_id, inventory_batch_id)",
		"CREATE INDEX IF NOT EXISTS idx_sales_cost_allocations_report ON sales_cost_allocations(tenant_id, product_code, inventory_batch_id, deleted_at)",
		"CREATE UNIQUE INDEX IF NOT EXISTS ux_sales_order_deletions_order ON sales_order_deletions(tenant_id, original_order_id)",
		"CREATE INDEX IF NOT EXISTS idx_sales_order_deletions_time ON sales_order_deletions(tenant_id, deleted_at)",
		"CREATE UNIQUE INDEX IF NOT EXISTS ux_document_delete_records_doc ON document_delete_records(tenant_id, document_type, document_id)",
		"CREATE INDEX IF NOT EXISTS idx_document_delete_records_no ON document_delete_records(tenant_id, document_no)",
		"CREATE INDEX IF NOT EXISTS idx_document_delete_records_time ON document_delete_records(tenant_id, delete_time)",
		"CREATE INDEX IF NOT EXISTS idx_document_delete_records_user ON document_delete_records(tenant_id, delete_user_id)",
		"CREATE INDEX IF NOT EXISTS idx_document_delete_records_status ON document_delete_records(tenant_id, delete_status)",
		"CREATE INDEX IF NOT EXISTS idx_document_delete_details_record ON document_delete_details(tenant_id, record_id, detail_type)",
		"CREATE INDEX IF NOT EXISTS idx_inventory_stocks_product ON inventory_stocks(tenant_id, product_id, warehouse)",
		"CREATE INDEX IF NOT EXISTS idx_inventory_stocks_time ON inventory_stocks(tenant_id, stock_time)",
		"CREATE INDEX IF NOT EXISTS idx_inventory_movements_source ON inventory_movements(tenant_id, source_type, occurred_at)",
		"CREATE INDEX IF NOT EXISTS idx_inventory_movements_product_code ON inventory_movements(tenant_id, product_code)",
		"CREATE INDEX IF NOT EXISTS idx_inventory_movements_product_time ON inventory_movements(tenant_id, product_code, occurred_at)",
		"CREATE INDEX IF NOT EXISTS idx_repair_orders_status ON repair_orders(tenant_id, repair_status, registered_at)",
		"CREATE INDEX IF NOT EXISTS idx_repair_orders_profit ON repair_orders(tenant_id, registered_at, profit_amount)",
		"CREATE INDEX IF NOT EXISTS idx_repair_items_order ON repair_order_items(tenant_id, repair_order_id, item_type)",
		"CREATE INDEX IF NOT EXISTS idx_repair_items_batch ON repair_order_items(tenant_id, inventory_batch_id)",
		"CREATE INDEX IF NOT EXISTS idx_project_projects_status ON project_projects(tenant_id, status, progress)",
		"CREATE INDEX IF NOT EXISTS idx_project_projects_dates ON project_projects(tenant_id, start_date, end_date)",
		"CREATE INDEX IF NOT EXISTS idx_finance_records_type_time ON finance_records(tenant_id, record_type, occurred_at)",
		"CREATE INDEX IF NOT EXISTS idx_finance_records_source ON finance_records(tenant_id, source_type, source_id)",
		"CREATE INDEX IF NOT EXISTS idx_finance_accounts_type ON finance_accounts(tenant_id, account_type, status)",
		"CREATE INDEX IF NOT EXISTS idx_receivables_customer_due ON receivables(tenant_id, customer_id, due_date, status)",
		"CREATE INDEX IF NOT EXISTS idx_receivables_source ON receivables(tenant_id, source_type, source_no)",
		"CREATE INDEX IF NOT EXISTS idx_customer_statements_customer_period ON customer_statements(tenant_id, customer_id, start_date, end_date, status)",
		"DROP INDEX IF EXISTS ux_customer_statement_items_sale_open",
		"CREATE INDEX IF NOT EXISTS idx_customer_statement_items_statement ON customer_statement_items(tenant_id, statement_id)",
		"CREATE INDEX IF NOT EXISTS idx_customer_statement_items_receivable ON customer_statement_items(tenant_id, receivable_id)",
		"CREATE UNIQUE INDEX IF NOT EXISTS ux_customer_statement_items_source_open ON customer_statement_items(tenant_id, source_type, source_id) WHERE deleted_at IS NULL AND source_type IS NOT NULL AND source_id IS NOT NULL",
		"CREATE INDEX IF NOT EXISTS idx_payables_supplier_due ON payables(tenant_id, supplier_id, due_date, status)",
		"CREATE INDEX IF NOT EXISTS idx_payables_source ON payables(tenant_id, source_type, source_no)",
		"CREATE INDEX IF NOT EXISTS idx_business_photos_biz ON business_photos(tenant_id, module, business_id, scene)",
		"CREATE INDEX IF NOT EXISTS idx_engineer_check_ins_user_time ON engineer_check_ins(tenant_id, user_id, check_in_at)",
		"CREATE INDEX IF NOT EXISTS idx_engineer_check_ins_created ON engineer_check_ins(tenant_id, created_at)",
		"CREATE INDEX IF NOT EXISTS idx_customers_search ON customers(tenant_id, name, phone, tax_no)",
		"CREATE INDEX IF NOT EXISTS idx_role_permissions_role ON role_permissions(role_id)",
		"CREATE INDEX IF NOT EXISTS idx_role_permissions_permission ON role_permissions(permission_id)",
		"CREATE INDEX IF NOT EXISTS idx_role_menus_role ON role_menus(role_id)",
		"CREATE INDEX IF NOT EXISTS idx_role_menus_menu ON role_menus(menu_id)",
		"CREATE INDEX IF NOT EXISTS idx_user_roles_user ON user_roles(user_id)",
		"CREATE INDEX IF NOT EXISTS idx_user_roles_role ON user_roles(role_id)",
		"CREATE INDEX IF NOT EXISTS idx_roles_parent ON roles(parent_id)",
		"CREATE INDEX IF NOT EXISTS idx_roles_data_scope ON roles(data_scope)",
		"CREATE INDEX IF NOT EXISTS idx_permissions_type ON permissions(type)",
		"CREATE INDEX IF NOT EXISTS idx_operation_logs_user_time ON operation_logs(user_id, created_at)",
		"CREATE INDEX IF NOT EXISTS idx_login_logs_user_time ON login_logs(user_id, created_at)",
	}
	for _, statement := range statements {
		if err := db.Exec(statement).Error; err != nil {
			return err
		}
	}
	return nil
}

func applyMissingColumns(db *gorm.DB) error {
	if db.Migrator().HasTable("repair_order_items") && !db.Migrator().HasColumn("repair_order_items", "inventory_batch_id") {
		if err := db.Exec("ALTER TABLE repair_order_items ADD COLUMN inventory_batch_id integer").Error; err != nil {
			return err
		}
	}
	return nil
}

func batchTimestampExpr() string {
	return "strftime('%Y%m%d%H%M%S', 'now', 'localtime')"
}

func applyDataFixes(db *gorm.DB) error {
	if err := dropProductPriceColumns(db); err != nil {
		return err
	}
	if err := seedFinanceAccounts(db); err != nil {
		return err
	}
	if err := repairFinanceAccounts(db); err != nil {
		return err
	}
	statements := []string{
		"UPDATE sales_orders SET profit_rate = CASE WHEN total_amount > 0 THEN profit_amount * 100.0 / total_amount ELSE 0 END WHERE total_amount > 0 AND (profit_rate IS NULL OR profit_rate = 0)",
		`INSERT INTO purchase_order_items (
			tenant_id, purchase_order_id, product_id, product_code, product_name, quantity, price, amount, warehouse,
			created_by, updated_by, created_at, updated_at
		)
		SELECT po.tenant_id, po.id, po.product_id, po.product_code, po.product_name, po.quantity, po.price, po.total_amount, '主仓库',
			po.created_by, po.updated_by, po.created_at, po.updated_at
		FROM purchase_orders po
		WHERE po.deleted_at IS NULL
			AND po.product_name <> ''
			AND NOT EXISTS (
				SELECT 1 FROM purchase_order_items poi
				WHERE poi.tenant_id = po.tenant_id AND poi.purchase_order_id = po.id AND poi.deleted_at IS NULL
			)`,
		`INSERT INTO inventory_batches (
			tenant_id, batch_no, product_id, product_code, product_name, warehouse,
			purchase_order_id, purchase_order_item_id, purchase_order_no,
			supplier_id, supplier_name, purchase_date, purchase_price,
			inbound_quantity, remaining_quantity, status,
			created_by, updated_by, created_at, updated_at
		)
		SELECT po.tenant_id, 'BAT' || ` + batchTimestampExpr() + ` || po.id, po.product_id, po.product_code, po.product_name, '主仓库',
			po.id, poi.id, po.order_no,
			po.supplier_id, po.supplier_name, po.order_date, po.price,
			po.quantity, po.quantity, 'available',
			po.created_by, po.updated_by, po.created_at, po.updated_at
		FROM purchase_orders po
		JOIN purchase_order_items poi ON poi.tenant_id = po.tenant_id AND poi.purchase_order_id = po.id AND poi.deleted_at IS NULL
		WHERE po.deleted_at IS NULL
			AND po.product_name <> ''
			AND NOT EXISTS (
				SELECT 1 FROM inventory_batches ib
				WHERE ib.tenant_id = po.tenant_id AND ib.purchase_order_id = po.id AND ib.deleted_at IS NULL
			)`,
		`UPDATE inventory_batches
		SET status = CASE
			WHEN remaining_quantity <= 0 THEN '已销售完成'
			WHEN inbound_quantity > 0 AND remaining_quantity < inbound_quantity THEN '销售中'
			ELSE '未销售'
		END
		WHERE deleted_at IS NULL`,
		`UPDATE inventory_stocks
		SET quantity = (
				SELECT SUM(s2.quantity)
				FROM inventory_stocks s2
				WHERE s2.tenant_id = inventory_stocks.tenant_id
					AND s2.warehouse = inventory_stocks.warehouse
					AND COALESCE(NULLIF(s2.product_code, ''), 'ID:' || COALESCE(CAST(s2.product_id AS TEXT), CAST(s2.id AS TEXT))) =
						COALESCE(NULLIF(inventory_stocks.product_code, ''), 'ID:' || COALESCE(CAST(inventory_stocks.product_id AS TEXT), CAST(inventory_stocks.id AS TEXT)))
					AND s2.deleted_at IS NULL
			),
			amount = (
				SELECT SUM(s2.amount)
				FROM inventory_stocks s2
				WHERE s2.tenant_id = inventory_stocks.tenant_id
					AND s2.warehouse = inventory_stocks.warehouse
					AND COALESCE(NULLIF(s2.product_code, ''), 'ID:' || COALESCE(CAST(s2.product_id AS TEXT), CAST(s2.id AS TEXT))) =
						COALESCE(NULLIF(inventory_stocks.product_code, ''), 'ID:' || COALESCE(CAST(inventory_stocks.product_id AS TEXT), CAST(inventory_stocks.id AS TEXT)))
					AND s2.deleted_at IS NULL
			),
			avg_cost = CASE
				WHEN (
					SELECT SUM(s2.quantity)
					FROM inventory_stocks s2
					WHERE s2.tenant_id = inventory_stocks.tenant_id
						AND s2.warehouse = inventory_stocks.warehouse
						AND COALESCE(NULLIF(s2.product_code, ''), 'ID:' || COALESCE(CAST(s2.product_id AS TEXT), CAST(s2.id AS TEXT))) =
							COALESCE(NULLIF(inventory_stocks.product_code, ''), 'ID:' || COALESCE(CAST(inventory_stocks.product_id AS TEXT), CAST(inventory_stocks.id AS TEXT)))
						AND s2.deleted_at IS NULL
				) > 0 THEN (
					SELECT SUM(s2.amount) / SUM(s2.quantity)
					FROM inventory_stocks s2
					WHERE s2.tenant_id = inventory_stocks.tenant_id
						AND s2.warehouse = inventory_stocks.warehouse
						AND COALESCE(NULLIF(s2.product_code, ''), 'ID:' || COALESCE(CAST(s2.product_id AS TEXT), CAST(s2.id AS TEXT))) =
							COALESCE(NULLIF(inventory_stocks.product_code, ''), 'ID:' || COALESCE(CAST(inventory_stocks.product_id AS TEXT), CAST(inventory_stocks.id AS TEXT)))
						AND s2.deleted_at IS NULL
				)
				ELSE 0
			END,
			stock_time = (
				SELECT MAX(s2.stock_time)
				FROM inventory_stocks s2
				WHERE s2.tenant_id = inventory_stocks.tenant_id
					AND s2.warehouse = inventory_stocks.warehouse
					AND COALESCE(NULLIF(s2.product_code, ''), 'ID:' || COALESCE(CAST(s2.product_id AS TEXT), CAST(s2.id AS TEXT))) =
						COALESCE(NULLIF(inventory_stocks.product_code, ''), 'ID:' || COALESCE(CAST(inventory_stocks.product_id AS TEXT), CAST(inventory_stocks.id AS TEXT)))
					AND s2.deleted_at IS NULL
			),
			updated_at = CURRENT_TIMESTAMP
		WHERE deleted_at IS NULL
			AND id IN (
				SELECT MIN(id)
				FROM inventory_stocks
				WHERE deleted_at IS NULL
				GROUP BY tenant_id, warehouse, COALESCE(NULLIF(product_code, ''), 'ID:' || COALESCE(CAST(product_id AS TEXT), CAST(id AS TEXT)))
			)`,
		`UPDATE inventory_stocks
		SET deleted_at = CURRENT_TIMESTAMP, updated_at = CURRENT_TIMESTAMP
		WHERE deleted_at IS NULL
			AND id NOT IN (
				SELECT MIN(id)
				FROM inventory_stocks
				WHERE deleted_at IS NULL
				GROUP BY tenant_id, warehouse, COALESCE(NULLIF(product_code, ''), 'ID:' || COALESCE(CAST(product_id AS TEXT), CAST(id AS TEXT)))
			)`,
		"UPDATE sales_orders SET cost_method = 'purchase_source' WHERE cost_method IN ('fifo', 'moving_average', 'moving_avg', 'specified_batch', 'batch', '') OR cost_method IS NULL",
		"UPDATE sales_order_items SET purchase_source = 'purchase_order' WHERE purchase_source IN ('fifo', 'moving_average', 'moving_avg', 'specified_batch', 'batch', '') OR purchase_source IS NULL",
	}
	for _, statement := range statements {
		if err := db.Exec(statement).Error; err != nil {
			return err
		}
	}
	return nil
}

func seedFinanceAccounts(db *gorm.DB) error {
	now := time.Now()
	accounts := []map[string]any{
		{"code": "CASH", "name": "现金", "account_type": "cash"},
		{"code": "WECHAT", "name": "微信", "account_type": "wechat"},
		{"code": "ALIPAY", "name": "支付宝", "account_type": "alipay"},
		{"code": "BANK", "name": "银行卡", "account_type": "bank"},
	}
	for _, account := range accounts {
		var count int64
		if err := db.Table("finance_accounts").Where("tenant_id = ? AND code = ? AND deleted_at IS NULL", 1, account["code"]).Count(&count).Error; err != nil {
			return err
		}
		if count > 0 {
			if err := db.Table("finance_accounts").Where("tenant_id = ? AND code = ? AND deleted_at IS NULL", 1, account["code"]).Updates(map[string]any{
				"name":         account["name"],
				"account_type": account["account_type"],
				"updated_at":   now,
			}).Error; err != nil {
				return err
			}
			continue
		}
		account["tenant_id"] = 1
		account["opening_balance"] = 0
		account["balance"] = 0
		account["status"] = "active"
		account["created_at"] = now
		account["updated_at"] = now
		if err := db.Table("finance_accounts").Create(account).Error; err != nil {
			return err
		}
	}
	return nil
}

func repairFinanceAccounts(db *gorm.DB) error {
	now := time.Now()
	var cashID uint64
	if err := db.Table("finance_accounts").
		Select("id").
		Where("tenant_id = ? AND code = ? AND deleted_at IS NULL", 1, "CASH").
		Scan(&cashID).Error; err != nil {
		return err
	}
	if cashID == 0 {
		return nil
	}
	badAccountWhere := "tenant_id = ? AND code LIKE ? AND name IN (?, ?) AND deleted_at IS NULL"
	var badBalance float64
	if err := db.Table("finance_accounts").
		Select("COALESCE(SUM(balance), 0)").
		Where(badAccountWhere, 1, "ACC%", "??", "").
		Scan(&badBalance).Error; err != nil {
		return err
	}
	if badBalance != 0 {
		if err := db.Table("finance_accounts").
			Where("id = ?", cashID).
			Updates(map[string]any{
				"balance":    gorm.Expr("balance + ?", badBalance),
				"updated_at": now,
			}).Error; err != nil {
			return err
		}
	}
	if err := db.Exec(`
		UPDATE finance_records
		SET account_id = ?, account_name = ?, updated_at = ?
		WHERE tenant_id = ?
		  AND deleted_at IS NULL
		  AND (
		    account_name IN (?, ?)
		    OR account_id IN (
		      SELECT id FROM finance_accounts
		      WHERE tenant_id = ? AND code LIKE ? AND name IN (?, ?) AND deleted_at IS NULL
		    )
		  )
	`, cashID, "现金", now, 1, "??", "", 1, "ACC%", "??", "").Error; err != nil {
		return err
	}
	return db.Table("finance_accounts").
		Where(badAccountWhere, 1, "ACC%", "??", "").
		Updates(map[string]any{
			"deleted_at": now,
			"updated_at": now,
		}).Error
}

func dropProductPriceColumns(db *gorm.DB) error {
	if !db.Migrator().HasTable("products") {
		return nil
	}
	for _, column := range []string{"purchase_price", "sale_price", "cost_price", "minimum_price", "wholesale_price"} {
		if db.Migrator().HasColumn("products", column) {
			if err := db.Migrator().DropColumn("products", column); err != nil {
				return err
			}
		}
	}
	return nil
}

func seedAdmin(db *gorm.DB, cfg config.AdminConfig) error {
	return db.Transaction(func(tx *gorm.DB) error {
		ctx := context.Background()
		var role authmodel.Role
		err := tx.WithContext(ctx).Where("tenant_id = ? AND code = ?", 1, "super_admin").First(&role).Error
		if err != nil {
			if !stderrors.Is(err, gorm.ErrRecordNotFound) {
				return err
			}
			role = authmodel.Role{
				Code: "super_admin", Name: "超级管理员", Status: "active", Sort: 1,
			}
			role.TenantID = 1
			if err := tx.WithContext(ctx).Create(&role).Error; err != nil {
				return err
			}
		}

		permissions := defaultPermissions()
		for i := range permissions {
			var existing authmodel.Permission
			err := tx.WithContext(ctx).Where("tenant_id = ? AND code = ?", 1, permissions[i].Code).First(&existing).Error
			if err != nil {
				if !stderrors.Is(err, gorm.ErrRecordNotFound) {
					return err
				}
				permissions[i].TenantID = 1
				if err := tx.WithContext(ctx).Create(&permissions[i]).Error; err != nil {
					return err
				}
				existing = permissions[i]
			}
			var count int64
			tx.Model(&authmodel.RolePermission{}).Where("role_id = ? AND permission_id = ?", role.ID, existing.ID).Count(&count)
			if count == 0 {
				if err := tx.Create(&authmodel.RolePermission{RoleID: role.ID, PermissionID: existing.ID, CreatedAt: time.Now()}).Error; err != nil {
					return err
				}
			}
		}

		menus := defaultMenus()
		for i := range menus {
			var existing authmodel.Menu
			err := tx.WithContext(ctx).Where("tenant_id = ? AND name = ?", 1, menus[i].Name).First(&existing).Error
			if err != nil {
				if !stderrors.Is(err, gorm.ErrRecordNotFound) {
					return err
				}
				menus[i].TenantID = 1
				if err := tx.WithContext(ctx).Create(&menus[i]).Error; err != nil {
					return err
				}
				existing = menus[i]
			}
			var count int64
			tx.Model(&authmodel.RoleMenu{}).Where("role_id = ? AND menu_id = ?", role.ID, existing.ID).Count(&count)
			if count == 0 {
				if err := tx.Create(&authmodel.RoleMenu{RoleID: role.ID, MenuID: existing.ID, CreatedAt: time.Now()}).Error; err != nil {
					return err
				}
			}
		}
		if err := tx.WithContext(ctx).Model(&authmodel.Menu{}).
			Where("tenant_id = ? AND name = ?", 1, "inventory-movements").
			Update("visible", false).Error; err != nil {
			return err
		}
		if err := tx.WithContext(ctx).Exec(`
INSERT INTO menus (tenant_id, name, title, path, component, icon, type, permission_code, sort, visible, status, created_at, updated_at)
SELECT 1, 'document-delete-records', '单据删除记录', '/document-delete-records', 'mobile/FeaturePage', 'Document', 'menu', 'system.audit.view', 165, true, 'active', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP
WHERE NOT EXISTS (SELECT 1 FROM menus WHERE tenant_id = 1 AND name = 'document-delete-records')
`).Error; err != nil {
			return err
		}

		var user authmodel.User
		err = tx.WithContext(ctx).Where("tenant_id = ? AND username = ?", 1, cfg.Username).First(&user).Error
		if err != nil {
			if !stderrors.Is(err, gorm.ErrRecordNotFound) {
				return err
			}
			hash, err := utils.HashPassword(cfg.Password)
			if err != nil {
				return err
			}
			user = authmodel.User{
				Username: cfg.Username, PasswordHash: hash, RealName: cfg.RealName, Status: "active",
			}
			user.TenantID = 1
			if err := tx.WithContext(ctx).Create(&user).Error; err != nil {
				return err
			}
		} else if usesDefaultAdminPassword(user.PasswordHash) && cfg.Password != "" {
			now := time.Now()
			hash, err := utils.HashPassword(cfg.Password)
			if err != nil {
				return err
			}
			if err := tx.WithContext(ctx).Model(&authmodel.User{}).
				Where("id = ?", user.ID).
				Updates(map[string]any{
					"password_hash":        hash,
					"must_change_password": false,
					"password_changed_at":  now,
					"password_version":     gorm.Expr("password_version + 1"),
					"updated_at":           now,
				}).Error; err != nil {
				return err
			}
			user.PasswordHash = hash
		}
		var userRoleCount int64
		tx.Model(&authmodel.UserRole{}).Where("user_id = ? AND role_id = ?", user.ID, role.ID).Count(&userRoleCount)
		if userRoleCount == 0 {
			if err := tx.Create(&authmodel.UserRole{UserID: user.ID, RoleID: role.ID, CreatedAt: time.Now()}).Error; err != nil {
				return err
			}
		}

		settings := []systemmodel.SystemSetting{
			{GroupName: "company", SettingKey: "company.name", SettingValue: "ERP Pro 管理系统", ValueType: "string", IsPublic: true},
			{GroupName: "security", SettingKey: "password.min_length", SettingValue: "8", ValueType: "int", IsPublic: false},
		}
		for i := range settings {
			var count int64
			tx.Model(&systemmodel.SystemSetting{}).Where("tenant_id = ? AND setting_key = ?", 1, settings[i].SettingKey).Count(&count)
			if count == 0 {
				settings[i].TenantID = 1
				if err := tx.Create(&settings[i]).Error; err != nil {
					return err
				}
			}
		}
		return nil
	})
}

func usesDefaultAdminPassword(hash string) bool {
	for _, password := range []string{"admin888", "admin123", "change-this-admin-password"} {
		if utils.CheckPassword(hash, password) {
			return true
		}
	}
	return false
}

func defaultPermissions() []authmodel.Permission {
	items := []struct {
		Code   string
		Name   string
		Module string
		Type   string
		Method string
		Path   string
	}{
		{"*", "全部权限", "system", "api", "", ""},
		{"auth.user.manage", "用户管理", "auth", "api", "", "/api/v1/users"},
		{"auth.user.reset_password", "重置用户密码", "auth", "button", "PUT", "/api/v1/users/:id/password"},
		{"auth.role.manage", "角色管理", "auth", "api", "", "/api/v1/roles"},
		{"auth.permission.manage", "权限管理", "auth", "api", "", "/api/v1/permissions"},
		{"auth.menu.manage", "菜单管理", "auth", "api", "", "/api/v1/menus"},
		{"system.setting.manage", "系统配置", "system", "api", "", "/api/v1/settings"},
		{"system.audit.view", "审计日志", "system", "api", "GET", "/api/v1/audit"},
		{"customer.manage", "客户管理", "customer", "api", "", "/api/v1/customers"},
		{"customer.create", "新增客户按钮", "customer", "button", "POST", "/api/v1/customers"},
		{"customer.edit", "编辑客户按钮", "customer", "button", "PUT", "/api/v1/customers/:id"},
		{"customer.delete", "删除客户按钮", "customer", "button", "DELETE", "/api/v1/customers/:id"},
		{"dashboard.boss.view", "老板驾驶舱", "dashboard", "api", "GET", "/api/v1/dashboard/boss"},
		{"product.manage", "商品管理", "product", "api", "", "/api/v1/products"},
	}
	result := make([]authmodel.Permission, 0, len(items))
	for _, item := range items {
		result = append(result, authmodel.Permission{
			Code: item.Code, Name: item.Name, Module: item.Module, Type: item.Type, Method: item.Method, Path: item.Path, Status: "active",
		})
	}
	result = append(result, defaultBusinessPermissions()...)
	return result
}

func defaultBusinessPermissions() []authmodel.Permission {
	return []authmodel.Permission{
		{Code: "sales.manage", Name: "销售管理", Module: "sales", Type: "api", Path: "/api/v1/sales", Status: "active"},
		{Code: "product.manage", Name: "商品管理", Module: "product", Type: "api", Path: "/api/v1/products", Status: "active"},
		{Code: "supplier.manage", Name: "供应商管理", Module: "supplier", Type: "api", Path: "/api/v1/suppliers", Status: "active"},
		{Code: "purchase.manage", Name: "采购管理", Module: "purchase", Type: "api", Path: "/api/v1/purchase", Status: "active"},
		{Code: "inventory.manage", Name: "库存管理", Module: "inventory", Type: "api", Path: "/api/v1/inventory", Status: "active"},
		{Code: "repair.manage", Name: "维修管理", Module: "repair", Type: "api", Path: "/api/v1/repair", Status: "active"},
		{Code: "project.manage", Name: "工程管理", Module: "project", Type: "api", Path: "/api/v1/project", Status: "active"},
		{Code: "finance.manage", Name: "财务管理", Module: "finance", Type: "api", Path: "/api/v1/finance", Status: "active"},
		{Code: "document_delete", Name: "单据删除", Module: "document", Type: "api", Method: "DELETE", Path: "/api/v1/:module/:id", Status: "active"},
	}
}

func defaultMenus() []authmodel.Menu {
	return []authmodel.Menu{
		{Name: "boss-dashboard", Title: "老板驾驶舱", Path: "/dashboard", Component: "mobile/Dashboard", Icon: "DataLine", Type: "menu", PermissionCode: "dashboard.boss.view", Sort: 1, Visible: true, Status: "active"},
		{Name: "system", Title: "系统管理", Path: "/system", Component: "Layout", Icon: "Setting", Type: "directory", PermissionCode: "", Sort: 10, Visible: true, Status: "active"},
		{Name: "users", Title: "用户管理", Path: "/system/users", Component: "system/users/index", Icon: "User", Type: "menu", PermissionCode: "auth.user.manage", Sort: 10, Visible: true, Status: "active"},
		{Name: "roles", Title: "角色管理", Path: "/system/roles", Component: "system/roles/index", Icon: "UserFilled", Type: "menu", PermissionCode: "auth.role.manage", Sort: 20, Visible: true, Status: "active"},
		{Name: "permissions", Title: "权限管理", Path: "/system/permissions", Component: "system/permissions/index", Icon: "Lock", Type: "menu", PermissionCode: "auth.permission.manage", Sort: 30, Visible: true, Status: "active"},
		{Name: "menus", Title: "菜单管理", Path: "/system/menus", Component: "system/menus/index", Icon: "Menu", Type: "menu", PermissionCode: "auth.menu.manage", Sort: 40, Visible: true, Status: "active"},
		{Name: "settings", Title: "系统配置", Path: "/system/settings", Component: "system/settings/index", Icon: "Tools", Type: "menu", PermissionCode: "system.setting.manage", Sort: 50, Visible: true, Status: "active"},
		{Name: "audit", Title: "审计日志", Path: "/system/audit", Component: "system/audit/index", Icon: "Document", Type: "menu", PermissionCode: "system.audit.view", Sort: 60, Visible: true, Status: "active"},
		{Name: "customers", Title: "客户管理", Path: "/customers", Component: "customer/index", Icon: "OfficeBuilding", Type: "menu", PermissionCode: "customer.manage", Sort: 100, Visible: true, Status: "active"},
		{Name: "suppliers", Title: "供应商管理", Path: "/suppliers", Component: "supplier/index", Icon: "Box", Type: "menu", PermissionCode: "supplier.manage", Sort: 110, Visible: true, Status: "active"},
		{Name: "products", Title: "商品管理", Path: "/products", Component: "product/index", Icon: "Goods", Type: "menu", PermissionCode: "product.manage", Sort: 120, Visible: true, Status: "active"},
		{Name: "sales", Title: "销售管理", Path: "/sales", Component: "mobile/FeaturePage", Icon: "Sell", Type: "menu", PermissionCode: "sales.manage", Sort: 130, Visible: true, Status: "active"},
		{Name: "purchase", Title: "采购管理", Path: "/purchase", Component: "mobile/FeaturePage", Icon: "ShoppingCart", Type: "menu", PermissionCode: "purchase.manage", Sort: 140, Visible: true, Status: "active"},
		{Name: "inventory", Title: "库存查询", Path: "/inventory", Component: "mobile/FeaturePage", Icon: "Box", Type: "menu", PermissionCode: "inventory.manage", Sort: 150, Visible: true, Status: "active"},
		{Name: "inventory-movements", Title: "库存流水", Path: "/inventory-movements", Component: "mobile/FeaturePage", Icon: "Document", Type: "menu", PermissionCode: "inventory.manage", Sort: 160, Visible: false, Status: "active"},
		{Name: "repair", Title: "维修管理", Path: "/repair", Component: "mobile/FeaturePage", Icon: "Tools", Type: "menu", PermissionCode: "repair.manage", Sort: 170, Visible: true, Status: "active"},
		{Name: "project", Title: "工程管理", Path: "/project", Component: "mobile/FeaturePage", Icon: "Briefcase", Type: "menu", PermissionCode: "project.manage", Sort: 180, Visible: true, Status: "active"},
	}
}
