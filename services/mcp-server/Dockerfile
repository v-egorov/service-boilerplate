# Build stage
FROM golang:1.25-alpine AS builder

WORKDIR /app



# Copy go mod files from project root
COPY ../../go.mod ../../go.sum ./
RUN GOPROXY=$GOPROXY go mod download

# Copy source code
COPY ../.. .

# Tidy go modules
RUN GOPROXY=$GOPROXY go mod tidy

# Build the application
RUN cd services/mcp-server && CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o mcp-server ./cmd

# Final stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates curl
WORKDIR /root/

# Copy the binary from builder stage
COPY --from=builder /app/services/mcp-server/mcp-server .

# Expose port
EXPOSE 8095

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
  CMD curl -f http://localhost:8095/health || exit 1

# Run the binary
CMD ["./mcp-server"]
