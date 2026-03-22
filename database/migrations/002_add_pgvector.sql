-- Migration: 002_add_pgvector.sql
-- Purpose: Add pgvector extension and embedding tables
-- Prerequisites: pgvector extension installed
-- Date: March 2026

-- Create pgvector extension
CREATE EXTENSION IF NOT EXISTS vector;

-- Add embedding column to knowledge table
ALTER TABLE knowledge ADD COLUMN embedding vector(1536);  -- OpenAI embeddings are 1536-dim

-- Create vector similarity index (using IVFFLAT for performance on large datasets)
-- Note: IVFFLAT is good for millions of vectors; for smaller datasets, HNSW is also available
CREATE INDEX ON knowledge USING IVFFLAT (embedding vector_cosine_ops) WITH (lists = 100);

-- For immediate performance in development, you can also create:
-- CREATE INDEX embedding_hnsw_idx ON knowledge USING HNSW (embedding vector_cosine_ops);

-- Create a helper function for semantic search
CREATE OR REPLACE FUNCTION search_similar_knowledge(
    query_embedding vector,
    top_k INT DEFAULT 10,
    similarity_threshold DECIMAL DEFAULT 0.3
)
RETURNS TABLE(
    knowledge_id UUID,
    summary TEXT,
    similarity_score DECIMAL,
    tags JSONB,
    decisions JSONB,
    source event_source,
    created_at TIMESTAMP
) AS $$
BEGIN
    RETURN QUERY
    SELECT 
        k.id,
        k.summary,
        (k.embedding <=> query_embedding)::DECIMAL * -1 + 1,  -- Convert distance to similarity (0-1)
        k.tags,
        k.decisions,
        e.source,
        k.created_at
    FROM knowledge k
    JOIN events e ON k.event_id = e.id
    WHERE (k.embedding <=> query_embedding) < (1 - similarity_threshold)
    ORDER BY k.embedding <=> query_embedding
    LIMIT top_k;
END;
$$ LANGUAGE plpgsql;

-- Create function for batch embedding updates
CREATE OR REPLACE FUNCTION mark_embedding_processed(
    knowledge_id UUID,
    embedding_vector vector,
    model VARCHAR(100)
)
RETURNS VOID AS $$
BEGIN
    UPDATE knowledge
    SET embedding = embedding_vector,
        model_used = model,
        updated_at = CURRENT_TIMESTAMP
    WHERE id = knowledge_id;
END;
$$ LANGUAGE plpgsql;

-- Create function to find knowledge without embeddings (for processing)
CREATE OR REPLACE FUNCTION get_unembedded_knowledge(limit_rows INT DEFAULT 100)
RETURNS TABLE(
    id UUID,
    summary TEXT,
    raw_text TEXT,
    event_id UUID
) AS $$
BEGIN
    RETURN QUERY
    SELECT k.id, k.summary, k.raw_text, k.event_id
    FROM knowledge k
    WHERE k.embedding IS NULL
    ORDER BY k.created_at ASC
    LIMIT limit_rows;
END;
$$ LANGUAGE plpgsql;

-- Create stats view for embedding coverage
CREATE MATERIALIZED VIEW embedding_coverage_stats AS
    SELECT 
        COUNT(*) as total_knowledge_items,
        COUNT(CASE WHEN embedding IS NOT NULL THEN 1 END) as with_embeddings,
        CASE WHEN COUNT(*) > 0 THEN ROUND(100 * COUNT(CASE WHEN embedding IS NOT NULL THEN 1 END)::DECIMAL / COUNT(*), 2) ELSE 0 END as coverage_percent,
        COUNT(CASE WHEN embedding IS NULL THEN 1 END) as pending_embeddings
    FROM knowledge;
