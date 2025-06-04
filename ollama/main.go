package main

import (
	"context"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os/exec"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
)

func main() {
	// Start Ollama
	cmd := exec.Command("ollama", "serve")
	cmd.Stdout = log.Writer()
	cmd.Stderr = log.Writer()

	if err := cmd.Start(); err != nil {
		log.Fatalf("failed to start Ollama: %v", err)
	}

	// Init OpenTelemetry
	shutdown := InitTracer()
	defer shutdown()

	// Setup reverse proxy to Ollama
	target, _ := url.Parse("http://localhost:11434")
	proxy := httputil.NewSingleHostReverseProxy(target)

	// Wrap proxy with tracing
	http.HandleFunc("/api/generate", func(w http.ResponseWriter, r *http.Request) {
		ctx, span := otel.Tracer("ollama-proxy").Start(r.Context(), "proxy /api/generate")
		defer span.End()

		// Trace metadata, if needed
		span.SetAttributes()

		r = r.WithContext(ctx)
		proxy.ServeHTTP(w, r)
	})

	log.Println("Proxy server running on :8080 -> Ollama :11434")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatalf("proxy failed: %v", err)
	}
}

func InitTracer() func() {
	ctx := context.Background()

	// Exporter to OTEL collector
	exporter, err := otlptracehttp.New(ctx,
		otlptracehttp.WithEndpoint("otel-collector:4318"),
		otlptracehttp.WithInsecure(),
	)
	if err != nil {
		log.Fatalf("failed to create OTLP exporter: %v", err)
	}

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceName("ollama-proxy"),
		)),
	)

	otel.SetTracerProvider(tp)

	return func() {
		if err := tp.Shutdown(ctx); err != nil {
			log.Fatalf("failed to shut down tracer provider: %v", err)
		}
	}
}
