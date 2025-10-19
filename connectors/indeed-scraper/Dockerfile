# Build stage
FROM golang:1.21-alpine AS builder

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the Indeed Scraper plugin
RUN CGO_ENABLED=0 GOOS=linux go build -o plugin-indeed-scraper ./cmd/plugin-indeed-scraper

# Runtime stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Copy binary from builder
COPY --from=builder /app/plugin-indeed-scraper .

# Expose port
EXPOSE 8086

# Add warning label
LABEL description="Experimental Indeed scraper - web scraping based"
LABEL warning="Check robots.txt before production use"
LABEL method="web_scraping"

# Run the plugin
CMD ["./plugin-indeed-scraper"]
