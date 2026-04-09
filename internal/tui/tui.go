// Package tui provides the root TUI application model.
package tui

import (
	tea "charm.land/bubbletea/v2"
	"github.com/ariasmn/ugm/internal/tui/group"
	"github.com/ariasmn/ugm/internal/tui/manage"
	"github.com/ariasmn/ugm/internal/tui/user"
)

type state int

const (
	showUserView state = iota
	showGroupView
	showManageView
)

type model struct {
	state         state
	bu            user.BubbleUser
	bg            group.BubbleGroup
	bm            manage.BubbleManage
	width, height int
}

func (m model) Init() tea.Cmd {
	return nil
}

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
		if m.state != showManageView {
			switch msg.String() {
			case "tab":
				return updateByState(m)
			case "q":
				return m, tea.Quit
			case "m":
				if m.state == showUserView {
					if u := m.bu.SelectedUsername(); u != "" {
						m.bm.SetTarget(u)
						m.state = showManageView
						return m, nil
					}
				}
			}
		}
	}

	switch m.state {
	case showUserView:
		m.bu, cmd = m.bu.Update(msg)
	case showGroupView:
		m.bg, cmd = m.bg.Update(msg)
	case showManageView:
		m.bm, cmd = m.bm.Update(msg)
		if m.bm.IsDone() {
			m.bu.RefreshUsers()
			m.state = showUserView
			m.bu, cmd = m.bu.Update(tea.WindowSizeMsg{Width: m.width, Height: m.height})
		}
	}

	return m, cmd
}

func (m model) View() tea.View {
	var content string
	switch m.state {
	case showGroupView:
		content = m.bg.View()
	case showManageView:
		content = m.bm.View()
	default:
		content = m.bu.View()
	}
	v := tea.NewView(content)
	v.AltScreen = true
	return v
}

// InitialModel creates the root TUI model.
func InitialModel() tea.Model {
	return model{
		state: showUserView,
		bu:    user.InitialModel(),
		bg:    group.InitialModel(),
		bm:    manage.InitialModel(),
	}
}

func updateByState(m model) (model, tea.Cmd) {
	var cmd tea.Cmd
	windowSizeMsg := tea.WindowSizeMsg{
		Width:  m.width,
		Height: m.height,
	}

	if m.state == showUserView {
		m.state = showGroupView
		m.bg, cmd = m.bg.Update(windowSizeMsg)
	} else {
		m.state = showUserView
		m.bu, cmd = m.bu.Update(windowSizeMsg)
	}

	return m, cmd
}
