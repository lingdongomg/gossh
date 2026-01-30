package views

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"gossh/internal/ui/styles"
)

// HelpModel is the help screen
type HelpModel struct {
	width  int
	height int
}

// NewHelpModel creates a new help model
func NewHelpModel() HelpModel {
	return HelpModel{}
}

// SetSize sets the view dimensions
func (m *HelpModel) SetSize(width, height int) {
	m.width = width
	m.height = height
}

// Init initializes the help model
func (m HelpModel) Init() tea.Cmd {
	return nil
}

// Update handles messages for the help model
func (m HelpModel) Update(msg tea.Msg) (HelpModel, tea.Cmd) {
	return m, nil
}

// View renders the help screen
func (m HelpModel) View() string {
	var b strings.Builder

	b.WriteString(styles.TitleStyle.Render("GoSSH Help - v1.0"))
	b.WriteString("\n\n")

	sections := []struct {
		title string
		keys  []struct {
			key  string
			desc string
		}
	}{
		{
			title: "Navigation",
			keys: []struct {
				key  string
				desc string
			}{
				{"up/k", "Move up"},
				{"down/j", "Move down"},
				{"g", "Jump to top"},
				{"G", "Jump to bottom"},
				{"/", "Search connections"},
				{"Enter", "Connect to selected server"},
			},
		},
		{
			title: "Connection Management",
			keys: []struct {
				key  string
				desc string
			}{
				{"a", "Add new connection"},
				{"e", "Edit selected connection"},
				{"d", "Delete selected connection"},
			},
		},
		{
			title: "Form Navigation",
			keys: []struct {
				key  string
				desc string
			}{
				{"Tab", "Next field"},
				{"Shift+Tab", "Previous field"},
				{"Space", "Toggle auth method / Cycle group"},
				{"Enter", "Save"},
				{"Esc", "Cancel"},
			},
		},
		{
			title: "Search Mode",
			keys: []struct {
				key  string
				desc string
			}{
				{"Type", "Filter by name, host, user, group, or tags"},
				{"Enter", "Confirm search and connect"},
				{"Esc", "Cancel search"},
			},
		},
		{
			title: "General",
			keys: []struct {
				key  string
				desc string
			}{
				{"?", "Show this help"},
				{"q", "Quit application"},
				{"Esc", "Go back / Cancel"},
			},
		},
		{
			title: "CLI Commands",
			keys: []struct {
				key  string
				desc string
			}{
				{"gossh list", "List all connections"},
				{"gossh connect <name>", "Connect by name"},
				{"gossh export [file]", "Export connections"},
				{"gossh import <file>", "Import connections"},
			},
		},
	}

	for _, section := range sections {
		b.WriteString(styles.LabelStyle.Render(section.title))
		b.WriteString("\n")
		for _, k := range section.keys {
			b.WriteString("  ")
			b.WriteString(styles.SelectedStyle.Render(k.key))
			b.WriteString("  ")
			b.WriteString(k.desc)
			b.WriteString("\n")
		}
		b.WriteString("\n")
	}

	b.WriteString(styles.HelpStyle.Render("Press Esc or ? to return"))

	return b.String()
}
