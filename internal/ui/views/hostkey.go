package views

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"gossh/internal/i18n"
	"gossh/internal/ssh"
	"gossh/internal/ui/styles"
)

// HostKeyModel is the host key verification dialog
type HostKeyModel struct {
	result    *ssh.HostKeyResult
	width     int
	height    int
	selected  int  // 0 = reject, 1 = accept
	accepted  bool // Whether user accepted
	update    bool // Whether to update the key (for changed keys)
	completed bool // Whether dialog is completed
}

// NewHostKeyModel creates a new host key verification dialog
func NewHostKeyModel() HostKeyModel {
	return HostKeyModel{
		selected: 0, // Default to reject for safety
	}
}

// SetResult sets the host key result to display
func (m *HostKeyModel) SetResult(result *ssh.HostKeyResult) {
	m.result = result
	m.selected = 0
	m.accepted = false
	m.update = false
	m.completed = false
}

// SetSize sets the view dimensions
func (m *HostKeyModel) SetSize(width, height int) {
	m.width = width
	m.height = height
}

// IsCompleted returns true if the dialog has been completed
func (m *HostKeyModel) IsCompleted() bool {
	return m.completed
}

// IsAccepted returns true if the user accepted the key
func (m *HostKeyModel) IsAccepted() bool {
	return m.accepted
}

// ShouldUpdate returns true if the key should be updated
func (m *HostKeyModel) ShouldUpdate() bool {
	return m.update
}

// Reset resets the dialog state
func (m *HostKeyModel) Reset() {
	m.result = nil
	m.selected = 0
	m.accepted = false
	m.update = false
	m.completed = false
}

// Init initializes the model
func (m HostKeyModel) Init() tea.Cmd {
	return nil
}

// Update handles messages
func (m HostKeyModel) Update(msg tea.Msg) (HostKeyModel, tea.Cmd) {
	if m.result == nil {
		return m, nil
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, key.NewBinding(key.WithKeys("y"))):
			m.accepted = true
			if m.result.Status == ssh.HostKeyChanged {
				m.update = true
			}
			m.completed = true
			return m, nil
		case key.Matches(msg, key.NewBinding(key.WithKeys("n"))):
			m.accepted = false
			m.completed = true
			return m, nil
		case key.Matches(msg, key.NewBinding(key.WithKeys("left", "h"))):
			m.selected = 0
		case key.Matches(msg, key.NewBinding(key.WithKeys("right", "l"))):
			m.selected = 1
		case key.Matches(msg, key.NewBinding(key.WithKeys("tab"))):
			m.selected = (m.selected + 1) % 2
		case key.Matches(msg, key.NewBinding(key.WithKeys("enter"))):
			if m.selected == 1 {
				m.accepted = true
				if m.result.Status == ssh.HostKeyChanged {
					m.update = true
				}
			}
			m.completed = true
			return m, nil
		case key.Matches(msg, key.NewBinding(key.WithKeys("esc"))):
			m.accepted = false
			m.completed = true
			return m, nil
		}
	}

	return m, nil
}

// View renders the host key dialog
func (m HostKeyModel) View() string {
	if m.result == nil {
		return ""
	}

	var b strings.Builder

	// Title based on status
	if m.result.Status == ssh.HostKeyNew {
		b.WriteString(styles.TitleStyle.Render(i18n.T("hostkey.unknown")))
		b.WriteString("\n\n")
		b.WriteString(fmt.Sprintf(i18n.T("hostkey.unknown.msg"), m.result.Host))
	} else {
		b.WriteString(styles.ErrorStyle.Render(i18n.T("hostkey.changed")))
		b.WriteString("\n\n")
		b.WriteString(styles.ErrorStyle.Render(fmt.Sprintf(i18n.T("hostkey.changed.msg"), m.result.Host)))
	}
	b.WriteString("\n\n")

	// Key info
	b.WriteString(styles.LabelStyle.Render(i18n.T("hostkey.keytype") + ":"))
	b.WriteString(" " + m.result.KeyType)
	b.WriteString("\n\n")

	b.WriteString(styles.LabelStyle.Render(i18n.T("hostkey.fingerprint") + ":"))
	b.WriteString("\n")
	b.WriteString(styles.DimStyle.Render("  " + m.result.Fingerprint))
	b.WriteString("\n\n")

	// Old key for changed status
	if m.result.Status == ssh.HostKeyChanged && m.result.OldKey != "" {
		b.WriteString(styles.WarningStyle.Render("Previous fingerprint:"))
		b.WriteString("\n")
		b.WriteString(styles.DimStyle.Render("  " + m.result.OldKey))
		b.WriteString("\n\n")
	}

	// Question
	b.WriteString(i18n.T("hostkey.trust"))
	b.WriteString("\n\n")

	// Buttons
	rejectLabel := i18n.T("hostkey.reject")
	acceptLabel := i18n.T("hostkey.accept")
	if m.result.Status == ssh.HostKeyChanged {
		acceptLabel = i18n.T("hostkey.update")
	}

	rejectBtn := styles.ButtonStyle.Render("[ " + rejectLabel + " ]")
	acceptBtn := styles.ButtonStyle.Render("[ " + acceptLabel + " ]")

	if m.selected == 0 {
		rejectBtn = styles.ActiveButtonStyle.Render("[ " + rejectLabel + " ]")
	} else {
		acceptBtn = styles.ActiveButtonStyle.Render("[ " + acceptLabel + " ]")
	}

	b.WriteString(rejectBtn + "  " + acceptBtn)
	b.WriteString("\n\n")

	// Help
	help := styles.HelpStyle.Render(i18n.T("hostkey.help"))
	b.WriteString(help)

	return styles.DialogStyle.Render(b.String())
}
