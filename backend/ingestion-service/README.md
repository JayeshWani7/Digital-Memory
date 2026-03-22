# Ingestion Service

Go service for ingesting events from Slack and GitHub

## Building

```bash
go mod download
go build -o ingestion-service cmd/main.go
```

## Running

```bash
go run cmd/main.go
```

## Environment Variables

- `PORT` - HTTP port (default: 8001)
- `DATABASE_URL` - PostgreSQL connection string
- `REDIS_URL` - Redis connection string
-SLACK_SIGNING_SECRET` - Slack webhook secret
- `GITHUB_TOKEN` - GitHub API token
- `GITHUB_WEBHOOK_SECRET` - GitHub webhook secret
