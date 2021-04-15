package grok

import (
	"bytes"
	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"
)

// Authorize ...
func Authorize(scope string) gin.HandlerFunc {
	return func(c *gin.Context) {
		permissions, exists := c.Get("permissions")

		if !exists {
			c.AbortWithStatus(http.StatusForbidden)
			return
		}

		for _, permission := range permissions.([]interface{}) {
			if permission == scope {
				c.Next()
				return
			}
		}

		c.AbortWithStatus(http.StatusForbidden)
	}
}

type InternalAuthorize interface {
	Authorize(scope string) gin.HandlerFunc
}

type APIAuthorize struct {
	settings *InternalAuth
}

// CreateAuthorize ...
func CreateAuthorize(settings *InternalAuth) InternalAuthorize {
	if settings.Fake {
		success := true
		if settings.Success != nil {
			success = *settings.Success
		}
		return NewFakeAuthorize(success)
	}

	return NewInternalAuthorize(settings)
}

func NewInternalAuthorize(settings *InternalAuth) InternalAuthorize {
	return &APIAuthorize{
		settings: settings,
	}
}

func (a *APIAuthorize) Authorize(scope string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if a.settings == nil {
			c.AbortWithStatus(http.StatusForbidden)
			return
		}

		payload := struct {
			Permission string `json:"permission,omitempty"`
		}{
			Permission: scope,
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
		req.Header.Set("X-Current-Identity", c.Request.Header.Get("X-Current-Identity"))

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
