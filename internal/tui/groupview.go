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

	"github.com/mrofisr/ugm-tui/internal/group"
)

// GroupView displays the list of system groups and their details.
type GroupView struct {
	list     list.Model
	viewport viewport.Model
	width    int
}

type groupItem group.Group

func (i groupItem) FilterValue() string { return i.Details.Name }

type groupDelegate struct{}

func (d groupDelegate) Height() int                             { return 1 }
func (d groupDelegate) Spacing() int                            { return 0 }
func (d groupDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }

func (d groupDelegate) Render(w io.Writer, m list.Model, index int, item list.Item) {
	g, ok := item.(groupItem)
	if !ok {
		return
	}
	line := g.Details.Name
	if index == m.Index() {
		line = _listSelectedStyle.Render("> " + line)
	} else {
		line = _listItemStyle.Render(line)
	}
	_, _ = fmt.Fprint(w, line)
}

func newGroupView(groups []group.Group) GroupView {
	items := make([]list.Item, len(groups))
	for i, g := range groups {
		items[i] = groupItem(g)
	}
	l := list.New(items, groupDelegate{}, 0, 0)
	l.Title = "Groups"
	l.SetShowHelp(false)

	return GroupView{list: l}
}

func (v GroupView) update(msg tea.Msg) (GroupView, tea.Cmd) {
	if msg, ok := msg.(tea.WindowSizeMsg); ok {
		v.width = msg.Width
		lw := listWidth(msg.Width)
		ls := listStyle(lw)
		h, vert := ls.GetFrameSize()
		ph := lipgloss.Height(v.list.Paginator.View())
		v.list.SetSize(lw-h, msg.Height-vert-ph-2)
		v.viewport = viewport.New(viewport.WithWidth(msg.Width-lw-4), viewport.WithHeight(msg.Height-3))
		v.viewport.SetContent(v.detailView())
	}
	var cmd tea.Cmd
	v.list, cmd = v.list.Update(msg)
	return v, cmd
}

func (v GroupView) view() string {
	v.viewport.SetContent(v.detailView())
	return lipgloss.JoinHorizontal(lipgloss.Top, v.listView(), v.viewport.View())
}

func (v GroupView) listView() string {
	v.list.Styles.Title = _listTitleStyle
	return listStyle(listWidth(v.width)).Render(v.list.View())
}

func (v GroupView) detailView() string {
	var b strings.Builder
	w := v.viewport.Width()

	if it := v.list.SelectedItem(); it != nil {
		g := it.(groupItem)
		b.WriteString(_headerStyle.Render("Details"))
		fmt.Fprintf(&b, "\n\nGID: %s\n", g.Details.Gid)
		fmt.Fprintf(&b, "Name: %s\n", g.Details.Name)
		b.WriteString(_dividerStyle.Render(strings.Repeat("-", w)) + "\n")
		b.WriteString(_headerStyle.Render("Current members"))
		fmt.Fprintf(&b, "\n\n%s", renderUserTable(g.Users))
	}

	return _detailStyle.Render(wordWrap(b.String(), w))
}

func renderUserTable(users []*user.User) string {
	cols := []table.Column{
		{Title: "Username", Width: 16},
		{Title: "Fullname", Width: 20},
		{Title: "UID", Width: 10},
		{Title: "GID", Width: 10},
		{Title: "Home directory", Width: 25},
	}
	var rows []table.Row
	for _, u := range users {
		if u == nil {
			return "No users in this group"
		}
		rows = append(rows, table.Row{u.Username, u.Name, u.Uid, u.Gid, u.HomeDir})
	}
	t := table.New(table.WithColumns(cols), table.WithRows(rows), table.WithHeight(len(rows)+1))
	s := table.DefaultStyles()
	s.Header = _tableHeaderStyle
	s.Cell = _tableStyle
	t.SetStyles(s)
	return t.View()
}
