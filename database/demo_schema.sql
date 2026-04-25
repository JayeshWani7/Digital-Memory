-- ============================================================
--  Digital-Memory: Demo Schema (No pgvector required)
--  Creates all tables with embedding stored as TEXT (for demo)
-- ============================================================

-- Enums
DO $$ BEGIN
  CREATE TYPE event_source AS ENUM ('slack', 'github');
EXCEPTION WHEN duplicate_object THEN NULL; END $$;

DO $$ BEGIN
  CREATE TYPE event_type AS ENUM ('message', 'pr_created', 'pr_updated', 'commit');
EXCEPTION WHEN duplicate_object THEN NULL; END $$;

DO $$ BEGIN
  CREATE TYPE processing_status AS ENUM ('pending', 'processing', 'completed', 'failed');
EXCEPTION WHEN duplicate_object THEN NULL; END $$;

DO $$ BEGIN
  CREATE TYPE entity_type AS ENUM ('service', 'api', 'person', 'tool', 'decision', 'architecture');
EXCEPTION WHEN duplicate_object THEN NULL; END $$;

-- events
CREATE TABLE IF NOT EXISTS events (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    source          event_source NOT NULL,
    source_id       VARCHAR(255) NOT NULL,
    event_type      event_type NOT NULL,
    raw_data        JSONB NOT NULL DEFAULT '{}',
    author          VARCHAR(255),
    channel         VARCHAR(255),
    received_at     TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    processing_status processing_status DEFAULT 'pending',
    processed_at    TIMESTAMP,
    error_message   TEXT,
    error_count     INT DEFAULT 0,
    last_error_at   TIMESTAMP,
    CONSTRAINT unique_source_id UNIQUE(source, source_id),
    CONSTRAINT max_retries CHECK(error_count < 5)
);

-- knowledge  (embedding stored as TEXT to avoid pgvector requirement)
CREATE TABLE IF NOT EXISTS knowledge (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    event_id        UUID NOT NULL REFERENCES events(id) ON DELETE CASCADE,
    summary         TEXT NOT NULL,
    raw_text        TEXT NOT NULL,
    decisions       JSONB DEFAULT '[]',
    tags            JSONB DEFAULT '[]',
    confidence      DECIMAL(3,2) CHECK(confidence >= 0 AND confidence <= 1),
    processed_at    TIMESTAMP,
    model_used      VARCHAR(100),
    embedding       TEXT,           -- stored as TEXT for demo (no pgvector needed)
    similarity_score DECIMAL(5,4) DEFAULT 0,  -- pre-computed score for demo
    created_at      TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at      TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- entities
CREATE TABLE IF NOT EXISTS entities (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name                VARCHAR(255) NOT NULL,
    entity_type         entity_type NOT NULL,
    description         TEXT,
    metadata            JSONB,
    first_mentioned_at  TIMESTAMP,
    last_mentioned_at   TIMESTAMP,
    mention_count       INT DEFAULT 1,
    created_at          TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at          TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT unique_entity_name_type UNIQUE(name, entity_type)
);

-- queries (analytics)
CREATE TABLE IF NOT EXISTS queries (
    id               UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id          VARCHAR(255),
    query_text       TEXT NOT NULL,
    query_type       VARCHAR(50),
    top_k            INT,
    results_count    INT,
    response_time_ms INT,
    created_at       TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- processing_errors
CREATE TABLE IF NOT EXISTS processing_errors (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    event_id      UUID REFERENCES events(id) ON DELETE SET NULL,
    service       VARCHAR(100),
    error_type    VARCHAR(100),
    error_message TEXT,
    stack_trace   TEXT,
    occurred_at   TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    resolved      BOOLEAN DEFAULT FALSE
);

-- Minimal indexes
CREATE INDEX IF NOT EXISTS idx_events_status     ON events(processing_status);
CREATE INDEX IF NOT EXISTS idx_events_source     ON events(source);
CREATE INDEX IF NOT EXISTS idx_events_received   ON events(received_at DESC);
CREATE INDEX IF NOT EXISTS idx_knowledge_event   ON knowledge(event_id);
CREATE INDEX IF NOT EXISTS idx_knowledge_score   ON knowledge(similarity_score DESC);

-- Materialized view for stats
CREATE MATERIALIZED VIEW IF NOT EXISTS events_summary AS
    SELECT source, event_type, processing_status,
           COUNT(*) as count,
           COUNT(CASE WHEN error_count > 0 THEN 1 END) as error_count,
           MAX(received_at) as latest
    FROM events
    GROUP BY source, event_type, processing_status;
