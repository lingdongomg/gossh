package views

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"gossh/internal/ui/styles"
)

const maxUnlockAttempts = 3

// UnlockModel is the unlock view for entering master password
type UnlockModel struct {
	password textinput.Model
	attempts int
	err      error
	width    int
	height   int
}

// NewUnlockModel creates a new unlock model
func NewUnlockModel() UnlockModel {
	password := textinput.New()
	password.Placeholder = "Enter master password"
	password.EchoMode = textinput.EchoPassword
	password.CharLimit = 100
	password.Width = 40
	password.Focus()

	return UnlockModel{
		password: password,
		attempts: 0,
	}
}

// SetSize sets the view dimensions
func (m *UnlockModel) SetSize(width, height int) {
	m.width = width
	m.height = height
}

// GetPassword returns the entered password
func (m *UnlockModel) GetPassword() string {
	return m.password.Value()
}

// IncrementAttempts increments and returns the number of failed attempts
func (m *UnlockModel) IncrementAttempts() int {
	m.attempts++
	return m.attempts
}

// MaxAttemptsReached returns true if max attempts have been reached
func (m *UnlockModel) MaxAttemptsReached() bool {
	return m.attempts >= maxUnlockAttempts
}

// SetError sets an error message
func (m *UnlockModel) SetError(err error) {
	m.err = err
}

// Reset clears the password field
func (m *UnlockModel) Reset() {
	m.password.SetValue("")
	m.password.Focus()
}

// Init initializes the unlock model
func (m UnlockModel) Init() tea.Cmd {
	return textinput.Blink
}

// Update handles messages for the unlock model
func (m UnlockModel) Update(msg tea.Msg) (UnlockModel, tea.Cmd) {
	var cmd tea.Cmd
	m.password, cmd = m.password.Update(msg)
	return m, cmd
}

// View renders the unlock view
func (m UnlockModel) View() string {
	var b strings.Builder

	b.WriteString(styles.TitleStyle.Render("GoSSH Locked"))
	b.WriteString("\n\n")

	b.WriteString("Enter master password to unlock:\n\n")

	b.WriteString(styles.LabelStyle.Render("Password:") + "\n")
	b.WriteString(m.password.View())
	b.WriteString("\n\n")

	// Attempt counter
	remaining := maxUnlockAttempts - m.attempts
	if m.attempts > 0 {
		attemptStr := fmt.Sprintf("[Attempt %d/%d]", m.attempts, maxUnlockAttempts)
		if remaining <= 1 {
			b.WriteString(styles.ErrorStyle.Render(attemptStr))
		} else {
			b.WriteString(styles.WarningStyle.Render(attemptStr))
		}
		b.WriteString("\n")
	}

	// Error message
	if m.err != nil {
		b.WriteString("\n")
		b.WriteString(styles.ErrorStyle.Render("Error: " + m.err.Error()))
		b.WriteString("\n")
	}

	// Help
	b.WriteString("\n")
	help := styles.HelpStyle.Render("enter:unlock  esc:exit")
	b.WriteString(help)

	return b.String()
}
