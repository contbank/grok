package grok

import (
	"bytes"
	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"
)

type TransactionalToken interface {
	Validate() gin.HandlerFunc
}

type InternalTransactionalToken struct {
	settings *TransactionalTokenSettings
}

// CreateAuthorize ...
func CreateTransactionalPassword(settings *TransactionalTokenSettings) TransactionalToken {
	if settings.Fake {
		success := true
		if settings.Success != nil {
			success = *settings.Success
		}
		return NewFakeTransactionToken(success)
	}

	return NewInternalTransactionalToken(settings)
}

func NewInternalTransactionalToken(settings *TransactionalTokenSettings) TransactionalToken {
	return &InternalTransactionalToken{
		settings: settings,
	}
}

func (a *InternalTransactionalToken) Validate() gin.HandlerFunc {
	return func(c *gin.Context) {
		if a.settings == nil {
			c.AbortWithStatus(http.StatusForbidden)
			return
		}

		token := c.Request.Header.Get("X-Transaction-Token")

		if len(token) <= 0 {
			c.AbortWithStatus(http.StatusForbidden)
			return
		}

		payload := struct {
			Permission string `json:"password,omitempty"`
		}{
			Permission: token,
		}

		b, err := json.Marshal(payload)
		if err != nil {
			c.AbortWithStatus(http.StatusForbidden)
			return
		}

		req, err := http.NewRequest("POST", a.settings.URL, bytes.NewReader(b))
		if err != nil {
			c.AbortWithStatus(http.StatusForbidden)
			return
		}

		jwt := c.Request.Header.Get("authorization")
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", jwt)

		client := http.Client{}
		resp, err := client.Do(req)

		if err != nil {
			c.AbortWithStatus(http.StatusForbidden)
			return
		}

		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			c.AbortWithStatus(http.StatusForbidden)
			return
		}

		c.Next()
	}
}
