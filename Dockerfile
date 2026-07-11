# Build stage
FROM golang:1.22-alpine AS builder

# Install git for version info
RUN apk add --no-cache git

WORKDIR /app

# Copy go mod files first for caching
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the binary
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o yggchat

# Runtime stage
FROM alpine:latest

# Install ca-certificates for HTTPS
RUN apk --no-cache add ca-certificates tzdata

# Create non-root user
RUN adduser -D -g '' yggchat

WORKDIR /app

# Copy binary from builder
COPY --from=builder /app/yggchat .

# Copy web assets
COPY --from=builder /app/web ./web

# Create downloads directory
RUN mkdir -p downloads && chown -R yggchat:yggchat /app

# Switch to non-root user
USER yggchat

# Expose ports
# 8080: Web Console HTTP
# 9000: Yggdrasil peering
EXPOSE 8080 9000

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget -qO- http://localhost:8080/api/state || exit 1

# Run the application
CMD ["./yggchat"]
