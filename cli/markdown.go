// markdown.go
package main

import (
	"strings"

	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/glamour/ansi"
)

// ── Pointer helpers for ansi.StyleConfig fields ───────────────────────

func strPtr(s string) *string { return &s }
func uintPtr(u uint) *uint    { return &u }
func boolPtr(b bool) *bool    { return &b }

// ── Build the full style config from scratch ──────────────────────────

func pokebotGlamourStyle() ansi.StyleConfig {
	gold := "#e0af68"
	blue := "#7aa2f7"
	cyan := "#7dcfff"
	teal := "#73daca"
	green := "#9ece6a"
	purple := "#bb9af7"
	orange := "#ff9e64"
	red := "#f7768e"
	surface := "#24283b"
	text := "#c0caf5"
	dim := "#565f89"
	faint := "#3b4261"

	return ansi.StyleConfig{

		// ── Document ──────────────────────────────────────────
		Document: ansi.StyleBlock{
			StylePrimitive: ansi.StylePrimitive{
				Color: &text,
			},
			Margin: uintPtr(0),
		},

		// ── Block quote ───────────────────────────────────────
		BlockQuote: ansi.StyleBlock{
			StylePrimitive: ansi.StylePrimitive{
				Color:  &dim,
				Italic: boolPtr(true),
			},
			Indent:      uintPtr(2),
			IndentToken: strPtr("▏ "),
		},

		// ── Paragraph ─────────────────────────────────────────
		Paragraph: ansi.StyleBlock{
			StylePrimitive: ansi.StylePrimitive{
				Color: &text,
			},
			Margin: uintPtr(1),
		},

		// ── Lists ─────────────────────────────────────────────
		List: ansi.StyleList{
			LevelIndent: 2,
			StyleBlock: ansi.StyleBlock{
				StylePrimitive: ansi.StylePrimitive{Color: &text},
			},
		},

		Item: ansi.StylePrimitive{
			BlockPrefix: "• ",
			Color:       &text,
		},

		Enumeration: ansi.StylePrimitive{
			BlockPrefix: ". ",
			Color:       &text,
		},

		// ── Headings ──────────────────────────────────────────
		Heading: ansi.StyleBlock{
			StylePrimitive: ansi.StylePrimitive{
				Color: &gold,
				Bold:  boolPtr(true),
			},
		},
		H1: ansi.StyleBlock{
			StylePrimitive: ansi.StylePrimitive{
				Color:  &gold,
				Bold:   boolPtr(true),
				Prefix: "◈ ",
			},
		},
		H2: ansi.StyleBlock{
			StylePrimitive: ansi.StylePrimitive{
				Color:  &blue,
				Bold:   boolPtr(true),
				Prefix: "◇ ",
			},
		},
		H3: ansi.StyleBlock{
			StylePrimitive: ansi.StylePrimitive{
				Color:  &cyan,
				Bold:   boolPtr(true),
				Prefix: "▸ ",
			},
		},
		H4: ansi.StyleBlock{
			StylePrimitive: ansi.StylePrimitive{
				Color: &teal,
				Bold:  boolPtr(true),
			},
		},
		H5: ansi.StyleBlock{
			StylePrimitive: ansi.StylePrimitive{
				Color: &teal,
				Bold:  boolPtr(true),
			},
		},
		H6: ansi.StyleBlock{
			StylePrimitive: ansi.StylePrimitive{
				Color: &dim,
				Bold:  boolPtr(true),
			},
		},

		// ── Inline styles ─────────────────────────────────────
		Strong: ansi.StylePrimitive{
			Color: &gold,
			Bold:  boolPtr(true),
		},
		Emph: ansi.StylePrimitive{
			Color:  &purple,
			Italic: boolPtr(true),
		},
		Strikethrough: ansi.StylePrimitive{
			CrossedOut: boolPtr(true),
			Color:      &dim,
		},

		// ── Horizontal rule ───────────────────────────────────
		HorizontalRule: ansi.StylePrimitive{
			Color:  &faint,
			Format: "\n──────────────────────────────────\n",
		},

		// ── Links ─────────────────────────────────────────────
		Link: ansi.StylePrimitive{
			Color:     &cyan,
			Underline: boolPtr(true),
		},
		LinkText: ansi.StylePrimitive{
			Color: &cyan,
			Bold:  boolPtr(true),
		},
		Image: ansi.StylePrimitive{
			Color:     &cyan,
			Underline: boolPtr(true),
		},
		ImageText: ansi.StylePrimitive{
			Color: &cyan,
		},

		// ── Inline code ───────────────────────────────────────
		Code: ansi.StyleBlock{
			StylePrimitive: ansi.StylePrimitive{
				Color:           &green,
				BackgroundColor: &surface,
			},
		},

		// ── Code block + Chroma syntax highlighting ───────────
		CodeBlock: ansi.StyleCodeBlock{
			StyleBlock: ansi.StyleBlock{
				StylePrimitive: ansi.StylePrimitive{
					Color: &green,
				},
				Margin: uintPtr(1),
			},
			Chroma: &ansi.Chroma{
				Text:  ansi.StylePrimitive{Color: &text},
				Error: ansi.StylePrimitive{Color: &red},
				Comment: ansi.StylePrimitive{
					Color:  &dim,
					Italic: boolPtr(true),
				},
				CommentPreproc: ansi.StylePrimitive{Color: &cyan},
				Keyword: ansi.StylePrimitive{
					Color: &purple,
					Bold:  boolPtr(true),
				},
				KeywordReserved:  ansi.StylePrimitive{Color: &purple},
				KeywordNamespace: ansi.StylePrimitive{Color: &cyan},
				KeywordType:      ansi.StylePrimitive{Color: &cyan},
				Operator:         ansi.StylePrimitive{Color: &cyan},
				Punctuation:      ansi.StylePrimitive{Color: &text},
				Name:             ansi.StylePrimitive{Color: &cyan},
				NameBuiltin:      ansi.StylePrimitive{Color: &cyan},
				NameTag:          ansi.StylePrimitive{Color: &red},
				NameAttribute:    ansi.StylePrimitive{Color: &purple},
				NameClass: ansi.StylePrimitive{
					Color: &gold,
					Bold:  boolPtr(true),
				},
				NameConstant:  ansi.StylePrimitive{Color: &orange},
				NameDecorator: ansi.StylePrimitive{Color: &gold},
				NameException: ansi.StylePrimitive{Color: &red},
				NameFunction: ansi.StylePrimitive{
					Color: &blue,
					Bold:  boolPtr(true),
				},
				NameOther:           ansi.StylePrimitive{Color: &text},
				Literal:             ansi.StylePrimitive{Color: &orange},
				LiteralNumber:       ansi.StylePrimitive{Color: &orange},
				LiteralDate:         ansi.StylePrimitive{Color: &orange},
				LiteralString:       ansi.StylePrimitive{Color: &green},
				LiteralStringEscape: ansi.StylePrimitive{Color: &orange},
				GenericDeleted:      ansi.StylePrimitive{Color: &red},
				GenericEmph:         ansi.StylePrimitive{Italic: boolPtr(true)},
				GenericInserted:     ansi.StylePrimitive{Color: &green},
				GenericStrong:       ansi.StylePrimitive{Bold: boolPtr(true)},
				GenericSubheading:   ansi.StylePrimitive{Color: &purple},
				Background:          ansi.StylePrimitive{BackgroundColor: &surface},
			},
		},

		// ── Table ─────────────────────────────────────────────
		Table: ansi.StyleTable{
			StyleBlock: ansi.StyleBlock{
				StylePrimitive: ansi.StylePrimitive{Color: &text},
			},
			CenterSeparator: strPtr("┼"),
			ColumnSeparator: strPtr("│"),
			RowSeparator:    strPtr("─"),
		},

		// ── Definitions ───────────────────────────────────────
		DefinitionTerm: ansi.StylePrimitive{
			Color: &blue,
			Bold:  boolPtr(true),
		},
		DefinitionDescription: ansi.StylePrimitive{
			Color: &text,
		},
	}
}

// ── Render markdown → styled terminal string ──────────────────────────

func renderMarkdown(content string, width int) string {
	if strings.TrimSpace(content) == "" {
		return ""
	}
	if width < 20 {
		width = 20
	}

	r, err := glamour.NewTermRenderer(
		glamour.WithStyles(pokebotGlamourStyle()),
		glamour.WithWordWrap(width),
	)
	if err != nil {
		return content
	}

	out, err := r.Render(content)
	if err != nil {
		return content
	}

	out = strings.TrimRight(out, "\n ")
	out = strings.TrimLeft(out, "\n")
	return out
}
