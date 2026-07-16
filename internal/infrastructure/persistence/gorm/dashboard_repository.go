package gormrepo

import (
	"context"
	"time"

	dashboardmodel "erp/internal/domain/dashboard/model"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

type BossDashboardRepository struct {
	db *gorm.DB
}

func NewBossDashboardRepository(db *gorm.DB) *BossDashboardRepository {
	return &BossDashboardRepository{db: db}
}

func (r *BossDashboardRepository) GetBossDashboard(ctx context.Context, tenantID uint64) (*dashboardmodel.BossDashboard, error) {
	result := &dashboardmodel.BossDashboard{
		TodaySales:          decimal.Zero,
		MonthProfit:         decimal.Zero,
		PendingReceivables:  decimal.Zero,
		InventoryValue:      decimal.Zero,
		InventoryQuantity:   decimal.Zero,
		ProjectProgress:     []dashboardmodel.ProjectProgressItem{},
		RecentOrders:        []dashboardmodel.RecentOrder{},
		SalesRanking:        []dashboardmodel.RankingItem{},
		EmployeePerformance: []dashboardmodel.RankingItem{},
		ProfitTrend:         emptyProfitTrend(),
		EngineerLocations:   []dashboardmodel.EngineerLocation{},
	}
	for _, fill := range []func(context.Context, uint64, *dashboardmodel.BossDashboard) error{
		r.fillSales, r.fillFinance, r.fillInventory, r.fillRepair, r.fillProject, r.fillEngineerLocations, r.fillCustomers,
	} {
		if err := fill(ctx, tenantID, result); err != nil {
			return nil, err
		}
	}
	return result, nil
}

func (r *BossDashboardRepository) GetPriceProfitAnalysis(ctx context.Context, tenantID uint64, days int) (*dashboardmodel.PriceProfitAnalysis, error) {
	if days <= 0 || days > 366 {
		days = 30
	}
	result := &dashboardmodel.PriceProfitAnalysis{
		Daily:    []dashboardmodel.DailyPriceProfit{},
		Products: []dashboardmodel.ProductProfitItem{},
	}
	if !r.db.Migrator().HasTable("sales_orders") {
		return result, nil
	}
	start := time.Now().AddDate(0, 0, -days+1).Format("2006-01-02")
	if err := r.fillPriceProfitSummary(ctx, tenantID, start, result); err != nil {
		return nil, err
	}
	if err := r.fillDailyPriceProfit(ctx, tenantID, start, result); err != nil {
		return nil, err
	}
	if err := r.fillProductProfit(ctx, tenantID, start, result); err != nil {
		return nil, err
	}
	return result, nil
}

func (r *BossDashboardRepository) SignIn(ctx context.Context, tenantID uint64, userID uint64, latitude float64, longitude float64, address string, device string) (*dashboardmodel.EngineerLocation, error) {
	now := time.Now()
	operatorID := userID
	record := &dashboardmodel.EngineerCheckIn{
		UserID:    userID,
		Latitude:  latitude,
		Longitude: longitude,
		Address:   address,
		Device:    device,
		CheckInAt: now,
	}
	record.TenantID = tenantID
	record.CreatedBy = &operatorID
	record.UpdatedBy = &operatorID
	if err := r.db.WithContext(ctx).Create(record).Error; err != nil {
		return nil, err
	}
	location := dashboardmodel.EngineerLocation{
		UserID:    userID,
		Latitude:  latitude,
		Longitude: longitude,
		Address:   address,
		UpdatedAt: now.Format("2006-01-02 15:04:05"),
	}
	if r.db.Migrator().HasTable("users") {
		_ = r.db.WithContext(ctx).Table("users").
			Select("COALESCE(NULLIF(real_name, ''), username, ?) AS name", "User").
			Where("tenant_id = ? AND id = ?", tenantID, userID).
			Scan(&location).Error
	}
	if location.Name == "" {
		location.Name = "User"
	}
	return &location, nil
}

func (r *BossDashboardRepository) ListEngineerCheckIns(ctx context.Context, tenantID uint64, userID uint64, page int, pageSize int) ([]dashboardmodel.EngineerCheckInHistory, int64, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 || pageSize > 200 {
		pageSize = 20
	}
	rows := []dashboardmodel.EngineerCheckInHistory{}
	if !r.db.Migrator().HasTable("engineer_check_ins") {
		return rows, 0, nil
	}
	query := r.db.WithContext(ctx).Table("engineer_check_ins AS e").
		Where("e.tenant_id = ? AND e.deleted_at IS NULL", tenantID)
	if userID > 0 {
		query = query.Where("e.user_id = ?", userID)
	}
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if total == 0 {
		return rows, 0, nil
	}
	err := query.
		Select(`e.id,
			e.user_id,
			COALESCE(NULLIF(u.real_name, ''), u.username, 'User') AS name,
			e.latitude,
			e.longitude,
			COALESCE(e.address, '') AS address,
			COALESCE(e.device, '') AS device,
			strftime('%Y-%m-%d %H:%M:%S', e.check_in_at) AS check_in_at`).
		Joins("LEFT JOIN users u ON u.tenant_id = e.tenant_id AND u.id = e.user_id").
		Order("e.check_in_at DESC").
		Limit(pageSize).
		Offset((page - 1) * pageSize).
		Scan(&rows).Error
	if err != nil {
		return nil, 0, err
	}
	return rows, total, nil
}

func (r *BossDashboardRepository) fillPriceProfitSummary(ctx context.Context, tenantID uint64, start string, result *dashboardmodel.PriceProfitAnalysis) error {
	if err := r.db.WithContext(ctx).Table("sales_orders").
		Select(`COALESCE(SUM(total_amount), 0) AS sales_amount,
			COALESCE(SUM(cost_amount), 0) AS cost_amount,
			COALESCE(SUM(profit_amount), 0) AS profit_amount,
			COALESCE(SUM(quantity), 0) AS sales_quantity,
			COALESCE(AVG(price), 0) AS avg_sales_price,
			COALESCE(AVG(cost_price), 0) AS avg_cost_price,
			CASE WHEN COALESCE(SUM(total_amount), 0) > 0 THEN COALESCE(SUM(profit_amount), 0) * 100.0 / SUM(total_amount) ELSE 0 END AS avg_profit_rate,
			COUNT(1) AS sales_order_count`).
		Where("tenant_id = ? AND order_date >= ?", tenantID, start).
		Scan(&result.Summary).Error; err != nil {
		return err
	}
	if !r.db.Migrator().HasTable("purchase_orders") {
		return nil
	}
	purchase := dashboardmodel.PriceProfitSummary{}
	if err := r.db.WithContext(ctx).Table("purchase_orders").
		Select(`COALESCE(SUM(total_amount), 0) AS purchase_amount,
			COALESCE(SUM(quantity), 0) AS purchase_quantity,
			COALESCE(AVG(price), 0) AS avg_purchase_price,
			COUNT(1) AS purchase_order_count`).
		Where("tenant_id = ? AND order_date >= ?", tenantID, start).
		Scan(&purchase).Error; err != nil {
		return err
	}
	result.Summary.PurchaseAmount = purchase.PurchaseAmount
	result.Summary.PurchaseQuantity = purchase.PurchaseQuantity
	result.Summary.AvgPurchasePrice = purchase.AvgPurchasePrice
	result.Summary.PurchaseOrderCount = purchase.PurchaseOrderCount
	return nil
}

func (r *BossDashboardRepository) fillDailyPriceProfit(ctx context.Context, tenantID uint64, start string, result *dashboardmodel.PriceProfitAnalysis) error {
	rows := []dashboardmodel.DailyPriceProfit{}
	if err := r.db.WithContext(ctx).Raw(`
SELECT
  COALESCE(s.date, p.date) AS date,
  COALESCE(s.sales_amount, 0) AS sales_amount,
  COALESCE(p.purchase_amount, 0) AS purchase_amount,
  COALESCE(s.cost_amount, 0) AS cost_amount,
  COALESCE(s.profit_amount, 0) AS profit_amount,
  COALESCE(s.avg_sales_price, 0) AS avg_sales_price,
  COALESCE(p.avg_purchase_price, 0) AS avg_purchase_price,
  COALESCE(s.avg_cost_price, 0) AS avg_cost_price,
  COALESCE(s.avg_profit_rate, 0) AS avg_profit_rate,
  COALESCE(s.sales_quantity, 0) AS sales_quantity,
  COALESCE(p.purchase_quantity, 0) AS purchase_quantity
FROM (
  SELECT DATE(order_date) AS date,
    SUM(total_amount) AS sales_amount,
    SUM(cost_amount) AS cost_amount,
    SUM(profit_amount) AS profit_amount,
    AVG(price) AS avg_sales_price,
    AVG(cost_price) AS avg_cost_price,
    CASE WHEN SUM(total_amount) > 0 THEN SUM(profit_amount) * 100.0 / SUM(total_amount) ELSE 0 END AS avg_profit_rate,
    SUM(quantity) AS sales_quantity
  FROM sales_orders
  WHERE tenant_id = ? AND order_date >= ?
  GROUP BY DATE(order_date)
) s
LEFT JOIN (
  SELECT DATE(order_date) AS date,
    SUM(total_amount) AS purchase_amount,
    AVG(price) AS avg_purchase_price,
    SUM(quantity) AS purchase_quantity
  FROM purchase_orders
  WHERE tenant_id = ? AND order_date >= ?
  GROUP BY DATE(order_date)
) p ON p.date = s.date
UNION
SELECT
  p.date AS date,
  0 AS sales_amount,
  p.purchase_amount,
  0 AS cost_amount,
  0 AS profit_amount,
  0 AS avg_sales_price,
  p.avg_purchase_price,
  0 AS avg_cost_price,
  0 AS avg_profit_rate,
  0 AS sales_quantity,
  p.purchase_quantity
FROM (
  SELECT DATE(order_date) AS date,
    SUM(total_amount) AS purchase_amount,
    AVG(price) AS avg_purchase_price,
    SUM(quantity) AS purchase_quantity
  FROM purchase_orders
  WHERE tenant_id = ? AND order_date >= ?
  GROUP BY DATE(order_date)
) p
LEFT JOIN (
  SELECT DATE(order_date) AS date
  FROM sales_orders
  WHERE tenant_id = ? AND order_date >= ?
  GROUP BY DATE(order_date)
) s ON s.date = p.date
WHERE s.date IS NULL
ORDER BY date ASC`, tenantID, start, tenantID, start, tenantID, start, tenantID, start).Scan(&rows).Error; err != nil {
		return err
	}
	result.Daily = rows
	return nil
}

func (r *BossDashboardRepository) fillProductProfit(ctx context.Context, tenantID uint64, start string, result *dashboardmodel.PriceProfitAnalysis) error {
	return r.db.WithContext(ctx).Table("sales_orders").
		Select(`product_name,
			COALESCE(SUM(total_amount), 0) AS sales_amount,
			COALESCE(SUM(cost_amount), 0) AS cost_amount,
			COALESCE(SUM(profit_amount), 0) AS profit_amount,
			COALESCE(AVG(price), 0) AS avg_sales_price,
			COALESCE(AVG(cost_price), 0) AS avg_cost_price,
			CASE WHEN COALESCE(SUM(total_amount), 0) > 0 THEN COALESCE(SUM(profit_amount), 0) * 100.0 / SUM(total_amount) ELSE 0 END AS avg_profit_rate,
			COALESCE(SUM(quantity), 0) AS sales_quantity,
			COUNT(1) AS order_count`).
		Where("tenant_id = ? AND order_date >= ?", tenantID, start).
		Group("product_name").
		Order("profit_amount DESC").
		Limit(20).
		Scan(&result.Products).Error
}

func (r *BossDashboardRepository) fillSales(ctx context.Context, tenantID uint64, result *dashboardmodel.BossDashboard) error {
	if !r.db.Migrator().HasTable("sales_orders") {
		return nil
	}
	today := time.Now().Format("2006-01-02")
	monthStart := time.Now().Format("2006-01") + "-01"
	if err := r.db.WithContext(ctx).Table("sales_orders").
		Select("COALESCE(SUM(total_amount), 0)").
		Where("tenant_id = ? AND DATE(order_date) = ?", tenantID, today).
		Scan(&result.TodaySales).Error; err != nil {
		return err
	}
	if err := r.db.WithContext(ctx).Table("sales_orders").
		Select("COALESCE(SUM(profit_amount), 0)").
		Where("tenant_id = ? AND order_date >= ?", tenantID, monthStart).
		Scan(&result.MonthProfit).Error; err != nil {
		return err
	}
	if err := r.db.WithContext(ctx).Table("sales_orders").
		Select("id, order_no, customer_name AS customer, total_amount AS amount, status, CAST(created_at AS TEXT) AS created_at").
		Where("tenant_id = ?", tenantID).
		Order("created_at DESC").
		Limit(6).
		Scan(&result.RecentOrders).Error; err != nil {
		return err
	}
	if r.db.Migrator().HasTable("sales_order_items") {
		if err := r.db.WithContext(ctx).Table("sales_order_items AS soi").
			Select("COALESCE(soi.product_name, 'Unknown Product') AS name, COALESCE(SUM(soi.amount), 0) AS value, COUNT(1) AS count").
			Joins("JOIN sales_orders so ON so.id = soi.sales_order_id").
			Where("so.tenant_id = ? AND so.order_date >= ?", tenantID, monthStart).
			Group("soi.product_name").
			Order("value DESC").
			Limit(5).
			Scan(&result.SalesRanking).Error; err != nil {
			return err
		}
	}
	if err := r.db.WithContext(ctx).Table("sales_orders AS so").
		Select("COALESCE(u.real_name, u.username, 'Unassigned') AS name, COALESCE(SUM(so.total_amount), 0) AS value, COUNT(1) AS count").
		Joins("LEFT JOIN users u ON u.id = so.created_by").
		Where("so.tenant_id = ? AND so.order_date >= ?", tenantID, monthStart).
		Group("u.id, u.real_name, u.username").
		Order("value DESC").
		Limit(5).
		Scan(&result.EmployeePerformance).Error; err != nil {
		return err
	}
	var trend []dashboardmodel.TrendItem
	if err := r.db.WithContext(ctx).Table("sales_orders").
		Select("DATE(order_date) AS date, COALESCE(SUM(profit_amount), 0) AS value").
		Where("tenant_id = ? AND order_date >= ?", tenantID, time.Now().AddDate(0, 0, -6).Format("2006-01-02")).
		Group("DATE(order_date)").
		Order("DATE(order_date) ASC").
		Scan(&trend).Error; err != nil {
		return err
	}
	if len(trend) > 0 {
		result.ProfitTrend = trend
	}
	return nil
}

func (r *BossDashboardRepository) fillFinance(ctx context.Context, tenantID uint64, result *dashboardmodel.BossDashboard) error {
	if !r.db.Migrator().HasTable("customers") {
		return nil
	}
	return r.db.WithContext(ctx).Table("customers").
		Select("COALESCE(SUM(receivable_balance), 0)").
		Where("tenant_id = ?", tenantID).
		Scan(&result.PendingReceivables).Error
}

func (r *BossDashboardRepository) fillInventory(ctx context.Context, tenantID uint64, result *dashboardmodel.BossDashboard) error {
	if !r.db.Migrator().HasTable("inventory_stocks") {
		return nil
	}
	if err := r.db.WithContext(ctx).Table("inventory_stocks").
		Where("tenant_id = ? AND quantity <= min_stock", tenantID).
		Count(&result.InventoryAlerts).Error; err != nil {
		return err
	}
	if r.db.Migrator().HasTable("inventory_batches") {
		values := struct {
			InventoryQuantity decimal.Decimal
			InventoryValue    decimal.Decimal
		}{}
		if err := r.db.WithContext(ctx).Table("inventory_batches").
			Select(`COALESCE(SUM(remaining_quantity), 0) AS inventory_quantity,
				COALESCE(SUM(remaining_quantity * purchase_price), 0) AS inventory_value`).
			Where("tenant_id = ? AND deleted_at IS NULL AND remaining_quantity > 0", tenantID).
			Scan(&values).Error; err != nil {
			return err
		}
		result.InventoryQuantity = values.InventoryQuantity
		result.InventoryValue = values.InventoryValue
	}
	return nil
}

func (r *BossDashboardRepository) fillRepair(ctx context.Context, tenantID uint64, result *dashboardmodel.BossDashboard) error {
	if !r.db.Migrator().HasTable("repair_orders") {
		return nil
	}
	return r.db.WithContext(ctx).Table("repair_orders").
		Where("tenant_id = ? AND repair_status NOT IN ?", tenantID, []string{"completed", "cancelled", "已完成", "已取消"}).
		Where("repair_status NOT IN ?", []string{"已完成", "已取消", "维修完成", "结算完成"}).
		Count(&result.PendingRepairs).Error
}

func (r *BossDashboardRepository) fillProject(ctx context.Context, tenantID uint64, result *dashboardmodel.BossDashboard) error {
	if !r.db.Migrator().HasTable("project_projects") {
		return nil
	}
	return r.db.WithContext(ctx).Table("project_projects").
		Select("id, name, status, progress").
		Where("tenant_id = ?", tenantID).
		Order("updated_at DESC").
		Limit(5).
		Scan(&result.ProjectProgress).Error
}

func (r *BossDashboardRepository) fillEngineerLocations(ctx context.Context, tenantID uint64, result *dashboardmodel.BossDashboard) error {
	if !r.db.Migrator().HasTable("engineer_check_ins") {
		return nil
	}
	return r.db.WithContext(ctx).Raw(`
SELECT
  e.user_id,
  COALESCE(NULLIF(u.real_name, ''), u.username, 'User') AS name,
  e.latitude,
  e.longitude,
  COALESCE(e.address, '') AS address,
  strftime('%Y-%m-%d %H:%M:%S', e.check_in_at) AS updated_at
FROM engineer_check_ins e
JOIN (
  SELECT user_id, MAX(check_in_at) AS latest_check_in_at
  FROM engineer_check_ins
  WHERE tenant_id = ? AND deleted_at IS NULL
  GROUP BY user_id
) latest ON latest.user_id = e.user_id AND latest.latest_check_in_at = e.check_in_at
LEFT JOIN users u ON u.tenant_id = e.tenant_id AND u.id = e.user_id
WHERE e.tenant_id = ? AND e.deleted_at IS NULL
ORDER BY e.check_in_at DESC
LIMIT 20`, tenantID, tenantID).Scan(&result.EngineerLocations).Error
}

func (r *BossDashboardRepository) fillCustomers(ctx context.Context, tenantID uint64, result *dashboardmodel.BossDashboard) error {
	if !r.db.Migrator().HasTable("customers") {
		return nil
	}
	today := time.Now().Format("2006-01-02")
	return r.db.WithContext(ctx).Table("customers").
		Where("tenant_id = ? AND DATE(created_at) = ?", tenantID, today).
		Count(&result.TodayNewCustomers).Error
}

func emptyProfitTrend() []dashboardmodel.TrendItem {
	result := make([]dashboardmodel.TrendItem, 0, 7)
	for i := 6; i >= 0; i-- {
		result = append(result, dashboardmodel.TrendItem{Date: time.Now().AddDate(0, 0, -i).Format("01-02"), Value: decimal.Zero})
	}
	return result
}
