package security

import (
	"time"

	"erp/internal/config"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type Claims struct {
	UserID          uint64 `json:"userId"`
	TenantID        uint64 `json:"tenantId"`
	Username        string `json:"username"`
	TokenUse        string `json:"tokenUse"`
	PasswordVersion int    `json:"passwordVersion"`
	jwt.RegisteredClaims
}

type JWTManager struct {
	secret        []byte
	accessExpire  time.Duration
	refreshExpire time.Duration
}

func NewJWTManager(cfg config.JWTConfig) *JWTManager {
	return &JWTManager{
		secret:        []byte(cfg.Secret),
		accessExpire:  time.Duration(cfg.AccessExpireMinutes) * time.Minute,
		refreshExpire: time.Duration(cfg.RefreshExpireHours) * time.Hour,
	}
}

func (m *JWTManager) Generate(userID uint64, tenantID uint64, username string, passwordVersion int) (string, string, int64, error) {
	access, err := m.generate(userID, tenantID, username, "access", m.accessExpire, passwordVersion)
	if err != nil {
		return "", "", 0, err
	}
	refresh, err := m.generate(userID, tenantID, username, "refresh", m.refreshExpire, passwordVersion)
	if err != nil {
		return "", "", 0, err
	}
	return access, refresh, int64(m.accessExpire.Seconds()), nil
}

func (m *JWTManager) Parse(tokenValue string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenValue, &Claims{}, func(token *jwt.Token) (any, error) {
		return m.secret, nil
	})
	if err != nil {
		return nil, err
	}
	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, jwt.ErrTokenInvalidClaims
	}
	return claims, nil
}

func (m *JWTManager) generate(userID uint64, tenantID uint64, username string, tokenUse string, expire time.Duration, passwordVersion int) (string, error) {
	now := time.Now()
	claims := Claims{
		UserID:          userID,
		TenantID:        tenantID,
		Username:        username,
		TokenUse:        tokenUse,
		PasswordVersion: passwordVersion,
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        uuid.NewString(),
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(expire)),
		},
	}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(m.secret)
}
