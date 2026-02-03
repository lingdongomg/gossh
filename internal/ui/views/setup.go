package views

import (
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"gossh/internal/crypto"
	"gossh/internal/i18n"
	"gossh/internal/ui/styles"
)

// SetupStep represents the current step in setup
type SetupStep int

const (
	StepChooseMode SetupStep = iota
	StepSetPassword
)

// SetupModel is the first-time setup view for master password
type SetupModel struct {
	step            SetupStep
	selectedOption  int // 0: enable password, 1: skip password
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

	confirm := textinput.New()
	confirm.Placeholder = "Confirm master password"
	confirm.EchoMode = textinput.EchoPassword
	confirm.CharLimit = 100
	confirm.Width = 40

	return SetupModel{
		step:            StepChooseMode,
		selectedOption:  0,
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

// SkipPassword returns true if user chose to skip password protection
func (m *SetupModel) SkipPassword() bool {
	return m.selectedOption == 1
}

// IsChoosingMode returns true if user is on the mode selection step
func (m *SetupModel) IsChoosingMode() bool {
	return m.step == StepChooseMode
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
		if m.step == StepChooseMode {
			switch msg.String() {
			case "up", "k":
				if m.selectedOption > 0 {
					m.selectedOption--
				}
				return m, nil
			case "down", "j":
				if m.selectedOption < 1 {
					m.selectedOption++
				}
				return m, nil
			case "1":
				m.selectedOption = 0
				return m, nil
			case "2":
				m.selectedOption = 1
				return m, nil
			}
		} else {
			switch msg.String() {
			case "tab", "down":
				m.nextField()
				return m, nil
			case "shift+tab", "up":
				m.prevField()
				return m, nil
			}
		}
	}

	// Update focused input in password step
	if m.step == StepSetPassword {
		var cmd tea.Cmd
		if m.focusIndex == 0 {
			m.password, cmd = m.password.Update(msg)
		} else {
			m.confirmPassword, cmd = m.confirmPassword.Update(msg)
		}
		return m, cmd
	}
	return m, nil
}

// ProceedToPassword moves to password entry step
func (m *SetupModel) ProceedToPassword() {
	m.step = StepSetPassword
	m.password.Focus()
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

	b.WriteString(styles.TitleStyle.Render(i18n.T("setup.title")))
	b.WriteString("\n\n")

	if m.step == StepChooseMode {
		b.WriteString(i18n.T("setup.desc") + "\n\n")

		// Option 1: Enable password protection
		option1 := "  " + i18n.T("setup.option.password") + "\n"
		option1 += "      " + i18n.T("setup.option.password.desc") + "\n"
		if m.selectedOption == 0 {
			b.WriteString(styles.SelectedStyle.Render("> " + option1[2:]))
		} else {
			b.WriteString(styles.DimStyle.Render(option1))
		}
		b.WriteString("\n")

		// Option 2: Skip password protection
		option2 := "  " + i18n.T("setup.option.nopassword") + "\n"
		option2 += "      " + i18n.T("setup.option.nopassword.desc") + "\n"
		if m.selectedOption == 1 {
			b.WriteString(styles.SelectedStyle.Render("> " + option2[2:]))
		} else {
			b.WriteString(styles.DimStyle.Render(option2))
		}

		b.WriteString("\n\n")
		help := styles.HelpStyle.Render(i18n.T("setup.help.choose"))
		b.WriteString(help)
	} else {
		b.WriteString(i18n.T("setup.password.title") + "\n")
		b.WriteString(styles.DimStyle.Render(i18n.T("setup.password.desc")))
		b.WriteString("\n\n")

		// Password field
		label := i18n.T("setup.password.prompt") + ":"
		if m.focusIndex == 0 {
			label = styles.SelectedStyle.Render(label)
		} else {
			label = styles.LabelStyle.Render(label)
		}
		b.WriteString(label + "\n")
		b.WriteString(m.password.View())
		b.WriteString("\n\n")

		// Confirm password field
		label = i18n.T("setup.password.confirm") + ":"
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
			b.WriteString(i18n.T("setup.password.strength") + ": " + bar + " " + desc)
			b.WriteString("\n")
		}

		// Error message
		if m.err != nil {
			b.WriteString("\n")
			b.WriteString(styles.ErrorStyle.Render(i18n.T("common.error") + ": " + m.err.Error()))
			b.WriteString("\n")
		}

		// Help
		b.WriteString("\n")
		help := styles.HelpStyle.Render(i18n.T("setup.help.password"))
		b.WriteString(help)
	}

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
