# API Documentation

## Base URL

- **Development**: `http://localhost:8000`
- **Production**: Configure per deployment

---

## Authentication

For Phase 1 (MVP), there is no authentication required.

Production: Add API key or JWT authentication.

---

## Endpoints

### 1. Health Check

**Endpoint**
```
GET /health
```

**Response**
```json
{
  "status": "healthy",
  "timestamp": 1711097400
}
```

---

### 2. Service Status

**Endpoint**
```
GET /status
```

**Response**
```json
{
  "status": "operational",
  "uptime_seconds": 3600,
  "database": {
    "total_events": 150,
    "processed_events": 145,
    "total_knowledge": 145,
    "with_embeddings": 145
  }
}
```

---

### 3. Metrics

**Endpoint**
```
GET /metrics
```

**Response**
```json
{
  "total_events": 150,
  "processed_events": 145,
  "total_knowledge": 145,
  "with_embeddings": 145,
  "embedding_coverage": 100.0
}
```

---

### 4. Semantic Query

**Endpoint**
```
POST /api/v1/query
```

**Request Body**
```json
{
  "query": "What database decisions were made?",
  "top_k": 5,
  "filter": {
    "source": "slack"
  }
}
```

**Query Parameters**
- `query` (required): Natural language search query
- `top_k` (optional): Number of results (default: 5, max: 50)
- `filter` (optional): Filter results by source: `slack`, `github`, etc.

**Response**
```json
{
  "query": "What database decisions were made?",
  "results": [
    {
      "id": "550e8400-e29b-41d4-a716-446655440000",
      "summary": "Team decided to migrate from MongoDB to PostgreSQL for better consistency",
      "similarity_score": 0.92,
      "source": "slack",
      "channel": "#engineering",
      "author": "alice",
      "tags": ["database", "architecture", "migration"],
      "decisions": [
        "Migrate to PostgreSQL",
        "Use pgvector for embeddings"
      ],
      "entities": [
        {
          "name": "PostgreSQL",
          "type": "tool",
          "context": "new database choice",
          "confidence": 0.95
        },
        {
          "name": "MongoDB",
          "type": "tool",
          "context": "legacy database",
          "confidence": 0.90
        }
      ],
      "created_at": "2024-03-22T10:30:00Z"
    },
    {
      "id": "660e8400-e29b-41d4-a716-446655440001",
      "summary": "Discussed schema design patterns for PostgreSQL migration",
      "similarity_score": 0.88,
      "source": "github",
      "channel": "org/backend-repo",
      "author": "bob",
      "tags": ["schema", "database", "design"],
      "decisions": ["Normalize schema", "Add indexes on foreign keys"],
      "entities": [
        {
          "name": "schema_migration",
          "type": "decision",
          "context": "database design task",
          "confidence": 0.85
        }
      ],
      "created_at": "2024-03-22T09:15:00Z"
    }
  ],
  "count": 2,
  "duration": "245ms"
}
```

**Response Codes**
- `200 OK`: Successful search
- `400 Bad Request`: Invalid query
- `500 Internal Server Error`: Server error

---

### 5. Event History

**Endpoint**
```
GET /api/v1/history
```

**Query Parameters**
- `limit` (optional): Results per page (default: 20, max: 100)
- `offset` (optional): Pagination offset (default: 0)

**Response**
```json
{
  "events": [
    {
      "id": "550e8400-e29b-41d4-a716-446655440000",
      "source": "slack",
      "event_type": "message",
      "author": "alice",
      "channel": "#engineering",
      "processing_status": "completed",
      "received_at": "2024-03-22T10:30:00Z",
      "processed_at": "2024-03-22T10:31:15Z"
    }
  ],
  "total": 150,
  "offset": 0,
  "limit": 20
}
```

---

### 6. List Entities

**Endpoint**
```
GET /api/v1/entities
```

**Query Parameters**
- `limit` (optional): Results per page (default: 50, max: 200)
- `offset` (optional): Pagination offset (default: 0)

**Response**
```json
{
  "entities": [
    {
      "id": "550e8400-e29b-41d4-a716-446655440000",
      "name": "PostgreSQL",
      "type": "tool",
      "description": "Open-source relational database",
      "first_mentioned_at": "2024-03-20T08:00:00Z",
      "last_mentioned_at": "2024-03-22T10:30:00Z",
      "mention_count": 15
    },
    {
      "id": "660e8400-e29b-41d4-a716-446655440001",
      "name": "alice",
      "type": "person",
      "description": "Engineering team member",
      "first_mentioned_at": "2024-03-19T09:00:00Z",
      "last_mentioned_at": "2024-03-22T11:45:00Z",
      "mention_count": 42
    }
  ],
  "total": 87,
  "offset": 0,
  "limit": 50
}
```

---

### 7. Get Entity Details

**Endpoint**
```
GET /api/v1/entities/:name
```

**URL Parameters**
- `name`: Entity name (URL encoded if contains special characters)

**Response**
```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "name": "PostgreSQL",
  "type": "tool",
  "description": "Open-source relational database",
  "first_mentioned_at": "2024-03-20T08:00:00Z",
  "last_mentioned_at": "2024-03-22T10:30:00Z",
  "mention_count": 15
}
```

**Response Codes**
- `200 OK`: Entity found
- `404 Not Found`: Entity doesn't exist
- `500 Internal Server Error`: Server error

---

## Error Handling

All errors follow this format:

```json
{
  "error": "Description of what went wrong"
}
```

**Common Errors**
- `400 Bad Request`: Invalid input (missing required fields, invalid JSON)
- `404 Not Found`: Resource not found
- `500 Internal Server Error`: Server-side error (check service logs)

---

## Rate Limiting

Not implemented in MVP. Will be added in production.

---

## Example Workflows

### Workflow 1: Search Related to Architecture

```bash
# Query
curl -X POST http://localhost:8000/api/v1/query \
  -H "Content-Type: application/json" \
  -d '{
    "query": "What architectural decisions impact our API design?",
    "top_k": 10
  }'

# Result contains summaries, decisions, and entities related to API architecture
```

### Workflow 2: Find All Mentions of a Service

```bash
# Get entity details
curl "http://localhost:8000/api/v1/entities/my-microservice"

# Response shows when it was mentioned, who mentioned it, in what context
```

### Workflow 3: Get Recent Activity

```bash
# Get recent events
curl "http://localhost:8000/api/v1/history?limit=20&offset=0"

# Shows last 20 ingested events and their processing status
```

---

## Best Practices

1. **Query Optimization**
   - Use specific queries for better results: "What database did we choose?" vs "database"
   - Adjust `top_k` based on your needs (5-10 for focused results, 20+ for exploration)

2. **Pagination**
   - Use `limit` and `offset` for large result sets
   - Default limit is 20; max is 100

3. **Caching**
   - Results are computed fresh from embeddings
   - Consider caching on client side for repeated queries

4. **Error Handling**
   - Always check HTTP status code
   - Log error messages for debugging
   - Retry 500 errors with exponential backoff

---

## Response Times

- `/api/v1/query`: 100-500ms (depends on vector DB performance)
- `/api/v1/history`: 50-150ms
- `/api/v1/entities`: 50-150ms

---

## Integration Examples

### JavaScript/Node.js

```javascript
async function queryMemory(query, topK = 5) {
  const response = await fetch('http://localhost:8000/api/v1/query', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json'
    },
    body: JSON.stringify({
      query: query,
      top_k: topK
    })
  });

  if (!response.ok) {
    throw new Error(`API error: ${response.status}`);
  }

  return await response.json();
}

// Usage
const results = await queryMemory('What API design decisions were made?');
console.log(results.results[0].summary);
```

### Python

```python
import requests

def query_memory(query: str, top_k: int = 5):
    response = requests.post(
        'http://localhost:8000/api/v1/query',
        json={
            'query': query,
            'top_k': top_k
        }
    )
    response.raise_for_status()
    return response.json()

# Usage
results = query_memory('What API design decisions were made?')
print(results['results'][0]['summary'])
```

### CURL

```bash
curl -X POST http://localhost:8000/api/v1/query \
  -H "Content-Type: application/json" \
  -d '{
    "query": "What API design decisions were made?",
    "top_k": 5
  }' | jq '.results[0]'
```

---

## Future Enhancements

- JWT authentication
- Rate limiting per API key
- Advanced filtering (date ranges, entity types)
- Result ranking customization
- Webhook subscriptions for new knowledge
- Batch query endpoint

