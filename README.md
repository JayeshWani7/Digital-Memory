# Digital Memory Layer - Phase 1

A production-grade backend system that ingests data from Slack and GitHub, extracts structured knowledge using LLMs, stores embeddings in a vector database, and enables semantic querying.

## 🎯 What This Does

- **Ingest**: Captures messages and events from Slack & GitHub
- **Process**: Uses OpenAI LLM to extract summaries, decisions, and entities
- **Store**: Saves embeddings in PostgreSQL (with PgVector extension)
- **Query**: Provides semantic search via natural language

## 📁 Project Structure

```
digital-memory/
├── SYSTEM_DESIGN.md           # Complete architecture & design
├── .env.example               # Environment configuration template
├── docker-compose.yml         # Local development environment
│
├── backend/
│   ├── ingestion-service/     # Go service for data ingestion
│   ├── api-service/           # Go service for querying
│   └── ai-service/            # Python service for LLM processing
│
├── database/
│   └── migrations/            # PostgreSQL migration scripts
│
├── docs/
│   ├── SETUP.md              # Step-by-step setup guide
│   ├── API.md                # API documentation
│   └── EXAMPLES.md           # Example usage & queries
│
└── test_data/
    ├── sample_slack_messages.json
    ├── sample_github_events.json
    └── load_test_data.py
```

## 🚀 Quick Start

### Prerequisites
- Docker & Docker Compose
- Go 1.21+
- Python 3.10+
- PostgreSQL CLI (psql)
- OpenAI API key

### 1. Clone & Setup
```bash
cd "Digital Memory"
cp .env.example .env
# Edit .env with your API keys and configuration
```

### 2. Start Infrastructure
```bash
docker-compose up -d
```

### 3. Run Database Migrations
```bash
# See docs/SETUP.md for detailed migration steps
```

### 4. Start Services
```bash
# Terminal 1: Ingestion Service
cd backend/ingestion-service
go run cmd/main.go

# Terminal 2: AI Service
cd backend/ai-service
python -m app.main

# Terminal 3: API Service
cd backend/api-service
go run cmd/main.go
```

### 5. Send Test Data
```bash
python test_data/load_test_data.py
```

### 6. Query
```bash
curl -X POST http://localhost:8000/api/v1/query \
  -H "Content-Type: application/json" \
  -d '{"query": "What database decisions were made?", "top_k": 5}'
```

## 📚 Documentation

- **[SYSTEM_DESIGN.md](SYSTEM_DESIGN.md)** - Architecture, design decisions, data models
- **[docs/SETUP.md](docs/SETUP.md)** - Detailed setup guide (WIP)
- **[docs/API.md](docs/API.md)** - Complete API reference (WIP)
- **[docs/EXAMPLES.md](docs/EXAMPLES.md)** - Usage examples and test cases (WIP)

## 🏗️ Architecture Overview

### Three Microservices

1. **Go Ingestion Service** (Port 8001)
   - Accepts Slack webhooks and GitHub events
   - Validates and stores raw data
   - Publishes to message queue

2. **Python AI Service** (Port 8002)
   - Consumes events from queue
   - Calls OpenAI API for processing
   - Generates embeddings
   - Stores results in PostgreSQL + PgVector

3. **Go API Service** (Port 8000)
   - Provides `/query` endpoint for semantic search
   - Converts queries to embeddings
   - Returns relevant results with metadata

### Data Flow

```
Slack/GitHub Events
  ↓
[Ingestion Service] → PostgreSQL + Redis Queue
  ↓
[AI Service] → OpenAI API + Embedding Generation
  ↓
PostgreSQL + PgVector
  ↓
[API Service] ← /query endpoint
  ↓
Client Response
```

## 🔧 Tech Stack

| Component | Technology | Why |
|-----------|-----------|-----|
| Ingestion | Go | Fast, concurrent, minimal overhead |
| Queue | Redis Streams | Lightweight, MVP-friendly |
| Processing | Python + FastAPI | Rich LLM ecosystem |
| LLM | OpenAI | Reliable, high quality |
| Vector DB | PostgreSQL + PgVector | Simplified deployment, single DB |
| API | Go | High performance |
| Metadata DB | PostgreSQL | Proven, scalable |

## 📊 Data Models

### Events Table
Raw incoming data from sources

### Knowledge Table
Processed knowledge with:
- Summary
- Entities (services, people, APIs)
- Decisions
- Tags
- Embeddings

## 🔐 Security

- Webhook signature validation (Slack/GitHub)
- JWT authentication for API
- Rate limiting
- Environment-based secrets
- HTTPS ready

## 🧪 Testing

```bash
# Run all tests
./scripts/test.sh

# Test ingestion service
cd backend/ingestion-service && go test ./...

# Test API service
cd backend/api-service && go test ./...

# Test AI service
cd backend/ai-service && pytest tests/
```

## 📈 Monitoring & Logging

All services include:
- Structured logging (JSON format)
- Health check endpoints
- Metrics endpoints
- Error tracking

## 🚗 Roadmap

### Phase 1 (Current) ✅
- Slack & GitHub ingestion
- LLM-based knowledge extraction
- Vector embeddings + semantic search

### Phase 2 🔮
- Knowledge graph construction
- Temporal reasoning
- Entity relationship discovery
- Causal analysis

### Phase 3 🎯
- Reasoning engine (LangChain agents)
- Question-answering over knowledge
- Integration recommendations

## 🤝 Contributing

This is a production-ready system template. Feel free to:
- Extend with new data sources
- Add more processing services
- Customize LLM prompts
- Deploy to production infrastructure

## 📝 License

MIT

## 📞 Support

For issues or questions about the implementation, refer to:
1. SYSTEM_DESIGN.md for architecture questions
2. docs/SETUP.md for setup issues
3. docs/API.md for API questions
4. Service logs for runtime issues

---

**Status**: Phase 1 Implementation in Progress

Last Updated: March 2026
