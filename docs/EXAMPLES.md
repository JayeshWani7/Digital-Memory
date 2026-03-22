# Examples and Use Cases

## Query Examples

### 1. Architecture Decisions

```bash
curl -X POST http://localhost:8000/api/v1/query \
  -H "Content-Type: application/json" \
  -d '{
    "query": "What architectural decisions have we made?",
    "top_k": 5
  }'
```

**Expected Results**: Finds decisions about microservices, database choices, caching strategies

---

### 2. Technology Stack Choices

```bash
curl -X POST http://localhost:8000/api/v1/query \
  -H "Content-Type: application/json" \
  -d '{
    "query": "What database and cache technologies are we using?",
    "top_k": 10
  }'
```

**Expected Results**: PostgreSQL, Redis, mentions of migration from MongoDB

---

### 3. Performance Improvements

```bash
curl -X POST http://localhost:8000/api/v1/query \
  -H "Content-Type: application/json" \
  -d '{
    "query": "What performance optimizations have been implemented?",
    "top_k": 5
  }'
```

**Expected Results**: Caching improvements, response time optimization, connection pooling

---

### 4. Team Discussions

```bash
curl -X POST http://localhost:8000/api/v1/query \
  -H "Content-Type: application/json" \
  -d '{
    "query": "What did the team discuss about scaling?",
    "top_k": 5
  }'
```

**Expected Results**: Microservices, independent scaling, Kubernetes

---

## REST API Examples

### Check Service Health

```bash
# All services
curl http://localhost:8001/health  # Ingestion
curl http://localhost:8002/health  # AI
curl http://localhost:8000/health  # API
```

### Get Service Status

```bash
# Full status with metrics
curl http://localhost:8000/status | jq .
```

### Get Recent Events

```bash
# Last 10 events
curl "http://localhost:8000/api/v1/history?limit=10" | jq .

# With pagination
curl "http://localhost:8000/api/v1/history?limit=20&offset=20" | jq .
```

### List All Entities

```bash
# All entities found in knowledge
curl "http://localhost:8000/api/v1/entities?limit=20" | jq .

# Get details on specific entity
curl "http://localhost:8000/api/v1/entities/PostgreSQL" | jq .
curl "http://localhost:0:0/api/v1/entities/alice" | jq .
```

---

## Integration Examples

### JavaScript/TypeScript

```javascript
// lib/memoryLayer.ts
class DigitalMemoryClient {
  private baseUrl: string;

  constructor(baseUrl = 'http://localhost:8000') {
    this.baseUrl = baseUrl;
  }

  async query(question: string, topK = 5): Promise<QueryResponse> {
    const response = await fetch(`${this.baseUrl}/api/v1/query`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({
        query: question,
        top_k: topK
      })
    });

    if (!response.ok) {
      throw new Error(`Query failed: ${response.statusText}`);
    }

    return response.json();
  }

  async getHistory(limit = 20, offset = 0): Promise<HistoryResponse> {
    const response = await fetch(
      `${this.baseUrl}/api/v1/history?limit=${limit}&offset=${offset}`
    );

    if (!response.ok) {
      throw new Error(`Failed to get history: ${response.statusText}`);
    }

    return response.json();
  }

  async listEntities(limit = 50): Promise<EntityResponse> {
    const response = await fetch(
      `${this.baseUrl}/api/v1/entities?limit=${limit}`
    );

    if (!response.ok) {
      throw new Error(`Failed to get entities: ${response.statusText}`);
    }

    return response.json();
  }
}

// Usage
const memory = new DigitalMemoryClient();

// Ask a question
const results = await memory.query('What architecture decisions did we make?');
console.log(results.results[0].summary);

// Get recent events
const history = await memory.getHistory(10);
console.log(`Latest event: ${history.events[0].event_type}`);

// Find all software being used
const entities = await memory.listEntities();
const tools = entities.entities.filter(e => e.type === 'tool');
console.log('Technologies:', tools.map(t => t.name).join(', '));
```

### Python

```python
# digital_memory/client.py
import requests
from typing import List, Dict, Any

class DigitalMemoryClient:
    def __init__(self, base_url: str = 'http://localhost:8000'):
        self.base_url = base_url

    def query(self, question: str, top_k: int = 5) -> Dict[str, Any]:
        """Query the knowledge base semantically"""
        response = requests.post(
            f'{self.base_url}/api/v1/query',
            json={
                'query': question,
                'top_k': top_k
            }
        )
        response.raise_for_status()
        return response.json()

    def get_history(self, limit: int = 20, offset: int = 0) -> Dict[str, Any]:
        """Get event history"""
        response = requests.get(
            f'{self.base_url}/api/v1/history',
            params={'limit': limit, 'offset': offset}
        )
        response.raise_for_status()
        return response.json()

    def list_entities(self, limit: int = 50) -> Dict[str, Any]:
        """List all entities"""
        response = requests.get(
            f'{self.base_url}/api/v1/entities',
            params={'limit': limit}
        )
        response.raise_for_status()
        return response.json()

# Usage
if __name__ == '__main__':
    memory = DigitalMemoryClient()

    # Ask a question
    results = memory.query('What database decisions were made?')
    for result in results['results']:
        print(f"Summary: {result['summary']}")
        print(f"Source: {result['source']} by {result['author']}")
        print(f"Decisions: {', '.join(result['decisions'])}")
        print()

    # Get recent events
    history = memory.get_history(limit=5)
    print(f"Total events: {history['total']}")

    # Find all services mentioned
    entities = memory.list_entities()
    services = [e for e in entities['entities'] if e['type'] == 'service']
    print(f"Services: {[s['name'] for s in services]}")
```

### Go

```go
// internal/memory/client.go
package memory

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type Client struct {
	baseURL string
	client  *http.Client
}

type QueryRequest struct {
	Query string `json:"query"`
	TopK  int    `json:"top_k"`
}

type QueryResult struct {
	ID              string        `json:"id"`
	Summary         string        `json:"summary"`
	SimilarityScore float64       `json:"similarity_score"`
	Source          string        `json:"source"`
	Author          string        `json:"author"`
	Tags            []string      `json:"tags"`
	Decisions       []string      `json:"decisions"`
}

type QueryResponse struct {
	Query   string        `json:"query"`
	Results []QueryResult `json:"results"`
	Count   int           `json:"count"`
	Duration string       `json:"duration"`
}

func NewClient(baseURL string) *Client {
	return &Client{
		baseURL: baseURL,
		client:  &http.Client{},
	}
}

func (c *Client) Query(question string, topK int) (*QueryResponse, error) {
	reqBody := QueryRequest{
		Query: question,
		TopK:  topK,
	}

	data, _ := json.Marshal(reqBody)
	resp, err := c.client.Post(
		fmt.Sprintf("%s/api/v1/query", c.baseURL),
		"application/json",
		bytes.NewBuffer(data),
	)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var queryResp QueryResponse
	json.Unmarshal(body, &queryResp)

	return &queryResp, nil
}

// Usage
func main() {
	client := memory.NewClient("http://localhost:8000")

	results, _ := client.Query("What database did we choose?", 5)
	for _, result := range results.Results {
		fmt.Printf("Summary: %s\n", result.Summary)
		fmt.Printf("Author: %s\n", result.Author)
		fmt.Println()
	}
}
```

---

## Common Use Cases

### 1. Onboarding New Team Members

```bash
# "What's our tech stack?"
curl -X POST http://localhost:8000/api/v1/query \
  -H "Content-Type: application/json" \
  -d '{"query": "What is our technology stack and why did we choose it?", "top_k": 10}'

# "What are our architectural principles?"
curl -X POST http://localhost:8000/api/v1/query \
  -H "Content-Type: application/json" \
  -d '{"query": "What are our core architectural principles and design patterns?", "top_k": 10}'
```

### 2. Technical Decision Context

```bash
# "Why did we make certain choices?"
curl -X POST http://localhost:8000/api/v1/query \
  -H "Content-Type: application/json" \
  -d '{"query": "Why did we migrate away from our previous database?", "top_k": 5}'
```

### 3. Knowledge Discovery

```bash
# "What has been accomplished?"
curl -X POST http://localhost:8000/api/v1/query \
  -H "Content-Type: application/json" \
  -d '{"query": "What performance improvements have we made?", "top_k": 20}'
```

### 4. Entity-Specific Queries

```bash
# Get all mentions of a specific service
curl "http://localhost:8000/api/v1/entities/OrderService"
```

### 5. Timeline/History

```bash
# See what changed over time
curl "http://localhost:0:0/api/v1/history?limit=50" | jq '.events | group_by(.received_at) | reverse'
```

---

## Response Format Guide

### Query Response

```json
{
  "query": "What database decisions were made?",
  "results": [
    {
      "id": "uuid",
      "summary": "Team decided to migrate to PostgreSQL...",
      "similarity_score": 0.92,
      "source": "slack",
      "channel": "#engineering",
      "author": "alice",
      "tags": ["database", "migration", "architecture"],
      "decisions": ["Migrate to PostgreSQL", "Use pgvector"],
      "entities": [
        {
          "name": "PostgreSQL",
          "type": "tool"
        }
      ],
      "created_at": "2024-03-22T10:30:00Z"
    }
  ],
  "count": 1,
  "duration": "125ms"
}
```

### Key Insights

- **Similarity Score**: Higher = more relevant (0-1)
- **Tags**: Auto-generated topic tags for result
- **Decisions**: Specific decisions extracted by LLM
- **Entities**: People, services, tools mentioned
- **Duration**: Query response time

---

## Debugging Tips

### Check if data was processed

```bash
psql -h localhost -U memory_user -d digital_memory

# See processing status
SELECT processing_status, COUNT(*) FROM events GROUP BY processing_status;

# See if embeddings were generated
SELECT COUNT(*) FROM knowledge WHERE embedding IS NOT NULL;

# See specific results
SELECT id, summary FROM knowledge LIMIT 5;
```

### Monitor services

```bash
# Check ingestion service
curl http://localhost:8001/status | jq .

# Check AI service
curl http://localhost:8002/status | jq .

# Check API service
curl http://localhost:8000/status | jq .
```

---

## Performance Tips

1. **Adjust `top_k`** based on results needed
   - 1-3: Very focused results
   - 5-10: Good balance
   - 20+: Exploratory search

2. **Use specific queries**
   - Good: "What database migration decisions were made?"
   - Poor: "database"

3. **Cache results** on your client for repeated queries

4. **Paginate history** to avoid large result sets

---

## Next Steps

- Integrate these APIs into your application
- Customize the LLM prompts in `backend/ai-service/app/llm/processor.py`
- Add authentication for production
- Setup monitoring and alerting
- Plan Phase 2 features (knowledge graphs, reasoning)

