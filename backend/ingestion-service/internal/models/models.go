package models

import (
	"time"
)

// EventSource represents the source of the event
type EventSource string

const (
	SourceSlack  EventSource = "slack"
	SourceGitHub EventSource = "github"
)

// EventType represents the type of event
type EventType string

const (
	EventMessage   EventType = "message"
	EventPRCreated EventType = "pr_created"
	EventPRUpdated EventType = "pr_updated"
	EventCommit    EventType = "commit"
)

// ProcessingStatus represents the processing status
type ProcessingStatus string

const (
	StatusPending    ProcessingStatus = "pending"
	StatusProcessing ProcessingStatus = "processing"
	StatusCompleted  ProcessingStatus = "completed"
	StatusFailed     ProcessingStatus = "failed"
)

// Event represents a raw event from a source
type Event struct {
	ID               string                 `json:"id" db:"id"`
	Source           EventSource            `json:"source" db:"source"`
	SourceID         string                 `json:"source_id" db:"source_id"`
	EventType        EventType              `json:"event_type" db:"event_type"`
	RawData          map[string]interface{} `json:"raw_data" db:"raw_data"`
	Author           string                 `json:"author" db:"author"`
	Channel          string                 `json:"channel" db:"channel"`
	ReceivedAt       time.Time              `json:"received_at" db:"received_at"`
	ProcessingStatus ProcessingStatus       `json:"processing_status" db:"processing_status"`
	ProcessedAt      *time.Time             `json:"processed_at" db:"processed_at"`
	ErrorMessage     string                 `json:"error_message" db:"error_message"`
	ErrorCount       int                    `json:"error_count" db:"error_count"`
	LastErrorAt      *time.Time             `json:"last_error_at" db:"last_error_at"`
}

// SlackEvent represents a Slack webhook event
type SlackEvent struct {
	Token     string         `json:"token"`
	TeamID    string         `json:"team_id"`
	APIAPPID  string         `json:"api_app_id"`
	Event     SlackEventData `json:"event"`
	Type      string         `json:"type"`
	EventID   string         `json:"event_id"`
	EventTime int64          `json:"event_time"`
}

type SlackEventData struct {
	Type      string `json:"type"`
	User      string `json:"user"`
	Text      string `json:"text"`
	Timestamp string `json:"ts"`
	Channel   string `json:"channel"`
	EventTs   string `json:"event_ts"`
	ThreadTs  string `json:"thread_ts,omitempty"`
}

// GitHubEvent represents a GitHub webhook event
type GitHubEvent struct {
	Action      string       `json:"action"`
	Number      int          `json:"number"`
	PullRequest *PullRequest `json:"pull_request,omitempty"`
	Repository  Repository   `json:"repository"`
	Sender      User         `json:"sender"`
	Push        *PushEvent   `json:"push,omitempty"`
}

type PullRequest struct {
	ID     int    `json:"id"`
	Number int    `json:"number"`
	Title  string `json:"title"`
	Body   string `json:"body"`
	User   User   `json:"user"`
	State  string `json:"state"`
	URL    string `json:"html_url"`
	Diff   string `json:"diff_url"`
}

type PushEvent struct {
	Ref        string   `json:"ref"`
	Commits    []Commit `json:"commits"`
	HeadCommit *Commit  `json:"head_commit"`
}

type Commit struct {
	ID        string `json:"id"`
	Message   string `json:"message"`
	Timestamp string `json:"timestamp"`
	Author    User   `json:"author"`
	URL       string `json:"url"`
	Diff      string `json:"diff,omitempty"`
}

type Repository struct {
	ID       int    `json:"id"`
	Name     string `json:"name"`
	FullName string `json:"full_name"`
	URL      string `json:"html_url"`
}

type User struct {
	Login string `json:"login"`
	ID    int    `json:"id"`
	URL   string `json:"html_url"`
	Email string `json:"email"`
}

// QueueMessage represents a message published to the queue (Redis)
type QueueMessage struct {
	EventID   string                 `json:"event_id"`
	Source    EventSource            `json:"source"`
	EventType EventType              `json:"event_type"`
	Timestamp time.Time              `json:"timestamp"`
	Data      map[string]interface{} `json:"data"`
}

// KafkaMessage represents a message published to Kafka
type KafkaMessage struct {
	EventID       string                 `json:"event_id"`
	Source        EventSource            `json:"source"`
	EventType     EventType              `json:"event_type"`
	Timestamp     time.Time              `json:"timestamp"`
	Data          map[string]interface{} `json:"data"`
	CorrelationID string                 `json:"correlation_id"`
	Version       string                 `json:"version"`
}

// ValidationResult represents the result of event validation
type ValidationResult struct {
	Valid    bool
	Errors   []string
	Warnings []string
}
