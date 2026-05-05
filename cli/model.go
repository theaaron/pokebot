// cli/model.go
package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const (
	sidebarMinW    = 28
	sidebarMaxW    = 38
	splashDuration = 60
)

// ── app state ─────────────────────────────────────────────────────────

type appState int

const (
	stateSplash appState = iota
	stateMain
)

type focusState int

const (
	focusChat focusState = iota
	focusSidebar
)

var splashPhases = []string{
	"Connecting to PokéBot API…",
	"Loading conversations…",
	"Preparing your adventure…",
	"Ready!",
}

type splashTickMsg time.Time

func splashTick() tea.Cmd {
	return tea.Tick(50*time.Millisecond, func(t time.Time) tea.Msg {
		return splashTickMsg(t)
	})
}

// ── model ─────────────────────────────────────────────────────────────

type model struct {
	width, height int
	state         appState
	focus         focusState
	sidebarCursor int

	splashProgress float64
	splashTick     int
	progressBar    progress.Model

	viewport  viewport.Model
	textInput textinput.Model
	spinner   spinner.Model
	help      help.Model
	keys      keyMap
	showHelp  bool

	// backend state
	apiURL     string
	convs      []ConvSummary // sidebar list
	active     *ConvFull     // currently displayed conversation
	loading    bool
	statusText string
	toolLog    []string
	initErr    string
}

func newModel(apiURL string) model {
	ti := textinput.New()
	ti.Placeholder = "Ask anything about Pokémon…"
	ti.CharLimit = 600
	ti.PromptStyle = inputPromptStyle
	ti.Prompt = "⚡ "
	ti.TextStyle = lipgloss.NewStyle().Foreground(colText)
	ti.PlaceholderStyle = lipgloss.NewStyle().Foreground(colTextDim).Faint(true)
	ti.Cursor.Style = lipgloss.NewStyle().Foreground(colAccent)
	ti.Blur()

	sp := spinner.New()
	sp.Spinner = spinner.Dot
	sp.Style = spinnerStyle

	pb := progress.New(
		progress.WithGradient("#42a5f5", "#e0af68"),
		progress.WithoutPercentage(),
	)
	pb.Width = 40

	vp := viewport.New(80, 20)
	vp.MouseWheelEnabled = true
	vp.MouseWheelDelta = 3

	return model{
		state:       stateSplash,
		viewport:    vp,
		textInput:   ti,
		spinner:     sp,
		progressBar: pb,
		help:        newHelp(),
		keys:        defaultKeys(),
		apiURL:      apiURL,
	}
}

func (m model) Init() tea.Cmd {
	return tea.Batch(splashTick(), m.spinner.Tick, textinput.Blink, loadConvListCmd(m.apiURL))
}

// ── Update ────────────────────────────────────────────────────────────

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.progressBar.Width = clamp(m.width/3, 20, 50)
		if m.state == stateMain {
			m.recalcLayout()
			m.refreshViewport()
		}
		return m, nil

	case splashTickMsg:
		if m.state != stateSplash {
			return m, nil
		}
		m.splashTick++
		m.splashProgress = float64(m.splashTick) / float64(splashDuration)
		if m.splashProgress > 1.0 {
			m.splashProgress = 1.0
		}
		if m.splashTick >= splashDuration {
			return m.transitionToMain()
		}
		return m, splashTick()

	// ── backend responses ──────────────────────────────────────────
	case convListMsg:
		m.convs = msg.convs
		if len(m.convs) > 0 && m.active == nil {
			return m, loadConvCmd(m.apiURL, m.convs[0].ID)
		}
		return m, nil

	case fetchErrMsg:
		m.initErr = msg.err.Error()
		if m.state == stateSplash {
			return m.transitionToMain()
		}
		return m, nil

	case convLoadedMsg:
		if msg.err != nil {
			m.initErr = msg.err.Error()
		} else {
			m.active = &msg.conv
			m.refreshViewport()
		}
		if m.state == stateSplash {
			return m.transitionToMain()
		}
		return m, nil

	case convCreatedMsg:
		if msg.err == nil {
			m.active = &msg.conv
			m.convs = append([]ConvSummary{{
				ID:        msg.conv.ID,
				Title:     msg.conv.Title,
				StartedAt: msg.conv.StartedAt,
			}}, m.convs...)
			m.sidebarCursor = 0
			m.refreshViewport()
		}
		return m, nil

	case convDeletedMsg:
		if msg.err == nil {
			return m, loadConvListCmd(m.apiURL)
		}
		return m, nil

	case chatEventsMsg:
		return m.applyChatEvents(msg.events)

	case chatDoneMsg:
		m.loading = false
		m.statusText = ""
		m.toolLog = nil
		if msg.err != nil {
			m.appendMessage(ChatMessage{Role: "error", Text: msg.err.Error()})
		}
		m.refreshViewport()
		m.textInput.Focus()
		return m, nil

	// ── UI events ──────────────────────────────────────────────────
	case tea.KeyMsg:
		if m.state == stateSplash {
			return m.transitionToMain()
		}
		switch {
		case msg.String() == "ctrl+c":
			return m, tea.Quit
		case msg.String() == "tab":
			if m.focus == focusChat {
				m.focus = focusSidebar
				m.textInput.Blur()
			} else {
				m.focus = focusChat
				if !m.loading {
					m.textInput.Focus()
				}
			}
			return m, nil
		case msg.String() == "ctrl+n":
			return m, createConvCmd(m.apiURL)
		case msg.String() == "?":
			if m.focus == focusSidebar || m.textInput.Value() == "" {
				m.showHelp = !m.showHelp
				return m, nil
			}
		case msg.String() == "esc":
			if m.showHelp {
				m.showHelp = false
				return m, nil
			}
		}
		if m.focus == focusSidebar {
			return m.updateSidebar(msg)
		}
		return m.updateChat(msg)

	case tea.MouseMsg:
		if m.state == stateSplash {
			return m.transitionToMain()
		}
		sw := m.sidebarWidth()
		switch msg.Type {
		case tea.MouseWheelUp:
			if msg.X <= sw {
				if m.sidebarCursor > 0 {
					m.sidebarCursor--
				}
			} else {
				m.viewport.LineUp(3)
			}
		case tea.MouseWheelDown:
			if msg.X <= sw {
				if m.sidebarCursor < len(m.convs)-1 {
					m.sidebarCursor++
				}
			} else {
				m.viewport.LineDown(3)
			}
		case tea.MouseLeft:
			if msg.X <= sw {
				m.focus = focusSidebar
				m.textInput.Blur()
			} else {
				m.focus = focusChat
				if !m.loading {
					m.textInput.Focus()
				}
			}
		}
		return m, nil

	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

// ── sub-updates ───────────────────────────────────────────────────────

func (m model) updateSidebar(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	n := len(m.convs)
	switch msg.String() {
	case "up", "k":
		if m.sidebarCursor > 0 {
			m.sidebarCursor--
		}
	case "down", "j":
		if m.sidebarCursor < n-1 {
			m.sidebarCursor++
		}
	case "enter":
		if m.sidebarCursor < n {
			return m, loadConvCmd(m.apiURL, m.convs[m.sidebarCursor].ID)
		}
	case "n":
		return m, createConvCmd(m.apiURL)
	case "d", "backspace", "ctrl+d":
		if n > 1 && m.sidebarCursor < n {
			id := m.convs[m.sidebarCursor].ID
			m.convs = append(m.convs[:m.sidebarCursor], m.convs[m.sidebarCursor+1:]...)
			if m.sidebarCursor >= len(m.convs) {
				m.sidebarCursor = len(m.convs) - 1
			}
			// pick next
			if len(m.convs) > 0 {
				return m, tea.Batch(deleteConvCmd(m.apiURL, id), loadConvCmd(m.apiURL, m.convs[m.sidebarCursor].ID))
			}
			return m, deleteConvCmd(m.apiURL, id)
		}
	}
	return m, nil
}

func (m model) updateChat(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "enter":
		text := strings.TrimSpace(m.textInput.Value())
		if text == "" || m.loading || m.active == nil {
			return m, nil
		}
		m.textInput.SetValue("")
		m.loading = true
		m.statusText = "Thinking…"
		m.toolLog = nil
		m.textInput.Blur()

		// optimistic user message
		m.appendMessage(ChatMessage{Role: "user", Text: text, Timestamp: time.Now()})
		m.refreshViewport()
		return m, chatStreamCmd(m.apiURL, m.active.ID, text)
	case "pgup", "ctrl+u":
		m.viewport.HalfViewUp()
		return m, nil
	case "pgdown":
		m.viewport.HalfViewDown()
		return m, nil
	}
	var cmd tea.Cmd
	m.textInput, cmd = m.textInput.Update(msg)
	return m, cmd
}

// ── process collected SSE events ──────────────────────────────────────

func (m model) applyChatEvents(events []sseEvent) (tea.Model, tea.Cmd) {
	for _, ev := range events {
		switch ev.Type {
		case "status":
			m.statusText, _ = ev.Data["text"].(string)
		case "tool_call":
			name, _ := ev.Data["name"].(string)
			args, _ := ev.Data["args"].(map[string]any)
			m.toolLog = append(m.toolLog, name)
			m.appendMessage(ChatMessage{Role: "tool_call", ToolName: name, ToolArgs: args, Timestamp: time.Now()})
		case "tool_result":
			name, _ := ev.Data["name"].(string)
			m.appendMessage(ChatMessage{Role: "tool_result", ToolName: name, Timestamp: time.Now()})
		case "bot_message":
			text, _ := ev.Data["text"].(string)
			var types []string
			if arr, ok := ev.Data["types"].([]any); ok {
				for _, t := range arr {
					if s, ok := t.(string); ok {
						types = append(types, s)
					}
				}
			}
			m.appendMessage(ChatMessage{Role: "bot", Text: text, Types: types, Timestamp: time.Now()})
			// update summary
			if m.active != nil {
				for i := range m.convs {
					if m.convs[i].ID == m.active.ID {
						m.convs[i].DetectedTypes = appendUnique(m.convs[i].DetectedTypes, types)
					}
				}
			}
		case "error":
			text, _ := ev.Data["text"].(string)
			m.appendMessage(ChatMessage{Role: "error", Text: text, Timestamp: time.Now()})
		case "done":
			m.loading = false
			m.statusText = ""
			m.toolLog = nil
		}
	}
	m.refreshViewport()
	if !m.loading {
		m.textInput.Focus()
	}
	return m, nil
}

// ── helpers ───────────────────────────────────────────────────────────

func (m *model) appendMessage(msg ChatMessage) {
	if m.active == nil {
		return
	}
	m.active.Messages = append(m.active.Messages, msg)
}

func appendUnique(base, add []string) []string {
	for _, a := range add {
		found := false
		for _, b := range base {
			if a == b {
				found = true
				break
			}
		}
		if !found {
			base = append(base, a)
		}
	}
	return base
}

func (m model) transitionToMain() (tea.Model, tea.Cmd) {
	m.splashProgress = 1.0
	m.state = stateMain
	m.recalcLayout()
	m.refreshViewport()
	if !m.loading {
		m.textInput.Focus()
	}
	return m, tea.Batch(textinput.Blink, m.spinner.Tick)
}

func (m *model) recalcLayout() {
	cw, ch := m.chatDimensions()
	m.viewport.Width = cw - 4
	m.viewport.Height = ch - 8
	m.textInput.Width = cw - 8
}

func (m *model) chatDimensions() (int, int) {
	sw := m.sidebarWidth()
	cw := m.width - sw
	if cw < 40 {
		cw = 40
	}
	return cw, m.height
}

func (m *model) sidebarWidth() int {
	w := m.width / 4
	if w < sidebarMinW {
		w = sidebarMinW
	}
	if w > sidebarMaxW {
		w = sidebarMaxW
	}
	return w
}

func (m *model) refreshViewport() {
	if m.active == nil {
		m.viewport.SetContent(m.renderWelcome(m.viewport.Width - 2))
		return
	}
	chatW := m.viewport.Width - 2
	if len(m.active.Messages) == 0 {
		m.viewport.SetContent(m.renderWelcome(chatW))
		m.viewport.GotoTop()
		return
	}
	var b strings.Builder
	for _, msg := range m.active.Messages {
		b.WriteString(m.renderChatMessage(msg, chatW))
		b.WriteString("\n")
	}
	if m.loading && m.statusText != "" {
		loading := lipgloss.JoinHorizontal(lipgloss.Center,
			m.spinner.View(), " ",
			spinnerLabelStyle.Render(m.statusText),
		)
		if len(m.toolLog) > 0 {
			chain := lipgloss.NewStyle().
				Foreground(colTextDim).Faint(true).
				Render(fmt.Sprintf("  [%s]", strings.Join(m.toolLog, " → ")))
			loading += chain
		}
		b.WriteString("\n" + loading + "\n")
	}
	m.viewport.SetContent(b.String())
	m.viewport.GotoBottom()
}
