package model

import (
	"time"

	"github.com/shopspring/decimal"

	shared "erp/internal/shared/model"
)

type Product struct {
	shared.BaseModel
	Code     string          `json:"code" gorm:"type:varchar(64);not null;index"`
	Name     string          `json:"name" gorm:"type:varchar(128);not null;index"`
	Category string          `json:"category" gorm:"type:varchar(64);index"`
	Brand    string          `json:"brand" gorm:"type:varchar(64);index"`
	Spec     string          `json:"spec" gorm:"type:varchar(128)"`
	Unit     string          `json:"unit" gorm:"type:varchar(32);not null;default:台"`
	Barcode  string          `json:"barcode" gorm:"type:varchar(128);index"`
	QRCode   string          `json:"qrCode" gorm:"type:varchar(255);index"`
	ImageURL string          `json:"imageUrl" gorm:"type:varchar(500)"`
	MinStock decimal.Decimal `json:"minStock" gorm:"type:decimal(18,4);not null;default:0"`
	Status   string          `json:"status" gorm:"type:varchar(32);not null;default:active;index"`
}

func (Product) TableName() string { return "products" }

type SalesOrder struct {
	shared.BaseModel
	OrderNo          string          `json:"orderNo" gorm:"type:varchar(64);not null;index"`
	CustomerID       *uint64         `json:"customerId" gorm:"index"`
	CustomerName     string          `json:"customerName" gorm:"type:varchar(128);index"`
	ProductID        *uint64         `json:"productId" gorm:"index"`
	ProductCode      string          `json:"productCode" gorm:"type:varchar(64);index"`
	ProductName      string          `json:"productName" gorm:"type:varchar(128);index"`
	Quantity         decimal.Decimal `json:"quantity" gorm:"type:decimal(18,4);not null;default:0"`
	Price            decimal.Decimal `json:"price" gorm:"type:decimal(18,4);not null;default:0"`
	CostPrice        decimal.Decimal `json:"costPrice" gorm:"type:decimal(18,4);not null;default:0"`
	OrderDate        time.Time       `json:"orderDate" gorm:"not null;index"`
	Status           string          `json:"status" gorm:"type:varchar(32);not null;default:draft;index"`
	TotalAmount      decimal.Decimal `json:"totalAmount" gorm:"type:decimal(18,4);not null;default:0"`
	ReceivedAmount   decimal.Decimal `json:"receivedAmount" gorm:"type:decimal(18,4);not null;default:0"`
	ReceivableAmount decimal.Decimal `json:"receivableAmount" gorm:"type:decimal(18,4);not null;default:0"`
	CostAmount       decimal.Decimal `json:"costAmount" gorm:"type:decimal(18,4);not null;default:0"`
	ProfitAmount     decimal.Decimal `json:"profitAmount" gorm:"type:decimal(18,4);not null;default:0;index"`
	ProfitRate       decimal.Decimal `json:"profitRate" gorm:"type:decimal(10,4);not null;default:0;index"`
	CostMethod       string          `json:"costMethod" gorm:"type:varchar(32);not null;default:purchase_source;index"`
	InventoryBatchID *uint64         `json:"inventoryBatchId" gorm:"index"`
	PurchaseOrderID  *uint64         `json:"purchaseOrderId" gorm:"index"`
	PurchaseOrderNo  string          `json:"purchaseOrderNo" gorm:"type:varchar(64);index"`
	SupplierID       *uint64         `json:"supplierId" gorm:"index"`
	SupplierName     string          `json:"supplierName" gorm:"type:varchar(128);index"`
	PurchaseDate     *time.Time      `json:"purchaseDate" gorm:"index"`
	PurchasePrice    decimal.Decimal `json:"purchasePrice" gorm:"type:decimal(18,4);not null;default:0"`
	DeletedBy        *uint64         `json:"deletedBy" gorm:"index"`
	DeleteReason     string          `json:"deleteReason" gorm:"type:varchar(500)"`
}

func (SalesOrder) TableName() string { return "sales_orders" }

type SalesOrderDeletion struct {
	ID              uint64    `json:"id" gorm:"primaryKey;autoIncrement"`
	TenantID        uint64    `json:"tenantId" gorm:"not null;default:1;index"`
	OriginalOrderID uint64    `json:"originalOrderId" gorm:"not null;index"`
	OrderNo         string    `json:"orderNo" gorm:"type:varchar(64);not null;index"`
	CustomerID      *uint64   `json:"customerId" gorm:"index"`
	CustomerName    string    `json:"customerName" gorm:"type:varchar(128);index"`
	ItemsJSON       string    `json:"itemsJson" gorm:"type:text;not null"`
	DeletedBy       uint64    `json:"deletedBy" gorm:"not null;index"`
	DeletedAt       time.Time `json:"deletedAt" gorm:"not null;index"`
	DeleteReason    string    `json:"deleteReason" gorm:"type:varchar(500);not null"`
	SnapshotJSON    string    `json:"snapshotJson" gorm:"type:text;not null"`
	CreatedAt       time.Time `json:"createdAt" gorm:"not null;index"`
}

func (SalesOrderDeletion) TableName() string { return "sales_order_deletions" }

type DocumentDeleteRecord struct {
	shared.BaseModel
	DocumentType     string     `json:"documentType" gorm:"type:varchar(32);not null;index"`
	DocumentID       uint64     `json:"documentId" gorm:"not null;index"`
	DocumentNo       string     `json:"documentNo" gorm:"type:varchar(64);not null;index"`
	DeleteReason     string     `json:"deleteReason" gorm:"type:varchar(500);not null"`
	DeleteUserID     uint64     `json:"deleteUserId" gorm:"not null;index"`
	DeleteUserName   string     `json:"deleteUserName" gorm:"type:varchar(64);index"`
	DeleteTime       time.Time  `json:"deleteTime" gorm:"not null;index"`
	DeleteStatus     string     `json:"deleteStatus" gorm:"type:varchar(32);not null;default:WAITING;index"`
	BeforeData       string     `json:"beforeData" gorm:"type:text;not null"`
	StockProcessed   bool       `json:"stockProcessed" gorm:"not null;default:false;index"`
	FinanceProcessed bool       `json:"financeProcessed" gorm:"not null;default:false;index"`
	IPAddress        string     `json:"ipAddress" gorm:"type:varchar(64)"`
	CompletedAt      *time.Time `json:"completedAt" gorm:"index"`
	FailedReason     string     `json:"failedReason" gorm:"type:varchar(500)"`
}

func (DocumentDeleteRecord) TableName() string { return "document_delete_records" }

type DocumentDeleteDetail struct {
	shared.BaseModel
	RecordID     uint64  `json:"recordId" gorm:"not null;index"`
	DetailType   string  `json:"detailType" gorm:"type:varchar(32);not null;index"`
	SKUID        *uint64 `json:"skuId" gorm:"column:sku_id;index"`
	SKUCode      string  `json:"skuCode" gorm:"column:sku_code;type:varchar(64);index"`
	SKUName      string  `json:"skuName" gorm:"column:sku_name;type:varchar(128);index"`
	WarehouseID  *uint64 `json:"warehouseId" gorm:"column:warehouse_id;index"`
	Warehouse    string  `json:"warehouse" gorm:"type:varchar(64);index"`
	Quantity     float64 `json:"qty" gorm:"type:decimal(18,4);not null;default:0"`
	StockChange  float64 `json:"stockChange" gorm:"type:decimal(18,4);not null;default:0"`
	FinanceType  string  `json:"financeType" gorm:"type:varchar(32);index"`
	FinanceNo    string  `json:"financeNo" gorm:"type:varchar(64);index"`
	Amount       float64 `json:"amount" gorm:"type:decimal(18,4);not null;default:0"`
	BusinessType string  `json:"businessType" gorm:"type:varchar(32);index"`
	BusinessID   *uint64 `json:"businessId" gorm:"index"`
	BusinessNo   string  `json:"businessNo" gorm:"type:varchar(64);index"`
}

func (DocumentDeleteDetail) TableName() string { return "document_delete_details" }

type SalesOrderItem struct {
	shared.BaseModel
	SalesOrderID        uint64          `json:"salesOrderId" gorm:"not null;index"`
	ProductID           *uint64         `json:"productId" gorm:"index"`
	ProductCode         string          `json:"productCode" gorm:"type:varchar(64);index"`
	ProductName         string          `json:"productName" gorm:"type:varchar(128);not null;index"`
	Quantity            decimal.Decimal `json:"quantity" gorm:"type:decimal(18,4);not null;default:0"`
	Price               decimal.Decimal `json:"price" gorm:"type:decimal(18,4);not null;default:0"`
	Amount              decimal.Decimal `json:"amount" gorm:"type:decimal(18,4);not null;default:0"`
	PurchaseSource      string          `json:"purchaseSource" gorm:"type:varchar(32);not null;default:purchase_order;index"`
	InventoryBatchID    *uint64         `json:"inventoryBatchId" gorm:"index"`
	PurchaseOrderID     *uint64         `json:"purchaseOrderId" gorm:"index"`
	PurchaseOrderItemID *uint64         `json:"purchaseOrderItemId" gorm:"index"`
	PurchaseOrderNo     string          `json:"purchaseOrderNo" gorm:"type:varchar(64);index"`
	SupplierID          *uint64         `json:"supplierId" gorm:"index"`
	SupplierName        string          `json:"supplierName" gorm:"type:varchar(128);index"`
	PurchaseDate        *time.Time      `json:"purchaseDate" gorm:"index"`
	PurchasePrice       decimal.Decimal `json:"purchasePrice" gorm:"type:decimal(18,4);not null;default:0"`
	CostPrice           decimal.Decimal `json:"costPrice" gorm:"type:decimal(18,4);not null;default:0"`
	CostAmount          decimal.Decimal `json:"costAmount" gorm:"type:decimal(18,4);not null;default:0"`
	ProfitAmount        decimal.Decimal `json:"profitAmount" gorm:"type:decimal(18,4);not null;default:0"`
}

func (SalesOrderItem) TableName() string { return "sales_order_items" }

type SalesCostAllocation struct {
	shared.BaseModel
	SalesOrderID     uint64          `json:"salesOrderId" gorm:"not null;index"`
	SalesOrderItemID *uint64         `json:"salesOrderItemId" gorm:"index"`
	InventoryBatchID *uint64         `json:"inventoryBatchId" gorm:"index"`
	ProductID        *uint64         `json:"productId" gorm:"index"`
	ProductCode      string          `json:"productCode" gorm:"type:varchar(64);index"`
	ProductName      string          `json:"productName" gorm:"type:varchar(128);not null;index"`
	Quantity         decimal.Decimal `json:"quantity" gorm:"type:decimal(18,4);not null;default:0"`
	CostPrice        decimal.Decimal `json:"costPrice" gorm:"type:decimal(18,4);not null;default:0"`
	CostAmount       decimal.Decimal `json:"costAmount" gorm:"type:decimal(18,4);not null;default:0"`
	PurchaseOrderID  *uint64         `json:"purchaseOrderId" gorm:"index"`
	PurchaseOrderNo  string          `json:"purchaseOrderNo" gorm:"type:varchar(64);index"`
	SupplierID       *uint64         `json:"supplierId" gorm:"index"`
	SupplierName     string          `json:"supplierName" gorm:"type:varchar(128);index"`
	PurchaseDate     *time.Time      `json:"purchaseDate" gorm:"index"`
	PurchasePrice    decimal.Decimal `json:"purchasePrice" gorm:"type:decimal(18,4);not null;default:0"`
}

func (SalesCostAllocation) TableName() string { return "sales_cost_allocations" }

type PurchaseOrder struct {
	shared.BaseModel
	OrderNo       string          `json:"orderNo" gorm:"type:varchar(64);not null;index"`
	SupplierID    *uint64         `json:"supplierId" gorm:"index"`
	SupplierName  string          `json:"supplierName" gorm:"type:varchar(128);index"`
	ProductID     *uint64         `json:"productId" gorm:"index"`
	ProductCode   string          `json:"productCode" gorm:"type:varchar(64);index"`
	ProductName   string          `json:"productName" gorm:"type:varchar(128);index"`
	Quantity      decimal.Decimal `json:"quantity" gorm:"type:decimal(18,4);not null;default:0"`
	Price         decimal.Decimal `json:"price" gorm:"type:decimal(18,4);not null;default:0"`
	OrderDate     time.Time       `json:"orderDate" gorm:"not null;index"`
	Status        string          `json:"status" gorm:"type:varchar(32);not null;default:draft;index"`
	TotalAmount   decimal.Decimal `json:"totalAmount" gorm:"type:decimal(18,4);not null;default:0"`
	PaidAmount    decimal.Decimal `json:"paidAmount" gorm:"type:decimal(18,4);not null;default:0"`
	PayableAmount decimal.Decimal `json:"payableAmount" gorm:"type:decimal(18,4);not null;default:0"`
	DeletedBy     *uint64         `json:"deletedBy" gorm:"index"`
	DeleteReason  string          `json:"deleteReason" gorm:"type:varchar(500)"`
}

func (PurchaseOrder) TableName() string { return "purchase_orders" }

type PurchaseOrderItem struct {
	shared.BaseModel
	PurchaseOrderID uint64          `json:"purchaseOrderId" gorm:"not null;index"`
	ProductID       *uint64         `json:"productId" gorm:"index"`
	ProductCode     string          `json:"productCode" gorm:"type:varchar(64);index"`
	ProductName     string          `json:"productName" gorm:"type:varchar(128);not null;index"`
	Quantity        decimal.Decimal `json:"quantity" gorm:"type:decimal(18,4);not null;default:0"`
	Price           decimal.Decimal `json:"price" gorm:"type:decimal(18,4);not null;default:0"`
	Amount          decimal.Decimal `json:"amount" gorm:"type:decimal(18,4);not null;default:0"`
	Warehouse       string          `json:"warehouse" gorm:"type:varchar(64);not null;default:主仓库;index"`
}

func (PurchaseOrderItem) TableName() string { return "purchase_order_items" }

type InventoryBatch struct {
	shared.BaseModel
	BatchNo             string          `json:"batchNo" gorm:"type:varchar(64);not null;index"`
	ProductID           *uint64         `json:"productId" gorm:"index"`
	ProductCode         string          `json:"productCode" gorm:"type:varchar(64);index"`
	ProductName         string          `json:"productName" gorm:"type:varchar(128);not null;index"`
	Warehouse           string          `json:"warehouse" gorm:"type:varchar(64);not null;default:主仓库;index"`
	PurchaseOrderID     *uint64         `json:"purchaseOrderId" gorm:"index"`
	PurchaseOrderItemID *uint64         `json:"purchaseOrderItemId" gorm:"index"`
	PurchaseOrderNo     string          `json:"purchaseOrderNo" gorm:"type:varchar(64);index"`
	SupplierID          *uint64         `json:"supplierId" gorm:"index"`
	SupplierName        string          `json:"supplierName" gorm:"type:varchar(128);index"`
	PurchaseDate        time.Time       `json:"purchaseDate" gorm:"not null;index"`
	PurchasePrice       decimal.Decimal `json:"purchasePrice" gorm:"type:decimal(18,4);not null;default:0"`
	InboundQuantity     decimal.Decimal `json:"inboundQuantity" gorm:"type:decimal(18,4);not null;default:0"`
	RemainingQuantity   decimal.Decimal `json:"remainingQuantity" gorm:"type:decimal(18,4);not null;default:0;index"`
	Status              string          `json:"status" gorm:"type:varchar(32);not null;default:available;index"`
}

func (InventoryBatch) TableName() string { return "inventory_batches" }

type InventoryStock struct {
	shared.BaseModel
	ProductID   *uint64         `json:"productId" gorm:"index"`
	ProductCode string          `json:"productCode" gorm:"type:varchar(64);index"`
	ProductName string          `json:"productName" gorm:"type:varchar(128);not null;index"`
	Warehouse   string          `json:"warehouse" gorm:"type:varchar(64);not null;default:main;index"`
	Quantity    decimal.Decimal `json:"quantity" gorm:"type:decimal(18,4);not null;default:0"`
	AvgCost     decimal.Decimal `json:"avgCost" gorm:"type:decimal(18,4);not null;default:0"`
	Amount      decimal.Decimal `json:"amount" gorm:"type:decimal(18,4);not null;default:0"`
	MinStock    decimal.Decimal `json:"minStock" gorm:"type:decimal(18,4);not null;default:0;index"`
	Status      string          `json:"status" gorm:"type:varchar(32);not null;default:normal;index"`
	StockTime   *time.Time      `json:"stockTime" gorm:"index"`
}

func (InventoryStock) TableName() string { return "inventory_stocks" }

type InventoryMovement struct {
	shared.BaseModel
	MovementNo       string          `json:"movementNo" gorm:"type:varchar(64);not null;index"`
	ProductID        *uint64         `json:"productId" gorm:"index"`
	InventoryBatchID *uint64         `json:"inventoryBatchId" gorm:"index"`
	ProductCode      string          `json:"productCode" gorm:"type:varchar(64);index"`
	ProductName      string          `json:"productName" gorm:"type:varchar(128);not null;index"`
	Warehouse        string          `json:"warehouse" gorm:"type:varchar(64);index"`
	SourceType       string          `json:"sourceType" gorm:"type:varchar(32);not null;index"`
	SourceID         *uint64         `json:"sourceId" gorm:"index"`
	Direction        string          `json:"direction" gorm:"type:varchar(16);not null;index"`
	Quantity         decimal.Decimal `json:"quantity" gorm:"type:decimal(18,4);not null;default:0"`
	BeforeQuantity   decimal.Decimal `json:"beforeQuantity" gorm:"type:decimal(18,4);not null;default:0"`
	AfterQuantity    decimal.Decimal `json:"afterQuantity" gorm:"type:decimal(18,4);not null;default:0"`
	UnitCost         decimal.Decimal `json:"unitCost" gorm:"type:decimal(18,4);not null;default:0"`
	Amount           decimal.Decimal `json:"amount" gorm:"type:decimal(18,4);not null;default:0"`
	OperatorName     string          `json:"operatorName" gorm:"type:varchar(64);index"`
	OccurredAt       time.Time       `json:"occurredAt" gorm:"not null;index"`
}

func (InventoryMovement) TableName() string { return "inventory_movements" }

type RepairOrder struct {
	shared.BaseModel
	OrderNo         string          `json:"orderNo" gorm:"type:varchar(64);not null;index"`
	CustomerID      *uint64         `json:"customerId" gorm:"index"`
	CustomerName    string          `json:"customerName" gorm:"type:varchar(128);index"`
	DeviceName      string          `json:"deviceName" gorm:"type:varchar(128);not null;index"`
	FaultDesc       string          `json:"faultDesc" gorm:"type:text"`
	RepairStatus    string          `json:"repairStatus" gorm:"type:varchar(32);not null;default:registered;index"`
	ProductID       *uint64         `json:"productId" gorm:"index"`
	ProductCode     string          `json:"productCode" gorm:"type:varchar(64);index"`
	ProductName     string          `json:"productName" gorm:"type:varchar(128);index"`
	Quantity        decimal.Decimal `json:"quantity" gorm:"type:decimal(18,4);not null;default:0"`
	Price           decimal.Decimal `json:"price" gorm:"type:decimal(18,4);not null;default:0"`
	CostPrice       decimal.Decimal `json:"costPrice" gorm:"type:decimal(18,4);not null;default:0"`
	ServiceAmount   decimal.Decimal `json:"serviceAmount" gorm:"type:decimal(18,4);not null;default:0"`
	OnsiteFee       decimal.Decimal `json:"onsiteFee" gorm:"type:decimal(18,4);not null;default:0"`
	DetectionFee    decimal.Decimal `json:"detectionFee" gorm:"type:decimal(18,4);not null;default:0"`
	InstallationFee decimal.Decimal `json:"installationFee" gorm:"type:decimal(18,4);not null;default:0"`
	PartsAmount     decimal.Decimal `json:"partsAmount" gorm:"type:decimal(18,4);not null;default:0"`
	PartsCost       decimal.Decimal `json:"partsCost" gorm:"type:decimal(18,4);not null;default:0"`
	OutsourceCost   decimal.Decimal `json:"outsourceCost" gorm:"type:decimal(18,4);not null;default:0"`
	LaborCost       decimal.Decimal `json:"laborCost" gorm:"type:decimal(18,4);not null;default:0"`
	TransportCost   decimal.Decimal `json:"transportCost" gorm:"type:decimal(18,4);not null;default:0"`
	OtherCost       decimal.Decimal `json:"otherCost" gorm:"type:decimal(18,4);not null;default:0"`
	TotalAmount     decimal.Decimal `json:"totalAmount" gorm:"type:decimal(18,4);not null;default:0"`
	CostAmount      decimal.Decimal `json:"costAmount" gorm:"type:decimal(18,4);not null;default:0"`
	ProfitAmount    decimal.Decimal `json:"profitAmount" gorm:"type:decimal(18,4);not null;default:0;index"`
	InventoryDone   bool            `json:"inventoryDone" gorm:"not null;default:false;index"`
	SettlementDone  bool            `json:"settlementDone" gorm:"not null;default:false;index"`
	RegisteredAt    time.Time       `json:"registeredAt" gorm:"not null;index"`
}

func (RepairOrder) TableName() string { return "repair_orders" }

type RepairOrderItem struct {
	shared.BaseModel
	RepairOrderID    uint64          `json:"repairOrderId" gorm:"not null;index"`
	ItemType         string          `json:"itemType" gorm:"type:varchar(32);not null;index"`
	ItemName         string          `json:"itemName" gorm:"type:varchar(128);not null;index"`
	SupplierID       *uint64         `json:"supplierId" gorm:"index"`
	SupplierName     string          `json:"supplierName" gorm:"type:varchar(128);index"`
	ServiceProject   string          `json:"serviceProject" gorm:"type:varchar(128);index"`
	ProductID        *uint64         `json:"productId" gorm:"index"`
	InventoryBatchID *uint64         `json:"inventoryBatchId" gorm:"index"`
	ProductCode      string          `json:"productCode" gorm:"type:varchar(64);index"`
	ProductName      string          `json:"productName" gorm:"type:varchar(128);index"`
	Quantity         decimal.Decimal `json:"quantity" gorm:"type:decimal(18,4);not null;default:0"`
	Price            decimal.Decimal `json:"price" gorm:"type:decimal(18,4);not null;default:0"`
	Amount           decimal.Decimal `json:"amount" gorm:"type:decimal(18,4);not null;default:0"`
	CostPrice        decimal.Decimal `json:"costPrice" gorm:"type:decimal(18,4);not null;default:0"`
	CostAmount       decimal.Decimal `json:"costAmount" gorm:"type:decimal(18,4);not null;default:0"`
	Remark           string          `json:"remark" gorm:"type:varchar(255)"`
}

func (RepairOrderItem) TableName() string { return "repair_order_items" }

type ProjectProject struct {
	shared.BaseModel
	ProjectNo      string          `json:"projectNo" gorm:"type:varchar(64);not null;index"`
	Name           string          `json:"name" gorm:"type:varchar(128);not null;index"`
	CustomerID     *uint64         `json:"customerId" gorm:"index"`
	CustomerName   string          `json:"customerName" gorm:"type:varchar(128);index"`
	ContractNo     string          `json:"contractNo" gorm:"type:varchar(64);index"`
	ContractName   string          `json:"contractName" gorm:"type:varchar(128);index"`
	ContractAmount decimal.Decimal `json:"contractAmount" gorm:"type:decimal(18,4);not null;default:0"`
	ReceivedAmount decimal.Decimal `json:"receivedAmount" gorm:"type:decimal(18,4);not null;default:0"`
	Status         string          `json:"status" gorm:"type:varchar(32);not null;default:planning;index"`
	Progress       int             `json:"progress" gorm:"not null;default:0;index"`
	BudgetAmount   decimal.Decimal `json:"budgetAmount" gorm:"type:decimal(18,4);not null;default:0"`
	CostAmount     decimal.Decimal `json:"costAmount" gorm:"type:decimal(18,4);not null;default:0"`
	SettleAmount   decimal.Decimal `json:"settleAmount" gorm:"type:decimal(18,4);not null;default:0"`
	StartDate      *time.Time      `json:"startDate" gorm:"index"`
	EndDate        *time.Time      `json:"endDate" gorm:"index"`
}

func (ProjectProject) TableName() string { return "project_projects" }

type FinanceAccount struct {
	shared.BaseModel
	Code           string          `json:"code" gorm:"type:varchar(64);not null;index"`
	Name           string          `json:"name" gorm:"type:varchar(128);not null;index"`
	AccountType    string          `json:"accountType" gorm:"type:varchar(32);not null;index"`
	OpeningBalance decimal.Decimal `json:"openingBalance" gorm:"type:decimal(18,4);not null;default:0"`
	Balance        decimal.Decimal `json:"balance" gorm:"type:decimal(18,4);not null;default:0;index"`
	Status         string          `json:"status" gorm:"type:varchar(32);not null;default:active;index"`
}

func (FinanceAccount) TableName() string { return "finance_accounts" }

type FinanceRecord struct {
	shared.BaseModel
	RecordNo     string          `json:"recordNo" gorm:"type:varchar(64);not null;index"`
	RecordType   string          `json:"recordType" gorm:"type:varchar(32);not null;index"`
	AccountID    *uint64         `json:"accountId" gorm:"index"`
	AccountName  string          `json:"accountName" gorm:"type:varchar(128);index"`
	TargetName   string          `json:"targetName" gorm:"type:varchar(128);index"`
	Amount       decimal.Decimal `json:"amount" gorm:"type:decimal(18,4);not null;default:0"`
	Status       string          `json:"status" gorm:"type:varchar(32);not null;default:confirmed;index"`
	BusinessType string          `json:"businessType" gorm:"type:varchar(32);index"`
	SourceType   string          `json:"sourceType" gorm:"type:varchar(32);index"`
	SourceID     *uint64         `json:"sourceId" gorm:"index"`
	BusinessNo   string          `json:"businessNo" gorm:"type:varchar(64);index"`
	OccurredAt   time.Time       `json:"occurredAt" gorm:"not null;index"`
}

func (FinanceRecord) TableName() string { return "finance_records" }

type Receivable struct {
	shared.BaseModel
	ReceivableNo   string          `json:"receivableNo" gorm:"type:varchar(64);not null;index"`
	CustomerID     *uint64         `json:"customerId" gorm:"index"`
	CustomerName   string          `json:"customerName" gorm:"type:varchar(128);not null;index"`
	SourceType     string          `json:"sourceType" gorm:"type:varchar(32);not null;index"`
	SourceID       *uint64         `json:"sourceId" gorm:"index"`
	SourceNo       string          `json:"sourceNo" gorm:"type:varchar(64);index"`
	TotalAmount    decimal.Decimal `json:"totalAmount" gorm:"type:decimal(18,4);not null;default:0"`
	ReceivedAmount decimal.Decimal `json:"receivedAmount" gorm:"type:decimal(18,4);not null;default:0"`
	BalanceAmount  decimal.Decimal `json:"balanceAmount" gorm:"type:decimal(18,4);not null;default:0;index"`
	InvoiceDate    time.Time       `json:"invoiceDate" gorm:"not null;index"`
	DueDate        time.Time       `json:"dueDate" gorm:"not null;index"`
	SettlementMode string          `json:"settlementMode" gorm:"type:varchar(32);not null;index"`
	Status         string          `json:"status" gorm:"type:varchar(32);not null;default:unpaid;index"`
}

func (Receivable) TableName() string { return "receivables" }

type CustomerStatement struct {
	shared.BaseModel
	StatementNo    string          `json:"statementNo" gorm:"type:varchar(64);not null;index"`
	CustomerID     uint64          `json:"customerId" gorm:"not null;index"`
	CustomerName   string          `json:"customerName" gorm:"type:varchar(128);not null;index"`
	ContactName    string          `json:"contactName" gorm:"type:varchar(64)"`
	ContactPhone   string          `json:"contactPhone" gorm:"type:varchar(32)"`
	StartDate      time.Time       `json:"startDate" gorm:"not null;index"`
	EndDate        time.Time       `json:"endDate" gorm:"not null;index"`
	TotalAmount    decimal.Decimal `json:"totalAmount" gorm:"type:decimal(18,4);not null;default:0"`
	ReceivedAmount decimal.Decimal `json:"receivedAmount" gorm:"type:decimal(18,4);not null;default:0"`
	UnpaidAmount   decimal.Decimal `json:"unpaidAmount" gorm:"type:decimal(18,4);not null;default:0"`
	CumulativeDebt decimal.Decimal `json:"cumulativeDebt" gorm:"type:decimal(18,4);not null;default:0"`
	Status         string          `json:"status" gorm:"type:varchar(32);not null;default:unconfirmed;index"`
	ConfirmedAt    *time.Time      `json:"confirmedAt" gorm:"index"`
	SettledAt      *time.Time      `json:"settledAt" gorm:"index"`
}

func (CustomerStatement) TableName() string { return "customer_statements" }

type CustomerStatementItem struct {
	shared.BaseModel
	StatementID      uint64          `json:"statementId" gorm:"not null;index"`
	SaleID           uint64          `json:"saleId" gorm:"not null;index"`
	SaleNo           string          `json:"saleNo" gorm:"type:varchar(64);not null;index"`
	SaleDate         time.Time       `json:"saleDate" gorm:"not null;index"`
	SourceType       string          `json:"sourceType" gorm:"type:varchar(32);index"`
	SourceID         *uint64         `json:"sourceId" gorm:"index"`
	SourceNo         string          `json:"sourceNo" gorm:"type:varchar(64);index"`
	ReceivableID     *uint64         `json:"receivableId" gorm:"index"`
	ProductName      string          `json:"productName" gorm:"type:varchar(128)"`
	Quantity         decimal.Decimal `json:"quantity" gorm:"type:decimal(18,4);not null;default:0"`
	TotalAmount      decimal.Decimal `json:"totalAmount" gorm:"type:decimal(18,4);not null;default:0"`
	ReceivedAmount   decimal.Decimal `json:"receivedAmount" gorm:"type:decimal(18,4);not null;default:0"`
	UnpaidAmount     decimal.Decimal `json:"unpaidAmount" gorm:"type:decimal(18,4);not null;default:0"`
	SettlementStatus string          `json:"settlementStatus" gorm:"type:varchar(32);index"`
}

func (CustomerStatementItem) TableName() string { return "customer_statement_items" }

type Payable struct {
	shared.BaseModel
	PayableNo     string          `json:"payableNo" gorm:"type:varchar(64);not null;index"`
	SupplierID    *uint64         `json:"supplierId" gorm:"index"`
	SupplierName  string          `json:"supplierName" gorm:"type:varchar(128);not null;index"`
	SourceType    string          `json:"sourceType" gorm:"type:varchar(32);not null;index"`
	SourceID      *uint64         `json:"sourceId" gorm:"index"`
	SourceNo      string          `json:"sourceNo" gorm:"type:varchar(64);index"`
	TotalAmount   decimal.Decimal `json:"totalAmount" gorm:"type:decimal(18,4);not null;default:0"`
	PaidAmount    decimal.Decimal `json:"paidAmount" gorm:"type:decimal(18,4);not null;default:0"`
	BalanceAmount decimal.Decimal `json:"balanceAmount" gorm:"type:decimal(18,4);not null;default:0;index"`
	BillDate      time.Time       `json:"billDate" gorm:"not null;index"`
	DueDate       time.Time       `json:"dueDate" gorm:"not null;index"`
	Status        string          `json:"status" gorm:"type:varchar(32);not null;default:unpaid;index"`
}

func (Payable) TableName() string { return "payables" }

type BusinessPhoto struct {
	shared.BaseModel
	Module      string    `json:"module" gorm:"type:varchar(32);not null;index"`
	BusinessID  uint64    `json:"businessId" gorm:"not null;index"`
	BusinessNo  string    `json:"businessNo" gorm:"type:varchar(64);index"`
	Scene       string    `json:"scene" gorm:"type:varchar(32);not null;default:general;index"`
	FileName    string    `json:"fileName" gorm:"type:varchar(255);not null"`
	FilePath    string    `json:"filePath" gorm:"type:varchar(500);not null"`
	FileURL     string    `json:"fileUrl" gorm:"type:varchar(500);not null"`
	ContentType string    `json:"contentType" gorm:"type:varchar(128)"`
	FileSize    int64     `json:"fileSize" gorm:"not null;default:0"`
	TakenAt     time.Time `json:"takenAt" gorm:"not null;index"`
}

func (BusinessPhoto) TableName() string { return "business_photos" }
