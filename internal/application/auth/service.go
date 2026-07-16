package auth

import (
	"context"
	stderrors "errors"
	"net/http"
	"time"
	"unicode"

	authmodel "erp/internal/domain/auth/model"
	authrepo "erp/internal/domain/auth/repository"
	"erp/internal/infrastructure/security"
	apperrors "erp/internal/shared/errors"
	"erp/internal/shared/query"
	"erp/internal/shared/utils"
	"gorm.io/gorm"
)

type Service struct {
	users       authrepo.UserRepository
	roles       authrepo.RoleRepository
	permissions authrepo.PermissionRepository
	menus       authrepo.MenuRepository
	audit       authrepo.AuditRepository
	jwt         *security.JWTManager
}

func NewService(users authrepo.UserRepository, roles authrepo.RoleRepository, permissions authrepo.PermissionRepository, menus authrepo.MenuRepository, audit authrepo.AuditRepository, jwt *security.JWTManager) *Service {
	return &Service{users: users, roles: roles, permissions: permissions, menus: menus, audit: audit, jwt: jwt}
}

func (s *Service) Login(ctx context.Context, tenantID uint64, req LoginRequest, ip string, userAgent string) (*TokenResponse, error) {
	loginLog := &authmodel.LoginLog{TenantID: tenantID, Username: req.Username, IP: ip, UserAgent: userAgent, CreatedAt: time.Now()}
	user, err := s.users.FindByUsername(ctx, tenantID, req.Username)
	if err != nil {
		if stderrors.Is(err, gorm.ErrRecordNotFound) {
			loginLog.Message = "用户名或密码错误"
			_ = s.audit.CreateLoginLog(ctx, loginLog)
			return nil, apperrors.ErrInvalidCredential
		}
		return nil, err
	}
	loginLog.UserID = &user.ID
	if !user.IsActive() {
		loginLog.Message = "用户已禁用"
		_ = s.audit.CreateLoginLog(ctx, loginLog)
		return nil, apperrors.ErrDisabledUser
	}
	if !utils.CheckPassword(user.PasswordHash, req.Password) {
		loginLog.Message = "用户名或密码错误"
		_ = s.audit.CreateLoginLog(ctx, loginLog)
		return nil, apperrors.ErrInvalidCredential
	}
	access, refresh, expiresIn, err := s.jwt.Generate(user.ID, user.TenantID, user.Username, user.PasswordVersion)
	if err != nil {
		return nil, err
	}
	_ = s.users.UpdateLoginInfo(ctx, tenantID, user.ID, ip)
	loginLog.Success = true
	loginLog.Message = "登录成功"
	_ = s.audit.CreateLoginLog(ctx, loginLog)
	return &TokenResponse{AccessToken: access, RefreshToken: refresh, TokenType: "Bearer", ExpiresIn: expiresIn, MustChangePassword: user.MustChangePassword}, nil
}

func (s *Service) Refresh(ctx context.Context, tokenValue string) (*TokenResponse, error) {
	claims, err := s.jwt.Parse(tokenValue)
	if err != nil || claims.TokenUse != "refresh" {
		return nil, apperrors.ErrInvalidToken
	}
	user, err := s.users.FindByID(ctx, claims.TenantID, claims.UserID)
	if err != nil {
		return nil, err
	}
	if !user.IsActive() {
		return nil, apperrors.ErrDisabledUser
	}
	if claims.PasswordVersion != user.PasswordVersion {
		return nil, apperrors.ErrInvalidToken
	}
	access, refresh, expiresIn, err := s.jwt.Generate(user.ID, user.TenantID, user.Username, user.PasswordVersion)
	if err != nil {
		return nil, err
	}
	return &TokenResponse{AccessToken: access, RefreshToken: refresh, TokenType: "Bearer", ExpiresIn: expiresIn, MustChangePassword: user.MustChangePassword}, nil
}

func (s *Service) ValidateAccessToken(ctx context.Context, claims *security.Claims) error {
	user, err := s.users.FindByID(ctx, claims.TenantID, claims.UserID)
	if err != nil {
		return err
	}
	if !user.IsActive() {
		return apperrors.ErrDisabledUser
	}
	if claims.PasswordVersion != user.PasswordVersion {
		return apperrors.ErrInvalidToken
	}
	return nil
}

func (s *Service) CurrentUser(ctx context.Context, tenantID uint64, userID uint64) (*CurrentUserResponse, error) {
	user, err := s.users.FindByID(ctx, tenantID, userID)
	if err != nil {
		return nil, err
	}
	codes, err := s.users.GetPermissionCodes(ctx, userID)
	if err != nil {
		return nil, err
	}
	menus, err := s.users.GetMenus(ctx, userID)
	if err != nil {
		return nil, err
	}
	dataScopes, err := s.users.GetDataScopes(ctx, userID)
	if err != nil {
		return nil, err
	}
	return &CurrentUserResponse{
		ID: user.ID, Username: user.Username, RealName: user.RealName, Avatar: user.Avatar,
		MustChangePassword: user.MustChangePassword,
		Permissions:        codes, DataScopes: dataScopes, Menus: BuildMenuTree(menus),
	}, nil
}

func (s *Service) HasPermission(ctx context.Context, userID uint64, code string) bool {
	if code == "" {
		return true
	}
	codes, err := s.users.GetPermissionCodes(ctx, userID)
	if err != nil {
		return false
	}
	for _, item := range codes {
		if item == "*" || item == code {
			return true
		}
	}
	return false
}

func (s *Service) DataScopes(ctx context.Context, userID uint64) ([]string, error) {
	return s.users.GetDataScopes(ctx, userID)
}

func (s *Service) ListUsers(ctx context.Context, tenantID uint64, q query.PageQuery) ([]authmodel.User, int64, error) {
	return s.users.List(ctx, tenantID, q)
}

func (s *Service) CreateUser(ctx context.Context, tenantID uint64, req UserCreateRequest, operatorID uint64) (*authmodel.User, error) {
	if err := validatePasswordPolicy(req.Password); err != nil {
		return nil, err
	}
	hash, err := utils.HashPassword(req.Password)
	if err != nil {
		return nil, err
	}
	now := time.Now()
	status := defaultStatus(req.Status)
	user := &authmodel.User{
		Username: req.Username, PasswordHash: hash, RealName: req.RealName, Phone: req.Phone,
		Email: req.Email, Avatar: req.Avatar, Status: status, MustChangePassword: req.MustChangePassword, PasswordChangedAt: &now,
	}
	user.TenantID = tenantID
	user.Remark = req.Remark
	user.CreatedBy = &operatorID
	user.UpdatedBy = &operatorID
	if err := s.users.Create(ctx, user); err != nil {
		return nil, err
	}
	if err := s.users.ReplaceRoles(ctx, user.ID, req.RoleIDs); err != nil {
		return nil, err
	}
	return user, nil
}

func (s *Service) UpdateUser(ctx context.Context, tenantID uint64, id uint64, req UserUpdateRequest, operatorID uint64) (*authmodel.User, error) {
	user, err := s.users.FindByID(ctx, tenantID, id)
	if err != nil {
		return nil, err
	}
	user.RealName = req.RealName
	user.Phone = req.Phone
	user.Email = req.Email
	user.Avatar = req.Avatar
	user.MustChangePassword = req.MustChangePassword
	if user.Username == "admin" && defaultStatus(req.Status) != "active" {
		return nil, apperrors.ErrForbidden
	}
	user.Status = defaultStatus(req.Status)
	user.Remark = req.Remark
	user.UpdatedBy = &operatorID
	if req.Password != "" {
		if err := validatePasswordPolicy(req.Password); err != nil {
			return nil, err
		}
		hash, err := utils.HashPassword(req.Password)
		if err != nil {
			return nil, err
		}
		now := time.Now()
		user.PasswordHash = hash
		user.PasswordChangedAt = &now
		user.PasswordVersion++
		user.MustChangePassword = req.MustChangePassword
	}
	if err := s.users.Update(ctx, user); err != nil {
		return nil, err
	}
	if !(user.Username == "admin" && operatorID == user.ID) {
		if err := s.users.ReplaceRoles(ctx, user.ID, req.RoleIDs); err != nil {
			return nil, err
		}
	}
	return user, nil
}

func (s *Service) DeleteUser(ctx context.Context, tenantID uint64, id uint64) error {
	user, err := s.users.FindByID(ctx, tenantID, id)
	if err != nil {
		return err
	}
	if user.Username == "admin" {
		return apperrors.ErrForbidden
	}
	return s.users.Delete(ctx, tenantID, id)
}

func (s *Service) ChangePassword(ctx context.Context, tenantID uint64, userID uint64, req ChangePasswordRequest) error {
	user, err := s.users.FindByID(ctx, tenantID, userID)
	if err != nil {
		return err
	}
	if !utils.CheckPassword(user.PasswordHash, req.OldPassword) {
		return apperrors.ErrInvalidCredential
	}
	if req.OldPassword == req.NewPassword {
		return apperrors.New(40001, "新密码不能与旧密码一致", http.StatusBadRequest)
	}
	if req.ConfirmPassword != "" && req.NewPassword != req.ConfirmPassword {
		return apperrors.New(40001, "两次密码必须一致", http.StatusBadRequest)
	}
	if err := validatePasswordPolicy(req.NewPassword); err != nil {
		return err
	}
	hash, err := utils.HashPassword(req.NewPassword)
	if err != nil {
		return err
	}
	now := time.Now()
	user.PasswordHash = hash
	user.PasswordChangedAt = &now
	user.PasswordVersion++
	user.MustChangePassword = false
	return s.users.Update(ctx, user)
}

func (s *Service) ResetPassword(ctx context.Context, tenantID uint64, id uint64, req ResetPasswordRequest) error {
	user, err := s.users.FindByID(ctx, tenantID, id)
	if err != nil {
		return err
	}
	if err := validatePasswordPolicy(req.Password); err != nil {
		return err
	}
	hash, err := utils.HashPassword(req.Password)
	if err != nil {
		return err
	}
	now := time.Now()
	user.PasswordHash = hash
	user.PasswordChangedAt = &now
	user.PasswordVersion++
	user.MustChangePassword = req.MustChangePassword
	return s.users.Update(ctx, user)
}

func (s *Service) UpdateProfile(ctx context.Context, tenantID uint64, userID uint64, req ProfileRequest) (*authmodel.User, error) {
	user, err := s.users.FindByID(ctx, tenantID, userID)
	if err != nil {
		return nil, err
	}
	user.RealName = req.RealName
	user.Phone = req.Phone
	user.Email = req.Email
	user.Avatar = req.Avatar
	return user, s.users.Update(ctx, user)
}

func (s *Service) ListRoles(ctx context.Context, tenantID uint64, q query.PageQuery) ([]authmodel.Role, int64, error) {
	return s.roles.List(ctx, tenantID, q)
}

func (s *Service) CreateRole(ctx context.Context, tenantID uint64, req RoleRequest, operatorID uint64) (*authmodel.Role, error) {
	role := &authmodel.Role{ParentID: req.ParentID, Code: req.Code, Name: req.Name, DataScope: defaultDataScope(req.DataScope), DataScopeDeptIDs: req.DataScopeDeptIDs, Sort: req.Sort, Status: defaultStatus(req.Status)}
	role.TenantID = tenantID
	role.Remark = req.Remark
	role.CreatedBy = &operatorID
	role.UpdatedBy = &operatorID
	if err := s.roles.Create(ctx, role); err != nil {
		return nil, err
	}
	if err := s.roles.ReplacePermissions(ctx, role.ID, req.PermissionIDs); err != nil {
		return nil, err
	}
	if err := s.roles.ReplaceMenus(ctx, role.ID, req.MenuIDs); err != nil {
		return nil, err
	}
	return role, nil
}

func (s *Service) UpdateRole(ctx context.Context, tenantID uint64, id uint64, req RoleRequest, operatorID uint64) (*authmodel.Role, error) {
	role, err := s.roles.FindByID(ctx, tenantID, id)
	if err != nil {
		return nil, err
	}
	if role.Code == "super_admin" {
		req.Code = role.Code
		req.ParentID = nil
		req.DataScope = "all"
		req.Status = "active"
		req.PermissionIDs = nil
		req.MenuIDs = nil
	}
	role.ParentID, role.Code, role.Name, role.DataScope, role.DataScopeDeptIDs = req.ParentID, req.Code, req.Name, defaultDataScope(req.DataScope), req.DataScopeDeptIDs
	role.Sort, role.Status, role.Remark, role.UpdatedBy = req.Sort, defaultStatus(req.Status), req.Remark, &operatorID
	if err := s.roles.Update(ctx, role); err != nil {
		return nil, err
	}
	if role.Code != "super_admin" || len(req.PermissionIDs) > 0 {
		if err := s.roles.ReplacePermissions(ctx, role.ID, req.PermissionIDs); err != nil {
			return nil, err
		}
	}
	if role.Code != "super_admin" || len(req.MenuIDs) > 0 {
		if err := s.roles.ReplaceMenus(ctx, role.ID, req.MenuIDs); err != nil {
			return nil, err
		}
	}
	return role, nil
}

func (s *Service) DeleteRole(ctx context.Context, tenantID uint64, id uint64) error {
	return s.roles.Delete(ctx, tenantID, id)
}

func (s *Service) ListPermissions(ctx context.Context, tenantID uint64, q query.PageQuery) ([]authmodel.Permission, int64, error) {
	return s.permissions.List(ctx, tenantID, q)
}

func (s *Service) CreatePermission(ctx context.Context, tenantID uint64, req PermissionRequest, operatorID uint64) (*authmodel.Permission, error) {
	permission := &authmodel.Permission{Code: req.Code, Name: req.Name, Module: req.Module, Type: defaultPermissionType(req.Type), Method: req.Method, Path: req.Path, Status: defaultStatus(req.Status)}
	permission.TenantID = tenantID
	permission.Remark = req.Remark
	permission.CreatedBy = &operatorID
	permission.UpdatedBy = &operatorID
	return permission, s.permissions.Create(ctx, permission)
}

func (s *Service) UpdatePermission(ctx context.Context, tenantID uint64, id uint64, req PermissionRequest, operatorID uint64) (*authmodel.Permission, error) {
	permission, err := s.permissions.FindByID(ctx, tenantID, id)
	if err != nil {
		return nil, err
	}
	permission.Code, permission.Name, permission.Module, permission.Type, permission.Method, permission.Path = req.Code, req.Name, req.Module, defaultPermissionType(req.Type), req.Method, req.Path
	permission.Status, permission.Remark, permission.UpdatedBy = defaultStatus(req.Status), req.Remark, &operatorID
	return permission, s.permissions.Update(ctx, permission)
}

func (s *Service) DeletePermission(ctx context.Context, tenantID uint64, id uint64) error {
	return s.permissions.Delete(ctx, tenantID, id)
}

func (s *Service) ListMenus(ctx context.Context, tenantID uint64, q query.PageQuery) ([]authmodel.Menu, int64, error) {
	return s.menus.List(ctx, tenantID, q)
}

func (s *Service) MenuTree(ctx context.Context, tenantID uint64) ([]authmodel.Menu, error) {
	menus, err := s.menus.Tree(ctx, tenantID)
	if err != nil {
		return nil, err
	}
	return BuildMenuTree(menus), nil
}

func (s *Service) CreateMenu(ctx context.Context, tenantID uint64, req MenuRequest, operatorID uint64) (*authmodel.Menu, error) {
	visible := true
	if req.Visible != nil {
		visible = *req.Visible
	}
	menuType := req.Type
	if menuType == "" {
		menuType = "menu"
	}
	menu := &authmodel.Menu{ParentID: req.ParentID, Name: req.Name, Title: req.Title, Path: req.Path, Component: req.Component, Icon: req.Icon, Type: menuType, PermissionCode: req.PermissionCode, Sort: req.Sort, Visible: visible, Status: defaultStatus(req.Status)}
	menu.TenantID = tenantID
	menu.Remark = req.Remark
	menu.CreatedBy = &operatorID
	menu.UpdatedBy = &operatorID
	return menu, s.menus.Create(ctx, menu)
}

func (s *Service) UpdateMenu(ctx context.Context, tenantID uint64, id uint64, req MenuRequest, operatorID uint64) (*authmodel.Menu, error) {
	menu, err := s.menus.FindByID(ctx, tenantID, id)
	if err != nil {
		return nil, err
	}
	menu.ParentID, menu.Name, menu.Title, menu.Path, menu.Component, menu.Icon = req.ParentID, req.Name, req.Title, req.Path, req.Component, req.Icon
	menu.Type, menu.PermissionCode, menu.Sort, menu.Status, menu.Remark, menu.UpdatedBy = req.Type, req.PermissionCode, req.Sort, defaultStatus(req.Status), req.Remark, &operatorID
	if req.Visible != nil {
		menu.Visible = *req.Visible
	}
	return menu, s.menus.Update(ctx, menu)
}

func (s *Service) DeleteMenu(ctx context.Context, tenantID uint64, id uint64) error {
	return s.menus.Delete(ctx, tenantID, id)
}

func (s *Service) ListLoginLogs(ctx context.Context, tenantID uint64, q query.PageQuery) ([]authmodel.LoginLog, int64, error) {
	return s.audit.ListLoginLogs(ctx, tenantID, q)
}

func (s *Service) ListOperationLogs(ctx context.Context, tenantID uint64, q query.PageQuery) ([]authmodel.OperationLog, int64, error) {
	return s.audit.ListOperationLogs(ctx, tenantID, q)
}

func defaultStatus(status string) string {
	if status == "" {
		return "active"
	}
	return status
}

func defaultDataScope(scope string) string {
	if scope == "" {
		return "all"
	}
	return scope
}

func defaultPermissionType(permissionType string) string {
	if permissionType == "" {
		return "api"
	}
	return permissionType
}

func validatePasswordPolicy(password string) error {
	if len(password) < 8 {
		return apperrors.New(40001, "新密码不少于8位", http.StatusBadRequest)
	}
	var hasLetter bool
	var hasDigit bool
	for _, r := range password {
		if unicode.IsLetter(r) {
			hasLetter = true
		}
		if unicode.IsDigit(r) {
			hasDigit = true
		}
	}
	if !hasLetter || !hasDigit {
		return apperrors.New(40001, "密码必须包含字母和数字", http.StatusBadRequest)
	}
	return nil
}

func BuildMenuTree(menus []authmodel.Menu) []authmodel.Menu {
	nodes := make(map[uint64]*authmodel.Menu, len(menus))
	roots := make([]authmodel.Menu, 0)
	for i := range menus {
		item := menus[i]
		item.Children = []authmodel.Menu{}
		nodes[item.ID] = &item
	}
	for _, node := range nodes {
		if node.ParentID == nil || *node.ParentID == 0 {
			continue
		}
		if parent, ok := nodes[*node.ParentID]; ok {
			parent.Children = append(parent.Children, *node)
		}
	}
	for _, node := range nodes {
		if node.ParentID == nil || *node.ParentID == 0 {
			roots = append(roots, *node)
			continue
		}
		if _, ok := nodes[*node.ParentID]; !ok {
			roots = append(roots, *node)
		}
	}
	return roots
}
