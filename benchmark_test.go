package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func BenchmarkHealthEndpoint(b *testing.B) {
	router := setupRouter()
	
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			w := httptest.NewRecorder()
			req, _ := http.NewRequest("GET", "/health", nil)
			router.ServeHTTP(w, req)
		}
	})
}

func BenchmarkPingEndpoint(b *testing.B) {
	router := setupRouter()
	
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			w := httptest.NewRecorder()
			req, _ := http.NewRequest("GET", "/ping", nil)
			router.ServeHTTP(w, req)
		}
	})
}

func BenchmarkTagsEndpoint(b *testing.B) {
	router := setupRouter()
	
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			w := httptest.NewRecorder()
			req, _ := http.NewRequest("GET", "/api/tags", nil)
			router.ServeHTTP(w, req)
		}
	})
}

func BenchmarkPsEndpoint(b *testing.B) {
	router := setupRouter()
	
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			w := httptest.NewRecorder()
			req, _ := http.NewRequest("GET", "/api/ps", nil)
			router.ServeHTTP(w, req)
		}
	})
}

func BenchmarkStatusEndpoint(b *testing.B) {
	router := setupRouter()
	
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			w := httptest.NewRecorder()
			req, _ := http.NewRequest("GET", "/api/status", nil)
			router.ServeHTTP(w, req)
		}
	})
}

func BenchmarkShowEndpoint(b *testing.B) {
	router := setupRouter()
	showReq := ShowRequest{Name: "amazon-q"}
	jsonData, _ := json.Marshal(showReq)
	
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			w := httptest.NewRecorder()
			req, _ := http.NewRequest("POST", "/api/show", bytes.NewBuffer(jsonData))
			req.Header.Set("Content-Type", "application/json")
			router.ServeHTTP(w, req)
		}
	})
}

func BenchmarkGenerateEndpoint(b *testing.B) {
	router := setupRouter()
	genReq := GenerateRequest{
		Model:  "amazon-q",
		Prompt: "Hello world",
		Stream: false,
	}
	jsonData, _ := json.Marshal(genReq)
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/api/generate", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)
	}
}

func BenchmarkChatEndpoint(b *testing.B) {
	router := setupRouter()
	chatReq := ChatRequest{
		Model: "amazon-q",
		Messages: []Message{
			{Role: "user", Content: "Hello"},
		},
		Stream: false,
	}
	jsonData, _ := json.Marshal(chatReq)
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/api/chat", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)
	}
}

func BenchmarkJSONMarshaling(b *testing.B) {
	response := GenerateResponse{
		Model:         "amazon-q",
		Response:      "This is a test response that simulates a typical Amazon Q response",
		Done:          true,
		TotalDuration: 1234567890,
		EvalCount:     25,
		EvalDuration:  987654321,
	}
	
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, _ = json.Marshal(response)
		}
	})
}

func BenchmarkJSONUnmarshaling(b *testing.B) {
	jsonData := `{
		"model": "amazon-q",
		"prompt": "This is a test prompt for benchmarking JSON unmarshaling performance",
		"stream": false,
		"images": [],
		"options": {"temperature": 0.7}
	}`
	
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			var req GenerateRequest
			_ = json.Unmarshal([]byte(jsonData), &req)
		}
	})
}

func BenchmarkCORSMiddleware(b *testing.B) {
	router := setupRouter()
	
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			w := httptest.NewRecorder()
			req, _ := http.NewRequest("OPTIONS", "/api/generate", nil)
			req.Header.Set("Origin", "http://localhost:3000")
			router.ServeHTTP(w, req)
		}
	})
}

func BenchmarkErrorHandling(b *testing.B) {
	router := setupRouter()
	
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			w := httptest.NewRecorder()
			req, _ := http.NewRequest("POST", "/api/generate", bytes.NewBufferString("invalid json"))
			req.Header.Set("Content-Type", "application/json")
			router.ServeHTTP(w, req)
		}
	})
}

func BenchmarkMemoryAllocation(b *testing.B) {
	router := setupRouter()
	genReq := GenerateRequest{
		Model:  "amazon-q",
		Prompt: "Test prompt",
	}
	jsonData, _ := json.Marshal(genReq)
	
	b.ResetTimer()
	b.ReportAllocs()
	
	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/api/generate", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)
	}
}

// Benchmark concurrent requests
func BenchmarkConcurrentHealthRequests(b *testing.B) {
	router := setupRouter()
	
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			w := httptest.NewRecorder()
			req, _ := http.NewRequest("GET", "/health", nil)
			router.ServeHTTP(w, req)
		}
	})
}

func BenchmarkConcurrentTagsRequests(b *testing.B) {
	router := setupRouter()
	
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			w := httptest.NewRecorder()
			req, _ := http.NewRequest("GET", "/api/tags", nil)
			router.ServeHTTP(w, req)
		}
	})
}

// Benchmark different payload sizes
func BenchmarkSmallPayload(b *testing.B) {
	router := setupRouter()
	genReq := GenerateRequest{
		Model:  "amazon-q",
		Prompt: "Hi",
	}
	jsonData, _ := json.Marshal(genReq)
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/api/generate", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)
	}
}

func BenchmarkLargePayload(b *testing.B) {
	router := setupRouter()
	
	// Create a large prompt
	largePrompt := ""
	for i := 0; i < 1000; i++ {
		largePrompt += "This is a large prompt for testing performance with bigger payloads. "
	}
	
	genReq := GenerateRequest{
		Model:  "amazon-q",
		Prompt: largePrompt,
	}
	jsonData, _ := json.Marshal(genReq)
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/api/generate", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)
	}
}

func BenchmarkComplexChatPayload(b *testing.B) {
	router := setupRouter()
	
	// Create a complex chat request with multiple messages
	messages := make([]Message, 10)
	for i := 0; i < 10; i++ {
		role := "user"
		if i%2 == 1 {
			role = "assistant"
		}
		messages[i] = Message{
			Role:    role,
			Content: "This is message number " + string(rune(i)) + " in a complex conversation for benchmarking purposes.",
		}
	}
	
	chatReq := ChatRequest{
		Model:    "amazon-q",
		Messages: messages,
		Stream:   false,
	}
	jsonData, _ := json.Marshal(chatReq)
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/api/chat", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)
	}
}
