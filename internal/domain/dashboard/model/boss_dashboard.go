package model

import (
	"time"

	shared "erp/internal/shared/model"
	"github.com/shopspring/decimal"
)

type BossDashboard struct {
	TodaySales          decimal.Decimal       `json:"todaySales"`
	MonthProfit         decimal.Decimal       `json:"monthProfit"`
	PendingReceivables  decimal.Decimal       `json:"pendingReceivables"`
	InventoryValue      decimal.Decimal       `json:"inventoryValue"`
	InventoryQuantity   decimal.Decimal       `json:"inventoryQuantity"`
	InventoryAlerts     int64                 `json:"inventoryAlerts"`
	PendingRepairs      int64                 `json:"pendingRepairs"`
	ProjectProgress     []ProjectProgressItem `json:"projectProgress"`
	TodayNewCustomers   int64                 `json:"todayNewCustomers"`
	RecentOrders        []RecentOrder         `json:"recentOrders"`
	SalesRanking        []RankingItem         `json:"salesRanking"`
	EmployeePerformance []RankingItem         `json:"employeePerformance"`
	ProfitTrend         []TrendItem           `json:"profitTrend"`
	EngineerLocations   []EngineerLocation    `json:"engineerLocations"`
}

type RecentOrder struct {
	ID        uint64          `json:"id"`
	OrderNo   string          `json:"orderNo"`
	Customer  string          `json:"customer"`
	Amount    decimal.Decimal `json:"amount"`
	Status    string          `json:"status"`
	CreatedAt string          `json:"createdAt"`
}

type RankingItem struct {
	Name   string          `json:"name"`
	Value  decimal.Decimal `json:"value"`
	Count  int64           `json:"count"`
	Avatar string          `json:"avatar"`
}

type TrendItem struct {
	Date  string          `json:"date"`
	Value decimal.Decimal `json:"value"`
}

type ProjectProgressItem struct {
	ID       uint64 `json:"id"`
	Name     string `json:"name"`
	Status   string `json:"status"`
	Progress int    `json:"progress"`
}

type EngineerLocation struct {
	UserID    uint64  `json:"userId"`
	Name      string  `json:"name"`
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	Address   string  `json:"address"`
	UpdatedAt string  `json:"updatedAt"`
}

type EngineerCheckIn struct {
	shared.BaseModel
	UserID    uint64    `json:"userId" gorm:"not null;index"`
	Latitude  float64   `json:"latitude" gorm:"not null"`
	Longitude float64   `json:"longitude" gorm:"not null"`
	Address   string    `json:"address" gorm:"type:varchar(255)"`
	Device    string    `json:"device" gorm:"type:varchar(255)"`
	CheckInAt time.Time `json:"checkInAt" gorm:"not null;index"`
}

func (EngineerCheckIn) TableName() string { return "engineer_check_ins" }

type EngineerCheckInHistory struct {
	ID        uint64  `json:"id"`
	UserID    uint64  `json:"userId"`
	Name      string  `json:"name"`
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	Address   string  `json:"address"`
	Device    string  `json:"device"`
	CheckInAt string  `json:"checkInAt"`
}

type PriceProfitAnalysis struct {
	Summary  PriceProfitSummary  `json:"summary"`
	Daily    []DailyPriceProfit  `json:"daily"`
	Products []ProductProfitItem `json:"products"`
}

type PriceProfitSummary struct {
	SalesAmount        decimal.Decimal `json:"salesAmount"`
	PurchaseAmount     decimal.Decimal `json:"purchaseAmount"`
	CostAmount         decimal.Decimal `json:"costAmount"`
	ProfitAmount       decimal.Decimal `json:"profitAmount"`
	AvgSalesPrice      decimal.Decimal `json:"avgSalesPrice"`
	AvgPurchasePrice   decimal.Decimal `json:"avgPurchasePrice"`
	AvgCostPrice       decimal.Decimal `json:"avgCostPrice"`
	AvgProfitRate      decimal.Decimal `json:"avgProfitRate"`
	SalesQuantity      decimal.Decimal `json:"salesQuantity"`
	PurchaseQuantity   decimal.Decimal `json:"purchaseQuantity"`
	SalesOrderCount    int64           `json:"salesOrderCount"`
	PurchaseOrderCount int64           `json:"purchaseOrderCount"`
}

type DailyPriceProfit struct {
	Date             string          `json:"date"`
	SalesAmount      decimal.Decimal `json:"salesAmount"`
	PurchaseAmount   decimal.Decimal `json:"purchaseAmount"`
	CostAmount       decimal.Decimal `json:"costAmount"`
	ProfitAmount     decimal.Decimal `json:"profitAmount"`
	AvgSalesPrice    decimal.Decimal `json:"avgSalesPrice"`
	AvgPurchasePrice decimal.Decimal `json:"avgPurchasePrice"`
	AvgCostPrice     decimal.Decimal `json:"avgCostPrice"`
	AvgProfitRate    decimal.Decimal `json:"avgProfitRate"`
	SalesQuantity    decimal.Decimal `json:"salesQuantity"`
	PurchaseQuantity decimal.Decimal `json:"purchaseQuantity"`
}

type ProductProfitItem struct {
	ProductName   string          `json:"productName"`
	SalesAmount   decimal.Decimal `json:"salesAmount"`
	CostAmount    decimal.Decimal `json:"costAmount"`
	ProfitAmount  decimal.Decimal `json:"profitAmount"`
	AvgSalesPrice decimal.Decimal `json:"avgSalesPrice"`
	AvgCostPrice  decimal.Decimal `json:"avgCostPrice"`
	AvgProfitRate decimal.Decimal `json:"avgProfitRate"`
	SalesQuantity decimal.Decimal `json:"salesQuantity"`
	OrderCount    int64           `json:"orderCount"`
}
