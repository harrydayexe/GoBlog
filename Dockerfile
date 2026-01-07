# Build stage
FROM golang:1.25-alpine AS builder

WORKDIR /build

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the binary
RUN CGO_ENABLED=0 GOOS=linux go build -o goblog ./cmd/goblog

# Runtime stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /app

# Copy binary from builder
COPY --from=builder /build/goblog .

# Create directory for posts
RUN mkdir -p /posts

# Expose default port
EXPOSE 8080

# Use ENTRYPOINT for the binary, CMD for default args
ENTRYPOINT ["./goblog", "serve"]
CMD ["/posts"]
