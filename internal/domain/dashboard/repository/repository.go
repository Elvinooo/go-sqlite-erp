package repository

import (
	"context"

	dashboardmodel "erp/internal/domain/dashboard/model"
)

type BossDashboardRepository interface {
	GetBossDashboard(ctx context.Context, tenantID uint64) (*dashboardmodel.BossDashboard, error)
	GetPriceProfitAnalysis(ctx context.Context, tenantID uint64, days int) (*dashboardmodel.PriceProfitAnalysis, error)
	SignIn(ctx context.Context, tenantID uint64, userID uint64, latitude float64, longitude float64, address string, device string) (*dashboardmodel.EngineerLocation, error)
	ListEngineerCheckIns(ctx context.Context, tenantID uint64, userID uint64, page int, pageSize int) ([]dashboardmodel.EngineerCheckInHistory, int64, error)
}
