package main

import (
	"bufio"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// OLLAMA API request/response structures
type GenerateRequest struct {
	Model    string                 `json:"model"`
	Prompt   string                 `json:"prompt"`
	Images   []string               `json:"images,omitempty"`
	Format   string                 `json:"format,omitempty"`
	Options  map[string]interface{} `json:"options,omitempty"`
	System   string                 `json:"system,omitempty"`
	Template string                 `json:"template,omitempty"`
	Context  []int                  `json:"context,omitempty"`
	Stream   bool                   `json:"stream,omitempty"`
	Raw      bool                   `json:"raw,omitempty"`
}

type ChatRequest struct {
	Model    string                 `json:"model"`
	Messages []Message              `json:"messages"`
	Format   string                 `json:"format,omitempty"`
	Options  map[string]interface{} `json:"options,omitempty"`
	Stream   bool                   `json:"stream,omitempty"`
	Tools    []Tool                 `json:"tools,omitempty"`
}

type Message struct {
	Role     string   `json:"role"`
	Content  string   `json:"content"`
	Images   []string `json:"images,omitempty"`
	ToolCall *ToolCall `json:"tool_calls,omitempty"`
}

type Tool struct {
	Type     string   `json:"type"`
	Function Function `json:"function"`
}

type Function struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Parameters  map[string]interface{} `json:"parameters"`
}

type ToolCall struct {
	Function FunctionCall `json:"function"`
}

type FunctionCall struct {
	Name      string `json:"name"`
	Arguments string `json:"arguments"`
}

type GenerateResponse struct {
	Model              string    `json:"model"`
	Response           string    `json:"response"`
	Done               bool      `json:"done"`
	Context            []int     `json:"context,omitempty"`
	TotalDuration      int64     `json:"total_duration,omitempty"`
	LoadDuration       int64     `json:"load_duration,omitempty"`
	PromptEvalCount    int       `json:"prompt_eval_count,omitempty"`
	PromptEvalDuration int64     `json:"prompt_eval_duration,omitempty"`
	EvalCount          int       `json:"eval_count,omitempty"`
	EvalDuration       int64     `json:"eval_duration,omitempty"`
	CreatedAt          time.Time `json:"created_at"`
}

type ChatResponse struct {
	Model              string    `json:"model"`
	Message            Message   `json:"message"`
	Done               bool      `json:"done"`
	TotalDuration      int64     `json:"total_duration,omitempty"`
	LoadDuration       int64     `json:"load_duration,omitempty"`
	PromptEvalCount    int       `json:"prompt_eval_count,omitempty"`
	PromptEvalDuration int64     `json:"prompt_eval_duration,omitempty"`
	EvalCount          int       `json:"eval_count,omitempty"`
	EvalDuration       int64     `json:"eval_duration,omitempty"`
	CreatedAt          time.Time `json:"created_at"`
}

type TagsResponse struct {
	Models []ModelInfo `json:"models"`
}

type ModelInfo struct {
	Name       string            `json:"name"`
	Model      string            `json:"model"`
	ModifiedAt time.Time         `json:"modified_at"`
	Size       int64             `json:"size"`
	Digest     string            `json:"digest"`
	Details    ModelDetails      `json:"details"`
	ExpiresAt  *time.Time        `json:"expires_at,omitempty"`
	SizeVram   int64             `json:"size_vram,omitempty"`
}

type ModelDetails struct {
	ParentModel       string   `json:"parent_model,omitempty"`
	Format            string   `json:"format"`
	Family            string   `json:"family"`
	Families          []string `json:"families,omitempty"`
	ParameterSize     string   `json:"parameter_size"`
	QuantizationLevel string   `json:"quantization_level"`
}

type CreateRequest struct {
	Name      string `json:"name"`
	Modelfile string `json:"modelfile,omitempty"`
	Stream    bool   `json:"stream,omitempty"`
	Path      string `json:"path,omitempty"`
}

type PullRequest struct {
	Name     string `json:"name"`
	Insecure bool   `json:"insecure,omitempty"`
	Stream   bool   `json:"stream,omitempty"`
}

type PushRequest struct {
	Name     string `json:"name"`
	Insecure bool   `json:"insecure,omitempty"`
	Stream   bool   `json:"stream,omitempty"`
}

type CopyRequest struct {
	Source      string `json:"source"`
	Destination string `json:"destination"`
}

type DeleteRequest struct {
	Name string `json:"name"`
}

type ShowRequest struct {
	Name    string `json:"name"`
	Verbose bool   `json:"verbose,omitempty"`
}

type ShowResponse struct {
	License    string       `json:"license,omitempty"`
	Modelfile  string       `json:"modelfile,omitempty"`
	Parameters string       `json:"parameters,omitempty"`
	Template   string       `json:"template,omitempty"`
	System     string       `json:"system,omitempty"`
	Details    ModelDetails `json:"details"`
	Messages   []Message    `json:"messages,omitempty"`
}

type EmbeddingsRequest struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
}

type EmbeddingsResponse struct {
	Embedding []float64 `json:"embedding"`
}

type BlobsRequest struct {
	Digest string `json:"digest"`
}

// Execute Amazon Q CLI command with optional file attachments
func executeQCommand(prompt string, images []string) (string, error) {
	args := []string{"chat", "--message", prompt}
	
	// Handle image attachments by saving them temporarily and using file paths
	var tempFiles []string
	defer func() {
		// Clean up temporary files
		for _, file := range tempFiles {
			os.Remove(file)
		}
	}()

	for i, imageData := range images {
		// Decode base64 image data
		data, err := base64.StdEncoding.DecodeString(imageData)
		if err != nil {
			continue // Skip invalid images
		}

		// Create temporary file
		tempFile := filepath.Join(os.TempDir(), fmt.Sprintf("q_image_%d_%d.png", time.Now().Unix(), i))
		if err := os.WriteFile(tempFile, data, 0644); err != nil {
			continue // Skip if can't write file
		}
		
		tempFiles = append(tempFiles, tempFile)
		// Add file argument to Q CLI (assuming Q supports file attachments)
		args = append(args, "--file", tempFile)
	}

	cmd := exec.Command("q", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("q command failed: %v, output: %s", err, string(output))
	}
	return strings.TrimSpace(string(output)), nil
}

// Handle /api/generate endpoint
func handleGenerate(c *gin.Context) {
	var req GenerateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.Stream {
		handleStreamingGenerate(c, req)
		return
	}

	startTime := time.Now()
	response, err := executeQCommand(req.Prompt, req.Images)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	duration := time.Since(startTime)

	c.JSON(http.StatusOK, GenerateResponse{
		Model:         "amazon-q",
		Response:      response,
		Done:          true,
		TotalDuration: duration.Nanoseconds(),
		EvalCount:     len(strings.Fields(response)),
		EvalDuration:  duration.Nanoseconds(),
		CreatedAt:     time.Now(),
	})
}

// Handle streaming generate requests
func handleStreamingGenerate(c *gin.Context, req GenerateRequest) {
	c.Header("Content-Type", "application/x-ndjson")
	c.Header("Transfer-Encoding", "chunked")

	args := []string{"chat", "--message", req.Prompt}
	cmd := exec.Command("q", args...)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if err := cmd.Start(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	scanner := bufio.NewScanner(stdout)
	for scanner.Scan() {
		line := scanner.Text()
		if line != "" {
			response := GenerateResponse{
				Model:     "amazon-q",
				Response:  line,
				Done:      false,
				CreatedAt: time.Now(),
			}
			jsonData, _ := json.Marshal(response)
			c.Writer.Write(jsonData)
			c.Writer.Write([]byte("\n"))
			c.Writer.Flush()
		}
	}

	// Send final response
	finalResponse := GenerateResponse{
		Model:     "amazon-q",
		Response:  "",
		Done:      true,
		CreatedAt: time.Now(),
	}
	jsonData, _ := json.Marshal(finalResponse)
	c.Writer.Write(jsonData)
	c.Writer.Write([]byte("\n"))
	c.Writer.Flush()

	cmd.Wait()
}

// Handle /api/chat endpoint
func handleChat(c *gin.Context) {
	var req ChatRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Extract the last user message and any images
	var userMessage string
	var images []string
	for i := len(req.Messages) - 1; i >= 0; i-- {
		if req.Messages[i].Role == "user" {
			userMessage = req.Messages[i].Content
			images = req.Messages[i].Images
			break
		}
	}

	if userMessage == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No user message found"})
		return
	}

	startTime := time.Now()
	response, err := executeQCommand(userMessage, images)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	duration := time.Since(startTime)

	c.JSON(http.StatusOK, ChatResponse{
		Model: "amazon-q",
		Message: Message{
			Role:    "assistant",
			Content: response,
		},
		Done:          true,
		TotalDuration: duration.Nanoseconds(),
		EvalCount:     len(strings.Fields(response)),
		EvalDuration:  duration.Nanoseconds(),
		CreatedAt:     time.Now(),
	})
}

// Handle /api/tags endpoint
func handleTags(c *gin.Context) {
	c.JSON(http.StatusOK, TagsResponse{
		Models: []ModelInfo{
			{
				Name:       "amazon-q:latest",
				Model:      "amazon-q",
				ModifiedAt: time.Now(),
				Size:       0, // Amazon Q is a service, not a local model
				Digest:     "sha256:amazon-q-service",
				Details: ModelDetails{
					Format:            "amazon-q-service",
					Family:            "amazon-q",
					ParameterSize:     "unknown",
					QuantizationLevel: "unknown",
				},
			},
		},
	})
}

// Handle /api/create endpoint
func handleCreate(c *gin.Context) {
	var req CreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusNotImplemented, gin.H{
		"error": "Model creation not supported for Amazon Q service",
	})
}

// Handle /api/pull endpoint
func handlePull(c *gin.Context) {
	var req PullRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusNotImplemented, gin.H{
		"error": "Model pulling not supported for Amazon Q service",
	})
}

// Handle /api/push endpoint
func handlePush(c *gin.Context) {
	var req PushRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusNotImplemented, gin.H{
		"error": "Model pushing not supported for Amazon Q service",
	})
}

// Handle /api/delete endpoint
func handleDelete(c *gin.Context) {
	var req DeleteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusNotImplemented, gin.H{
		"error": "Model deletion not supported for Amazon Q service",
	})
}

// Handle /api/copy endpoint
func handleCopy(c *gin.Context) {
	var req CopyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusNotImplemented, gin.H{
		"error": "Model copying not supported for Amazon Q service",
	})
}

// Handle /api/show endpoint
func handleShow(c *gin.Context) {
	var req ShowRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, ShowResponse{
		Modelfile: "# Amazon Q Service Model\nFROM amazon-q-service",
		Template:  "{{ .Prompt }}",
		Details: ModelDetails{
			Format:            "amazon-q-service",
			Family:            "amazon-q",
			ParameterSize:     "unknown",
			QuantizationLevel: "unknown",
		},
	})
}

// Handle /api/embeddings endpoint
func handleEmbeddings(c *gin.Context) {
	var req EmbeddingsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusNotImplemented, gin.H{
		"error": "Embeddings not supported for Amazon Q service",
	})
}

// Handle /api/blobs/:digest endpoint
func handleBlobs(c *gin.Context) {
	digest := c.Param("digest")
	if digest == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing digest parameter"})
		return
	}

	c.JSON(http.StatusNotFound, gin.H{
		"error": "Blob storage not supported for Amazon Q service",
	})
}

// Handle HEAD /api/blobs/:digest endpoint
func handleBlobsHead(c *gin.Context) {
	digest := c.Param("digest")
	if digest == "" {
		c.Status(http.StatusBadRequest)
		return
	}

	c.Status(http.StatusNotFound)
}

// Handle POST /api/blobs/:digest endpoint
func handleBlobsPost(c *gin.Context) {
	digest := c.Param("digest")
	if digest == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing digest parameter"})
		return
	}

	c.JSON(http.StatusNotImplemented, gin.H{
		"error": "Blob upload not supported for Amazon Q service",
	})
}

// Handle file upload endpoint
func handleUpload(c *gin.Context) {
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No file uploaded"})
		return
	}
	defer file.Close()

	// Create temporary file
	tempFile := filepath.Join(os.TempDir(), fmt.Sprintf("q_upload_%d_%s", time.Now().Unix(), header.Filename))
	out, err := os.Create(tempFile)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create temporary file"})
		return
	}
	defer out.Close()

	// Copy uploaded file to temporary location
	_, err = io.Copy(out, file)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save file"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":  "File uploaded successfully",
		"filename": header.Filename,
		"path":     tempFile,
	})
}

// Additional OLLAMA endpoint structures
type ProcessRequest struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
}

type ProcessResponse struct {
	Response string `json:"response"`
}

type PsResponse struct {
	Models []RunningModel `json:"models"`
}

type RunningModel struct {
	Name       string    `json:"name"`
	Model      string    `json:"model"`
	Size       int64     `json:"size"`
	Digest     string    `json:"digest"`
	Details    ModelDetails `json:"details"`
	ExpiresAt  time.Time `json:"expires_at"`
	SizeVram   int64     `json:"size_vram"`
}

type StatusResponse struct {
	Status string `json:"status"`
	Models []RunningModel `json:"models,omitempty"`
}

// Handle /api/ps endpoint - List running models
func handlePs(c *gin.Context) {
	c.JSON(http.StatusOK, PsResponse{
		Models: []RunningModel{
			{
				Name:      "amazon-q:latest",
				Model:     "amazon-q",
				Size:      0,
				Digest:    "sha256:amazon-q-service",
				ExpiresAt: time.Now().Add(24 * time.Hour),
				SizeVram:  0,
				Details: ModelDetails{
					Format:            "amazon-q-service",
					Family:            "amazon-q",
					ParameterSize:     "unknown",
					QuantizationLevel: "unknown",
				},
			},
		},
	})
}

// Handle /api/embed endpoint - Generate embeddings (alternative to /api/embeddings)
func handleEmbed(c *gin.Context) {
	var req EmbeddingsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusNotImplemented, gin.H{
		"error": "Embeddings not supported for Amazon Q service",
	})
}

// Handle /api/list endpoint - Alternative to /api/tags
func handleList(c *gin.Context) {
	handleTags(c)
}

// Handle /api/status endpoint - Server status
func handleStatus(c *gin.Context) {
	c.JSON(http.StatusOK, StatusResponse{
		Status: "running",
		Models: []RunningModel{
			{
				Name:      "amazon-q:latest",
				Model:     "amazon-q",
				Size:      0,
				Digest:    "sha256:amazon-q-service",
				ExpiresAt: time.Now().Add(24 * time.Hour),
				SizeVram:  0,
				Details: ModelDetails{
					Format:            "amazon-q-service",
					Family:            "amazon-q",
					ParameterSize:     "unknown",
					QuantizationLevel: "unknown",
				},
			},
		},
	})
}

// Handle streaming chat endpoint
func handleChatStream(c *gin.Context, req ChatRequest) {
	c.Header("Content-Type", "application/x-ndjson")
	c.Header("Transfer-Encoding", "chunked")

	// Extract the last user message and any images
	var userMessage string
	for i := len(req.Messages) - 1; i >= 0; i-- {
		if req.Messages[i].Role == "user" {
			userMessage = req.Messages[i].Content
			break
		}
	}

	if userMessage == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No user message found"})
		return
	}

	args := []string{"chat", "--message", userMessage}
	cmd := exec.Command("q", args...)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if err := cmd.Start(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	scanner := bufio.NewScanner(stdout)
	for scanner.Scan() {
		line := scanner.Text()
		if line != "" {
			response := ChatResponse{
				Model: "amazon-q",
				Message: Message{
					Role:    "assistant",
					Content: line,
				},
				Done:      false,
				CreatedAt: time.Now(),
			}
			jsonData, _ := json.Marshal(response)
			c.Writer.Write(jsonData)
			c.Writer.Write([]byte("\n"))
			c.Writer.Flush()
		}
	}

	// Send final response
	finalResponse := ChatResponse{
		Model: "amazon-q",
		Message: Message{
			Role:    "assistant",
			Content: "",
		},
		Done:      true,
		CreatedAt: time.Now(),
	}
	jsonData, _ := json.Marshal(finalResponse)
	c.Writer.Write(jsonData)
	c.Writer.Write([]byte("\n"))
	c.Writer.Flush()

	cmd.Wait()
}

// Update handleChat to support streaming
func handleChatWithStreaming(c *gin.Context) {
	var req ChatRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.Stream {
		handleChatStream(c, req)
		return
	}

	// Extract the last user message and any images
	var userMessage string
	var images []string
	for i := len(req.Messages) - 1; i >= 0; i-- {
		if req.Messages[i].Role == "user" {
			userMessage = req.Messages[i].Content
			images = req.Messages[i].Images
			break
		}
	}

	if userMessage == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No user message found"})
		return
	}

	startTime := time.Now()
	response, err := executeQCommand(userMessage, images)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	duration := time.Since(startTime)

	c.JSON(http.StatusOK, ChatResponse{
		Model: "amazon-q",
		Message: Message{
			Role:    "assistant",
			Content: response,
		},
		Done:          true,
		TotalDuration: duration.Nanoseconds(),
		EvalCount:     len(strings.Fields(response)),
		EvalDuration:  duration.Nanoseconds(),
		CreatedAt:     time.Now(),
	})
}

// CORS middleware for browser compatibility
func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS, HEAD")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
		c.Header("Access-Control-Expose-Headers", "Content-Length")
		c.Header("Access-Control-Allow-Credentials", "true")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}
