// backend/state.go
package main

import (
	"fmt"
	"strings"
	"time"
)

// ── Message roles ─────────────────────────────────────────────────────

type MessageRole string

const (
	RoleUser       MessageRole = "user"
	RoleBot        MessageRole = "bot"
	RoleToolCall   MessageRole = "tool_call"
	RoleToolResult MessageRole = "tool_result"
	RoleError      MessageRole = "error"
)

// ── Chat message ──────────────────────────────────────────────────────

type ChatMessage struct {
	Role      MessageRole    `json:"role"`
	Text      string         `json:"text,omitempty"`
	ToolName  string         `json:"tool_name,omitempty"`
	ToolArgs  map[string]any `json:"tool_args,omitempty"`
	Types     []string       `json:"types,omitempty"`
	Timestamp time.Time      `json:"timestamp"`
}

// ── Conversation ──────────────────────────────────────────────────────

type Conversation struct {
	ID            string
	Title         string
	StartedAt     time.Time
	Messages      []ChatMessage
	GeminiHistory []map[string]any // kept in memory / persisted, not sent to API
	DetectedTypes []string
}

func newConversation() *Conversation {
	return &Conversation{
		ID:        fmt.Sprintf("%08x", time.Now().UnixNano()%0xFFFFFFFF),
		StartedAt: time.Now(),
		Title:     "New conversation",
	}
}

// ── API response shapes ──────────────────────────────────────────────

type ConvSummary struct {
	ID            string    `json:"id"`
	Title         string    `json:"title"`
	StartedAt     time.Time `json:"started_at"`
	MessageCount  int       `json:"message_count"`
	DetectedTypes []string  `json:"detected_types"`
}

type ConvFull struct {
	ID            string        `json:"id"`
	Title         string        `json:"title"`
	StartedAt     time.Time     `json:"started_at"`
	Messages      []ChatMessage `json:"messages"`
	DetectedTypes []string      `json:"detected_types"`
}

func toSummary(c *Conversation) ConvSummary {
	return ConvSummary{
		ID:            c.ID,
		Title:         c.Title,
		StartedAt:     c.StartedAt,
		MessageCount:  len(c.Messages),
		DetectedTypes: c.DetectedTypes,
	}
}

func toFull(c *Conversation) ConvFull {
	return ConvFull{
		ID:            c.ID,
		Title:         c.Title,
		StartedAt:     c.StartedAt,
		Messages:      c.Messages,
		DetectedTypes: c.DetectedTypes,
	}
}

// ── SSE event types ──────────────────────────────────────────────────

type SSEEventType string

const (
	EventStatus     SSEEventType = "status"
	EventToolCall   SSEEventType = "tool_call"
	EventToolResult SSEEventType = "tool_result"
	EventBotMsg     SSEEventType = "bot_message"
	EventError      SSEEventType = "error"
	EventDone       SSEEventType = "done"
)

type SSEEvent struct {
	Type SSEEventType
	Data map[string]any
}

// ── helpers ───────────────────────────────────────────────────────────

func norm(s string) string {
	return strings.ToLower(strings.TrimSpace(strings.ReplaceAll(s, " ", "-")))
}

func strArg(args map[string]any, key string) string {
	if v, ok := args[key]; ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}
