package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"time"

	"erp/internal/config"
	"erp/internal/infrastructure/database"
	gormrepo "erp/internal/infrastructure/persistence/gorm"

	"gorm.io/gorm"
)

const (
	tenantID    uint64 = 1
	operatorID  uint64 = 1
	warehouse          = "main"
	accountName        = "FLOW-CASH"
)

type check struct {
	Name   string `json:"name"`
	Pass   bool   `json:"pass"`
	Detail string `json:"detail"`
}

type flow struct {
	db     *gorm.DB
	repo   *gormrepo.ModuleRepository
	ctx    context.Context
	checks []check
}

func main() {
	cfg, err := config.Load("config.yaml")
	if err != nil {
		log.Fatal(err)
	}
	db, err := database.Open(cfg.Database)
	if err != nil {
		log.Fatal(err)
	}
	if err := database.AutoMigrate(db, cfg.Admin); err != nil {
		log.Fatal(err)
	}
	f := &flow{db: db, repo: gormrepo.NewModuleRepository(db), ctx: context.Background(), checks: make([]check, 0, 32)}
	if err := f.run(); err != nil {
		log.Fatal(err)
	}
	failed := 0
	for _, item := range f.checks {
		if !item.Pass {
			failed++
		}
	}
	out, _ := json.MarshalIndent(map[string]any{
		"summary": map[string]any{"passed": len(f.checks) - failed, "failed": failed},
		"checks":  f.checks,
	}, "", "  ")
	fmt.Println(string(out))
	if failed > 0 {
		log.Fatalf("finance flow has %d failed checks", failed)
	}
}

func (f *flow) run() error {
	if err := f.clearData(); err != nil {
		return err
	}
	customerID, supplierID, productID, err := f.seedBaseData()
	if err != nil {
		return err
	}
	if err := f.purchasePayableFlow(supplierID, productID); err != nil {
		return err
	}
	if err := f.salesStatementReceiptFlow(customerID, productID); err != nil {
		return err
	}
	if _, err := f.repo.Create(f.ctx, "finance", tenantID, operatorID, map[string]any{
		"recordType":  "daily income",
		"accountName": accountName,
		"amount":      100,
	}); err == nil {
		f.add("manual finance create rejected", false, "manual finance create unexpectedly succeeded")
	} else {
		f.add("manual finance create rejected", true, err.Error())
	}
	return nil
}

func (f *flow) clearData() error {
	tables := []string{
		"customer_statement_items", "customer_statements",
		"receivables", "payables", "finance_records", "finance_accounts",
		"sales_cost_allocations", "sales_order_items", "sales_orders",
		"purchase_order_items", "purchase_orders",
		"inventory_movements", "inventory_batches", "inventory_stocks",
		"document_delete_details", "document_delete_records", "sales_order_deletions",
		"business_photos", "repair_order_items", "repair_orders", "project_projects",
		"customers", "suppliers", "products", "login_logs", "operation_logs",
	}
	return f.db.Transaction(func(tx *gorm.DB) error {
		for _, table := range tables {
			if tx.Migrator().HasTable(table) {
				if err := tx.Exec("DELETE FROM " + table).Error; err != nil {
					return err
				}
			}
		}
		return nil
	})
}

func (f *flow) seedBaseData() (uint64, uint64, uint64, error) {
	now := time.Now()
	if _, err := f.repo.Create(f.ctx, "finance-accounts", tenantID, operatorID, map[string]any{
		"code":           "FLOWCASH",
		"name":           accountName,
		"accountType":    "cash",
		"openingBalance": 10000,
		"status":         "active",
	}); err != nil {
		return 0, 0, 0, err
	}
	customer := map[string]any{
		"tenant_id": 1, "code": "FLOW-C001", "name": "Flow Customer", "type": "company",
		"payment_method": "monthly", "billing_cycle": "monthly", "credit_days": 30,
		"credit_limit": 100000, "receivable_balance": 0, "status": "active",
		"created_at": now, "updated_at": now,
	}
	if err := f.db.Table("customers").Create(customer).Error; err != nil {
		return 0, 0, 0, err
	}
	supplier := map[string]any{
		"tenant_id": 1, "code": "FLOW-S001", "name": "Flow Supplier", "contact_name": "Supplier Contact",
		"payable_balance": 0, "status": "active", "created_at": now, "updated_at": now,
	}
	if err := f.db.Table("suppliers").Create(supplier).Error; err != nil {
		return 0, 0, 0, err
	}
	product := map[string]any{
		"tenant_id": 1, "code": "FLOW-P001", "name": "Flow Product", "category": "Test",
		"brand": "Flow", "spec": "STD", "unit": "pcs", "status": "active",
		"created_at": now, "updated_at": now,
	}
	if err := f.db.Table("products").Create(product).Error; err != nil {
		return 0, 0, 0, err
	}
	var customerID, supplierID, productID uint64
	if err := f.db.Table("customers").Where("code = ?", "FLOW-C001").Select("id").Scan(&customerID).Error; err != nil {
		return 0, 0, 0, err
	}
	if err := f.db.Table("suppliers").Where("code = ?", "FLOW-S001").Select("id").Scan(&supplierID).Error; err != nil {
		return 0, 0, 0, err
	}
	if err := f.db.Table("products").Where("code = ?", "FLOW-P001").Select("id").Scan(&productID).Error; err != nil {
		return 0, 0, 0, err
	}
	return customerID, supplierID, productID, nil
}

func (f *flow) purchasePayableFlow(supplierID, productID uint64) error {
	purchase, err := f.repo.Create(f.ctx, "purchase", tenantID, operatorID, map[string]any{
		"supplierId":   supplierID,
		"supplierName": "Flow Supplier",
		"productId":    productID,
		"quantity":     10,
		"price":        800,
		"paidAmount":   0,
		"warehouse":    warehouse,
		"orderDate":    time.Now(),
	})
	if err != nil {
		return err
	}
	purchaseID := idOf(purchase)
	f.addAmount("purchase payable generated", f.sum("payables", "balance_amount", "source_id = ?", purchaseID), 8000)
	f.addAmount("purchase stock increased", f.sum("inventory_stocks", "quantity", "product_id = ?", productID), 10)

	payable := map[string]any{}
	if err := f.db.Table("payables").Where("source_id = ?", purchaseID).Take(&payable).Error; err != nil {
		return err
	}
	if _, err := f.repo.Action(f.ctx, "payables", tenantID, operatorID, "pay", map[string]any{
		"payableId":    idOf(payable),
		"supplierId":   supplierID,
		"supplierName": "Flow Supplier",
		"amount":       3000,
		"accountName":  accountName,
		"businessNo":   payable["payable_no"],
		"businessType": "payable payment",
	}); err != nil {
		return err
	}
	f.addAmount("payable balance after partial payment", f.sum("payables", "balance_amount", "id = ?", idOf(payable)), 5000)
	f.addAmount("purchase paid amount synced", f.sum("purchase_orders", "paid_amount", "id = ?", purchaseID), 3000)
	f.addAmount("cash after payable payment", f.accountBalance(), 7000)
	f.addAmount("payment finance flow generated", f.sum("finance_records", "amount", "source_type = ?", "expense"), 3000)
	return nil
}

func (f *flow) salesStatementReceiptFlow(customerID, productID uint64) error {
	batchID := uint64(0)
	if err := f.db.Table("inventory_batches").Where("product_id = ? AND remaining_quantity > 0", productID).Order("id ASC").Select("id").Scan(&batchID).Error; err != nil {
		return err
	}
	sale, err := f.repo.Create(f.ctx, "sales", tenantID, operatorID, map[string]any{
		"customerId":       customerID,
		"customerName":     "Flow Customer",
		"productId":        productID,
		"quantity":         4,
		"price":            2000,
		"receivedAmount":   0,
		"warehouse":        warehouse,
		"inventoryBatchId": batchID,
		"orderDate":        time.Now(),
	})
	if err != nil {
		return err
	}
	saleID := idOf(sale)
	f.addAmount("sales receivable generated", f.sum("receivables", "balance_amount", "source_id = ?", saleID), 8000)
	f.addAmount("sales stock deducted", f.sum("inventory_stocks", "quantity", "product_id = ?", productID), 6)

	statement, err := f.repo.Action(f.ctx, "customer-statements", tenantID, operatorID, "generate", map[string]any{
		"customerId": customerID,
		"startDate":  "2026-01-01",
		"endDate":    "2026-12-31",
		"saleIds":    []any{saleID},
	})
	if err != nil {
		return err
	}
	statementID := idOf(statement)
	f.addAmount("statement unpaid amount generated", number(statement["unpaidAmount"]), 8000)
	confirmed, err := f.repo.Action(f.ctx, "customer-statements", tenantID, operatorID, "confirm", map[string]any{"id": statementID})
	if err != nil {
		return err
	}
	f.add("statement confirmed", text(confirmed["status"]) == "confirmed", fmt.Sprintf("status=%s", text(confirmed["status"])))

	settled, err := f.repo.Action(f.ctx, "customer-statements", tenantID, operatorID, "settle", map[string]any{
		"id":          statementID,
		"amount":      8000,
		"accountName": accountName,
	})
	if err != nil {
		return err
	}
	f.add("statement settled", text(settled["status"]) == "settled", fmt.Sprintf("status=%s", text(settled["status"])))
	f.addAmount("statement unpaid cleared", number(settled["unpaidAmount"]), 0)
	f.addAmount("receivable cleared after statement receipt", f.sum("receivables", "balance_amount", "source_id = ?", saleID), 0)
	f.addAmount("sales received synced", f.sum("sales_orders", "received_amount", "id = ?", saleID), 8000)
	f.addAmount("customer receivable balance synced", f.sum("customers", "receivable_balance", "id = ?", customerID), 0)
	f.addAmount("receipt finance flow generated", f.sum("finance_records", "amount", "source_type = ?", "income"), 8000)
	f.addAmount("cash after statement receipt", f.accountBalance(), 15000)
	return nil
}

func (f *flow) add(name string, pass bool, detail string) {
	f.checks = append(f.checks, check{Name: name, Pass: pass, Detail: detail})
}

func (f *flow) addAmount(name string, got, want float64) {
	f.add(name, math.Abs(got-want) < 0.0001, fmt.Sprintf("got=%.2f want=%.2f", got, want))
}

func (f *flow) sum(table, column, where string, args ...any) float64 {
	var value float64
	db := f.db.Table(table)
	if where != "" {
		db = db.Where(where, args...)
	}
	_ = db.Select("COALESCE(SUM(" + column + "), 0)").Scan(&value).Error
	return value
}

func (f *flow) accountBalance() float64 {
	var value float64
	_ = f.db.Table("finance_accounts").Where("name = ?", accountName).Select("balance").Scan(&value).Error
	return value
}

func idOf(row map[string]any) uint64 {
	for _, key := range []string{"id", "ID"} {
		switch v := row[key].(type) {
		case uint64:
			return v
		case uint:
			return uint64(v)
		case int:
			return uint64(v)
		case int64:
			return uint64(v)
		case float64:
			return uint64(v)
		}
	}
	return 0
}

func number(value any) float64 {
	switch v := value.(type) {
	case int:
		return float64(v)
	case int64:
		return float64(v)
	case uint64:
		return float64(v)
	case float64:
		return v
	case float32:
		return float64(v)
	default:
		var out float64
		_, _ = fmt.Sscan(fmt.Sprint(v), &out)
		return out
	}
}

func text(value any) string {
	if value == nil {
		return ""
	}
	return fmt.Sprint(value)
}
