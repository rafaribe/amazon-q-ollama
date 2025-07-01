#!/bin/sh

# Comprehensive API tests for Amazon Q OLLAMA
# This script runs inside a container to test all endpoints

set -e

# Configuration
API_HOST=${API_HOST:-"amazon-q-ollama"}
API_PORT=${API_PORT:-"11434"}
BASE_URL="http://${API_HOST}:${API_PORT}"

# Colors for output (if supported)
if [ -t 1 ]; then
    RED='\033[0;31m'
    GREEN='\033[0;32m'
    YELLOW='\033[1;33m'
    BLUE='\033[0;34m'
    NC='\033[0m'
else
    RED=''
    GREEN=''
    YELLOW=''
    BLUE=''
    NC=''
fi

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
            printf "${BLUE}[INFO]${NC} %s\n" "$message"
            ;;
        "SUCCESS")
            printf "${GREEN}[SUCCESS]${NC} %s\n" "$message"
            ;;
        "ERROR")
            printf "${RED}[ERROR]${NC} %s\n" "$message"
            ;;
        "WARNING")
            printf "${YELLOW}[WARNING]${NC} %s\n" "$message"
            ;;
    esac
}

# Function to run a test
run_test() {
    local test_name=$1
    local method=${2:-"GET"}
    local endpoint=$3
    local data=$4
    local expected_status=${5:-200}
    local validation_pattern=$6
    
    TOTAL_TESTS=$((TOTAL_TESTS + 1))
    
    print_status "INFO" "Running test: $test_name"
    
    # Build curl command
    local curl_cmd="curl -s -w '%{http_code}' -o /tmp/response.json"
    
    if [ "$method" = "POST" ] || [ "$method" = "PUT" ] || [ "$method" = "DELETE" ]; then
        curl_cmd="$curl_cmd -X $method"
    fi
    
    if [ -n "$data" ]; then
        curl_cmd="$curl_cmd -H 'Content-Type: application/json' -d '$data'"
    fi
    
    curl_cmd="$curl_cmd '$BASE_URL$endpoint'"
    
    # Execute curl command
    status_code=$(eval "$curl_cmd" 2>/dev/null || echo "000")
    
    # Read response
    response=""
    if [ -f "/tmp/response.json" ]; then
        response=$(cat /tmp/response.json 2>/dev/null || echo "")
    fi
    
    # Check status code
    if [ "$status_code" = "$expected_status" ]; then
        # Check response pattern if provided
        if [ -n "$validation_pattern" ] && [ -n "$response" ]; then
            if echo "$response" | grep -q "$validation_pattern"; then
                print_status "SUCCESS" "✓ $test_name (HTTP $status_code)"
                PASSED_TESTS=$((PASSED_TESTS + 1))
            else
                print_status "ERROR" "✗ $test_name (HTTP $status_code but response validation failed)"
                print_status "ERROR" "Expected pattern: $validation_pattern"
                print_status "ERROR" "Actual response: $response"
                FAILED_TESTS=$((FAILED_TESTS + 1))
            fi
        else
            print_status "SUCCESS" "✓ $test_name (HTTP $status_code)"
            PASSED_TESTS=$((PASSED_TESTS + 1))
        fi
    else
        print_status "ERROR" "✗ $test_name (Expected HTTP $expected_status, got HTTP $status_code)"
        if [ -n "$response" ]; then
            print_status "ERROR" "Response: $response"
        fi
        FAILED_TESTS=$((FAILED_TESTS + 1))
    fi
    
    # Clean up
    rm -f /tmp/response.json
}

# Wait for API to be ready
wait_for_api() {
    local max_attempts=30
    local attempt=1
    
    print_status "INFO" "Waiting for API to be ready..."
    
    while [ $attempt -le $max_attempts ]; do
        if curl -s -f "$BASE_URL/health" > /dev/null 2>&1; then
            print_status "SUCCESS" "API is ready!"
            return 0
        fi
        
        print_status "INFO" "Attempt $attempt/$max_attempts - API not ready yet..."
        sleep 2
        attempt=$((attempt + 1))
    done
    
    print_status "ERROR" "API failed to become ready after $max_attempts attempts"
    return 1
}

# Run all tests
run_all_tests() {
    print_status "INFO" "Starting comprehensive API tests..."
    print_status "INFO" "Testing against: $BASE_URL"
    
    # Basic utility endpoints
    run_test "Health Check" "GET" "/health" "" 200 '"status":"ok"'
    run_test "Ping Endpoint" "GET" "/ping" "" 200
    run_test "Root Endpoint" "GET" "/" "" 200 '"message":"Amazon Q OLLAMA - OLLAMA Compatible API"'
    run_test "Version Endpoint" "GET" "/api/version" "" 200 '"version":"amazon-q-ollama-1.0.0"'
    run_test "Metrics Endpoint" "GET" "/metrics" "" 200
    
    # Model information endpoints
    run_test "Tags Endpoint" "GET" "/api/tags" "" 200 '"amazon-q:latest"'
    run_test "List Endpoint" "GET" "/api/list" "" 200 '"amazon-q:latest"'
    run_test "PS Endpoint" "GET" "/api/ps" "" 200 '"amazon-q:latest"'
    run_test "Status Endpoint" "GET" "/api/status" "" 200 '"status":"running"'
    
    # Show endpoint
    run_test "Show Endpoint" "POST" "/api/show" '{"name": "amazon-q"}' 200 '"modelfile"'
    
    # Generation endpoints (will fail without Q CLI but should handle gracefully)
    run_test "Generate Endpoint" "POST" "/api/generate" '{"model": "amazon-q", "prompt": "Hello"}' 500
    run_test "Chat Endpoint" "POST" "/api/chat" '{"model": "amazon-q", "messages": [{"role": "user", "content": "Hello"}]}' 500
    
    # Chat endpoint validation
    run_test "Chat No User Message" "POST" "/api/chat" '{"model": "amazon-q", "messages": [{"role": "system", "content": "You are helpful"}]}' 400
    
    # Streaming endpoints
    run_test "Streaming Generate" "POST" "/api/generate" '{"model": "amazon-q", "prompt": "Hello", "stream": true}' 200
    run_test "Streaming Chat" "POST" "/api/chat" '{"model": "amazon-q", "messages": [{"role": "user", "content": "Hello"}], "stream": true}' 200
    
    # Model management endpoints (should return 501)
    run_test "Create Endpoint" "POST" "/api/create" '{"name": "test-model"}' 501
    run_test "Pull Endpoint" "POST" "/api/pull" '{"name": "test-model"}' 501
    run_test "Push Endpoint" "POST" "/api/push" '{"name": "test-model"}' 501
    run_test "Delete Endpoint" "DELETE" "/api/delete" '{"name": "test-model"}' 501
    run_test "Copy Endpoint" "POST" "/api/copy" '{"source": "model1", "destination": "model2"}' 501
    
    # Embedding endpoints (should return 501)
    run_test "Embeddings Endpoint" "POST" "/api/embeddings" '{"model": "amazon-q", "prompt": "test"}' 501
    run_test "Embed Endpoint" "POST" "/api/embed" '{"model": "amazon-q", "prompt": "test"}' 501
    
    # Blob endpoints
    run_test "Blobs GET" "GET" "/api/blobs/sha256:test" "" 404
    run_test "Blobs POST" "POST" "/api/blobs/sha256:test" "" 501
    
    # Error handling tests
    run_test "Invalid JSON" "POST" "/api/generate" 'invalid json' 400
    run_test "Missing Content-Type" "POST" "/api/generate" '{"model": "amazon-q"}' 400
    
    # CORS test
    print_status "INFO" "Testing CORS preflight..."
    TOTAL_TESTS=$((TOTAL_TESTS + 1))
    cors_status=$(curl -s -w '%{http_code}' -o /dev/null \
        -X OPTIONS "$BASE_URL/api/generate" \
        -H "Origin: http://localhost:3000" \
        -H "Access-Control-Request-Method: POST" \
        -H "Access-Control-Request-Headers: Content-Type" 2>/dev/null || echo "000")
    
    if [ "$cors_status" = "204" ]; then
        print_status "SUCCESS" "✓ CORS Preflight (HTTP 204)"
        PASSED_TESTS=$((PASSED_TESTS + 1))
    else
        print_status "ERROR" "✗ CORS Preflight (Expected HTTP 204, got HTTP $cors_status)"
        FAILED_TESTS=$((FAILED_TESTS + 1))
    fi
}

# Performance tests
run_performance_tests() {
    print_status "INFO" "Running performance tests..."
    
    local endpoints="/health /ping /api/tags /api/ps /api/status"
    
    for endpoint in $endpoints; do
        print_status "INFO" "Testing response time for $endpoint"
        
        local total_time=0
        local successful_requests=0
        local max_time=0
        local min_time=999999
        
        for i in $(seq 1 10); do
            local start_time=$(date +%s%N 2>/dev/null || echo "0")
            if [ "$start_time" = "0" ]; then
                # Fallback for systems without nanosecond precision
                start_time=$(date +%s)000000000
            fi
            
            if curl -s -f "$BASE_URL$endpoint" > /dev/null 2>&1; then
                local end_time=$(date +%s%N 2>/dev/null || echo "0")
                if [ "$end_time" = "0" ]; then
                    end_time=$(date +%s)000000000
                fi
                
                local request_time=$(( (end_time - start_time) / 1000000 ))
                total_time=$((total_time + request_time))
                successful_requests=$((successful_requests + 1))
                
                if [ $request_time -gt $max_time ]; then
                    max_time=$request_time
                fi
                if [ $request_time -lt $min_time ]; then
                    min_time=$request_time
                fi
            fi
        done
        
        if [ $successful_requests -gt 0 ]; then
            local avg_time=$((total_time / successful_requests))
            print_status "SUCCESS" "$endpoint - Avg: ${avg_time}ms, Min: ${min_time}ms, Max: ${max_time}ms (${successful_requests}/10 successful)"
        else
            print_status "ERROR" "$endpoint - All requests failed"
        fi
    done
}

# Main execution
main() {
    print_status "INFO" "Amazon Q OLLAMA Container Tests"
    print_status "INFO" "============================="
    
    # Wait for API to be ready
    if ! wait_for_api; then
        print_status "ERROR" "API is not ready, exiting"
        exit 1
    fi
    
    # Run tests
    run_all_tests
    
    # Run performance tests
    run_performance_tests
    
    # Print summary
    print_status "INFO" "============================="
    print_status "INFO" "Test Summary:"
    print_status "INFO" "Total Tests: $TOTAL_TESTS"
    print_status "SUCCESS" "Passed: $PASSED_TESTS"
    
    if [ $FAILED_TESTS -gt 0 ]; then
        print_status "ERROR" "Failed: $FAILED_TESTS"
        print_status "ERROR" "Some tests failed! ❌"
        exit 1
    else
        print_status "SUCCESS" "All tests passed! 🎉"
        exit 0
    fi
}

# Run main function
main "$@"
