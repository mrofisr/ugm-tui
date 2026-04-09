// Package user provides the user list TUI component.
package user

import (
	"charm.land/bubbles/v2/list"
	"charm.land/bubbles/v2/viewport"
	"github.com/ariasmn/ugm/userparser"
)

type item userparser.User

func (i item) FilterValue() string { return i.Details.Username }

// BubbleUser is the user list TUI model.
type BubbleUser struct {
	list     list.Model
	viewport viewport.Model
}

// InitialModel creates the initial user list model.
func InitialModel() BubbleUser {
	items := userToItem(userparser.GetUsers())
	l := list.New(items, itemDelegate{}, 0, 0)
	l.Title = "Users"
	l.SetShowHelp(false)

	return BubbleUser{list: l}
}

// RefreshUsers re-parses /etc/passwd and updates the list.
func (bu *BubbleUser) RefreshUsers() {
	userparser.ParseUsers("/etc/passwd")
	items := userToItem(userparser.GetUsers())
	bu.list.SetItems(items)
}

// SelectedUsername returns the username of the currently selected item.
func (bu BubbleUser) SelectedUsername() string {
	if it := bu.list.SelectedItem(); it != nil {
		return it.(item).Details.Username
	}
	return ""
}

func userToItem(users []userparser.User) []list.Item {
	items := make([]list.Item, len(users))
	for i, v := range users {
		items[i] = item(v)
	}
	return items
}
