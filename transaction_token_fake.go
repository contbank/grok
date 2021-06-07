package grok

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// FakeTransactionToken ...
type FakeTransactionToken struct {
	alwaysSuccess bool
}

// NewFakeTransactionToken ...
func NewFakeTransactionToken(success bool) TransactionalToken {
	return &FakeTransactionToken{
		alwaysSuccess: success,
	}
}

// Validate ...
func (a *FakeTransactionToken) Validate() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !a.alwaysSuccess {
			c.AbortWithStatus(http.StatusForbidden)
			return
		}
		c.Next()
	}
}
