-- Migration: 001_init_schema.sql
-- Purpose: Create initial database schema for Digital Memory Layer
-- Date: March 2026

-- Create enums for event types and statuses
CREATE EXTENSION IF NOT EXISTS vector;
CREATE TYPE event_source AS ENUM ('slack', 'github');
CREATE TYPE event_type AS ENUM ('message', 'pr_created', 'pr_updated', 'commit');
CREATE TYPE processing_status AS ENUM ('pending', 'processing', 'completed', 'failed');
CREATE TYPE entity_type AS ENUM ('service', 'api', 'person', 'tool', 'decision', 'architecture');

-- Create events table (raw incoming data)
CREATE TABLE IF NOT EXISTS events (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    source event_source NOT NULL,
    source_id VARCHAR(255) NOT NULL,  -- External ID (Slack ts, GitHub pr #, etc.)
    event_type event_type NOT NULL,
    raw_data JSONB NOT NULL,           -- Full raw event data
    author VARCHAR(255),                -- Who created this
    channel VARCHAR(255),               -- Channel/repo for context
    received_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    processing_status processing_status DEFAULT 'pending',
    processed_at TIMESTAMP,
    error_message TEXT,
    error_count INT DEFAULT 0,
    last_error_at TIMESTAMP,
    
    CONSTRAINT unique_source_id UNIQUE(source, source_id),
    CONSTRAINT max_retries CHECK(error_count < 5)
);

CREATE INDEX idx_events_status ON events(processing_status);
CREATE INDEX idx_events_source ON events(source);
CREATE INDEX idx_events_received_at ON events(received_at DESC);
CREATE INDEX idx_events_raw_data ON events USING GIN(raw_data);

-- Create entities table (extracted entities)
CREATE TABLE IF NOT EXISTS entities (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    entity_type entity_type NOT NULL,
    description TEXT,
    metadata JSONB,                    -- Additional context (url, email, etc.)
    first_mentioned_at TIMESTAMP,
    last_mentioned_at TIMESTAMP,
    mention_count INT DEFAULT 1,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    CONSTRAINT unique_entity_name_type UNIQUE(name, entity_type)
);

CREATE INDEX idx_entities_type ON entities(entity_type);
CREATE INDEX idx_entities_name ON entities(name);

-- Create knowledge table (processed knowledge with embeddings)
CREATE TABLE IF NOT EXISTS knowledge (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    event_id UUID NOT NULL REFERENCES events(id) ON DELETE CASCADE,
    
    -- Extracted content
    summary TEXT NOT NULL,
    raw_text TEXT NOT NULL,
    
    -- Structured data
    decisions JSONB DEFAULT '[]',      -- Array of decision strings
    tags JSONB DEFAULT '[]',           -- Array of tag strings
    
    -- Processing metadata
    confidence DECIMAL(3,2) CHECK(confidence >= 0 AND confidence <= 1),
    processed_at TIMESTAMP,
    model_used VARCHAR(100),           -- e.g., "gpt-3.5-turbo"
    
    -- For tracking
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_knowledge_event_id ON knowledge(event_id);
CREATE INDEX idx_knowledge_created_at ON knowledge(created_at DESC);
CREATE INDEX idx_knowledge_tags ON knowledge USING GIN(tags);
CREATE INDEX idx_knowledge_decisions ON knowledge USING GIN(decisions);

-- Create knowledge_entities junction table (many-to-many)
CREATE TABLE IF NOT EXISTS knowledge_entities (
    knowledge_id UUID NOT NULL REFERENCES knowledge(id) ON DELETE CASCADE,
    entity_id UUID NOT NULL REFERENCES entities(id) ON DELETE CASCADE,
    context VARCHAR(255),              -- How entity was mentioned
    confidence DECIMAL(3,2),
    
    PRIMARY KEY(knowledge_id, entity_id)
);

CREATE INDEX idx_knowledge_entities_entity ON knowledge_entities(entity_id);

-- Create queries table (for auditing and analytics)
CREATE TABLE IF NOT EXISTS queries (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id VARCHAR(255),              -- Optional user identifier
    query_text TEXT NOT NULL,
    query_type VARCHAR(50),            -- 'semantic_search', 'entity_search', etc.
    top_k INT,
    results_count INT,
    response_time_ms INT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_queries_created_at ON queries(created_at DESC);
CREATE INDEX idx_queries_user_id ON queries(user_id);

-- Create query_results table (for tracking result relevance)
CREATE TABLE IF NOT EXISTS query_results (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    query_id UUID NOT NULL REFERENCES queries(id) ON DELETE CASCADE,
    knowledge_id UUID NOT NULL REFERENCES knowledge(id) ON DELETE SET NULL,
    rank INT,
    similarity_score DECIMAL(4,3),
    was_relevant BOOLEAN,              -- For future feedback
    feedback_at TIMESTAMP
);

CREATE INDEX idx_query_results_query_id ON query_results(query_id);

-- Create processing_errors table (for debugging)
CREATE TABLE IF NOT EXISTS processing_errors (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    event_id UUID REFERENCES events(id) ON DELETE SET NULL,
    service VARCHAR(100),              -- 'ingestion', 'ai', 'api'
    error_type VARCHAR(100),
    error_message TEXT,
    stack_trace TEXT,
    occurred_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    resolved BOOLEAN DEFAULT FALSE
);

CREATE INDEX idx_processing_errors_occurred_at ON processing_errors(occurred_at DESC);
CREATE INDEX idx_processing_errors_event_id ON processing_errors(event_id);

-- Create system_stats table (for monitoring)
CREATE TABLE IF NOT EXISTS system_stats (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    metric_name VARCHAR(100) NOT NULL,
    metric_value DECIMAL(15,2) NOT NULL,
    tags JSONB,
    recorded_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_system_stats_metric ON system_stats(metric_name, recorded_at DESC);

-- Create materialized view for quick stats
CREATE MATERIALIZED VIEW events_summary AS
    SELECT 
        source,
        event_type,
        processing_status,
        COUNT(*) as count,
        COUNT(CASE WHEN error_count > 0 THEN 1 END) as error_count,
        MAX(received_at) as latest
    FROM events
    GROUP BY source, event_type, processing_status;

CREATE INDEX idx_events_summary_source ON events_summary(source);
