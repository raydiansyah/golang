# Multi-stage build untuk Go
# Stage 1: Build
FROM golang:1.25-alpine AS builder

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum* ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build aplikasi
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o kasir-api .

# Stage 2: Runtime
FROM alpine:latest

WORKDIR /app

# Copy binary dari builder stage
COPY --from=builder /app/kasir-api .

# Expose port
EXPOSE 8080

# Run aplikasi
CMD ["./kasir-api"]