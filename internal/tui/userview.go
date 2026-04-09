package tui

import (
	"fmt"
	"io"
	"os/user"
	"strings"

	"charm.land/bubbles/v2/list"
	"charm.land/bubbles/v2/table"
	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/mrofisr/ugm-tui/internal/passwd"
)

// UserView displays the list of system users and their details.
type UserView struct {
	list     list.Model
	viewport viewport.Model
	users    []passwd.User
}

type userItem passwd.User

func (i userItem) FilterValue() string { return i.Details.Username }

type userDelegate struct{}

func (d userDelegate) Height() int                             { return 1 }
func (d userDelegate) Spacing() int                            { return 0 }
func (d userDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }

func (d userDelegate) Render(w io.Writer, m list.Model, index int, item list.Item) {
	u, ok := item.(userItem)
	if !ok {
		return
	}
	line := u.Details.Username
	if index == m.Index() {
		line = _listSelectedStyle.Render("> " + line)
	} else {
		line = _listItemStyle.Render(line)
	}
	_, _ = fmt.Fprint(w, line)
}

func newUserView(users []passwd.User) UserView {
	items := make([]list.Item, len(users))
	for i, u := range users {
		items[i] = userItem(u)
	}
	l := list.New(items, userDelegate{}, 0, 0)
	l.Title = "Users"
	l.SetShowHelp(false)

	return UserView{list: l, users: users}
}

func (v *UserView) refresh(users []passwd.User) {
	v.users = users
	items := make([]list.Item, len(users))
	for i, u := range users {
		items[i] = userItem(u)
	}
	v.list.SetItems(items)
}

func (v UserView) selectedUsername() string {
	if it := v.list.SelectedItem(); it != nil {
		return it.(userItem).Details.Username
	}
	return ""
}

func (v UserView) update(msg tea.Msg) (UserView, tea.Cmd) {
	if msg, ok := msg.(tea.WindowSizeMsg); ok {
		h, vert := _listStyle.GetFrameSize()
		ph := lipgloss.Height(v.list.Paginator.View())
		v.list.SetSize(msg.Width-h, msg.Height-vert-ph)
		v.viewport = viewport.New(viewport.WithWidth(msg.Width), viewport.WithHeight(msg.Height))
		v.viewport.SetContent(v.detailView())
	}
	var cmd tea.Cmd
	v.list, cmd = v.list.Update(msg)
	return v, cmd
}

func (v UserView) view() string {
	v.viewport.SetContent(v.detailView())
	return lipgloss.JoinHorizontal(lipgloss.Top, v.listView(), v.viewport.View())
}

func (v UserView) listView() string {
	v.list.Styles.Title = _listTitleStyle
	return _listStyle.Render(v.list.View())
}

func (v UserView) detailView() string {
	var b strings.Builder
	w := v.viewport.Width()

	if it := v.list.SelectedItem(); it != nil {
		u := it.(userItem)
		b.WriteString(_headerStyle.Render("Details"))
		fmt.Fprintf(&b, "\n\nUsername: %s\n", u.Details.Username)
		fmt.Fprintf(&b, "Fullname: %s\n", u.Details.Name)
		fmt.Fprintf(&b, "UID: %s\nGID: %s\n", u.Details.Uid, u.Details.Gid)
		fmt.Fprintf(&b, "Home directory: %s\n", u.Details.HomeDir)
		b.WriteString(_dividerStyle.Render(strings.Repeat("-", w)) + "\n")
		b.WriteString(_headerStyle.Render("Member of"))
		fmt.Fprintf(&b, "\n\n%s", renderGroupTable(u.Groups))
	}

	return _detailStyle.Render(wordWrap(b.String(), w))
}

func renderGroupTable(groups []*user.Group) string {
	cols := []table.Column{
		{Title: "GID", Width: 10},
		{Title: "Name", Width: 16},
	}
	var rows []table.Row
	for _, g := range groups {
		if g == nil {
			continue
		}
		rows = append(rows, table.Row{g.Gid, g.Name})
	}
	t := table.New(table.WithColumns(cols), table.WithRows(rows), table.WithHeight(len(rows)+1))
	s := table.DefaultStyles()
	s.Header = _tableHeaderStyle
	s.Cell = _tableStyle
	t.SetStyles(s)
	return t.View()
}
