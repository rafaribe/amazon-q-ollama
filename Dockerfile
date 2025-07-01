# syntax=docker/dockerfile:1

FROM golang:1.24-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o amazon-q-ollama .

# Use your existing Amazon Q container as base
FROM ghcr.io/rafaribe/amazon-q:2025.07.01

# Switch to root to install additional packages and copy binaries
USER root

# Install additional dependencies needed for the API server
RUN apt-get update && apt-get install -y --no-install-recommends \
    ca-certificates \
    && rm -rf /var/lib/apt/lists/* /tmp/* /var/tmp/*

# Copy the compiled Go binary from builder stage
COPY --from=builder /app/amazon-q-ollama /usr/local/bin/amazon-q-ollama
RUN chmod +x /usr/local/bin/amazon-q-ollama

# Create directory for the API server
RUN mkdir -p /app && chown dev:dev /app

# Switch back to non-root user
USER dev
WORKDIR /app

# Expose the API port (OLLAMA default port)
EXPOSE 11434

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD curl -f http://localhost:11434/health || exit 1

# Start the API server instead of the default q chat
ENTRYPOINT ["/usr/local/bin/amazon-q-ollama"]
