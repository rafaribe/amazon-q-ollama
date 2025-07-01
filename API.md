# Amazon Q OLLAMA - Complete OLLAMA Compatibility

This document provides comprehensive documentation for all available endpoints in the Amazon Q OLLAMA wrapper.

## Base URL
```
http://localhost:11434
```

## Authentication
Uses AWS credentials configured in the container environment.

## Complete Endpoint Reference

### Core Chat & Generation Endpoints

#### POST /api/generate
Generate text completions with optional streaming.

**Request Body:**
```json
{
  "model": "amazon-q",
  "prompt": "Your prompt here",
  "images": ["base64_encoded_image_data"],
  "format": "json",
  "options": {
    "temperature": 0.7
  },
  "system": "System prompt",
  "template": "Template string",
  "context": [1, 2, 3],
  "stream": false,
  "raw": false
}
```

**Response:**
```json
{
  "model": "amazon-q",
  "response": "Generated response",
  "done": true,
  "context": [1, 2, 3],
  "total_duration": 1234567890,
  "load_duration": 123456,
  "prompt_eval_count": 10,
  "prompt_eval_duration": 987654,
  "eval_count": 25,
  "eval_duration": 1234567,
  "created_at": "2025-07-01T22:00:00Z"
}
```

#### POST /api/chat
Chat conversations with message history and image support.

**Request Body:**
```json
{
  "model": "amazon-q",
  "messages": [
    {
      "role": "user",
      "content": "Hello!",
      "images": ["base64_encoded_image_data"]
    }
  ],
  "format": "json",
  "options": {
    "temperature": 0.7
  },
  "stream": false,
  "tools": [
    {
      "type": "function",
      "function": {
        "name": "get_weather",
        "description": "Get weather information",
        "parameters": {
          "type": "object",
          "properties": {
            "location": {"type": "string"}
          }
        }
      }
    }
  ]
}
```

**Response:**
```json
{
  "model": "amazon-q",
  "message": {
    "role": "assistant",
    "content": "Hello! How can I help you?",
    "images": [],
    "tool_calls": null
  },
  "done": true,
  "total_duration": 1234567890,
  "created_at": "2025-07-01T22:00:00Z"
}
```

### Model Information Endpoints

#### GET /api/tags
List available models.

**Response:**
```json
{
  "models": [
    {
      "name": "amazon-q:latest",
      "model": "amazon-q",
      "modified_at": "2025-07-01T22:00:00Z",
      "size": 0,
      "digest": "sha256:amazon-q-service",
      "details": {
        "parent_model": "",
        "format": "amazon-q-service",
        "family": "amazon-q",
        "families": ["amazon-q"],
        "parameter_size": "unknown",
        "quantization_level": "unknown"
      },
      "expires_at": null,
      "size_vram": 0
    }
  ]
}
```

#### GET /api/list
Alternative endpoint for listing models (same as /api/tags).

#### POST /api/show
Show detailed model information.

**Request Body:**
```json
{
  "name": "amazon-q",
  "verbose": false
}
```

**Response:**
```json
{
  "license": "",
  "modelfile": "# Amazon Q Service Model\nFROM amazon-q-service",
  "parameters": "",
  "template": "{{ .Prompt }}",
  "system": "",
  "details": {
    "format": "amazon-q-service",
    "family": "amazon-q",
    "parameter_size": "unknown",
    "quantization_level": "unknown"
  },
  "messages": []
}
```

### Process Management Endpoints

#### GET /api/ps
List running models and their status.

**Response:**
```json
{
  "models": [
    {
      "name": "amazon-q:latest",
      "model": "amazon-q",
      "size": 0,
      "digest": "sha256:amazon-q-service",
      "details": {
        "format": "amazon-q-service",
        "family": "amazon-q",
        "parameter_size": "unknown",
        "quantization_level": "unknown"
      },
      "expires_at": "2025-07-02T22:00:00Z",
      "size_vram": 0
    }
  ]
}
```

#### GET /api/status
Server status and running models.

**Response:**
```json
{
  "status": "running",
  "models": [
    {
      "name": "amazon-q:latest",
      "model": "amazon-q",
      "size": 0,
      "digest": "sha256:amazon-q-service",
      "expires_at": "2025-07-02T22:00:00Z",
      "size_vram": 0
    }
  ]
}
```

### Model Management Endpoints (Compatibility Layer)

#### POST /api/create
Create a new model (returns not implemented for Amazon Q).

**Request Body:**
```json
{
  "name": "my-model",
  "modelfile": "FROM amazon-q\nSYSTEM You are a helpful assistant",
  "stream": false,
  "path": "/path/to/modelfile"
}
```

#### POST /api/pull
Pull a model from registry (returns not implemented).

**Request Body:**
```json
{
  "name": "model-name",
  "insecure": false,
  "stream": false
}
```

#### POST /api/push
Push a model to registry (returns not implemented).

**Request Body:**
```json
{
  "name": "model-name",
  "insecure": false,
  "stream": false
}
```

#### DELETE /api/delete
Delete a model (returns not implemented).

**Request Body:**
```json
{
  "name": "model-name"
}
```

#### POST /api/copy
Copy a model (returns not implemented).

**Request Body:**
```json
{
  "source": "source-model",
  "destination": "destination-model"
}
```

### Embedding Endpoints

#### POST /api/embeddings
Generate text embeddings (returns not implemented).

**Request Body:**
```json
{
  "model": "amazon-q",
  "prompt": "Text to embed"
}
```

#### POST /api/embed
Alternative embedding endpoint (returns not implemented).

### Blob Storage Endpoints

#### GET /api/blobs/:digest
Retrieve a blob by digest (returns not found).

#### HEAD /api/blobs/:digest
Check if a blob exists (returns not found).

#### POST /api/blobs/:digest
Upload a blob (returns not implemented).

### File Handling Endpoints

#### POST /upload
Upload files for processing.

**Request:** Multipart form data with file field.

**Response:**
```json
{
  "message": "File uploaded successfully",
  "filename": "example.txt",
  "path": "/tmp/q_upload_1234567890_example.txt"
}
```

### Utility Endpoints

#### GET /health
Health check endpoint.

**Response:**
```json
{
  "status": "ok"
}
```

#### GET /ping
Simple ping endpoint.

**Response:** `pong` (text/plain)

#### HEAD /
Alternative health check (returns 200 OK).

#### GET /api/version
API version information.

**Response:**
```json
{
  "version": "amazon-q-ollama-1.0.0"
}
```

#### GET /metrics
Basic metrics endpoint (Prometheus format).

**Response:**
```
# Amazon Q OLLAMA Metrics
amazon_q_ollama_up 1
```

#### GET /
Root endpoint with API information.

**Response:**
```json
{
  "message": "Amazon Q OLLAMA - OLLAMA Compatible API",
  "version": "1.0.0",
  "endpoints": ["...list of all endpoints..."]
}
```

## Streaming Support

Both `/api/generate` and `/api/chat` support streaming responses when `"stream": true` is included in the request body.

**Streaming Response Format:**
- Content-Type: `application/x-ndjson`
- Each line contains a JSON object
- Final response has `"done": true`

**Example Streaming Response:**
```
{"model":"amazon-q","response":"Hello","done":false,"created_at":"2025-07-01T22:00:00Z"}
{"model":"amazon-q","response":" there!","done":false,"created_at":"2025-07-01T22:00:01Z"}
{"model":"amazon-q","response":"","done":true,"created_at":"2025-07-01T22:00:02Z"}
```

## Image Support

Images can be included in chat messages as base64-encoded strings:

```json
{
  "model": "amazon-q",
  "messages": [
    {
      "role": "user",
      "content": "What do you see in this image?",
      "images": ["data:image/jpeg;base64,/9j/4AAQSkZJRgABAQAAAQ..."]
    }
  ]
}
```

## Error Responses

All endpoints return appropriate HTTP status codes and error messages:

```json
{
  "error": "Error description"
}
```

Common status codes:
- `200` - Success
- `400` - Bad Request
- `404` - Not Found
- `500` - Internal Server Error
- `501` - Not Implemented

## CORS Support

The API includes CORS headers for browser compatibility:
- `Access-Control-Allow-Origin: *`
- `Access-Control-Allow-Methods: GET, POST, PUT, DELETE, OPTIONS, HEAD`
- `Access-Control-Allow-Headers: Origin, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization`

## Rate Limiting

No rate limiting is currently implemented. Consider adding rate limiting for production deployments.

## File Upload Limits

- Maximum file size: 32MB
- Supported via multipart form data
- Files are stored temporarily and cleaned up automatically
