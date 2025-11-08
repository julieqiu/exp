# Use official Go image as base
FROM golang:1.23-alpine AS builder

# Set working directory
WORKDIR /build

# Copy go.mod and go.sum
COPY go.mod go.sum* ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the binary
RUN CGO_ENABLED=0 GOOS=linux go build -o testcontainer .

# Use minimal alpine image for final container
FROM alpine:latest

# Install ca-certificates for HTTPS
RUN apk --no-cache add ca-certificates

# Create directories for mounts
RUN mkdir -p /source /librarian /input /output /repo

# Copy binary from builder
COPY --from=builder /build/testcontainer /usr/local/bin/testcontainer

# Set entrypoint
ENTRYPOINT ["/usr/local/bin/testcontainer"]

# Default command (can be overridden)
CMD ["--version"]
