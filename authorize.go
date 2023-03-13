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
	PARTNERS_SCOPE = "read:partners"
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

		accountID := c.Param("account_id")
		if len(accountID) != 0 {

			response, err := getAccounts(c, accountID, a.settings.URLs[1])
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

			//c.Header("X-Current-Identity", *identifier)
			c.Request.Header.Set("X-Current-Identity", *identifier)
		}

		response, err := postAuthorization(c, scope, a.settings.URLs[0])
		if err != nil {
			c.AbortWithStatus(http.StatusForbidden)
			return
		}

		defer response.Body.Close()

		if response.StatusCode != http.StatusOK {
			c.AbortWithStatus(http.StatusForbidden)
			return
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

func postAuthorization(c *gin.Context, scope string, URL string) (*http.Response, error) {

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
	req.Header.Set("X-Current-Identity", c.Request.Header.Get("X-Current-Identity"))

	client := http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func getAccounts(c *gin.Context, accountID string, URL string) (*http.Response, error) {

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
