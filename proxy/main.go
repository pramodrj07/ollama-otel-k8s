package main

import (
	"context"
	"io"
	"log"
	"net/http"

	"github.com/go-redis/redis/v8"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/sdk/trace"
)

var (
	redisClient *redis.Client
	ctx         = context.Background()
)

func main() {
	// Redis setup
	redisClient = redis.NewClient(&redis.Options{
		Addr: "redis:6379",
	})

	// Tracing setup
	exporter, _ := otlptracehttp.New(ctx, otlptracehttp.WithEndpoint("otel-collector:4318"), otlptracehttp.WithInsecure())
	tp := trace.NewTracerProvider(trace.WithBatcher(exporter))
	otel.SetTracerProvider(tp)

	http.HandleFunc("/ask", handleRequest)
	http.HandleFunc("/history", handleHistory)
	// Pull API that can pull required models from Ollama
	http.HandleFunc("/pull", pullHandler)

	log.Println("Proxy listening on :8080")
	http.ListenAndServe(":8080", nil)
}

// pullHandler is a placeholder for the pull API that can pull required models from Ollama
func pullHandler(w http.ResponseWriter, r *http.Request) {
	tracer := otel.Tracer("frontend-proxy")
	ctx, span := tracer.Start(r.Context(), "pullHandler")
	defer span.End()

	// Create a new request to Ollama and copy headers
	ollamaReq, err := http.NewRequestWithContext(ctx, "POST", "http://ollama:11434/api/pull", r.Body)
	if err != nil {
		http.Error(w, "Failed to create request to Ollama", http.StatusInternalServerError)
		return
	}
	ollamaReq.Header = r.Header.Clone()

	// Perform the request to Ollama
	client := &http.Client{}
	resp, err := client.Do(ollamaReq)
	if err != nil {
		log.Printf("Error calling Ollama: %v", err)
		span.RecordError(err)
		http.Error(w, "Ollama call failed", http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	// Forward headers and status from Ollama
	for k, vv := range resp.Header {
		for _, v := range vv {
			w.Header().Add(k, v)
		}
	}
	w.WriteHeader(resp.StatusCode)
}

func handleRequest(w http.ResponseWriter, r *http.Request) {
	tracer := otel.Tracer("frontend-proxy")
	ctx, span := tracer.Start(r.Context(), "handleRequest")
	defer span.End()

	// Create a new request to Ollama and copy headers
	ollamaReq, err := http.NewRequestWithContext(ctx, "POST", "http://ollama:11434/api/generate", r.Body)
	if err != nil {
		http.Error(w, "Failed to create request to Ollama", http.StatusInternalServerError)
		return
	}
	ollamaReq.Header = r.Header.Clone()

	// Perform the request to Ollama
	client := &http.Client{}
	resp, err := client.Do(ollamaReq)
	if err != nil {
		//print error
		log.Printf("Error calling Ollama: %v", err)
		span.RecordError(err)
		http.Error(w, "Ollama call failed", http.StatusBadGateway)
		return
	}
	// print resp
	body, _ := io.ReadAll(resp.Body)
	log.Printf("Response from Ollama: %s", body)

	defer resp.Body.Close()

	// Save request to Redis
	redisClient.LPush(ctx, "requests", r.URL.RawQuery)
	redisClient.LTrim(ctx, "requests", 0, 99)

	// Forward headers and status from Ollama
	for k, vv := range resp.Header {
		for _, v := range vv {
			w.Header().Add(k, v)
		}
	}
	w.WriteHeader(resp.StatusCode)

	// Copy response body from Ollama to client
	io.Copy(w, resp.Body)
}

func handleHistory(w http.ResponseWriter, r *http.Request) {
	// Read last 100
	history, _ := redisClient.LRange(ctx, "requests", 0, 99).Result()
	for _, h := range history {
		w.Write([]byte(h + "\n"))
	}
}
