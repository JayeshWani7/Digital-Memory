/**
 * TypeScript type definitions for Digital Memory API
 */

export interface SearchResult {
  id: string;
  title: string;
  content: string;
  source: 'slack' | 'github';
  relevance_score: number;
  embedding: number[];
  metadata: {
    channel?: string;
    author?: string;
    timestamp?: number;
    url?: string;
    [key: string]: any;
  };
  created_at: string;
  updated_at: string;
}

export interface SearchQuery {
  query: string;
  top_k?: number;
  filters?: {
    source?: 'slack' | 'github';
    author?: string;
    date_range?: {
      start: string;
      end: string;
    };
  };
}

export interface SearchResponse {
  results: SearchResult[];
  total_count: number;
  query_time_ms: number;
}

export interface HealthCheck {
  status: 'healthy' | 'unhealthy';
  timestamp: number;
}

export interface ServiceStatus {
  status: 'operational' | 'degraded' | 'down';
  uptime_seconds: number;
  database: {
    total_events: number;
    processed_events: number;
    total_knowledge: number;
    with_embeddings: number;
  };
}

export interface ErrorResponse {
  error: string;
  code: string;
  details?: string;
}

export interface SearchFilters {
  source: 'all' | 'slack' | 'github';
  author?: string;
  startDate?: Date;
  endDate?: Date;
}
