package models

import "time"

// QueryRequest represents a semantic search query
type QueryRequest struct {
	Query  string `json:"query" binding:"required"`
	TopK   int    `json:"top_k"`
	Filter map[string]interface{} `json:"filter,omitempty"`
}

// SearchMatch represents a ranked similarity match before result hydration.
type SearchMatch struct {
	ID    string
	Score float64
}

// QueryResult represents a single search result
type QueryResult struct {
	ID             string                 `json:"id"`
	Summary        string                 `json:"summary"`
	RawText        string                 `json:"raw_text,omitempty"`
	SimilarityScore float64                `json:"similarity_score"`
	Source         string                 `json:"source"`
	Channel        string                 `json:"channel"`
	Author         string                 `json:"author"`
	Tags           []string               `json:"tags"`
	Decisions      []string               `json:"decisions"`
	Entities       []map[string]interface{} `json:"entities"`
	CreatedAt      time.Time              `json:"created_at"`
}

// QueryResponse represents the response to a semantic search
type QueryResponse struct {
	Query   string         `json:"query"`
	Results []QueryResult  `json:"results"`
	Count   int            `json:"count"`
	Duration string        `json:"duration"`
}

// Entity represents a knowledge entity
type Entity struct {
	ID               string `json:"id"`
	Name             string `json:"name"`
	Type             string `json:"type"`
	Description      string `json:"description"`
	FirstMentionedAt time.Time `json:"first_mentioned_at"`
	LastMentionedAt  time.Time `json:"last_mentioned_at"`
	MentionCount     int    `json:"mention_count"`
}

// HistoryEvent represents a past event
type HistoryEvent struct {
	ID               string    `json:"id"`
	Source           string    `json:"source"`
	EventType        string    `json:"event_type"`
	Author           string    `json:"author"`
	Channel          string    `json:"channel"`
	ProcessingStatus string    `json:"processing_status"`
	ReceivedAt       time.Time `json:"received_at"`
	ProcessedAt      *time.Time `json:"processed_at"`
}

// HistoryResponse represents event history
type HistoryResponse struct {
	Events []HistoryEvent `json:"events"`
	Total  int            `json:"total"`
	Offset int            `json:"offset"`
	Limit  int            `json:"limit"`
}

// StatusResponse represents service status
type StatusResponse struct {
	Status string      `json:"status"`
	Uptime int64       `json:"uptime_seconds"`
	DB     DBStats     `json:"database"`
	Queue  QueueStats  `json:"queue"`
}

// DBStats represents database statistics
type DBStats struct {
	TotalEvents     int `json:"total_events"`
	ProcessedEvents int `json:"processed_events"`
	TotalKnowledge  int `json:"total_knowledge"`
	WithEmbeddings  int `json:"with_embeddings"`
}

// QueueStats represents queue statistics
type QueueStats struct {
	PendingEvents int `json:"pending_events"`
	FailedEvents  int `json:"failed_events"`
}
