package main

import (
	"fmt"
	"log"
	"time"

	"erp/internal/config"
	"erp/internal/infrastructure/database"
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
	if err := seedChineseDemo(db); err != nil {
		log.Fatal(err)
	}
	fmt.Println("中文测试数据已更新")
}

func seedChineseDemo(db *gorm.DB) error {
	return db.Transaction(func(tx *gorm.DB) error {
		now := time.Now()
		if err := seedAuthText(tx); err != nil {
			return err
		}
		if err := seedCustomers(tx, now); err != nil {
			return err
		}
		if err := seedSuppliers(tx, now); err != nil {
			return err
		}
		if err := seedProducts(tx, now); err != nil {
			return err
		}
		if err := seedInventory(tx, now); err != nil {
			return err
		}
		if err := seedBusiness(tx, now); err != nil {
			return err
		}
		return localizeBusinessStatuses(tx)
	})
}

func seedAuthText(tx *gorm.DB) error {
	statements := []struct {
		sql  string
		args []any
	}{
		{"UPDATE users SET real_name = ? WHERE tenant_id = 1 AND username = ?", []any{"系统管理员", "admin"}},
		{"UPDATE roles SET name = ? WHERE tenant_id = 1 AND code = ?", []any{"超级管理员", "super_admin"}},
		{"UPDATE system_settings SET setting_value = ? WHERE tenant_id = 1 AND setting_key = ?", []any{"ERP Pro 管理系统", "company.name"}},
	}
	for _, item := range statements {
		if err := tx.Exec(item.sql, item.args...).Error; err != nil {
			return err
		}
	}
	permissionNames := map[string]string{
		"*":                        "全部权限",
		"auth.user.manage":         "用户管理",
		"auth.user.reset_password": "重置用户密码",
		"auth.role.manage":         "角色管理",
		"auth.permission.manage":   "权限管理",
		"auth.menu.manage":         "菜单管理",
		"system.setting.manage":    "系统配置",
		"system.audit.view":        "审计日志",
		"customer.manage":          "客户管理",
		"customer.create":          "新增客户按钮",
		"customer.edit":            "编辑客户按钮",
		"customer.delete":          "删除客户按钮",
		"dashboard.boss.view":      "老板驾驶舱",
		"product.manage":           "商品管理",
		"sales.manage":             "销售管理",
		"supplier.manage":          "供应商管理",
		"purchase.manage":          "采购管理",
		"inventory.manage":         "库存管理",
		"repair.manage":            "维修管理",
		"project.manage":           "工程管理",
		"finance.manage":           "财务管理",
	}
	for code, name := range permissionNames {
		if err := tx.Exec("UPDATE permissions SET name = ? WHERE tenant_id = 1 AND code = ?", name, code).Error; err != nil {
			return err
		}
	}
	menuTitles := map[string]string{
		"boss-dashboard": "老板驾驶舱",
		"system":         "系统管理",
		"users":          "用户管理",
		"roles":          "角色管理",
		"permissions":    "权限管理",
		"menus":          "菜单管理",
		"settings":       "系统配置",
		"audit":          "审计日志",
		"customers":      "客户管理",
		"products":       "商品管理",
	}
	for name, title := range menuTitles {
		if err := tx.Exec("UPDATE menus SET title = ? WHERE tenant_id = 1 AND name = ?", title, name).Error; err != nil {
			return err
		}
	}
	return nil
}

func seedCustomers(tx *gorm.DB, now time.Time) error {
	customers := []map[string]any{
		{"code": "CUST001", "name": "杭州启明办公设备有限公司", "type": "company", "level": "重点客户", "phone": "0571-88001122", "email": "service@qiming.example", "tax_no": "91330100MA8QIMING", "address": "杭州市西湖区文三路 188 号", "credit_limit": 80000, "receivable_balance": 12500, "status": "active"},
		{"code": "CUST002", "name": "宁波蓝海网络工程有限公司", "type": "company", "level": "普通客户", "phone": "0574-66008899", "email": "it@lanhai.example", "tax_no": "91330200MA8LANHAI", "address": "宁波市鄞州区中山东路 900 号", "credit_limit": 60000, "receivable_balance": 8000, "status": "active"},
		{"code": "CUST003", "name": "绍兴越城智慧园区", "type": "company", "level": "重点客户", "phone": "0575-88996655", "email": "ops@yuecheng.example", "tax_no": "91330600MA8YUECHENG", "address": "绍兴市越城区科创路 66 号", "credit_limit": 120000, "receivable_balance": 0, "status": "active"},
	}
	for _, row := range customers {
		if err := upsertByKey(tx, "customers", "code", row, now); err != nil {
			return err
		}
	}
	if err := tx.Exec("UPDATE customers SET name = ?, level = ?, phone = ?, email = ?, address = ? WHERE tenant_id = 1 AND code LIKE ?",
		"杭州启明办公设备有限公司", "重点客户", "0571-88001122", "service@qiming.example", "杭州市西湖区文三路 188 号", "CUST-%-01").Error; err != nil {
		return err
	}
	if err := tx.Exec("UPDATE customers SET name = ?, level = ?, phone = ?, email = ?, address = ? WHERE tenant_id = 1 AND code LIKE ?",
		"宁波蓝海网络工程有限公司", "普通客户", "0574-66008899", "it@lanhai.example", "宁波市鄞州区中山东路 900 号", "CUST-%-02").Error; err != nil {
		return err
	}
	if err := tx.Exec("UPDATE customers SET name = ?, level = ?, phone = ?, email = ?, address = ? WHERE tenant_id = 1 AND code LIKE ?",
		"绍兴越城智慧园区", "重点客户", "0575-88996655", "ops@yuecheng.example", "绍兴市越城区科创路 66 号", "CUST-%-X").Error; err != nil {
		return err
	}
	if err := tx.Exec("UPDATE customers SET name = ?, address = ? WHERE tenant_id = 1 AND code = ?", "杭州测试企业", "杭州市滨江区长河路 99 号", "001").Error; err != nil {
		return err
	}
	return replaceText(tx, map[string]string{
		"Alpha Office Ltd": "杭州启明办公设备有限公司",
		"Beta Network Co":  "宁波蓝海网络工程有限公司",
		"Debug Customer":   "绍兴越城智慧园区",
	}, []tableColumn{
		{"customers", "name"},
		{"sales_orders", "customer_name"},
		{"repair_orders", "customer_name"},
		{"project_projects", "customer_name"},
	})
}

func seedSuppliers(tx *gorm.DB, now time.Time) error {
	suppliers := []map[string]any{
		{"code": "SUP001", "name": "上海联创打印设备有限公司", "contact_name": "赵经理", "phone": "021-55667788", "email": "sales@lianchuang.example", "tax_no": "91310100MA8LIANCHUANG", "address": "上海市闵行区申长路 1688 号", "payable_balance": 9500, "status": "active"},
		{"code": "SUP002", "name": "深圳安视达监控科技有限公司", "contact_name": "陈经理", "phone": "0755-88990011", "email": "order@anshida.example", "tax_no": "91440300MA8ANSHIDA", "address": "深圳市南山区科技园北区 12 栋", "payable_balance": 6200, "status": "active"},
	}
	for _, row := range suppliers {
		if err := upsertByKey(tx, "suppliers", "code", row, now); err != nil {
			return err
		}
	}
	if err := tx.Exec("UPDATE suppliers SET name = ?, contact_name = ?, phone = ?, email = ?, address = ? WHERE tenant_id = 1 AND code LIKE ?",
		"上海联创打印设备有限公司", "赵经理", "021-55667788", "sales@lianchuang.example", "上海市闵行区申长路 1688 号", "SUP-%-01").Error; err != nil {
		return err
	}
	if err := tx.Exec("UPDATE suppliers SET name = ?, contact_name = ?, phone = ?, email = ?, address = ? WHERE tenant_id = 1 AND code LIKE ?",
		"上海联创打印设备有限公司", "赵经理", "021-55667788", "sales@lianchuang.example", "上海市闵行区申长路 1688 号", "SUP-%-X").Error; err != nil {
		return err
	}
	return replaceText(tx, map[string]string{
		"Office Device Supplier":  "上海联创打印设备有限公司",
		"Network Camera Supplier": "深圳安视达监控科技有限公司",
		"Debug Supplier":          "上海联创打印设备有限公司",
		"Gamma Supplier":          "上海联创打印设备有限公司",
	}, []tableColumn{
		{"suppliers", "name"},
		{"purchase_orders", "supplier_name"},
	})
}

func seedProducts(tx *gorm.DB, now time.Time) error {
	products := []map[string]any{
		{"code": "PRD001", "name": "联想商务台式电脑 M460", "category": "电脑设备", "brand": "联想", "spec": "i5/16G/512G", "unit": "台", "barcode": "690100000001", "status": "active"},
		{"code": "PRD002", "name": "惠普黑白激光打印机 P1108", "category": "打印设备", "brand": "惠普", "spec": "A4 黑白激光", "unit": "台", "barcode": "690100000002", "status": "active"},
		{"code": "PRD003", "name": "海康 400 万网络摄像机", "category": "监控设备", "brand": "海康威视", "spec": "400万 POE", "unit": "台", "barcode": "690100000003", "status": "active"},
		{"code": "PRD004", "name": "千兆企业路由器", "category": "网络设备", "brand": "锐捷", "spec": "千兆双WAN", "unit": "台", "barcode": "690100000004", "status": "active"},
	}
	for _, row := range products {
		if err := upsertByKey(tx, "products", "code", row, now); err != nil {
			return err
		}
	}
	return replaceText(tx, map[string]string{"??POE???": "测试POE交换机"}, []tableColumn{
		{"products", "name"},
		{"purchase_orders", "product_name"},
		{"inventory_stocks", "product_name"},
		{"inventory_movements", "product_name"},
	})
}

func seedInventory(tx *gorm.DB, now time.Time) error {
	items := []map[string]any{
		{"product_code": "PRD001", "product_name": "联想商务台式电脑 M460", "warehouse": "主仓库", "quantity": 18, "avg_cost": 3600, "amount": 64800, "min_stock": 5, "status": "normal"},
		{"product_code": "PRD002", "product_name": "惠普黑白激光打印机 P1108", "warehouse": "主仓库", "quantity": 3, "avg_cost": 980, "amount": 2940, "min_stock": 5, "status": "warning"},
		{"product_code": "PRD003", "product_name": "海康 400 万网络摄像机", "warehouse": "工程仓", "quantity": 24, "avg_cost": 260, "amount": 6240, "min_stock": 10, "status": "normal"},
		{"product_code": "PRD004", "product_name": "千兆企业路由器", "warehouse": "工程仓", "quantity": 2, "avg_cost": 430, "amount": 860, "min_stock": 4, "status": "warning"},
	}
	for _, row := range items {
		if err := upsertInventory(tx, row, now); err != nil {
			return err
		}
	}
	if err := replaceText(tx, map[string]string{
		"Office Printer Model P1": "惠普黑白激光打印机 P1108",
		"Network Camera C2":       "海康 400 万网络摄像机",
		"Debug Product":           "联想商务台式电脑 M460",
	}, []tableColumn{
		{"inventory_stocks", "product_name"},
		{"inventory_movements", "product_name"},
		{"sales_orders", "product_name"},
		{"sales_order_items", "product_name"},
		{"purchase_orders", "product_name"},
	}); err != nil {
		return err
	}
	if err := tx.Exec("UPDATE inventory_stocks SET warehouse = ? WHERE tenant_id = 1 AND warehouse = ?", "主仓库", "main").Error; err != nil {
		return err
	}
	return linkProductIDs(tx)
}

func seedBusiness(tx *gorm.DB, now time.Time) error {
	if err := upsertSales(tx, now); err != nil {
		return err
	}
	if err := upsertPurchase(tx, now); err != nil {
		return err
	}
	if err := upsertRepairProjectFinance(tx, now); err != nil {
		return err
	}
	return linkProductIDs(tx)
}

func upsertSales(tx *gorm.DB, now time.Time) error {
	rows := []map[string]any{
		{"order_no": "SO-CN-001", "customer_name": "杭州启明办公设备有限公司", "product_code": "PRD002", "product_name": "惠普黑白激光打印机 P1108", "quantity": 2, "price": 1800, "cost_price": 1200, "order_date": now, "status": "已完成", "total_amount": 3600, "received_amount": 2400, "receivable_amount": 1200, "cost_amount": 2400, "profit_amount": 1200},
		{"order_no": "SO-CN-002", "customer_name": "宁波蓝海网络工程有限公司", "product_code": "PRD003", "product_name": "海康 400 万网络摄像机", "quantity": 3, "price": 680, "cost_price": 380, "order_date": now, "status": "已送货", "total_amount": 2040, "received_amount": 1000, "receivable_amount": 1040, "cost_amount": 1140, "profit_amount": 900},
	}
	for _, row := range rows {
		if err := upsertByKey(tx, "sales_orders", "order_no", row, now); err != nil {
			return err
		}
		var orderID uint64
		if err := tx.Table("sales_orders").Where("tenant_id = 1 AND order_no = ?", row["order_no"]).Select("id").Scan(&orderID).Error; err != nil {
			return err
		}
		if orderID > 0 {
			item := map[string]any{"sales_order_id": orderID, "product_name": row["product_name"], "quantity": row["quantity"], "price": row["price"], "cost_price": row["cost_price"], "amount": row["total_amount"]}
			if err := upsertSalesItem(tx, item, now); err != nil {
				return err
			}
		}
	}
	return nil
}

func linkProductIDs(tx *gorm.DB) error {
	statements := []string{
		"UPDATE sales_orders SET product_id = (SELECT id FROM products WHERE products.tenant_id = sales_orders.tenant_id AND products.code = sales_orders.product_code LIMIT 1) WHERE tenant_id = 1",
		"UPDATE purchase_orders SET product_id = (SELECT id FROM products WHERE products.tenant_id = purchase_orders.tenant_id AND products.code = purchase_orders.product_code LIMIT 1) WHERE tenant_id = 1",
		"UPDATE inventory_stocks SET product_id = (SELECT id FROM products WHERE products.tenant_id = inventory_stocks.tenant_id AND products.code = inventory_stocks.product_code LIMIT 1) WHERE tenant_id = 1",
	}
	for _, statement := range statements {
		if err := tx.Exec(statement).Error; err != nil {
			return err
		}
	}
	return nil
}

func upsertPurchase(tx *gorm.DB, now time.Time) error {
	rows := []map[string]any{
		{"order_no": "PO-CN-001", "supplier_name": "上海联创打印设备有限公司", "product_code": "PRD002", "product_name": "惠普黑白激光打印机 P1108", "quantity": 5, "price": 1180, "order_date": now, "status": "已入库", "total_amount": 5900, "paid_amount": 3000, "payable_amount": 2900},
		{"order_no": "PO-CN-002", "supplier_name": "深圳安视达监控科技有限公司", "product_code": "PRD003", "product_name": "海康 400 万网络摄像机", "quantity": 10, "price": 360, "order_date": now, "status": "待付款", "total_amount": 3600, "paid_amount": 0, "payable_amount": 3600},
	}
	for _, row := range rows {
		if err := upsertByKey(tx, "purchase_orders", "order_no", row, now); err != nil {
			return err
		}
	}
	return nil
}

func upsertRepairProjectFinance(tx *gorm.DB, now time.Time) error {
	repairs := []map[string]any{
		{"order_no": "WX-CN-001", "customer_name": "杭州启明办公设备有限公司", "device_name": "财务室打印机卡纸", "fault_desc": "连续打印时卡纸，进纸组件磨损", "repair_status": "待维修", "parts_cost": 120, "labor_cost": 180, "total_amount": 300, "registered_at": now},
		{"order_no": "WX-CN-002", "customer_name": "宁波蓝海网络工程有限公司", "device_name": "监控摄像头离线", "fault_desc": "仓库 3 号摄像头无法连接", "repair_status": "维修中", "parts_cost": 80, "labor_cost": 220, "total_amount": 300, "registered_at": now},
	}
	for _, row := range repairs {
		if err := upsertByKey(tx, "repair_orders", "order_no", row, now); err != nil {
			return err
		}
	}
	projects := []map[string]any{
		{"project_no": "GC-CN-001", "name": "越城园区监控改造工程", "customer_name": "绍兴越城智慧园区", "status": "施工中", "progress": 65, "budget_amount": 86000, "cost_amount": 42000, "settle_amount": 0, "start_date": now.AddDate(0, 0, -5)},
		{"project_no": "GC-CN-002", "name": "蓝海办公网络升级项目", "customer_name": "宁波蓝海网络工程有限公司", "status": "验收中", "progress": 90, "budget_amount": 52000, "cost_amount": 33000, "settle_amount": 48000, "start_date": now.AddDate(0, 0, -12)},
	}
	for _, row := range projects {
		if err := upsertByKey(tx, "project_projects", "project_no", row, now); err != nil {
			return err
		}
	}
	records := []map[string]any{
		{"record_no": "CW-CN-001", "record_type": "收款", "account_name": "工商银行基本户", "target_name": "杭州启明办公设备有限公司", "amount": 2400, "status": "已确认", "business_type": "销售收款", "business_no": "SO-CN-001", "occurred_at": now},
		{"record_no": "CW-CN-002", "record_type": "付款", "account_name": "工商银行基本户", "target_name": "上海联创打印设备有限公司", "amount": 3000, "status": "已确认", "business_type": "采购付款", "business_no": "PO-CN-001", "occurred_at": now},
	}
	for _, row := range records {
		if err := upsertByKey(tx, "finance_records", "record_no", row, now); err != nil {
			return err
		}
	}
	return nil
}

type tableColumn struct {
	table  string
	column string
}

func replaceText(tx *gorm.DB, replacements map[string]string, columns []tableColumn) error {
	for oldText, newText := range replacements {
		for _, col := range columns {
			sql := fmt.Sprintf("UPDATE %s SET %s = ? WHERE tenant_id = 1 AND %s = ?", col.table, col.column, col.column)
			if err := tx.Exec(sql, newText, oldText).Error; err != nil {
				return err
			}
		}
	}
	return nil
}

func upsertByKey(tx *gorm.DB, table, key string, row map[string]any, now time.Time) error {
	var count int64
	if err := tx.Table(table).Where("tenant_id = 1 AND "+key+" = ?", row[key]).Count(&count).Error; err != nil {
		return err
	}
	row["tenant_id"] = 1
	row["updated_at"] = now
	if count == 0 {
		row["created_at"] = now
		return tx.Table(table).Create(row).Error
	}
	return tx.Table(table).Where("tenant_id = 1 AND "+key+" = ?", row[key]).Updates(row).Error
}

func upsertInventory(tx *gorm.DB, row map[string]any, now time.Time) error {
	var count int64
	if err := tx.Table("inventory_stocks").
		Where("tenant_id = 1 AND product_code = ? AND warehouse = ?", row["product_code"], row["warehouse"]).
		Count(&count).Error; err != nil {
		return err
	}
	row["tenant_id"] = 1
	row["updated_at"] = now
	if _, ok := row["stock_time"]; !ok {
		row["stock_time"] = now
	}
	if count == 0 {
		row["created_at"] = now
		return tx.Table("inventory_stocks").Create(row).Error
	}
	return tx.Table("inventory_stocks").
		Where("tenant_id = 1 AND product_code = ? AND warehouse = ?", row["product_code"], row["warehouse"]).
		Updates(row).Error
}

func upsertSalesItem(tx *gorm.DB, row map[string]any, now time.Time) error {
	var count int64
	if err := tx.Table("sales_order_items").Where("tenant_id = 1 AND sales_order_id = ?", row["sales_order_id"]).Count(&count).Error; err != nil {
		return err
	}
	row["tenant_id"] = 1
	row["updated_at"] = now
	if count == 0 {
		row["created_at"] = now
		return tx.Table("sales_order_items").Create(row).Error
	}
	return tx.Table("sales_order_items").Where("tenant_id = 1 AND sales_order_id = ?", row["sales_order_id"]).Updates(row).Error
}

func localizeBusinessStatuses(tx *gorm.DB) error {
	if err := tx.Exec(`
		UPDATE sales_orders
		SET status = '草稿',
			price = CASE WHEN price = 0 AND product_code = 'PRD003' THEN 680 ELSE price END,
			total_amount = CASE WHEN total_amount = 0 AND product_code = 'PRD003' THEN quantity * 680 ELSE total_amount END,
			cost_amount = CASE WHEN cost_amount = 0 AND product_code = 'PRD003' THEN quantity * cost_price ELSE cost_amount END,
			profit_amount = CASE WHEN profit_amount = 0 AND product_code = 'PRD003' THEN quantity * (680 - cost_price) ELSE profit_amount END,
			receivable_amount = CASE WHEN receivable_amount = 0 AND product_code = 'PRD003' THEN quantity * 680 - received_amount ELSE receivable_amount END
		WHERE tenant_id = 1 AND (status = '' OR status IS NULL OR total_amount = 0)
	`).Error; err != nil {
		return err
	}
	statements := []struct {
		table  string
		column string
		from   string
		to     string
	}{
		{"customers", "status", "active", "启用"},
		{"customers", "status", "inactive", "停用"},
		{"suppliers", "status", "active", "启用"},
		{"suppliers", "status", "inactive", "停用"},
		{"sales_orders", "status", "draft", "草稿"},
		{"sales_orders", "status", "confirmed", "已确认"},
		{"sales_orders", "status", "delivered", "已送货"},
		{"sales_orders", "status", "completed", "已完成"},
		{"purchase_orders", "status", "draft", "草稿"},
		{"purchase_orders", "status", "confirmed", "已确认"},
		{"purchase_orders", "status", "completed", "已完成"},
		{"inventory_stocks", "status", "normal", "正常"},
		{"inventory_stocks", "status", "warning", "预警"},
		{"repair_orders", "repair_status", "registered", "已登记"},
		{"repair_orders", "repair_status", "in_progress", "维修中"},
		{"repair_orders", "repair_status", "completed", "已完成"},
		{"repair_orders", "repair_status", "cancelled", "已取消"},
		{"project_projects", "status", "planning", "计划中"},
		{"project_projects", "status", "construction", "施工中"},
		{"project_projects", "status", "acceptance", "验收中"},
		{"project_projects", "status", "settled", "已结算"},
		{"finance_records", "status", "confirmed", "已确认"},
	}
	for _, item := range statements {
		sql := fmt.Sprintf("UPDATE %s SET %s = ? WHERE tenant_id = 1 AND %s = ?", item.table, item.column, item.column)
		if err := tx.Exec(sql, item.to, item.from).Error; err != nil {
			return err
		}
	}
	return nil
}
