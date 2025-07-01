package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

// IntegrationTestSuite contains integration tests
type IntegrationTestSuite struct {
	suite.Suite
	router *gin.Engine
}

func (suite *IntegrationTestSuite) SetupSuite() {
	suite.router = setupRouter()
}

func (suite *IntegrationTestSuite) TestCompleteAPIWorkflow() {
	// Test 1: Check server health
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/health", nil)
	suite.router.ServeHTTP(w, req)
	assert.Equal(suite.T(), 200, w.Code)

	// Test 2: Get available models
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/api/tags", nil)
	suite.router.ServeHTTP(w, req)
	assert.Equal(suite.T(), 200, w.Code)

	var tagsResponse TagsResponse
	err := json.Unmarshal(w.Body.Bytes(), &tagsResponse)
	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), tagsResponse.Models, 1)
	assert.Equal(suite.T(), "amazon-q:latest", tagsResponse.Models[0].Name)

	// Test 3: Show model details
	showReq := ShowRequest{Name: "amazon-q"}
	jsonData, _ := json.Marshal(showReq)
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("POST", "/api/show", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	suite.router.ServeHTTP(w, req)
	assert.Equal(suite.T(), 200, w.Code)

	// Test 4: Check running processes
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/api/ps", nil)
	suite.router.ServeHTTP(w, req)
	assert.Equal(suite.T(), 200, w.Code)

	// Test 5: Check server status
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/api/status", nil)
	suite.router.ServeHTTP(w, req)
	assert.Equal(suite.T(), 200, w.Code)

	var statusResponse StatusResponse
	err = json.Unmarshal(w.Body.Bytes(), &statusResponse)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "running", statusResponse.Status)
}

func (suite *IntegrationTestSuite) TestAllEndpointsRespond() {
	endpoints := []struct {
		method   string
		path     string
		body     interface{}
		expected int
	}{
		{"GET", "/", nil, 200},
		{"GET", "/health", nil, 200},
		{"GET", "/ping", nil, 200},
		{"HEAD", "/", nil, 200},
		{"GET", "/metrics", nil, 200},
		{"GET", "/api/version", nil, 200},
		{"GET", "/api/tags", nil, 200},
		{"GET", "/api/list", nil, 200},
		{"GET", "/api/ps", nil, 200},
		{"GET", "/api/status", nil, 200},
		{"POST", "/api/show", ShowRequest{Name: "amazon-q"}, 200},
		{"POST", "/api/create", CreateRequest{Name: "test"}, 501},
		{"POST", "/api/pull", PullRequest{Name: "test"}, 501},
		{"POST", "/api/push", PushRequest{Name: "test"}, 501},
		{"DELETE", "/api/delete", DeleteRequest{Name: "test"}, 501},
		{"POST", "/api/copy", CopyRequest{Source: "a", Destination: "b"}, 501},
		{"POST", "/api/embeddings", EmbeddingsRequest{Model: "amazon-q", Prompt: "test"}, 501},
		{"POST", "/api/embed", EmbeddingsRequest{Model: "amazon-q", Prompt: "test"}, 501},
		{"GET", "/api/blobs/sha256:test", nil, 404},
		{"HEAD", "/api/blobs/sha256:test", nil, 404},
		{"POST", "/api/blobs/sha256:test", nil, 501},
	}

	for _, endpoint := range endpoints {
		suite.T().Run(fmt.Sprintf("%s %s", endpoint.method, endpoint.path), func(t *testing.T) {
			var reqBody *bytes.Buffer
			if endpoint.body != nil {
				jsonData, _ := json.Marshal(endpoint.body)
				reqBody = bytes.NewBuffer(jsonData)
			}

			w := httptest.NewRecorder()
			var req *http.Request
			if reqBody != nil {
				req, _ = http.NewRequest(endpoint.method, endpoint.path, reqBody)
				req.Header.Set("Content-Type", "application/json")
			} else {
				req, _ = http.NewRequest(endpoint.method, endpoint.path, nil)
			}

			suite.router.ServeHTTP(w, req)
			assert.Equal(t, endpoint.expected, w.Code, "Endpoint %s %s should return %d", endpoint.method, endpoint.path, endpoint.expected)
		})
	}
}

func (suite *IntegrationTestSuite) TestCORSFunctionality() {
	// Test CORS preflight
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("OPTIONS", "/api/generate", nil)
	req.Header.Set("Origin", "http://localhost:3000")
	req.Header.Set("Access-Control-Request-Method", "POST")
	req.Header.Set("Access-Control-Request-Headers", "Content-Type")

	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), 204, w.Code)
	assert.Equal(suite.T(), "*", w.Header().Get("Access-Control-Allow-Origin"))
	assert.Contains(suite.T(), w.Header().Get("Access-Control-Allow-Methods"), "POST")
	assert.Contains(suite.T(), w.Header().Get("Access-Control-Allow-Headers"), "Content-Type")

	// Test actual CORS request
	genReq := GenerateRequest{Model: "amazon-q", Prompt: "test"}
	jsonData, _ := json.Marshal(genReq)
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("POST", "/api/generate", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Origin", "http://localhost:3000")

	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), "*", w.Header().Get("Access-Control-Allow-Origin"))
}

func (suite *IntegrationTestSuite) TestErrorHandling() {
	// Test invalid JSON
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/generate", bytes.NewBufferString("invalid json"))
	req.Header.Set("Content-Type", "application/json")
	suite.router.ServeHTTP(w, req)
	assert.Equal(suite.T(), 400, w.Code)

	// Test missing required fields
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("POST", "/api/chat", bytes.NewBufferString("{}"))
	req.Header.Set("Content-Type", "application/json")
	suite.router.ServeHTTP(w, req)
	assert.Equal(suite.T(), 400, w.Code)

	// Test chat without user message
	chatReq := ChatRequest{
		Model: "amazon-q",
		Messages: []Message{
			{Role: "system", Content: "You are helpful"},
		},
	}
	jsonData, _ := json.Marshal(chatReq)
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("POST", "/api/chat", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	suite.router.ServeHTTP(w, req)
	assert.Equal(suite.T(), 400, w.Code)

	var errorResponse map[string]string
	err := json.Unmarshal(w.Body.Bytes(), &errorResponse)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "No user message found", errorResponse["error"])
}

func (suite *IntegrationTestSuite) TestResponseFormats() {
	// Test JSON responses
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/tags", nil)
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), 200, w.Code)
	assert.Equal(suite.T(), "application/json; charset=utf-8", w.Header().Get("Content-Type"))

	var response TagsResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)

	// Test text responses
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/ping", nil)
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), 200, w.Code)
	assert.Equal(suite.T(), "text/plain; charset=utf-8", w.Header().Get("Content-Type"))
	assert.Equal(suite.T(), "pong", w.Body.String())

	// Test metrics format
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/metrics", nil)
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), 200, w.Code)
	assert.Contains(suite.T(), w.Body.String(), "amazon_q_ollama_up 1")
}

func (suite *IntegrationTestSuite) TestPerformance() {
	// Test response times for critical endpoints
	endpoints := []string{"/health", "/ping", "/api/tags", "/api/ps"}

	for _, endpoint := range endpoints {
		suite.T().Run(fmt.Sprintf("Performance_%s", endpoint), func(t *testing.T) {
			start := time.Now()
			w := httptest.NewRecorder()
			req, _ := http.NewRequest("GET", endpoint, nil)
			suite.router.ServeHTTP(w, req)
			duration := time.Since(start)

			assert.Equal(t, 200, w.Code)
			assert.Less(t, duration, 50*time.Millisecond, "Endpoint %s should respond quickly", endpoint)
		})
	}
}

func (suite *IntegrationTestSuite) TestConcurrentRequests() {
	// Test concurrent health checks
	const numRequests = 10
	results := make(chan int, numRequests)

	for i := 0; i < numRequests; i++ {
		go func() {
			w := httptest.NewRecorder()
			req, _ := http.NewRequest("GET", "/health", nil)
			suite.router.ServeHTTP(w, req)
			results <- w.Code
		}()
	}

	// Collect results
	for i := 0; i < numRequests; i++ {
		select {
		case code := <-results:
			assert.Equal(suite.T(), 200, code)
		case <-time.After(5 * time.Second):
			suite.T().Fatal("Request timed out")
		}
	}
}

func TestIntegrationSuite(t *testing.T) {
	suite.Run(t, new(IntegrationTestSuite))
}
