# Implementation Guide - Quick Start

## What Have We Built?

You now have a **production-grade, event-driven architecture** for capturing organizational knowledge:

```
Slack/GitHub Events
        ↓
[Ingestion Service] - Port 8001 (Go)
    ↓ Validates & stores in PostgreSQL
[Redis Queue] - Async processing
    ↓ Events waiting for AI processing
[AI Service] - Port 8002 (Python)
    ↓ Calls OpenAI API for extraction
PostgreSQL + PgVector - Stores knowledge with embeddings
    ↓ 
[API Service] - Port 8000 (Go)
    ↓ Semantic search queries
Client Applications
```

---

## Getting Started (Next 15 Minutes)

### 1. Prerequisites Setup

```bash
# Check Go is installed
go version  # Should be 1.21+

# Check Python
python --version  # Should be 3.10+

# Check Docker
docker --version
docker-compose --version
```

### 2. Configure & Start

```bash
cd "Digital Memory"

# Copy and edit environment
cp .env.example .env

# IMPORTANT: Set OPENAI_API_KEY in .env
# Get key from: https://platform.openai.com/api-keys
nano .env    # or your favorite editor
# Set: OPENAI_API_KEY=sk-your-actual-key
```

### 3. Start Infrastructure (3 minutes)

```bash
# Start PostgreSQL, Redis, and admin UIs
docker-compose up -d

# Verify health
docker-compose ps
# All should show "Up"
```

### 4. Initialize Database (2 minutes)

```bash
# Run migrations
psql -h localhost -U memory_user -d digital_memory \
  -f database/migrations/001_init_schema.sql

psql -h localhost -U memory_user -d digital_memory \
  -f database/migrations/002_add_pgvector.sql

psql -h localhost -U memory_user -d digital_memory \
  -f database/migrations/003_create_indexes.sql

# Password: secure_password_change_me
```

### 5. Start Services (Local Development - Better for debugging)

**Terminal 1 - Ingestion Service:**
```bash
cd backend/ingestion-service
go mod download
go run cmd/main.go
# Should see: "Starting ingestion service" on port 8001
```

**Terminal 2 - API Service:**
```bash
cd backend/api-service
go mod download
go run cmd/main.go
# Should see: "Starting API service" on port 8000
```

**Terminal 3 - AI Service:**
```bash
cd backend/ai-service
pip install -r requirements.txt
python -m app.main
# Should see: "Started server process" on port 8002
```

### 6. Load Sample Data (1 minute)

```bash
# From project root, in a new terminal
python test_data/load_test_data.py

# You should see:
# ✓ Loaded Slack message from alice
# ✓ Loaded Slack message from bob
# ... etc
```

### 7. Test the System (2 minutes)

```bash
# Wait 30 seconds for AI processing...

# Test a query
curl -X POST http://localhost:8000/api/v1/query \
  -H "Content-Type: application/json" \
  -d '{
    "query": "What database decisions were made?",
    "top_k": 5
  }' | jq .

# Should return results with summaries and decisions
```

---

## Understanding the Architecture

### Services

| Service | Port | Language | Purpose |
|---------|------|----------|---------|
| **Ingestion** | 8001 | Go | Receives events from Slack/GitHub, validates, stores |
| **AI Processing** | 8002 | Python | Extracts knowledge using OpenAI, generates embeddings |
| **API** | 8000 | Go | Semantic search queries over knowledge |

### Data Flow for a Slack Message

```
1. Slack webhook → Ingestion Service (8001)
2. Validation & storage → PostgreSQL
3. Event published → Redis Stream
4. AI Service (8002) consumes → Calls OpenAI API
5. Knowledge extracted + embedding generated
6. Stored back → PostgreSQL + PgVector
7. Client queries API (8000) → Similarity search
```

### Databases

| Database | Purpose | Tech |
|----------|---------|------|
| **PostgreSQL** | All structured data + vectors | 5432 |
| **Redis** | Message queue for async processing | 6379 |

---

## Key Files to Understand

### Configuration

- `.env.example` → Environment variables template
- `.env` → Your actual configuration (create by copying example)

### Architecture Documentation

- `SYSTEM_DESIGN.md` → Complete system architecture (READ THIS FIRST!)
- `docs/SETUP.md` → Step-by-step setup guide
- `docs/API.md` → API endpoint reference
- `docs/EXAMPLES.md` → Usage examples in Python, JavaScript, Go

### Services

```
backend/ingestion-service/    → Event ingestion (Go)
├── cmd/main.go              → Entry point
├── internal/handlers/        → HTTP handlers
├── internal/database/        → PostgreSQL access
└── internal/queue/           → Redis producer

backend/api-service/          → Query API (Go)
├── cmd/main.go              → Entry point
├── internal/handlers/        → Query handlers
├── internal/database/        → PostgreSQL access
└── internal/vector_db/       → Vector similarity search

backend/ai-service/           → LLM Processing (Python)
├── app/main.py              → FastAPI app
├── app/llm/processor.py      → OpenAI integration
├── app/queue_consumer.py     → Redis consumer
└── app/database.py           → PostgreSQL access
```

### Database

```
database/migrations/
├── 001_init_schema.sql       → Core tables
├── 002_add_pgvector.sql      → Vector embeddings
└── 003_create_indexes.sql    → Performance indexes
```

### Testing & Examples

```
test_data/
├── load_test_data.py         → Load sample data
└── sample_data.py            → Sample data definitions

docs/
├── SETUP.md                  → This guide
├── API.md                    → API reference
├── EXAMPLES.md               → Code examples
└── PHASE_2_ROADMAP.md       → Future enhancements
```

---

## Customization Guide

### 1. Change the LLM Model

**File**: `backend/ai-service/app/config.py`

```python
# Change from gpt-3.5-turbo to gpt-4 for better accuracy
self.openai_model = os.getenv("OPENAI_MODEL", "gpt-4")
```

### 2. Customize Knowledge Extraction

**File**: `backend/ai-service/app/llm/processor.py`

Modify the prompts in `_generate_summary()`, `_extract_entities()`, etc. to fit your domain:

```python
async def _extract_entities(self, text: str, source: str) -> list:
    # Change this prompt to find different entity types
    prompt = f"""
    Extract entities specific to our domain from: {text}
    
    Look for: [YOUR CUSTOM ENTITY TYPES]
    ...
    """
```

### 3. Add New Webhook Sources

**File**: `backend/ingestion-service/internal/handlers/handlers.go`

```go
// Add new HandleJiraEvent(), HandleConfluenceEvent(), etc.
func (h *EventHandler) HandleCustomSource(c *gin.Context) {
    // Similar to HandleSlackEvent()
}

// Register route in main.go
router.POST("/webhook/custom", handler.HandleCustomSource)
```

### 4. Modify API Response Format

**File**: `backend/api-service/internal/models/models.go`

Add fields to `QueryResult` struct to return additional data your clients need.

---

## Monitoring & Debugging

### Check Service Health

```bash
curl http://localhost:8001/health
curl http://localhost:8002/health
curl http://localhost:8000/health

# All should return {"status": "healthy"}
```

### Check Processing Status

```bash
psql -h localhost -U memory_user -d digital_memory

# How many events are pending?
SELECT processing_status, COUNT(*) FROM events GROUP BY processing_status;

# How many embeddings generated?
SELECT COUNT(*) FROM knowledge WHERE embedding IS NOT NULL;

# View recent processed knowledge
SELECT id, summary FROM knowledge ORDER BY created_at DESC LIMIT 5;

# Exit
\q
```

### Watch Logs

```bash
# Ingestion service (if running in Docker)
docker-compose logs -f ingestion-service

# AI service
docker-compose logs -f ai-service

# API service
docker-compose logs -f api-service

# Or check terminal output if running locally
```

### Access Admin Panels

- **Database**: http://localhost:8080 (Adminer)
- **Redis**: http://localhost:8081 (Redis Commander)

---

## Common Issues & Solutions

### Issue: "OpenAI API key not valid"

```
Solution:
1. Check you set OPENAI_API_KEY in .env
2. Verify key is valid: https://platform.openai.com/account/api-keys
3. Check account has available quota
4. Key should start with "sk-"
```

### Issue: "database connection refused"

```
Solution:
1. Verify PostgreSQL is running: docker-compose ps
2. Check password in .env matches docker-compose.yml
3. Wait 10 seconds after starting (migration may be running)
4. Never use 'localhost', use '127.0.0.1' or container name
```

### Issue: "no results from query"

```
Solution:
1. Verify data was loaded: python test_data/load_test_data.py
2. Check embeddings were generated (see "Check Processing Status")
3. Wait 1-2 minutes for AI service to process events
4. Try a simple query: "database"
```

### Issue: "port already in use"

```
Solution:
# Kill process using port
lsof -i :8000  # Shows process using port 8000
kill -9 <PID>

# Or change port in .env
PORT=8100  # Use different port
```

---

## Production Deployment Checklist

- [ ] Set secure passwords in `.env`
- [ ] Enable HTTPS/TLS
- [ ] Setup API authentication (JWT or API keys)
- [ ] Configure rate limiting
- [ ] Setup monitoring & alerting
- [ ] Enable database backups
- [ ] Use managed PostgreSQL (AWS RDS, etc.)
- [ ] Use managed Redis (AWS ElastiCache, etc.)
- [ ] Deploy to Kubernetes or container service
- [ ] Setup CI/CD pipeline
- [ ] Configure CORS properly
- [ ] Add request validation
- [ ] Setup error tracking (Sentry, etc.)
- [ ] Enable structured logging
- [ ] Configure Slack/GitHub webhooks with real secrets

---

## Next Steps After Phase 1 MVP

1. **Integrate Real Data** (1-2 weeks)
   - Setup Slack workspace webhook
   - Configure GitHub repository webhooks
   - Start collecting real organizational knowledge

2. **Customize Extraction** (1 week)
   - Modify LLM prompts for your domain
   - Add custom entity types
   - Tune decision extraction

3. **Build UI** (2-3 weeks)
   - Web dashboard for searching
   - Knowledge browser
   - Analytics dashboard

4. **Add Authentication** (1 week)
   - JWT or OAuth integration
   - User roles and permissions
   - Audit logging

5. **Plan Phase 2** (2 weeks)
   - Knowledge graph design
   - Temporal reasoning
   - Reasoning engine

---

## Support & Documentation

**Quick Links**:
- **System Architecture**: [SYSTEM_DESIGN.md](../SYSTEM_DESIGN.md)
- **Setup Guide**: [docs/SETUP.md](../docs/SETUP.md)
- **API Reference**: [docs/API.md](../docs/API.md)
- **Code Examples**: [docs/EXAMPLES.md](../docs/EXAMPLES.md)
- **Phase 2 Roadmap**: [docs/PHASE_2_ROADMAP.md](../docs/PHASE_2_ROADMAP.md)

**Common Commands**:
```bash
# Start everything
docker-compose up -d && python test_data/load_test_data.py

# Query
curl -X POST http://localhost:8000/api/v1/query \
  -H "Content-Type: application/json" \
  -d '{"query": "What technical decisions were made?", "top_k": 5}'

# Check database
psql -h localhost -U memory_user -d digital_memory

# View logs
docker-compose logs -f

# Stop
docker-compose down
```

---

## Welcome to the Future of Organizational Memory! 🚀

You now have a production-ready system for capturing, processing, and intelligently querying your organization's knowledge. Use Phase 1 to establish the foundation, learn from the data, and plan for Phase 2 enhancements.

Good luck!

