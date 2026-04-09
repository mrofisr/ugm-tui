package user

import (
	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/ariasmn/ugm/internal/tui/common"
)

// Update handles messages for the user view.
func (bu BubbleUser) Update(msg tea.Msg) (BubbleUser, tea.Cmd) {
	if msg, ok := msg.(tea.WindowSizeMsg); ok {
		horizontal, vertical := common.ListStyle.GetFrameSize()
		paginatorHeight := lipgloss.Height(bu.list.Paginator.View())

		bu.list.SetSize(msg.Width-horizontal, msg.Height-vertical-paginatorHeight)
		bu.viewport = viewport.New(viewport.WithWidth(msg.Width), viewport.WithHeight(msg.Height))
		bu.viewport.SetContent(bu.detailView())
	}

	var cmd tea.Cmd
	bu.list, cmd = bu.list.Update(msg)

	return bu, cmd
}
