package grok_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/contbank/grok"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	logrustest "github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/assert"
)

func TestLogMiddlewareRestrictsAccessTokenAndAuthorization(t *testing.T) {
	gin.SetMode(gin.TestMode)
	hook := logrustest.NewGlobal()
	defer hook.Reset()

	engine := gin.New()
	engine.Use(grok.LogMiddleware([]string{grok.TransactionTokenHeader, "Authorization"}))
	engine.GET("/test", grok.SetAccessTokenInContext(), func(c *gin.Context) {
		c.Set("user_id", "61096795f3473c650f276e3b")
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer secret-jwt")
	req.Header.Set(grok.TransactionTokenHeader, "9876")
	engine.ServeHTTP(httptest.NewRecorder(), req)

	entry := hook.LastEntry()
	if !assert.NotNil(t, entry) {
		return
	}

	b, err := json.Marshal(entry.Data)
	assert.NoError(t, err)
	logged := string(b)

	assert.False(t, strings.Contains(logged, "secret-jwt"), logged)
	assert.False(t, strings.Contains(logged, "9876"), logged)
	assert.True(t, strings.Contains(logged, "61096795f3473c650f276e3b"), "outras claims continuam no log")

	claims := entry.Data["claims"].(map[string]interface{})
	assert.Equal(t, "RESTRICTED", claims["access_token"])
	assert.Equal(t, logrus.InfoLevel, entry.Level)
}
