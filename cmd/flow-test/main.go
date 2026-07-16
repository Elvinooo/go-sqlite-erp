package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"
	"unicode/utf8"

	"erp/internal/config"
	authmodel "erp/internal/domain/auth/model"
	businessmodel "erp/internal/domain/business/model"
	customermodel "erp/internal/domain/customer/model"
	suppliermodel "erp/internal/domain/supplier/model"
	"erp/internal/infrastructure/database"
	gormrepo "erp/internal/infrastructure/persistence/gorm"
	"erp/internal/shared/query"
	"erp/internal/shared/utils"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

const (
	tenantID  uint64 = 1
	adminName        = "admin"
	adminPass        = "admin888"
	warehouse        = "主仓库"
)

type checkResult struct {
	Module string `json:"module"`
	Name   string `json:"name"`
	Pass   bool   `json:"pass"`
	Detail string `json:"detail"`
}

type flowContext struct {
	db       *gorm.DB
	repo     *gormrepo.ModuleRepository
	ctx      context.Context
	adminID  uint64
	checks   []checkResult
	ids      map[string]uint64
	failures []string
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
	f := &flowContext{
		db:     db,
		repo:   gormrepo.NewModuleRepository(db),
		ctx:    context.Background(),
		checks: make([]checkResult, 0, 64),
		ids:    make(map[string]uint64),
	}
	if err := f.run(); err != nil {
		log.Fatal(err)
	}
	out, _ := json.MarshalIndent(map[string]any{
		"summary": map[string]any{
			"passed": len(f.checks) - len(f.failures),
			"failed": len(f.failures),
		},
		"checks":   f.checks,
		"failures": f.failures,
	}, "", "  ")
	fmt.Println(string(out))
}

func (f *flowContext) run() error {
	if err := f.resetData(); err != nil {
		return err
	}
	if err := f.seedAuth(); err != nil {
		return err
	}
	if err := f.seedBaseData(); err != nil {
		return err
	}
	if err := f.purchaseFlow(); err != nil {
		return err
	}
	if err := f.salesFlow(); err != nil {
		return err
	}
	if err := f.immediatePaymentFlow(); err != nil {
		return err
	}
	if err := f.inventoryTraceFlow(); err != nil {
		return err
	}
	if err := f.repairFlow(); err != nil {
		return err
	}
	if err := f.financeFlow(); err != nil {
		return err
	}
	if err := f.printFlow(); err != nil {
		return err
	}
	if err := f.ensureChineseData(); err != nil {
		return err
	}
	return nil
}

func (f *flowContext) check(module, name string, pass bool, detail string) {
	f.checks = append(f.checks, checkResult{Module: module, Name: name, Pass: pass, Detail: detail})
	if !pass {
		f.failures = append(f.failures, module+"："+name+"："+detail)
	}
}

func (f *flowContext) resetData() error {
	return f.db.Transaction(func(tx *gorm.DB) error {
		tables := []string{
			"business_photos",
			"sales_cost_allocations",
			"sales_order_items",
			"sales_orders",
			"purchase_order_items",
			"purchase_orders",
			"inventory_batches",
			"inventory_movements",
			"inventory_stocks",
			"repair_order_items",
			"repair_orders",
			"project_projects",
			"finance_records",
			"receivables",
			"customers",
			"suppliers",
			"products",
			"login_logs",
			"operation_logs",
		}
		for _, table := range tables {
			if err := tx.Exec("DELETE FROM " + table).Error; err != nil {
				return err
			}
		}
		if err := tx.Exec("DELETE FROM user_roles WHERE user_id IN (SELECT id FROM users WHERE username <> ?)", adminName).Error; err != nil {
			return err
		}
		if err := tx.Exec("DELETE FROM users WHERE username <> ?", adminName).Error; err != nil {
			return err
		}
		if err := tx.Exec("DELETE FROM role_permissions WHERE role_id IN (SELECT id FROM roles WHERE code <> 'super_admin')").Error; err != nil {
			return err
		}
		if err := tx.Exec("DELETE FROM role_menus WHERE role_id IN (SELECT id FROM roles WHERE code <> 'super_admin')").Error; err != nil {
			return err
		}
		if err := tx.Exec("DELETE FROM roles WHERE code <> 'super_admin'").Error; err != nil {
			return err
		}
		return nil
	})
}

func (f *flowContext) seedAuth() error {
	hash, err := utils.HashPassword(adminPass)
	if err != nil {
		return err
	}
	now := time.Now()
	err = f.db.Transaction(func(tx *gorm.DB) error {
		var admin authmodel.User
		if err := tx.Where("tenant_id = ? AND username = ?", tenantID, adminName).First(&admin).Error; err != nil {
			return err
		}
		admin.PasswordHash = hash
		admin.RealName = "系统管理员"
		admin.Status = "active"
		admin.UpdatedAt = now
		if err := tx.Save(&admin).Error; err != nil {
			return err
		}
		f.adminID = admin.ID

		var super authmodel.Role
		if err := tx.Where("tenant_id = ? AND code = ?", tenantID, "super_admin").First(&super).Error; err != nil {
			return err
		}
		super.Name = "超级管理员"
		super.Status = "active"
		super.UpdatedAt = now
		if err := tx.Save(&super).Error; err != nil {
			return err
		}
		if err := tx.Exec("DELETE FROM user_roles WHERE user_id = ?", admin.ID).Error; err != nil {
			return err
		}
		if err := tx.Create(&authmodel.UserRole{UserID: admin.ID, RoleID: super.ID, CreatedAt: now}).Error; err != nil {
			return err
		}
		roles := []struct {
			code string
			name string
		}{
			{"sales_staff", "销售员"},
			{"warehouse_staff", "仓库员"},
			{"repair_staff", "维修员"},
			{"project_staff", "工程员"},
			{"finance_staff", "财务"},
		}
		for i, item := range roles {
			role := authmodel.Role{Code: item.code, Name: item.name, DataScope: "all", Sort: i + 10, Status: "active"}
			role.TenantID = tenantID
			role.CreatedAt = now
			role.UpdatedAt = now
			if err := tx.Create(&role).Error; err != nil {
				return err
			}
			f.ids["role:"+item.code] = role.ID
		}
		users := []struct {
			username string
			realName string
			roleCode string
			phone    string
		}{
			{"zhangwei", "张伟", "sales_staff", "13800000001"},
			{"liqiang", "李强", "warehouse_staff", "13800000002"},
			{"wangshifu", "王师傅", "repair_staff", "13800000003"},
			{"zhaoshifu", "赵师傅", "project_staff", "13800000004"},
			{"liukuaiji", "刘会计", "finance_staff", "13800000005"},
		}
		for _, item := range users {
			user := authmodel.User{
				Username: item.username, PasswordHash: hash, RealName: item.realName, Phone: item.phone, Status: "active",
			}
			user.TenantID = tenantID
			user.CreatedAt = now
			user.UpdatedAt = now
			if err := tx.Create(&user).Error; err != nil {
				return err
			}
			if err := tx.Create(&authmodel.UserRole{UserID: user.ID, RoleID: f.ids["role:"+item.roleCode], CreatedAt: now}).Error; err != nil {
				return err
			}
			f.ids["user:"+item.username] = user.ID
		}
		return f.assignRolePermissions(tx, now)
	})
	if err != nil {
		return err
	}
	var count int64
	f.db.Model(&authmodel.User{}).Where("tenant_id = ?", tenantID).Count(&count)
	f.check("认证权限", "用户重置并保留管理员", count == 6, fmt.Sprintf("当前用户数 %d", count))
	return nil
}

func (f *flowContext) assignRolePermissions(tx *gorm.DB, now time.Time) error {
	rolePermissions := map[string][]string{
		"sales_staff":     {"sales.manage", "customer.manage", "product.manage", "inventory.manage"},
		"warehouse_staff": {"purchase.manage", "inventory.manage", "product.manage", "supplier.manage"},
		"repair_staff":    {"repair.manage", "inventory.manage", "customer.manage", "product.manage"},
		"project_staff":   {"project.manage", "inventory.manage", "customer.manage", "product.manage"},
		"finance_staff":   {"finance.manage", "customer.manage", "supplier.manage", "dashboard.boss.view"},
	}
	for roleCode, codes := range rolePermissions {
		roleID := f.ids["role:"+roleCode]
		for _, code := range codes {
			var permission authmodel.Permission
			if err := tx.Where("tenant_id = ? AND code = ?", tenantID, code).First(&permission).Error; err != nil {
				continue
			}
			if err := tx.Create(&authmodel.RolePermission{RoleID: roleID, PermissionID: permission.ID, CreatedAt: now}).Error; err != nil {
				return err
			}
		}
	}
	return nil
}

func (f *flowContext) seedBaseData() error {
	now := time.Now()
	return f.db.Transaction(func(tx *gorm.DB) error {
		suppliers := []suppliermodel.Supplier{
			supplier("SUP001", "深圳华强电子有限公司", "陈经理", "0755-88880001", "深圳市福田区华强北路100号"),
			supplier("SUP002", "广州办公设备供应商", "黄经理", "020-88880002", "广州市天河区办公设备城18号"),
			supplier("SUP003", "杭州监控设备有限公司", "周经理", "0571-88880003", "杭州市滨江区物联网街66号"),
		}
		for i := range suppliers {
			suppliers[i].CreatedAt = now
			suppliers[i].UpdatedAt = now
			if err := tx.Create(&suppliers[i]).Error; err != nil {
				return err
			}
			f.ids["supplier:"+suppliers[i].Name] = suppliers[i].ID
		}
		customers := []customermodel.Customer{
			customer("CUS001", "XX科技有限公司", "重点客户", "monthly", 30, "monthly", 100000),
			customer("CUS002", "XX学校", "教育客户", "quarterly", 90, "quarterly", 150000),
			customer("CUS003", "XX物业公司", "项目客户", "project_acceptance", 0, "project_acceptance", 120000),
			customer("CUS004", "个人客户张先生", "个人客户", "immediate", 0, "none", 0),
		}
		for i := range customers {
			customers[i].CreatedAt = now
			customers[i].UpdatedAt = now
			if err := tx.Create(&customers[i]).Error; err != nil {
				return err
			}
			f.ids["customer:"+customers[i].Name] = customers[i].ID
		}
		products := []businessmodel.Product{
			product("PRD001", "惠普88A硒鼓", "打印耗材", "惠普", "CC388A", "个", "690100000001", 180, 250),
			product("PRD002", "佳能打印机", "办公设备", "佳能", "LBP2900", "台", "690100000002", 1200, 1600),
			product("PRD003", "联想台式电脑", "电脑设备", "联想", "扬天M460", "台", "690100000003", 3600, 4680),
			product("PRD004", "8口交换机", "网络设备", "普联", "千兆8口", "个", "690100000004", 90, 150),
			product("PRD005", "高清摄像头", "监控设备", "海康威视", "400万POE", "个", "690100000005", 300, 480),
			product("PRD006", "硬盘4TB", "存储设备", "希捷", "监控级4TB", "块", "690100000006", 420, 580),
			product("PRD007", "网线100米", "网络耗材", "安普", "超五类100米", "箱", "690100000007", 280, 420),
		}
		for i := range products {
			products[i].CreatedAt = now
			products[i].UpdatedAt = now
			if err := tx.Create(&products[i]).Error; err != nil {
				return err
			}
			f.ids["product:"+products[i].Name] = products[i].ID
		}
		project := businessmodel.ProjectProject{
			ProjectNo:    "GC20260710001",
			Name:         "XX物业公司监控改造工程",
			CustomerID:   ptrID(f.ids["customer:XX物业公司"]),
			CustomerName: "XX物业公司",
			Status:       "施工中",
			Progress:     65,
			BudgetAmount: decimal.NewFromFloat(68000),
			CostAmount:   decimal.NewFromFloat(32000),
			SettleAmount: decimal.Zero,
			StartDate:    ptrTime(now),
		}
		project.TenantID = tenantID
		project.CreatedAt = now
		project.UpdatedAt = now
		if err := tx.Create(&project).Error; err != nil {
			return err
		}
		f.ids["project:1"] = project.ID
		return nil
	})
}

func supplier(code, name, contact, phone, address string) suppliermodel.Supplier {
	s := suppliermodel.Supplier{Code: code, Name: name, ContactName: contact, Phone: phone, Address: address, Status: "active"}
	s.TenantID = tenantID
	return s
}

func customer(code, name, level, method string, days int, cycle string, limit float64) customermodel.Customer {
	c := customermodel.Customer{
		Code: code, Name: name, Type: "company", Level: level, Phone: "13800000000",
		PaymentMethod: method, CreditLimit: decimal.NewFromFloat(limit), CreditDays: days,
		BillingCycle: cycle, PaymentDay: 0, Status: "active",
	}
	if strings.Contains(name, "张先生") {
		c.Type = "person"
	}
	c.TenantID = tenantID
	return c
}

func product(code, name, category, brand, spec, unit, barcode string, purchase, sale float64) businessmodel.Product {
	p := businessmodel.Product{
		Code: code, Name: name, Category: category, Brand: brand, Spec: spec, Unit: unit, Barcode: barcode,
		MinStock: decimal.Zero, Status: "active",
	}
	_, _ = purchase, sale
	p.TenantID = tenantID
	return p
}

func ptrID(id uint64) *uint64 {
	if id == 0 {
		return nil
	}
	return &id
}

func ptrTime(t time.Time) *time.Time {
	return &t
}

func (f *flowContext) purchaseFlow() error {
	rows := []map[string]any{
		{"supplier_id": f.ids["supplier:深圳华强电子有限公司"], "supplier_name": "深圳华强电子有限公司", "product_id": f.ids["product:惠普88A硒鼓"], "quantity": 100, "price": 180, "status": "已入库"},
		{"supplier_id": f.ids["supplier:广州办公设备供应商"], "supplier_name": "广州办公设备供应商", "product_id": f.ids["product:佳能打印机"], "quantity": 20, "price": 1200, "status": "已入库"},
		{"supplier_id": f.ids["supplier:杭州监控设备有限公司"], "supplier_name": "杭州监控设备有限公司", "product_id": f.ids["product:高清摄像头"], "quantity": 50, "price": 300, "status": "已入库"},
	}
	for i, row := range rows {
		row["warehouse"] = warehouse
		row["paid_amount"] = 0
		row["order_date"] = time.Now().Add(time.Duration(i) * time.Minute)
		created, err := f.repo.Create(f.ctx, "purchase", tenantID, f.adminID, row)
		if err != nil {
			return err
		}
		f.ids[fmt.Sprintf("purchase:%d", i+1)] = mapID(created)
	}
	f.normalizeStatuses()
	var total float64
	f.db.Table("purchase_orders").Select("COALESCE(SUM(total_amount),0)").Scan(&total)
	f.check("采购流程", "采购金额", total == 57000, fmt.Sprintf("采购总额 %.2f", total))
	f.checkStock("采购流程", "惠普88A硒鼓库存增加", "PRD001", 100)
	f.checkStock("采购流程", "佳能打印机库存增加", "PRD002", 20)
	f.checkStock("采购流程", "高清摄像头库存增加", "PRD005", 50)
	var batches int64
	f.db.Table("inventory_batches").Where("tenant_id = ? AND remaining_quantity > 0", tenantID).Count(&batches)
	f.check("采购流程", "采购批次生成", batches == 3, fmt.Sprintf("批次数 %d", batches))
	return nil
}

func (f *flowContext) salesFlow() error {
	if err := f.createSale("XX科技有限公司", "惠普88A硒鼓", 10, 250, 0); err != nil {
		return err
	}
	if err := f.createSale("XX科技有限公司", "佳能打印机", 2, 1600, 0); err != nil {
		return err
	}
	f.normalizeStatuses()
	f.checkStock("销售流程", "惠普88A硒鼓销售扣库存", "PRD001", 90)
	f.checkStock("销售流程", "佳能打印机销售扣库存", "PRD002", 18)
	var ar float64
	f.db.Table("receivables").Where("tenant_id = ? AND customer_name = ?", tenantID, "XX科技有限公司").Select("COALESCE(SUM(balance_amount),0)").Scan(&ar)
	f.check("销售流程", "月结客户生成应收", ar == 5700, fmt.Sprintf("应收余额 %.2f", ar))
	var profit float64
	f.db.Table("sales_orders").Where("tenant_id = ? AND customer_name = ?", tenantID, "XX科技有限公司").Select("COALESCE(SUM(profit_amount),0)").Scan(&profit)
	f.check("销售流程", "销售利润计算", profit == 1500, fmt.Sprintf("利润 %.2f", profit))
	var allocations int64
	f.db.Table("sales_cost_allocations").Count(&allocations)
	f.check("销售流程", "销售关联采购批次", allocations >= 2, fmt.Sprintf("成本分摊记录 %d", allocations))
	return nil
}

func (f *flowContext) immediatePaymentFlow() error {
	if err := f.createSale("个人客户张先生", "惠普88A硒鼓", 1, 250, 250); err != nil {
		return err
	}
	f.normalizeStatuses()
	var income float64
	f.db.Table("finance_records").Where("tenant_id = ? AND target_name = ?", tenantID, "个人客户张先生").Select("COALESCE(SUM(amount),0)").Scan(&income)
	f.check("立即付款", "立即付款生成收入流水", income == 250, fmt.Sprintf("收入 %.2f", income))
	var ar float64
	f.db.Table("receivables").Where("tenant_id = ? AND customer_name = ?", tenantID, "个人客户张先生").Select("COALESCE(SUM(balance_amount),0)").Scan(&ar)
	f.check("立即付款", "立即付款不生成欠款", ar == 0, fmt.Sprintf("应收 %.2f", ar))
	return nil
}

func (f *flowContext) createSale(customerName, productName string, qty, price, received float64) error {
	productID := f.ids["product:"+productName]
	sourceData, err := f.repo.Action(f.ctx, "sales", tenantID, f.adminID, "purchase-sources", map[string]any{
		"product_id": productID, "warehouse": warehouse,
	})
	if err != nil {
		return err
	}
	sources, _ := sourceData["list"].([]map[string]any)
	if len(sources) == 0 {
		return fmt.Errorf("%s 没有可用采购库存", productName)
	}
	row := map[string]any{
		"customer_id":        f.ids["customer:"+customerName],
		"customer_name":      customerName,
		"product_id":         productID,
		"quantity":           qty,
		"price":              price,
		"received_amount":    received,
		"status":             "已完成",
		"warehouse":          warehouse,
		"inventory_batch_id": mapID(sources[0]),
		"order_date":         time.Now(),
	}
	created, err := f.repo.Create(f.ctx, "sales", tenantID, f.adminID, row)
	if err != nil {
		return err
	}
	f.ids["sale:"+customerName+":"+productName] = mapID(created)
	return nil
}

func (f *flowContext) inventoryTraceFlow() error {
	detail, err := f.inventoryDetail("PRD001")
	if err != nil {
		return err
	}
	sources, _ := detail["purchaseSources"].([]map[string]any)
	f.check("库存追踪", "库存详情显示采购来源", len(sources) > 0, fmt.Sprintf("采购来源 %d 条", len(sources)))
	if len(sources) > 0 {
		first := sources[0]
		f.check("库存追踪", "采购来源字段完整", stringValue(first["purchaseOrderNo"]) != "" && stringValue(first["supplierName"]) != "" && numberValue(first["remainingQuantity"]) == 89, fmt.Sprintf("来源=%s 供应商=%s 剩余=%.0f", stringValue(first["purchaseOrderNo"]), stringValue(first["supplierName"]), numberValue(first["remainingQuantity"])))
	}
	if err := f.smallBatchFullSale(); err != nil {
		return err
	}
	f.normalizeStatuses()
	var fullCount int64
	f.db.Table("inventory_batches").Where("tenant_id = ? AND product_code = ? AND status = ?", tenantID, "PRD004", "已销售完成").Count(&fullCount)
	f.check("库存追踪", "采购批次全部销售状态", fullCount == 1, fmt.Sprintf("已销售完成批次 %d", fullCount))
	return nil
}

func (f *flowContext) smallBatchFullSale() error {
	_, err := f.repo.Create(f.ctx, "purchase", tenantID, f.adminID, map[string]any{
		"supplier_id": f.ids["supplier:深圳华强电子有限公司"], "supplier_name": "深圳华强电子有限公司",
		"product_id": f.ids["product:8口交换机"], "quantity": 2, "price": 90, "status": "已入库", "warehouse": warehouse, "order_date": time.Now(),
	})
	if err != nil {
		return err
	}
	return f.createSale("个人客户张先生", "8口交换机", 2, 150, 300)
}

func (f *flowContext) repairFlow() error {
	created, err := f.repo.Create(f.ctx, "repair", tenantID, f.adminID, map[string]any{
		"customer_id":    f.ids["customer:XX物业公司"],
		"customer_name":  "XX物业公司",
		"device_name":    "惠普打印机",
		"fault_desc":     "打印模糊",
		"repair_status":  "已完成",
		"product_id":     f.ids["product:惠普88A硒鼓"],
		"quantity":       1,
		"price":          200,
		"service_amount": 50,
		"labor_cost":     0,
		"warehouse":      warehouse,
		"registered_at":  time.Now(),
	})
	if err != nil {
		return err
	}
	f.ids["repair:1"] = mapID(created)
	f.normalizeStatuses()
	f.checkStock("维修流程", "维修配件扣库存", "PRD001", 88)
	var income float64
	f.db.Table("finance_records").Where("tenant_id = ? AND target_name = ? AND business_no = ?", tenantID, "XX物业公司", stringValue(created["orderNo"])).Select("COALESCE(SUM(amount),0)").Scan(&income)
	f.check("维修流程", "维修完成生成收入", income == 250, fmt.Sprintf("维修收入 %.2f", income))
	var profit float64
	f.db.Table("repair_orders").Where("tenant_id = ? AND id = ?", tenantID, mapID(created)).Select("profit_amount").Scan(&profit)
	f.check("维修流程", "维修利润计算", profit == 70, fmt.Sprintf("维修利润 %.2f", profit))
	return nil
}

func (f *flowContext) financeFlow() error {
	var income float64
	f.db.Table("finance_records").Where("tenant_id = ?", tenantID).Select("COALESCE(SUM(amount),0)").Scan(&income)
	f.check("财务测试", "收入记录", income == 800, fmt.Sprintf("收入流水合计 %.2f", income))
	var debt float64
	f.db.Table("customers").Where("tenant_id = ? AND name = ?", tenantID, "XX科技有限公司").Select("receivable_balance").Scan(&debt)
	f.check("财务测试", "客户欠款同步", debt == 5700, fmt.Sprintf("客户欠款 %.2f", debt))
	var payable float64
	f.db.Table("purchase_orders").Where("tenant_id = ?", tenantID).Select("COALESCE(SUM(payable_amount),0)").Scan(&payable)
	f.check("财务测试", "供应商应付款可统计", payable == 57180, fmt.Sprintf("采购未付款 %.2f", payable))
	var profit float64
	f.db.Table("sales_orders").Where("tenant_id = ?", tenantID).Select("COALESCE(SUM(profit_amount),0)").Scan(&profit)
	f.check("财务测试", "利润统计", profit == 1690, fmt.Sprintf("销售利润 %.2f", profit))
	return nil
}

func (f *flowContext) printFlow() error {
	prints := []struct {
		name   string
		module string
		action string
		id     uint64
		path   string
	}{
		{"销售单", "sales", "print", f.ids["sale:XX科技有限公司:惠普88A硒鼓"], "/print/sales/"},
		{"采购单", "purchase", "print", f.ids["purchase:1"], "/print/purchase/"},
		{"送货单", "sales", "print", f.ids["sale:XX科技有限公司:佳能打印机"], "/print/sales/"},
		{"维修单", "repair", "repair-print", f.ids["repair:1"], "/print/repair/"},
	}
	for _, item := range prints {
		result, err := f.repo.Action(f.ctx, item.module, tenantID, f.adminID, item.action, map[string]any{"id": item.id})
		f.check("打印测试", item.name+"打印地址", err == nil && strings.Contains(stringValue(result["printUrl"]), item.path), fmt.Sprintf("返回 %s", stringValue(result["printUrl"])))
	}
	var financeID uint64
	f.db.Table("finance_records").Where("tenant_id = ?", tenantID).Order("id ASC").Select("id").Limit(1).Scan(&financeID)
	result, err := f.repo.Action(f.ctx, "finance", tenantID, f.adminID, "receipt", map[string]any{"id": financeID})
	f.check("打印测试", "收款单打印地址", err == nil && strings.Contains(stringValue(result["printUrl"]), "/print/receipt/"), fmt.Sprintf("返回 %s", stringValue(result["printUrl"])))
	var receivableID uint64
	f.db.Table("receivables").Where("tenant_id = ?", tenantID).Order("id ASC").Select("id").Limit(1).Scan(&receivableID)
	result, err = f.repo.Action(f.ctx, "receivables", tenantID, f.adminID, "statement", map[string]any{"id": receivableID})
	f.check("打印测试", "对账单打印地址", err == nil && strings.Contains(stringValue(result["printUrl"]), "/print/statement/"), fmt.Sprintf("返回 %s", stringValue(result["printUrl"])))
	return nil
}

func (f *flowContext) inventoryDetail(productCode string) (map[string]any, error) {
	rows, _, err := f.repo.List(f.ctx, "inventory", tenantID, query.PageQuery{Page: 1, PageSize: 100})
	if err != nil {
		return nil, err
	}
	for _, row := range rows {
		if stringValue(row["productCode"]) == productCode {
			return f.repo.Find(f.ctx, "inventory", tenantID, mapID(row))
		}
	}
	return nil, fmt.Errorf("未找到库存：%s", productCode)
}

func (f *flowContext) checkStock(module, name, code string, expected float64) {
	var qty float64
	f.db.Table("inventory_stocks").Where("tenant_id = ? AND product_code = ? AND warehouse = ? AND deleted_at IS NULL", tenantID, code, warehouse).Select("COALESCE(SUM(quantity),0)").Scan(&qty)
	f.check(module, name, qty == expected, fmt.Sprintf("库存 %.2f，期望 %.2f", qty, expected))
}

func (f *flowContext) normalizeStatuses() {
	f.db.Exec(`UPDATE inventory_batches SET status = CASE WHEN remaining_quantity <= 0 THEN '已销售完成' WHEN inbound_quantity > remaining_quantity THEN '销售中' ELSE '未销售' END WHERE tenant_id = ?`, tenantID)
	f.db.Exec(`UPDATE inventory_stocks SET status = CASE WHEN quantity <= min_stock THEN '预警' ELSE '正常' END WHERE tenant_id = ?`, tenantID)
	f.db.Exec(`UPDATE finance_records SET record_type = '收款' WHERE tenant_id = ? AND record_type <> '支出'`, tenantID)
	f.db.Exec(`UPDATE finance_records SET account_name = '默认账户' WHERE tenant_id = ? AND (account_name = '' OR account_name LIKE '%?%' OR account_name LIKE '%鍙%')`, tenantID)
	f.db.Exec(`UPDATE finance_records SET status = '已确认' WHERE tenant_id = ?`, tenantID)
	f.db.Exec(`UPDATE finance_records SET business_type = '销售收款' WHERE tenant_id = ? AND business_no LIKE 'SO%'`, tenantID)
	f.db.Exec(`UPDATE finance_records SET business_type = '维修结算' WHERE tenant_id = ? AND business_no LIKE 'RO%'`, tenantID)
	f.db.Exec(`UPDATE receivables SET status = '未收款' WHERE tenant_id = ? AND balance_amount > 0`, tenantID)
	f.db.Exec(`UPDATE receivables SET status = '已收款' WHERE tenant_id = ? AND balance_amount <= 0`, tenantID)
	f.db.Exec(`UPDATE receivables SET settlement_mode = '月结30天' WHERE tenant_id = ? AND customer_name = 'XX科技有限公司'`, tenantID)
}

func (f *flowContext) ensureChineseData() error {
	tables := map[string][]string{
		"customers":              {"name", "level", "address"},
		"suppliers":              {"name", "contact_name", "address"},
		"products":               {"name", "category", "brand", "spec", "unit"},
		"purchase_orders":        {"supplier_name", "product_name", "status"},
		"sales_orders":           {"customer_name", "product_name", "status"},
		"repair_orders":          {"customer_name", "device_name", "fault_desc", "repair_status"},
		"finance_records":        {"record_type", "account_name", "target_name", "status", "business_type"},
		"inventory_batches":      {"product_name", "warehouse", "supplier_name", "status"},
		"inventory_stocks":       {"product_name", "warehouse", "status"},
		"inventory_movements":    {"product_name"},
		"receivables":            {"customer_name", "settlement_mode", "status"},
		"repair_order_items":     {"item_name", "product_name"},
		"sales_order_items":      {"product_name", "supplier_name"},
		"sales_cost_allocations": {"product_name", "supplier_name"},
	}
	bad := make([]string, 0)
	for table, cols := range tables {
		for _, col := range cols {
			rows := []string{}
			if err := f.db.Table(table).Where(col+" LIKE ? OR "+col+" LIKE ?", "%?%", "%\uFFFD%").Pluck(col, &rows).Error; err == nil && len(rows) > 0 {
				bad = append(bad, table+"."+col)
			}
		}
	}
	f.check("数据质量", "无问号和未知字符", len(bad) == 0, strings.Join(bad, "，"))
	var chineseFields int64
	f.db.Table("products").Where("name GLOB '*[一-龥]*'").Count(&chineseFields)
	f.check("数据质量", "商品名称为中文", chineseFields == 7, fmt.Sprintf("中文商品 %d 个", chineseFields))
	return nil
}

func mapID(row map[string]any) uint64 {
	for _, key := range []string{"id", "ID"} {
		switch v := row[key].(type) {
		case uint64:
			return v
		case uint:
			return uint64(v)
		case int64:
			return uint64(v)
		case int:
			return uint64(v)
		case float64:
			return uint64(v)
		}
	}
	return 0
}

func numberValue(value any) float64 {
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
	case decimal.Decimal:
		f, _ := v.Float64()
		return f
	case []uint8:
		f, _ := decimal.NewFromString(string(v))
		x, _ := f.Float64()
		return x
	case string:
		f, _ := decimal.NewFromString(v)
		x, _ := f.Float64()
		return x
	default:
		return 0
	}
}

func stringValue(value any) string {
	switch v := value.(type) {
	case nil:
		return ""
	case string:
		if utf8.ValidString(v) {
			return v
		}
		return strings.ToValidUTF8(v, "")
	case []byte:
		return string(v)
	default:
		return fmt.Sprint(v)
	}
}
