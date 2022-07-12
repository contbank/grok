package grok

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/gin-gonic/gin"
)

const (
	// TransactionTokenHeader ...
	TransactionTokenHeader = "X-Transaction-Token"
	// CurrentIdentityHeader ...
	CurrentIdentityHeader = "X-Current-Identity"
)

type TransactionalToken interface {
	Validate() gin.HandlerFunc
}

type InternalTransactionalToken struct {
	settings *TransactionalTokenSettings
}

// CreateTransactionalToken ...
func CreateTransactionalToken(settings *TransactionalTokenSettings) TransactionalToken {
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

		defaultError := Error{
			Code:     http.StatusBadRequest,
			Key:      "INVALID_PASSWORD",
			Messages: []string{"invalid password"},
		}

		if a.settings == nil {
			c.AbortWithStatusJSON(http.StatusForbidden, defaultError)
			return
		}

		// get token and current identity from header
		token, currentIdentity, err := getHeaderParameters(c)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusForbidden, err)
			return
		}

		payload := struct {
			Permission string `json:"password,omitempty"`
		}{
			Permission: *token,
		}

		b, err := json.Marshal(payload)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusForbidden, defaultError)
			return
		}

		// passwords api
		req, err := http.NewRequest("POST", a.settings.URL, bytes.NewReader(b))
		if err != nil {
			c.AbortWithStatusJSON(http.StatusForbidden, defaultError)
			return
		}

		// setting headers
		jwt := c.Request.Header.Get("authorization")
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", jwt)
		if currentIdentity != nil {
			req.Header.Set("X-Current-Identity", *currentIdentity)
		}

		client := http.Client{}
		resp, err := client.Do(req)

		if err != nil {
			c.AbortWithStatusJSON(http.StatusForbidden, defaultError)
			return
		}

		defer resp.Body.Close()

		if resp.StatusCode == http.StatusOK {
			c.Next()
			return
		}

		if resp.StatusCode == http.StatusConflict || resp.StatusCode == http.StatusBadRequest {

			var response Error
			body, err := ioutil.ReadAll(resp.Body)

			if err != nil {
				c.AbortWithStatusJSON(http.StatusForbidden, defaultError)
				return
			}

			err = json.Unmarshal(body, &response)

			if err != nil {
				c.AbortWithStatusJSON(http.StatusForbidden, defaultError)
				return
			}

			c.AbortWithStatusJSON(resp.StatusCode, response)
			return
		}

		c.AbortWithStatusJSON(http.StatusForbidden, defaultError)
	}
}

// getHeaderParameters ...
func getHeaderParameters(c *gin.Context) (*string, *string, error) {
	defaultError := Error{
		Code: http.StatusBadRequest,
	}

	// get token
	token := c.Request.Header.Get(TransactionTokenHeader)
	if len(token) <= 0 {
		defaultError.Key = "INVALID_PASSWORD"
		defaultError.Messages = []string{"invalid password"}
		return nil, nil, &defaultError
	}

	// get current identity
	currentIdentity := c.Request.Header.Get(CurrentIdentityHeader)
	if len(currentIdentity) <= 0 {
		// TODO Não permitir currentIdentity vazio (apenas após o front alterar para sempre enviar o X-Current-Identity)
		//// defaultError.Messages = []string{"invalid current identity"}
		//// return nil, nil, &defaultError
		return &token, nil, nil
	} else {
		currentIdentity = OnlyDigits(currentIdentity)
	}

	return &token, &currentIdentity, nil
}
