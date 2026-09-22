# This Source Code Form is subject to the terms of the Mozilla Public
# License, v. 2.0. If a copy of the MPL was not distributed with this
# file, You can obtain one at https://mozilla.org/MPL/2.0/.

# Build stage
FROM golang:1.26-alpine AS builder

WORKDIR /build

# Copy go mod files for both modules. The CLI module replaces the library with
# ../, so the root go.mod must be present before `go mod download` can resolve.
COPY go.mod go.sum ./
COPY cli/go.mod cli/go.sum ./cli/
RUN go -C cli mod download

# Copy source code
COPY . .

# Build the binary from the CLI module
RUN CGO_ENABLED=0 GOOS=linux go -C cli build -o /build/goblog ./cmd/goblog

# Runtime stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /app

# Copy binary from builder
COPY --from=builder /build/goblog .

# Create directory for posts
RUN mkdir -p /posts

# Expose the blog port and the admin port serving /metrics
EXPOSE 8080
EXPOSE 9090

# Healthchecks
HEALTHCHECK CMD wget --spider -q http://localhost:8080/healthz/startup || exit 1

# Use ENTRYPOINT for the binary, CMD for default args.
# The image opts into metrics because scraping is the usual reason to run it;
# port 9090 is only reachable if the operator publishes it.
ENTRYPOINT ["./goblog", "serve", "--health-checks", "--metrics"]
CMD ["/posts"]
