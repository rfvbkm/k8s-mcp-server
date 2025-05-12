FROM golang:1.24-alpine AS builder

WORKDIR /app

# Copy go.mod and go.sum
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o k8s-mcp-server main.go

# Use a minimal alpine image for the final stage
FROM alpine:3.19

WORKDIR /app

# Install CA certificates for HTTPS connections
RUN apk --no-cache add ca-certificates

# Copy the binary from the builder stage
COPY --from=builder /app/k8s-mcp-server /app/

# Expose the port the server listens on
EXPOSE 8080

# Set the entrypoint
ENTRYPOINT ["/app/k8s-mcp-server"]
