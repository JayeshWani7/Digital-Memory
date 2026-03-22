-- Migration: 003_create_indexes.sql
-- Purpose: Add performance indexes and optimizations
-- Date: March 2026

-- Full-text search index for summary and raw_text
ALTER TABLE knowledge ADD COLUMN fts_document tsvector;

CREATE INDEX idx_knowledge_fts ON knowledge USING GIN(fts_document);

-- Trigger to update full-text search index
CREATE OR REPLACE FUNCTION knowledge_fts_update()
RETURNS TRIGGER AS $$
BEGIN
    NEW.fts_document := 
        setweight(to_tsvector('english', COALESCE(NEW.summary, '')), 'A') ||
        setweight(to_tsvector('english', COALESCE(NEW.raw_text, '')), 'B');
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER knowledge_fts_trigger BEFORE INSERT OR UPDATE ON knowledge
FOR EACH ROW EXECUTE FUNCTION knowledge_fts_update();

-- Composite indexes for common queries
CREATE INDEX idx_knowledge_event_created ON knowledge(event_id, created_at DESC);
CREATE INDEX idx_events_source_status_received ON events(source, processing_status, received_at DESC);

-- Partial indexes for fast pending queries
CREATE INDEX idx_events_pending ON events(received_at DESC) WHERE processing_status = 'pending';
CREATE INDEX idx_events_failed ON events(received_at DESC) WHERE processing_status = 'failed';
CREATE INDEX idx_knowledge_no_embedding ON knowledge(created_at) WHERE embedding IS NULL;

-- Optimize for common date range queries
CREATE INDEX idx_knowledge_created_at_range ON knowledge(created_at DESC);
CREATE INDEX idx_events_received_at_range ON events(received_at DESC);

-- JSON query optimization
CREATE INDEX idx_knowledge_tags_gin ON knowledge USING GIN(tags);
CREATE INDEX idx_knowledge_decisions_gin ON knowledge USING GIN(decisions);

-- Entity lookup optimization
CREATE INDEX idx_entities_created_at ON entities(created_at DESC);

-- Query analytics indexes
CREATE INDEX idx_queries_date_range ON queries(created_at DESC);
CREATE INDEX idx_query_results_knowledge_id ON query_results(knowledge_id);

-- Analysis views for monitoring
CREATE MATERIALIZED VIEW knowledge_by_source AS
    SELECT 
        e.source,
        COUNT(DISTINCT k.id) as knowledge_count,
        COUNT(CASE WHEN k.embedding IS NOT NULL THEN 1 END) as embedded_count,
        AVG(k.confidence) as avg_confidence,
        MAX(k.created_at) as latest_knowledge
    FROM knowledge k
    JOIN events e ON k.event_id = e.id
    GROUP BY e.source;

CREATE MATERIALIZED VIEW entity_statistics AS
    SELECT 
        entity_type,
        COUNT(*) as total_entities,
        COUNT(DISTINCT ke.knowledge_id) as referenced_in_knowledge,
        AVG(mention_count) as avg_mentions,
        MAX(last_mentioned_at) as most_recent_mention
    FROM entities e
    LEFT JOIN knowledge_entities ke ON e.id = ke.entity_id
    GROUP BY entity_type;

-- Create view for processing pipeline status
CREATE MATERIALIZED VIEW processing_pipeline_status AS
    SELECT 
        'ingestion' as stage,
        COUNT(*) as total_items,
        COUNT(CASE WHEN processing_status = 'completed' THEN 1 END) as completed,
        COUNT(CASE WHEN processing_status = 'pending' THEN 1 END) as pending,
        COUNT(CASE WHEN processing_status = 'processing' THEN 1 END) as processing_now,
        COUNT(CASE WHEN processing_status = 'failed' THEN 1 END) as failed
    FROM events
    UNION ALL
    SELECT 
        'embedding' as stage,
        COUNT(*) as total_items,
        COUNT(CASE WHEN embedding IS NOT NULL THEN 1 END) as completed,
        COUNT(CASE WHEN embedding IS NULL THEN 1 END) as pending,
        0 as processing_now,
        0 as failed
    FROM knowledge;

-- Maintenance note: Consider running these periodically:
-- VACUUM ANALYZE;
-- REINDEX INDEX CONCURRENTLY idx_knowledge_embedding;  (after many updates)
