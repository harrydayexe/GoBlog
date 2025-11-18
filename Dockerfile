# Build stage
FROM golang:1.25-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git gcc musl-dev

# Set working directory
WORKDIR /build

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Install templ CLI for generating templates
RUN go install github.com/a-h/templ/cmd/templ@latest

# Generate templ templates
RUN templ generate

# Build the binary
RUN CGO_ENABLED=1 GOOS=linux go build -a -installsuffix cgo -ldflags="-w -s" -o goblogserv ./cmd/GoBlogServ

# Runtime stage
FROM alpine:latest

# Install ca-certificates for HTTPS
RUN apk --no-cache add ca-certificates

# Create non-root user
RUN addgroup -g 1000 goblog && \
    adduser -D -u 1000 -G goblog goblog

# Set working directory
WORKDIR /app

# Copy binary from builder
COPY --from=builder /build/goblogserv /app/goblogserv

# Create directories for content and data
RUN mkdir -p /app/posts /app/data && \
    chown -R goblog:goblog /app

# Switch to non-root user
USER goblog

# Expose port
EXPOSE 8080

# Environment variables with defaults
ENV GOBLOG_CONTENT_PATH=/app/posts \
    GOBLOG_SEARCH_INDEX_PATH=/app/data/blog.bleve \
    GOBLOG_CACHE_ENABLED=true \
    GOBLOG_CACHE_MAX_MB=100 \
    GOBLOG_CACHE_TTL=60m \
    GOBLOG_SEARCH_ENABLED=true \
    GOBLOG_REBUILD_INDEX=false \
    GOBLOG_POSTS_PER_PAGE=10 \
    GOBLOG_VERBOSE=false

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

# Run the binary
ENTRYPOINT ["/app/goblogserv"]

# Default arguments (can be overridden)
CMD ["-host", "0.0.0.0", "-port", "8080"]
