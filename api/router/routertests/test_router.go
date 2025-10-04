package routertests

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/jgfranco17/echoris/api/logging"
	"github.com/jgfranco17/echoris/api/router"
	"github.com/jgfranco17/echoris/api/router/system"
	v0 "github.com/jgfranco17/echoris/api/router/v0"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type ExampleHttpRequest struct {
	Method         string
	Endpoint       string
	ExpectedCode   int
	Payload        string
	ExpectedFields map[string]interface{}
	ExpectedLogs   string
}

func NewBasicExampleRequest(method string, endpoint string, statusCode int) ExampleHttpRequest {
	return ExampleHttpRequest{
		Method:         method,
		Endpoint:       endpoint,
		ExpectedCode:   statusCode,
		Payload:        "",
		ExpectedFields: make(map[string]interface{}),
	}
}

func (e *ExampleHttpRequest) WithPayload(payload string) ExampleHttpRequest {
	e.Payload = payload
	return *e
}

type TestServer struct {
	service *router.Service
	logs    bytes.Buffer
}

/*
Create a new Test Server.

Uses chaining-builder pattern to define routes.
*/
func NewTestServer(port int) *TestServer {
	baseRouter := gin.Default()

	var buf bytes.Buffer
	logger := logging.New(&buf)
	baseRouter.Use(func(c *gin.Context) {
		c.Set(string(logging.Logger), logger)
	})
	return &TestServer{
		service: &router.Service{
			Router: baseRouter,
			Port:   port,
		},
		logs: buf,
	}
}

func (s *TestServer) WithSystemRoutes() *TestServer {
	system.SetSystemRoutes(s.service.Router)
	return s
}

func (s *TestServer) WithV0Routes() *TestServer {
	mockClient := &v0.MockLogClient{}
	v0.SetRoutes(s.service.Router, mockClient)
	return s
}

func (s *TestServer) WithV0RoutesAndClient(client v0.LogClient) *TestServer {
	v0.SetRoutes(s.service.Router, client)
	return s
}

func (s *TestServer) RunRequests(t *testing.T, sampleRequests []ExampleHttpRequest, token string) {
	t.Helper()

	for _, r := range sampleRequests {
		// Create the request with the provided method, endpoint, and body (if any)
		var request *http.Request
		if r.Payload != "" {
			request = httptest.NewRequest(r.Method, r.Endpoint, bytes.NewBuffer([]byte(r.Payload)))
		} else {
			request = httptest.NewRequest(r.Method, r.Endpoint, nil)
		}
		request.Header.Set("Content-Type", "application/json")
		if token != "" {
			request.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
		}

		recorder := httptest.NewRecorder()
		s.service.Router.ServeHTTP(recorder, request)
		assert.Equalf(t, r.ExpectedCode, recorder.Code, "Expected status %d but got %d", r.ExpectedCode, recorder.Code)

		var responseBody map[string]interface{}
		err := json.Unmarshal(recorder.Body.Bytes(), &responseBody)
		require.NoErrorf(t, err, "Failed to unmarshal JSON response body")
		for key, value := range r.ExpectedFields {
			assert.Contains(t, responseBody, key, "Response is missing key: %s", key)
			assert.Equal(t, value, responseBody[key], "Expected value for key '%s'", key)
		}

		if r.ExpectedLogs != "" {
			assert.Contains(t, s.logs.String(), r.ExpectedLogs, "Expected logs not found")
		}
	}
}
