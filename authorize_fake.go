package grok

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// FakeAuthorize ...
type FakeAuthorize struct {
	alwaysSuccess bool
}

// NewFakeAuthorize ...
func NewFakeAuthorize(success bool) InternalAuthorize {
	return &FakeAuthorize{
		alwaysSuccess: success,
	}
}

// Authorize ...
// Deprecated: Use PermissionRequired or PermissionsRequired instead.
func (a *FakeAuthorize) Authorize(scope string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !a.alwaysSuccess {
			c.AbortWithStatus(http.StatusForbidden)
			return
		}
		c.Next()
	}
}

// PermissionRequired ...
func (a *FakeAuthorize) PermissionRequired(scope string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !a.alwaysSuccess {
			c.AbortWithStatus(http.StatusForbidden)
			return
		}
		c.Next()
	}
}

// PermissionsRequired ...
func (a *FakeAuthorize) PermissionsRequired(scopes []string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !a.alwaysSuccess {
			c.AbortWithStatus(http.StatusForbidden)
			return
		}
		c.Next()
	}
}
