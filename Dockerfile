# GPTCode CLI Docker Image
# Multi-stage build for minimal image size
#
# Usage:
#   docker build -t gptcode/cli:latest .
#   docker run -e GPTCODE_TOKEN=$TOKEN gptcode/cli:latest gptcode run --headless
#

# Stage 1: Build
FROM golang:1.24-alpine AS builder

RUN apk add --no-cache git ca-certificates

WORKDIR /build

# Copy go mod files first for cache
COPY go.mod go.sum ./
RUN go mod download

# Copy source
COPY . .

# Build static binary
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o gptcode ./cmd/gptcode

# Stage 2: Minimal runtime
FROM alpine:3.19

RUN apk add --no-cache ca-certificates tzdata git

# Create non-root user
RUN adduser -D -u 1000 gptcode
USER gptcode

WORKDIR /home/gptcode

# Copy binary
COPY --from=builder /build/gptcode /usr/local/bin/gptcode

# Default command
ENTRYPOINT ["gptcode"]
CMD ["--help"]
