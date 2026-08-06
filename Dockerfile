# Build Stage
FROM golang:1.21-alpine AS builder

WORKDIR /app

# Install build dependencies
RUN apk add --no-cache git

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the API binary (statically linked)
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o deployly-api api/main.go

# Final Stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata useradd

# Create non-root user for security
RUN adduser -D -g '' deploylyuser
USER deploylyuser

WORKDIR /home/deploylyuser/

# Copy binary from builder
COPY --from=builder /app/deployly-api .

# Expose API port
EXPOSE 8080

# Run the API
ENTRYPOINT ["./deployly-api"]
