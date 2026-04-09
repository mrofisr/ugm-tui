package group

import (
	"fmt"
	"os/user"
	"strings"

	"charm.land/bubbles/v2/table"
	"charm.land/lipgloss/v2"
	"github.com/ariasmn/ugm/internal/tui/common"
)

// View renders the group list and detail panel.
func (bg BubbleGroup) View() string {
	bg.viewport.SetContent(bg.detailView())

	return lipgloss.JoinHorizontal(
		lipgloss.Top, bg.listView(), bg.viewport.View())
}

func (bg BubbleGroup) listView() string {
	bg.list.Styles.Title = common.ListColorStyle

	return common.ListStyle.Render(bg.list.View())
}

func (bg BubbleGroup) detailView() string {
	builder := &strings.Builder{}
	vpWidth := bg.viewport.Width()
	divider := common.DividerStyle.Render(strings.Repeat("-", vpWidth)) + "\n"
	detailsHeader := common.HeaderStyle.Render("Details")
	memberOfHeader := common.HeaderStyle.Render("Current members")

	if it := bg.list.SelectedItem(); it != nil {
		builder.WriteString(detailsHeader)
		builder.WriteString(renderGroupDetails(it.(item).Details))
		builder.WriteString(divider)
		builder.WriteString(memberOfHeader)
		fmt.Fprintf(builder, "\n\n%s", renderUserTable(it.(item).Users))
	}
	details := common.WordWrap(builder.String(), vpWidth)

	return common.DetailStyle.Render(details)
}

func renderGroupDetails(group user.Group) string {
	gid := fmt.Sprintf("\n\nGID: %s\n", group.Gid)
	name := fmt.Sprintf("Name: %s\n", group.Name)

	return gid + name
}

func renderUserTable(users []*user.User) string {
	columns := []table.Column{
		{Title: "Username", Width: 16},
		{Title: "Fullname", Width: 20},
		{Title: "UID", Width: 10},
		{Title: "GID", Width: 10},
		{Title: "Home directory", Width: 25},
	}

	rows := []table.Row{}
	for _, u := range users {
		if u == nil {
			return "No users in this group"
		}
		rows = append(rows, table.Row{u.Username, u.Name, u.Uid, u.Gid, u.HomeDir})
	}

	t := table.New(
		table.WithColumns(columns),
		table.WithRows(rows),
		table.WithHeight(len(rows)+1),
	)
	s := table.DefaultStyles()
	s.Header = common.TableHeaderStyle
	s.Cell = common.TableMainStyle
	t.SetStyles(s)

	return t.View()
}
