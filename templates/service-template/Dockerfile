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
RUN cd services/SERVICE_NAME && CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o SERVICE_NAME ./cmd

# Final stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates curl
WORKDIR /root/

# Copy the binary from builder stage
COPY --from=builder /app/services/SERVICE_NAME/SERVICE_NAME .

# Expose port
EXPOSE PORT

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
  CMD curl -f http://localhost:PORT/health || exit 1

# Run the binary
CMD ["./SERVICE_NAME"]
