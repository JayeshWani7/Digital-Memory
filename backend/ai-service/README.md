# AI Service - Python FastAPI

LLM-based knowledge extraction and embedding generation service

## Build]

```bash
pip install -r requirements.txt
```

## Run

```bash
python -m app.main
```

## Environment Variables

- `PORT` - HTTP port (default: 8002)
- `DATABASE_URL` - PostgreSQL connection string
- `REDIS_URL` - Redis connection string
- `OPENAI_API_KEY` - OpenAI API key (required)
- `OPENAI_MODEL` - OpenAI model (default: gpt-3.5-turbo)
- `OPENAI_EMBEDDING_MODEL` - Embedding model (default: text-embedding-3-small)
