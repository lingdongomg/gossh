package views

import (
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"gossh/internal/i18n"
	"gossh/internal/ui/styles"
)

// ConfirmKeyMap defines key bindings for confirm dialog
type ConfirmKeyMap struct {
	Confirm key.Binding
	Cancel  key.Binding
	Back    key.Binding
}

// DefaultConfirmKeyMap returns default confirm key bindings
var DefaultConfirmKeyMap = ConfirmKeyMap{
	Confirm: key.NewBinding(
		key.WithKeys("y"),
		key.WithHelp("y", "yes"),
	),
	Cancel: key.NewBinding(
		key.WithKeys("n"),
		key.WithHelp("n", "no"),
	),
	Back: key.NewBinding(
		key.WithKeys("esc"),
		key.WithHelp("esc", "cancel"),
	),
}

// ConfirmModel is the delete confirmation dialog
type ConfirmModel struct {
	title    string
	message  string
	width    int
	height   int
	keys     ConfirmKeyMap
	selected int // 0 = No, 1 = Yes
}

// NewConfirmModel creates a new confirm dialog
func NewConfirmModel() ConfirmModel {
	return ConfirmModel{
		title:    "Confirm",
		message:  "Are you sure?",
		keys:     DefaultConfirmKeyMap,
		selected: 0,
	}
}

// SetMessage sets the dialog message
func (m *ConfirmModel) SetMessage(title, message string) {
	m.title = title
	m.message = message
	m.selected = 0
}

// SetSize sets the view dimensions
func (m *ConfirmModel) SetSize(width, height int) {
	m.width = width
	m.height = height
}

// IsConfirmed returns whether the user confirmed
func (m *ConfirmModel) IsConfirmed() bool {
	return m.selected == 1
}

// Init initializes the confirm model
func (m ConfirmModel) Init() tea.Cmd {
	return nil
}

// Update handles messages for the confirm model
func (m ConfirmModel) Update(msg tea.Msg) (ConfirmModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keys.Confirm):
			m.selected = 1
		case key.Matches(msg, m.keys.Cancel):
			m.selected = 0
		case msg.String() == "left", msg.String() == "h":
			m.selected = 0
		case msg.String() == "right", msg.String() == "l":
			m.selected = 1
		case msg.String() == "tab":
			m.selected = (m.selected + 1) % 2
		}
	}
	return m, nil
}

// View renders the confirm dialog
func (m ConfirmModel) View() string {
	var b strings.Builder

	b.WriteString(styles.TitleStyle.Render(m.title))
	b.WriteString("\n\n")
	b.WriteString(m.message)
	b.WriteString("\n\n")

	// Buttons
	noBtn := styles.ButtonStyle.Render("[ " + i18n.T("confirm.no") + " ]")
	yesBtn := styles.ButtonStyle.Render("[ " + i18n.T("confirm.yes") + " ]")

	if m.selected == 0 {
		noBtn = styles.ActiveButtonStyle.Render("[ " + i18n.T("confirm.no") + " ]")
	} else {
		yesBtn = styles.ActiveButtonStyle.Render("[ " + i18n.T("confirm.yes") + " ]")
	}

	b.WriteString(noBtn + "  " + yesBtn)
	b.WriteString("\n\n")

	help := styles.HelpStyle.Render(i18n.T("confirm.help"))
	b.WriteString(help)

	return styles.DialogStyle.Render(b.String())
}
