# Digital Memory Layer - Complete Setup Guide

## Prerequisites

- Docker & Docker Compose (for local development)
- Go 1.21+ (for building/running services locally)
- Python 3.10+ (for AI service)
- PostgreSQL CLI tools
- OpenAI API key
- (Optional) Slack & GitHub tokens for webhook configuration

---

## Step 1: Clone & Configure

```bash
cd "Digital Memory"

# Copy environment template
cp .env.example .env

# Edit .env with your configuration
#  - OPENAI_API_KEY (required for LLM processing)
#  - Database credentials
#  - Slack/GitHub tokens (optional for Phase 1)
```

### Key Environment Variables to Set

```env
# Required
OPENAI_API_KEY=sk-your-key-here

# Optional (defaults provided)
SLACK_SIGNING_SECRET=your-secret
GITHUB_TOKEN=ghp_your-token
GITHUB_WEBHOOK_SECRET=your-secret

# Database (change for production)
DB_PASSWORD=secure_password_change_me
```

---

## Step 2: Start Docker Infrastructure

This starts PostgreSQL, Redis, and includes optional admin panels.

```bash
docker-compose up -d
```

Verify all services are healthy:

```bash
docker-compose ps
```

You should see all services with status "Up":
- `digital-memory-postgres` (port 5432)
- `digital-memory-redis` (port 6379)
- `digital-memory-adminer` (port 8080) - Database admin
- `digital-memory-redis-cli` (port 8081) - Redis admin

---

## Step 3: Initialize the Database

###3a. Run Migrations

Open a new terminal and connect to PostgreSQL:

```bash
# Option 1: Using psql directly
psql -h localhost -U memory_user -d digital_memory -f database/migrations/001_init_schema.sql
psql -h localhost -U memory_user -d digital_memory -f database/migrations/002_add_pgvector.sql
psql -h localhost -U memory_user -d digital_memory -f database/migrations/003_create_indexes.sql

# When prompted for password, enter: secure_password_change_me
```

### 3b: Verify Migration

```bash
# Connect to PostgreSQL
psql -h localhost -U memory_user -d digital_memory

# List tables
\dt

# Should see: events, knowledge, entities, queries, processing_errors, etc.

# Check pgvector extension
\dx

# Should include pgvector

# Exit
\q
```

### 3c: Quick Verification Query

```bash
psql -h localhost -U memory_user -d digital_memory -c "SELECT COUNT(*) FROM events;"

# Should return: 0
```

---

## Step 4: Start Services (Development Mode)

You have two options: Docker containers or local development.

### Option A: Using Docker Compose (Recommended for MVP)

```bash
# The services are already defined in docker-compose.yml
# Need to build the images first

# Build images
docker-compose build ingestion-service ai-service api-service

# Start services
docker-compose up ingestion-service ai-service api-service

# View logs
docker-compose logs -f
```

### Option B: Local Development (Debug Mode)

This allows hot-reload and easier debugging.

#### Terminal 1: Ingestion Service

```bash
cd backend/ingestion-service
go mod download
go run cmd/main.go

# You should see: "Starting ingestion service" on port 8001
```

#### Terminal 2: AI Service

```bash
cd backend/ai-service
pip install -r requirements.txt
python -m app.main

# You should see: "Started server process" on port 8002
```

#### Terminal 3: API Service

```bash
cd backend/api-service
go mod download
go run cmd/main.go

# You should see: "Starting API service" on port 8000
```

---

## Step 5: Verify Services are Running

### Health Checks

```bash
# Ingestion Service
curl http://localhost:8001/health

# AI Service
curl http://localhost:8002/health

# API Service
curl http://localhost:8000/health

# All should return: {"status": "healthy", "timestamp": ...}
```

### Status Endpoints

```bash
curl http://localhost:8001/status
curl http://localhost:8002/status
curl http://localhost:8000/status
```

---

## Step 6: Load Sample Data

First, ensure all services are running (they should all respond to /health).

```bash
# From project root
python test_data/load_test_data.py

# You should see:
# - "Loaded X Slack messages"
# - "Loaded Y GitHub events"
# - "Sample data loaded successfully"
```

### What This Does

The script generates:
- 5 sample Slack messages
- 3 sample GitHub PR events
- Posts to ingestion service
- Events are stored in PostgreSQL
- Queued for AI processing

---

## Step 7: Verify Data was Ingested

### Check Database

```bash
psql -h localhost -U memory_user -d digital_memory

# Count events
SELECT COUNT(*) FROM events;

# View recent events
SELECT source, event_type, author, received_at FROM events ORDER BY received_at DESC LIMIT 5;

# Check if processing started
SELECT COUNT(*) FROM knowledge WHERE id IS NOT NULL;

# Exit
\q
```

### Check Ingestion Service Status

```bash
curl http://localhost:8001/status

# Should show:
# - request_count > 0
# - success_count > 0
# - event_stats with counts
```

---

## Step 8: Wait for AI Processing

The AI service processes events asynchronously. This may take 30 seconds - 2 minutes depending on:
- Network latency to OpenAI API
- Queue depth
- OPENAI_API_KEY validity

### Monitor Progress

```bash
# Check how many events are still pending
psql -h localhost -U memory_user -d digital_memory -c \
  "SELECT processing_status, COUNT(*) FROM events GROUP BY processing_status;"

# Check knowledge count (processed events)
psql -h localhost -U memory_user -d digital_memory -c \
  "SELECT COUNT(*) FROM knowledge;"

# Check embeddings coverage
psql -h localhost -U memory_user -d digital_memory -c \
  "SELECT COUNT(*) as with_embedding FROM knowledge WHERE embedding IS NOT NULL;"
```

### Tail Service Logs (if using Docker)

```bash
docker-compose logs -f ai-service | grep -E "(processed|Error|exception)"
```

---

## Step 9: Test Semantic Query

Once AI service has processed events (check Step 8), test the query endpoint:

```bash
# Simple query
curl -X POST http://localhost:8000/api/v1/query \
  -H "Content-Type: application/json" \
  -d '{
    "query": "What technical decisions were made?",
    "top_k": 5
  }'

# Should return results with summary, source, author, etc.
```

### Expected Response

```json
{
  "query": "What technical decisions were made?",
  "results": [
    {
      "id": "uuid",
      "summary": "...",
      "similarity_score": 0.85,
      "source": "slack",
      "author": "john",
      "tags": ["architecture", "database"],
      "decisions": ["Migrate to PostgreSQL", "..."],
      "created_at": "2024-03-22T10:30:00Z"
    }
  ],
  "count": 5,
  "duration": "125ms"
}
```

---

## Step 10: Explore the System

### List Events

```bash
curl "http://localhost:8000/api/v1/history?limit=10&offset=0"
```

### List Entities

```bash
curl "http://localhost:8000/api/v1/entities?limit=20"
```

### Get Entity Details

```bash
curl "http://localhost:8000/api/v1/entities/PostgreSQL"
```

---

## Troubleshooting

### Issue: Services won't start

**Check logs:**
```bash
docker-compose logs <service-name>
# or
# Check terminal output if running locally
```

**Common causes:**
- Port already in use: Change `PORT` in .env
- Database not ready: Wait 10 seconds, try again
- Missing dependencies: Run `go mod download` or `pip install -r requirements.txt`

### Issue: OpenAI API errors

**Check:**
```bash
# Verify API key is set
echo $OPENAI_API_KEY

# Test API key
curl https://api.openai.com/v1/models \
  -H "Authorization: Bearer $OPENAI_API_KEY"
```

**Solutions:**
- Verify API key is valid
- Check OpenAI account has available quota
- Ensure model exists (e.g., gpt-3.5-turbo)

### Issue: No embeddings generated

**Check:**
```bash
# Verify AI service is running
curl http://localhost:8002/health

# Check AI service logs
docker-compose logs ai-service | grep -i error

# Verify Redis is connected
redis-cli -h localhost ping  # Should return PONG
```

### Issue: Queries return empty results

**Possible causes:**
- Embeddings not yet generated (wait for AI processing)
- No data loaded (run test_data/load_test_data.py)
- Database not initialized (re-run migrations)

**Debug:**
```bash
# Check we have knowledge
psql -h localhost -U memory_user -d digital_memory -c "SELECT COUNT(*) FROM knowledge;"

# Check we have embeddings
psql -h localhost -U memory_user -d digital_memory -c \
  "SELECT COUNT(*) FROM knowledge WHERE embedding IS NOT NULL;"
```

---

## Database Admin Panel

Access Adminer (web UI for PostgreSQL) at:
- **URL**: http://localhost:8080
- **Server**: postgres
- **User**: memory_user
- **Password**: secure_password_change_me
- **Database**: digital_memory

---

## Redis Admin Panel

Access Redis Commander (web UI for Redis) at:
- **URL**: http://localhost:8081

---

## Shutting Down

```bash
# Stop all Docker containers
docker-compose down

# Or if running locally:
# Ctrl+C in each terminal window
```

---

## Next Steps

Once the MVP is running:

1. **Integrate Real Data Sources**
   - Configure Slack webhook secrets
   - Setup GitHub repository webhooks
   - Create production API keys

2. **Customize LLM Processing**
   - Edit `backend/ai-service/app/llm/processor.py`
   - Tune extraction prompts
   - Add domain-specific entity types

3. **Optimize Vector Search**
   - Experiment with different embedding models
   - Adjust similarity thresholds
   - Add filtering/metadata search

4. **Production Deployment**
   - Use managed PostgreSQL (AWS RDS, Heroku)
   - Use managed Redis (AWS ElastiCache)
   - Deploy to Kubernetes or container service
   - Setup monitoring & alerting

---

## Quick Reference Commands

```bash
# Start everything
docker-compose up -d

# View logs
docker-compose logs -f

# Check status
curl http://localhost:8001/health
curl http://localhost:8002/health
curl http://localhost:8000/health

# Load sample data
python test_data/load_test_data.py

# Test query
curl -X POST http://localhost:8000/api/v1/query \
  -H "Content-Type: application/json" \
  -d '{"query": "What decisions were made?", "top_k": 5}'

# View database
psql -h localhost -U memory_user -d digital_memory

# Stop everything
docker-compose down
```

---

## Support

For issues:
1. Check logs: `docker-compose logs <service>`
2. Verify environment variables: Check `.env` file
3. Test connectivity: `curl` health endpoints
4. Check database: Use Adminer at localhost:8080
5. Review docs: See [SYSTEM_DESIGN.md](../SYSTEM_DESIGN.md) and [API.md](API.md)

