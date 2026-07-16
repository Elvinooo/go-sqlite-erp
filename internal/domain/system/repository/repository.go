package repository

import (
	"context"

	systemmodel "erp/internal/domain/system/model"
	"erp/internal/shared/query"
)

type SettingRepository interface {
	FindByID(ctx context.Context, tenantID uint64, id uint64) (*systemmodel.SystemSetting, error)
	FindByKey(ctx context.Context, tenantID uint64, key string) (*systemmodel.SystemSetting, error)
	List(ctx context.Context, tenantID uint64, q query.PageQuery) ([]systemmodel.SystemSetting, int64, error)
	Create(ctx context.Context, setting *systemmodel.SystemSetting) error
	Update(ctx context.Context, setting *systemmodel.SystemSetting) error
	Delete(ctx context.Context, tenantID uint64, id uint64) error
	RestoreTestData(ctx context.Context, tenantID uint64, operatorID uint64) (map[string]any, error)
}
