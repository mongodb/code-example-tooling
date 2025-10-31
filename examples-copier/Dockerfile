# Build stage
FROM golang:1.23.4-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git ca-certificates

WORKDIR /app

# Copy go mod files first for better caching
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the binary
# CGO_ENABLED=0 for static binary (no C dependencies)
# -ldflags="-w -s" strips debug info to reduce size
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o examples-copier .

# Runtime stage
FROM alpine:latest

# Install ca-certificates for HTTPS requests
RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Copy binary from builder
COPY --from=builder /app/examples-copier .

# Cloud Run sets PORT environment variable
# Our app reads it from config.Port (defaults to 8080)
EXPOSE 8080

# Run the binary
CMD ["./examples-copier"]

