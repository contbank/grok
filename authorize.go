package grok

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

const (
	PARTNERS_SCOPE     = "read:partners"
	X_CURRENT_IDENTITY = "X-Current-Identity"
	ACCOUNT_ID_PARAM   = "account_id"
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

		if a.RequestFullPathHasAccountID(c) {
			if a.settings.URLs == nil || len(a.settings.URLs) < 2 {
				c.AbortWithStatus(http.StatusInternalServerError)
				return
			}

			accountID := c.Param(ACCOUNT_ID_PARAM)
			response, err := a.GetAccounts(c, accountID, *a.settings.URLs[1])
			if err != nil {
				c.AbortWithStatus(http.StatusForbidden)
				return
			}

			defer response.Body.Close()

			if response.StatusCode != http.StatusOK {
				c.AbortWithStatus(http.StatusForbidden)
				return
			}

			identifier := new(string)
			responseBody, _ := ioutil.ReadAll(response.Body)
			if err := json.Unmarshal(responseBody, identifier); err != nil {
				c.AbortWithStatus(http.StatusForbidden)
				return
			}

			c.Request.Header.Set(X_CURRENT_IDENTITY, *identifier)
		}

		// current identity is required
		currentIdentity := c.Request.Header.Get(X_CURRENT_IDENTITY)
		if len(currentIdentity) == 0 {
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}

		var url *string
		if !a.RequestFullPathHasAccountID(c) && a.settings.URL != nil {
			url = a.settings.URL
		} else {
			url = a.settings.URLs[0]
		}

		jwt := c.Request.Header.Get("authorization")

		for _, elemScope := range scopes {
			if !a.verifyAuthorizationPermission(elemScope, jwt, currentIdentity, *url) {
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
func (a *APIAuthorize) verifyAuthorizationPermission(scope string, jwt string,
	currentIdentity string, url string) bool {

	payload := struct {
		Permission string `json:"permission,omitempty"`
	}{
		Permission: scope,
	}

	b, err := json.Marshal(payload)
	if err != nil {
		return false
	}

	req, err := http.NewRequest("POST", url, bytes.NewReader(b))
	if err != nil {
		return false
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", jwt)
	req.Header.Set(X_CURRENT_IDENTITY, currentIdentity)

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

func (a *APIAuthorize) PostAuthorization(c *gin.Context, scope string, URL string) (*http.Response, error) {

	payload := struct {
		Permission string `json:"permission,omitempty"`
	}{
		Permission: scope,
	}

	b, err := json.Marshal(payload)
	if err != nil {
		return nil, NewError(http.StatusForbidden, "SCOPE_NOT_FOUND", "error token scope not found")
	}

	req, err := http.NewRequest("POST", URL, bytes.NewReader(b))
	if err != nil {
		return nil, NewError(http.StatusForbidden, "ERROR_POST", "error on post to authorizations")
	}

	jwt := c.Request.Header.Get("authorization")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", jwt)
	req.Header.Set(X_CURRENT_IDENTITY, c.Request.Header.Get(X_CURRENT_IDENTITY))

	client := http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func (a *APIAuthorize) GetAccounts(c *gin.Context, accountID string, URL string) (*http.Response, error) {

	newURL := strings.Replace(URL, ":account_id", accountID, -1)
	req, err := http.NewRequest("GET", newURL, nil)
	if err != nil {
		return nil, NewError(http.StatusForbidden, "ERROR_GET", "error on get to accounts")
	}

	jwt := c.Request.Header.Get("authorization")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", jwt)

	client := http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func (a *APIAuthorize) RequestFullPathHasAccountID(c *gin.Context) bool {
	if len(c.Param(ACCOUNT_ID_PARAM)) != 0 {
		return true
	}
	return false
}
