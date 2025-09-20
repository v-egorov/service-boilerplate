# Build stage
FROM golang:1.23-alpine AS builder

WORKDIR /app

# Install Air for hot reloading (compatible with Go 1.23)
RUN GOPROXY=$GOPROXY go install github.com/air-verse/air@v1.52.3

# Copy go mod files from project root
COPY ../../go.mod ../../go.sum ./

RUN GOPROXY=$GOPROXY go mod download

# Copy source code
COPY ../.. .

# Tidy go modules
RUN GOPROXY=$GOPROXY go mod tidy

# Build the application
RUN cd api-gateway && CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o api-gateway ./cmd

# Final stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates curl
WORKDIR /root/

# Copy Air binary
COPY --from=builder /go/bin/air /usr/local/bin/air

# Copy the binary from builder stage
COPY --from=builder /app/api-gateway/api-gateway .

# Copy Air configuration
COPY --from=builder /app/api-gateway/.air.toml .

# Expose port
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
  CMD curl -f http://localhost:8080/health || exit 1

# Run the binary
CMD ["./api-gateway"]
