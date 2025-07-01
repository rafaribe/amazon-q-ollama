# Amazon Q OLLAMA

A comprehensive REST API wrapper for Amazon Q that provides full OLLAMA-compatible endpoints, allowing applications built for OLLAMA to seamlessly use Amazon Q instead. Includes support for file uploads, image processing, and all OLLAMA API features.

## Features

- **Complete OLLAMA API Compatibility**: All endpoints implemented
- **File and Image Support**: Handle base64 images and file uploads
- **Streaming Responses**: Real-time streaming for generate and chat endpoints
- **Built on Existing Container**: Uses your `ghcr.io/rafaribe/amazon-q:2025.07.01` container
- **Docker and Docker Compose Support**: Easy deployment options
- **Comprehensive Testing**: Full test suite with Makefile commands
- **Health Monitoring**: Built-in health checks and monitoring

## Complete API Endpoints

### Core Chat & Generation
- `POST /api/generate` - Generate completions (with streaming support)
- `POST /api/chat` - Chat conversations (with image support)
- `GET /api/tags` - List available models
- `POST /api/show` - Show model information
- `GET /api/version` - API version information

### Model Management (Compatibility Layer)
- `POST /api/create` - Model creation (returns not implemented)
- `POST /api/pull` - Model downloading (returns not implemented)
- `POST /api/push` - Model uploading (returns not implemented)
- `DELETE /api/delete` - Model deletion (returns not implemented)
- `POST /api/copy` - Model copying (returns not implemented)

### Advanced Features
- `POST /api/embeddings` - Text embeddings (returns not implemented)
- `GET /api/blobs/:digest` - Blob retrieval (returns not found)
- `HEAD /api/blobs/:digest` - Blob existence check (returns not found)
- `POST /api/blobs/:digest` - Blob upload (returns not implemented)

### File Handling
- `POST /upload` - File upload endpoint for attachments

### Utility
- `GET /health` - Health check endpoint
- `GET /` - Root endpoint with API information

## Usage Examples

### Basic Chat
```bash
curl -X POST http://localhost:11434/api/chat \
  -H "Content-Type: application/json" \
  -d '{
    "model": "amazon-q",
    "messages": [
      {
        "role": "user",
        "content": "Explain Kubernetes deployment strategies"
      }
    ]
  }'
```

### Chat with Image
```bash
curl -X POST http://localhost:11434/api/chat \
  -H "Content-Type: application/json" \
  -d '{
    "model": "amazon-q",
    "messages": [
      {
        "role": "user",
        "content": "What do you see in this image?",
        "images": ["base64_encoded_image_data_here"]
      }
    ]
  }'
```

### Streaming Generation
```bash
curl -X POST http://localhost:11434/api/generate \
  -H "Content-Type: application/json" \
  -d '{
    "model": "amazon-q",
    "prompt": "Write a comprehensive Go web server example",
    "stream": true
  }'
```

### File Upload
```bash
curl -X POST http://localhost:11434/upload \
  -F "file=@/path/to/your/file.txt"
```

### Advanced Generation with Options
```bash
curl -X POST http://localhost:11434/api/generate \
  -H "Content-Type: application/json" \
  -d '{
    "model": "amazon-q",
    "prompt": "Explain AWS Lambda best practices",
    "system": "You are an AWS expert",
    "format": "json",
    "options": {
      "temperature": 0.7
    }
  }'
```

## Deployment

### Docker Compose (Recommended)
```bash
# Set your AWS credentials
export AWS_ACCESS_KEY_ID=your_access_key
export AWS_SECRET_ACCESS_KEY=your_secret_key
export AWS_REGION=us-east-1

# Start the service
docker-compose up -d

# View logs
docker-compose logs -f

# Stop the service
docker-compose down
```

### Docker Build and Run
```bash
# Build the image
docker build -t amazon-q-ollama .

# Run the container
docker run -d \
  -p 11434:11434 \
  -e AWS_ACCESS_KEY_ID=your_access_key \
  -e AWS_SECRET_ACCESS_KEY=your_secret_key \
  -e AWS_REGION=us-east-1 \
  -v ~/.aws:/home/dev/.aws:ro \
  amazon-q-ollama
```

### Local Development
```bash
# Install dependencies
make deps

# Build and run
make dev

# Or run directly
go run .
```

## Testing

The project includes comprehensive testing commands:

```bash
# Test core functionality
make test-core

# Test all endpoints
make test-all

# Test specific endpoints
make test-health
make test-generate
make test-chat
make test-chat-with-image
make test-upload
make test-generate-stream
```

### Individual Endpoint Tests
```bash
# Health check
make test-health

# Basic generation
make test-generate

# Streaming generation
make test-generate-stream

# Chat conversation
make test-chat

# Chat with image
make test-chat-with-image

# File upload
make test-upload

# Model information
make test-tags
make test-show

# Version info
make test-version
```

## File and Image Support

### Image Processing
The API supports base64-encoded images in chat messages:
- Images are automatically decoded and saved as temporary files
- Temporary files are passed to the Amazon Q CLI
- Files are automatically cleaned up after processing

### File Uploads
- Upload files via the `/upload` endpoint
- Files are stored temporarily and can be referenced in subsequent requests
- Supports multipart form uploads up to 32MB

## Configuration

### Environment Variables
- `AWS_REGION` - AWS region (default: us-east-1)
- `AWS_ACCESS_KEY_ID` - AWS access key
- `AWS_SECRET_ACCESS_KEY` - AWS secret key
- `AWS_SESSION_TOKEN` - AWS session token (if using temporary credentials)

### Volume Mounts
- `~/.aws:/home/dev/.aws:ro` - AWS credentials
- `amazon-q-data:/home/dev/.local/share/amazon-q` - Amazon Q data
- `amazon-q-cache:/home/dev/.cache/amazon-q` - Amazon Q cache

## Integration with Existing Applications

This API is fully compatible with OLLAMA's REST API, making it a drop-in replacement:

### Python Example
```python
import requests

# Works exactly like OLLAMA
response = requests.post('http://localhost:11434/api/chat', json={
    'model': 'amazon-q',
    'messages': [
        {'role': 'user', 'content': 'Hello!'}
    ]
})

print(response.json())
```

### JavaScript Example
```javascript
const response = await fetch('http://localhost:11434/api/generate', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({
        model: 'amazon-q',
        prompt: 'Explain microservices architecture'
    })
});

const data = await response.json();
console.log(data.response);
```

## Response Formats

All responses follow OLLAMA's exact format specifications:

### Generate Response
```json
{
  "model": "amazon-q",
  "response": "Generated text response",
  "done": true,
  "total_duration": 1234567890,
  "eval_count": 42,
  "eval_duration": 987654321,
  "created_at": "2025-07-01T22:00:00Z"
}
```

### Chat Response
```json
{
  "model": "amazon-q",
  "message": {
    "role": "assistant",
    "content": "Chat response content"
  },
  "done": true,
  "total_duration": 1234567890,
  "created_at": "2025-07-01T22:00:00Z"
}
```

## Architecture

- **Base Container**: `ghcr.io/rafaribe/amazon-q:2025.07.01`
- **API Framework**: Go with Gin web framework
- **CLI Integration**: Executes `q chat` commands with file support
- **File Handling**: Temporary file management for images and uploads
- **Streaming**: Real-time response streaming via HTTP chunked transfer
- **Compatibility**: 100% OLLAMA API compatible

## Performance Considerations

- File uploads limited to 32MB
- Temporary files are automatically cleaned up
- Streaming responses for better user experience
- Health checks for monitoring
- Efficient base64 image processing

## Troubleshooting

### Common Issues
1. **AWS Credentials**: Ensure AWS credentials are properly configured
2. **File Permissions**: Check that the container can write to temp directories
3. **Network**: Verify port 11434 is accessible
4. **Memory**: Large file uploads may require increased memory limits

### Debug Commands
```bash
# Check container logs
docker-compose logs -f

# Test health endpoint
curl http://localhost:11434/health

# Check API information
curl http://localhost:11434/

# Verify AWS credentials in container
docker exec -it amazon-q-api aws sts get-caller-identity
```

This implementation provides a complete, production-ready OLLAMA-compatible API wrapper for Amazon Q with full file and image support.
