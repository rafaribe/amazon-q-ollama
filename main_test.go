package main

import (
	"bytes"
	"encoding/json"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func setupRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.Default()
	r.Use(corsMiddleware())
	r.MaxMultipartMemory = 32 << 20

	// OLLAMA-compatible API endpoints
	api := r.Group("/api")
	{
		api.POST("/generate", handleGenerate)
		api.POST("/chat", handleChatWithStreaming)
		api.GET("/tags", handleTags)
		api.POST("/show", handleShow)
		api.POST("/create", handleCreate)
		api.POST("/pull", handlePull)
		api.POST("/push", handlePush)
		api.DELETE("/delete", handleDelete)
		api.POST("/copy", handleCopy)
		api.GET("/ps", handlePs)
		api.GET("/status", handleStatus)
		api.POST("/embeddings", handleEmbeddings)
		api.POST("/embed", handleEmbed)
		api.GET("/list", handleList)
		api.GET("/blobs/:digest", handleBlobs)
		api.HEAD("/blobs/:digest", handleBlobsHead)
		api.POST("/blobs/:digest", handleBlobsPost)
		api.GET("/version", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"version": "amazon-q-ollama-1.0.0"})
		})
	}

	r.POST("/upload", handleUpload)
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})
	r.GET("/ping", func(c *gin.Context) {
		c.String(http.StatusOK, "pong")
	})
	r.HEAD("/", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})
	r.GET("/metrics", func(c *gin.Context) {
		c.String(http.StatusOK, "# Amazon Q OLLAMA Metrics\namazon_q_ollama_up 1\n")
	})
	r.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "Amazon Q OLLAMA - OLLAMA Compatible API",
			"version": "1.0.0",
		})
	})

	return r
}

func TestHealthEndpoint(t *testing.T) {
	router := setupRouter()
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/health", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
	
	var response map[string]string
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "ok", response["status"])
}

func TestPingEndpoint(t *testing.T) {
	router := setupRouter()
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/ping", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
	assert.Equal(t, "pong", w.Body.String())
}

func TestRootEndpoint(t *testing.T) {
	router := setupRouter()
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
	
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "Amazon Q OLLAMA - OLLAMA Compatible API", response["message"])
	assert.Equal(t, "1.0.0", response["version"])
}

func TestHeadRootEndpoint(t *testing.T) {
	router := setupRouter()
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("HEAD", "/", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
	assert.Empty(t, w.Body.String())
}

func TestVersionEndpoint(t *testing.T) {
	router := setupRouter()
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/version", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
	
	var response map[string]string
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "amazon-q-ollama-1.0.0", response["version"])
}

func TestMetricsEndpoint(t *testing.T) {
	router := setupRouter()
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/metrics", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
	assert.Contains(t, w.Body.String(), "amazon_q_ollama_up 1")
}

func TestTagsEndpoint(t *testing.T) {
	router := setupRouter()
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/tags", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
	
	var response TagsResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Len(t, response.Models, 1)
	assert.Equal(t, "amazon-q:latest", response.Models[0].Name)
	assert.Equal(t, "amazon-q", response.Models[0].Model)
}

func TestListEndpoint(t *testing.T) {
	router := setupRouter()
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/list", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
	
	var response TagsResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Len(t, response.Models, 1)
}

func TestPsEndpoint(t *testing.T) {
	router := setupRouter()
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/ps", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
	
	var response PsResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Len(t, response.Models, 1)
	assert.Equal(t, "amazon-q:latest", response.Models[0].Name)
}

func TestStatusEndpoint(t *testing.T) {
	router := setupRouter()
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/status", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
	
	var response StatusResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "running", response.Status)
	assert.Len(t, response.Models, 1)
}

func TestShowEndpoint(t *testing.T) {
	router := setupRouter()
	
	showReq := ShowRequest{
		Name:    "amazon-q",
		Verbose: false,
	}
	jsonData, _ := json.Marshal(showReq)
	
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/show", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
	
	var response ShowResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Contains(t, response.Modelfile, "Amazon Q Service Model")
	assert.Equal(t, "{{ .Prompt }}", response.Template)
}

func TestGenerateEndpoint(t *testing.T) {
	router := setupRouter()
	
	genReq := GenerateRequest{
		Model:  "amazon-q",
		Prompt: "Hello world",
		Stream: false,
	}
	jsonData, _ := json.Marshal(genReq)
	
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/generate", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	// Since we don't have actual Q CLI in test environment, expect error
	assert.True(t, w.Code == 500 || w.Code == 200)
	
	if w.Code == 200 {
		var response GenerateResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "amazon-q", response.Model)
		assert.True(t, response.Done)
	}
}

func TestChatEndpoint(t *testing.T) {
	router := setupRouter()
	
	chatReq := ChatRequest{
		Model: "amazon-q",
		Messages: []Message{
			{
				Role:    "user",
				Content: "Hello",
			},
		},
		Stream: false,
	}
	jsonData, _ := json.Marshal(chatReq)
	
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/chat", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	// Since we don't have actual Q CLI in test environment, expect error
	assert.True(t, w.Code == 500 || w.Code == 200)
	
	if w.Code == 200 {
		var response ChatResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "amazon-q", response.Model)
		assert.Equal(t, "assistant", response.Message.Role)
	}
}

func TestChatEndpointNoUserMessage(t *testing.T) {
	router := setupRouter()
	
	chatReq := ChatRequest{
		Model: "amazon-q",
		Messages: []Message{
			{
				Role:    "system",
				Content: "You are a helpful assistant",
			},
		},
	}
	jsonData, _ := json.Marshal(chatReq)
	
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/chat", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, 400, w.Code)
	
	var response map[string]string
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "No user message found", response["error"])
}

func TestCreateEndpoint(t *testing.T) {
	router := setupRouter()
	
	createReq := CreateRequest{
		Name:      "test-model",
		Modelfile: "FROM amazon-q",
	}
	jsonData, _ := json.Marshal(createReq)
	
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/create", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, 501, w.Code)
	
	var response map[string]string
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Contains(t, response["error"], "not supported")
}

func TestPullEndpoint(t *testing.T) {
	router := setupRouter()
	
	pullReq := PullRequest{
		Name: "test-model",
	}
	jsonData, _ := json.Marshal(pullReq)
	
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/pull", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, 501, w.Code)
}

func TestPushEndpoint(t *testing.T) {
	router := setupRouter()
	
	pushReq := PushRequest{
		Name: "test-model",
	}
	jsonData, _ := json.Marshal(pushReq)
	
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/push", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, 501, w.Code)
}

func TestDeleteEndpoint(t *testing.T) {
	router := setupRouter()
	
	deleteReq := DeleteRequest{
		Name: "test-model",
	}
	jsonData, _ := json.Marshal(deleteReq)
	
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("DELETE", "/api/delete", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, 501, w.Code)
}

func TestCopyEndpoint(t *testing.T) {
	router := setupRouter()
	
	copyReq := CopyRequest{
		Source:      "source-model",
		Destination: "dest-model",
	}
	jsonData, _ := json.Marshal(copyReq)
	
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/copy", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, 501, w.Code)
}

func TestEmbeddingsEndpoint(t *testing.T) {
	router := setupRouter()
	
	embReq := EmbeddingsRequest{
		Model:  "amazon-q",
		Prompt: "Hello world",
	}
	jsonData, _ := json.Marshal(embReq)
	
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/embeddings", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, 501, w.Code)
}

func TestEmbedEndpoint(t *testing.T) {
	router := setupRouter()
	
	embReq := EmbeddingsRequest{
		Model:  "amazon-q",
		Prompt: "Hello world",
	}
	jsonData, _ := json.Marshal(embReq)
	
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/embed", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, 501, w.Code)
}

func TestBlobsGetEndpoint(t *testing.T) {
	router := setupRouter()
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/blobs/sha256:test", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, 404, w.Code)
}

func TestBlobsHeadEndpoint(t *testing.T) {
	router := setupRouter()
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("HEAD", "/api/blobs/sha256:test", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, 404, w.Code)
}

func TestBlobsPostEndpoint(t *testing.T) {
	router := setupRouter()
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/blobs/sha256:test", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, 501, w.Code)
}

func TestUploadEndpoint(t *testing.T) {
	router := setupRouter()
	
	// Create a temporary file for testing
	content := "Hello, World!"
	
	// Create multipart form
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)
	part, err := writer.CreateFormFile("file", "test.txt")
	assert.NoError(t, err)
	
	_, err = io.WriteString(part, content)
	assert.NoError(t, err)
	
	err = writer.Close()
	assert.NoError(t, err)
	
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/upload", &buf)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
	
	var response map[string]string
	err = json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "File uploaded successfully", response["message"])
	assert.Equal(t, "test.txt", response["filename"])
	assert.Contains(t, response["path"], "test.txt")
	
	// Clean up the uploaded file
	if path, exists := response["path"]; exists {
		os.Remove(path)
	}
}

func TestUploadEndpointNoFile(t *testing.T) {
	router := setupRouter()
	
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/upload", nil)
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, 400, w.Code)
	
	var response map[string]string
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "No file uploaded", response["error"])
}

func TestCORSHeaders(t *testing.T) {
	router := setupRouter()
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("OPTIONS", "/api/generate", nil)
	req.Header.Set("Origin", "http://localhost:3000")
	router.ServeHTTP(w, req)

	assert.Equal(t, 204, w.Code)
	assert.Equal(t, "*", w.Header().Get("Access-Control-Allow-Origin"))
	assert.Contains(t, w.Header().Get("Access-Control-Allow-Methods"), "POST")
	assert.Contains(t, w.Header().Get("Access-Control-Allow-Headers"), "Content-Type")
}

func TestInvalidJSONRequest(t *testing.T) {
	router := setupRouter()
	
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/generate", strings.NewReader("invalid json"))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, 400, w.Code)
}

func TestGenerateWithImages(t *testing.T) {
	router := setupRouter()
	
	// Simple base64 encoded 1x1 pixel PNG
	base64Image := "iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mP8/5+hHgAHggJ/PchI7wAAAABJRU5ErkJggg=="
	
	genReq := GenerateRequest{
		Model:  "amazon-q",
		Prompt: "Describe this image",
		Images: []string{base64Image},
		Stream: false,
	}
	jsonData, _ := json.Marshal(genReq)
	
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/generate", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	// Since we don't have actual Q CLI, expect error or success
	assert.True(t, w.Code == 500 || w.Code == 200)
}

func TestChatWithImages(t *testing.T) {
	router := setupRouter()
	
	// Simple base64 encoded 1x1 pixel PNG
	base64Image := "iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mP8/5+hHgAHggJ/PchI7wAAAABJRU5ErkJggg=="
	
	chatReq := ChatRequest{
		Model: "amazon-q",
		Messages: []Message{
			{
				Role:    "user",
				Content: "What do you see?",
				Images:  []string{base64Image},
			},
		},
		Stream: false,
	}
	jsonData, _ := json.Marshal(chatReq)
	
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/chat", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	// Since we don't have actual Q CLI, expect error or success
	assert.True(t, w.Code == 500 || w.Code == 200)
}

func TestStreamingGenerate(t *testing.T) {
	router := setupRouter()
	
	genReq := GenerateRequest{
		Model:  "amazon-q",
		Prompt: "Count to 3",
		Stream: true,
	}
	jsonData, _ := json.Marshal(genReq)
	
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/generate", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	// For streaming, we expect either success or error
	assert.True(t, w.Code == 500 || w.Code == 200)
	
	if w.Code == 200 {
		assert.Equal(t, "application/x-ndjson", w.Header().Get("Content-Type"))
	}
}

func TestStreamingChat(t *testing.T) {
	router := setupRouter()
	
	chatReq := ChatRequest{
		Model: "amazon-q",
		Messages: []Message{
			{
				Role:    "user",
				Content: "Count to 3",
			},
		},
		Stream: true,
	}
	jsonData, _ := json.Marshal(chatReq)
	
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/chat", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	// For streaming, we expect either success or error
	assert.True(t, w.Code == 500 || w.Code == 200)
	
	if w.Code == 200 {
		assert.Equal(t, "application/x-ndjson", w.Header().Get("Content-Type"))
	}
}

// Test helper functions
func TestExecuteQCommandWithInvalidImages(t *testing.T) {
	// Test with invalid base64 data
	invalidImages := []string{"invalid-base64-data"}
	
	// This should not panic and should handle invalid images gracefully
	_, err := executeQCommand("test prompt", invalidImages)
	
	// We expect an error since Q CLI is not available in test environment
	assert.Error(t, err)
}

func TestCorsMiddleware(t *testing.T) {
	router := setupRouter()
	
	// Test preflight request
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("OPTIONS", "/api/chat", nil)
	req.Header.Set("Origin", "http://example.com")
	req.Header.Set("Access-Control-Request-Method", "POST")
	req.Header.Set("Access-Control-Request-Headers", "Content-Type")
	
	router.ServeHTTP(w, req)
	
	assert.Equal(t, 204, w.Code)
	assert.Equal(t, "*", w.Header().Get("Access-Control-Allow-Origin"))
	assert.Contains(t, w.Header().Get("Access-Control-Allow-Methods"), "POST")
	assert.Contains(t, w.Header().Get("Access-Control-Allow-Headers"), "Content-Type")
}

func TestResponseTiming(t *testing.T) {
	router := setupRouter()
	
	start := time.Now()
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/health", nil)
	router.ServeHTTP(w, req)
	duration := time.Since(start)
	
	assert.Equal(t, 200, w.Code)
	assert.Less(t, duration, 100*time.Millisecond, "Health endpoint should respond quickly")
}
