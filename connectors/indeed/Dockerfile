# Build stage
FROM golang:1.23-alpine AS builder

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the Indeed plugin
RUN CGO_ENABLED=0 GOOS=linux go build -o plugin-indeed ./cmd/plugin-indeed

# Runtime stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Copy binary from builder
COPY --from=builder /app/plugin-indeed .

# Expose port
EXPOSE 8085

# Run the plugin
CMD ["./plugin-indeed"]
