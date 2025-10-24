FROM golang:1.23-alpine AS builder

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o openjobs ./cmd/openjobs

# Final stage
FROM alpine:latest

# Install required packages: ca-certificates for HTTPS
RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Copy the binary
COPY --from=builder /app/openjobs .

# Note: Scheduling is handled by the Go application's internal scheduler
# Set CRON_SCHEDULE environment variable to control sync frequency
# Example: CRON_SCHEDULE=0 6 * * * (daily at 6 AM)

# Expose port
EXPOSE 8080

# Run the application directly
# The Go app has built-in scheduler that respects CRON_SCHEDULE env var
CMD ["./openjobs"]
