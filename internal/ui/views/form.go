package views

import (
	"strconv"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"gossh/internal/model"
	"gossh/internal/ui/styles"
)

// FormKeyMap defines key bindings for the form view
type FormKeyMap struct {
	Tab      key.Binding
	ShiftTab key.Binding
	Enter    key.Binding
	Escape   key.Binding
}

// DefaultFormKeyMap returns default form key bindings
var DefaultFormKeyMap = FormKeyMap{
	Tab: key.NewBinding(
		key.WithKeys("tab"),
		key.WithHelp("tab", "next field"),
	),
	ShiftTab: key.NewBinding(
		key.WithKeys("shift+tab"),
		key.WithHelp("shift+tab", "prev field"),
	),
	Enter: key.NewBinding(
		key.WithKeys("enter"),
		key.WithHelp("enter", "save"),
	),
	Escape: key.NewBinding(
		key.WithKeys("esc"),
		key.WithHelp("esc", "cancel"),
	),
}

// FormField represents the index of form fields
type FormField int

const (
	FieldName FormField = iota
	FieldHost
	FieldPort
	FieldUser
	FieldAuthMethod
	FieldPassword
	FieldKeyPath
	FieldKeyPassword
	FieldGroup
	FieldTags
	FieldStartupCommand
	FieldCount
)

// FormModel is the add/edit connection form
type FormModel struct {
	inputs       []textinput.Model
	authMethod   model.AuthType
	focusIndex   int
	width        int
	height       int
	Editing      bool
	editID       string
	err          error
	keys         FormKeyMap
	groups       []string
	groupIndex   int
}

// NewFormModel creates a new form model
func NewFormModel(groups []string) FormModel {
	inputs := make([]textinput.Model, FieldCount)

	// Name
	inputs[FieldName] = textinput.New()
	inputs[FieldName].Placeholder = "My Server"
	inputs[FieldName].CharLimit = 50
	inputs[FieldName].Width = 30
	inputs[FieldName].Prompt = ""

	// Host
	inputs[FieldHost] = textinput.New()
	inputs[FieldHost].Placeholder = "192.168.1.1 or example.com"
	inputs[FieldHost].CharLimit = 255
	inputs[FieldHost].Width = 30
	inputs[FieldHost].Prompt = ""

	// Port
	inputs[FieldPort] = textinput.New()
	inputs[FieldPort].Placeholder = "22"
	inputs[FieldPort].CharLimit = 5
	inputs[FieldPort].Width = 10
	inputs[FieldPort].Prompt = ""
	inputs[FieldPort].SetValue("22")

	// User
	inputs[FieldUser] = textinput.New()
	inputs[FieldUser].Placeholder = "root"
	inputs[FieldUser].CharLimit = 50
	inputs[FieldUser].Width = 30
	inputs[FieldUser].Prompt = ""

	// Auth method (display only, toggle with space)
	inputs[FieldAuthMethod] = textinput.New()
	inputs[FieldAuthMethod].Placeholder = "password"
	inputs[FieldAuthMethod].CharLimit = 20
	inputs[FieldAuthMethod].Width = 20
	inputs[FieldAuthMethod].Prompt = ""
	inputs[FieldAuthMethod].SetValue("password")

	// Password
	inputs[FieldPassword] = textinput.New()
	inputs[FieldPassword].Placeholder = "********"
	inputs[FieldPassword].CharLimit = 100
	inputs[FieldPassword].Width = 30
	inputs[FieldPassword].EchoMode = textinput.EchoPassword
	inputs[FieldPassword].Prompt = ""

	// Key path
	inputs[FieldKeyPath] = textinput.New()
	inputs[FieldKeyPath].Placeholder = "~/.ssh/id_rsa"
	inputs[FieldKeyPath].CharLimit = 255
	inputs[FieldKeyPath].Width = 40
	inputs[FieldKeyPath].Prompt = ""

	// Key password
	inputs[FieldKeyPassword] = textinput.New()
	inputs[FieldKeyPassword].Placeholder = "(optional)"
	inputs[FieldKeyPassword].CharLimit = 100
	inputs[FieldKeyPassword].Width = 30
	inputs[FieldKeyPassword].EchoMode = textinput.EchoPassword
	inputs[FieldKeyPassword].Prompt = ""

	// Group (display only, cycle with space)
	inputs[FieldGroup] = textinput.New()
	inputs[FieldGroup].Placeholder = "Ungrouped"
	inputs[FieldGroup].CharLimit = 50
	inputs[FieldGroup].Width = 20
	inputs[FieldGroup].Prompt = ""

	// Tags
	inputs[FieldTags] = textinput.New()
	inputs[FieldTags].Placeholder = "web, nginx, prod"
	inputs[FieldTags].CharLimit = 200
	inputs[FieldTags].Width = 40
	inputs[FieldTags].Prompt = ""

	// Startup command
	inputs[FieldStartupCommand] = textinput.New()
	inputs[FieldStartupCommand].Placeholder = "cd /app && source venv/bin/activate"
	inputs[FieldStartupCommand].CharLimit = 500
	inputs[FieldStartupCommand].Width = 50
	inputs[FieldStartupCommand].Prompt = ""

	// Focus first field
	inputs[FieldName].Focus()

	// Prepare groups list
	allGroups := append([]string{"Ungrouped"}, groups...)

	return FormModel{
		inputs:     inputs,
		authMethod: model.AuthPassword,
		focusIndex: 0,
		keys:       DefaultFormKeyMap,
		groups:     allGroups,
		groupIndex: 0,
	}
}

// SetConnection populates the form with an existing connection
func (m *FormModel) SetConnection(conn model.Connection) {
	m.Editing = true
	m.editID = conn.ID
	m.inputs[FieldName].SetValue(conn.Name)
	m.inputs[FieldHost].SetValue(conn.Host)
	m.inputs[FieldPort].SetValue(strconv.Itoa(conn.Port))
	m.inputs[FieldUser].SetValue(conn.User)
	m.authMethod = conn.AuthMethod
	if conn.AuthMethod == model.AuthKey {
		m.inputs[FieldAuthMethod].SetValue("key")
	} else {
		m.inputs[FieldAuthMethod].SetValue("password")
	}
	m.inputs[FieldPassword].SetValue(conn.Password)
	m.inputs[FieldKeyPath].SetValue(conn.KeyPath)
	m.inputs[FieldKeyPassword].SetValue(conn.KeyPassword)

	// Set group
	groupName := conn.Group
	if groupName == "" {
		groupName = "Ungrouped"
	}
	m.inputs[FieldGroup].SetValue(groupName)
	for i, g := range m.groups {
		if g == groupName {
			m.groupIndex = i
			break
		}
	}

	// Set tags
	if len(conn.Tags) > 0 {
		m.inputs[FieldTags].SetValue(strings.Join(conn.Tags, ", "))
	}

	// Set startup command
	m.inputs[FieldStartupCommand].SetValue(conn.StartupCommand)
}

// Reset clears the form
func (m *FormModel) Reset() {
	m.Editing = false
	m.editID = ""
	m.focusIndex = 0
	m.err = nil
	m.authMethod = model.AuthPassword
	m.groupIndex = 0

	for i := range m.inputs {
		m.inputs[i].SetValue("")
		m.inputs[i].Blur()
	}
	m.inputs[FieldPort].SetValue("22")
	m.inputs[FieldAuthMethod].SetValue("password")
	m.inputs[FieldGroup].SetValue("Ungrouped")
	m.inputs[FieldName].Focus()
}

// GetConnection returns the connection from form values
func (m *FormModel) GetConnection() (model.Connection, error) {
	port, err := strconv.Atoi(m.inputs[FieldPort].Value())
	if err != nil {
		port = 22
	}

	// Parse tags
	tagsStr := m.inputs[FieldTags].Value()
	var tags []string
	if tagsStr != "" {
		for _, tag := range strings.Split(tagsStr, ",") {
			tag = strings.TrimSpace(tag)
			if tag != "" {
				tags = append(tags, tag)
			}
		}
	}

	// Get group
	group := m.inputs[FieldGroup].Value()
	if group == "Ungrouped" {
		group = ""
	}

	conn := model.Connection{
		Name:           m.inputs[FieldName].Value(),
		Host:           m.inputs[FieldHost].Value(),
		Port:           port,
		User:           m.inputs[FieldUser].Value(),
		AuthMethod:     m.authMethod,
		Password:       m.inputs[FieldPassword].Value(),
		KeyPath:        m.inputs[FieldKeyPath].Value(),
		KeyPassword:    m.inputs[FieldKeyPassword].Value(),
		Group:          group,
		Tags:           tags,
		StartupCommand: m.inputs[FieldStartupCommand].Value(),
	}

	if m.Editing {
		conn.ID = m.editID
	} else {
		conn = model.NewConnection()
		conn.Name = m.inputs[FieldName].Value()
		conn.Host = m.inputs[FieldHost].Value()
		conn.Port = port
		conn.User = m.inputs[FieldUser].Value()
		conn.AuthMethod = m.authMethod
		conn.Password = m.inputs[FieldPassword].Value()
		conn.KeyPath = m.inputs[FieldKeyPath].Value()
		conn.KeyPassword = m.inputs[FieldKeyPassword].Value()
		conn.Group = group
		conn.Tags = tags
		conn.StartupCommand = m.inputs[FieldStartupCommand].Value()
	}

	if err := conn.Validate(); err != nil {
		return conn, err
	}

	return conn, nil
}

// SetSize sets the view dimensions
func (m *FormModel) SetSize(width, height int) {
	m.width = width
	m.height = height
}

// Init initializes the form model
func (m FormModel) Init() tea.Cmd {
	return textinput.Blink
}

// Update handles messages for the form model
func (m FormModel) Update(msg tea.Msg) (FormModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keys.Tab), msg.String() == "down":
			m.nextField()
		case key.Matches(msg, m.keys.ShiftTab), msg.String() == "up":
			m.prevField()
		case msg.String() == " " && m.focusIndex == int(FieldAuthMethod):
			// Toggle auth method
			if m.authMethod == model.AuthPassword {
				m.authMethod = model.AuthKey
				m.inputs[FieldAuthMethod].SetValue("key")
			} else {
				m.authMethod = model.AuthPassword
				m.inputs[FieldAuthMethod].SetValue("password")
			}
			return m, nil
		case msg.String() == " " && m.focusIndex == int(FieldGroup):
			// Cycle through groups
			m.groupIndex = (m.groupIndex + 1) % len(m.groups)
			m.inputs[FieldGroup].SetValue(m.groups[m.groupIndex])
			return m, nil
		default:
			// Handle input for current field
			var cmd tea.Cmd
			m.inputs[m.focusIndex], cmd = m.inputs[m.focusIndex].Update(msg)
			return m, cmd
		}
	}

	// Update focused input
	var cmd tea.Cmd
	m.inputs[m.focusIndex], cmd = m.inputs[m.focusIndex].Update(msg)
	return m, cmd
}

func (m *FormModel) nextField() {
	m.inputs[m.focusIndex].Blur()
	m.focusIndex++
	if m.focusIndex >= int(FieldCount) {
		m.focusIndex = 0
	}
	// Skip password field if using key auth, skip key fields if using password
	if m.authMethod == model.AuthKey && m.focusIndex == int(FieldPassword) {
		m.focusIndex = int(FieldKeyPath)
	}
	if m.authMethod == model.AuthPassword && (m.focusIndex == int(FieldKeyPath) || m.focusIndex == int(FieldKeyPassword)) {
		m.focusIndex = int(FieldGroup)
	}
	m.inputs[m.focusIndex].Focus()
}

func (m *FormModel) prevField() {
	m.inputs[m.focusIndex].Blur()
	m.focusIndex--
	if m.focusIndex < 0 {
		m.focusIndex = int(FieldCount) - 1
	}
	// Skip key fields if using password auth, skip password if using key
	if m.authMethod == model.AuthPassword && (m.focusIndex == int(FieldKeyPath) || m.focusIndex == int(FieldKeyPassword)) {
		m.focusIndex = int(FieldAuthMethod)
	}
	if m.authMethod == model.AuthKey && m.focusIndex == int(FieldPassword) {
		m.focusIndex = int(FieldAuthMethod)
	}
	m.inputs[m.focusIndex].Focus()
}

// View renders the form
func (m FormModel) View() string {
	var b strings.Builder

	title := "Add Connection"
	if m.Editing {
		title = "Edit Connection"
	}
	b.WriteString(styles.TitleStyle.Render(title))
	b.WriteString("\n\n")

	// Form fields
	fields := []struct {
		label string
		field FormField
		show  bool
		note  string
	}{
		{"Name", FieldName, true, ""},
		{"Host", FieldHost, true, ""},
		{"Port", FieldPort, true, ""},
		{"User", FieldUser, true, ""},
		{"Auth", FieldAuthMethod, true, "(space to toggle)"},
		{"Password", FieldPassword, m.authMethod == model.AuthPassword, ""},
		{"Key Path", FieldKeyPath, m.authMethod == model.AuthKey, ""},
		{"Key Password", FieldKeyPassword, m.authMethod == model.AuthKey, "(optional)"},
		{"Group", FieldGroup, true, "(space to cycle)"},
		{"Tags", FieldTags, true, "(comma separated)"},
		{"Startup Cmd", FieldStartupCommand, true, "(runs after connect)"},
	}

	for _, f := range fields {
		if !f.show {
			continue
		}

		label := styles.LabelStyle.Render(f.label + ":")
		if m.focusIndex == int(f.field) {
			label = styles.SelectedStyle.Render(f.label + ":")
		}

		if f.field == FieldAuthMethod {
			// Show as toggle
			authDisplay := "[password] / key"
			if m.authMethod == model.AuthKey {
				authDisplay = "password / [key]"
			}
			if m.focusIndex == int(FieldAuthMethod) {
				authDisplay = styles.SelectedStyle.Render(authDisplay)
			}
			b.WriteString(label + " " + authDisplay)
			if f.note != "" {
				b.WriteString(" " + styles.DimStyle.Render(f.note))
			}
			b.WriteString("\n")
		} else if f.field == FieldGroup {
			// Show as selector
			groupDisplay := "[" + m.inputs[FieldGroup].Value() + "]"
			if m.focusIndex == int(FieldGroup) {
				groupDisplay = styles.SelectedStyle.Render(groupDisplay)
			}
			b.WriteString(label + " " + groupDisplay)
			if f.note != "" {
				b.WriteString(" " + styles.DimStyle.Render(f.note))
			}
			b.WriteString("\n")
		} else {
			b.WriteString(label + " " + m.inputs[f.field].View())
			if f.note != "" {
				b.WriteString(" " + styles.DimStyle.Render(f.note))
			}
			b.WriteString("\n")
		}
	}

	// Error message
	if m.err != nil {
		b.WriteString("\n")
		b.WriteString(styles.ErrorStyle.Render("Error: " + m.err.Error()))
		b.WriteString("\n")
	}

	// Help
	b.WriteString("\n")
	help := styles.HelpStyle.Render("tab:next field  enter:save  esc:cancel")
	b.WriteString(help)

	return b.String()
}
