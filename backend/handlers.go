// backend/handlers.go
package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// ── App ───────────────────────────────────────────────────────────────

type App struct {
	gemini *GeminiClient
	tools  *ToolRegistry
	store  *Store
}

func NewApp(apiKey string) *App {
	return &App{
		gemini: newGeminiClient(apiKey),
		tools:  newToolRegistry(),
		store:  NewStore(),
	}
}

// ── helpers ───────────────────────────────────────────────────────────

func writeJSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(v)
}

func readBody(r *http.Request) (map[string]any, error) {
	data, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}
	var m map[string]any
	if err := json.Unmarshal(data, &m); err != nil {
		return nil, err
	}
	return m, nil
}

func writeSSE(w http.ResponseWriter, ev SSEEvent) {
	data, _ := json.Marshal(ev.Data)
	fmt.Fprintf(w, "event: %s\ndata: %s\n\n", ev.Type, string(data))
	if f, ok := w.(http.Flusher); ok {
		f.Flush()
	}
}

// ── GET /health ───────────────────────────────────────────────────────

func (app *App) health(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{"status": "ok"})
}

// ── GET /api/conversations ────────────────────────────────────────────

func (app *App) listConversations(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, app.store.List())
}

// ── POST /api/conversations ───────────────────────────────────────────

func (app *App) createConversation(w http.ResponseWriter, _ *http.Request) {
	c := app.store.Create()
	writeJSON(w, http.StatusCreated, toFull(c))
}

// ── GET /api/conversations/{id} ───────────────────────────────────────

func (app *App) getConversation(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	c, ok := app.store.Get(id)
	if !ok {
		writeJSON(w, http.StatusNotFound, map[string]any{"error": "not found"})
		return
	}
	writeJSON(w, http.StatusOK, toFull(c))
}

// ── DELETE /api/conversations/{id} ────────────────────────────────────

func (app *App) deleteConversation(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if app.store.Delete(id) {
		w.WriteHeader(http.StatusNoContent)
		return
	}
	writeJSON(w, http.StatusNotFound, map[string]any{"error": "not found"})
}

// ── POST /api/conversations/{id}/chat  (SSE) ─────────────────────────

func (app *App) chatStream(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	conv, ok := app.store.Get(id)
	if !ok {
		writeJSON(w, http.StatusNotFound, map[string]any{"error": "not found"})
		return
	}

	body, err := readBody(r)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": "bad body"})
		return
	}
	msg, _ := body["message"].(string)
	if msg == "" {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": "message is required"})
		return
	}

	// ── set up SSE response ──────────────────────────────────────
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.WriteHeader(http.StatusOK)
	if f, ok := w.(http.Flusher); ok {
		f.Flush()
	}

	// ── run agent, stream events ─────────────────────────────────
	ch := make(chan SSEEvent, 32)
	go func() {
		defer close(ch)
		app.RunAgentLoop(conv, msg, ch)
	}()

	for ev := range ch {
		writeSSE(w, ev)
	}

	// persist after the loop
	app.store.SaveAfterChat(conv)
}
