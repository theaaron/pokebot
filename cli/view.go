// cli/view.go
package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
)

func (m model) View() string {
	if m.width == 0 || m.height == 0 {
		return ""
	}
	if m.state == stateSplash {
		return m.renderSplash()
	}
	if m.initErr != "" {
		return m.renderInitError()
	}
	sidebar := m.renderSidebar()
	chat := m.renderChatPanel()
	main := lipgloss.JoinHorizontal(lipgloss.Top, sidebar, chat)
	if m.showHelp {
		return m.renderHelpOverlay(main)
	}
	return main
}

// ── init error screen ─────────────────────────────────────────────────

func (m model) renderInitError() string {
	card := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(colRed).
		Padding(2, 4).
		Render(lipgloss.JoinVertical(lipgloss.Center,
			lipgloss.NewStyle().Foreground(colRed).Bold(true).Render("✗  Connection Error"),
			"",
			lipgloss.NewStyle().Foreground(colTextSub).Render(m.initErr),
			"",
			lipgloss.NewStyle().Foreground(colTextDim).Faint(true).Render(
				"Make sure the backend is running:\n  cd backend && go run ."),
			"",
			lipgloss.NewStyle().Foreground(colTextFaint).Faint(true).Render("ctrl+c to quit"),
		))
	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, card,
		lipgloss.WithWhitespaceBackground(lipgloss.Color("#0f0f14")))
}

// ── splash ────────────────────────────────────────────────────────────

func (m model) renderSplash() string {
	w := m.width
	h := m.height

	dkBlu := lipgloss.NewStyle().Foreground(lipgloss.Color("#0d47a1")).Bold(true)
	mdBlu := lipgloss.NewStyle().Foreground(lipgloss.Color("#1565c0")).Bold(true)
	ltBlu := lipgloss.NewStyle().Foreground(lipgloss.Color("#42a5f5"))
	whtS := lipgloss.NewStyle().Foreground(lipgloss.Color("#e0e0e0"))
	dimS := lipgloss.NewStyle().Foreground(colBorderDim)
	accS := lipgloss.NewStyle().Foreground(colAccent).Bold(true)
	fntS := lipgloss.NewStyle().Foreground(colTextFaint)

	ball := lipgloss.JoinVertical(lipgloss.Center,
		dkBlu.Render("▄▄█████████████▄▄"),
		mdBlu.Render("▄███████████████████▄"),
		mdBlu.Render("████████")+ltBlu.Render("█████████")+mdBlu.Render("████████"),
		mdBlu.Render("██████")+ltBlu.Render("█████████████████")+mdBlu.Render("██████"),
		mdBlu.Render("█████████████████████████████████"),
		dimS.Render("━━━━━━━━━━━━━━━━")+accS.Render("◉")+dimS.Render("━━━━━━━━━━━━━━━━"),
		whtS.Render("█████████████████████████████████"),
		whtS.Render("█████████████████████████████"),
		whtS.Render("█████████████████████████"),
		whtS.Render("▀███████████████████▀"),
		whtS.Render("▀▀█████████████▀▀"),
	)

	title := lipgloss.NewStyle().Foreground(colAccent).Bold(true).Render("✦  P  O  K  É  B  O  T  ✦")
	subtitle := lipgloss.NewStyle().Foreground(colTextSub).Render("Your AI Pokémon Assistant")
	powered := fntS.Render("Gemini Flash Lite  ·  PokeAPI v2  ·  Bubble Tea")
	divW := clamp(w/3, 20, 50)
	divider := dimS.Render(strings.Repeat("─", divW))

	m.progressBar.Width = clamp(w/3, 20, 50)
	bar := m.progressBar.ViewAs(m.splashProgress)

	phaseIdx := int(m.splashProgress * float64(len(splashPhases)-1))
	if phaseIdx >= len(splashPhases) {
		phaseIdx = len(splashPhases) - 1
	}
	phase := lipgloss.NewStyle().Foreground(colPurple).Italic(true).
		Render(fmt.Sprintf("%s %s", m.spinner.View(), splashPhases[phaseIdx]))

	pct := lipgloss.NewStyle().Foreground(colTextDim).Faint(true).
		Render(fmt.Sprintf("%.0f%%", m.splashProgress*100))
	barLine := lipgloss.JoinHorizontal(lipgloss.Center, bar, "  ", pct)

	skip := ""
	if m.splashTick > 10 {
		skip = lipgloss.NewStyle().Foreground(colTextFaint).Faint(true).Render("Press any key to skip")
	}

	content := lipgloss.JoinVertical(lipgloss.Center,
		"", ball, "", title, subtitle, "", divider, "", powered, "",
		divider, "", barLine, phase, "", skip, "", "v1.0.0", "",
	)

	borderCol := colAccent
	if m.splashTick%20 < 10 {
		borderCol = colBorder
	}
	cardW := clamp(w-8, 44, 72)
	card := lipgloss.NewStyle().
		Border(lipgloss.DoubleBorder()).BorderForeground(borderCol).
		Padding(1, 2).Width(cardW).Align(lipgloss.Center).
		Render(content)

	return lipgloss.Place(w, h, lipgloss.Center, lipgloss.Center, card,
		lipgloss.WithWhitespaceBackground(lipgloss.Color("#0f0f14")),
		lipgloss.WithWhitespaceForeground(lipgloss.Color("#0f0f14")))
}

// ── sidebar ───────────────────────────────────────────────────────────

func (m model) renderSidebar() string {
	sw := m.sidebarWidth()
	h := m.height
	inner := sw - 2

	focused := m.focus == focusSidebar
	borderCol := colBorderDim
	if focused {
		borderCol = colAccent
	}

	title := sidebarTitleStyle.Width(inner).Render(
		fmt.Sprintf("📂 History (%d)", len(m.convs)))
	div := sidebarDivider.Render(strings.Repeat("─", inner))

	var items []string
	for i, c := range m.convs {
		isActive := m.active != nil && c.ID == m.active.ID
		isCursor := i == m.sidebarCursor && focused

		label := c.Title
		maxLabel := inner - 5
		if maxLabel < 10 {
			maxLabel = 10
		}
		if len(label) > maxLabel {
			label = label[:maxLabel] + "…"
		}

		dots := ""
		for _, t := range c.DetectedTypes {
			if len(dots) > 12 {
				dots += "…"
				break
			}
			dots += typeDot(t)
		}

		meta := formatTimeShort(c.StartedAt)
		if c.MessageCount > 0 {
			meta += fmt.Sprintf(" · %d msg", c.MessageCount)
		}

		marker := "  "
		if isActive {
			marker = "▸ "
		}

		line1 := marker + label
		line2 := "  " + meta
		if dots != "" {
			line2 += " " + dots
		}

		var item string
		switch {
		case isCursor:
			item = sidebarCursorItem.Width(inner).Render(line1 + "\n" + line2)
		case isActive:
			item = sidebarActiveItem.Width(inner).Render(line1 + "\n" + line2)
		default:
			l1 := lipgloss.NewStyle().Foreground(colTextSub).Render(line1)
			l2 := lipgloss.NewStyle().Foreground(colTextDim).Faint(true).Render(line2)
			item = sidebarNormalItem.Width(inner).Render(l1 + "\n" + l2)
		}
		items = append(items, item)
	}

	list := strings.Join(items, "\n"+sidebarDivider.Render(strings.Repeat("┄", inner))+"\n")

	newBtnText := "+ New conversation"
	if focused {
		newBtnText = "+ New conversation (n)"
	}
	newBtn := sidebarNewBtn.Width(inner).Render(newBtnText)

	helpText := "↑↓ scroll · ⏎ select · n new"
	if focused {
		helpText += "\nd delete · tab → chat"
	} else {
		helpText += "\ntab → focus · scroll ⇅"
	}
	helpLine := sidebarHelpStyle.Width(inner).Render(helpText)

	usedHeight := lipgloss.Height(title) + lipgloss.Height(div)*3 +
		lipgloss.Height(list) + lipgloss.Height(newBtn) + lipgloss.Height(helpLine) + 4
	padHeight := h - usedHeight - 2
	if padHeight < 0 {
		padHeight = 0
	}

	content := lipgloss.JoinVertical(lipgloss.Left,
		title, div, list, strings.Repeat("\n", padHeight), div, newBtn, div, helpLine)

	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).BorderForeground(borderCol).
		Width(inner).Height(h - 2).Render(content)
}

// ── chat panel ────────────────────────────────────────────────────────

func (m model) renderChatPanel() string {
	cw, ch := m.chatDimensions()
	inner := cw - 2

	focused := m.focus == focusChat
	borderCol := colBorder
	if focused {
		borderCol = colCyan
	}

	headerText := "✦ PokéBot"
	if m.active != nil && m.active.Title != "New conversation" {
		headerText += chatHeaderDim.Render("  ·  " + trunc(m.active.Title, inner-30))
	}
	badges := ""
	if m.active != nil {
		for i, t := range m.active.DetectedTypes {
			if i >= 3 {
				break
			}
			badges += " " + typeBadge(t)
		}
	}
	header := chatHeaderStyle.Width(inner).Render(headerText + badges)
	headerDiv := lipgloss.NewStyle().Foreground(borderCol).Render(strings.Repeat("─", inner))

	vpStr := m.viewport.View()

	inputDiv := lipgloss.NewStyle().Foreground(borderCol).Render(strings.Repeat("─", inner))

	var inputLine string
	if m.loading {
		inputLine = lipgloss.JoinHorizontal(lipgloss.Center,
			"  ", m.spinner.View(), " ", spinnerLabelStyle.Render(m.statusText))
	} else {
		inputLine = m.textInput.View()
	}

	scrollPct := fmt.Sprintf(" %3.f%%", m.viewport.ScrollPercent()*100)
	helpView := m.help.View(m.keys)
	rightW := lipgloss.Width(scrollPct) + 2
	leftW := inner - rightW
	if leftW < 10 {
		leftW = 10
	}
	statusLine := lipgloss.JoinHorizontal(lipgloss.Top,
		statusBarStyle.Width(leftW).Render(helpView),
		statusBarStyle.Render(scrollPct))

	content := lipgloss.JoinVertical(lipgloss.Left,
		header, headerDiv, vpStr, inputDiv, inputLine, statusLine)

	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).BorderForeground(borderCol).
		Width(inner).Height(ch - 2).Render(content)
}

// ── message rendering ─────────────────────────────────────────────────

func (m model) renderChatMessage(msg ChatMessage, maxW int) string {
	switch msg.Role {
	case "user":
		bubbleW := clamp(maxW*2/3, 20, maxW-4)
		label := userLabel.Render("You")
		ts := userTimestamp.Render(formatTimeFull(msg.Timestamp))
		header := lipgloss.JoinHorizontal(lipgloss.Bottom, label, "  ", ts)
		bubble := userBubble.Width(bubbleW).Render(msg.Text)
		block := lipgloss.JoinVertical(lipgloss.Right, header, bubble)
		return lipgloss.PlaceHorizontal(maxW, lipgloss.Right, block)

	case "bot":
		tc := typeColor(msg.Types)
		icon := typeIcon(msg.Types)
		bubbleW := clamp(maxW*3/4, 30, maxW-4)
		mdWidth := bubbleW - 4
		if mdWidth < 20 {
			mdWidth = 20
		}
		rendered := renderMarkdown(msg.Text, mdWidth)
		label := botLabel.Foreground(tc).Render(icon + " PokéBot")
		badges := ""
		for i, t := range msg.Types {
			if i >= 2 {
				break
			}
			badges += " " + typeBadge(t)
		}
		bubble := botBubble.BorderForeground(tc).Width(bubbleW).Render(rendered)
		return lipgloss.JoinVertical(lipgloss.Left, label+badges, bubble)

	case "tool_call":
		name := toolNameStyle.Render(msg.ToolName)
		args := toolArgStyle.Render(formatToolArgs(msg.ToolArgs))
		inner := fmt.Sprintf("⚙  %s  %s", name, args)
		boxW := clamp(lipgloss.Width(inner)+4, 20, maxW-4)
		return toolCallBox.Width(boxW).Render(inner)

	case "tool_result":
		return toolResultBox.Render(fmt.Sprintf("  ✓  %s → received", msg.ToolName))

	case "error":
		bubbleW := clamp(maxW-4, 30, 72)
		return errorBubble.Width(bubbleW).Render("✗  " + msg.Text)
	}
	return ""
}

// ── welcome ───────────────────────────────────────────────────────────

func (m model) renderWelcome(maxW int) string {
	cardW := clamp(maxW-4, 40, 68)

	pokeball := lipgloss.NewStyle().Foreground(lipgloss.Color("#42a5f5")).Render("      ◓") +
		lipgloss.NewStyle().Foreground(colTextDim).Render("━━━━") +
		lipgloss.NewStyle().Foreground(colAccent).Bold(true).Render("●") +
		lipgloss.NewStyle().Foreground(colTextDim).Render("━━━━") +
		lipgloss.NewStyle().Foreground(colText).Render("◓")

	title := welcomeTitle.Render("⚡ Welcome to PokéBot!")
	subtitle := welcomeSubtitle.Render(
		"Ask me anything about Pokémon. I have access to live\ndata from PokeAPI and can help with all your questions!")

	body := welcomeHint.Render("💡 Try asking:")
	hints := []string{
		"   → Where can I find Emolga in Pokémon Violet?",
		"   → What moves does Charizard learn?",
		"   → Show me the evolution chain for Eevee",
		"   → What are all the Dragon type Pokémon?",
		"   → Tell me about the Life Orb item",
		"   → What nature should I pick for Garchomp?",
	}
	var hintBlock strings.Builder
	for _, h := range hints {
		hintBlock.WriteString(welcomeHint.Render(h) + "\n")
	}

	divider := lipgloss.NewStyle().Foreground(colBorderDim).Render(strings.Repeat("─", cardW-8))
	kb := welcomeDim.Render("⌨️  Shortcuts")
	shortcuts := welcomeDim.Render(
		"   tab      toggle sidebar / chat\n" +
			"   ctrl+n   new conversation\n" +
			"   ?        toggle full help\n" +
			"   scroll   mouse wheel supported\n" +
			"   ctrl+c   quit")
	powered := lipgloss.NewStyle().Foreground(colTextFaint).Faint(true).
		Align(lipgloss.Center).Width(cardW - 8).
		Render("Gemini Flash Lite  ·  PokeAPI v2  ·  Go + Bubble Tea")

	content := lipgloss.JoinVertical(lipgloss.Left,
		"", pokeball, "", title, subtitle, "",
		divider, "", body, hintBlock.String(),
		divider, "", kb, shortcuts, "",
		divider, "", powered, "")

	return welcomeCard.Width(cardW).Render(content)
}

// ── help overlay ──────────────────────────────────────────────────────

func (m model) renderHelpOverlay(base string) string {
	overlayW := clamp(m.width/2, 40, 60)
	title := lipgloss.NewStyle().Bold(true).Foreground(colAccent).Render("⌨️  Keyboard Shortcuts")
	divider := lipgloss.NewStyle().Foreground(colBorderDim).Render(strings.Repeat("─", overlayW-4))

	sections := []struct {
		header string
		keys   []string
	}{
		{"General", []string{
			"tab        Switch between sidebar & chat",
			"ctrl+n     New conversation",
			"ctrl+c     Quit",
			"?          Toggle this help",
			"esc        Close help",
		}},
		{"Chat (when focused)", []string{
			"enter      Send message",
			"pgup/pgdn  Scroll chat",
			"ctrl+u     Scroll up (half page)",
			"mouse ⇅    Scroll with wheel",
		}},
		{"Sidebar (when focused)", []string{
			"↑/k ↓/j    Navigate conversations",
			"enter      Switch to conversation",
			"n          New conversation",
			"d          Delete conversation",
		}},
	}

	var body strings.Builder
	for _, sec := range sections {
		body.WriteString(lipgloss.NewStyle().Bold(true).Foreground(colBlue).Render(sec.header) + "\n")
		for _, k := range sec.keys {
			parts := strings.SplitN(k, " ", 2)
			keyStr := lipgloss.NewStyle().Foreground(colAccent).Bold(true).Width(12).Render(strings.TrimSpace(parts[0]))
			desc := lipgloss.NewStyle().Foreground(colTextSub).Render(strings.TrimSpace(parts[1]))
			body.WriteString("  " + keyStr + desc + "\n")
		}
		body.WriteString("\n")
	}

	content := lipgloss.JoinVertical(lipgloss.Left, "", title, "", divider, "", body.String())
	overlay := lipgloss.NewStyle().
		Border(lipgloss.DoubleBorder()).BorderForeground(colAccent).
		Background(colBg).Padding(1, 2).Width(overlayW).Render(content)

	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, overlay,
		lipgloss.WithWhitespaceBackground(lipgloss.Color("#0f0f14")))
}

// ── small helpers ─────────────────────────────────────────────────────

func formatToolArgs(args map[string]any) string {
	if args == nil {
		return ""
	}
	var parts []string
	for k, v := range args {
		parts = append(parts, fmt.Sprintf("%s: %v", k, v))
	}
	return strings.Join(parts, ", ")
}

func formatTimeShort(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	now := time.Now()
	if t.YearDay() == now.YearDay() && t.Year() == now.Year() {
		return t.Format("3:04 PM")
	}
	if now.Sub(t) < 7*24*time.Hour {
		return t.Format("Mon 3:04 PM")
	}
	return t.Format("Jan 2")
}

func formatTimeFull(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.Format("3:04 PM")
}

func trunc(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "…"
}

func clamp(val, lo, hi int) int {
	if val < lo {
		return lo
	}
	if val > hi {
		return hi
	}
	return val
}
