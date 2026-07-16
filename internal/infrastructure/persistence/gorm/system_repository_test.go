package gormrepo

import (
	"context"
	"testing"
	"time"

	authmodel "erp/internal/domain/auth/model"
	businessmodel "erp/internal/domain/business/model"
	customermodel "erp/internal/domain/customer/model"
	dashboardmodel "erp/internal/domain/dashboard/model"
	suppliermodel "erp/internal/domain/supplier/model"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func TestRestoreTestDataClearsSQLiteTablesAndSeedsDemoRows(t *testing.T) {
	db := newRestoreTestDB(t)
	repo := NewSettingRepository(db)
	now := time.Now()

	lastLogin := now.Add(-time.Hour)
	if err := db.Table("users").Create(map[string]any{
		"id": 9, "tenant_id": 1, "username": "demo", "password_hash": "hash", "status": "active",
		"last_login_at": lastLogin, "last_login_ip": "127.0.0.1", "created_at": now, "updated_at": now,
	}).Error; err != nil {
		t.Fatalf("seed user: %v", err)
	}
	if err := db.Table("login_logs").Create(map[string]any{
		"tenant_id": 1, "username": "demo", "success": true, "created_at": now,
	}).Error; err != nil {
		t.Fatalf("seed login log: %v", err)
	}
	if err := db.Table("operation_logs").Create(map[string]any{
		"tenant_id": 1, "username": "demo", "module": "settings", "action": "restore", "method": "POST", "path": "/settings/restore-test-data", "created_at": now,
	}).Error; err != nil {
		t.Fatalf("seed operation log: %v", err)
	}
	if err := db.Table("sales_orders").Create(map[string]any{
		"tenant_id": 1, "order_no": "SO-OLD", "customer_name": "old customer", "quantity": 1, "price": 1, "order_date": now, "status": "draft", "created_at": now, "updated_at": now,
	}).Error; err != nil {
		t.Fatalf("seed sales order: %v", err)
	}
	if err := db.Table("finance_accounts").Create(map[string]any{
		"tenant_id": 1, "code": "OLD", "name": "old account", "account_type": "cash", "opening_balance": 1, "balance": 1, "status": "active", "created_at": now, "updated_at": now,
	}).Error; err != nil {
		t.Fatalf("seed finance account: %v", err)
	}

	result, err := repo.RestoreTestData(context.Background(), 1, 9)
	if err != nil {
		t.Fatalf("restore test data: %v", err)
	}
	if result["customers"] != 4 || result["suppliers"] != 4 || result["products"] != 5 || result["financeAccounts"] != 4 {
		t.Fatalf("unexpected restore result: %#v", result)
	}
	if result["loginLogs"] != int64(1) || result["operationLogs"] != int64(1) || result["accountLoginLogs"] != int64(1) {
		t.Fatalf("unexpected cleared counts: %#v", result)
	}

	assertTableCount(t, db, "login_logs", 0)
	assertTableCount(t, db, "operation_logs", 0)
	assertTableCount(t, db, "sales_orders", 0)
	assertTableCount(t, db, "customers", 4)
	assertTableCount(t, db, "suppliers", 4)
	assertTableCount(t, db, "products", 5)
	assertTableCount(t, db, "finance_accounts", 4)

	user := map[string]any{}
	if err := db.Table("users").Where("id = ?", 9).Take(&user).Error; err != nil {
		t.Fatalf("load user: %v", err)
	}
	if user["last_login_at"] != nil || stringValue(user["last_login_ip"]) != "" {
		t.Fatalf("login info was not reset: %#v", user)
	}

	firstCustomer := map[string]any{}
	if err := db.Table("customers").Order("id").Take(&firstCustomer).Error; err != nil {
		t.Fatalf("load first customer: %v", err)
	}
	if uintValue(firstCustomer["id"]) != 1 {
		t.Fatalf("customer sequence was not reset: %#v", firstCustomer)
	}
	if stringValue(firstCustomer["name"]) != "杭州星河科技有限公司" {
		t.Fatalf("unexpected demo customer name: %#v", firstCustomer)
	}

	if _, err := repo.RestoreTestData(context.Background(), 1, 9); err != nil {
		t.Fatalf("restore test data second time: %v", err)
	}
	assertTableCount(t, db, "customers", 4)
	assertTableCount(t, db, "finance_accounts", 4)
}

func newRestoreTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open("file:restore_test?mode=memory&cache=shared"), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("get sql db: %v", err)
	}
	sqlDB.SetMaxOpenConns(1)
	t.Cleanup(func() {
		_ = sqlDB.Close()
	})
	if err := db.AutoMigrate(
		&authmodel.User{}, &authmodel.LoginLog{}, &authmodel.OperationLog{},
		&customermodel.Customer{}, &customermodel.CustomerContact{}, &customermodel.CustomerAddress{},
		&suppliermodel.Supplier{},
		&businessmodel.BusinessPhoto{}, &businessmodel.DocumentDeleteDetail{}, &businessmodel.DocumentDeleteRecord{}, &businessmodel.SalesOrderDeletion{},
		&businessmodel.CustomerStatementItem{}, &businessmodel.CustomerStatement{},
		&businessmodel.Receivable{}, &businessmodel.Payable{}, &businessmodel.FinanceRecord{}, &businessmodel.FinanceAccount{},
		&businessmodel.InventoryMovement{}, &businessmodel.InventoryStock{}, &businessmodel.InventoryBatch{},
		&businessmodel.SalesCostAllocation{}, &businessmodel.SalesOrderItem{}, &businessmodel.SalesOrder{},
		&businessmodel.PurchaseOrderItem{}, &businessmodel.PurchaseOrder{},
		&businessmodel.RepairOrderItem{}, &businessmodel.RepairOrder{},
		&dashboardmodel.EngineerCheckIn{},
		&businessmodel.ProjectProject{},
		&businessmodel.Product{},
	); err != nil {
		t.Fatalf("auto migrate: %v", err)
	}
	return db
}

func assertTableCount(t *testing.T, db *gorm.DB, table string, want int64) {
	t.Helper()
	var got int64
	if err := db.Table(table).Count(&got).Error; err != nil {
		t.Fatalf("count %s: %v", table, err)
	}
	if got != want {
		t.Fatalf("count %s = %d, want %d", table, got, want)
	}
}
