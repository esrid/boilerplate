# Stage 1: Build
FROM golang:1.22-alpine AS builder

# Enable Go Modules and configure basic dependencies
ENV CGO_ENABLED=0 \
    GOOS=linux \
    GOARCH=amd64

# Install git and other deps (if needed)
RUN apk add --no-cache git

# Set the working directory inside the container
WORKDIR /app

# Copy go mod and sum first, to cache deps
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the source code
COPY . .

# Build the Go binary
RUN go build -o main .

# Stage 2: Run - minimal clean image
FROM alpine:latest

# Install certificates (for HTTPS)
RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Copy the built binary from the builder stage
COPY --from=builder /app/main .

# Expose application port (adjust as needed)
EXPOSE 8080

# Run the Go app
CMD ["./main"]

