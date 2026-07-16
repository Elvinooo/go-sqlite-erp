package dashboard

import (
	"context"

	dashboardmodel "erp/internal/domain/dashboard/model"
	dashboardrepo "erp/internal/domain/dashboard/repository"
)

type Service struct {
	repo dashboardrepo.BossDashboardRepository
}

func NewService(repo dashboardrepo.BossDashboardRepository) *Service {
	return &Service{repo: repo}
}

func (s *Service) Boss(ctx context.Context, tenantID uint64) (*dashboardmodel.BossDashboard, error) {
	return s.repo.GetBossDashboard(ctx, tenantID)
}

func (s *Service) PriceProfit(ctx context.Context, tenantID uint64, days int) (*dashboardmodel.PriceProfitAnalysis, error) {
	return s.repo.GetPriceProfitAnalysis(ctx, tenantID, days)
}

func (s *Service) SignIn(ctx context.Context, tenantID uint64, userID uint64, latitude float64, longitude float64, address string, device string) (*dashboardmodel.EngineerLocation, error) {
	return s.repo.SignIn(ctx, tenantID, userID, latitude, longitude, address, device)
}

func (s *Service) SignInHistory(ctx context.Context, tenantID uint64, userID uint64, page int, pageSize int) ([]dashboardmodel.EngineerCheckInHistory, int64, error) {
	return s.repo.ListEngineerCheckIns(ctx, tenantID, userID, page, pageSize)
}
