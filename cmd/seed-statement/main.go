package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"erp/internal/config"
	"erp/internal/infrastructure/database"
	gormrepo "erp/internal/infrastructure/persistence/gorm"
	"gorm.io/gorm"
)

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
	if err := seedStatementDemo(db); err != nil {
		log.Fatal(err)
	}
	fmt.Println("customer statement test data seeded")
}

func seedStatementDemo(db *gorm.DB) error {
	return db.Transaction(func(tx *gorm.DB) error {
		now := time.Date(2026, 1, 31, 10, 0, 0, 0, time.Local)
		generatedCustomerID, err := upsertStatementCustomer(tx, "STMT001", "对账测试客户-已生成", "13800001001", now)
		if err != nil {
			return err
		}
		pendingCustomerID, err := upsertStatementCustomer(tx, "STMT002", "对账测试客户-待生成", "13800001002", now)
		if err != nil {
			return err
		}

		generatedSales, err := upsertStatementSales(tx, generatedCustomerID, "对账测试客户-已生成", []statementSaleSeed{
			{OrderNo: "STMT202601001", ProductName: "月结测试商品A", OrderDate: "2026-01-05", Total: 5000, Received: 0, Receivable: 5000},
			{OrderNo: "STMT202601002", ProductName: "月结测试商品B", OrderDate: "2026-01-15", Total: 3000, Received: 1000, Receivable: 2000},
			{OrderNo: "STMT202601003", ProductName: "月结测试商品C", OrderDate: "2026-01-28", Total: 2000, Received: 2000, Receivable: 0},
		}, now)
		if err != nil {
			return err
		}
		if err := upsertStatementSalesOnly(tx, pendingCustomerID, "对账测试客户-待生成", []statementSaleSeed{
			{OrderNo: "STMT202602001", ProductName: "待生成测试商品A", OrderDate: "2026-02-03", Total: 2600, Received: 0, Receivable: 0},
			{OrderNo: "STMT202602002", ProductName: "待生成测试商品B", OrderDate: "2026-02-12", Total: 1800, Received: 300, Receivable: 0},
		}, now); err != nil {
			return err
		}

		var linked int64
		if err := tx.Table("customer_statement_items").Where("tenant_id = ? AND sale_id IN ? AND deleted_at IS NULL", 1, generatedSales).Count(&linked).Error; err != nil {
			return err
		}
		if linked == 0 {
			repo := gormrepo.NewModuleRepository(tx)
			_, err = repo.Action(context.Background(), "customer-statements", 1, 1, "generate", map[string]any{
				"customer_id": generatedCustomerID,
				"start_date":  "2026-01-01",
				"end_date":    "2026-01-31",
				"sale_ids":    uintIDsToAny(generatedSales),
				"remark":      "seeded statement demo",
			})
			if err != nil {
				return err
			}
		}
		return printStatementSeedSummary(tx, generatedCustomerID, pendingCustomerID)
	})
}

func printStatementSeedSummary(tx *gorm.DB, generatedCustomerID uint64, pendingCustomerID uint64) error {
	var statementCount int64
	if err := tx.Table("customer_statements").Where("tenant_id = ? AND customer_id = ? AND deleted_at IS NULL", 1, generatedCustomerID).Count(&statementCount).Error; err != nil {
		return err
	}
	var pendingSales int64
	if err := tx.Table("sales_orders AS so").
		Joins("LEFT JOIN customer_statement_items AS csi ON csi.tenant_id = so.tenant_id AND csi.sale_id = so.id AND csi.deleted_at IS NULL").
		Where("so.tenant_id = ? AND so.customer_id = ? AND so.deleted_at IS NULL AND csi.id IS NULL", 1, pendingCustomerID).
		Count(&pendingSales).Error; err != nil {
		return err
	}
	var balance float64
	if err := tx.Table("customers").Where("tenant_id = ? AND id = ?", 1, generatedCustomerID).Select("receivable_balance").Scan(&balance).Error; err != nil {
		return err
	}
	fmt.Printf("seed summary: generatedStatements=%d pendingStatementSales=%d generatedCustomerBalance=%.2f\n", statementCount, pendingSales, balance)
	return nil
}

type statementSaleSeed struct {
	OrderNo     string
	ProductName string
	OrderDate   string
	Total       float64
	Received    float64
	Receivable  float64
}

func upsertStatementCustomer(tx *gorm.DB, code string, name string, phone string, now time.Time) (uint64, error) {
	row := map[string]any{}
	err := tx.Table("customers").Where("tenant_id = ? AND code = ? AND deleted_at IS NULL", 1, code).Take(&row).Error
	if err == nil {
		id := uintValue(row["id"])
		return id, tx.Table("customers").Where("id = ?", id).Updates(map[string]any{
			"name": name, "phone": phone, "payment_method": "monthly", "billing_cycle": "monthly",
			"credit_days": 30, "status": "active", "updated_at": now,
		}).Error
	}
	if err != gorm.ErrRecordNotFound {
		return 0, err
	}
	create := map[string]any{
		"tenant_id": 1, "code": code, "name": name, "type": "company", "level": "月结客户",
		"phone": phone, "payment_method": "monthly", "billing_cycle": "monthly", "credit_days": 30,
		"credit_limit": 100000, "receivable_balance": 0, "status": "active", "created_at": now, "updated_at": now,
	}
	if err := tx.Table("customers").Create(create).Error; err != nil {
		return 0, err
	}
	row = map[string]any{}
	if err := tx.Table("customers").Where("tenant_id = ? AND code = ?", 1, code).Take(&row).Error; err != nil {
		return 0, err
	}
	return uintValue(row["id"]), nil
}

func upsertStatementSales(tx *gorm.DB, customerID uint64, customerName string, items []statementSaleSeed, now time.Time) ([]uint64, error) {
	if err := upsertStatementSalesOnly(tx, customerID, customerName, items, now); err != nil {
		return nil, err
	}
	ids := make([]uint64, 0, len(items))
	for _, item := range items {
		row := map[string]any{}
		if err := tx.Table("sales_orders").Where("tenant_id = ? AND order_no = ?", 1, item.OrderNo).Take(&row).Error; err != nil {
			return nil, err
		}
		ids = append(ids, uintValue(row["id"]))
	}
	return ids, nil
}

func upsertStatementSalesOnly(tx *gorm.DB, customerID uint64, customerName string, items []statementSaleSeed, now time.Time) error {
	for _, item := range items {
		orderDate, err := time.ParseInLocation("2006-01-02", item.OrderDate, time.Local)
		if err != nil {
			return err
		}
		row := map[string]any{}
		err = tx.Table("sales_orders").Where("tenant_id = ? AND order_no = ?", 1, item.OrderNo).Take(&row).Error
		values := map[string]any{
			"customer_id": customerID, "customer_name": customerName, "product_name": item.ProductName,
			"quantity": 1, "price": item.Total, "order_date": orderDate, "status": "completed",
			"total_amount": item.Total, "received_amount": item.Received, "receivable_amount": item.Receivable,
			"cost_amount": item.Total * 0.6, "profit_amount": item.Total * 0.4, "profit_rate": 40,
			"cost_method": "manual", "updated_by": 1, "updated_at": now,
		}
		if err == nil {
			if err := tx.Table("sales_orders").Where("id = ?", uintValue(row["id"])).Updates(values).Error; err != nil {
				return err
			}
			continue
		}
		if err != gorm.ErrRecordNotFound {
			return err
		}
		values["tenant_id"] = 1
		values["order_no"] = item.OrderNo
		values["created_by"] = 1
		values["created_at"] = now
		if err := tx.Table("sales_orders").Create(values).Error; err != nil {
			return err
		}
	}
	return nil
}

func uintIDsToAny(ids []uint64) []any {
	out := make([]any, 0, len(ids))
	for _, id := range ids {
		out = append(out, id)
	}
	return out
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
	default:
		return 0
	}
}
