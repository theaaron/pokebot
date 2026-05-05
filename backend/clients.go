// backend/clients.go
package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

// ======================================================================
// Gemini Client
// ======================================================================

type GeminiClient struct {
	apiKey string
	model  string
	client *http.Client
}

func newGeminiClient(apiKey string) *GeminiClient {
	return &GeminiClient{
		apiKey: apiKey,
		model:  "gemini-flash-lite-latest",
		client: &http.Client{Timeout: 90 * time.Second},
	}
}

func (g *GeminiClient) Generate(
	contents []map[string]any,
	tools []map[string]any,
	systemInstruction map[string]any,
) (map[string]any, error) {
	endpoint := fmt.Sprintf(
		"https://generativelanguage.googleapis.com/v1beta/models/%s:generateContent?key=%s",
		g.model, g.apiKey,
	)

	body := map[string]any{"contents": contents, "tools": tools}
	if systemInstruction != nil {
		body["system_instruction"] = systemInstruction
	}

	jsonData, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("marshal: %w", err)
	}

	req, err := http.NewRequest("POST", endpoint, strings.NewReader(string(jsonData)))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := g.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, trunc(string(data), 300))
	}

	var result map[string]any
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("unmarshal: %w", err)
	}
	if errObj, ok := result["error"].(map[string]any); ok {
		if msg, ok := errObj["message"].(string); ok {
			return nil, fmt.Errorf("gemini: %s", msg)
		}
	}
	return result, nil
}

// ======================================================================
// PokeAPI Client
// ======================================================================

type PokeAPIClient struct {
	baseURL string
	client  *http.Client
	cache   sync.Map
}

func newPokeAPIClient() *PokeAPIClient {
	return &PokeAPIClient{
		baseURL: "https://pokeapi.co/api/v2",
		client:  &http.Client{Timeout: 15 * time.Second},
	}
}

func (p *PokeAPIClient) Get(path string, params map[string]string) (map[string]any, error) {
	cacheKey := path + fmt.Sprint(params)
	if v, ok := p.cache.Load(cacheKey); ok {
		return v.(map[string]any), nil
	}

	u, _ := url.Parse(fmt.Sprintf("%s/%s", p.baseURL, path))
	if params != nil {
		q := u.Query()
		for k, v := range params {
			q.Set(k, v)
		}
		u.RawQuery = q.Encode()
	}

	data, err := p.fetch(u.String())
	if err != nil {
		return nil, err
	}
	var result map[string]any
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("parse: %w", err)
	}
	p.cache.Store(cacheKey, result)
	return result, nil
}

func (p *PokeAPIClient) GetURL(rawURL string) (map[string]any, error) {
	if v, ok := p.cache.Load(rawURL); ok {
		return v.(map[string]any), nil
	}
	data, err := p.fetch(rawURL)
	if err != nil {
		return nil, err
	}
	var result map[string]any
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("parse: %w", err)
	}
	p.cache.Store(rawURL, result)
	return result, nil
}

func (p *PokeAPIClient) GetArray(rawURL string) ([]map[string]any, error) {
	if v, ok := p.cache.Load(rawURL); ok {
		return v.([]map[string]any), nil
	}
	data, err := p.fetch(rawURL)
	if err != nil {
		return nil, err
	}
	var raw []any
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("parse array: %w", err)
	}
	var result []map[string]any
	for _, item := range raw {
		if m, ok := item.(map[string]any); ok {
			result = append(result, m)
		}
	}
	p.cache.Store(rawURL, result)
	return result, nil
}

func (p *PokeAPIClient) fetch(rawURL string) ([]byte, error) {
	resp, err := p.client.Get(rawURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode == 404 {
		return nil, fmt.Errorf("not found: %s", rawURL)
	}
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("HTTP %d from PokeAPI", resp.StatusCode)
	}
	return data, nil
}

// ── utility ───────────────────────────────────────────────────────────

func trunc(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "…"
}
