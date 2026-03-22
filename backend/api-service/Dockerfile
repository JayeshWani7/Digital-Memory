# API Service Dockerfile
# Multi-stage build for minimal image size

FROM golang:1.21-alpine AS builder

WORKDIR /app

RUN apk add --no-cache git

COPY go.mod ./
COPY . .
RUN go mod tidy

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o api-service cmd/main.go

# Final stage
FROM alpine:3.19

RUN apk --no-cache add ca-certificates

WORKDIR /root/

COPY --from=builder /app/api-service .

EXPOSE 8000

CMD ["./api-service"]
