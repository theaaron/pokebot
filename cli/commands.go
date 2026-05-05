// cli/commands.go
package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

// ── shared types that mirror the backend JSON ────────────────────────

type ConvSummary struct {
	ID            string    `json:"id"`
	Title         string    `json:"title"`
	StartedAt     time.Time `json:"started_at"`
	MessageCount  int       `json:"message_count"`
	DetectedTypes []string  `json:"detected_types"`
}

type ChatMessage struct {
	Role      string         `json:"role"`
	Text      string         `json:"text,omitempty"`
	ToolName  string         `json:"tool_name,omitempty"`
	ToolArgs  map[string]any `json:"tool_args,omitempty"`
	Types     []string       `json:"types,omitempty"`
	Timestamp time.Time      `json:"timestamp"`
}

type ConvFull struct {
	ID            string        `json:"id"`
	Title         string        `json:"title"`
	StartedAt     time.Time     `json:"started_at"`
	Messages      []ChatMessage `json:"messages"`
	DetectedTypes []string      `json:"detected_types"`
}

// ── SSE event parsed from stream ──────────────────────────────────────

type sseEvent struct {
	Type string
	Data map[string]any
}

// ── tea messages ──────────────────────────────────────────────────────

type convListMsg struct{ convs []ConvSummary }
type convLoadedMsg struct {
	conv ConvFull
	err  error
}
type convCreatedMsg struct {
	conv ConvFull
	err  error
}
type convDeletedMsg struct{ err error }
type chatEventMsg struct{ ev sseEvent }
type chatDoneMsg struct{ err error }
type fetchErrMsg struct{ err error }

// ── HTTP helpers ──────────────────────────────────────────────────────

var httpClient = &http.Client{Timeout: 10 * time.Second}

func fetchJSON(url string, out any) error {
	resp, err := httpClient.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return fmt.Errorf("HTTP %d", resp.StatusCode)
	}
	return json.NewDecoder(resp.Body).Decode(out)
}

func postJSON(url string, body any, out any) error {
	data, _ := json.Marshal(body)
	resp, err := httpClient.Post(url, "application/json", bytes.NewReader(data))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		b, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(b))
	}
	return json.NewDecoder(resp.Body).Decode(out)
}

func deleteReq(url string) error {
	req, _ := http.NewRequest("DELETE", url, nil)
	resp, err := httpClient.Do(req)
	if err != nil {
		return err
	}
	resp.Body.Close()
	if resp.StatusCode >= 400 {
		return fmt.Errorf("HTTP %d", resp.StatusCode)
	}
	return nil
}

// ── tea.Cmd builders ──────────────────────────────────────────────────

func loadConvListCmd(apiURL string) tea.Cmd {
	return func() tea.Msg {
		var list []ConvSummary
		if err := fetchJSON(apiURL+"/api/conversations", &list); err != nil {
			return fetchErrMsg{err: err}
		}
		return convListMsg{convs: list}
	}
}

func loadConvCmd(apiURL, id string) tea.Cmd {
	return func() tea.Msg {
		var c ConvFull
		if err := fetchJSON(apiURL+"/api/conversations/"+id, &c); err != nil {
			return convLoadedMsg{err: err}
		}
		return convLoadedMsg{conv: c}
	}
}

func createConvCmd(apiURL string) tea.Cmd {
	return func() tea.Msg {
		var c ConvFull
		resp, err := httpClient.Post(apiURL+"/api/conversations", "application/json", nil)
		if err != nil {
			return convCreatedMsg{err: err}
		}
		defer resp.Body.Close()
		if err := json.NewDecoder(resp.Body).Decode(&c); err != nil {
			return convCreatedMsg{err: err}
		}
		return convCreatedMsg{conv: c}
	}
}

func deleteConvCmd(apiURL, id string) tea.Cmd {
	return func() tea.Msg {
		return convDeletedMsg{err: deleteReq(apiURL + "/api/conversations/" + id)}
	}
}

// ── SSE chat stream ──────────────────────────────────────────────────
// Opens the SSE connection, parses every event, and sends each one as a
// chatEventMsg.  A final chatDoneMsg is always sent.

func chatStreamCmd(apiURL, convID, message string) tea.Cmd {
	return func() tea.Msg {
		// We return a *batch* of the first event + a continuation command.
		// This keeps the tea loop responsive between events.
		events, err := openChatStream(apiURL, convID, message)
		if err != nil {
			return chatDoneMsg{err: err}
		}
		// Drain all events eagerly (the agentic loop is server-side).
		// Return the collected events as a single batch message so the
		// model can process them sequentially without extra plumbing.
		return chatEventsMsg{events: events}
	}
}

type chatEventsMsg struct{ events []sseEvent }

func openChatStream(apiURL, convID, message string) ([]sseEvent, error) {
	body, _ := json.Marshal(map[string]string{"message": message})
	resp, err := (&http.Client{Timeout: 120 * time.Second}).Post(
		apiURL+"/api/conversations/"+convID+"/chat",
		"application/json",
		bytes.NewReader(body),
	)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		b, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("chat HTTP %d: %s", resp.StatusCode, string(b))
	}

	var collected []sseEvent
	scanner := bufio.NewScanner(resp.Body)
	var eventType, dataStr string

	for scanner.Scan() {
		line := scanner.Text()

		switch {
		case strings.HasPrefix(line, "event:"):
			eventType = strings.TrimSpace(strings.TrimPrefix(line, "event:"))
		case strings.HasPrefix(line, "data:"):
			dataStr = strings.TrimSpace(strings.TrimPrefix(line, "data:"))
		case line == "":
			// end of one SSE event
			if eventType != "" && dataStr != "" {
				var d map[string]any
				if err := json.Unmarshal([]byte(dataStr), &d); err == nil {
					collected = append(collected, sseEvent{Type: eventType, Data: d})
				}
			}
			eventType, dataStr = "", ""
		}
	}
	return collected, scanner.Err()
}
