package gormrepo

import (
	"context"

	customermodel "erp/internal/domain/customer/model"
	customerrepo "erp/internal/domain/customer/repository"
	"erp/internal/shared/query"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

type CustomerRepository struct {
	db *gorm.DB
}

func NewCustomerRepository(db *gorm.DB) *CustomerRepository {
	return &CustomerRepository{db: db}
}

func (r *CustomerRepository) FindByID(ctx context.Context, tenantID uint64, id uint64) (*customermodel.Customer, error) {
	var customer customermodel.Customer
	err := r.db.WithContext(ctx).Where("tenant_id = ? AND id = ?", tenantID, id).First(&customer).Error
	return &customer, err
}

func (r *CustomerRepository) FindByCode(ctx context.Context, tenantID uint64, code string) (*customermodel.Customer, error) {
	var customer customermodel.Customer
	err := r.db.WithContext(ctx).Where("tenant_id = ? AND code = ?", tenantID, code).First(&customer).Error
	return &customer, err
}

func (r *CustomerRepository) List(ctx context.Context, tenantID uint64, q query.PageQuery) ([]customermodel.Customer, int64, error) {
	var customers []customermodel.Customer
	var total int64
	db := r.db.WithContext(ctx).Model(&customermodel.Customer{}).Where("tenant_id = ?", tenantID)
	if q.Keyword != "" {
		like := "%" + q.Keyword + "%"
		db = db.Where("code LIKE ? OR name LIKE ? OR phone LIKE ? OR tax_no LIKE ? OR address LIKE ?", like, like, like, like, like)
	}
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	err := db.Order(safeOrder(q)).Offset(q.Offset()).Limit(q.PageSize).Find(&customers).Error
	return customers, total, err
}

func (r *CustomerRepository) Create(ctx context.Context, customer *customermodel.Customer) error {
	return r.db.WithContext(ctx).Create(customer).Error
}

func (r *CustomerRepository) Update(ctx context.Context, customer *customermodel.Customer) error {
	return r.db.WithContext(ctx).Save(customer).Error
}

func (r *CustomerRepository) Delete(ctx context.Context, tenantID uint64, id uint64) error {
	return r.db.WithContext(ctx).Where("tenant_id = ? AND id = ?", tenantID, id).Delete(&customermodel.Customer{}).Error
}

func (r *CustomerRepository) BatchCreate(ctx context.Context, customers []customermodel.Customer) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for i := range customers {
			var existing customermodel.Customer
			err := tx.Where("tenant_id = ? AND code = ?", customers[i].TenantID, customers[i].Code).First(&existing).Error
			if err == nil {
				customers[i].ID = existing.ID
				customers[i].CreatedAt = existing.CreatedAt
				if err := tx.Model(&existing).Updates(map[string]any{
					"name": customers[i].Name, "type": customers[i].Type, "level": customers[i].Level,
					"phone": customers[i].Phone, "email": customers[i].Email, "tax_no": customers[i].TaxNo,
					"address": customers[i].Address, "credit_limit": customers[i].CreditLimit,
					"payment_method": customers[i].PaymentMethod, "credit_days": customers[i].CreditDays,
					"billing_cycle": customers[i].BillingCycle, "payment_day": customers[i].PaymentDay,
					"receivable_balance": customers[i].ReceivableBalance, "status": customers[i].Status,
					"updated_by": customers[i].UpdatedBy,
				}).Error; err != nil {
					return err
				}
				continue
			}
			if err != gorm.ErrRecordNotFound {
				return err
			}
			if err := tx.Create(&customers[i]).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func (r *CustomerRepository) Debt(ctx context.Context, tenantID uint64, id uint64) (decimal.Decimal, error) {
	if r.db.Migrator().HasTable("receivables") {
		var total decimal.Decimal
		err := r.db.WithContext(ctx).
			Table("receivables").
			Select("COALESCE(SUM(balance_amount), 0)").
			Where("tenant_id = ? AND customer_id = ? AND status <> ?", tenantID, id, "paid").
			Scan(&total).Error
		return total, err
	}
	customer, err := r.FindByID(ctx, tenantID, id)
	if err != nil {
		return decimal.Zero, err
	}
	return customer.ReceivableBalance, nil
}

func (r *CustomerRepository) OrderHistory(ctx context.Context, tenantID uint64, id uint64, q query.PageQuery) ([]customerrepo.CustomerOrderHistory, int64, error) {
	if !r.db.Migrator().HasTable("sales_orders") {
		return []customerrepo.CustomerOrderHistory{}, 0, nil
	}
	var total int64
	db := r.db.WithContext(ctx).Table("sales_orders").Where("tenant_id = ? AND customer_id = ?", tenantID, id)
	if q.Keyword != "" {
		db = db.Where("order_no LIKE ?", "%"+q.Keyword+"%")
	}
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var list []customerrepo.CustomerOrderHistory
	err := db.Select("id, order_no, order_date, status, total_amount, received_amount, receivable_amount, profit_amount").
		Order("order_date DESC, id DESC").
		Offset(q.Offset()).
		Limit(q.PageSize).
		Scan(&list).Error
	return list, total, err
}
