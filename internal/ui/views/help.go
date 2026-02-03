package views

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"gossh/internal/i18n"
	"gossh/internal/ui/styles"
)

// HelpModel is the help screen
type HelpModel struct {
	width   int
	height  int
	version string
}

// NewHelpModel creates a new help model
func NewHelpModel() HelpModel {
	return HelpModel{
		version: "1.2.0", // default version
	}
}

// SetVersion sets the version to display
func (m *HelpModel) SetVersion(version string) {
	m.version = version
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

	b.WriteString(styles.TitleStyle.Render(fmt.Sprintf("%s - v%s", i18n.T("help.title"), m.version)))
	b.WriteString("\n\n")

	sections := []struct {
		title string
		keys  []struct {
			key  string
			desc string
		}
	}{
		{
			title: i18n.T("help.navigation"),
			keys: []struct {
				key  string
				desc string
			}{
				{"up/k", i18n.T("help.key.up")},
				{"down/j", i18n.T("help.key.down")},
				{"g", i18n.T("help.key.top")},
				{"G", i18n.T("help.key.bottom")},
				{"/", i18n.T("help.key.search")},
				{"Enter", i18n.T("help.key.connect")},
			},
		},
		{
			title: i18n.T("help.connection"),
			keys: []struct {
				key  string
				desc string
			}{
				{"a", i18n.T("help.key.add")},
				{"e", i18n.T("help.key.edit")},
				{"d", i18n.T("help.key.delete")},
				{"t", i18n.T("help.key.test")},
			},
		},
		{
			title: i18n.T("help.form"),
			keys: []struct {
				key  string
				desc string
			}{
				{"Tab", i18n.T("help.key.tab")},
				{"Shift+Tab", i18n.T("help.key.shifttab")},
				{"Space", i18n.T("help.key.space")},
				{"Enter", i18n.T("help.key.save")},
				{"Esc", i18n.T("help.key.cancel")},
			},
		},
		{
			title: i18n.T("help.search"),
			keys: []struct {
				key  string
				desc string
			}{
				{"Type", i18n.T("help.key.type")},
				{"Enter", i18n.T("help.key.confirm")},
				{"Esc", i18n.T("help.key.cancel")},
			},
		},
		{
			title: i18n.T("help.general"),
			keys: []struct {
				key  string
				desc string
			}{
				{"s", i18n.T("help.key.settings")},
				{"?", i18n.T("help.key.help")},
				{"q", i18n.T("help.key.quit")},
				{"Esc", i18n.T("help.key.back")},
			},
		},
		{
			title: i18n.T("help.cli"),
			keys: []struct {
				key  string
				desc string
			}{
				{"gossh list", i18n.T("help.cli.list")},
				{"gossh connect <name>", i18n.T("help.cli.connect")},
				{"gossh export [file]", i18n.T("help.cli.export")},
				{"gossh import <file>", i18n.T("help.cli.import")},
				{"gossh check", i18n.T("help.cli.check")},
				{"gossh sftp <name>", i18n.T("help.cli.sftp")},
				{"gossh forward <name>", i18n.T("help.cli.forward")},
				{"gossh exec <cmd>", i18n.T("help.cli.exec")},
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

	b.WriteString(styles.HelpStyle.Render(i18n.T("help.return")))

	return b.String()
}
