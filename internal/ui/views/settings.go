package views

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"gossh/internal/config"
	"gossh/internal/i18n"
	"gossh/internal/ui/styles"
)

// SettingsState represents the current state of settings view
type SettingsState int

const (
	SettingsMain SettingsState = iota
	SettingsLanguage
	SettingsPasswordEnable
	SettingsPasswordChange
	SettingsPasswordDisable
)

// SettingsModel represents the settings view
type SettingsModel struct {
	cfg           *config.Manager
	state         SettingsState
	selectedIndex int
	width         int
	height        int
	version       string
	wantBack      bool // Flag to indicate user wants to go back
	
	// For password input
	passwordInput    textinput.Model
	confirmInput     textinput.Model
	currentInput     textinput.Model
	passwordFocused  int // 0: current, 1: new, 2: confirm
	
	// Settings values
	selectedLang  i18n.Language
	
	// Messages
	message     string
	messageType string // "success" or "error"
}

// NewSettingsModel creates a new settings model
func NewSettingsModel(cfg *config.Manager) SettingsModel {
	// Password input
	pwInput := textinput.New()
	pwInput.EchoMode = textinput.EchoPassword
	pwInput.EchoCharacter = '•'
	pwInput.CharLimit = 64

	confirmInput := textinput.New()
	confirmInput.EchoMode = textinput.EchoPassword
	confirmInput.EchoCharacter = '•'
	confirmInput.CharLimit = 64

	currentInput := textinput.New()
	currentInput.EchoMode = textinput.EchoPassword
	currentInput.EchoCharacter = '•'
	currentInput.CharLimit = 64

	return SettingsModel{
		cfg:           cfg,
		state:         SettingsMain,
		selectedIndex: 0,
		selectedLang:  i18n.GetLanguage(),
		version:       "1.2.0",
		wantBack:      false,
		passwordInput: pwInput,
		confirmInput:  confirmInput,
		currentInput:  currentInput,
	}
}

// SetVersion sets the version to display
func (m *SettingsModel) SetVersion(version string) {
	m.version = version
}

// Init initializes the model
func (m SettingsModel) Init() tea.Cmd {
	return nil
}

// Update handles messages
func (m SettingsModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case tea.KeyMsg:
		// Clear message on any key
		m.message = ""
		
		switch m.state {
		case SettingsMain:
			return m.updateMain(msg)
		case SettingsLanguage:
			return m.updateLanguage(msg)
		case SettingsPasswordEnable, SettingsPasswordChange:
			return m.updatePasswordInput(msg)
		case SettingsPasswordDisable:
			return m.updatePasswordDisable(msg)
		}
	}

	return m, nil
}

func (m SettingsModel) updateMain(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	menuItems := m.getMenuItems()
	
	switch {
	case key.Matches(msg, key.NewBinding(key.WithKeys("up", "k"))):
		if m.selectedIndex > 0 {
			m.selectedIndex--
		}
	case key.Matches(msg, key.NewBinding(key.WithKeys("down", "j"))):
		if m.selectedIndex < len(menuItems)-1 {
			m.selectedIndex++
		}
	case key.Matches(msg, key.NewBinding(key.WithKeys("enter"))):
		return m.handleMenuSelect()
	case key.Matches(msg, key.NewBinding(key.WithKeys("esc", "q"))):
		m.wantBack = true
		return m, nil
	}
	
	return m, nil
}

func (m SettingsModel) updateLanguage(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, key.NewBinding(key.WithKeys("up", "k"))):
		if m.selectedLang == i18n.LangZH {
			m.selectedLang = i18n.LangEN
		}
	case key.Matches(msg, key.NewBinding(key.WithKeys("down", "j"))):
		if m.selectedLang == i18n.LangEN {
			m.selectedLang = i18n.LangZH
		}
	case key.Matches(msg, key.NewBinding(key.WithKeys("enter"))):
		// Save language setting
		i18n.SetLanguage(m.selectedLang)
		if err := m.cfg.SetLanguage(string(m.selectedLang)); err != nil {
			m.message = fmt.Sprintf("Error: %v", err)
			m.messageType = "error"
		} else {
			m.message = i18n.T("settings.saved")
			m.messageType = "success"
		}
		m.state = SettingsMain
	case key.Matches(msg, key.NewBinding(key.WithKeys("esc"))):
		m.state = SettingsMain
	}
	
	return m, nil
}

func (m SettingsModel) updatePasswordInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, key.NewBinding(key.WithKeys("esc"))):
		m.state = SettingsMain
		m.resetPasswordInputs()
		return m, nil
	case key.Matches(msg, key.NewBinding(key.WithKeys("tab", "down"))):
		// Cycle through inputs forward
		if m.state == SettingsPasswordChange {
			m.passwordFocused = (m.passwordFocused + 1) % 3
		} else {
			m.passwordFocused = (m.passwordFocused + 1) % 2
		}
		m.updateInputFocus()
		return m, nil
	case key.Matches(msg, key.NewBinding(key.WithKeys("shift+tab", "up"))):
		// Cycle through inputs backward
		if m.state == SettingsPasswordChange {
			m.passwordFocused = (m.passwordFocused + 2) % 3
		} else {
			m.passwordFocused = (m.passwordFocused + 1) % 2
		}
		m.updateInputFocus()
		return m, nil
	case key.Matches(msg, key.NewBinding(key.WithKeys("enter"))):
		return m.handlePasswordSubmit()
	}
	
	// Update the focused input
	var cmd tea.Cmd
	switch m.passwordFocused {
	case 0:
		if m.state == SettingsPasswordChange {
			m.currentInput, cmd = m.currentInput.Update(msg)
		} else {
			m.passwordInput, cmd = m.passwordInput.Update(msg)
		}
	case 1:
		if m.state == SettingsPasswordChange {
			m.passwordInput, cmd = m.passwordInput.Update(msg)
		} else {
			m.confirmInput, cmd = m.confirmInput.Update(msg)
		}
	case 2:
		m.confirmInput, cmd = m.confirmInput.Update(msg)
	}
	
	return m, cmd
}

func (m SettingsModel) updatePasswordDisable(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, key.NewBinding(key.WithKeys("esc"))):
		m.state = SettingsMain
		m.resetPasswordInputs()
		return m, nil
	case key.Matches(msg, key.NewBinding(key.WithKeys("enter"))):
		// Verify current password and disable
		currentPassword := m.currentInput.Value()
		if currentPassword == "" {
			m.message = i18n.T("form.error.required")
			m.messageType = "error"
			return m, nil
		}
		
		if err := m.cfg.DisablePassword(currentPassword); err != nil {
			m.message = fmt.Sprintf("%s: %v", i18n.T("common.error"), err)
			m.messageType = "error"
			return m, nil
		}
		
		m.message = i18n.T("settings.saved")
		m.messageType = "success"
		m.state = SettingsMain
		m.resetPasswordInputs()
		return m, nil
	}
	
	var cmd tea.Cmd
	m.currentInput, cmd = m.currentInput.Update(msg)
	return m, cmd
}

func (m *SettingsModel) updateInputFocus() {
	m.currentInput.Blur()
	m.passwordInput.Blur()
	m.confirmInput.Blur()
	
	switch m.passwordFocused {
	case 0:
		if m.state == SettingsPasswordChange {
			m.currentInput.Focus()
		} else {
			m.passwordInput.Focus()
		}
	case 1:
		if m.state == SettingsPasswordChange {
			m.passwordInput.Focus()
		} else {
			m.confirmInput.Focus()
		}
	case 2:
		m.confirmInput.Focus()
	}
}

func (m *SettingsModel) resetPasswordInputs() {
	m.passwordInput.SetValue("")
	m.confirmInput.SetValue("")
	m.currentInput.SetValue("")
	m.passwordFocused = 0
}

func (m SettingsModel) handleMenuSelect() (tea.Model, tea.Cmd) {
	menuItems := m.getMenuItems()
	if m.selectedIndex >= len(menuItems) {
		return m, nil
	}
	
	item := menuItems[m.selectedIndex]
	switch item.action {
	case "language":
		m.state = SettingsLanguage
	case "enable_password":
		m.state = SettingsPasswordEnable
		m.passwordFocused = 0
		m.passwordInput.Focus()
	case "change_password":
		m.state = SettingsPasswordChange
		m.passwordFocused = 0
		m.currentInput.Focus()
	case "disable_password":
		m.state = SettingsPasswordDisable
		m.currentInput.Focus()
	case "back":
		m.wantBack = true
		return m, nil
	}
	
	return m, nil
}

func (m SettingsModel) handlePasswordSubmit() (tea.Model, tea.Cmd) {
	password := m.passwordInput.Value()
	confirm := m.confirmInput.Value()
	
	if password == "" {
		m.message = i18n.T("form.error.required")
		m.messageType = "error"
		return m, nil
	}
	
	if password != confirm {
		m.message = i18n.T("setup.password.mismatch")
		m.messageType = "error"
		return m, nil
	}
	
	if len(password) < 8 {
		m.message = i18n.T("setup.password.weak")
		m.messageType = "error"
		return m, nil
	}
	
	// Enable password protection
	if err := m.cfg.EnablePassword(password); err != nil {
		m.message = fmt.Sprintf("%s: %v", i18n.T("common.error"), err)
		m.messageType = "error"
		return m, nil
	}
	
	m.message = i18n.T("settings.saved")
	m.messageType = "success"
	m.state = SettingsMain
	m.resetPasswordInputs()
	
	return m, nil
}

type menuItem struct {
	label  string
	action string
}

func (m SettingsModel) getMenuItems() []menuItem {
	items := []menuItem{
		{label: i18n.T("settings.language"), action: "language"},
	}
	
	// Password related items based on current state
	if m.cfg.IsPasswordProtected() {
		items = append(items, menuItem{label: i18n.T("settings.password.change"), action: "change_password"})
		items = append(items, menuItem{label: i18n.T("settings.password.disable"), action: "disable_password"})
	} else {
		items = append(items, menuItem{label: i18n.T("settings.password.enable"), action: "enable_password"})
	}
	
	items = append(items, menuItem{label: i18n.T("common.back"), action: "back"})
	
	return items
}

// View renders the settings view
func (m SettingsModel) View() string {
	var b strings.Builder
	
	// Title
	title := styles.TitleStyle.Render(i18n.T("settings.title"))
	b.WriteString(title + "\n\n")
	
	switch m.state {
	case SettingsMain:
		b.WriteString(m.renderMainMenu())
	case SettingsLanguage:
		b.WriteString(m.renderLanguageSelection())
	case SettingsPasswordEnable:
		b.WriteString(m.renderPasswordEnable())
	case SettingsPasswordChange:
		b.WriteString(m.renderPasswordChange())
	case SettingsPasswordDisable:
		b.WriteString(m.renderPasswordDisable())
	}
	
	// Message
	if m.message != "" {
		style := styles.SuccessStyle
		if m.messageType == "error" {
			style = styles.ErrorStyle
		}
		b.WriteString("\n" + style.Render(m.message))
	}
	
	// Help - show different help based on state
	var helpText string
	switch m.state {
	case SettingsMain:
		helpText = i18n.T("settings.help")
	case SettingsLanguage:
		helpText = i18n.T("settings.help.language")
	case SettingsPasswordEnable, SettingsPasswordChange:
		helpText = i18n.T("settings.help.password")
	case SettingsPasswordDisable:
		helpText = i18n.T("settings.help.password.disable")
	}
	b.WriteString("\n\n" + styles.HelpStyle.Render(helpText))
	
	return b.String()
}

func (m SettingsModel) renderMainMenu() string {
	var b strings.Builder
	
	menuItems := m.getMenuItems()
	for i, item := range menuItems {
		cursor := "  "
		style := lipgloss.NewStyle()
		if i == m.selectedIndex {
			cursor = "▸ "
			style = styles.SelectedStyle
		}
		b.WriteString(cursor + style.Render(item.label) + "\n")
	}
	
	// About section
	b.WriteString("\n" + styles.SubtitleStyle.Render(i18n.T("settings.about")) + "\n")
	b.WriteString(fmt.Sprintf("  %s: v%s\n", i18n.T("app.version"), m.version))
	
	return b.String()
}

func (m SettingsModel) renderLanguageSelection() string {
	var b strings.Builder
	
	b.WriteString(styles.SubtitleStyle.Render(i18n.T("settings.language")) + "\n\n")
	
	languages := []struct {
		code i18n.Language
		name string
	}{
		{i18n.LangEN, "English"},
		{i18n.LangZH, "中文"},
	}
	
	for _, lang := range languages {
		cursor := "  "
		marker := "○"
		style := lipgloss.NewStyle()
		if lang.code == m.selectedLang {
			cursor = "▸ "
			marker = "●"
			style = styles.SelectedStyle
		}
		b.WriteString(fmt.Sprintf("%s%s %s\n", cursor, marker, style.Render(lang.name)))
	}
	
	return b.String()
}

func (m SettingsModel) renderPasswordEnable() string {
	var b strings.Builder
	
	// Update placeholders with current language
	m.passwordInput.Placeholder = ""
	m.confirmInput.Placeholder = ""
	
	b.WriteString(styles.SubtitleStyle.Render(i18n.T("settings.password.enable")) + "\n\n")
	b.WriteString(i18n.T("setup.password.prompt") + "\n")
	b.WriteString(m.passwordInput.View() + "\n\n")
	b.WriteString(i18n.T("setup.password.confirm") + "\n")
	b.WriteString(m.confirmInput.View() + "\n")
	
	return b.String()
}

func (m SettingsModel) renderPasswordChange() string {
	var b strings.Builder
	
	// Update placeholders with current language
	m.currentInput.Placeholder = ""
	m.passwordInput.Placeholder = ""
	m.confirmInput.Placeholder = ""
	
	b.WriteString(styles.SubtitleStyle.Render(i18n.T("settings.password.change")) + "\n\n")
	b.WriteString(i18n.T("unlock.prompt") + "\n")
	b.WriteString(m.currentInput.View() + "\n\n")
	b.WriteString(i18n.T("setup.password.prompt") + "\n")
	b.WriteString(m.passwordInput.View() + "\n\n")
	b.WriteString(i18n.T("setup.password.confirm") + "\n")
	b.WriteString(m.confirmInput.View() + "\n")
	
	return b.String()
}

func (m SettingsModel) renderPasswordDisable() string {
	var b strings.Builder
	
	// Update placeholder with current language
	m.currentInput.Placeholder = ""
	
	b.WriteString(styles.SubtitleStyle.Render(i18n.T("settings.password.disable")) + "\n\n")
	b.WriteString(i18n.T("unlock.prompt") + "\n")
	b.WriteString(m.currentInput.View() + "\n")
	
	return b.String()
}

// ShouldQuit returns true if the user wants to go back
func (m SettingsModel) ShouldQuit() bool {
	return m.wantBack
}
