# PokéBot

An agentic Pokémon assistant for your terminal. Ask anything — types, moves, evolutions, locations, items, natures — and get answers backed by live PokéAPI data.

Built with Go, [Charm](https://charm.sh) TUI libraries, and Google Gemini.

---

## How it works

The backend runs an agentic loop: Gemini decides which tools to call, fetches real data from PokéAPI, and streams the response back to the terminal client in real time.

```
┌──────────────────┐        HTTP/SSE        ┌──────────────────────┐
│   CLI (TUI)      │ ──────────────────────▶ │   Backend API        │
│  Bubble Tea UI   │                         │  Go HTTP + Gemini    │
│  Glamour markdown│ ◀────────────────────── │  PokeAPI client      │
└──────────────────┘    streaming events     └──────────────────────┘
```

- **`/backend`** — REST/SSE API server
- **`/cli`** — Terminal client

---

## Requirements

- Go 1.22+
- Gemini API key — free at [aistudio.google.com](https://aistudio.google.com/app/apikeys)

---

## Running

**Start the backend:**
```bash
cd backend
export GEMINI_API_KEY="your-api-key-here"
go run .
```

**Start the CLI (in a new terminal):**
```bash
cd cli
go run .
```

---

## Controls

| Key | Action |
|---|---|
| `Tab` / `Shift+Tab` | Switch between chat and sidebar |
| `↑` / `↓` | Navigate conversations |
| `Enter` | Send message / select conversation |
| `Delete` | Delete conversation |
| `?` | Help |
| `Ctrl+C` | Quit |

---

## Tech Stack

**Backend** — Go, [Gemini REST API](https://ai.google.dev/api/generate-content), [PokéAPI v2](https://pokeapi.co/)

**CLI** — [Bubble Tea](https://github.com/charmbracelet/bubbletea), [Lip Gloss](https://github.com/charmbracelet/lipgloss), [Glamour](https://github.com/charmbracelet/glamour) via [charm.land](https://charm.sh)
