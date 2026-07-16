package repository

import (
	"context"

	authmodel "erp/internal/domain/auth/model"
	"erp/internal/shared/query"
)

type UserRepository interface {
	FindByID(ctx context.Context, tenantID uint64, id uint64) (*authmodel.User, error)
	FindByUsername(ctx context.Context, tenantID uint64, username string) (*authmodel.User, error)
	List(ctx context.Context, tenantID uint64, q query.PageQuery) ([]authmodel.User, int64, error)
	Create(ctx context.Context, user *authmodel.User) error
	Update(ctx context.Context, user *authmodel.User) error
	Delete(ctx context.Context, tenantID uint64, id uint64) error
	ReplaceRoles(ctx context.Context, userID uint64, roleIDs []uint64) error
	GetPermissionCodes(ctx context.Context, userID uint64) ([]string, error)
	GetDataScopes(ctx context.Context, userID uint64) ([]string, error)
	GetMenus(ctx context.Context, userID uint64) ([]authmodel.Menu, error)
	UpdateLoginInfo(ctx context.Context, tenantID uint64, id uint64, ip string) error
}

type RoleRepository interface {
	FindByID(ctx context.Context, tenantID uint64, id uint64) (*authmodel.Role, error)
	List(ctx context.Context, tenantID uint64, q query.PageQuery) ([]authmodel.Role, int64, error)
	Create(ctx context.Context, role *authmodel.Role) error
	Update(ctx context.Context, role *authmodel.Role) error
	Delete(ctx context.Context, tenantID uint64, id uint64) error
	ReplacePermissions(ctx context.Context, roleID uint64, permissionIDs []uint64) error
	ReplaceMenus(ctx context.Context, roleID uint64, menuIDs []uint64) error
}

type AuditRepository interface {
	CreateLoginLog(ctx context.Context, log *authmodel.LoginLog) error
	ListLoginLogs(ctx context.Context, tenantID uint64, q query.PageQuery) ([]authmodel.LoginLog, int64, error)
	ListOperationLogs(ctx context.Context, tenantID uint64, q query.PageQuery) ([]authmodel.OperationLog, int64, error)
}

type PermissionRepository interface {
	FindByID(ctx context.Context, tenantID uint64, id uint64) (*authmodel.Permission, error)
	FindByCode(ctx context.Context, tenantID uint64, code string) (*authmodel.Permission, error)
	List(ctx context.Context, tenantID uint64, q query.PageQuery) ([]authmodel.Permission, int64, error)
	Create(ctx context.Context, permission *authmodel.Permission) error
	Update(ctx context.Context, permission *authmodel.Permission) error
	Delete(ctx context.Context, tenantID uint64, id uint64) error
}

type MenuRepository interface {
	FindByID(ctx context.Context, tenantID uint64, id uint64) (*authmodel.Menu, error)
	List(ctx context.Context, tenantID uint64, q query.PageQuery) ([]authmodel.Menu, int64, error)
	Tree(ctx context.Context, tenantID uint64) ([]authmodel.Menu, error)
	Create(ctx context.Context, menu *authmodel.Menu) error
	Update(ctx context.Context, menu *authmodel.Menu) error
	Delete(ctx context.Context, tenantID uint64, id uint64) error
}
