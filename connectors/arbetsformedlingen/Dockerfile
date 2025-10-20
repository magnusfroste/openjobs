# Dockerfile for Arbetsförmedlingen Plugin
# Build context should be project root (/)
FROM golang:1.23-alpine AS builder

# Force cache bust - update this timestamp to force rebuild
ARG CACHE_BUST=2025-10-20-08:37

# Set working directory
WORKDIR /app

# Copy go module files from project root
COPY go.mod go.sum ./
RUN go mod download

# Copy entire project source
COPY . .

# Build the plugin binary from project root context
RUN CGO_ENABLED=0 GOOS=linux go build -o plugin-arbetsformedlingen ./cmd/plugin-arbetsformedlingen

# Create minimal runtime image
FROM alpine:latest

# Install ca-certificates for HTTPS requests
RUN apk --no-cache add ca-certificates

# Set working directory
WORKDIR /root/

# Copy the binary from builder
COPY --from=builder /app/plugin-arbetsformedlingen .

# Expose port
EXPOSE 8081

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost:8081/health || exit 1

# Run the plugin
CMD ["./plugin-arbetsformedlingen"]
