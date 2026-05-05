// styles.go
package main

import "github.com/charmbracelet/lipgloss"

// ======================================================================
// Palette — Tokyo Night Storm
// ======================================================================

var (
	colBg        = lipgloss.Color("#1a1b26")
	colSidebarBg = lipgloss.Color("#16161e")
	colSurface   = lipgloss.Color("#24283b")
	colSurface2  = lipgloss.Color("#2f3549")
	colOverlay   = lipgloss.Color("#414868")
	colBorder    = lipgloss.Color("#3b4261")
	colBorderDim = lipgloss.Color("#292e42")
	colBorderLit = lipgloss.Color("#565f89")

	colText      = lipgloss.Color("#c0caf5")
	colTextSub   = lipgloss.Color("#a9b1d6")
	colTextDim   = lipgloss.Color("#565f89")
	colTextFaint = lipgloss.Color("#3b4261")

	colAccent  = lipgloss.Color("#e0af68")
	colGreen   = lipgloss.Color("#9ece6a")
	colRed     = lipgloss.Color("#f7768e")
	colOrange  = lipgloss.Color("#ff9e64")
	colBlue    = lipgloss.Color("#7aa2f7")
	colCyan    = lipgloss.Color("#7dcfff")
	colPurple  = lipgloss.Color("#bb9af7")
	colMagenta = lipgloss.Color("#ff007c")
	colTeal    = lipgloss.Color("#73daca")
	colPink    = lipgloss.Color("#f7768e")
)

// ======================================================================
// Sidebar
// ======================================================================

var (
	sidebarTitleStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(colAccent).
				Padding(0, 1).
				MarginBottom(0)

	sidebarActiveItem = lipgloss.NewStyle().
				Bold(true).
				Foreground(colAccent).
				Background(colSurface2).
				Padding(0, 1)

	sidebarCursorItem = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("#1a1b26")).
				Background(colAccent).
				Padding(0, 1)

	sidebarNormalItem = lipgloss.NewStyle().
				Foreground(colTextSub).
				Padding(0, 1)

	sidebarDimItem = lipgloss.NewStyle().
			Foreground(colTextDim).
			Faint(true).
			Padding(0, 1)

	sidebarDivider = lipgloss.NewStyle().
			Foreground(colBorderDim)

	sidebarHelpStyle = lipgloss.NewStyle().
				Foreground(colTextDim).
				Faint(true).
				Padding(0, 1)

	sidebarNewBtn = lipgloss.NewStyle().
			Foreground(colGreen).
			Bold(true).
			Padding(0, 1)

	sidebarCountBadge = lipgloss.NewStyle().
				Foreground(colTextDim).
				Faint(true)
)

// ======================================================================
// Chat panel
// ======================================================================

var (
	chatHeaderStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(colAccent).
			Padding(0, 1)

	chatHeaderDim = lipgloss.NewStyle().
			Foreground(colTextDim).
			Faint(true)
)

// ── User bubble ───────────────────────────────────────────────────────

var (
	userLabel = lipgloss.NewStyle().
			Bold(true).
			Foreground(colGreen).
			Padding(0, 0)

	userBubble = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(colGreen).
			Foreground(colText).
			Padding(0, 1)

	userTimestamp = lipgloss.NewStyle().
			Foreground(colTextFaint).
			Faint(true)
)

// ── Bot bubble ────────────────────────────────────────────────────────

var (
	botLabel = lipgloss.NewStyle().
			Bold(true).
			Padding(0, 0)

	botBubble = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			Foreground(colText).
			Padding(0, 1)
)

// ── Tool call / result ────────────────────────────────────────────────

var (
	toolCallBox = lipgloss.NewStyle().
			Border(lipgloss.Border{
			Top:         "─",
			Bottom:      "─",
			Left:        "│",
			Right:       "│",
			TopLeft:     "┌",
			TopRight:    "┐",
			BottomLeft:  "└",
			BottomRight: "┘",
		}).
		BorderForeground(colOverlay).
		Foreground(colOrange).
		Faint(true).
		Padding(0, 1)

	toolResultBox = lipgloss.NewStyle().
			Foreground(colGreen).
			Faint(true).
			Padding(0, 1)

	toolNameStyle = lipgloss.NewStyle().
			Foreground(colOrange).
			Bold(true)

	toolArgStyle = lipgloss.NewStyle().
			Foreground(colTextDim).
			Faint(true)
)

// ── System & error ────────────────────────────────────────────────────

var (
	systemStyle = lipgloss.NewStyle().
			Foreground(colTextDim).
			Italic(true).
			Faint(true).
			Padding(0, 1)

	errorBubble = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(colRed).
			Foreground(colRed).
			Padding(0, 1)
)

// ── Input ─────────────────────────────────────────────────────────────

var (
	inputPromptStyle = lipgloss.NewStyle().
				Foreground(colAccent).
				Bold(true)

	inputAreaStyle = lipgloss.NewStyle().
			Foreground(colText)
)

// ── Spinner ───────────────────────────────────────────────────────────

var (
	spinnerStyle = lipgloss.NewStyle().
			Foreground(colPurple)

	spinnerLabelStyle = lipgloss.NewStyle().
				Foreground(colPurple).
				Faint(true).
				Italic(true)
)

// ── Status bar / help ─────────────────────────────────────────────────

var (
	statusBarStyle = lipgloss.NewStyle().
			Foreground(colTextDim).
			Faint(true).
			Padding(0, 1)

	helpKeyStyle = lipgloss.NewStyle().
			Foreground(colTextDim).
			Bold(true)

	helpDescStyle = lipgloss.NewStyle().
			Foreground(colTextFaint)

	helpSepStyle = lipgloss.NewStyle().
			Foreground(colBorderDim)
)

// ── Welcome card ──────────────────────────────────────────────────────

var (
	welcomeCard = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(colBorderDim).
			Padding(1, 3)

	welcomeTitle = lipgloss.NewStyle().
			Bold(true).
			Foreground(colAccent)

	welcomeSubtitle = lipgloss.NewStyle().
			Foreground(colTextSub)

	welcomeHint = lipgloss.NewStyle().
			Foreground(colCyan)

	welcomeDim = lipgloss.NewStyle().
			Foreground(colTextDim)

	welcomeAccent = lipgloss.NewStyle().
			Foreground(colAccent).
			Bold(true)
)

// ======================================================================
// Pokémon type palette
// ======================================================================

var typeColors = map[string]lipgloss.Color{
	"normal":   "#a8a878",
	"fire":     "#f08030",
	"water":    "#6890f0",
	"grass":    "#78c850",
	"electric": "#f8d030",
	"ice":      "#98d8d8",
	"fighting": "#c03028",
	"poison":   "#a040a0",
	"ground":   "#e0c068",
	"flying":   "#a890f0",
	"psychic":  "#f85888",
	"bug":      "#a8b820",
	"rock":     "#b8a038",
	"ghost":    "#705898",
	"dragon":   "#7038f8",
	"dark":     "#705848",
	"steel":    "#b8b8d0",
	"fairy":    "#ee99ac",
}

var typeIcons = map[string]string{
	"normal": "⬜", "fire": "🔥", "water": "💧", "grass": "🌿",
	"electric": "⚡", "ice": "❄️", "fighting": "🥊", "poison": "☠️",
	"ground": "🌍", "flying": "🕊️", "psychic": "🔮", "bug": "🐛",
	"rock": "🪨", "ghost": "👻", "dragon": "🐉", "dark": "🌑",
	"steel": "⚙️", "fairy": "🧚",
}

func typeColor(types []string) lipgloss.Color {
	if len(types) > 0 {
		if c, ok := typeColors[types[0]]; ok {
			return c
		}
	}
	return colCyan
}

func typeIcon(types []string) string {
	if len(types) > 0 {
		if i, ok := typeIcons[types[0]]; ok {
			return i
		}
	}
	return "✦"
}

func typeBadge(t string) string {
	c, ok := typeColors[t]
	if !ok {
		c = colTextDim
	}
	return lipgloss.NewStyle().
		Foreground(lipgloss.Color("#1a1b26")).
		Background(c).
		Bold(true).
		Padding(0, 1).
		Render(t)
}

func typeDot(t string) string {
	c, ok := typeColors[t]
	if !ok {
		c = colTextDim
	}
	return lipgloss.NewStyle().Foreground(c).Render("●")
}
