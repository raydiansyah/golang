# Multi-stage build untuk Go
# Stage 1: Build
FROM golang:1.25-alpine AS builder

WORKDIR /app

# Install git needed for some go deps
RUN apk add --no-cache git

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

# Install certificates and timezone (Essential for connecting to valid SSL DBs like Neon/Supabase)
RUN apk --no-cache add ca-certificates tzdata

# Create a non-root user
RUN addgroup -S appgroup && adduser -S appuser -G appgroup

# Copy binary dari builder stage
COPY --from=builder /app/kasir-api .

# Switch to non-root user
USER appuser

# Expose port
EXPOSE 8080

# Run aplikasi
CMD ["./kasir-api"]