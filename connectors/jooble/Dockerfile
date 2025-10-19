# Dockerfile for Jooble Plugin
# Build context should be project root (/)
FROM golang:1.25-alpine AS builder

# Set working directory
WORKDIR /app

# Copy go module files from project root
COPY go.mod go.sum ./
RUN go mod download

# Copy entire project source
COPY . .

# Build the plugin binary from project root context
RUN CGO_ENABLED=0 GOOS=linux go build -o plugin-jooble ./cmd/plugin-jooble

# Runtime stage
FROM alpine:latest

# Install ca-certificates for HTTPS requests
RUN apk --no-cache add ca-certificates

# Set working directory
WORKDIR /root/

# Copy the binary from builder
COPY --from=builder /app/plugin-jooble .

# Expose port
EXPOSE 8088

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost:8088/health || exit 1

# Add labels
LABEL description="Jooble job aggregator plugin"
LABEL source="jooble"
LABEL type="api"

# Run the plugin
CMD ["./plugin-jooble"]
