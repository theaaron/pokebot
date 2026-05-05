// backend/persist.go
package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// ── on-disk schema ────────────────────────────────────────────────────

type persistData struct {
	Version       int           `json:"version"`
	Conversations []persistConv `json:"conversations"`
}

type persistConv struct {
	ID            string           `json:"id"`
	Title         string           `json:"title"`
	StartedAt     time.Time        `json:"started_at"`
	Messages      []ChatMessage    `json:"messages"`
	GeminiHistory []map[string]any `json:"gemini_history"`
	DetectedTypes []string         `json:"detected_types"`
}

// ── Store ─────────────────────────────────────────────────────────────

type Store struct {
	mu            sync.RWMutex
	conversations []*Conversation // index 0 = newest
}

func NewStore() *Store {
	s := &Store{}
	s.load()
	if len(s.conversations) == 0 {
		s.conversations = append(s.conversations, newConversation())
	}
	return s
}

// ── read ops ──────────────────────────────────────────────────────────

func (s *Store) List() []ConvSummary {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]ConvSummary, len(s.conversations))
	for i, c := range s.conversations {
		out[i] = toSummary(c)
	}
	return out
}

func (s *Store) Get(id string) (*Conversation, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, c := range s.conversations {
		if c.ID == id {
			return c, true
		}
	}
	return nil, false
}

// ── write ops ─────────────────────────────────────────────────────────

func (s *Store) Create() *Conversation {
	s.mu.Lock()
	defer s.mu.Unlock()
	c := newConversation()
	s.conversations = append([]*Conversation{c}, s.conversations...)
	s.save()
	return c
}

func (s *Store) Delete(id string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	for i, c := range s.conversations {
		if c.ID == id {
			s.conversations = append(s.conversations[:i], s.conversations[i+1:]...)
			if len(s.conversations) == 0 {
				s.conversations = append(s.conversations, newConversation())
			}
			s.save()
			return true
		}
	}
	return false
}

// called by agent after the loop completes (lock held externally via SaveAfterChat)
func (s *Store) SaveAfterChat(conv *Conversation) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.save()
	_ = conv // conv is already a pointer in the slice
}

// ── file I/O ──────────────────────────────────────────────────────────

func dataFilePath() string {
	dir, err := os.UserConfigDir()
	if err != nil {
		dir = "."
	}
	return filepath.Join(dir, "pokebot", "history.json")
}

func (s *Store) save() {
	pd := persistData{Version: 1}
	for _, c := range s.conversations {
		pd.Conversations = append(pd.Conversations, persistConv{
			ID:            c.ID,
			Title:         c.Title,
			StartedAt:     c.StartedAt,
			Messages:      c.Messages,
			GeminiHistory: c.GeminiHistory,
			DetectedTypes: c.DetectedTypes,
		})
	}
	data, err := json.MarshalIndent(pd, "", "  ")
	if err != nil {
		return
	}
	path := dataFilePath()
	_ = os.MkdirAll(filepath.Dir(path), 0o755)
	_ = os.WriteFile(path, data, 0o644)
}

func (s *Store) load() {
	data, err := os.ReadFile(dataFilePath())
	if err != nil {
		return
	}
	var pd persistData
	if err := json.Unmarshal(data, &pd); err != nil {
		return
	}
	for _, pc := range pd.Conversations {
		s.conversations = append(s.conversations, &Conversation{
			ID:            pc.ID,
			Title:         pc.Title,
			StartedAt:     pc.StartedAt,
			Messages:      pc.Messages,
			GeminiHistory: pc.GeminiHistory,
			DetectedTypes: pc.DetectedTypes,
		})
	}
}
