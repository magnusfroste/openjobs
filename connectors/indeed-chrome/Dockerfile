# Dockerfile for Indeed Chrome Plugin
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
RUN CGO_ENABLED=0 GOOS=linux go build -o plugin-indeed-chrome ./cmd/plugin-indeed-chrome

# Runtime stage - Use Chromium-based image
FROM zenika/alpine-chrome:latest

# Switch to root for setup
USER root

# Set working directory
WORKDIR /root/

# Copy the binary from builder
COPY --from=builder /app/plugin-indeed-chrome .

# Expose port
EXPOSE 8087

# Add labels
LABEL description="Indeed Chrome scraper - headless browser based"
LABEL method="headless_chrome"
LABEL advantage="Bypasses Cloudflare bot detection"

# Run the plugin
CMD ["./plugin-indeed-chrome"]
