package gormrepo

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"

	authmodel "erp/internal/domain/auth/model"
	businessmodel "erp/internal/domain/business/model"
	customermodel "erp/internal/domain/customer/model"
	suppliermodel "erp/internal/domain/supplier/model"
	apperrors "erp/internal/shared/errors"
	"erp/internal/shared/query"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func TestDeleteSalesOrderRestoresInventoryAndKeepsAudit(t *testing.T) {
	db := newDeleteTestDB(t)
	repo := NewModuleRepository(db)
	seedSalesOrderDeleteFixture(t, db, 101, "已完成")

	if err := repo.Delete(context.Background(), "sales", 1, 101, 7, "客户取消订单"); err != nil {
		t.Fatalf("delete sales order: %v", err)
	}

	order := map[string]any{}
	if err := db.Table("sales_orders").Where("id = ?", 101).Take(&order).Error; err != nil {
		t.Fatalf("load deleted order: %v", err)
	}
	if stringValue(order["status"]) != "已删除" || stringValue(order["delete_reason"]) != "客户取消订单" || uintValue(order["deleted_by"]) != 7 || empty(order["deleted_at"]) {
		t.Fatalf("unexpected deleted order: %#v", order)
	}
	assertNumber(t, "deleted sales received", order["received_amount"], 0)
	assertNumber(t, "deleted sales receivable", order["receivable_amount"], 0)

	snapshot := map[string]any{}
	if err := db.Table("sales_order_deletions").Where("original_order_id = ?", 101).Take(&snapshot).Error; err != nil {
		t.Fatalf("load deletion snapshot: %v", err)
	}
	var items []map[string]any
	if err := json.Unmarshal([]byte(stringValue(snapshot["items_json"])), &items); err != nil || len(items) != 1 {
		t.Fatalf("invalid items snapshot: items=%#v err=%v", items, err)
	}
	var fullSnapshot map[string]any
	if err := json.Unmarshal([]byte(stringValue(snapshot["snapshot_json"])), &fullSnapshot); err != nil || fullSnapshot["cost_allocations"] == nil {
		t.Fatalf("invalid full snapshot: snapshot=%#v err=%v", fullSnapshot, err)
	}

	stock := map[string]any{}
	if err := db.Table("inventory_stocks").Where("product_code = ?", "SKU-101").Take(&stock).Error; err != nil {
		t.Fatalf("load stock: %v", err)
	}
	assertNumber(t, "stock quantity", stock["quantity"], 10)
	batch := map[string]any{}
	if err := db.Table("inventory_batches").Where("id = ?", 301).Take(&batch).Error; err != nil {
		t.Fatalf("load batch: %v", err)
	}
	assertNumber(t, "batch remaining", batch["remaining_quantity"], 10)

	movement := map[string]any{}
	if err := db.Table("inventory_movements").Where("source_type = ? AND source_id = ?", "ORDER_DELETE", 101).Take(&movement).Error; err != nil {
		t.Fatalf("load inventory movement: %v", err)
	}
	if stringValue(movement["direction"]) != "in" {
		t.Fatalf("unexpected movement: %#v", movement)
	}
	if stringValue(movement["warehouse"]) != "主仓库" {
		t.Fatalf("unexpected movement warehouse: %#v", movement)
	}
	assertNumber(t, "movement quantity", movement["quantity"], 4)
	receivable := map[string]any{}
	if err := db.Table("receivables").Where("source_id = ?", 101).Take(&receivable).Error; err != nil {
		t.Fatalf("load receivable: %v", err)
	}
	if stringValue(receivable["status"]) != "已冲销" {
		t.Fatalf("unexpected receivable: %#v", receivable)
	}
	assertNumber(t, "receivable balance", receivable["balance_amount"], 0)
	statementItem := map[string]any{}
	if err := db.Table("customer_statement_items").Where("sale_id = ?", 101).Take(&statementItem).Error; err != nil {
		t.Fatalf("load statement item: %v", err)
	}
	if stringValue(statementItem["settlement_status"]) != "已冲销" {
		t.Fatalf("unexpected statement item: %#v", statementItem)
	}
	assertNumber(t, "statement item total", statementItem["total_amount"], 0)
	statement := map[string]any{}
	if err := db.Table("customer_statements").Where("id = ?", 501).Take(&statement).Error; err != nil {
		t.Fatalf("load statement: %v", err)
	}
	assertNumber(t, "statement unpaid", statement["unpaid_amount"], 0)
	var reversalCount int64
	_ = db.Table("finance_records").Where("business_type IN ?", []string{"客户对账单删除冲销", "销售删除冲销"}).Count(&reversalCount).Error
	if reversalCount != 1 {
		t.Fatalf("reversal finance count=%d", reversalCount)
	}
	account := map[string]any{}
	if err := db.Table("finance_accounts").Where("id = ?", 701).Take(&account).Error; err != nil {
		t.Fatalf("load finance account: %v", err)
	}
	assertNumber(t, "account balance after reversal", account["balance"], 1000)

	var operationLogs int64
	if err := db.Table("operation_logs").Where("module = ? AND action = ?", "sales", "delete").Count(&operationLogs).Error; err != nil || operationLogs != 1 {
		t.Fatalf("operation log count=%d err=%v", operationLogs, err)
	}
	list, total, err := repo.List(context.Background(), "sales", 1, query.PageQuery{Page: 1, PageSize: 20, DeletedOnly: true})
	if err != nil || total != 1 || len(list) != 1 || stringValue(list[0]["deleteReason"]) != "客户取消订单" {
		t.Fatalf("deleted list total=%d list=%#v err=%v", total, list, err)
	}

	err = repo.Delete(context.Background(), "sales", 1, 101, 7, "再次删除")
	assertAppError(t, err, 40034, "订单已经删除")
	var movementCount int64
	_ = db.Table("inventory_movements").Where("source_type = ? AND source_id = ?", "ORDER_DELETE", 101).Count(&movementCount).Error
	if movementCount != 1 {
		t.Fatalf("duplicate delete restored inventory %d times", movementCount)
	}
}

func TestDeletePurchaseOrderReversesStockFinanceAndKeepsAudit(t *testing.T) {
	db := newDeleteTestDB(t)
	repo := NewModuleRepository(db)
	seedPurchaseOrderDeleteFixture(t, db, 201)

	if err := repo.Delete(context.Background(), "purchase", 1, 201, 7, "duplicate entry"); err != nil {
		t.Fatalf("delete purchase order: %v", err)
	}

	order := map[string]any{}
	if err := db.Table("purchase_orders").Where("id = ?", 201).Take(&order).Error; err != nil {
		t.Fatalf("load deleted purchase order: %v", err)
	}
	if stringValue(order["status"]) != "已删除" || stringValue(order["delete_reason"]) != "duplicate entry" || uintValue(order["deleted_by"]) != 7 || empty(order["deleted_at"]) {
		t.Fatalf("unexpected deleted purchase order: %#v", order)
	}
	record := map[string]any{}
	if err := db.Table("document_delete_records").Where("document_type = ? AND document_id = ?", "PURCHASE_ORDER", 201).Take(&record).Error; err != nil {
		t.Fatalf("load document delete record: %v", err)
	}
	if stringValue(record["delete_status"]) != "SUCCESS" || !boolValue(record["stock_processed"]) || !boolValue(record["finance_processed"]) {
		t.Fatalf("unexpected delete record: %#v", record)
	}
	var detailCount int64
	_ = db.Table("document_delete_details").Where("record_id = ?", mapID(record)).Count(&detailCount).Error
	if detailCount < 2 {
		t.Fatalf("detail count=%d", detailCount)
	}
	stock := map[string]any{}
	if err := db.Table("inventory_stocks").Where("product_code = ?", "PUR-201").Take(&stock).Error; err != nil {
		t.Fatalf("load stock: %v", err)
	}
	assertNumber(t, "purchase stock quantity", stock["quantity"], 0)
	movement := map[string]any{}
	if err := db.Table("inventory_movements").Where("source_type = ? AND source_id = ?", "PURCHASE_DELETE", 201).Take(&movement).Error; err != nil {
		t.Fatalf("load purchase delete movement: %v", err)
	}
	if stringValue(movement["direction"]) != "out" {
		t.Fatalf("unexpected purchase delete movement: %#v", movement)
	}
	payable := map[string]any{}
	if err := db.Table("payables").Where("source_id = ?", 201).Take(&payable).Error; err != nil {
		t.Fatalf("load payable: %v", err)
	}
	if stringValue(payable["status"]) != "已冲销" {
		t.Fatalf("unexpected payable: %#v", payable)
	}
	var financeCount int64
	_ = db.Table("finance_records").Where("business_no = ?", "PO-201").Count(&financeCount).Error
	if financeCount != 2 {
		t.Fatalf("finance reversal count=%d", financeCount)
	}
	err := repo.Delete(context.Background(), "purchase", 1, 201, 7, "again")
	assertAppError(t, err, 40034, "订单已经删除")
}

func TestDeletePurchaseOrderRejectsSoldBatch(t *testing.T) {
	db := newDeleteTestDB(t)
	repo := NewModuleRepository(db)
	seedPurchaseOrderDeleteFixture(t, db, 202)
	if err := db.Table("inventory_batches").Where("purchase_order_id = ?", 202).Updates(map[string]any{
		"remaining_quantity": 3,
		"status":             "销售中",
	}).Error; err != nil {
		t.Fatalf("mark batch sold: %v", err)
	}

	err := repo.Delete(context.Background(), "purchase", 1, 202, 7, "已销售不可删")
	assertAppError(t, err, 40038, "采购订单库存已被销售或领用，无法删除")
	order := map[string]any{}
	if err := db.Table("purchase_orders").Where("id = ?", 202).Take(&order).Error; err != nil {
		t.Fatalf("load purchase order: %v", err)
	}
	if !empty(order["deleted_at"]) || stringValue(order["status"]) == "已删除" {
		t.Fatalf("purchase order should not be deleted: %#v", order)
	}
	var recordCount int64
	_ = db.Table("document_delete_records").Where("document_type = ? AND document_id = ?", "PURCHASE_ORDER", 202).Count(&recordCount).Error
	if recordCount != 0 {
		t.Fatalf("unexpected delete record count=%d", recordCount)
	}
}

func TestDeleteSalesOrderRejectsOutboundAndRollsBackFailures(t *testing.T) {
	t.Run("outbound order", func(t *testing.T) {
		db := newDeleteTestDB(t)
		repo := NewModuleRepository(db)
		seedSalesOrderDeleteFixture(t, db, 102, "已出库")

		err := repo.Delete(context.Background(), "sales", 1, 102, 7, "测试删除")
		assertAppError(t, err, 40035, "订单已经出库，无法删除")
		assertDeleteSideEffects(t, db, 102, 0, 6)
	})

	t.Run("missing batch rolls back", func(t *testing.T) {
		db := newDeleteTestDB(t)
		repo := NewModuleRepository(db)
		seedSalesOrderDeleteFixture(t, db, 103, "已完成")
		if err := db.Exec("DELETE FROM inventory_batches WHERE id = ?", 303).Error; err != nil {
			t.Fatalf("remove batch: %v", err)
		}

		err := repo.Delete(context.Background(), "sales", 1, 103, 7, "触发回滚")
		if err == nil {
			t.Fatal("expected delete failure")
		}
		assertDeleteSideEffects(t, db, 103, 0, 6)
	})
}

func TestDeleteSalesOrderConcurrentRequestsRestoreOnce(t *testing.T) {
	db := newDeleteTestDB(t)
	repo := NewModuleRepository(db)
	seedSalesOrderDeleteFixture(t, db, 104, "已完成")

	start := make(chan struct{})
	errs := make(chan error, 2)
	var wg sync.WaitGroup
	for i := 0; i < 2; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start
			errs <- repo.Delete(context.Background(), "sales", 1, 104, 7, "并发删除")
		}()
	}
	close(start)
	wg.Wait()
	close(errs)

	successes := 0
	alreadyDeleted := 0
	for err := range errs {
		if err == nil {
			successes++
			continue
		}
		var appErr *apperrors.AppError
		if errors.As(err, &appErr) && appErr.Code == 40034 {
			alreadyDeleted++
			continue
		}
		t.Fatalf("unexpected concurrent error: %v", err)
	}
	if successes != 1 || alreadyDeleted != 1 {
		t.Fatalf("successes=%d alreadyDeleted=%d", successes, alreadyDeleted)
	}
	stock := map[string]any{}
	_ = db.Table("inventory_stocks").Where("product_code = ?", "SKU-104").Take(&stock).Error
	assertNumber(t, "concurrent stock quantity", stock["quantity"], 10)
	var movementCount int64
	_ = db.Table("inventory_movements").Where("source_type = ? AND source_id = ?", "ORDER_DELETE", 104).Count(&movementCount).Error
	if movementCount != 1 {
		t.Fatalf("concurrent delete movement count=%d", movementCount)
	}
}

func newDeleteTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
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
		&authmodel.User{}, &authmodel.OperationLog{},
		&customermodel.Customer{}, &suppliermodel.Supplier{},
		&businessmodel.SalesOrder{}, &businessmodel.SalesOrderItem{}, &businessmodel.SalesCostAllocation{}, &businessmodel.SalesOrderDeletion{},
		&businessmodel.DocumentDeleteRecord{}, &businessmodel.DocumentDeleteDetail{},
		&businessmodel.PurchaseOrder{}, &businessmodel.PurchaseOrderItem{},
		&businessmodel.InventoryBatch{}, &businessmodel.InventoryStock{}, &businessmodel.InventoryMovement{},
		&businessmodel.FinanceAccount{}, &businessmodel.FinanceRecord{}, &businessmodel.Receivable{},
		&businessmodel.CustomerStatement{}, &businessmodel.CustomerStatementItem{}, &businessmodel.Payable{},
	); err != nil {
		t.Fatalf("auto migrate: %v", err)
	}
	if err := db.Exec("CREATE UNIQUE INDEX IF NOT EXISTS ux_sales_order_deletions_order ON sales_order_deletions(tenant_id, original_order_id)").Error; err != nil {
		t.Fatalf("create unique index: %v", err)
	}
	if err := db.Exec("CREATE UNIQUE INDEX IF NOT EXISTS ux_document_delete_records_doc ON document_delete_records(tenant_id, document_type, document_id)").Error; err != nil {
		t.Fatalf("create document delete unique index: %v", err)
	}
	now := time.Now()
	if err := db.Table("users").Create(map[string]any{
		"id": 7, "tenant_id": 1, "username": "tester", "password_hash": "test", "status": "active", "created_at": now, "updated_at": now,
	}).Error; err != nil {
		t.Fatalf("create user: %v", err)
	}
	return db
}

func seedSalesOrderDeleteFixture(t *testing.T, db *gorm.DB, orderID uint64, status string) {
	t.Helper()
	now := time.Now()
	suffix := orderID - 100
	batchID := uint64(300) + suffix
	productCode := fmt.Sprintf("SKU-%d", orderID)
	order := map[string]any{
		"id": orderID, "tenant_id": 1, "order_no": fmt.Sprintf("SO-%d", orderID), "customer_id": 11, "customer_name": "测试客户",
		"product_id": 21, "product_code": productCode, "product_name": "测试商品", "quantity": 4, "price": 20, "cost_price": 10,
		"order_date": now, "status": status, "total_amount": 80, "cost_amount": 40, "profit_amount": 40, "cost_method": "purchase_source",
		"inventory_batch_id": batchID, "created_by": 7, "updated_by": 7, "created_at": now, "updated_at": now,
	}
	if err := db.Table("sales_orders").Create(order).Error; err != nil {
		t.Fatalf("create sales order: %v", err)
	}
	item := map[string]any{
		"tenant_id": 1, "sales_order_id": orderID, "product_id": 21, "product_code": productCode, "product_name": "测试商品",
		"quantity": 4, "price": 20, "amount": 80, "purchase_source": "purchase_order", "cost_price": 10, "cost_amount": 40,
		"created_by": 7, "updated_by": 7, "created_at": now, "updated_at": now,
	}
	if err := db.Table("sales_order_items").Create(item).Error; err != nil {
		t.Fatalf("create sales item: %v", err)
	}
	batch := map[string]any{
		"id": batchID, "tenant_id": 1, "batch_no": fmt.Sprintf("BAT-%d", orderID), "product_id": 21, "product_code": productCode,
		"product_name": "测试商品", "warehouse": "主仓库", "purchase_order_id": 41, "purchase_order_no": "PO-1",
		"purchase_date": now, "purchase_price": 10, "inbound_quantity": 10, "remaining_quantity": 6, "status": "销售中",
		"created_by": 7, "updated_by": 7, "created_at": now, "updated_at": now,
	}
	if err := db.Table("inventory_batches").Create(batch).Error; err != nil {
		t.Fatalf("create batch: %v", err)
	}
	allocation := map[string]any{
		"tenant_id": 1, "sales_order_id": orderID, "sales_order_item_id": item["@id"], "inventory_batch_id": batchID,
		"product_id": 21, "product_code": productCode, "product_name": "测试商品", "quantity": 4, "cost_price": 10, "cost_amount": 40,
		"purchase_order_id": 41, "purchase_order_no": "PO-1", "purchase_price": 10,
		"created_by": 7, "updated_by": 7, "created_at": now, "updated_at": now,
	}
	if err := db.Table("sales_cost_allocations").Create(allocation).Error; err != nil {
		t.Fatalf("create allocation: %v", err)
	}
	stock := map[string]any{
		"tenant_id": 1, "product_id": 21, "product_code": productCode, "product_name": "测试商品", "warehouse": "主仓库",
		"quantity": 6, "avg_cost": 10, "amount": 60, "min_stock": 0, "status": "正常", "stock_time": now,
		"created_by": 7, "updated_by": 7, "created_at": now, "updated_at": now,
	}
	if err := db.Table("inventory_stocks").Create(stock).Error; err != nil {
		t.Fatalf("create stock: %v", err)
	}
	if orderID == 101 {
		account := map[string]any{
			"id": 701, "tenant_id": 1, "code": "CASH", "name": "现金", "account_type": "cash",
			"opening_balance": 1000, "balance": 1030, "status": "active", "created_at": now, "updated_at": now,
		}
		if err := db.Table("finance_accounts").Create(account).Error; err != nil {
			t.Fatalf("create sales finance account: %v", err)
		}
		receivable := map[string]any{
			"tenant_id": 1, "receivable_no": "AR-101", "customer_id": 11, "customer_name": "测试客户",
			"source_type": "sales", "source_id": orderID, "source_no": "SO-101",
			"total_amount": 80, "received_amount": 30, "balance_amount": 50,
			"invoice_date": now, "due_date": now.AddDate(0, 0, 30), "settlement_mode": "monthly", "status": "partial",
			"created_by": 7, "updated_by": 7, "created_at": now, "updated_at": now,
		}
		if err := db.Table("receivables").Create(receivable).Error; err != nil {
			t.Fatalf("create sales receivable: %v", err)
		}
		if err := db.Table("sales_orders").Where("id = ?", orderID).Updates(map[string]any{"received_amount": 30, "receivable_amount": 50}).Error; err != nil {
			t.Fatalf("update sales collection: %v", err)
		}
		statement := map[string]any{
			"id": 501, "tenant_id": 1, "statement_no": "CS-101", "customer_id": 11, "customer_name": "测试客户",
			"start_date": now.AddDate(0, 0, -1), "end_date": now.AddDate(0, 0, 1),
			"total_amount": 80, "received_amount": 30, "unpaid_amount": 50, "cumulative_debt": 50, "status": "confirmed",
			"created_by": 7, "updated_by": 7, "created_at": now, "updated_at": now,
		}
		if err := db.Table("customer_statements").Create(statement).Error; err != nil {
			t.Fatalf("create customer statement: %v", err)
		}
		statementItem := map[string]any{
			"tenant_id": 1, "statement_id": 501, "sale_id": orderID, "sale_no": "SO-101", "sale_date": now,
			"source_type": "sales", "source_id": orderID, "source_no": "SO-101",
			"product_name": "测试商品", "quantity": 4, "total_amount": 80, "received_amount": 30, "unpaid_amount": 50, "settlement_status": "partial",
			"created_by": 7, "updated_by": 7, "created_at": now, "updated_at": now,
		}
		if err := db.Table("customer_statement_items").Create(statementItem).Error; err != nil {
			t.Fatalf("create customer statement item: %v", err)
		}
		finance := map[string]any{
			"tenant_id": 1, "record_no": "FIN-SO-101", "record_type": "收款", "account_id": 701, "account_name": "现金",
			"target_name": "测试客户", "amount": 30, "status": "已确认", "business_type": "客户对账单收款",
			"source_type": "income", "business_no": "CS-101", "occurred_at": now, "created_by": 7, "updated_by": 7,
			"created_at": now, "updated_at": now,
		}
		if err := db.Table("finance_records").Create(finance).Error; err != nil {
			t.Fatalf("create sales finance record: %v", err)
		}
	}
}

func seedPurchaseOrderDeleteFixture(t *testing.T, db *gorm.DB, orderID uint64) {
	t.Helper()
	now := time.Now()
	order := map[string]any{
		"id": orderID, "tenant_id": 1, "order_no": "PO-201", "supplier_id": 31, "supplier_name": "Supplier A",
		"product_id": 51, "product_code": "PUR-201", "product_name": "Purchase Item", "quantity": 5, "price": 12,
		"order_date": now, "status": "done", "total_amount": 60, "paid_amount": 20, "payable_amount": 40,
		"created_by": 7, "updated_by": 7, "created_at": now, "updated_at": now,
	}
	if err := db.Table("purchase_orders").Create(order).Error; err != nil {
		t.Fatalf("create purchase order: %v", err)
	}
	item := map[string]any{
		"tenant_id": 1, "purchase_order_id": orderID, "product_id": 51, "product_code": "PUR-201",
		"product_name": "Purchase Item", "quantity": 5, "price": 12, "amount": 60, "warehouse": "main",
		"created_by": 7, "updated_by": 7, "created_at": now, "updated_at": now,
	}
	if err := db.Table("purchase_order_items").Create(item).Error; err != nil {
		t.Fatalf("create purchase item: %v", err)
	}
	batch := map[string]any{
		"id": 401, "tenant_id": 1, "batch_no": "BAT-201", "product_id": 51, "product_code": "PUR-201",
		"product_name": "Purchase Item", "warehouse": "main", "purchase_order_id": orderID, "purchase_order_item_id": item["@id"],
		"purchase_order_no": "PO-201", "supplier_id": 31, "supplier_name": "Supplier A", "purchase_date": now,
		"purchase_price": 12, "inbound_quantity": 5, "remaining_quantity": 5, "status": "available",
		"created_by": 7, "updated_by": 7, "created_at": now, "updated_at": now,
	}
	if err := db.Table("inventory_batches").Create(batch).Error; err != nil {
		t.Fatalf("create purchase batch: %v", err)
	}
	stock := map[string]any{
		"tenant_id": 1, "product_id": 51, "product_code": "PUR-201", "product_name": "Purchase Item", "warehouse": "main",
		"quantity": 5, "avg_cost": 12, "amount": 60, "min_stock": 0, "status": "normal", "stock_time": now,
		"created_by": 7, "updated_by": 7, "created_at": now, "updated_at": now,
	}
	if err := db.Table("inventory_stocks").Create(stock).Error; err != nil {
		t.Fatalf("create purchase stock: %v", err)
	}
	account := map[string]any{
		"id": 801, "tenant_id": 1, "code": "CASH", "name": "现金", "account_type": "cash",
		"opening_balance": 1000, "balance": 980, "status": "active", "created_at": now, "updated_at": now,
	}
	if err := db.Table("finance_accounts").Create(account).Error; err != nil {
		t.Fatalf("create finance account: %v", err)
	}
	finance := map[string]any{
		"tenant_id": 1, "record_no": "FIN-201", "record_type": "付款", "account_id": 801, "account_name": "现金",
		"target_name": "Supplier A", "amount": 20, "status": "confirmed", "business_type": "purchase payment",
		"source_type": "expense", "business_no": "PO-201", "occurred_at": now, "created_by": 7, "updated_by": 7,
		"created_at": now, "updated_at": now,
	}
	if err := db.Table("finance_records").Create(finance).Error; err != nil {
		t.Fatalf("create finance record: %v", err)
	}
	payable := map[string]any{
		"tenant_id": 1, "payable_no": "AP-201", "supplier_id": 31, "supplier_name": "Supplier A", "source_type": "purchase",
		"source_id": orderID, "source_no": "PO-201", "total_amount": 60, "paid_amount": 20, "balance_amount": 40,
		"bill_date": now, "due_date": now.AddDate(0, 0, 30), "status": "unpaid", "created_by": 7, "updated_by": 7,
		"created_at": now, "updated_at": now,
	}
	if err := db.Table("payables").Create(payable).Error; err != nil {
		t.Fatalf("create payable: %v", err)
	}
}

func assertDeleteSideEffects(t *testing.T, db *gorm.DB, orderID uint64, snapshots int64, stockQuantity float64) {
	t.Helper()
	var snapshotCount int64
	_ = db.Table("sales_order_deletions").Where("original_order_id = ?", orderID).Count(&snapshotCount).Error
	if snapshotCount != snapshots {
		t.Fatalf("snapshot count=%d want=%d", snapshotCount, snapshots)
	}
	order := map[string]any{}
	_ = db.Table("sales_orders").Where("id = ?", orderID).Take(&order).Error
	if !empty(order["deleted_at"]) {
		t.Fatalf("order should remain active: %#v", order)
	}
	stock := map[string]any{}
	_ = db.Table("inventory_stocks").Where("product_code = ?", fmt.Sprintf("SKU-%d", orderID)).Take(&stock).Error
	assertNumber(t, "rollback stock quantity", stock["quantity"], stockQuantity)
	var movementCount int64
	_ = db.Table("inventory_movements").Where("source_type = ? AND source_id = ?", "ORDER_DELETE", orderID).Count(&movementCount).Error
	if movementCount != 0 {
		t.Fatalf("rollback movement count=%d", movementCount)
	}
}

func assertAppError(t *testing.T, err error, code int, message string) {
	t.Helper()
	var appErr *apperrors.AppError
	if !errors.As(err, &appErr) || appErr.Code != code || appErr.Message != message {
		t.Fatalf("error=%v want code=%d message=%q", err, code, message)
	}
}

func assertNumber(t *testing.T, name string, value any, want float64) {
	t.Helper()
	got := numberValue(value)
	if got < want-0.000001 || got > want+0.000001 {
		t.Fatalf("%s=%v want=%v", name, got, want)
	}
}
