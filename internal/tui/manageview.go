package tui

import (
	"fmt"
	"strings"

	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/mrofisr/ugm-tui/internal/audit"
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
	_actionCreateGroup
	_actionDeleteGroup
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
	"Create Group (new role)",
	"Delete Group",
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
		return v.updateConfirm(msg, "delete", "userdel -r "+v.targetUser, func() error { return usermgmt.DeleteUser(v.targetUser) })
	case _actionLock:
		return v.updateConfirm(msg, "lock", "usermod --lock "+v.targetUser, func() error { return usermgmt.LockUser(v.targetUser) })
	case _actionUnlock:
		return v.updateConfirm(msg, "unlock", "usermod --unlock "+v.targetUser, func() error { return usermgmt.UnlockUser(v.targetUser) })
	case _actionExpiry:
		return v.updateExpiry(msg)
	case _actionAddGroup:
		return v.updateGroupInput(msg, "add", usermgmt.AddToGroup)
	case _actionRemoveGroup:
		return v.updateGroupInput(msg, "remove", usermgmt.RemoveFromGroup)
	case _actionPasswdAging:
		if _, ok := msg.(tea.KeyPressMsg); ok {
			v.action = _actionMenu
			v.infoText = ""
		}
		return v, nil
	case _actionCreateGroup:
		return v.updateSingleInput(msg, "create-group", usermgmt.CreateGroup)
	case _actionDeleteGroup:
		return v.updateSingleInput(msg, "delete-group", usermgmt.DeleteGroup)
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
		s = v.viewConfirmPrompt("Delete User",
			fmt.Sprintf("Delete user '%s' and their home directory?", v.targetUser),
			"userdel -r "+v.targetUser)
	case _actionLock:
		s = v.viewConfirmPrompt("Lock User",
			fmt.Sprintf("Lock user '%s'? This will disable login.", v.targetUser),
			"usermod --lock "+v.targetUser)
	case _actionUnlock:
		s = v.viewConfirmPrompt("Unlock User",
			fmt.Sprintf("Unlock user '%s'?", v.targetUser),
			"usermod --unlock "+v.targetUser)
	case _actionExpiry:
		s = v.viewExpiry()
	case _actionAddGroup:
		s = v.viewGroupInput("Add to Group")
	case _actionRemoveGroup:
		s = v.viewGroupInput("Remove from Group")
	case _actionPasswdAging:
		s = v.viewInfo("Password Aging Info")
	case _actionCreateGroup:
		s = v.viewSingleInput("Create Group", "group name")
	case _actionDeleteGroup:
		s = v.viewSingleInput("Delete Group", "group name")
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
			case 8:
				v.action = _actionCreateGroup
				v.initSingleInput("e.g. devops, docker, deploy")
			case 9:
				v.action = _actionDeleteGroup
				v.initSingleInput("group to delete")
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

// --- Generic confirm (y/n) with command preview ---

func (v ManageView) updateConfirm(msg tea.Msg, verb, cmdPreview string, fn func() error) (ManageView, tea.Cmd) {
	if msg, ok := msg.(tea.KeyPressMsg); ok {
		switch msg.String() {
		case "y":
			if err := fn(); err != nil {
				v.status = _errorStyle.Render(fmt.Sprintf("%s failed: %s", verb, err))
			} else {
				v.status = _successStyle.Render(fmt.Sprintf("User '%s' %sed!", v.targetUser, verb))
				audit.Log(verb, v.targetUser, cmdPreview)
			}
			v.action = _actionMenu
		case "n":
			v.action = _actionMenu
		}
	}
	return v, nil
}

func (v ManageView) viewConfirmPrompt(title, prompt, cmdPreview string) string {
	return fmt.Sprintf("%s\n\n%s\n\n%s\n\n%s",
		_headerStyle.Render(title),
		_promptStyle.Render(prompt),
		_previewStyle.Render("$ "+cmdPreview),
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
	audit.Log("create-user", username, fmt.Sprintf("useradd -m -s %s %s", shell, username))

	if secret != "" {
		var err error
		if v.authIsSSH {
			err = usermgmt.SetSSHKey(username, secret)
			if err == nil {
				audit.Log("set-ssh-key", username, "wrote authorized_keys")
			}
		} else {
			err = usermgmt.SetPassword(username, secret)
			if err == nil {
				audit.Log("set-password", username, "chpasswd")
			}
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

	username := v.inputs[0].Value()
	shell := v.inputs[1].Value()
	if username == "" {
		username = "<username>"
	}
	if shell == "" {
		shell = "/bin/bash"
	}
	preview := fmt.Sprintf("useradd -m -s %s %s", shell, username)

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
	b.WriteString("\n" + _previewStyle.Render("$ "+preview))
	b.WriteString("\n\n" + _dividerStyle.Render("tab next field (on Auth: toggle mode) • enter submit • esc back"))
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
			cmd := fmt.Sprintf("chage --expiredate %s %s", date, v.targetUser)
			if err := usermgmt.SetExpiry(v.targetUser, date); err != nil {
				v.status = _errorStyle.Render("Set expiry failed: " + err.Error())
			} else {
				v.status = _successStyle.Render(fmt.Sprintf("Expiry for '%s' set to %s", v.targetUser, date))
				audit.Log("set-expiry", v.targetUser, cmd)
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
	date := v.inputs[0].Value()
	if date == "" {
		date = "YYYY-MM-DD"
	}
	preview := fmt.Sprintf("chage --expiredate %s %s", date, v.targetUser)
	return fmt.Sprintf("%s\n\n  User: %s\n  Date: %s\n\n%s\n\n%s",
		_headerStyle.Render("Set Expiry Date"),
		_promptStyle.Render(v.targetUser),
		v.inputs[0].View(),
		_previewStyle.Render("$ "+preview),
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
				audit.Log(verb+"-group", v.targetUser, "group="+grp)
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
	grp := v.inputs[0].Value()
	var preview string
	if title == "Add to Group" {
		if grp == "" {
			grp = "<group>"
		}
		preview = fmt.Sprintf("usermod -aG %s %s", grp, v.targetUser)
	} else {
		if grp == "" {
			grp = "<group>"
		}
		preview = fmt.Sprintf("gpasswd -d %s %s", v.targetUser, grp)
	}
	return fmt.Sprintf("%s\n\n  User: %s\n  Group: %s\n\n%s\n\n%s",
		_headerStyle.Render(title),
		_promptStyle.Render(v.targetUser),
		v.inputs[0].View(),
		_previewStyle.Render("$ "+preview),
		_dividerStyle.Render("enter submit • esc back"),
	)
}

// --- Single input (group create/delete) ---

func (v *ManageView) initSingleInput(placeholder string) {
	v.focusIdx = 0
	v.inputs = make([]textinput.Model, 1)
	v.inputs[0] = textinput.New()
	v.inputs[0].Placeholder = placeholder
	v.inputs[0].Focus()
}

func (v ManageView) updateSingleInput(msg tea.Msg, verb string, fn func(string) error) (ManageView, tea.Cmd) {
	if msg, ok := msg.(tea.KeyPressMsg); ok {
		if msg.String() == "enter" {
			val := v.inputs[0].Value()
			if val == "" {
				v.status = _errorStyle.Render("Value is required")
				return v, nil
			}
			if err := fn(val); err != nil {
				v.status = _errorStyle.Render(fmt.Sprintf("%s failed: %s", verb, err))
			} else {
				v.status = _successStyle.Render(fmt.Sprintf("'%s' %sd!", val, verb))
				audit.Log(verb, val, "")
			}
			v.action = _actionMenu
			return v, nil
		}
	}

	var cmd tea.Cmd
	v.inputs[0], cmd = v.inputs[0].Update(msg)
	return v, cmd
}

func (v ManageView) viewSingleInput(title, label string) string {
	val := v.inputs[0].Value()
	var preview string
	if title == "Create Group" {
		if val == "" {
			val = "<name>"
		}
		preview = "groupadd " + val
	} else {
		if val == "" {
			val = "<name>"
		}
		preview = "groupdel " + val
	}
	return fmt.Sprintf("%s\n\n  %s: %s\n\n%s\n\n%s",
		_headerStyle.Render(title),
		label,
		v.inputs[0].View(),
		_previewStyle.Render("$ "+preview),
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
