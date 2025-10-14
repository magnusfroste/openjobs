FROM golang:1.21-alpine AS builder

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

# Install required packages: ca-certificates for HTTPS, supercronic for cron
RUN apk --no-cache add ca-certificates curl && \
    curl -fsSLO https://github.com/aptible/supercronic/releases/download/v0.2.29/supercronic-linux-amd64 && \
    mv supercronic-linux-amd64 /usr/local/bin/supercronic && \
    chmod +x /usr/local/bin/supercronic

WORKDIR /root/

# Copy the binary
COPY --from=builder /app/openjobs .

# Create crontab file for job sync (every 6 hours)
RUN echo "0 */6 * * * curl -X POST http://localhost:8080/sync/manual || echo 'Sync failed'" > /root/crontab

# Create startup script
RUN echo '#!/bin/sh' > /root/start.sh && \
    echo 'echo "🚀 Starting OpenJobs API server..."' >> /root/start.sh && \
    echo './openjobs &' >> /root/start.sh && \
    echo 'API_PID=$!' >> /root/start.sh && \
    echo 'sleep 5' >> /root/start.sh && \
    echo 'echo "⏰ Starting cron scheduler (job sync every 6 hours)..."' >> /root/start.sh && \
    echo 'supercronic /root/crontab &' >> /root/start.sh && \
    echo 'CRON_PID=$!' >> /root/start.sh && \
    echo 'echo "✅ All services started!"' >> /root/start.sh && \
    echo 'echo "   API PID: $API_PID"' >> /root/start.sh && \
    echo 'echo "   Cron PID: $CRON_PID"' >> /root/start.sh && \
    echo 'wait $API_PID' >> /root/start.sh && \
    chmod +x /root/start.sh

# Expose port
EXPOSE 8080

# Run the startup script
CMD ["/root/start.sh"]