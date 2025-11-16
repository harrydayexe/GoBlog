# Build stage
FROM golang:1.24-alpine AS builder

WORKDIR /build

# Install build dependencies
RUN apk add --no-cache git ca-certificates

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the binary
RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-s -w" \
    -o goblogserv \
    ./cmd/GoBlogServ

# Runtime stage
FROM alpine:latest

# Install runtime dependencies
RUN apk --no-cache add ca-certificates tzdata

# Create non-root user
RUN addgroup -g 1000 goblog && \
    adduser -D -u 1000 -G goblog goblog

WORKDIR /app

# Copy binary from builder
COPY --from=builder /build/goblogserv /usr/local/bin/goblogserv

# Set ownership
RUN chown -R goblog:goblog /app

# Switch to non-root user
USER goblog

# Create directory for blog posts
VOLUME ["/posts"]

# Expose default port
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

# Run the server
ENTRYPOINT ["goblogserv"]
CMD ["-content", "/posts", "-port", "8080"]
