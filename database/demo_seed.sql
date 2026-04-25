-- ============================================================
--  Digital-Memory: Demo Seed Data
--  5 knowledge items with different pre-ranked similarity scores
--  Shows the sort.Slice fix: highest score first in every response
-- ============================================================

-- Insert demo events
INSERT INTO events (id, source, source_id, event_type, raw_data, author, channel, processing_status)
VALUES
  ('a0000000-0000-0000-0000-000000000001', 'slack', 'slack-001', 'message',
   '{"text":"Decided to use Redis for caching layer"}', 'alice', 'architecture', 'completed'),
  ('a0000000-0000-0000-0000-000000000002', 'github', 'github-002', 'pr_created',
   '{"title":"Fix: use sorted slice for search ranking"}', 'bob', 'backend', 'completed'),
  ('a0000000-0000-0000-0000-000000000003', 'slack', 'slack-003', 'message',
   '{"text":"API rate limits confirmed at 1000 req/min"}', 'carol', 'api-team', 'completed'),
  ('a0000000-0000-0000-0000-000000000004', 'github', 'github-004', 'commit',
   '{"message":"Refactor database connection pooling"}', 'dave', 'infra', 'completed'),
  ('a0000000-0000-0000-0000-000000000005', 'slack', 'slack-005', 'message',
   '{"text":"Onboarding new microservice team next week"}', 'eve', 'general', 'completed')
ON CONFLICT (source, source_id) DO NOTHING;

-- Insert 5 knowledge items with deliberately varied similarity scores
-- This is the data that proves ranking works correctly
INSERT INTO knowledge (id, event_id, summary, raw_text, decisions, tags, confidence, model_used, similarity_score, processed_at)
VALUES
  (
    'b0000000-0000-0000-0000-000000000001',
    'a0000000-0000-0000-0000-000000000002',
    'Fix: Replace unordered map with sorted slice for semantic search ranking',
    'The original implementation used a Go map to collect search results, which caused non-deterministic ordering. Replaced with a sorted slice using sort.Slice to guarantee descending similarity order.',
    '["Use sort.Slice over map for ordered results","Preserve similarity scores through full pipeline"]',
    '["ranking","semantic-search","go","bugfix","sort"]',
    0.98, 'demo', 0.9821, NOW()
  ),
  (
    'b0000000-0000-0000-0000-000000000003',
    'a0000000-0000-0000-0000-000000000001',
    'Architecture Decision: Redis as caching layer for API responses',
    'Team decided to use Redis as the primary caching layer. Cache TTL set to 5 minutes for search results. This reduces load on PostgreSQL for repeated queries.',
    '["Use Redis for caching","5-minute TTL for search results"]',
    '["redis","caching","architecture","performance"]',
    0.91, 'demo', 0.7654, NOW()
  ),
  (
    'b0000000-0000-0000-0000-000000000004',
    'a0000000-0000-0000-0000-000000000003',
    'API Rate Limiting: 1000 requests per minute confirmed',
    'Confirmed API rate limits at 1000 requests per minute per client. Implemented sliding window rate limiting using Redis counters.',
    '["Rate limit at 1000 req/min","Use sliding window algorithm"]',
    '["api","rate-limiting","redis","backend"]',
    0.85, 'demo', 0.6203, NOW()
  ),
  (
    'b0000000-0000-0000-0000-000000000005',
    'a0000000-0000-0000-0000-000000000004',
    'Database Connection Pool Refactoring',
    'Refactored connection pooling to use pgxpool. Max connections set to 25, idle connections to 5, with a 5-minute lifetime.',
    '["Use pgxpool for connection management","Max 25 connections"]',
    '["database","postgresql","performance","backend"]',
    0.78, 'demo', 0.4891, NOW()
  ),
  (
    'b0000000-0000-0000-0000-000000000006',
    'a0000000-0000-0000-0000-000000000005',
    'Team Announcement: New microservice team onboarding',
    'New team joining next week to work on the notification microservice. Will need access to Slack, GitHub, and the internal wiki.',
    '["Provision access for new team","Schedule onboarding session"]',
    '["team","onboarding","announcement"]',
    0.62, 'demo', 0.2134, NOW()
  )
ON CONFLICT (id) DO UPDATE SET
  similarity_score = EXCLUDED.similarity_score,
  summary = EXCLUDED.summary;

-- Refresh the materialized view
REFRESH MATERIALIZED VIEW events_summary;

-- Verify data
SELECT
  k.id,
  k.summary,
  k.similarity_score,
  k.confidence
FROM knowledge k
ORDER BY k.similarity_score DESC;
