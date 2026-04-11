# Build stage
FROM golang:1.25-alpine AS builder

# Install git (needed for some Go modules)
RUN apk add --no-cache git

# Set working directory
WORKDIR /build

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN GOPROXY=$GOPROXY go mod download

# Copy source code
COPY . .

# Tidy go modules
RUN GOPROXY=$GOPROXY go mod tidy

# Build the migrate-wrapper binary
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -installsuffix cgo -o migrate-wrapper .

# Runtime stage
FROM alpine:latest

# Install required packages
RUN apk add --no-cache postgresql-client jq

# Copy golang-migrate binary from the official image
COPY --from=migrate/migrate:latest /migrate /usr/local/bin/migrate

# Copy our wrapper binary
COPY --from=builder /build/migrate-wrapper /usr/local/bin/migrate-wrapper

# Make sure binaries are executable
RUN chmod +x /usr/local/bin/migrate /usr/local/bin/migrate-wrapper

# Create a non-root user
RUN addgroup -g 1001 -S appgroup && \
    adduser -u 1001 -S appuser -G appgroup

# Set working directory to root so volume mounts work correctly
WORKDIR /

# Add /usr/local/bin to PATH
ENV PATH="/usr/local/bin:$PATH"

# Switch to non-root user
USER appuser

# Set the entrypoint
ENTRYPOINT ["migrate-wrapper"]
