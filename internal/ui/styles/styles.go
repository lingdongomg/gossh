package styles

import (
	"github.com/charmbracelet/lipgloss"
)

var (
	// Colors
	PrimaryColor   = lipgloss.Color("#7D56F4")
	SecondaryColor = lipgloss.Color("#5A4FCF")
	AccentColor    = lipgloss.Color("#FF6B6B")
	SuccessColor   = lipgloss.Color("#4CAF50")
	WarningColor   = lipgloss.Color("#FFC107")
	ErrorColor     = lipgloss.Color("#F44336")
	MutedColor     = lipgloss.Color("#666666")
	BgColor        = lipgloss.Color("#1A1A2E")
	FgColor        = lipgloss.Color("#EAEAEA")

	// Base styles
	BaseStyle = lipgloss.NewStyle().
			Foreground(FgColor)

	// Title
	TitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(PrimaryColor).
			MarginBottom(1)

	// Subtitle
	SubtitleStyle = lipgloss.NewStyle().
			Foreground(MutedColor).
			Italic(true)

	// Selected item in list
	SelectedStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FFFFFF")).
			Background(PrimaryColor).
			Padding(0, 1)

	// Normal item in list
	NormalStyle = lipgloss.NewStyle().
			Foreground(FgColor).
			Padding(0, 1)

	// Dimmed text
	DimStyle = lipgloss.NewStyle().
			Foreground(MutedColor)

	// Success message
	SuccessStyle = lipgloss.NewStyle().
			Foreground(SuccessColor)

	// Error message
	ErrorStyle = lipgloss.NewStyle().
			Foreground(ErrorColor).
			Bold(true)

	// Warning message
	WarningStyle = lipgloss.NewStyle().
			Foreground(WarningColor)

	// Help text
	HelpStyle = lipgloss.NewStyle().
			Foreground(MutedColor).
			MarginTop(1)

	// Border
	BorderStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(PrimaryColor).
			Padding(1, 2)

	// Form label
	LabelStyle = lipgloss.NewStyle().
			Foreground(PrimaryColor).
			Bold(true)

	// Form input
	InputStyle = lipgloss.NewStyle().
			Foreground(FgColor)

	// Form focused input
	FocusedInputStyle = lipgloss.NewStyle().
				Foreground(FgColor).
				Background(SecondaryColor)

	// Status bar
	StatusBarStyle = lipgloss.NewStyle().
			Foreground(FgColor).
			Background(lipgloss.Color("#333333")).
			Padding(0, 1)

	// Connection status - connected
	ConnectedStyle = lipgloss.NewStyle().
			Foreground(SuccessColor).
			Bold(true)

	// Connection status - disconnected
	DisconnectedStyle = lipgloss.NewStyle().
				Foreground(ErrorColor)

	// Tag style
	TagStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFFFF")).
			Background(SecondaryColor).
			Padding(0, 1).
			MarginRight(1)

	// Dialog box
	DialogStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(AccentColor).
			Padding(1, 2).
			Width(50)

	// Button
	ButtonStyle = lipgloss.NewStyle().
			Foreground(FgColor).
			Background(MutedColor).
			Padding(0, 2).
			MarginRight(1)

	// Active button
	ActiveButtonStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#FFFFFF")).
				Background(PrimaryColor).
				Padding(0, 2).
				MarginRight(1)
)
