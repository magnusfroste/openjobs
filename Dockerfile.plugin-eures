# Dockerfile for EURES Plugin
FROM golang:1.21-alpine AS builder

# Set working directory
WORKDIR /app

# Copy go module files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the plugin binary
RUN CGO_ENABLED=0 GOOS=linux go build -o plugin-eures ./cmd/plugin-eures

# Create minimal runtime image
FROM alpine:latest

# Install ca-certificates for HTTPS requests
RUN apk --no-cache add ca-certificates

# Set working directory
WORKDIR /root/

# Copy the binary from builder
COPY --from=builder /app/plugin-eures .

# Expose port
EXPOSE 8082

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost:8082/health || exit 1

# Run the plugin
CMD ["./plugin-eures"]
