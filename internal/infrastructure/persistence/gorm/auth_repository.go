package gormrepo

import (
	"context"
	"strconv"
	"time"

	authmodel "erp/internal/domain/auth/model"
	"erp/internal/shared/query"
	"gorm.io/gorm"
)

type UserRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) FindByID(ctx context.Context, tenantID uint64, id uint64) (*authmodel.User, error) {
	var user authmodel.User
	err := r.db.WithContext(ctx).Preload("Roles").Where("tenant_id = ? AND id = ?", tenantID, id).First(&user).Error
	return &user, err
}

func (r *UserRepository) FindByUsername(ctx context.Context, tenantID uint64, username string) (*authmodel.User, error) {
	var user authmodel.User
	err := r.db.WithContext(ctx).Where("tenant_id = ? AND username = ?", tenantID, username).First(&user).Error
	return &user, err
}

func (r *UserRepository) List(ctx context.Context, tenantID uint64, q query.PageQuery) ([]authmodel.User, int64, error) {
	var users []authmodel.User
	var total int64
	db := r.db.WithContext(ctx).Model(&authmodel.User{}).Where("tenant_id = ?", tenantID)
	if q.Keyword != "" {
		like := "%" + q.Keyword + "%"
		db = db.Where("username LIKE ? OR real_name LIKE ? OR phone LIKE ?", like, like, like)
	}
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	err := db.Preload("Roles").Order(safeOrder(q)).Offset(q.Offset()).Limit(q.PageSize).Find(&users).Error
	return users, total, err
}

func (r *UserRepository) Create(ctx context.Context, user *authmodel.User) error {
	return r.db.WithContext(ctx).Create(user).Error
}

func (r *UserRepository) Update(ctx context.Context, user *authmodel.User) error {
	return r.db.WithContext(ctx).Omit("Roles").Save(user).Error
}

func (r *UserRepository) Delete(ctx context.Context, tenantID uint64, id uint64) error {
	return r.db.WithContext(ctx).Where("tenant_id = ? AND id = ?", tenantID, id).Delete(&authmodel.User{}).Error
}

func (r *UserRepository) ReplaceRoles(ctx context.Context, userID uint64, roleIDs []uint64) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("user_id = ?", userID).Delete(&authmodel.UserRole{}).Error; err != nil {
			return err
		}
		for _, roleID := range roleIDs {
			if err := tx.Create(&authmodel.UserRole{UserID: userID, RoleID: roleID, CreatedAt: time.Now()}).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func (r *UserRepository) GetPermissionCodes(ctx context.Context, userID uint64) ([]string, error) {
	var username string
	if err := r.db.WithContext(ctx).Table("users").Select("username").Where("id = ?", userID).Scan(&username).Error; err != nil {
		return nil, err
	}
	if username == "admin" {
		return []string{"*"}, nil
	}
	var codes []string
	err := r.db.WithContext(ctx).Raw(`
WITH RECURSIVE inherited_roles(id, parent_id) AS (
    SELECT roles.id, roles.parent_id
    FROM roles
    JOIN user_roles ON user_roles.role_id = roles.id
    WHERE user_roles.user_id = ? AND roles.status = 'active'
    UNION
    SELECT parent.id, parent.parent_id
    FROM roles parent
    JOIN inherited_roles child ON child.parent_id = parent.id
    WHERE parent.status = 'active'
)
SELECT DISTINCT permissions.code
FROM permissions
JOIN role_permissions ON role_permissions.permission_id = permissions.id
JOIN inherited_roles ON inherited_roles.id = role_permissions.role_id
WHERE permissions.status = 'active'
`, userID).Scan(&codes).Error
	return codes, err
}

func (r *UserRepository) GetDataScopes(ctx context.Context, userID uint64) ([]string, error) {
	var username string
	if err := r.db.WithContext(ctx).Table("users").Select("username").Where("id = ?", userID).Scan(&username).Error; err != nil {
		return nil, err
	}
	if username == "admin" {
		return []string{"all"}, nil
	}
	var scopes []string
	err := r.db.WithContext(ctx).Raw(`
WITH RECURSIVE inherited_roles(id, parent_id, data_scope) AS (
    SELECT roles.id, roles.parent_id, roles.data_scope
    FROM roles
    JOIN user_roles ON user_roles.role_id = roles.id
    WHERE user_roles.user_id = ? AND roles.status = 'active'
    UNION
    SELECT parent.id, parent.parent_id, parent.data_scope
    FROM roles parent
    JOIN inherited_roles child ON child.parent_id = parent.id
    WHERE parent.status = 'active'
)
SELECT DISTINCT data_scope FROM inherited_roles
`, userID).Scan(&scopes).Error
	return scopes, err
}

func (r *UserRepository) GetMenus(ctx context.Context, userID uint64) ([]authmodel.Menu, error) {
	var menus []authmodel.Menu
	var username string
	if err := r.db.WithContext(ctx).Table("users").Select("username").Where("id = ?", userID).Scan(&username).Error; err != nil {
		return nil, err
	}
	if username == "admin" {
		err := r.db.WithContext(ctx).Where("status = ? AND visible = ?", "active", true).Order("sort ASC, id ASC").Find(&menus).Error
		return menus, err
	}
	err := r.db.WithContext(ctx).Raw(`
WITH RECURSIVE inherited_roles(id, parent_id) AS (
    SELECT roles.id, roles.parent_id
    FROM roles
    JOIN user_roles ON user_roles.role_id = roles.id
    WHERE user_roles.user_id = ? AND roles.status = 'active'
    UNION
    SELECT parent.id, parent.parent_id
    FROM roles parent
    JOIN inherited_roles child ON child.parent_id = parent.id
    WHERE parent.status = 'active'
)
SELECT DISTINCT menus.*
FROM menus
JOIN role_menus ON role_menus.menu_id = menus.id
JOIN inherited_roles ON inherited_roles.id = role_menus.role_id
WHERE menus.status = 'active' AND menus.visible = true
ORDER BY menus.sort ASC, menus.id ASC
`, userID).Scan(&menus).Error
	return menus, err
}

func (r *UserRepository) UpdateLoginInfo(ctx context.Context, tenantID uint64, id uint64, ip string) error {
	return r.db.WithContext(ctx).Model(&authmodel.User{}).
		Where("tenant_id = ? AND id = ?", tenantID, id).
		Updates(map[string]any{"last_login_at": time.Now(), "last_login_ip": ip}).Error
}

type RoleRepository struct {
	db *gorm.DB
}

func NewRoleRepository(db *gorm.DB) *RoleRepository {
	return &RoleRepository{db: db}
}

func (r *RoleRepository) FindByID(ctx context.Context, tenantID uint64, id uint64) (*authmodel.Role, error) {
	var role authmodel.Role
	err := r.db.WithContext(ctx).Preload("Parent").Preload("Permissions").Preload("Menus").Where("tenant_id = ? AND id = ?", tenantID, id).First(&role).Error
	return &role, err
}

func (r *RoleRepository) List(ctx context.Context, tenantID uint64, q query.PageQuery) ([]authmodel.Role, int64, error) {
	var roles []authmodel.Role
	var total int64
	db := r.db.WithContext(ctx).Model(&authmodel.Role{}).Where("tenant_id = ?", tenantID)
	if q.Keyword != "" {
		like := "%" + q.Keyword + "%"
		db = db.Where("code LIKE ? OR name LIKE ?", like, like)
	}
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	err := db.Preload("Parent").Preload("Permissions").Preload("Menus").Order(safeOrder(q)).Offset(q.Offset()).Limit(q.PageSize).Find(&roles).Error
	return roles, total, err
}

func (r *RoleRepository) Create(ctx context.Context, role *authmodel.Role) error {
	return r.db.WithContext(ctx).Create(role).Error
}

func (r *RoleRepository) Update(ctx context.Context, role *authmodel.Role) error {
	return r.db.WithContext(ctx).Omit("Parent", "Menus", "Permissions").Save(role).Error
}

func (r *RoleRepository) Delete(ctx context.Context, tenantID uint64, id uint64) error {
	return r.db.WithContext(ctx).Where("tenant_id = ? AND id = ?", tenantID, id).Delete(&authmodel.Role{}).Error
}

func (r *RoleRepository) ReplacePermissions(ctx context.Context, roleID uint64, permissionIDs []uint64) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("role_id = ?", roleID).Delete(&authmodel.RolePermission{}).Error; err != nil {
			return err
		}
		for _, permissionID := range permissionIDs {
			if err := tx.Create(&authmodel.RolePermission{RoleID: roleID, PermissionID: permissionID, CreatedAt: time.Now()}).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func (r *RoleRepository) ReplaceMenus(ctx context.Context, roleID uint64, menuIDs []uint64) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("role_id = ?", roleID).Delete(&authmodel.RoleMenu{}).Error; err != nil {
			return err
		}
		for _, menuID := range menuIDs {
			if err := tx.Create(&authmodel.RoleMenu{RoleID: roleID, MenuID: menuID, CreatedAt: time.Now()}).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

type PermissionRepository struct {
	db *gorm.DB
}

func NewPermissionRepository(db *gorm.DB) *PermissionRepository {
	return &PermissionRepository{db: db}
}

func (r *PermissionRepository) FindByID(ctx context.Context, tenantID uint64, id uint64) (*authmodel.Permission, error) {
	var permission authmodel.Permission
	err := r.db.WithContext(ctx).Where("tenant_id = ? AND id = ?", tenantID, id).First(&permission).Error
	return &permission, err
}

func (r *PermissionRepository) FindByCode(ctx context.Context, tenantID uint64, code string) (*authmodel.Permission, error) {
	var permission authmodel.Permission
	err := r.db.WithContext(ctx).Where("tenant_id = ? AND code = ?", tenantID, code).First(&permission).Error
	return &permission, err
}

func (r *PermissionRepository) List(ctx context.Context, tenantID uint64, q query.PageQuery) ([]authmodel.Permission, int64, error) {
	var permissions []authmodel.Permission
	var total int64
	db := r.db.WithContext(ctx).Model(&authmodel.Permission{}).Where("tenant_id = ?", tenantID)
	if q.Keyword != "" {
		like := "%" + q.Keyword + "%"
		db = db.Where("code LIKE ? OR name LIKE ? OR module LIKE ?", like, like, like)
	}
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	err := db.Order(safeOrder(q)).Offset(q.Offset()).Limit(q.PageSize).Find(&permissions).Error
	return permissions, total, err
}

func (r *PermissionRepository) Create(ctx context.Context, permission *authmodel.Permission) error {
	return r.db.WithContext(ctx).Create(permission).Error
}

func (r *PermissionRepository) Update(ctx context.Context, permission *authmodel.Permission) error {
	return r.db.WithContext(ctx).Save(permission).Error
}

func (r *PermissionRepository) Delete(ctx context.Context, tenantID uint64, id uint64) error {
	return r.db.WithContext(ctx).Where("tenant_id = ? AND id = ?", tenantID, id).Delete(&authmodel.Permission{}).Error
}

type MenuRepository struct {
	db *gorm.DB
}

func NewMenuRepository(db *gorm.DB) *MenuRepository {
	return &MenuRepository{db: db}
}

func (r *MenuRepository) FindByID(ctx context.Context, tenantID uint64, id uint64) (*authmodel.Menu, error) {
	var menu authmodel.Menu
	err := r.db.WithContext(ctx).Where("tenant_id = ? AND id = ?", tenantID, id).First(&menu).Error
	return &menu, err
}

func (r *MenuRepository) List(ctx context.Context, tenantID uint64, q query.PageQuery) ([]authmodel.Menu, int64, error) {
	var menus []authmodel.Menu
	var total int64
	db := r.db.WithContext(ctx).Model(&authmodel.Menu{}).Where("tenant_id = ?", tenantID)
	if q.Keyword != "" {
		like := "%" + q.Keyword + "%"
		db = db.Where("name LIKE ? OR title LIKE ? OR path LIKE ?", like, like, like)
	}
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	err := db.Order("sort ASC, id ASC").Offset(q.Offset()).Limit(q.PageSize).Find(&menus).Error
	return menus, total, err
}

func (r *MenuRepository) Tree(ctx context.Context, tenantID uint64) ([]authmodel.Menu, error) {
	var menus []authmodel.Menu
	err := r.db.WithContext(ctx).Where("tenant_id = ?", tenantID).Order("sort ASC, id ASC").Find(&menus).Error
	return menus, err
}

func (r *MenuRepository) Create(ctx context.Context, menu *authmodel.Menu) error {
	return r.db.WithContext(ctx).Create(menu).Error
}

func (r *MenuRepository) Update(ctx context.Context, menu *authmodel.Menu) error {
	return r.db.WithContext(ctx).Save(menu).Error
}

func (r *MenuRepository) Delete(ctx context.Context, tenantID uint64, id uint64) error {
	return r.db.WithContext(ctx).Where("tenant_id = ? AND id = ?", tenantID, id).Delete(&authmodel.Menu{}).Error
}

func safeOrder(q query.PageQuery) string {
	allowed := map[string]bool{"id": true, "created_at": true, "updated_at": true, "sort": true, "username": true, "name": true, "code": true}
	if !allowed[q.SortBy] {
		q.SortBy = "id"
	}
	if q.Order != "asc" {
		q.Order = "desc"
	}
	return q.SortBy + " " + q.Order
}

type AuditRepository struct {
	db *gorm.DB
}

func NewAuditRepository(db *gorm.DB) *AuditRepository {
	return &AuditRepository{db: db}
}

func (r *AuditRepository) CreateLoginLog(ctx context.Context, log *authmodel.LoginLog) error {
	return r.db.WithContext(ctx).Create(log).Error
}

func (r *AuditRepository) ListLoginLogs(ctx context.Context, tenantID uint64, q query.PageQuery) ([]authmodel.LoginLog, int64, error) {
	var logs []authmodel.LoginLog
	var total int64
	db := r.db.WithContext(ctx).Model(&authmodel.LoginLog{}).Where("tenant_id = ?", tenantID)
	if q.Keyword != "" {
		like := "%" + q.Keyword + "%"
		db = db.Where("username LIKE ? OR ip LIKE ? OR message LIKE ?", like, like, like)
	}
	if q.UserID > 0 {
		db = db.Where("user_id = ?", q.UserID)
	}
	if q.Username != "" {
		db = db.Where("username = ?", q.Username)
	}
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	err := db.Order("created_at DESC").Offset(q.Offset()).Limit(q.PageSize).Find(&logs).Error
	return logs, total, err
}

func (r *AuditRepository) ListOperationLogs(ctx context.Context, tenantID uint64, q query.PageQuery) ([]authmodel.OperationLog, int64, error) {
	var logs []authmodel.OperationLog
	var total int64
	db := r.db.WithContext(ctx).Model(&authmodel.OperationLog{}).Where("tenant_id = ?", tenantID)
	if q.Keyword != "" {
		like := "%" + q.Keyword + "%"
		db = db.Where("username LIKE ? OR path LIKE ? OR module LIKE ? OR action LIKE ? OR request_body LIKE ?", like, like, like, like, like)
	}
	if q.Module != "" {
		db = db.Where("module = ?", q.Module)
	}
	if q.Action != "" {
		db = db.Where("action = ?", q.Action)
	}
	if q.Method != "" {
		db = db.Where("method = ?", q.Method)
	}
	if q.StatusCode != "" {
		if code, err := strconv.Atoi(q.StatusCode); err == nil {
			db = db.Where("status_code = ?", code)
		}
	}
	if q.StartDate != "" {
		if start, err := time.Parse(time.RFC3339, q.StartDate); err == nil {
			db = db.Where("created_at >= ?", start)
		}
	}
	if q.EndDate != "" {
		if end, err := time.Parse(time.RFC3339, q.EndDate); err == nil {
			db = db.Where("created_at <= ?", end)
		}
	}
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	err := db.Order("created_at DESC").Offset(q.Offset()).Limit(q.PageSize).Find(&logs).Error
	return logs, total, err
}
