package tui

import (
	"fmt"
	"io"
	"os/exec"
	"os/user"
	"strconv"
	"strings"

	"charm.land/bubbles/v2/list"
	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/mrofisr/ugm-tui/internal/passwd"
	"github.com/mrofisr/ugm-tui/internal/usermgmt"
)

type detailTab int

const (
	_tabOverview detailTab = iota
	_tabSSHKeys
	_tabSudoRules
	_tabActivity
)

var _detailTabs = []string{"overview", "ssh keys", "sudo rules", "activity"}

// UserView displays the list of system users and their details.
type UserView struct {
	list         list.Model
	viewport     viewport.Model
	users        []passwd.User
	showSystem   bool
	detailTab    detailTab
	width        int
	cachedDetail string
	lastSelected int // track selection changes
}

type userItem struct {
	user   passwd.User
	status string
	isSudo bool
	isSys  bool
}

func (i userItem) FilterValue() string { return i.user.Details.Username }

type userDelegate struct{}

func (d userDelegate) Height() int                             { return 1 }
func (d userDelegate) Spacing() int                            { return 0 }
func (d userDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }

func (d userDelegate) Render(w io.Writer, m list.Model, index int, item list.Item) {
	u, ok := item.(userItem)
	if !ok {
		return
	}

	var tag string
	switch {
	case u.status == "locked":
		tag = _badgeLockedStyle.Render("LCK")
	case strings.HasPrefix(u.status, "expires"):
		tag = _badgeExpiryStyle.Render("EXP")
	case u.isSys:
		tag = _badgeSysStyle.Render("SYS")
	default:
		tag = _badgeActiveStyle.Render(" " + _badgeActive + " ")
	}

	name := u.user.Details.Username
	uid := u.user.Details.Uid
	sudo := ""
	if u.isSudo {
		sudo = " " + _badgeSudoStyle.Render("S")
	}

	line := fmt.Sprintf("[%s] %-12s uid:%-5s%s", tag, name, uid, sudo)
	if index == m.Index() {
		line = _listSelectedStyle.Render("▶" + line)
	} else {
		line = _listItemStyle.Render(line)
	}
	_, _ = fmt.Fprint(w, line)
}

func newUserView(users []passwd.User) UserView {
	v := UserView{users: users}
	v.rebuildList()
	return v
}

func (v *UserView) rebuildList() {
	var items []list.Item
	for _, u := range v.users {
		uid, _ := strconv.Atoi(u.Details.Uid)
		isSys := uid > 0 && uid < 1000
		if !v.showSystem && isSys {
			continue
		}
		status := usermgmt.AccountStatus(u.Details.Username)
		isSudo := hasSudoGroup(u.Groups)
		items = append(items, userItem{user: u, status: status, isSudo: isSudo, isSys: isSys})
	}
	l := list.New(items, userDelegate{}, 0, 0)
	l.Title = "/ search users..."
	l.SetShowHelp(false)
	v.list = l
}

func hasSudoGroup(groups []*user.Group) bool {
	for _, g := range groups {
		if g != nil && (g.Name == "sudo" || g.Name == "wheel" || g.Name == "root") {
			return true
		}
	}
	return false
}

func (v *UserView) refresh(users []passwd.User) {
	v.users = users
	v.rebuildList()
}

func (v UserView) selectedUsername() string {
	if it := v.list.SelectedItem(); it != nil {
		return it.(userItem).user.Details.Username
	}
	return ""
}

func (v UserView) stats() (total, locked, expiring, sys int) {
	for _, item := range v.list.Items() {
		total++
		if ui, ok := item.(userItem); ok {
			switch {
			case ui.status == "locked":
				locked++
			case strings.HasPrefix(ui.status, "expires"):
				expiring++
			}
			if ui.isSys {
				sys++
			}
		}
	}
	return total, locked, expiring, sys
}

func (v UserView) update(msg tea.Msg) (UserView, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		v.width = msg.Width
		lw := listWidth(msg.Width)
		ls := listStyle(lw)
		h, vert := ls.GetFrameSize()
		ph := lipgloss.Height(v.list.Paginator.View())
		v.list.SetSize(lw-h, msg.Height-vert-ph-4)
		v.viewport = viewport.New(viewport.WithWidth(msg.Width-lw-4), viewport.WithHeight(msg.Height-5))
		v.refreshDetail()
	case tea.KeyPressMsg:
		switch msg.String() {
		case "s":
			v.showSystem = !v.showSystem
			v.rebuildList()
			v.refreshDetail()
			return v, nil
		case "1":
			v.detailTab = _tabOverview
			v.refreshDetail()
			return v, nil
		case "2":
			v.detailTab = _tabSSHKeys
			v.refreshDetail()
			return v, nil
		case "3":
			v.detailTab = _tabSudoRules
			v.refreshDetail()
			return v, nil
		case "4":
			v.detailTab = _tabActivity
			v.refreshDetail()
			return v, nil
		}
	}
	var cmd tea.Cmd
	v.list, cmd = v.list.Update(msg)
	// Refresh detail if selection changed
	if v.list.Index() != v.lastSelected {
		v.refreshDetail()
	}
	return v, cmd
}

func (v UserView) view() string {
	v.viewport.SetContent(v.cachedDetail)
	tabBar := renderTabBar(_detailTabs, int(v.detailTab))
	detailPanel := tabBar + v.viewport.View()
	return lipgloss.JoinHorizontal(lipgloss.Top, v.listView(), detailPanel)
}

func (v UserView) listView() string {
	v.list.Styles.Title = _listTitleStyle
	return listStyle(listWidth(v.width)).Render(v.list.View())
}

func (v *UserView) refreshDetail() {
	v.cachedDetail = v.computeDetail()
	v.lastSelected = v.list.Index()
}

func (v UserView) computeDetail() string {
	it := v.list.SelectedItem()
	if it == nil {
		return ""
	}
	ui := it.(userItem)

	switch v.detailTab {
	case _tabSSHKeys:
		return v.renderSSHKeys(ui.user)
	case _tabSudoRules:
		return v.renderSudoRules(ui.user)
	case _tabActivity:
		return v.renderActivity(ui.user)
	default:
		return v.renderOverview(ui)
	}
}

func (v UserView) renderOverview(ui userItem) string {
	u := ui.user
	var b strings.Builder
	w := v.viewport.Width()

	// Header line
	statusText := _badgeActiveStyle.Render("● active")
	switch {
	case ui.status == "locked":
		statusText = _badgeLockedStyle.Render("● locked")
	case strings.HasPrefix(ui.status, "expires"):
		statusText = _badgeExpiryStyle.Render("● " + ui.status)
	}
	fmt.Fprintf(&b, "%s  uid:%s gid:%s  %s\n", u.Details.Username, u.Details.Uid, u.Details.Gid, statusText)

	// Account status section
	b.WriteString("\n" + _dividerStyle.Render("── account status "+strings.Repeat("─", max(0, w-19))) + "\n")
	fmt.Fprintf(&b, "  status         %s\n", statusText)
	fmt.Fprintf(&b, "  last login     %s\n", usermgmt.LastLogin(u.Details.Username))
	fmt.Fprintf(&b, "  shell          %s\n", shellName(u))
	fmt.Fprintf(&b, "  home           %s\n", u.Details.HomeDir)

	// Password aging section
	b.WriteString("\n" + _dividerStyle.Render("── password aging "+strings.Repeat("─", max(0, w-19))) + "\n")
	aging, err := usermgmt.PasswordAging(u.Details.Username)
	if err == nil {
		for _, line := range strings.Split(aging, "\n") {
			fmt.Fprintf(&b, "  %s\n", line)
		}
	}

	// Groups section
	b.WriteString("\n" + _dividerStyle.Render("── groups "+strings.Repeat("─", max(0, w-11))) + "\n")
	b.WriteString("  " + renderGroupPills(u.Groups) + "\n")

	return _detailStyle.Render(b.String())
}

func (v UserView) renderSSHKeys(u passwd.User) string {
	var b strings.Builder
	b.WriteString(_headerStyle.Render("SSH Authorized Keys") + "\n\n")

	// Read authorized_keys
	home := u.Details.HomeDir
	out, err := exec.Command("cat", home+"/.ssh/authorized_keys").CombinedOutput()
	if err != nil {
		b.WriteString(_previewStyle.Render("  No authorized_keys found or not readable"))
		return _detailStyle.Render(b.String())
	}

	keys := strings.TrimSpace(string(out))
	if keys == "" {
		b.WriteString(_previewStyle.Render("  No keys configured"))
		return _detailStyle.Render(b.String())
	}

	for i, key := range strings.Split(keys, "\n") {
		parts := strings.Fields(key)
		keyType := ""
		comment := ""
		if len(parts) >= 1 {
			keyType = parts[0]
		}
		if len(parts) >= 3 {
			comment = parts[2]
		}

		// Get fingerprint
		fp := ""
		fpOut, fpErr := exec.Command("ssh-keygen", "-lf", "-", "-E", "sha256").Output()
		if fpErr == nil {
			fp = strings.TrimSpace(string(fpOut))
		}

		fmt.Fprintf(&b, "  %d. %s", i+1, _promptStyle.Render(keyType))
		if comment != "" {
			fmt.Fprintf(&b, "  %s", comment)
		}
		if fp != "" {
			fmt.Fprintf(&b, "\n     %s", _previewStyle.Render(fp))
		}
		b.WriteString("\n")
	}

	return _detailStyle.Render(b.String())
}

func (v UserView) renderSudoRules(u passwd.User) string {
	var b strings.Builder
	b.WriteString(_headerStyle.Render("Sudo Rules") + "\n\n")

	// Check /etc/sudoers and /etc/sudoers.d/* for this user
	username := u.Details.Username
	out, err := exec.Command("grep", "-rh", username, "/etc/sudoers", "/etc/sudoers.d/").CombinedOutput()
	if err != nil || strings.TrimSpace(string(out)) == "" {
		// Also check group-based rules
		var groupRules []string
		for _, g := range u.Groups {
			if g == nil {
				continue
			}
			gOut, gErr := exec.Command("grep", "-rh", "%"+g.Name, "/etc/sudoers", "/etc/sudoers.d/").CombinedOutput()
			if gErr == nil && strings.TrimSpace(string(gOut)) != "" {
				for _, line := range strings.Split(strings.TrimSpace(string(gOut)), "\n") {
					line = strings.TrimSpace(line)
					if line != "" && !strings.HasPrefix(line, "#") {
						groupRules = append(groupRules, fmt.Sprintf("  (via %s) %s", _badgeSudoStyle.Render(g.Name), line))
					}
				}
			}
		}
		if len(groupRules) == 0 {
			b.WriteString(_previewStyle.Render("  No sudo rules found for this user"))
		} else {
			b.WriteString(strings.Join(groupRules, "\n") + "\n")
		}
		return _detailStyle.Render(b.String())
	}

	for _, line := range strings.Split(strings.TrimSpace(string(out)), "\n") {
		line = strings.TrimSpace(line)
		if line != "" && !strings.HasPrefix(line, "#") {
			fmt.Fprintf(&b, "  %s\n", line)
		}
	}

	return _detailStyle.Render(b.String())
}

func (v UserView) renderActivity(u passwd.User) string {
	var b strings.Builder
	username := u.Details.Username
	b.WriteString(_headerStyle.Render("Recent Activity") + "\n\n")

	// Last logins (last -n 5)
	b.WriteString(_dividerStyle.Render("── recent logins") + "\n")
	out, err := exec.Command("last", "-n", "5", username).CombinedOutput()
	if err == nil {
		lines := strings.Split(strings.TrimSpace(string(out)), "\n")
		for _, line := range lines {
			if strings.TrimSpace(line) != "" && !strings.HasPrefix(line, "wtmp") {
				fmt.Fprintf(&b, "  %s\n", line)
			}
		}
		if len(lines) == 0 || (len(lines) == 1 && strings.TrimSpace(lines[0]) == "") {
			b.WriteString(_previewStyle.Render("  No login history") + "\n")
		}
	}

	// Failed logins (lastb -n 5, may need root)
	b.WriteString("\n" + _dividerStyle.Render("── failed logins") + "\n")
	out, err = exec.Command("lastb", "-n", "5", username).CombinedOutput()
	if err == nil && strings.TrimSpace(string(out)) != "" {
		lines := strings.Split(strings.TrimSpace(string(out)), "\n")
		for _, line := range lines {
			if strings.TrimSpace(line) != "" && !strings.HasPrefix(line, "btmp") {
				fmt.Fprintf(&b, "  %s\n", _badgeLockedStyle.Render(line))
			}
		}
	} else {
		b.WriteString(_previewStyle.Render("  No failed logins") + "\n")
	}

	// Recent sudo usage
	b.WriteString("\n" + _dividerStyle.Render("── sudo usage") + "\n")
	out, err = exec.Command("grep", username, "/var/log/auth.log").CombinedOutput()
	if err != nil {
		// Try journalctl as fallback
		out, err = exec.Command("journalctl", "_COMM=sudo", "-n", "5", "--no-pager", "-q").CombinedOutput()
	}
	if err == nil && strings.TrimSpace(string(out)) != "" {
		lines := strings.Split(strings.TrimSpace(string(out)), "\n")
		count := 5
		if len(lines) < count {
			count = len(lines)
		}
		for _, line := range lines[len(lines)-count:] {
			if strings.Contains(line, username) {
				fmt.Fprintf(&b, "  %s\n", strings.TrimSpace(line))
			}
		}
	} else {
		b.WriteString(_previewStyle.Render("  No sudo activity found") + "\n")
	}

	return _detailStyle.Render(b.String())
}

func shellName(u passwd.User) string {
	if u.Shell != "" {
		return u.Shell
	}
	return "unknown"
}

func renderGroupPills(groups []*user.Group) string {
	if len(groups) == 0 {
		return "No groups"
	}
	var pills []string
	for _, g := range groups {
		if g == nil {
			continue
		}
		switch g.Name {
		case "sudo", "wheel", "root":
			pills = append(pills, _badgeSudoStyle.Render("["+g.Name+"]"))
		case "docker", "lxd", "libvirt":
			pills = append(pills, lipgloss.NewStyle().Foreground(lipgloss.Color("#56b6c2")).Render("["+g.Name+"]"))
		default:
			pills = append(pills, "["+g.Name+"]")
		}
	}
	return strings.Join(pills, " ")
}
