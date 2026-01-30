package ui

import (
	"fmt"
	"io"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"gossh/internal/config"
	"gossh/internal/model"
	"gossh/internal/ssh"
	"gossh/internal/ui/styles"
	"gossh/internal/ui/views"
)

// ViewState represents the current view
type ViewState int

const (
	ViewSetup ViewState = iota
	ViewUnlock
	ViewList
	ViewForm
	ViewConfirm
	ViewHelp
	ViewConnecting
)

// KeyMap defines the key bindings for the application
type KeyMap struct {
	Up      key.Binding
	Down    key.Binding
	Enter   key.Binding
	Add     key.Binding
	Edit    key.Binding
	Delete  key.Binding
	Help    key.Binding
	Quit    key.Binding
	Back    key.Binding
	Search  key.Binding
	Confirm key.Binding
	Cancel  key.Binding
}

// DefaultKeyMap returns the default key bindings
var DefaultKeyMap = KeyMap{
	Up: key.NewBinding(
		key.WithKeys("up", "k"),
		key.WithHelp("up/k", "up"),
	),
	Down: key.NewBinding(
		key.WithKeys("down", "j"),
		key.WithHelp("down/j", "down"),
	),
	Enter: key.NewBinding(
		key.WithKeys("enter"),
		key.WithHelp("enter", "connect"),
	),
	Add: key.NewBinding(
		key.WithKeys("a"),
		key.WithHelp("a", "add"),
	),
	Edit: key.NewBinding(
		key.WithKeys("e"),
		key.WithHelp("e", "edit"),
	),
	Delete: key.NewBinding(
		key.WithKeys("d"),
		key.WithHelp("d", "delete"),
	),
	Help: key.NewBinding(
		key.WithKeys("?"),
		key.WithHelp("?", "help"),
	),
	Quit: key.NewBinding(
		key.WithKeys("q", "ctrl+c"),
		key.WithHelp("q", "quit"),
	),
	Back: key.NewBinding(
		key.WithKeys("esc"),
		key.WithHelp("esc", "back"),
	),
	Search: key.NewBinding(
		key.WithKeys("/"),
		key.WithHelp("/", "search"),
	),
	Confirm: key.NewBinding(
		key.WithKeys("y"),
		key.WithHelp("y", "confirm"),
	),
	Cancel: key.NewBinding(
		key.WithKeys("n"),
		key.WithHelp("n", "cancel"),
	),
}

// Model is the main Bubbletea model
type Model struct {
	state     ViewState
	setup     views.SetupModel
	unlock    views.UnlockModel
	list      views.ListModel
	form      views.FormModel
	confirm   views.ConfirmModel
	help      views.HelpModel
	config    *config.Manager
	keys      KeyMap
	width     int
	height    int
	err       error
	statusMsg string
	deleteID  string
	sshConn   model.Connection
}

// NewModel creates a new app model
func NewModel(cfg *config.Manager) Model {
	m := Model{
		setup:   views.NewSetupModel(),
		unlock:  views.NewUnlockModel(),
		list:    views.NewListModel(),
		form:    views.NewFormModel(cfg.GroupNames()),
		confirm: views.NewConfirmModel(),
		help:    views.NewHelpModel(),
		config:  cfg,
		keys:    DefaultKeyMap,
	}

	// Determine initial state
	if cfg.IsFirstRun() {
		m.state = ViewSetup
	} else if !cfg.IsUnlocked() {
		m.state = ViewUnlock
	} else {
		m.state = ViewList
		m.list.SetConnections(cfg.Connections())
	}

	return m
}

// Init initializes the model
func (m Model) Init() tea.Cmd {
	return nil
}

// Update handles messages
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.setup.SetSize(msg.Width, msg.Height)
		m.unlock.SetSize(msg.Width, msg.Height)
		m.list.SetSize(msg.Width, msg.Height)
		m.form.SetSize(msg.Width, msg.Height)
		m.confirm.SetSize(msg.Width, msg.Height)
		m.help.SetSize(msg.Width, msg.Height)
		return m, nil

	case tea.KeyMsg:
		// Global quit on ctrl+c
		if msg.String() == "ctrl+c" {
			return m, tea.Quit
		}

		// Handle based on current view
		switch m.state {
		case ViewSetup:
			return m.updateSetup(msg)
		case ViewUnlock:
			return m.updateUnlock(msg)
		case ViewList:
			return m.updateList(msg)
		case ViewForm:
			return m.updateForm(msg)
		case ViewConfirm:
			return m.updateConfirm(msg)
		case ViewHelp:
			return m.updateHelp(msg)
		}

	case sshDoneMsg:
		m.state = ViewList
		if msg.err != nil {
			m.err = msg.err
			m.statusMsg = "Connection error: " + msg.err.Error()
			m.config.UpdateConnectionStatus(m.sshConn.ID, model.ConnStatusFailed)
		} else {
			m.statusMsg = "Disconnected"
			m.config.UpdateConnectionStatus(m.sshConn.ID, model.ConnStatusSuccess)
		}
		m.list.SetConnections(m.config.Connections())
		return m, nil
	}

	return m, nil
}

func (m Model) updateSetup(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, m.keys.Back):
		return m, tea.Quit

	case key.Matches(msg, m.keys.Enter):
		password, err := m.setup.GetPassword()
		if err != nil {
			m.err = err
			return m, nil
		}

		if err := m.config.SetupMasterPassword(password); err != nil {
			m.err = err
			return m, nil
		}

		m.state = ViewList
		m.list.SetConnections(m.config.Connections())
		m.statusMsg = "Master password set successfully"
		m.err = nil
		return m, nil

	default:
		var cmd tea.Cmd
		m.setup, cmd = m.setup.Update(msg)
		return m, cmd
	}
}

func (m Model) updateUnlock(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, m.keys.Back):
		return m, tea.Quit

	case key.Matches(msg, m.keys.Enter):
		password := m.unlock.GetPassword()
		if err := m.config.Unlock(password); err != nil {
			m.unlock.IncrementAttempts()
			m.unlock.SetError(err)
			m.unlock.Reset()

			if m.unlock.MaxAttemptsReached() {
				return m, tea.Quit
			}
			return m, nil
		}

		m.state = ViewList
		m.list.SetConnections(m.config.Connections())
		m.statusMsg = "Unlocked successfully"
		m.err = nil
		return m, nil

	default:
		var cmd tea.Cmd
		m.unlock, cmd = m.unlock.Update(msg)
		return m, cmd
	}
}

func (m Model) updateList(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Check if in search mode
	if m.list.IsSearching() {
		switch {
		case key.Matches(msg, m.keys.Back):
			m.list.ClearSearch()
			return m, nil
		case key.Matches(msg, m.keys.Enter):
			// If search has results and user presses enter, connect
			if conn, ok := m.list.Selected(); ok {
				m.sshConn = conn
				m.state = ViewConnecting
				return m, m.connectSSH(conn)
			}
			return m, nil
		default:
			var cmd tea.Cmd
			m.list, cmd = m.list.Update(msg)
			return m, cmd
		}
	}

	switch {
	case key.Matches(msg, m.keys.Quit):
		return m, tea.Quit

	case key.Matches(msg, m.keys.Search):
		m.list.StartSearch()
		return m, nil

	case key.Matches(msg, m.keys.Add):
		m.form.Reset()
		m.state = ViewForm
		return m, nil

	case key.Matches(msg, m.keys.Edit):
		if conn, ok := m.list.Selected(); ok {
			m.form.Reset()
			m.form.SetConnection(conn)
			m.state = ViewForm
		}
		return m, nil

	case key.Matches(msg, m.keys.Delete):
		if conn, ok := m.list.Selected(); ok {
			m.deleteID = conn.ID
			m.confirm.SetMessage("Delete Connection", fmt.Sprintf("Delete '%s'?", conn.Name))
			m.state = ViewConfirm
		}
		return m, nil

	case key.Matches(msg, m.keys.Enter):
		if conn, ok := m.list.Selected(); ok {
			m.sshConn = conn
			m.state = ViewConnecting
			return m, m.connectSSH(conn)
		}
		return m, nil

	case key.Matches(msg, m.keys.Help):
		m.state = ViewHelp
		return m, nil

	default:
		var cmd tea.Cmd
		m.list, cmd = m.list.Update(msg)
		return m, cmd
	}
}

func (m Model) updateForm(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, m.keys.Back):
		m.state = ViewList
		return m, nil

	case key.Matches(msg, m.keys.Enter):
		conn, err := m.form.GetConnection()
		if err != nil {
			m.err = err
			return m, nil
		}

		if m.form.Editing {
			if err := m.config.UpdateConnection(conn); err != nil {
				m.err = err
				return m, nil
			}
			m.statusMsg = "Connection updated"
		} else {
			if err := m.config.AddConnection(conn); err != nil {
				m.err = err
				return m, nil
			}
			m.statusMsg = "Connection added"
		}

		m.list.SetConnections(m.config.Connections())
		m.state = ViewList
		m.err = nil
		return m, nil

	default:
		var cmd tea.Cmd
		m.form, cmd = m.form.Update(msg)
		return m, cmd
	}
}

func (m Model) updateConfirm(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, m.keys.Back):
		m.state = ViewList
		return m, nil

	case key.Matches(msg, m.keys.Enter):
		if m.confirm.IsConfirmed() {
			if err := m.config.DeleteConnection(m.deleteID); err != nil {
				m.err = err
			} else {
				m.statusMsg = "Connection deleted"
				m.list.SetConnections(m.config.Connections())
			}
		}
		m.state = ViewList
		return m, nil

	default:
		var cmd tea.Cmd
		m.confirm, cmd = m.confirm.Update(msg)
		return m, cmd
	}
}

func (m Model) updateHelp(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if key.Matches(msg, m.keys.Back) || key.Matches(msg, m.keys.Help) {
		m.state = ViewList
	}
	return m, nil
}

// sshDoneMsg is sent when SSH session ends
type sshDoneMsg struct {
	err error
}

func (m Model) connectSSH(conn model.Connection) tea.Cmd {
	c := &sshExecModel{
		conn: conn,
	}
	return tea.Exec(c, func(err error) tea.Msg {
		return sshDoneMsg{err: err}
	})
}

// sshExecModel implements tea.ExecCommand for SSH connections
type sshExecModel struct {
	conn model.Connection
}

func (c *sshExecModel) Run() error {
	terminal := ssh.NewTerminal(c.conn)
	return terminal.Run()
}

func (c *sshExecModel) SetStdin(r io.Reader)  {}
func (c *sshExecModel) SetStdout(w io.Writer) {}
func (c *sshExecModel) SetStderr(w io.Writer) {}

// View renders the UI
func (m Model) View() string {
	switch m.state {
	case ViewSetup:
		return m.setup.View()
	case ViewUnlock:
		return m.unlock.View()
	case ViewForm:
		return m.form.View()
	case ViewConfirm:
		return m.confirm.View()
	case ViewHelp:
		return m.help.View()
	case ViewConnecting:
		return "Connecting to " + m.sshConn.Host + "..."
	default:
		view := m.list.View()
		if m.statusMsg != "" {
			view += "\n" + styles.DimStyle.Render(m.statusMsg)
		}
		if m.err != nil {
			view += "\n" + styles.ErrorStyle.Render("Error: "+m.err.Error())
		}
		return view
	}
}
