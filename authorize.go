package grok

import (
	"bytes"
	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"
)

const (
	PARTNERS_SCOPE = "read:partners"
)

// Authorize ...
// Deprecated: Use TokenScopeRequired or TokenScopesRequired instead.
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

// TokenScopeRequired ...
func TokenScopeRequired(scope string) gin.HandlerFunc {
	return TokenScopesRequired([]string{scope})
}

// TokenScopesRequired ...
func TokenScopesRequired(scopes []string) gin.HandlerFunc {
	return func(c *gin.Context) {
		permissions, exists := c.Get("permissions")

		if !exists {
			c.AbortWithStatus(http.StatusForbidden)
			return
		}

		valid := false

		for pos, elem := range scopes {
			internalValid := false
			for _, permission := range permissions.([]interface{}) {
				if permission == elem {
					internalValid = true
				}
			}
			if !internalValid {
				valid = false
				return
			} else if pos == len(scopes)-1 {
				valid = true
			}
		}

		if valid {
			c.Next()
			return
		}

		c.AbortWithStatus(http.StatusForbidden)
	}
}

type InternalAuthorize interface {
	Authorize(scope string) gin.HandlerFunc // deprecated
	PermissionRequired(scope string) gin.HandlerFunc
	PermissionsRequired(scopes []string) gin.HandlerFunc
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

// Authorize ...
// Deprecated: Use PermissionRequired or PermissionsRequired instead.
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

// PermissionRequired ...
func (a *APIAuthorize) PermissionRequired(scope string) gin.HandlerFunc {
	return a.PermissionsRequired([]string{scope})
}

// PermissionsRequired ...
func (a *APIAuthorize) PermissionsRequired(scopes []string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if a.settings == nil {
			c.AbortWithStatus(http.StatusForbidden)
			return
		}

		valid := true

		jwt := c.Request.Header.Get("authorization")
		currentIdentity := c.Request.Header.Get("X-Current-Identity")

		for _, elemScope := range scopes {
			if !a.verifyAuthorizationPermission(elemScope, jwt, currentIdentity) {
				valid = false
				c.AbortWithStatus(http.StatusForbidden)
				return
			}
		}

		if !valid {
			c.AbortWithStatus(http.StatusForbidden)
		}

		c.Next()
	}
}

// IsPartner ...
func IsPartner(c *gin.Context) bool {
	permissions, exists := c.Get("permissions")

	if !exists {
		return false
	}

	for _, permission := range permissions.([]interface{}) {
		if permission == PARTNERS_SCOPE {
			return true
		}
	}

	return false
}

// verifyAuthorizationPermission ...
func (a *APIAuthorize) verifyAuthorizationPermission(scope string, jwt string, currentIdentity string) bool {
	payload := struct {
		Permission string `json:"permission,omitempty"`
	}{
		Permission: scope,
	}

	b, err := json.Marshal(payload)
	if err != nil {
		return false
	}

	req, err := http.NewRequest("POST", a.settings.URL, bytes.NewReader(b))
	if err != nil {
		return false
	}

	//// jwt := c.Request.Header.Get("authorization")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", jwt)
	//// req.Header.Set("X-Current-Identity", c.Request.Header.Get("X-Current-Identity"))
	req.Header.Set("X-Current-Identity", currentIdentity)

	client := http.Client{}

	resp, err := client.Do(req)
	if err != nil {
		return false
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return false
	}

	return true
}
