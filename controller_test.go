package grok_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/contbank/grok"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type APIControllerTestSuite struct {
	suite.Suite
	assert   *assert.Assertions
	settings *grok.Settings
	server   *grok.API
}

type testContainer struct{}

func (c *testContainer) Controllers() []grok.APIController {
	return nil
}

func (c *testContainer) Close() error {
	return nil
}

func TestAPIControllerTestSuite(t *testing.T) {
	suite.Run(t, new(APIControllerTestSuite))
}

func (s *APIControllerTestSuite) SetupTest() {
	container := &testContainer{}
	s.assert = assert.New(s.T())
	s.settings = &grok.Settings{}
	grok.FromYAML("tests/config.yaml", s.settings)
	s.server = grok.New(
		grok.WithSettings(s.settings),
		grok.WithSwagger(TestSwaggerSpec, "/custom/swagger/"),
		grok.WithContainer(container))
}

func (s *APIControllerTestSuite) TestNotFound() {
	req := httptest.NewRequest("GET", "/notfound", nil)
	response := httptest.NewRecorder()

	s.server.Engine.ServeHTTP(response, req)

	s.assert.Equal(http.StatusNotFound, response.Code)
}

func (s *APIControllerTestSuite) TestPostNotFound() {
	reader := strings.NewReader("{\"secret\": \"A secret value\", \"test\": \"test\"}")
	req := httptest.NewRequest("POST", "/notfound", reader)
	response := httptest.NewRecorder()

	s.server.Engine.ServeHTTP(response, req)

	s.assert.Equal(http.StatusNotFound, response.Code)
}

func (s *APIControllerTestSuite) TestSwagger() {
	req := httptest.NewRequest("GET", "/custom/swagger/index.html", nil)
	response := httptest.NewRecorder()

	s.server.Engine.ServeHTTP(response, req)

	s.assert.Equal(http.StatusOK, response.Code)
}
