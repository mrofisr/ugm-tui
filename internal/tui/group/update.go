package group

import (
	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/ariasmn/ugm/internal/tui/common"
)

// Update handles messages for the group view.
func (bg BubbleGroup) Update(msg tea.Msg) (BubbleGroup, tea.Cmd) {
	if msg, ok := msg.(tea.WindowSizeMsg); ok {
		horizontal, vertical := common.ListStyle.GetFrameSize()
		paginatorHeight := lipgloss.Height(bg.list.Paginator.View())

		bg.list.SetSize(msg.Width-horizontal, msg.Height-vertical-paginatorHeight)
		bg.viewport = viewport.New(viewport.WithWidth(msg.Width), viewport.WithHeight(msg.Height))
		bg.viewport.SetContent(bg.detailView())
	}

	var cmd tea.Cmd
	bg.list, cmd = bg.list.Update(msg)

	return bg, cmd
}
