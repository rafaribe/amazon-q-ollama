package main

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()

	// Add CORS middleware for browser compatibility
	r.Use(corsMiddleware())

	// Set max multipart memory for file uploads (32 MB)
	r.MaxMultipartMemory = 32 << 20

	// OLLAMA-compatible API endpoints
	api := r.Group("/api")
	{
		// Core endpoints
		api.POST("/generate", handleGenerate)
		api.POST("/chat", handleChatWithStreaming)
		api.GET("/tags", handleTags)
		api.POST("/show", handleShow)
		
		// Model management endpoints (not supported but implemented for compatibility)
		api.POST("/create", handleCreate)
		api.POST("/pull", handlePull)
		api.POST("/push", handlePush)
		api.DELETE("/delete", handleDelete)
		api.POST("/copy", handleCopy)
		
		// Process management endpoints
		api.GET("/ps", handlePs)
		api.GET("/status", handleStatus)
		
		// Embedding endpoints
		api.POST("/embeddings", handleEmbeddings)
		api.POST("/embed", handleEmbed)
		
		// Alternative endpoints
		api.GET("/list", handleList)
		
		// Blob endpoints
		api.GET("/blobs/:digest", handleBlobs)
		api.HEAD("/blobs/:digest", handleBlobsHead)
		api.POST("/blobs/:digest", handleBlobsPost)
		
		// Version endpoint
		api.GET("/version", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{
				"version": "amazon-q-ollama-1.0.0",
			})
		})
	}

	// File upload endpoint
	r.POST("/upload", handleUpload)

	// Health check endpoint
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// Ping endpoint (OLLAMA compatibility)
	r.GET("/ping", func(c *gin.Context) {
		c.String(http.StatusOK, "pong")
	})

	// Alternative health check
	r.HEAD("/", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	// Metrics endpoint (basic)
	r.GET("/metrics", func(c *gin.Context) {
		c.String(http.StatusOK, "# Amazon Q OLLAMA Metrics\namazon_q_ollama_up 1\n")
	})

	// Root endpoint
	r.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "Amazon Q OLLAMA - OLLAMA Compatible API",
			"version": "1.0.0",
			"endpoints": []string{
				"POST /api/generate",
				"POST /api/chat",
				"GET /api/tags",
				"GET /api/list",
				"POST /api/show",
				"POST /api/create",
				"POST /api/pull",
				"POST /api/push",
				"DELETE /api/delete",
				"POST /api/copy",
				"GET /api/ps",
				"GET /api/status",
				"POST /api/embeddings",
				"POST /api/embed",
				"GET /api/blobs/:digest",
				"HEAD /api/blobs/:digest",
				"POST /api/blobs/:digest",
				"GET /api/version",
				"POST /upload",
				"GET /health",
				"GET /ping",
				"HEAD /",
				"GET /metrics",
			},
		})
	})

	log.Println("Amazon Q OLLAMA server starting on :11434")
	log.Println("OLLAMA-compatible endpoints available")
	log.Println("Complete API compatibility with streaming support")
	log.Println("CORS enabled for browser compatibility")
	log.Println("Visit http://localhost:11434 for endpoint information")
	
	if err := r.Run(":11434"); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
