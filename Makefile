.PHONY: build run test clean docker-build docker-run docker-compose-up docker-compose-down

# Go commands
build:
	go build -o amazon-q-ollama .

run:
	go run .

test:
	go test -v ./...

test-unit:
	go test -v -run "^Test" ./...

test-integration:
	go test -v -run "^TestIntegration" ./...

test-benchmark:
	go test -v -run "^Benchmark" -bench=. ./...

test-coverage:
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

test-race:
	go test -v -race ./...

test-short:
	go test -v -short ./...

test-verbose:
	go test -v -count=1 ./...

clean:
	rm -f amazon-q-ollama coverage.out coverage.html

# Docker commands
docker-build:
	docker build -t amazon-q-ollama .

docker-run:
	docker run -d \
		-p 11434:11434 \
		-e AWS_REGION=${AWS_REGION} \
		-e AWS_ACCESS_KEY_ID=${AWS_ACCESS_KEY_ID} \
		-e AWS_SECRET_ACCESS_KEY=${AWS_SECRET_ACCESS_KEY} \
		-v ~/.aws:/home/dev/.aws:ro \
		--name amazon-q-ollama \
		amazon-q-ollama

docker-stop:
	docker stop amazon-q-ollama || true
	docker rm amazon-q-ollama || true

# Docker Compose commands
docker-compose-up:
	docker-compose up -d

docker-compose-down:
	docker-compose down

docker-compose-logs:
	docker-compose logs -f

# Development helpers
dev: clean build run

deps:
	go mod tidy
	go mod download

# Live testing endpoints (requires running server)
test-health:
	@echo "Testing health endpoint..."
	curl -f http://localhost:11434/health

test-ping:
	@echo "Testing ping endpoint..."
	curl http://localhost:11434/ping

test-version:
	@echo "Testing version endpoint..."
	curl http://localhost:11434/api/version

test-root:
	@echo "Testing root endpoint..."
	curl http://localhost:11434/

test-generate:
	@echo "Testing generate endpoint..."
	curl -X POST http://localhost:11434/api/generate \
		-H "Content-Type: application/json" \
		-d '{"model": "amazon-q", "prompt": "Hello, how are you?"}'

test-generate-stream:
	@echo "Testing streaming generate endpoint..."
	curl -X POST http://localhost:11434/api/generate \
		-H "Content-Type: application/json" \
		-d '{"model": "amazon-q", "prompt": "Write a simple Go function", "stream": true}'

test-chat:
	@echo "Testing chat endpoint..."
	curl -X POST http://localhost:11434/api/chat \
		-H "Content-Type: application/json" \
		-d '{"model": "amazon-q", "messages": [{"role": "user", "content": "What is Go programming language?"}]}'

test-chat-stream:
	@echo "Testing streaming chat endpoint..."
	curl -X POST http://localhost:11434/api/chat \
		-H "Content-Type: application/json" \
		-d '{"model": "amazon-q", "messages": [{"role": "user", "content": "Count to 5"}], "stream": true}'

test-chat-with-image:
	@echo "Testing chat endpoint with image..."
	curl -X POST http://localhost:11434/api/chat \
		-H "Content-Type: application/json" \
		-d '{"model": "amazon-q", "messages": [{"role": "user", "content": "Describe this image", "images": ["iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mP8/5+hHgAHggJ/PchI7wAAAABJRU5ErkJggg=="]}]}'

test-tags:
	@echo "Testing tags endpoint..."
	curl http://localhost:11434/api/tags

test-list:
	@echo "Testing list endpoint..."
	curl http://localhost:11434/api/list

test-show:
	@echo "Testing show endpoint..."
	curl -X POST http://localhost:11434/api/show \
		-H "Content-Type: application/json" \
		-d '{"name": "amazon-q"}'

test-ps:
	@echo "Testing ps endpoint..."
	curl http://localhost:11434/api/ps

test-status:
	@echo "Testing status endpoint..."
	curl http://localhost:11434/api/status

test-create:
	@echo "Testing create endpoint (should return not implemented)..."
	curl -X POST http://localhost:11434/api/create \
		-H "Content-Type: application/json" \
		-d '{"name": "test-model", "modelfile": "FROM amazon-q"}'

test-pull:
	@echo "Testing pull endpoint (should return not implemented)..."
	curl -X POST http://localhost:11434/api/pull \
		-H "Content-Type: application/json" \
		-d '{"name": "test-model"}'

test-push:
	@echo "Testing push endpoint (should return not implemented)..."
	curl -X POST http://localhost:11434/api/push \
		-H "Content-Type: application/json" \
		-d '{"name": "test-model"}'

test-delete:
	@echo "Testing delete endpoint (should return not implemented)..."
	curl -X DELETE http://localhost:11434/api/delete \
		-H "Content-Type: application/json" \
		-d '{"name": "test-model"}'

test-copy:
	@echo "Testing copy endpoint (should return not implemented)..."
	curl -X POST http://localhost:11434/api/copy \
		-H "Content-Type: application/json" \
		-d '{"source": "test-model", "destination": "test-model-copy"}'

test-embeddings:
	@echo "Testing embeddings endpoint (should return not implemented)..."
	curl -X POST http://localhost:11434/api/embeddings \
		-H "Content-Type: application/json" \
		-d '{"model": "amazon-q", "prompt": "Hello world"}'

test-embed:
	@echo "Testing embed endpoint (should return not implemented)..."
	curl -X POST http://localhost:11434/api/embed \
		-H "Content-Type: application/json" \
		-d '{"model": "amazon-q", "prompt": "Hello world"}'

test-blobs:
	@echo "Testing blobs endpoint (should return not found)..."
	curl http://localhost:11434/api/blobs/sha256:test

test-upload:
	@echo "Testing file upload endpoint..."
	echo "Hello World" > /tmp/test.txt
	curl -X POST http://localhost:11434/upload \
		-F "file=@/tmp/test.txt"
	rm -f /tmp/test.txt

test-head-root:
	@echo "Testing HEAD / endpoint..."
	curl -I http://localhost:11434/

test-metrics:
	@echo "Testing metrics endpoint..."
	curl http://localhost:11434/metrics

# Run all live tests (requires running server)
test-all-live: test-health test-ping test-version test-root test-generate test-generate-stream test-chat test-chat-stream test-chat-with-image test-tags test-list test-show test-ps test-status test-create test-pull test-push test-delete test-copy test-embeddings test-embed test-blobs test-upload test-head-root test-metrics

# Test core functionality only (requires running server)
test-core-live: test-health test-ping test-generate test-chat test-tags test-show test-ps test-status

# Performance testing
test-performance:
	@echo "Running performance tests..."
	go test -bench=. -benchmem -count=3 ./...

# Load testing (requires hey tool: go install github.com/rakyll/hey@latest)
test-load-health:
	@echo "Load testing health endpoint..."
	hey -n 1000 -c 10 http://localhost:11434/health

test-load-tags:
	@echo "Load testing tags endpoint..."
	hey -n 1000 -c 10 http://localhost:11434/api/tags

# Security testing
test-security:
	@echo "Running security tests..."
	@echo "Testing CORS..."
	curl -H "Origin: http://malicious.com" -H "Access-Control-Request-Method: POST" -H "Access-Control-Request-Headers: X-Requested-With" -X OPTIONS http://localhost:11434/api/generate
	@echo "Testing invalid methods..."
	curl -X PATCH http://localhost:11434/api/generate
	@echo "Testing large payloads..."
	curl -X POST http://localhost:11434/api/generate -H "Content-Type: application/json" -d '{"model": "amazon-q", "prompt": "'$(python3 -c "print('A' * 10000)")'"}' || true

# Comprehensive test suite
test-comprehensive: test test-race test-coverage test-performance

# CI/CD pipeline tests
test-ci: test-unit test-integration test-race test-coverage

# Development workflow
test-dev: test-short test-verbose

# Container testing
test-container:
	@echo "Running comprehensive container tests..."
	./test-container.sh

test-container-compose:
	@echo "Running tests with Docker Compose..."
	docker-compose -f docker-compose.test.yml up --build --abort-on-container-exit --exit-code-from test-runner

test-container-manual:
	@echo "Building and running container for manual testing..."
	docker build -t amazon-q-ollama .
	docker run -d --name amazon-q-ollama-manual -p 11434:11434 amazon-q-ollama
	@echo "Container is running on http://localhost:11434"
	@echo "Run 'make stop-container-manual' to stop and remove the container"

stop-container-manual:
	@echo "Stopping and removing manual test container..."
	docker stop amazon-q-ollama-manual || true
	docker rm amazon-q-ollama-manual || true

test-container-logs:
	@echo "Showing container logs..."
	docker logs amazon-q-ollama-manual

# Container health check
test-container-health:
	@echo "Testing container health..."
	@if docker ps | grep -q amazon-q-ollama-manual; then \
		echo "Container is running"; \
		curl -f http://localhost:11434/health && echo " - Health check passed" || echo " - Health check failed"; \
	else \
		echo "Container is not running"; \
	fi

# Quick container test
test-container-quick:
	@echo "Running quick container test..."
	docker build -t amazon-q-ollama . && \
	docker run --rm -d --name amazon-q-ollama-quick -p 11435:11434 amazon-q-ollama && \
	sleep 5 && \
	curl -f http://localhost:11435/health && \
	echo " - Quick test passed" && \
	docker stop amazon-q-ollama-quick || true

# Help
help:
	@echo "Available commands:"
	@echo ""
	@echo "Build & Run:"
	@echo "  build               - Build the Go binary"
	@echo "  run                 - Run the application locally"
	@echo "  dev                 - Clean, build and run"
	@echo "  deps                - Install Go dependencies"
	@echo ""
	@echo "Go Testing:"
	@echo "  test                - Run all Go tests"
	@echo "  test-unit           - Run unit tests only"
	@echo "  test-integration    - Run integration tests only"
	@echo "  test-benchmark      - Run benchmark tests"
	@echo "  test-coverage       - Run tests with coverage report"
	@echo "  test-race           - Run tests with race detection"
	@echo "  test-performance    - Run performance benchmarks"
	@echo "  test-comprehensive  - Run all test types"
	@echo "  test-ci             - Run CI/CD pipeline tests"
	@echo ""
	@echo "Container Testing:"
	@echo "  test-container      - Run comprehensive container tests"
	@echo "  test-container-compose - Run tests with Docker Compose"
	@echo "  test-container-manual - Start container for manual testing"
	@echo "  test-container-quick - Quick container health test"
	@echo "  test-container-health - Check running container health"
	@echo "  stop-container-manual - Stop manual test container"
	@echo ""
	@echo "Live Testing (requires running server):"
	@echo "  test-all-live       - Test all endpoints on running server"
	@echo "  test-core-live      - Test core endpoints on running server"
	@echo "  test-health         - Test health endpoint"
	@echo "  test-generate       - Test generate endpoint"
	@echo "  test-chat           - Test chat endpoint"
	@echo ""
	@echo "Docker:"
	@echo "  docker-build        - Build Docker image"
	@echo "  docker-run          - Run Docker container"
	@echo "  docker-stop         - Stop Docker container"
	@echo "  docker-compose-up   - Start with Docker Compose"
	@echo "  docker-compose-down - Stop Docker Compose"
