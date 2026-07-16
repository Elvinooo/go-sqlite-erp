package customer

import "github.com/shopspring/decimal"

type CustomerRequest struct {
	Code              string          `json:"code" binding:"required,max=64"`
	Name              string          `json:"name" binding:"required,max=128"`
	Type              string          `json:"type" binding:"max=32"`
	Level             string          `json:"level" binding:"max=32"`
	Phone             string          `json:"phone" binding:"max=32"`
	Email             string          `json:"email" binding:"max=128"`
	TaxNo             string          `json:"taxNo" binding:"max=64"`
	Address           string          `json:"address" binding:"max=255"`
	PaymentMethod     string          `json:"paymentMethod" binding:"max=32"`
	CreditLimit       decimal.Decimal `json:"creditLimit"`
	CreditDays        int             `json:"creditDays"`
	BillingCycle      string          `json:"billingCycle" binding:"max=32"`
	PaymentDay        int             `json:"paymentDay"`
	ReceivableBalance decimal.Decimal `json:"receivableBalance"`
	Status            string          `json:"status"`
	Remark            string          `json:"remark"`
}

type CustomerDebtResponse struct {
	CustomerID        uint64          `json:"customerId"`
	CustomerName      string          `json:"customerName"`
	ReceivableBalance decimal.Decimal `json:"receivableBalance"`
}

type ImportResult struct {
	Total   int      `json:"total"`
	Success int      `json:"success"`
	Failed  int      `json:"failed"`
	Errors  []string `json:"errors"`
}
