package grok

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type FakeTrustedDevice struct {
	alwaysSuccess bool
}

func NewFakeTrustedDevice(success bool) TrustedDevice {
	return &FakeTrustedDevice{alwaysSuccess: success}
}

func (a *FakeTrustedDevice) Validate() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !a.alwaysSuccess {
			c.AbortWithStatusJSON(http.StatusForbidden, Error{
				Code:     http.StatusForbidden,
				Key:      "DEVICE_NOT_TRUSTED",
				Messages: []string{"aparelho não cadastrado ou sem permissão"},
			})
			return
		}
		c.Next()
	}
}
