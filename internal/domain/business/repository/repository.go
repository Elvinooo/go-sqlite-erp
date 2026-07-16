package repository

import (
	"context"

	"erp/internal/shared/query"
)

type ModuleRepository interface {
	List(ctx context.Context, module string, tenantID uint64, q query.PageQuery) ([]map[string]any, int64, error)
	Find(ctx context.Context, module string, tenantID uint64, id uint64) (map[string]any, error)
	Create(ctx context.Context, module string, tenantID uint64, operatorID uint64, data map[string]any) (map[string]any, error)
	Update(ctx context.Context, module string, tenantID uint64, id uint64, operatorID uint64, data map[string]any) (map[string]any, error)
	Delete(ctx context.Context, module string, tenantID uint64, id uint64, operatorID uint64, reason string) error
	Action(ctx context.Context, module string, tenantID uint64, operatorID uint64, action string, data map[string]any) (map[string]any, error)
	CreatePhoto(ctx context.Context, module string, tenantID uint64, operatorID uint64, businessID uint64, photo map[string]any) (map[string]any, error)
	ListPhotos(ctx context.Context, module string, tenantID uint64, businessID uint64) ([]map[string]any, error)
}
