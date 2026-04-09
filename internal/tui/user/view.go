package user

import (
	"fmt"
	"os/user"
	"strings"

	"charm.land/bubbles/v2/table"
	"charm.land/lipgloss/v2"
	"github.com/ariasmn/ugm/internal/tui/common"
)

// View renders the user list and detail panel.
func (bu BubbleUser) View() string {
	bu.viewport.SetContent(bu.detailView())

	return lipgloss.JoinHorizontal(
		lipgloss.Top, bu.listView(), bu.viewport.View())
}

func (bu BubbleUser) listView() string {
	bu.list.Styles.Title = common.ListColorStyle

	return common.ListStyle.Render(bu.list.View())
}

func (bu BubbleUser) detailView() string {
	builder := &strings.Builder{}
	vpWidth := bu.viewport.Width()
	divider := common.DividerStyle.Render(strings.Repeat("-", vpWidth)) + "\n"
	detailsHeader := common.HeaderStyle.Render("Details")
	memberOfHeader := common.HeaderStyle.Render("Member of")

	if it := bu.list.SelectedItem(); it != nil {
		builder.WriteString(detailsHeader)
		builder.WriteString(renderUserDetails(it.(item).Details))
		builder.WriteString(divider)
		builder.WriteString(memberOfHeader)
		fmt.Fprintf(builder, "\n\n%s", renderGroupTable(it.(item).Groups))
	}
	details := common.WordWrap(builder.String(), vpWidth)

	return common.DetailStyle.Render(details)
}

func renderUserDetails(u user.User) string {
	username := fmt.Sprintf("\n\nUsername: %s\n", u.Username)
	fullname := fmt.Sprintf("Fullname: %s\n", u.Name)
	identificators := fmt.Sprintf("UID: %s\nGID: %s\n", u.Uid, u.Gid)
	homeDirectory := fmt.Sprintf("Home directory: %s\n", u.HomeDir)

	return username + fullname + identificators + homeDirectory
}

func renderGroupTable(groups []*user.Group) string {
	columns := []table.Column{
		{Title: "GID", Width: 10},
		{Title: "Name", Width: 16},
	}

	rows := []table.Row{}
	for _, group := range groups {
		if group == nil {
			continue
		}
		rows = append(rows, table.Row{group.Gid, group.Name})
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
