# Ollama OpenTelemetry Kubernetes

A Kubernetes deployment for running Ollama with OpenTelemetry observability.

## Overview

This project provides a complete setup for deploying Ollama in Kubernetes with integrated OpenTelemetry monitoring and observability features.

## Quick Start

### Prerequisites

- Kubernetes cluster
- kubectl configured
- Docker (for building images)

### Deployment

1. Deploy the application to your Kubernetes cluster
2. Set up port forwarding to access services locally

### Accessing Services

Forward the proxy service to access the web interface:
```bash
kubectl port-forward svc/proxy 8080:80
```

Forward the Ollama service to interact with the API:
```bash
kubectl port-forward svc/ollama 11434:11434
```

### Using Ollama

Pull a model using the Ollama API:
```bash
curl -X POST http://localhost:11434/api/pull -d '{
    "name": "llama3"
}'
```

Generate responses:
```bash
curl -X POST http://localhost:11434/api/generate -d '{
    "model": "llama3",
    "prompt": "Hello, how are you?"
}'
```

## Features

- 🚀 Ollama LLM server deployment
- 📊 OpenTelemetry observability
- 🌐 Kubernetes-native architecture
- 🔍 Distributed tracing
- 📈 Metrics collection

## Components

- **Ollama Service**: Core LLM inference server
- **Proxy Service**: Load balancer and routing
- **OpenTelemetry**: Observability and monitoring

## Contributing

Feel free to submit issues and enhancement requests!