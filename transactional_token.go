package grok

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/gin-gonic/gin"
)

const (
	//TransactionTokenHeader ...
	TransactionTokenHeader = "X-Transaction-Token"
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
			Messages: []string{"invalid password"},
		}

		if a.settings == nil {
			c.AbortWithStatusJSON(http.StatusForbidden, defaultError)
			return
		}

		token := c.Request.Header.Get(TransactionTokenHeader)

		if len(token) <= 0 {
			c.AbortWithStatusJSON(http.StatusForbidden, defaultError)
			return
		}

		payload := struct {
			Permission string `json:"password,omitempty"`
		}{
			Permission: token,
		}

		b, err := json.Marshal(payload)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusForbidden, defaultError)
			return
		}

		req, err := http.NewRequest("POST", a.settings.URL, bytes.NewReader(b))
		if err != nil {
			c.AbortWithStatusJSON(http.StatusForbidden, defaultError)
			return
		}

		jwt := c.Request.Header.Get("authorization")
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", jwt)

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
