package utils

import (
	"strconv"

	"github.com/gin-gonic/gin"
)

const (
	ContextUserID     = "userID"
	ContextTenantID   = "tenantID"
	ContextUsername   = "username"
	ContextDataScopes = "dataScopes"
)

func GetUserID(c *gin.Context) uint64 {
	value, exists := c.Get(ContextUserID)
	if !exists {
		return 0
	}
	switch v := value.(type) {
	case uint64:
		return v
	case string:
		id, _ := strconv.ParseUint(v, 10, 64)
		return id
	default:
		return 0
	}
}

func GetDataScopes(c *gin.Context) []string {
	value, exists := c.Get(ContextDataScopes)
	if !exists {
		return []string{}
	}
	scopes, ok := value.([]string)
	if !ok {
		return []string{}
	}
	return scopes
}

func GetTenantID(c *gin.Context) uint64 {
	value, exists := c.Get(ContextTenantID)
	if !exists {
		return 1
	}
	if id, ok := value.(uint64); ok && id > 0 {
		return id
	}
	return 1
}
