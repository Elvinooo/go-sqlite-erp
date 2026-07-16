package auth

type LoginRequest struct {
	Username string `json:"username" binding:"required,min=2,max=64"`
	Password string `json:"password" binding:"required,min=6,max=64"`
}

type RefreshTokenRequest struct {
	RefreshToken string `json:"refreshToken" binding:"required"`
}

type TokenResponse struct {
	AccessToken        string `json:"accessToken"`
	RefreshToken       string `json:"refreshToken"`
	TokenType          string `json:"tokenType"`
	ExpiresIn          int64  `json:"expiresIn"`
	MustChangePassword bool   `json:"mustChangePassword"`
}

type CurrentUserResponse struct {
	ID                 uint64   `json:"id"`
	Username           string   `json:"username"`
	RealName           string   `json:"realName"`
	Avatar             string   `json:"avatar"`
	MustChangePassword bool     `json:"mustChangePassword"`
	Permissions        []string `json:"permissions"`
	DataScopes         []string `json:"dataScopes"`
	Menus              any      `json:"menus"`
}

type UserCreateRequest struct {
	Username           string   `json:"username" binding:"required,min=2,max=64"`
	Password           string   `json:"password" binding:"required,min=8,max=64"`
	RealName           string   `json:"realName" binding:"max=64"`
	Phone              string   `json:"phone" binding:"max=32"`
	Email              string   `json:"email" binding:"max=128"`
	Avatar             string   `json:"avatar" binding:"max=255"`
	Status             string   `json:"status"`
	MustChangePassword bool     `json:"mustChangePassword"`
	RoleIDs            []uint64 `json:"roleIds"`
	Remark             string   `json:"remark"`
}

type UserUpdateRequest struct {
	RealName           string   `json:"realName" binding:"max=64"`
	Phone              string   `json:"phone" binding:"max=32"`
	Email              string   `json:"email" binding:"max=128"`
	Avatar             string   `json:"avatar" binding:"max=255"`
	Status             string   `json:"status"`
	Password           string   `json:"password" binding:"omitempty,min=8,max=64"`
	MustChangePassword bool     `json:"mustChangePassword"`
	RoleIDs            []uint64 `json:"roleIds"`
	Remark             string   `json:"remark"`
}

type RoleRequest struct {
	ParentID         *uint64  `json:"parentId"`
	Code             string   `json:"code" binding:"required,min=2,max=64"`
	Name             string   `json:"name" binding:"required,min=2,max=64"`
	DataScope        string   `json:"dataScope"`
	DataScopeDeptIDs string   `json:"dataScopeDeptIds"`
	Sort             int      `json:"sort"`
	Status           string   `json:"status"`
	PermissionIDs    []uint64 `json:"permissionIds"`
	MenuIDs          []uint64 `json:"menuIds"`
	Remark           string   `json:"remark"`
}

type PermissionRequest struct {
	Code   string `json:"code" binding:"required,min=2,max=128"`
	Name   string `json:"name" binding:"required,min=2,max=128"`
	Module string `json:"module" binding:"required,max=64"`
	Type   string `json:"type"`
	Method string `json:"method" binding:"max=16"`
	Path   string `json:"path" binding:"max=255"`
	Status string `json:"status"`
	Remark string `json:"remark"`
}

type ChangePasswordRequest struct {
	OldPassword     string `json:"oldPassword" binding:"required,min=6,max=64"`
	NewPassword     string `json:"newPassword" binding:"required,min=8,max=64"`
	ConfirmPassword string `json:"confirmPassword" binding:"omitempty,min=8,max=64"`
}

type ResetPasswordRequest struct {
	Password           string `json:"password" binding:"required,min=8,max=64"`
	MustChangePassword bool   `json:"mustChangePassword"`
}

type ProfileRequest struct {
	RealName string `json:"realName" binding:"max=64"`
	Phone    string `json:"phone" binding:"max=32"`
	Email    string `json:"email" binding:"max=128"`
	Avatar   string `json:"avatar" binding:"max=255"`
}

type MenuRequest struct {
	ParentID       *uint64 `json:"parentId"`
	Name           string  `json:"name" binding:"required,min=2,max=64"`
	Title          string  `json:"title" binding:"required,min=2,max=64"`
	Path           string  `json:"path" binding:"max=255"`
	Component      string  `json:"component" binding:"max=255"`
	Icon           string  `json:"icon" binding:"max=64"`
	Type           string  `json:"type"`
	PermissionCode string  `json:"permissionCode" binding:"max=128"`
	Sort           int     `json:"sort"`
	Visible        *bool   `json:"visible"`
	Status         string  `json:"status"`
	Remark         string  `json:"remark"`
}
