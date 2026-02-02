# Phase 4: Dockerfile for Go + MongoDB CRUD API
# Multi-stage build for optimal image size

# Stage 1: Build the Go application
FROM golang:1.21-alpine AS builder

# Set working directory
WORKDIR /app

# Install build dependencies
RUN apk add --no-cache git

# Copy go mod files first (for better caching)
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy the entire source code
COPY . .

# Build the application
# CGO_ENABLED=0 creates a statically linked binary
# -ldflags="-w -s" strips debug information (smaller binary)
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o /app/server ./cmd/api

# Stage 2: Create minimal runtime image
FROM alpine:latest

# Install CA certificates for HTTPS requests (if needed)
RUN apk --no-cache add ca-certificates tzdata

# Create non-root user for security
RUN addgroup -g 1000 appuser && \
    adduser -D -u 1000 -G appuser appuser

# Set working directory
WORKDIR /app

# Copy the binary from builder stage
COPY --from=builder /app/server .

# Change ownership to non-root user
RUN chown -R appuser:appuser /app

# Switch to non-root user
USER appuser

# Expose port 8080
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

# Run the application
CMD ["./server"]