package tui

import (
	"fmt"
	"strings"

	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/mrofisr/ugm-tui/internal/usermgmt"
)

type manageAction int

const (
	_actionMenu manageAction = iota
	_actionCreate
	_actionDelete
	_actionLock
	_actionUnlock
	_actionExpiry
	_actionAddGroup
	_actionRemoveGroup
	_actionPasswdAging
)

// ManageView provides user management actions.
type ManageView struct {
	action     manageAction
	menuIndex  int
	inputs     []textinput.Model
	focusIdx   int
	authIsSSH  bool
	targetUser string
	status     string
	infoText   string
	done       bool
}

var _menuItems = []string{
	"Create New User",
	"Delete User",
	"Lock User (revoke access)",
	"Unlock User",
	"Set Expiry Date",
	"Add to Group (assign role)",
	"Remove from Group",
	"View Password Aging Info",
}

func newManageView() ManageView {
	return ManageView{action: _actionMenu}
}

func (v *ManageView) setTarget(username string) {
	v.targetUser = username
	v.action = _actionMenu
	v.menuIndex = 0
	v.status = ""
	v.infoText = ""
	v.done = false
}

func (v ManageView) update(msg tea.Msg) (ManageView, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		return v, nil
	case tea.KeyPressMsg:
		if msg.String() == "esc" {
			if v.action != _actionMenu {
				v.action = _actionMenu
				v.status = ""
				v.infoText = ""
				return v, nil
			}
			v.done = true
			return v, nil
		}
	}

	switch v.action {
	case _actionMenu:
		return v.updateMenu(msg)
	case _actionCreate:
		return v.updateCreate(msg)
	case _actionDelete:
		return v.updateConfirm(msg, "delete", func() error { return usermgmt.DeleteUser(v.targetUser) })
	case _actionLock:
		return v.updateConfirm(msg, "lock", func() error { return usermgmt.LockUser(v.targetUser) })
	case _actionUnlock:
		return v.updateConfirm(msg, "unlock", func() error { return usermgmt.UnlockUser(v.targetUser) })
	case _actionExpiry:
		return v.updateExpiry(msg)
	case _actionAddGroup:
		return v.updateGroupInput(msg, "add", usermgmt.AddToGroup)
	case _actionRemoveGroup:
		return v.updateGroupInput(msg, "remove", usermgmt.RemoveFromGroup)
	case _actionPasswdAging:
		// read-only view, any key goes back
		if _, ok := msg.(tea.KeyPressMsg); ok {
			v.action = _actionMenu
			v.infoText = ""
		}
		return v, nil
	}
	return v, nil
}

func (v ManageView) view() string {
	var s string
	switch v.action {
	case _actionMenu:
		s = v.viewMenu()
	case _actionCreate:
		s = v.viewCreate()
	case _actionDelete:
		s = v.viewConfirmPrompt("Delete User", fmt.Sprintf("Delete user '%s' and their home directory?", v.targetUser))
	case _actionLock:
		s = v.viewConfirmPrompt("Lock User", fmt.Sprintf("Lock user '%s'? This will disable login.", v.targetUser))
	case _actionUnlock:
		s = v.viewConfirmPrompt("Unlock User", fmt.Sprintf("Unlock user '%s'?", v.targetUser))
	case _actionExpiry:
		s = v.viewExpiry()
	case _actionAddGroup:
		s = v.viewGroupInput("Add to Group")
	case _actionRemoveGroup:
		s = v.viewGroupInput("Remove from Group")
	case _actionPasswdAging:
		s = v.viewInfo("Password Aging Info")
	}
	if v.status != "" {
		s += "\n\n" + v.status
	}
	return lipgloss.NewStyle().Padding(1, 2).Render(s)
}

// --- Menu ---

func (v ManageView) updateMenu(msg tea.Msg) (ManageView, tea.Cmd) {
	if msg, ok := msg.(tea.KeyPressMsg); ok {
		switch msg.String() {
		case "up", "k":
			if v.menuIndex > 0 {
				v.menuIndex--
			}
		case "down", "j":
			if v.menuIndex < len(_menuItems)-1 {
				v.menuIndex++
			}
		case "enter":
			v.status = ""
			v.infoText = ""
			switch v.menuIndex {
			case 0:
				v.action = _actionCreate
				v.initCreateInputs()
			case 1:
				v.action = _actionDelete
			case 2:
				v.action = _actionLock
			case 3:
				v.action = _actionUnlock
			case 4:
				v.action = _actionExpiry
				v.initExpiryInput()
			case 5:
				v.action = _actionAddGroup
				v.initGroupInput()
			case 6:
				v.action = _actionRemoveGroup
				v.initGroupInput()
			case 7:
				v.action = _actionPasswdAging
				info, err := usermgmt.PasswordAging(v.targetUser)
				if err != nil {
					v.status = _errorStyle.Render(err.Error())
					v.action = _actionMenu
				} else {
					v.infoText = info
				}
			}
		}
	}
	return v, nil
}

func (v ManageView) viewMenu() string {
	title := _headerStyle.Render(fmt.Sprintf("Manage User: %s", v.targetUser))
	var b strings.Builder
	b.WriteString(title + "\n\n")
	for i, item := range _menuItems {
		cursor := "  "
		style := _listItemStyle
		if i == v.menuIndex {
			cursor = "> "
			style = _listSelectedStyle
		}
		b.WriteString(style.Render(cursor+item) + "\n")
	}
	b.WriteString("\n" + _dividerStyle.Render("↑/↓ navigate • enter select • esc back"))
	return b.String()
}

// --- Generic confirm (y/n) ---

func (v ManageView) updateConfirm(msg tea.Msg, verb string, fn func() error) (ManageView, tea.Cmd) {
	if msg, ok := msg.(tea.KeyPressMsg); ok {
		switch msg.String() {
		case "y":
			if err := fn(); err != nil {
				v.status = _errorStyle.Render(fmt.Sprintf("%s failed: %s", verb, err))
			} else {
				v.status = _successStyle.Render(fmt.Sprintf("User '%s' %sed!", v.targetUser, verb))
			}
			v.action = _actionMenu
		case "n":
			v.action = _actionMenu
		}
	}
	return v, nil
}

func (v ManageView) viewConfirmPrompt(title, prompt string) string {
	return fmt.Sprintf("%s\n\n%s\n\n%s",
		_headerStyle.Render(title),
		_promptStyle.Render(prompt),
		_dividerStyle.Render("y confirm • n cancel • esc back"),
	)
}

// --- Create User ---

func (v *ManageView) initCreateInputs() {
	v.authIsSSH = false
	v.focusIdx = 0
	v.inputs = make([]textinput.Model, 4)

	v.inputs[0] = textinput.New()
	v.inputs[0].Placeholder = "username"
	v.inputs[0].Focus()

	v.inputs[1] = textinput.New()
	v.inputs[1].Placeholder = "/bin/bash"
	v.inputs[1].SetValue("/bin/bash")

	v.inputs[2] = textinput.New()
	v.inputs[2].Placeholder = "password or sshkey"

	v.inputs[3] = textinput.New()
	v.inputs[3].Placeholder = "password"
	v.inputs[3].EchoMode = textinput.EchoPassword
}

func (v ManageView) updateCreate(msg tea.Msg) (ManageView, tea.Cmd) {
	if msg, ok := msg.(tea.KeyPressMsg); ok {
		switch msg.String() {
		case "tab":
			if v.focusIdx == 2 {
				v.authIsSSH = !v.authIsSSH
				if v.authIsSSH {
					v.inputs[3].Placeholder = "ssh-rsa AAAA..."
					v.inputs[3].EchoMode = textinput.EchoNormal
				} else {
					v.inputs[3].Placeholder = "password"
					v.inputs[3].EchoMode = textinput.EchoPassword
				}
				return v, nil
			}
			v.focusIdx = (v.focusIdx + 1) % 4
			for i := range v.inputs {
				if i == v.focusIdx {
					v.inputs[i].Focus()
				} else {
					v.inputs[i].Blur()
				}
			}
			return v, nil
		case "enter":
			return v.submitCreate()
		}
	}

	var cmd tea.Cmd
	v.inputs[v.focusIdx], cmd = v.inputs[v.focusIdx].Update(msg)
	return v, cmd
}

func (v ManageView) submitCreate() (ManageView, tea.Cmd) {
	username := v.inputs[0].Value()
	shell := v.inputs[1].Value()
	secret := v.inputs[3].Value()

	if username == "" {
		v.status = _errorStyle.Render("Username is required")
		return v, nil
	}
	if err := usermgmt.CreateUser(username, shell); err != nil {
		v.status = _errorStyle.Render("Create failed: " + err.Error())
		return v, nil
	}
	if secret != "" {
		var err error
		if v.authIsSSH {
			err = usermgmt.SetSSHKey(username, secret)
		} else {
			err = usermgmt.SetPassword(username, secret)
		}
		if err != nil {
			v.status = _errorStyle.Render("Auth setup failed: " + err.Error())
			return v, nil
		}
	}
	v.status = _successStyle.Render(fmt.Sprintf("User '%s' created successfully!", username))
	v.action = _actionMenu
	return v, nil
}

func (v ManageView) viewCreate() string {
	title := _headerStyle.Render("Create New User")
	authLabel := "Auth: [Password]"
	if v.authIsSSH {
		authLabel = "Auth: [SSH Key]"
	}

	labels := []string{"Username", "Shell", authLabel, "Secret"}
	var b strings.Builder
	b.WriteString(title + "\n\n")
	for i, input := range v.inputs {
		cursor := "  "
		if i == v.focusIdx {
			cursor = "> "
		}
		fmt.Fprintf(&b, "%s%s: %s\n", cursor, labels[i], input.View())
	}
	b.WriteString("\n" + _dividerStyle.Render("tab next field (on Auth: toggle mode) • enter submit • esc back"))
	return b.String()
}

// --- Set Expiry ---

func (v *ManageView) initExpiryInput() {
	v.focusIdx = 0
	v.inputs = make([]textinput.Model, 1)
	v.inputs[0] = textinput.New()
	v.inputs[0].Placeholder = "YYYY-MM-DD"
	v.inputs[0].Focus()
}

func (v ManageView) updateExpiry(msg tea.Msg) (ManageView, tea.Cmd) {
	if msg, ok := msg.(tea.KeyPressMsg); ok {
		if msg.String() == "enter" {
			date := v.inputs[0].Value()
			if date == "" {
				v.status = _errorStyle.Render("Date is required (YYYY-MM-DD)")
				return v, nil
			}
			if err := usermgmt.SetExpiry(v.targetUser, date); err != nil {
				v.status = _errorStyle.Render("Set expiry failed: " + err.Error())
			} else {
				v.status = _successStyle.Render(fmt.Sprintf("Expiry for '%s' set to %s", v.targetUser, date))
			}
			v.action = _actionMenu
			return v, nil
		}
	}

	var cmd tea.Cmd
	v.inputs[0], cmd = v.inputs[0].Update(msg)
	return v, cmd
}

func (v ManageView) viewExpiry() string {
	return fmt.Sprintf("%s\n\n  User: %s\n  Date: %s\n\n%s",
		_headerStyle.Render("Set Expiry Date"),
		_promptStyle.Render(v.targetUser),
		v.inputs[0].View(),
		_dividerStyle.Render("enter submit • esc back"),
	)
}

// --- Add/Remove Group ---

func (v *ManageView) initGroupInput() {
	v.focusIdx = 0
	v.inputs = make([]textinput.Model, 1)
	v.inputs[0] = textinput.New()
	v.inputs[0].Placeholder = "group name (e.g. docker, sudo)"
	v.inputs[0].Focus()
}

func (v ManageView) updateGroupInput(msg tea.Msg, verb string, fn func(string, string) error) (ManageView, tea.Cmd) {
	if msg, ok := msg.(tea.KeyPressMsg); ok {
		if msg.String() == "enter" {
			grp := v.inputs[0].Value()
			if grp == "" {
				v.status = _errorStyle.Render("Group name is required")
				return v, nil
			}
			if err := fn(v.targetUser, grp); err != nil {
				v.status = _errorStyle.Render(fmt.Sprintf("%s group failed: %s", verb, err))
			} else {
				v.status = _successStyle.Render(fmt.Sprintf("User '%s' %sed group '%s'", v.targetUser, verb, grp))
			}
			v.action = _actionMenu
			return v, nil
		}
	}

	var cmd tea.Cmd
	v.inputs[0], cmd = v.inputs[0].Update(msg)
	return v, cmd
}

func (v ManageView) viewGroupInput(title string) string {
	return fmt.Sprintf("%s\n\n  User: %s\n  Group: %s\n\n%s",
		_headerStyle.Render(title),
		_promptStyle.Render(v.targetUser),
		v.inputs[0].View(),
		_dividerStyle.Render("enter submit • esc back"),
	)
}

// --- Info view (read-only) ---

func (v ManageView) viewInfo(title string) string {
	return fmt.Sprintf("%s\n\n  User: %s\n\n%s\n\n%s",
		_headerStyle.Render(title),
		_promptStyle.Render(v.targetUser),
		v.infoText,
		_dividerStyle.Render("any key to go back"),
	)
}
