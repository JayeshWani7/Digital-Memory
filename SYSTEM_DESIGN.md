# Digital Memory Layer - Phase 1 System Design

## 🏗️ System Architecture Overview

```
┌─────────────────────────────────────────────────────────────────────┐
│                         EXTERNAL SOURCES                             │
│                   (Slack, GitHub, etc.)                              │
└────────────────────────┬────────────────────────────────────────────┘
                         │
         ┌───────────────┼───────────────┐
         │               │               │
         ▼               ▼               ▼
   ┌──────────┐    ┌──────────┐    ┌────────────┐
   │  Slack   │    │ GitHub   │    │ (Future)   │
   │Webhooks  │    │  Client  │    │  Sources   │
   └────┬─────┘    └─────┬────┘    └────┬───────┘
        │                │              │
        └────────────────┼──────────────┘
                         │
        ┌────────────────▼──────────────┐
        │                               │
        │  GO INGESTION SERVICE         │
        │  (Port: 8001)                 │
        │                               │
        │ - Webhook handler             │
        │ - GitHub API poller           │
        │ - Event validation            │
        │ - PostgreSQL storage          │
        │ - Kafka producer              │
        └────────────────┬──────────────┘
                         │
                    [KAFKA QUEUE]
                  (in-memory for MVP)
                         │
        ┌────────────────▼──────────────┐
        │                               │
        │  PYTHON AI SERVICE            │
        │  (Port: 8002)                 │
        │                               │
        │ - Event consumer              │
        │ - LLM processing (OpenAI)     │
        │ - Embedding generation        │
        │ - Pinecone/Weaviate upload    │
        │ - Metadata storage            │
        └────────────────┬──────────────┘
                         │
         ┌───────────────┼───────────────┐
         │               │               │
         ▼               ▼               ▼
    [PostgreSQL]   [Vector DB]      [Cache]
    (Metadata)   (Embeddings)      (Optional)
                                         
        ┌────────────────▼──────────────┐
        │                               │
        │  GO API SERVICE               │
        │  (Port: 8000)                 │
        │                               │
        │ - /query endpoint             │
        │ - Authentication              │
        │ - Response formatting         │
        │ - Rate limiting               │
        └────────────────┬──────────────┘
                         │
                    [CLIENTS]
```

---

## 📋 Component Details

### 1. **Go Ingestion Service** (Port: 8001)
**Responsibility**: Accept data from external sources

#### Endpoints:
- `POST /webhook/slack` - Receive Slack events
- `POST /webhook/github` - Receive GitHub events (via webhook)
- `POST /health` - Health check

#### Features:
- Webhook signature validation (Slack, GitHub)
- Rate limiting
- Database persistence
- Event publishing to Kafka
- Structured logging
- Retry logic for failed publishes

#### Data Flow:
```
External Event → Validation → DB Store → Kafka Publish → ACK
```

---

### 2. **Kafka Queue** (MVP: Use Redis Streams or in-memory alternative)

**Rationale for MVP**: Use `Redis Streams` for simplicity rather than full Kafka
- Easier to setup locally
- Sufficient for Phase 1
- Can migrate to Kafka later

**Topics**:
- `events.slack.new`
- `events.github.pr_created`
- `events.github.commit_pushed`
- `events.github.pr_updated`

---

### 3. **Python AI Processing Service** (Port: 8002)
**Responsibility**: Extract knowledge from events

#### Features:
- Consume events from Redis/Kafka
- Call OpenAI API (gpt-4-turbo or gpt-3.5-turbo for cost)
- Generate structured output:
  ```json
  {
    "summary": "string",
    "decisions": ["array of key decisions"],
    "entities": [
      {"name": "string", "type": "service|api|person|tool", "context": "string"}
    ],
    "tags": ["array of tags"],
    "raw_text": "original content",
    "processed_at": "ISO8601"
  }
  ```
- Generate embeddings (via OpenAI or open-source)
- Store in vector DB with metadata
- Health checks and error handling

#### Processing Logic:
```
Event from Queue
  → Fetch full content (if needed)
  → Summarize with LLM
  → Extract entities/decisions
  → Generate embedding
  → Store in vector DB
  → Update metadata DB
  → ACK
```

---

### 4. **Vector Database** 
**Options for MVP**:
- **Pinecone** (free tier: 1M vectors)
- **Weaviate** (open-source, self-hosted)
- **Milvus** (open-source, lightweight)
- **PgVector** (PostgreSQL extension - simplest!)

**Recommendation**: Start with **PgVector** (PostgreSQL extension) for Phase 1
- No external dependency
- Easier deployment
- All data in one place
- Can migrate to dedicated vector DB later

**Schema**:
```
embeddings table:
  id (pk)
  source (slack|github)
  source_id (external ID)
  content (full text)
  summary (LLM summary)
  embedding (vector)
  entities (jsonb)
  tags (jsonb)
  decisions (jsonb)
  created_at
  updated_at
```

---

### 5. **Go API Service** (Port: 8000)
**Responsibility**: Enable querying the knowledge

#### Endpoints:
- `POST /api/v1/query` - Semantic search
- `GET /api/v1/history` - Event history
- `GET /api/v1/entities` - List known entities
- `GET /api/v1/health` - Health check

#### Query Flow:
```
Natural Language Query
  → Generate embedding
  → Vector DB similarity search
  → Fetch metadata from PostgreSQL
  → Rank and format results
  → Return to client
```

---

## 📊 Data Models

### Raw Events (PostgreSQL - events table)
```json
{
  "id": "uuid",
  "source": "slack|github",
  "source_id": "external_id",
  "event_type": "message|pr|commit",
  "raw_data": {},
  "received_at": "ISO8601",
  "processing_status": "pending|processing|completed|failed"
}
```

### Processed Knowledge (PostgreSQL - knowledge table + PgVector)
```json
{
  "id": "uuid",
  "event_id": "fk to events",
  "summary": "string",
  "entities": [{"name": "", "type": "", "context": ""}],
  "decisions": ["string"],
  "tags": ["string"],
  "embedding": "vector",
  "confidence": 0.95,
  "processed_at": "ISO8601"
}
```

---

## 🔄 Event Flow Example

### Slack Message Ingestion
```
1. User posts message in #engineering channel
2. Slack sends webhook to ingestion-service:8001/webhook/slack
3. Service validates signature + stores in PostgreSQL (events table)
4. Service publishes to Redis stream: events.slack.new
5. Python service consumes event
6. Python service calls OpenAI GPT-4:
   - Extracts summary
   - Identifies entities (if any)
   - Extracts key decisions (if any)
7. Python service generates embedding via OpenAI
8. Python service stores in PostgreSQL + PgVector
9. Data is now queryable via /query endpoint
```

### Query Example
```
User asks: "What decisions were made about database migrations?"

1. API service generates embedding for query
2. Vector DB performs similarity search
3. Returns top-N matching embeddings + metadata
4. Format and return results to user
```

---

## 🛠️ Technology Stack Summary

| Component | Technology | Rationale |
|-----------|-----------|-----------|
| Ingestion | Go | Fast, concurrent, minimal resources |
| Message Queue | Redis Streams | Lightweight, good for MVP |
| AI Processing | Python + FastAPI | Rich ecosystem (LangChain, etc.) |
| LLM | OpenAI API | Reliable, good quality |
| Vector Storage | PgVector (PostgreSQL) | Single DB, simpler deployment |
| Metadata DB | PostgreSQL | Proven, scalable |
| API Service | Go | Fast, efficient |

---

## 📁 Folder Structure

```
digital-memory/
├── backend/
│   ├── ingestion-service/
│   │   ├── cmd/
│   │   │   └── main.go
│   │   ├── internal/
│   │   │   ├── handlers/
│   │   │   ├── models/
│   │   │   ├── database/
│   │   │   ├── queue/
│   │   │   └── middleware/
│   │   ├── config/
│   │   ├── go.mod
│   │   └── Dockerfile
│   │
│   ├── api-service/
│   │   ├── cmd/
│   │   │   └── main.go
│   │   ├── internal/
│   │   │   ├── handlers/
│   │   │   ├── models/
│   │   │   ├── database/
│   │   │   ├── vector_db/
│   │   │   └── middleware/
│   │   ├── config/
│   │   ├── go.mod
│   │   └── Dockerfile
│   │
│   └── ai-service/
│       ├── main.py
│       ├── app/
│       │   ├── __init__.py
│       │   ├── config.py
│       │   ├── models.py
│       │   ├── llm/
│       │   ├── embeddings/
│       │   ├── queue_consumer.py
│       │   └── vector_db.py
│       ├── requirements.txt
│       ├── Dockerfile
│       └── tests/
│
├── database/
│   └── migrations/
│       ├── 001_init_schema.sql
│       ├── 002_add_pgvector.sql
│       └── 003_create_indexes.sql
│
├── docs/
│   ├── SETUP.md
│   ├── API.md
│   ├── ARCHITECTURE.md
│   └── EXAMPLES.md
│
├── test_data/
│   ├── sample_slack_messages.json
│   ├── sample_github_events.json
│   └── load_test_data.py
│
├── docker-compose.yml
├── .env.example
└── README.md
```

---

## 🔐 Security Considerations

1. **Webhook Validation**: Verify signatures from Slack/GitHub
2. **API Authentication**: Use JWT or API keys
3. **Rate Limiting**: Prevent abuse
4. **Data Encryption**: At rest (PostgreSQL) and in transit (HTTPS)
5. **Secrets Management**: Use environment variables
6. **Access Control**: Role-based access for different APIs

---

## 🚀 Deployment Strategy (Phase 1)

**Local Development**:
- Docker Compose for all services
- Redis for local queue
- PostgreSQL locally

**Production Ready**:
- Kubernetes (optional for Phase 1)
- Managed PostgreSQL (AWS RDS)
- Managed Redis (AWS ElastiCache)
- OpenAI API

---

## 📈 Scalability Considerations

### Bottlenecks & Solutions:

| Bottleneck | Phase 1 | Phase 2+ |
|-----------|---------|----------|
| Single queue consumer | Single Python service | Multiple workers (Celery) |
| Database throughput | PostgreSQL limits | Read replicas, caching |
| Vector DB | PgVector | Dedicated vector DB (Pinecone) |
| API throughput | Single Go service | Load balancing |

---

## 🔮 Phase 2 Extensions

### Knowledge Graph
```
Nodes: Entities (services, people, APIs, decisions)
Edges: Relationships ("decided_by", "affects", "implemented_in")

Tools: Neo4j or PostgreSQL with recursive queries
```

### Reasoning Engine
```
- Temporal reasoning: "What changed over time?"
- Causal reasoning: "Why was this decision made?"
- Impact analysis: "What depends on this?"

Tools: LangChain agents, Graph algorithms
```

---

## ✅ MVP Success Criteria

1. ✅ Ingest Slack messages and GitHub PRs
2. ✅ Extract summaries and entities with LLM
3. ✅ Store embeddings in PgVector
4. ✅ Query with natural language (top-5 results)
5. ✅ All services dockerized and testable locally
6. ✅ Production-level error handling & logging

---

## 📅 Implementation Timeline (Estimated)

- **Step 1-2**: Project setup, database schema (2 hours)
- **Step 3-4**: Go ingestion service (3 hours)
- **Step 5-6**: Python AI service (4 hours)
- **Step 7-8**: Go API service (3 hours)
- **Step 9-10**: Integration, testing, documentation (3 hours)

**Total**: ~15 hours for complete Phase 1

