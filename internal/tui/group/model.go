// Package group provides the group list TUI component.
package group

import (
	"charm.land/bubbles/v2/list"
	"charm.land/bubbles/v2/viewport"
	"github.com/ariasmn/ugm/groupparser"
)

type item groupparser.Group

func (i item) FilterValue() string { return i.Details.Name }

// BubbleGroup is the group list TUI model.
type BubbleGroup struct {
	list     list.Model
	viewport viewport.Model
}

// InitialModel creates the initial group list model.
func InitialModel() BubbleGroup {
	items := groupToItem(groupparser.GetGroups())
	l := list.New(items, itemDelegate{}, 0, 0)
	l.Title = "Groups"
	l.SetShowHelp(false)

	return BubbleGroup{list: l}
}

func groupToItem(groups []groupparser.Group) []list.Item {
	items := make([]list.Item, len(groups))
	for i, v := range groups {
		items[i] = item(v)
	}
	return items
}
