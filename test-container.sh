#!/bin/bash

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
CONTAINER_NAME="amazon-q-ollama-test"
IMAGE_NAME="amazon-q-ollama"
PORT="11435"  # Changed from 11434 to avoid conflicts
BASE_URL="http://localhost:${PORT}"

# Test counters
TOTAL_TESTS=0
PASSED_TESTS=0
FAILED_TESTS=0

# Function to print colored output
print_status() {
    local status=$1
    local message=$2
    case $status in
        "INFO")
            echo -e "${BLUE}[INFO]${NC} $message"
            ;;
        "SUCCESS")
            echo -e "${GREEN}[SUCCESS]${NC} $message"
            ;;
        "ERROR")
            echo -e "${RED}[ERROR]${NC} $message"
            ;;
        "WARNING")
            echo -e "${YELLOW}[WARNING]${NC} $message"
            ;;
    esac
}

# Function to run a test
run_test() {
    local test_name=$1
    local curl_command=$2
    local expected_status=${3:-200}
    
    TOTAL_TESTS=$((TOTAL_TESTS + 1))
    
    print_status "INFO" "Running test: $test_name"
    
    # Execute curl command and capture response
    response=$(eval "$curl_command" 2>/dev/null)
    status_code=$(eval "$curl_command -w '%{http_code}' -o /dev/null -s" 2>/dev/null)
    
    if [ "$status_code" = "$expected_status" ]; then
        print_status "SUCCESS" "✓ $test_name (HTTP $status_code)"
        PASSED_TESTS=$((PASSED_TESTS + 1))
        return 0
    else
        print_status "ERROR" "✗ $test_name (Expected HTTP $expected_status, got HTTP $status_code)"
        FAILED_TESTS=$((FAILED_TESTS + 1))
        return 1
    fi
}

# Function to run a test with response validation
run_test_with_validation() {
    local test_name=$1
    local curl_command=$2
    local expected_status=${3:-200}
    local validation_pattern=$4
    
    TOTAL_TESTS=$((TOTAL_TESTS + 1))
    
    print_status "INFO" "Running test: $test_name"
    
    # Execute curl command and capture response
    response=$(eval "$curl_command" 2>/dev/null)
    status_code=$(eval "$curl_command -w '%{http_code}' -o /dev/null -s" 2>/dev/null)
    
    if [ "$status_code" = "$expected_status" ]; then
        if [ -n "$validation_pattern" ] && ! echo "$response" | grep -q "$validation_pattern"; then
            print_status "ERROR" "✗ $test_name (HTTP $status_code but response validation failed)"
            print_status "ERROR" "Expected pattern: $validation_pattern"
            print_status "ERROR" "Actual response: $response"
            FAILED_TESTS=$((FAILED_TESTS + 1))
            return 1
        else
            print_status "SUCCESS" "✓ $test_name (HTTP $status_code)"
            PASSED_TESTS=$((PASSED_TESTS + 1))
            return 0
        fi
    else
        print_status "ERROR" "✗ $test_name (Expected HTTP $expected_status, got HTTP $status_code)"
        FAILED_TESTS=$((FAILED_TESTS + 1))
        return 1
    fi
}

# Function to wait for container to be ready
wait_for_container() {
    local max_attempts=30
    local attempt=1
    
    print_status "INFO" "Waiting for container to be ready..."
    
    while [ $attempt -le $max_attempts ]; do
        if curl -s -f "$BASE_URL/health" > /dev/null 2>&1; then
            print_status "SUCCESS" "Container is ready!"
            return 0
        fi
        
        print_status "INFO" "Attempt $attempt/$max_attempts - Container not ready yet..."
        sleep 2
        attempt=$((attempt + 1))
    done
    
    print_status "ERROR" "Container failed to become ready after $max_attempts attempts"
    return 1
}

# Function to cleanup
cleanup() {
    print_status "INFO" "Cleaning up..."
    docker stop "$CONTAINER_NAME" >/dev/null 2>&1 || true
    docker rm "$CONTAINER_NAME" >/dev/null 2>&1 || true
}

# Function to build and run container
setup_container() {
    print_status "INFO" "Building Docker image..."
    if ! docker build -t "$IMAGE_NAME" .; then
        print_status "ERROR" "Failed to build Docker image"
        exit 1
    fi
    
    print_status "INFO" "Starting container..."
    if ! docker run -d \
        --name "$CONTAINER_NAME" \
        -p "$PORT:11434" \
        -e AWS_REGION=us-east-1 \
        "$IMAGE_NAME"; then
        print_status "ERROR" "Failed to start container"
        exit 1
    fi
    
    # Wait for container to be ready
    if ! wait_for_container; then
        cleanup
        exit 1
    fi
}

# Function to run all tests
run_all_tests() {
    print_status "INFO" "Starting comprehensive API tests..."
    
    # Basic health and utility endpoints
    run_test_with_validation "Health Check" \
        "curl -s '$BASE_URL/health'" \
        200 \
        '"status":"ok"'
    
    run_test "Ping Endpoint" \
        "curl -s '$BASE_URL/ping'" \
        200
    
    run_test_with_validation "Root Endpoint" \
        "curl -s '$BASE_URL/'" \
        200 \
        '"message":"Amazon Q OLLAMA - OLLAMA Compatible API"'
    
    run_test "HEAD Root Endpoint" \
        "curl -s -I '$BASE_URL/'" \
        200
    
    run_test_with_validation "Version Endpoint" \
        "curl -s '$BASE_URL/api/version'" \
        200 \
        '"version":"amazon-q-ollama-1.0.0"'
    
    run_test "Metrics Endpoint" \
        "curl -s '$BASE_URL/metrics'" \
        200
    
    # Model information endpoints
    run_test_with_validation "Tags Endpoint" \
        "curl -s '$BASE_URL/api/tags'" \
        200 \
        '"amazon-q:latest"'
    
    run_test_with_validation "List Endpoint" \
        "curl -s '$BASE_URL/api/list'" \
        200 \
        '"amazon-q:latest"'
    
    run_test_with_validation "Show Endpoint" \
        "curl -s -X POST '$BASE_URL/api/show' -H 'Content-Type: application/json' -d '{\"name\": \"amazon-q\"}'" \
        200 \
        '"modelfile"'
    
    # Process management endpoints
    run_test_with_validation "PS Endpoint" \
        "curl -s '$BASE_URL/api/ps'" \
        200 \
        '"amazon-q:latest"'
    
    run_test_with_validation "Status Endpoint" \
        "curl -s '$BASE_URL/api/status'" \
        200 \
        '"status":"running"'
    
    # Generation endpoints (these will fail without actual Q CLI, but should return proper error responses)
    run_test "Generate Endpoint" \
        "curl -s -X POST '$BASE_URL/api/generate' -H 'Content-Type: application/json' -d '{\"model\": \"amazon-q\", \"prompt\": \"Hello\"}'" \
        500
    
    run_test "Chat Endpoint" \
        "curl -s -X POST '$BASE_URL/api/chat' -H 'Content-Type: application/json' -d '{\"model\": \"amazon-q\", \"messages\": [{\"role\": \"user\", \"content\": \"Hello\"}]}'" \
        500
    
    # Chat endpoint with no user message (should return 400)
    run_test "Chat Endpoint - No User Message" \
        "curl -s -X POST '$BASE_URL/api/chat' -H 'Content-Type: application/json' -d '{\"model\": \"amazon-q\", \"messages\": [{\"role\": \"system\", \"content\": \"You are helpful\"}]}'" \
        400
    
    # Model management endpoints (should return 501 Not Implemented)
    run_test "Create Endpoint" \
        "curl -s -X POST '$BASE_URL/api/create' -H 'Content-Type: application/json' -d '{\"name\": \"test-model\"}'" \
        501
    
    run_test "Pull Endpoint" \
        "curl -s -X POST '$BASE_URL/api/pull' -H 'Content-Type: application/json' -d '{\"name\": \"test-model\"}'" \
        501
    
    run_test "Push Endpoint" \
        "curl -s -X POST '$BASE_URL/api/push' -H 'Content-Type: application/json' -d '{\"name\": \"test-model\"}'" \
        501
    
    run_test "Delete Endpoint" \
        "curl -s -X DELETE '$BASE_URL/api/delete' -H 'Content-Type: application/json' -d '{\"name\": \"test-model\"}'" \
        501
    
    run_test "Copy Endpoint" \
        "curl -s -X POST '$BASE_URL/api/copy' -H 'Content-Type: application/json' -d '{\"source\": \"model1\", \"destination\": \"model2\"}'" \
        501
    
    # Embedding endpoints (should return 501 Not Implemented)
    run_test "Embeddings Endpoint" \
        "curl -s -X POST '$BASE_URL/api/embeddings' -H 'Content-Type: application/json' -d '{\"model\": \"amazon-q\", \"prompt\": \"test\"}'" \
        501
    
    run_test "Embed Endpoint" \
        "curl -s -X POST '$BASE_URL/api/embed' -H 'Content-Type: application/json' -d '{\"model\": \"amazon-q\", \"prompt\": \"test\"}'" \
        501
    
    # Blob endpoints (should return 404 Not Found)
    run_test "Blobs GET Endpoint" \
        "curl -s '$BASE_URL/api/blobs/sha256:test'" \
        404
    
    run_test "Blobs HEAD Endpoint" \
        "curl -s -I '$BASE_URL/api/blobs/sha256:test'" \
        404
    
    run_test "Blobs POST Endpoint" \
        "curl -s -X POST '$BASE_URL/api/blobs/sha256:test'" \
        501
    
    # File upload endpoint
    run_test_with_validation "File Upload Endpoint" \
        "curl -s -X POST '$BASE_URL/upload' -F 'file=@/etc/hostname'" \
        200 \
        '"File uploaded successfully"'
    
    # CORS preflight request
    run_test "CORS Preflight" \
        "curl -s -X OPTIONS '$BASE_URL/api/generate' -H 'Origin: http://localhost:3000' -H 'Access-Control-Request-Method: POST'" \
        204
    
    # Invalid JSON request (should return 400)
    run_test "Invalid JSON Request" \
        "curl -s -X POST '$BASE_URL/api/generate' -H 'Content-Type: application/json' -d 'invalid json'" \
        400
    
    # Streaming endpoints
    run_test "Streaming Generate" \
        "curl -s -X POST '$BASE_URL/api/generate' -H 'Content-Type: application/json' -d '{\"model\": \"amazon-q\", \"prompt\": \"Hello\", \"stream\": true}'" \
        200
    
    run_test "Streaming Chat" \
        "curl -s -X POST '$BASE_URL/api/chat' -H 'Content-Type: application/json' -d '{\"model\": \"amazon-q\", \"messages\": [{\"role\": \"user\", \"content\": \"Hello\"}], \"stream\": true}'" \
        200
}

# Function to run performance tests
run_performance_tests() {
    print_status "INFO" "Running performance tests..."
    
    # Test response times for critical endpoints
    local endpoints=("/health" "/ping" "/api/tags" "/api/ps" "/api/status")
    
    for endpoint in "${endpoints[@]}"; do
        print_status "INFO" "Testing response time for $endpoint"
        
        # Run 10 requests and measure average response time
        local total_time=0
        local successful_requests=0
        
        for i in {1..10}; do
            local start_time=$(date +%s%N)
            if curl -s -f "$BASE_URL$endpoint" > /dev/null; then
                local end_time=$(date +%s%N)
                local request_time=$(( (end_time - start_time) / 1000000 )) # Convert to milliseconds
                total_time=$((total_time + request_time))
                successful_requests=$((successful_requests + 1))
            fi
        done
        
        if [ $successful_requests -gt 0 ]; then
            local avg_time=$((total_time / successful_requests))
            print_status "SUCCESS" "$endpoint - Average response time: ${avg_time}ms (${successful_requests}/10 successful)"
        else
            print_status "ERROR" "$endpoint - All requests failed"
        fi
    done
}

# Function to run load tests
run_load_tests() {
    print_status "INFO" "Running basic load tests..."
    
    # Test concurrent requests to health endpoint
    print_status "INFO" "Testing 20 concurrent requests to /health"
    
    local pids=()
    local start_time=$(date +%s)
    
    # Start 20 concurrent requests
    for i in {1..20}; do
        (curl -s -f "$BASE_URL/health" > /dev/null 2>&1 && echo "SUCCESS" || echo "FAILED") &
        pids+=($!)
    done
    
    # Wait for all requests to complete
    local successful=0
    local failed=0
    
    for pid in "${pids[@]}"; do
        wait $pid
        result=$(jobs -p | wc -l)
    done
    
    local end_time=$(date +%s)
    local duration=$((end_time - start_time))
    
    print_status "SUCCESS" "Load test completed in ${duration}s"
}

# Main execution
main() {
    print_status "INFO" "Starting Amazon Q OLLAMA Container Tests"
    print_status "INFO" "========================================"
    
    # Trap to ensure cleanup on exit
    trap cleanup EXIT
    
    # Setup
    cleanup  # Clean up any existing containers
    setup_container
    
    # Run tests
    run_all_tests
    
    # Run performance tests
    run_performance_tests
    
    # Run load tests
    run_load_tests
    
    # Print summary
    print_status "INFO" "========================================"
    print_status "INFO" "Test Summary:"
    print_status "INFO" "Total Tests: $TOTAL_TESTS"
    print_status "SUCCESS" "Passed: $PASSED_TESTS"
    print_status "ERROR" "Failed: $FAILED_TESTS"
    
    if [ $FAILED_TESTS -eq 0 ]; then
        print_status "SUCCESS" "All tests passed! 🎉"
        exit 0
    else
        print_status "ERROR" "Some tests failed! ❌"
        exit 1
    fi
}

# Check if Docker is available
if ! command -v docker &> /dev/null; then
    print_status "ERROR" "Docker is not installed or not in PATH"
    exit 1
fi

# Check if curl is available
if ! command -v curl &> /dev/null; then
    print_status "ERROR" "curl is not installed or not in PATH"
    exit 1
fi

# Run main function
main "$@"
