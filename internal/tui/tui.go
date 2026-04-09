package tui

import (
	"fmt"
	"log"
	"os/exec"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/mrofisr/ugm-tui/internal/group"
	"github.com/mrofisr/ugm-tui/internal/passwd"
)

type sidebarTab int

const (
	_sidebarUsers sidebarTab = iota
	_sidebarGroups
	_sidebarAudit
	_sidebarManage // not a real tab, but a state
)

var _sidebarTabs = []string{"users", "groups", "audit"}

type model struct {
	sidebar       sidebarTab
	users         UserView
	groups        GroupView
	manage        ManageView
	auditContent  string
	showHelp      bool
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
		sidebar: _sidebarUsers,
		users:   newUserView(users),
		groups:  newGroupView(groups),
		manage:  newManageView(),
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

		// Help overlay
		if msg.String() == "?" {
			m.showHelp = !m.showHelp
			return m, nil
		}
		if m.showHelp {
			m.showHelp = false
			return m, nil
		}

		if m.sidebar != _sidebarManage {
			switch msg.String() {
			case "tab":
				return m.cycleSidebar()
			case "q":
				return m, tea.Quit
			case "m":
				if m.sidebar == _sidebarUsers {
					if u := m.users.selectedUsername(); u != "" {
						m.manage.setTarget(u)
						m.sidebar = _sidebarManage
						return m, nil
					}
				}
			}
		}
	}

	switch m.sidebar {
	case _sidebarUsers:
		m.users, cmd = m.users.update(msg)
	case _sidebarGroups:
		m.groups, cmd = m.groups.update(msg)
	case _sidebarAudit:
		// Audit is read-only, no update needed
	case _sidebarManage:
		m.manage, cmd = m.manage.update(msg)
		if m.manage.done {
			users, _ := passwd.Load()
			m.users.refresh(users)
			m.sidebar = _sidebarUsers
			m.users, cmd = m.users.update(tea.WindowSizeMsg{Width: m.width, Height: m.height})
		}
	}

	return m, cmd
}

func (m model) View() tea.View {
	var content string

	switch {
	case m.showHelp:
		content = m.renderHelp()
	case m.sidebar == _sidebarManage:
		content = lipgloss.JoinVertical(lipgloss.Left, m.manage.view(), m.statusBar())
	default:
		// Sidebar tab bar
		sidebarTabBar := renderTabBar(_sidebarTabs, int(m.sidebar))

		var body string
		switch m.sidebar {
		case _sidebarGroups:
			body = m.groups.view()
		case _sidebarAudit:
			body = m.renderAuditView()
		default:
			body = m.users.view()
		}

		content = lipgloss.JoinVertical(lipgloss.Left, sidebarTabBar, body, m.statusBar())
	}

	v := tea.NewView(content)
	v.AltScreen = true
	return v
}

func (m model) cycleSidebar() (model, tea.Cmd) {
	var cmd tea.Cmd
	sz := tea.WindowSizeMsg{Width: m.width, Height: m.height}

	switch m.sidebar {
	case _sidebarUsers:
		m.sidebar = _sidebarGroups
		m.groups, cmd = m.groups.update(sz)
	case _sidebarGroups:
		m.sidebar = _sidebarAudit
		m.auditContent = loadAuditLog()
	default:
		m.sidebar = _sidebarUsers
		m.users, cmd = m.users.update(sz)
	}
	return m, cmd
}

func (m model) statusBar() string {
	var left string
	switch m.sidebar {
	case _sidebarUsers:
		total, locked, expiring, sys := m.users.stats()
		left = fmt.Sprintf(" users:%d  act:%d  %s lck:%d  %s exp:%d  sys:%d",
			total, total-locked-expiring,
			_badgeLockedStyle.Render(_badgeLocked), locked,
			_badgeExpiryStyle.Render(_badgeExpiry), expiring,
			sys,
		)
	case _sidebarGroups:
		left = fmt.Sprintf(" %d groups", len(m.groups.list.Items()))
	case _sidebarAudit:
		left = " audit log"
	case _sidebarManage:
		left = fmt.Sprintf(" managing: %s", m.manage.targetUser)
	}

	keys := []struct{ key, desc string }{
		{"m", "manage"},
		{"tab", "switch"},
		{"j/k", "nav"},
		{"?", "help"},
		{"q", "quit"},
	}
	right := make([]string, 0, len(keys))
	for _, k := range keys {
		right = append(right, _statusKeyStyle.Render(k.key)+" "+k.desc)
	}

	leftRendered := _statusBarStyle.Render(left)
	rightRendered := _statusBarStyle.Render(strings.Join(right, "  "))

	gap := m.width - lipgloss.Width(leftRendered) - lipgloss.Width(rightRendered)
	if gap < 0 {
		gap = 0
	}

	return leftRendered + _statusBarStyle.Render(strings.Repeat(" ", gap)) + rightRendered
}

func (m model) renderAuditView() string {
	var b strings.Builder
	b.WriteString(_headerStyle.Render("Audit Log") + "\n\n")

	// ugm audit log
	b.WriteString(_dividerStyle.Render("── ugm actions (/var/log/ugm-audit.log)") + "\n")
	if m.auditContent != "" {
		lines := strings.Split(m.auditContent, "\n")
		start := 0
		if len(lines) > 20 {
			start = len(lines) - 20
		}
		for _, line := range lines[start:] {
			if strings.TrimSpace(line) != "" {
				fmt.Fprintf(&b, "  %s\n", line)
			}
		}
	} else {
		b.WriteString(_previewStyle.Render("  No audit entries yet") + "\n")
	}

	// Recent auth events
	b.WriteString("\n" + _dividerStyle.Render("── recent auth events") + "\n")
	out, err := exec.Command("last", "-n", "10").CombinedOutput()
	if err == nil {
		for _, line := range strings.Split(strings.TrimSpace(string(out)), "\n") {
			if strings.TrimSpace(line) != "" && !strings.HasPrefix(line, "wtmp") {
				fmt.Fprintf(&b, "  %s\n", line)
			}
		}
	}

	return _detailStyle.Render(b.String())
}

func loadAuditLog() string {
	out, err := exec.Command("cat", "/var/log/ugm-audit.log").CombinedOutput()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}

func (m model) renderHelp() string {
	help := `
  Navigation
  ──────────────────────────────────
  ↑/k          Previous item
  ↓/j          Next item
  ←/h          Previous page
  →/l          Next page
  /            Search
  Enter        Apply search
  Tab          Cycle sidebar: users → groups → audit
  Esc          Back / Exit manage view
  q            Quit
  Ctrl+c       Force quit

  User List
  ──────────────────────────────────
  m            Manage selected user
  s            Toggle system users (UID < 1000)
  1            Detail: overview
  2            Detail: SSH keys
  3            Detail: sudo rules
  4            Detail: activity

  Management (after pressing m)
  ──────────────────────────────────
  ↑/↓          Navigate menu
  Enter         Select action / Submit form
  Tab           Next field / Toggle auth method
  y/n           Confirm / Cancel destructive action
  Esc           Back to menu / Exit management

  Badges
  ──────────────────────────────────
  ` + _badgeActiveStyle.Render("●") + ` active    ` + _badgeLockedStyle.Render("LCK") + ` locked    ` + _badgeExpiryStyle.Render("EXP") + ` expiring
  ` + _badgeSudoStyle.Render("S") + ` sudo      ` + _badgeSysStyle.Render("SYS") + ` system account

  Press any key to close this help.
`
	return lipgloss.Place(m.width, m.height,
		lipgloss.Center, lipgloss.Center,
		_helpOverlayStyle.Render(help),
	)
}
