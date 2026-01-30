package views

import (
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"gossh/internal/crypto"
	"gossh/internal/ui/styles"
)

// SetupModel is the first-time setup view for master password
type SetupModel struct {
	password        textinput.Model
	confirmPassword textinput.Model
	focusIndex      int
	err             error
	width           int
	height          int
}

// NewSetupModel creates a new setup model
func NewSetupModel() SetupModel {
	password := textinput.New()
	password.Placeholder = "Enter master password"
	password.EchoMode = textinput.EchoPassword
	password.CharLimit = 100
	password.Width = 40
	password.Focus()

	confirm := textinput.New()
	confirm.Placeholder = "Confirm master password"
	confirm.EchoMode = textinput.EchoPassword
	confirm.CharLimit = 100
	confirm.Width = 40

	return SetupModel{
		password:        password,
		confirmPassword: confirm,
		focusIndex:      0,
	}
}

// SetSize sets the view dimensions
func (m *SetupModel) SetSize(width, height int) {
	m.width = width
	m.height = height
}

// GetPassword returns the entered password if valid
func (m *SetupModel) GetPassword() (string, error) {
	pwd := m.password.Value()
	confirm := m.confirmPassword.Value()

	if len(pwd) < 8 {
		return "", crypto.ErrPasswordTooWeak
	}

	if pwd != confirm {
		return "", errPasswordMismatch
	}

	return pwd, nil
}

// Reset clears the form
func (m *SetupModel) Reset() {
	m.password.SetValue("")
	m.confirmPassword.SetValue("")
	m.password.Focus()
	m.confirmPassword.Blur()
	m.focusIndex = 0
	m.err = nil
}

// Init initializes the setup model
func (m SetupModel) Init() tea.Cmd {
	return textinput.Blink
}

// Update handles messages for the setup model
func (m SetupModel) Update(msg tea.Msg) (SetupModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "tab", "down":
			m.nextField()
			return m, nil
		case "shift+tab", "up":
			m.prevField()
			return m, nil
		}
	}

	// Update focused input
	var cmd tea.Cmd
	if m.focusIndex == 0 {
		m.password, cmd = m.password.Update(msg)
	} else {
		m.confirmPassword, cmd = m.confirmPassword.Update(msg)
	}
	return m, cmd
}

func (m *SetupModel) nextField() {
	if m.focusIndex == 0 {
		m.password.Blur()
		m.confirmPassword.Focus()
		m.focusIndex = 1
	}
}

func (m *SetupModel) prevField() {
	if m.focusIndex == 1 {
		m.confirmPassword.Blur()
		m.password.Focus()
		m.focusIndex = 0
	}
}

// View renders the setup view
func (m SetupModel) View() string {
	var b strings.Builder

	b.WriteString(styles.TitleStyle.Render("Welcome to GoSSH"))
	b.WriteString("\n\n")

	b.WriteString("First time setup: Please create a master password\n")
	b.WriteString(styles.DimStyle.Render("This password will be used to encrypt your saved credentials."))
	b.WriteString("\n")
	b.WriteString(styles.DimStyle.Render("Make sure to remember it!"))
	b.WriteString("\n\n")

	// Password field
	label := "Master Password:"
	if m.focusIndex == 0 {
		label = styles.SelectedStyle.Render(label)
	} else {
		label = styles.LabelStyle.Render(label)
	}
	b.WriteString(label + "\n")
	b.WriteString(m.password.View())
	b.WriteString("\n\n")

	// Confirm password field
	label = "Confirm Password:"
	if m.focusIndex == 1 {
		label = styles.SelectedStyle.Render(label)
	} else {
		label = styles.LabelStyle.Render(label)
	}
	b.WriteString(label + "\n")
	b.WriteString(m.confirmPassword.View())
	b.WriteString("\n\n")

	// Password strength indicator
	if m.password.Value() != "" {
		score, desc := crypto.PasswordStrength(m.password.Value())
		bar := renderStrengthBar(score)
		b.WriteString("Password Strength: " + bar + " " + desc)
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
	help := styles.HelpStyle.Render("tab:next field  enter:confirm  esc:exit")
	b.WriteString(help)

	return b.String()
}

func renderStrengthBar(score int) string {
	filled := score + 1
	empty := 4 - score
	bar := strings.Repeat("█", filled) + strings.Repeat("░", empty)

	switch score {
	case 0:
		return styles.ErrorStyle.Render(bar)
	case 1:
		return styles.WarningStyle.Render(bar)
	case 2, 3:
		return styles.SuccessStyle.Render(bar)
	default:
		return styles.SuccessStyle.Render(bar)
	}
}

var errPasswordMismatch = &passwordError{message: "passwords do not match"}

type passwordError struct {
	message string
}

func (e *passwordError) Error() string {
	return e.message
}
