package customer

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"strconv"
	"strings"

	customermodel "erp/internal/domain/customer/model"
	customerrepo "erp/internal/domain/customer/repository"
	"erp/internal/shared/query"
	"github.com/shopspring/decimal"
	"github.com/xuri/excelize/v2"
)

type Service struct {
	customers customerrepo.CustomerRepository
}

func NewService(customers customerrepo.CustomerRepository) *Service {
	return &Service{customers: customers}
}

func (s *Service) List(ctx context.Context, tenantID uint64, q query.PageQuery) ([]customermodel.Customer, int64, error) {
	return s.customers.List(ctx, tenantID, q)
}

func (s *Service) Get(ctx context.Context, tenantID uint64, id uint64) (*customermodel.Customer, error) {
	return s.customers.FindByID(ctx, tenantID, id)
}

func (s *Service) Create(ctx context.Context, tenantID uint64, req CustomerRequest, operatorID uint64) (*customermodel.Customer, error) {
	customer := fromRequest(req)
	customer.TenantID = tenantID
	customer.CreatedBy = &operatorID
	customer.UpdatedBy = &operatorID
	return customer, s.customers.Create(ctx, customer)
}

func (s *Service) Update(ctx context.Context, tenantID uint64, id uint64, req CustomerRequest, operatorID uint64) (*customermodel.Customer, error) {
	customer, err := s.customers.FindByID(ctx, tenantID, id)
	if err != nil {
		return nil, err
	}
	customer.Code = req.Code
	customer.Name = req.Name
	customer.Type = defaultType(req.Type)
	customer.Level = req.Level
	customer.Phone = req.Phone
	customer.Email = req.Email
	customer.TaxNo = req.TaxNo
	customer.Address = req.Address
	customer.PaymentMethod = defaultPaymentMethod(req.PaymentMethod)
	customer.CreditLimit = req.CreditLimit
	customer.CreditDays = req.CreditDays
	customer.BillingCycle = defaultBillingCycle(req.BillingCycle)
	customer.PaymentDay = req.PaymentDay
	customer.ReceivableBalance = req.ReceivableBalance
	customer.Status = defaultStatus(req.Status)
	customer.Remark = req.Remark
	customer.UpdatedBy = &operatorID
	return customer, s.customers.Update(ctx, customer)
}

func (s *Service) Delete(ctx context.Context, tenantID uint64, id uint64) error {
	return s.customers.Delete(ctx, tenantID, id)
}

func (s *Service) Debt(ctx context.Context, tenantID uint64, id uint64) (*CustomerDebtResponse, error) {
	customer, err := s.customers.FindByID(ctx, tenantID, id)
	if err != nil {
		return nil, err
	}
	debt, err := s.customers.Debt(ctx, tenantID, id)
	if err != nil {
		return nil, err
	}
	return &CustomerDebtResponse{CustomerID: customer.ID, CustomerName: customer.Name, ReceivableBalance: debt}, nil
}

func (s *Service) OrderHistory(ctx context.Context, tenantID uint64, id uint64, q query.PageQuery) ([]customerrepo.CustomerOrderHistory, int64, error) {
	return s.customers.OrderHistory(ctx, tenantID, id, q)
}

func (s *Service) ImportExcel(ctx context.Context, tenantID uint64, operatorID uint64, reader io.Reader) (*ImportResult, error) {
	body, err := io.ReadAll(reader)
	if err != nil {
		return nil, err
	}
	file, err := excelize.OpenReader(bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	defer file.Close()

	sheet := file.GetSheetName(0)
	rows, err := file.GetRows(sheet)
	if err != nil {
		return nil, err
	}
	result := &ImportResult{Errors: []string{}}
	customers := make([]customermodel.Customer, 0)
	for index, row := range rows {
		if index == 0 {
			continue
		}
		result.Total++
		if len(row) < 2 {
			result.Failed++
			result.Errors = append(result.Errors, fmt.Sprintf("第%d行：客户编码和客户名称必填", index+1))
			continue
		}
		code := strings.TrimSpace(cell(row, 0))
		name := strings.TrimSpace(cell(row, 1))
		if code == "" || name == "" {
			result.Failed++
			result.Errors = append(result.Errors, fmt.Sprintf("第%d行：客户编码和客户名称必填", index+1))
			continue
		}
		creditLimit, _ := decimal.NewFromString(defaultZero(cell(row, 9)))
		creditDays, _ := strconv.Atoi(defaultZero(cell(row, 10)))
		paymentDay, _ := strconv.Atoi(defaultZero(cell(row, 12)))
		receivableBalance, _ := decimal.NewFromString(defaultZero(cell(row, 13)))
		item := customermodel.Customer{
			Code: code, Name: name, Type: defaultType(cell(row, 2)), Level: cell(row, 3),
			Phone: cell(row, 4), Email: cell(row, 5), TaxNo: cell(row, 6), Address: cell(row, 7),
			PaymentMethod: defaultPaymentMethod(cell(row, 8)), CreditLimit: creditLimit, CreditDays: creditDays,
			BillingCycle: defaultBillingCycle(cell(row, 11)), PaymentDay: paymentDay,
			ReceivableBalance: receivableBalance, Status: defaultStatus(cell(row, 14)),
		}
		item.TenantID = tenantID
		item.CreatedBy = &operatorID
		item.UpdatedBy = &operatorID
		customers = append(customers, item)
	}
	if len(customers) > 0 {
		if err := s.customers.BatchCreate(ctx, customers); err != nil {
			return nil, err
		}
		result.Success = len(customers)
	}
	return result, nil
}

func (s *Service) ExportExcel(ctx context.Context, tenantID uint64, q query.PageQuery) ([]byte, error) {
	q.Page = 1
	q.PageSize = 100000
	customers, _, err := s.customers.List(ctx, tenantID, q)
	if err != nil {
		return nil, err
	}
	file := excelize.NewFile()
	defer file.Close()
	sheet := "客户"
	file.SetSheetName("Sheet1", sheet)
	headers := []string{"客户编码", "客户名称", "类型", "等级", "电话", "邮箱", "税号", "地址", "结算方式", "信用额度", "信用天数", "账期", "付款日", "应收余额", "状态", "备注"}
	for i, header := range headers {
		cellName, _ := excelize.CoordinatesToCellName(i+1, 1)
		_ = file.SetCellValue(sheet, cellName, header)
	}
	for r, item := range customers {
		values := []any{item.Code, item.Name, item.Type, item.Level, item.Phone, item.Email, item.TaxNo, item.Address, item.PaymentMethod, item.CreditLimit.StringFixed(2), item.CreditDays, item.BillingCycle, item.PaymentDay, item.ReceivableBalance.StringFixed(2), item.Status, item.Remark}
		for c, value := range values {
			cellName, _ := excelize.CoordinatesToCellName(c+1, r+2)
			_ = file.SetCellValue(sheet, cellName, value)
		}
	}
	buffer, err := file.WriteToBuffer()
	if err != nil {
		return nil, err
	}
	return buffer.Bytes(), nil
}

func fromRequest(req CustomerRequest) *customermodel.Customer {
	item := &customermodel.Customer{
		Code: req.Code, Name: req.Name, Type: defaultType(req.Type), Level: req.Level,
		Phone: req.Phone, Email: req.Email, TaxNo: req.TaxNo, Address: req.Address,
		PaymentMethod: defaultPaymentMethod(req.PaymentMethod), CreditLimit: req.CreditLimit,
		CreditDays: req.CreditDays, BillingCycle: defaultBillingCycle(req.BillingCycle), PaymentDay: req.PaymentDay,
		ReceivableBalance: req.ReceivableBalance,
		Status:            defaultStatus(req.Status),
	}
	item.Remark = req.Remark
	return item
}

func defaultPaymentMethod(value string) string {
	if strings.TrimSpace(value) == "" {
		return "immediate"
	}
	return value
}

func defaultBillingCycle(value string) string {
	if strings.TrimSpace(value) == "" {
		return "none"
	}
	return value
}

func defaultType(value string) string {
	if strings.TrimSpace(value) == "" {
		return "company"
	}
	return value
}

func defaultStatus(value string) string {
	if strings.TrimSpace(value) == "" {
		return "active"
	}
	return value
}

func cell(row []string, index int) string {
	if index >= len(row) {
		return ""
	}
	return strings.TrimSpace(row[index])
}

func defaultZero(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return "0"
	}
	return value
}
