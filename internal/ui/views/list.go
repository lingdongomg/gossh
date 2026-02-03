package views

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"gossh/internal/i18n"
	"gossh/internal/model"
	"gossh/internal/ui/styles"
)

// ListKeyMap defines key bindings for the list view
type ListKeyMap struct {
	Up     key.Binding
	Down   key.Binding
	Enter  key.Binding
	Add    key.Binding
	Edit   key.Binding
	Delete key.Binding
	Help   key.Binding
	Quit   key.Binding
	Search key.Binding
	Top    key.Binding
	Bottom key.Binding
}

// DefaultListKeyMap returns default list key bindings
var DefaultListKeyMap = ListKeyMap{
	Up: key.NewBinding(
		key.WithKeys("up", "k"),
		key.WithHelp("↑/k", "up"),
	),
	Down: key.NewBinding(
		key.WithKeys("down", "j"),
		key.WithHelp("↓/j", "down"),
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
	Search: key.NewBinding(
		key.WithKeys("/"),
		key.WithHelp("/", "search"),
	),
	Top: key.NewBinding(
		key.WithKeys("g"),
		key.WithHelp("g", "top"),
	),
	Bottom: key.NewBinding(
		key.WithKeys("G"),
		key.WithHelp("G", "bottom"),
	),
}

// ListModel is the connection list view
type ListModel struct {
	connections []model.Connection
	filtered    []model.Connection
	cursor      int
	width       int
	height      int
	keys        ListKeyMap
	searchInput textinput.Model
	searching   bool
	searchQuery string
	groupView   bool // If true, show grouped by group
}

// NewListModel creates a new list model
func NewListModel() ListModel {
	search := textinput.New()
	search.Placeholder = "Search..."
	search.CharLimit = 50
	search.Width = 30
	search.Prompt = "/ "

	return ListModel{
		connections: []model.Connection{},
		filtered:    []model.Connection{},
		cursor:      0,
		keys:        DefaultListKeyMap,
		searchInput: search,
		groupView:   true,
	}
}

// SetConnections updates the connections list
func (m *ListModel) SetConnections(conns []model.Connection) {
	m.connections = conns
	m.applyFilter()
}

// applyFilter filters connections based on search query
func (m *ListModel) applyFilter() {
	if m.searchQuery == "" {
		m.filtered = m.connections
	} else {
		m.filtered = make([]model.Connection, 0)
		for _, conn := range m.connections {
			if conn.MatchesFilter(m.searchQuery) {
				m.filtered = append(m.filtered, conn)
			}
		}
	}

	// Adjust cursor if needed
	if m.cursor >= len(m.filtered) && len(m.filtered) > 0 {
		m.cursor = len(m.filtered) - 1
	}
	if m.cursor < 0 {
		m.cursor = 0
	}
}

// SetSize sets the view dimensions
func (m *ListModel) SetSize(width, height int) {
	m.width = width
	m.height = height
}

// Selected returns the currently selected connection
func (m *ListModel) Selected() (model.Connection, bool) {
	if len(m.filtered) == 0 || m.cursor < 0 || m.cursor >= len(m.filtered) {
		return model.Connection{}, false
	}
	return m.filtered[m.cursor], true
}

// IsSearching returns true if in search mode
func (m *ListModel) IsSearching() bool {
	return m.searching
}

// StartSearch enters search mode
func (m *ListModel) StartSearch() {
	m.searching = true
	m.searchInput.SetValue("")
	m.searchInput.Focus()
}

// ClearSearch exits search mode and clears filter
func (m *ListModel) ClearSearch() {
	m.searching = false
	m.searchQuery = ""
	m.searchInput.SetValue("")
	m.searchInput.Blur()
	m.applyFilter()
}

// Init initializes the list model
func (m ListModel) Init() tea.Cmd {
	return nil
}

// Update handles messages for the list model
func (m ListModel) Update(msg tea.Msg) (ListModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if m.searching {
			// Handle search input
			switch msg.String() {
			case "enter":
				m.searching = false
				m.searchInput.Blur()
				return m, nil
			case "esc":
				m.ClearSearch()
				return m, nil
			default:
				var cmd tea.Cmd
				m.searchInput, cmd = m.searchInput.Update(msg)
				m.searchQuery = m.searchInput.Value()
				m.applyFilter()
				return m, cmd
			}
		}

		switch {
		case key.Matches(msg, m.keys.Up):
			if m.cursor > 0 {
				m.cursor--
			}
		case key.Matches(msg, m.keys.Down):
			if m.cursor < len(m.filtered)-1 {
				m.cursor++
			}
		case key.Matches(msg, m.keys.Top):
			m.cursor = 0
		case key.Matches(msg, m.keys.Bottom):
			if len(m.filtered) > 0 {
				m.cursor = len(m.filtered) - 1
			}
		}
	}
	return m, nil
}

// View renders the list
func (m ListModel) View() string {
	var b strings.Builder

	// Title
	title := styles.TitleStyle.Render(i18n.T("list.title"))
	b.WriteString(title)
	b.WriteString("\n\n")

	// Search bar if searching
	if m.searching {
		b.WriteString(m.searchInput.View())
		b.WriteString("\n\n")
	} else if m.searchQuery != "" {
		b.WriteString(styles.DimStyle.Render(fmt.Sprintf(i18n.T("list.filter"), m.searchQuery)))
		b.WriteString("\n\n")
	}

	if len(m.filtered) == 0 {
		if m.searchQuery != "" {
			b.WriteString(styles.DimStyle.Render(i18n.T("list.empty.search")))
		} else {
			b.WriteString(styles.DimStyle.Render(i18n.T("list.empty")))
		}
		b.WriteString("\n")
	} else if m.groupView {
		// Group by group name
		groups := make(map[string][]model.Connection)
		groupOrder := []string{}
		for _, conn := range m.filtered {
			group := conn.Group
			if group == "" {
				group = i18n.T("list.ungrouped")
			}
			if _, exists := groups[group]; !exists {
				groupOrder = append(groupOrder, group)
			}
			groups[group] = append(groups[group], conn)
		}

		// Track absolute index for cursor
		idx := 0
		for _, groupName := range groupOrder {
			conns := groups[groupName]
			// Group header
			groupStyle := styles.LabelStyle
			b.WriteString(groupStyle.Render("▾ " + groupName))
			b.WriteString(styles.DimStyle.Render(fmt.Sprintf(" (%d)", len(conns))))
			b.WriteString("\n")

			for _, conn := range conns {
				line := m.renderConnectionLine(conn, idx == m.cursor)
				b.WriteString("  " + line + "\n")
				idx++
			}
			b.WriteString("\n")
		}
	} else {
		// Flat list
		for i, conn := range m.filtered {
			line := m.renderConnectionLine(conn, i == m.cursor)
			b.WriteString(line + "\n")
		}
	}

	// Stats
	b.WriteString(styles.DimStyle.Render(fmt.Sprintf(i18n.T("list.total"), len(m.connections))))
	if m.searchQuery != "" {
		b.WriteString(styles.DimStyle.Render(fmt.Sprintf(i18n.T("list.showing"), len(m.filtered))))
	}
	b.WriteString("\n")

	// Help
	b.WriteString("\n")
	var help string
	if m.searching {
		help = styles.HelpStyle.Render(i18n.T("list.help.search"))
	} else {
		help = styles.HelpStyle.Render(i18n.T("list.help"))
	}
	b.WriteString(help)

	return b.String()
}

func (m *ListModel) renderConnectionLine(conn model.Connection, selected bool) string {
	cursor := "  "
	style := styles.NormalStyle
	if selected {
		cursor = "> "
		style = styles.SelectedStyle
	}

	// Status indicator
	statusIcon := "○"
	switch conn.LastStatus {
	case model.ConnStatusSuccess:
		statusIcon = styles.SuccessStyle.Render("●")
	case model.ConnStatusFailed:
		statusIcon = styles.ErrorStyle.Render("●")
	}

	// Format: name (user@host:port)
	name := style.Render(conn.Name)
	details := styles.DimStyle.Render(fmt.Sprintf("%s@%s:%d", conn.User, conn.Host, conn.Port))

	// Auth indicator
	authIcon := "[key]"
	if conn.AuthMethod == model.AuthPassword {
		authIcon = "[pwd]"
	}

	// Tags
	var tags string
	if len(conn.Tags) > 0 {
		tags = styles.DimStyle.Render(" [" + strings.Join(conn.Tags, ", ") + "]")
	}

	return fmt.Sprintf("%s%s %s %s %s%s", cursor, statusIcon, name, details, authIcon, tags)
}
