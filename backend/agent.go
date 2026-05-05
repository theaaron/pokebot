// backend/agent.go
package main

import (
	"fmt"
	"strings"
	"time"
)

const maxAgentIter = 10

const systemPrompt = `You are PokéBot, an expert Pokémon assistant with live data from PokeAPI.

Rules:
- Always use tools to get real data. Never make up facts.
- Be friendly and enthusiastic about Pokémon!
- Use markdown formatting in your responses: headers, bold, bullet lists, code blocks.
- If the user asks about a specific game version, highlight relevant results.
- Format responses clearly with sections and bullet points.
- You can reference earlier messages — full chat history is available.
- If a Pokémon is not found, suggest checking spelling or use search_pokemon.
- You may call multiple tools to answer complex questions.`

// RunAgentLoop executes the full agentic loop for one user message.
// Events are pushed onto the channel; the caller must drain it.
// The conversation's Messages and GeminiHistory are mutated in place.
func (app *App) RunAgentLoop(conv *Conversation, userText string, out chan<- SSEEvent) {
	// ── add user message to both stores ─────────────────────────────
	conv.Messages = append(conv.Messages, ChatMessage{
		Role:      RoleUser,
		Text:      userText,
		Timestamp: time.Now(),
	})
	if conv.Title == "New conversation" {
		t := userText
		if len(t) > 45 {
			t = t[:45] + "…"
		}
		conv.Title = t
	}

	conv.GeminiHistory = append(conv.GeminiHistory, map[string]any{
		"role":  "user",
		"parts": []map[string]any{{"text": userText}},
	})

	sysPayload := map[string]any{
		"parts": []map[string]any{{"text": systemPrompt}},
	}

	// ── agentic loop ─────────────────────────────────────────────────
	for iter := 0; iter < maxAgentIter; iter++ {
		out <- SSEEvent{Type: EventStatus, Data: map[string]any{"text": "Thinking…"}}

		resp, err := app.gemini.Generate(conv.GeminiHistory, app.tools.GeminiPayload(), sysPayload)
		if err != nil {
			pushError(conv, out, err.Error())
			return
		}

		// ── parse response parts ───────────────────────────────────
		rawParts, err := extractParts(resp)
		if err != nil {
			pushError(conv, out, err.Error())
			return
		}
		conv.GeminiHistory = append(conv.GeminiHistory, map[string]any{
			"role": "model", "parts": rawParts,
		})

		var fCalls []FunctionCall
		var texts []string
		for _, pm := range rawParts {
			if fc, ok := pm["functionCall"].(map[string]any); ok {
				name, _ := fc["name"].(string)
				args, _ := fc["args"].(map[string]any)
				if args == nil {
					args = map[string]any{}
				}
				fCalls = append(fCalls, FunctionCall{Name: name, Args: args})
			}
			if t, ok := pm["text"].(string); ok && strings.TrimSpace(t) != "" {
				texts = append(texts, t)
			}
		}

		// ── text-only → final response ─────────────────────────────
		if len(texts) > 0 && len(fCalls) == 0 {
			types := app.tools.LastDetectedTypes
			if len(types) == 0 {
				types = conv.DetectedTypes
			}
			msg := ChatMessage{
				Role:      RoleBot,
				Text:      strings.Join(texts, "\n"),
				Types:     types,
				Timestamp: time.Now(),
			}
			conv.Messages = append(conv.Messages, msg)
			out <- SSEEvent{Type: EventBotMsg, Data: map[string]any{
				"text": msg.Text, "types": types,
			}}
			out <- SSEEvent{Type: EventDone, Data: map[string]any{}}
			return
		}

		// ── tool calls ─────────────────────────────────────────────
		if len(fCalls) > 0 {
			var responseParts []map[string]any

			for _, fc := range fCalls {
				// record tool_call message
				conv.Messages = append(conv.Messages, ChatMessage{
					Role:      RoleToolCall,
					ToolName:  fc.Name,
					ToolArgs:  fc.Args,
					Timestamp: time.Now(),
				})
				out <- SSEEvent{Type: EventToolCall, Data: map[string]any{
					"name": fc.Name, "args": fc.Args,
				}}

				out <- SSEEvent{Type: EventStatus, Data: map[string]any{
					"text": fmt.Sprintf("Running %s…", fc.Name),
				}}

				// execute
				result := app.tools.Call(fc.Name, fc.Args)

				// record tool_result message
				conv.Messages = append(conv.Messages, ChatMessage{
					Role:      RoleToolResult,
					ToolName:  fc.Name,
					Timestamp: time.Now(),
				})
				if len(app.tools.LastDetectedTypes) > 0 {
					addDetectedTypes(conv, app.tools.LastDetectedTypes)
				}
				out <- SSEEvent{Type: EventToolResult, Data: map[string]any{
					"name": fc.Name,
				}}

				responseParts = append(responseParts, map[string]any{
					"functionResponse": map[string]any{"name": fc.Name, "response": result},
				})
			}

			conv.GeminiHistory = append(conv.GeminiHistory, map[string]any{
				"role": "user", "parts": responseParts,
			})
			continue // next iteration
		}

		// ── empty response ─────────────────────────────────────────
		pushError(conv, out, "Empty response from Gemini.")
		return
	}

	// exceeded max iterations
	pushError(conv, out, "Exceeded maximum agent iterations.")
}

// ── helpers ───────────────────────────────────────────────────────────

type FunctionCall struct {
	Name string
	Args map[string]any
}

func extractParts(resp map[string]any) ([]map[string]any, error) {
	candidates, ok := resp["candidates"].([]any)
	if !ok || len(candidates) == 0 {
		return nil, fmt.Errorf("no candidates in response")
	}
	first, _ := candidates[0].(map[string]any)
	content, ok := first["content"].(map[string]any)
	if !ok {
		return nil, fmt.Errorf("no content in candidate")
	}
	parts, ok := content["parts"].([]any)
	if !ok {
		return nil, fmt.Errorf("no parts in content")
	}
	var raw []map[string]any
	for _, p := range parts {
		if pm, ok := p.(map[string]any); ok {
			raw = append(raw, pm)
		}
	}
	return raw, nil
}

func pushError(conv *Conversation, out chan<- SSEEvent, msg string) {
	conv.Messages = append(conv.Messages, ChatMessage{
		Role:      RoleError,
		Text:      msg,
		Timestamp: time.Now(),
	})
	out <- SSEEvent{Type: EventError, Data: map[string]any{"text": msg}}
	out <- SSEEvent{Type: EventDone, Data: map[string]any{}}
}

func addDetectedTypes(conv *Conversation, types []string) {
	for _, t := range types {
		found := false
		for _, e := range conv.DetectedTypes {
			if e == t {
				found = true
				break
			}
		}
		if !found {
			conv.DetectedTypes = append(conv.DetectedTypes, t)
		}
	}
}
