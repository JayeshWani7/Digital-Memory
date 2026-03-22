# Ingestion Service Dockerfile

FROM golang:1.21-alpine AS builder

WORKDIR /app

# Install build dependencies
RUN apk add --no-cache git

# Copy go mod file
COPY go.mod ./

# Copy source code
COPY . .

# Download dependencies and tidy
RUN go mod tidy

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o ingestion-service cmd/main.go

# Final stage
FROM alpine:3.19

RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Copy binary from builder
COPY --from=builder /app/ingestion-service .

EXPOSE 8001

CMD ["./ingestion-service"]
