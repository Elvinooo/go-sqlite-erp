package gormrepo

import (
	"context"
	"testing"
	"time"

	"gorm.io/gorm"
)

func TestCustomerStatementGeneratesFromSalesAndSyncsReceivables(t *testing.T) {
	db := newDeleteTestDB(t)
	repo := NewModuleRepository(db)
	ctx := context.Background()
	now := time.Date(2026, 1, 15, 10, 0, 0, 0, time.Local)
	customerID := uint64(501)
	customerName := "Statement Test Customer"
	if err := db.Table("customers").Create(map[string]any{
		"id": customerID, "tenant_id": 1, "code": "C-STMT", "name": customerName,
		"payment_method": "monthly", "billing_cycle": "monthly", "credit_days": 30,
		"status": "active", "created_at": now, "updated_at": now,
	}).Error; err != nil {
		t.Fatalf("create customer: %v", err)
	}
	seedStatementSale(t, db, 6101, customerID, customerName, "STMT-SO-001", "Product A", now.AddDate(0, 0, -10), 5000, 0, 0)
	seedStatementSale(t, db, 6102, customerID, customerName, "STMT-SO-002", "Product B", now.AddDate(0, 0, -5), 3000, 1000, 0)
	seedStatementSale(t, db, 6103, 0, customerName, "STMT-SO-003", "Product C", now.AddDate(0, 0, -2), 1200, 0, 0)

	candidates, err := repo.Action(ctx, "customer-statements", 1, 7, "sales-candidates", map[string]any{
		"customer_id": customerID,
		"start_date":  "2026-01-01",
		"end_date":    "2026-01-31",
		"page":        1,
		"page_size":   20,
	})
	if err != nil {
		t.Fatalf("sales candidates: %v", err)
	}
	if total := int(numberValue(candidates["total"])); total != 3 {
		t.Fatalf("candidate total=%d data=%#v", total, candidates)
	}
	list := candidates["list"].([]map[string]any)
	assertNumber(t, "candidate one unpaid", statementCandidateByNo(t, list, "STMT-SO-001")["unpaidAmount"], 5000)
	assertNumber(t, "candidate two unpaid", statementCandidateByNo(t, list, "STMT-SO-002")["unpaidAmount"], 2000)

	created, err := repo.Action(ctx, "customer-statements", 1, 7, "generate", map[string]any{
		"customer_id": customerID,
		"start_date":  "2026-01-01",
		"end_date":    "2026-01-31",
		"sale_ids":    []any{float64(6101), float64(6102), float64(6103)},
	})
	if err != nil {
		t.Fatalf("generate statement: %v", err)
	}
	assertNumber(t, "statement total", created["totalAmount"], 9200)
	assertNumber(t, "statement received", created["receivedAmount"], 1000)
	assertNumber(t, "statement unpaid", created["unpaidAmount"], 8200)
	assertNumber(t, "statement cumulative debt", created["cumulativeDebt"], 8200)

	confirmed, err := repo.Action(ctx, "customer-statements", 1, 7, "confirm", map[string]any{"id": created["id"]})
	if err != nil {
		t.Fatalf("confirm statement: %v", err)
	}
	if stringValue(confirmed["status"]) != "confirmed" {
		t.Fatalf("confirmed status=%#v", confirmed)
	}

	var receivableCount int64
	if err := db.Table("receivables").Where("customer_id = ? AND deleted_at IS NULL", customerID).Count(&receivableCount).Error; err != nil {
		t.Fatalf("count receivables: %v", err)
	}
	if receivableCount != 3 {
		t.Fatalf("receivable count=%d", receivableCount)
	}
	customer := map[string]any{}
	if err := db.Table("customers").Where("id = ?", customerID).Take(&customer).Error; err != nil {
		t.Fatalf("load customer: %v", err)
	}
	assertNumber(t, "customer balance", customer["receivable_balance"], 8200)

	settled, err := repo.Action(ctx, "customer-statements", 1, 7, "settle", map[string]any{"id": created["id"], "account_name": "现金", "amount": 8200})
	if err != nil {
		t.Fatalf("settle statement: %v", err)
	}
	if stringValue(settled["status"]) != "settled" {
		t.Fatalf("settled status=%#v", settled)
	}
	assertNumber(t, "settled unpaid", settled["unpaidAmount"], 0)
	assertNumber(t, "settled received", settled["receivedAmount"], 9200)
	var receivableBalance float64
	if err := db.Table("receivables").Where("customer_id = ? AND deleted_at IS NULL", customerID).Select("COALESCE(SUM(balance_amount), 0)").Scan(&receivableBalance).Error; err != nil {
		t.Fatalf("sum receivable balance: %v", err)
	}
	assertNumber(t, "receivable balance after settle", receivableBalance, 0)
	finance := map[string]any{}
	if err := db.Table("finance_records").Where("business_no = ?", settled["statementNo"]).Take(&finance).Error; err != nil {
		t.Fatalf("load statement finance record: %v", err)
	}
	assertNumber(t, "finance amount", finance["amount"], 8200)
	account := map[string]any{}
	if err := db.Table("finance_accounts").Where("name = ?", "现金").Take(&account).Error; err != nil {
		t.Fatalf("load finance account: %v", err)
	}
	assertNumber(t, "cash balance", account["balance"], 8200)
	if _, err := repo.Create(ctx, "finance", 1, 7, map[string]any{
		"record_type":  "daily income",
		"account_name": "cash",
		"target_name":  customerName,
		"amount":       100,
	}); err == nil {
		t.Fatalf("manual finance create should be rejected")
	}
	if _, err := repo.Update(ctx, "finance-accounts", 1, uintValue(account["id"]), 7, map[string]any{
		"name":    account["name"],
		"balance": 999999,
	}); err != nil {
		t.Fatalf("update finance account metadata: %v", err)
	}
	accountAfterUpdate := map[string]any{}
	if err := db.Table("finance_accounts").Where("id = ?", account["id"]).Take(&accountAfterUpdate).Error; err != nil {
		t.Fatalf("reload finance account: %v", err)
	}
	assertNumber(t, "cash balance ignores manual edit", accountAfterUpdate["balance"], 8200)

	after, err := repo.Action(ctx, "customer-statements", 1, 7, "sales-candidates", map[string]any{
		"customer_id": customerID,
		"start_date":  "2026-01-01",
		"end_date":    "2026-01-31",
		"page":        1,
		"page_size":   20,
	})
	if err != nil {
		t.Fatalf("sales candidates after statement: %v", err)
	}
	if total := int(numberValue(after["total"])); total != 0 {
		t.Fatalf("candidate total after generate=%d data=%#v", total, after)
	}
}

func seedStatementSale(t *testing.T, db *gorm.DB, id uint64, customerID uint64, customerName string, orderNo string, productName string, orderDate time.Time, total float64, received float64, receivable float64) {
	t.Helper()
	row := map[string]any{
		"id": id, "tenant_id": 1, "order_no": orderNo,
		"customer_name": customerName, "product_name": productName,
		"quantity": 1, "price": total, "order_date": orderDate, "status": "completed",
		"total_amount": total, "received_amount": received, "receivable_amount": receivable,
		"cost_amount": total * 0.6, "profit_amount": total * 0.4, "profit_rate": 40,
		"cost_method": "manual", "created_by": 7, "updated_by": 7, "created_at": orderDate, "updated_at": orderDate,
	}
	if customerID > 0 {
		row["customer_id"] = customerID
	}
	if err := db.Table("sales_orders").Create(row).Error; err != nil {
		t.Fatalf("create sale %s: %v", orderNo, err)
	}
}

func statementCandidateByNo(t *testing.T, list []map[string]any, saleNo string) map[string]any {
	t.Helper()
	for _, item := range list {
		if stringValue(item["saleNo"]) == saleNo {
			return item
		}
	}
	t.Fatalf("candidate %s not found in %#v", saleNo, list)
	return nil
}
