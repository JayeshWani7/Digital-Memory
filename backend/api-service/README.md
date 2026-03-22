# API Service - Go

Semantic search query service

## Build

```bash
go mod download
go build -o api-service cmd/main.go
```

## Run

```bash
go run cmd/main.go
```

## Environment Variables

- `PORT` - HTTP port (default: 8000)
- `DATABASE_URL` - PostgreSQL connection string
