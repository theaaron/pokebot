package main

import (
	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
)

// ── Key bindings ──────────────────────────────────────────────────────

type keyMap struct {
	Send     key.Binding
	Tab      key.Binding
	NewChat  key.Binding
	Delete   key.Binding
	Quit     key.Binding
	Up       key.Binding
	Down     key.Binding
	Enter    key.Binding
	ScrollUp key.Binding
	ScrollDn key.Binding
	Help     key.Binding
}

func defaultKeys() keyMap {
	return keyMap{
		Send: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("⏎", "send"),
		),
		Tab: key.NewBinding(
			key.WithKeys("tab"),
			key.WithHelp("tab", "switch panel"),
		),
		NewChat: key.NewBinding(
			key.WithKeys("ctrl+n"),
			key.WithHelp("c-n", "new chat"),
		),
		Delete: key.NewBinding(
			key.WithKeys("ctrl+d"),
			key.WithHelp("c-d", "delete chat"),
		),
		Quit: key.NewBinding(
			key.WithKeys("ctrl+c"),
			key.WithHelp("c-c", "quit"),
		),
		Up: key.NewBinding(
			key.WithKeys("up", "k"),
			key.WithHelp("↑/k", "up"),
		),
		Down: key.NewBinding(
			key.WithKeys("down", "j"),
			key.WithHelp("↓/j", "down"),
		),
		Enter: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("⏎", "select"),
		),
		ScrollUp: key.NewBinding(
			key.WithKeys("pgup", "ctrl+u"),
			key.WithHelp("pgup", "scroll up"),
		),
		ScrollDn: key.NewBinding(
			key.WithKeys("pgdown", "ctrl+d"),
			key.WithHelp("pgdn", "scroll down"),
		),
		Help: key.NewBinding(
			key.WithKeys("?"),
			key.WithHelp("?", "help"),
		),
	}
}

// Satisfy the help.KeyMap interface
func (k keyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Send, k.Tab, k.NewChat, k.Quit}
}

func (k keyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Send, k.Tab, k.NewChat, k.Delete},
		{k.Up, k.Down, k.ScrollUp, k.ScrollDn},
		{k.Help, k.Quit},
	}
}

// ── Help model wrapper ────────────────────────────────────────────────

func newHelp() help.Model {
	h := help.New()
	h.ShortSeparator = "  ·  "
	h.Styles.ShortKey = helpKeyStyle
	h.Styles.ShortDesc = helpDescStyle
	h.Styles.ShortSeparator = helpSepStyle
	h.Styles.FullKey = helpKeyStyle
	h.Styles.FullDesc = helpDescStyle
	h.Styles.FullSeparator = helpSepStyle
	return h
}
