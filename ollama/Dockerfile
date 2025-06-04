FROM ubuntu:22.04

# Install Ollama (official method)
RUN apt-get update && apt-get install -y curl \
    && curl -fsSL https://ollama.com/install.sh | sh

EXPOSE 11434
CMD ["ollama", "serve"]