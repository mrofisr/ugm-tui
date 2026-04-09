package tui

import (
	"log"

	tea "charm.land/bubbletea/v2"

	"github.com/mrofisr/ugm-tui/internal/group"
	"github.com/mrofisr/ugm-tui/internal/passwd"
)

type state int

const (
	_stateUser state = iota
	_stateGroup
	_stateManage
)

type model struct {
	state         state
	users         UserView
	groups        GroupView
	manage        ManageView
	width, height int
}

// New creates the root TUI model.
func New() tea.Model {
	users, err := passwd.Load()
	if err != nil {
		log.Fatalf("load users: %v", err)
	}
	groups, err := group.Load()
	if err != nil {
		log.Fatalf("load groups: %v", err)
	}

	return model{
		state:  _stateUser,
		users:  newUserView(users),
		groups: newGroupView(groups),
		manage: newManageView(),
	}
}

func (m model) Init() tea.Cmd { return nil }

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	case tea.KeyPressMsg:
		if msg.String() == "ctrl+c" {
			return m, tea.Quit
		}
		if m.state != _stateManage {
			switch msg.String() {
			case "tab":
				return m.switchView()
			case "q":
				return m, tea.Quit
			case "m":
				if m.state == _stateUser {
					if u := m.users.selectedUsername(); u != "" {
						m.manage.setTarget(u)
						m.state = _stateManage
						return m, nil
					}
				}
			}
		}
	}

	switch m.state {
	case _stateUser:
		m.users, cmd = m.users.update(msg)
	case _stateGroup:
		m.groups, cmd = m.groups.update(msg)
	case _stateManage:
		m.manage, cmd = m.manage.update(msg)
		if m.manage.done {
			users, _ := passwd.Load()
			m.users.refresh(users)
			m.state = _stateUser
			m.users, cmd = m.users.update(tea.WindowSizeMsg{Width: m.width, Height: m.height})
		}
	}

	return m, cmd
}

func (m model) View() tea.View {
	var content string
	switch m.state {
	case _stateGroup:
		content = m.groups.view()
	case _stateManage:
		content = m.manage.view()
	default:
		content = m.users.view()
	}
	v := tea.NewView(content)
	v.AltScreen = true
	return v
}

func (m model) switchView() (model, tea.Cmd) {
	var cmd tea.Cmd
	sz := tea.WindowSizeMsg{Width: m.width, Height: m.height}

	if m.state == _stateUser {
		m.state = _stateGroup
		m.groups, cmd = m.groups.update(sz)
	} else {
		m.state = _stateUser
		m.users, cmd = m.users.update(sz)
	}
	return m, cmd
}
