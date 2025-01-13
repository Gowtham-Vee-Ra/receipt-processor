# Build stage
FROM golang:1.21-alpine AS builder

# Set working directory
WORKDIR /app

# Copy go mod files
COPY go.mod ./
COPY go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o receipt-processor ./cmd/server

# Final stage
FROM alpine:latest

# Add non root user
RUN adduser -D -g '' appuser

# Copy binary from builder
COPY --from=builder /app/receipt-processor /app/receipt-processor

# Use non root user
USER appuser

# Expose port
EXPOSE 8080

# Run the binary
CMD ["/app/receipt-processor"]