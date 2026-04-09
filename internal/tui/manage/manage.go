// Package manage provides the user management TUI component.
package manage

import (
	"fmt"
	"strings"

	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/ariasmn/ugm/internal/tui/common"
	"github.com/ariasmn/ugm/usermgmt"
)

type action int

const (
	actionMenu action = iota
	actionCreateUser
	actionLockUser
	actionSetExpiry
)

// BubbleManage is the user management TUI model.
type BubbleManage struct {
	action       action
	menuIndex    int
	inputs       []textinput.Model
	focusedInput int
	authIsSSH    bool
	targetUser   string
	status       string
	done         bool
	width        int
	height       int
}

// DoneMsg signals that the management view is finished.
type DoneMsg struct{}

var menuItems = []string{
	"Create New User",
	"Lock User (revoke access)",
	"Set Expiry Date",
}

// InitialModel creates the initial management model.
func InitialModel() BubbleManage {
	return BubbleManage{action: actionMenu}
}

// SetTarget sets the user to manage and resets the view.
func (bm *BubbleManage) SetTarget(username string) {
	bm.targetUser = username
	bm.action = actionMenu
	bm.menuIndex = 0
	bm.status = ""
	bm.done = false
}

// Init implements tea.Model.
func (bm BubbleManage) Init() tea.Cmd { return nil }

// Update handles messages for the management view.
func (bm BubbleManage) Update(msg tea.Msg) (BubbleManage, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		bm.width = msg.Width
		bm.height = msg.Height
		return bm, nil
	case tea.KeyPressMsg:
		if msg.String() == "esc" {
			if bm.action != actionMenu {
				bm.action = actionMenu
				bm.status = ""
				return bm, nil
			}
			bm.done = true
			return bm, nil
		}
	}

	switch bm.action {
	case actionMenu:
		return bm.updateMenu(msg)
	case actionCreateUser:
		return bm.updateCreateUser(msg)
	case actionLockUser:
		return bm.updateLockUser(msg)
	case actionSetExpiry:
		return bm.updateSetExpiry(msg)
	}
	return bm, nil
}

// IsDone returns whether the user exited the management view.
func (bm BubbleManage) IsDone() bool { return bm.done }

// View renders the management view.
func (bm BubbleManage) View() string {
	var s string
	switch bm.action {
	case actionMenu:
		s = bm.viewMenu()
	case actionCreateUser:
		s = bm.viewCreateUser()
	case actionLockUser:
		s = bm.viewLockUser()
	case actionSetExpiry:
		s = bm.viewSetExpiry()
	}
	if bm.status != "" {
		s += "\n\n" + bm.status
	}
	return lipgloss.NewStyle().Padding(1, 2).Render(s)
}

// --- Menu ---

func (bm BubbleManage) updateMenu(msg tea.Msg) (BubbleManage, tea.Cmd) {
	if msg, ok := msg.(tea.KeyPressMsg); ok {
		switch msg.String() {
		case "up", "k":
			if bm.menuIndex > 0 {
				bm.menuIndex--
			}
		case "down", "j":
			if bm.menuIndex < len(menuItems)-1 {
				bm.menuIndex++
			}
		case "enter":
			bm.status = ""
			switch bm.menuIndex {
			case 0:
				bm.action = actionCreateUser
				bm.initCreateInputs()
			case 1:
				bm.action = actionLockUser
			case 2:
				bm.action = actionSetExpiry
				bm.initExpiryInput()
			}
		}
	}
	return bm, nil
}

func (bm BubbleManage) viewMenu() string {
	title := common.HeaderStyle.Render(fmt.Sprintf("Manage User: %s", bm.targetUser))
	var b strings.Builder
	b.WriteString(title + "\n\n")
	for i, item := range menuItems {
		cursor := "  "
		style := common.ListItemStyle
		if i == bm.menuIndex {
			cursor = "> "
			style = common.ListSelectedListItemStyle
		}
		b.WriteString(style.Render(cursor+item) + "\n")
	}
	b.WriteString("\n" + common.DividerStyle.Render("↑/↓ navigate • enter select • esc back"))
	return b.String()
}

// --- Create User ---

func (bm *BubbleManage) initCreateInputs() {
	bm.authIsSSH = false
	bm.focusedInput = 0
	bm.inputs = make([]textinput.Model, 4)

	bm.inputs[0] = textinput.New()
	bm.inputs[0].Placeholder = "username"
	bm.inputs[0].Focus()

	bm.inputs[1] = textinput.New()
	bm.inputs[1].Placeholder = "/bin/bash"
	bm.inputs[1].SetValue("/bin/bash")

	bm.inputs[2] = textinput.New()
	bm.inputs[2].Placeholder = "password or sshkey"

	bm.inputs[3] = textinput.New()
	bm.inputs[3].Placeholder = "password"
	bm.inputs[3].EchoMode = textinput.EchoPassword
}

func (bm BubbleManage) updateCreateUser(msg tea.Msg) (BubbleManage, tea.Cmd) {
	if msg, ok := msg.(tea.KeyPressMsg); ok {
		switch msg.String() {
		case "tab":
			if bm.focusedInput == 2 {
				bm.authIsSSH = !bm.authIsSSH
				if bm.authIsSSH {
					bm.inputs[3].Placeholder = "ssh-rsa AAAA..."
					bm.inputs[3].EchoMode = textinput.EchoNormal
				} else {
					bm.inputs[3].Placeholder = "password"
					bm.inputs[3].EchoMode = textinput.EchoPassword
				}
				return bm, nil
			}
			bm.focusedInput = (bm.focusedInput + 1) % 4
			for i := range bm.inputs {
				if i == bm.focusedInput {
					bm.inputs[i].Focus()
				} else {
					bm.inputs[i].Blur()
				}
			}
			return bm, nil
		case "enter":
			username := bm.inputs[0].Value()
			shell := bm.inputs[1].Value()
			secret := bm.inputs[3].Value()
			if username == "" {
				bm.status = common.ErrorStyle.Render("Username is required")
				return bm, nil
			}
			if err := usermgmt.CreateUser(username, shell); err != nil {
				bm.status = common.ErrorStyle.Render("Create failed: " + err.Error())
				return bm, nil
			}
			if secret != "" {
				if bm.authIsSSH {
					if err := usermgmt.SetSSHKey(username, secret); err != nil {
						bm.status = common.ErrorStyle.Render("SSH key failed: " + err.Error())
						return bm, nil
					}
				} else {
					if err := usermgmt.SetPassword(username, secret); err != nil {
						bm.status = common.ErrorStyle.Render("Password failed: " + err.Error())
						return bm, nil
					}
				}
			}
			bm.status = common.SuccessStyle.Render(fmt.Sprintf("User '%s' created successfully!", username))
			bm.action = actionMenu
			return bm, nil
		}
	}

	var cmd tea.Cmd
	bm.inputs[bm.focusedInput], cmd = bm.inputs[bm.focusedInput].Update(msg)
	return bm, cmd
}

func (bm BubbleManage) viewCreateUser() string {
	title := common.HeaderStyle.Render("Create New User")
	authLabel := "Auth: [Password]"
	if bm.authIsSSH {
		authLabel = "Auth: [SSH Key]"
	}

	labels := []string{"Username", "Shell", authLabel, "Secret"}
	var b strings.Builder
	b.WriteString(title + "\n\n")
	for i, input := range bm.inputs {
		cursor := "  "
		if i == bm.focusedInput {
			cursor = "> "
		}
		fmt.Fprintf(&b, "%s%s: %s\n", cursor, labels[i], input.View())
	}
	b.WriteString("\n" + common.DividerStyle.Render("tab next field (on Auth: toggle mode) • enter submit • esc back"))
	return b.String()
}

// --- Lock User ---

func (bm BubbleManage) updateLockUser(msg tea.Msg) (BubbleManage, tea.Cmd) {
	if msg, ok := msg.(tea.KeyPressMsg); ok {
		switch msg.String() {
		case "y":
			if err := usermgmt.LockUser(bm.targetUser); err != nil {
				bm.status = common.ErrorStyle.Render("Lock failed: " + err.Error())
			} else {
				bm.status = common.SuccessStyle.Render(fmt.Sprintf("User '%s' locked!", bm.targetUser))
			}
			bm.action = actionMenu
		case "n":
			bm.action = actionMenu
		}
	}
	return bm, nil
}

func (bm BubbleManage) viewLockUser() string {
	title := common.HeaderStyle.Render("Lock User")
	return fmt.Sprintf("%s\n\n%s\n\n%s",
		title,
		common.PromptStyle.Render(fmt.Sprintf("Lock user '%s'? This will disable login.", bm.targetUser)),
		common.DividerStyle.Render("y confirm • n cancel • esc back"),
	)
}

// --- Set Expiry ---

func (bm *BubbleManage) initExpiryInput() {
	bm.focusedInput = 0
	bm.inputs = make([]textinput.Model, 1)
	bm.inputs[0] = textinput.New()
	bm.inputs[0].Placeholder = "YYYY-MM-DD"
	bm.inputs[0].Focus()
}

func (bm BubbleManage) updateSetExpiry(msg tea.Msg) (BubbleManage, tea.Cmd) {
	if msg, ok := msg.(tea.KeyPressMsg); ok {
		if msg.String() == "enter" {
			date := bm.inputs[0].Value()
			if date == "" {
				bm.status = common.ErrorStyle.Render("Date is required (YYYY-MM-DD)")
				return bm, nil
			}
			if err := usermgmt.SetExpiry(bm.targetUser, date); err != nil {
				bm.status = common.ErrorStyle.Render("Set expiry failed: " + err.Error())
			} else {
				bm.status = common.SuccessStyle.Render(fmt.Sprintf("Expiry for '%s' set to %s", bm.targetUser, date))
			}
			bm.action = actionMenu
			return bm, nil
		}
	}

	var cmd tea.Cmd
	bm.inputs[0], cmd = bm.inputs[0].Update(msg)
	return bm, cmd
}

func (bm BubbleManage) viewSetExpiry() string {
	title := common.HeaderStyle.Render("Set Expiry Date")
	return fmt.Sprintf("%s\n\n  User: %s\n  Date: %s\n\n%s",
		title,
		common.PromptStyle.Render(bm.targetUser),
		bm.inputs[0].View(),
		common.DividerStyle.Render("enter submit • esc back"),
	)
}
