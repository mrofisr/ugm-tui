package tui

import (
	"fmt"
	"strings"

	"charm.land/bubbles/v2/spinner"
	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/mrofisr/ugm-tui/internal/audit"
	"github.com/mrofisr/ugm-tui/internal/usermgmt"
)

// cmdResultMsg carries the result of an async command execution.
type cmdResultMsg struct {
	status  string
	isError bool
}

// passwdAgingMsg carries password aging info text.
type passwdAgingMsg string

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
	action        manageAction
	menuIndex     int
	inputs        []textinput.Model
	focusIdx      int
	authIsSSH     bool
	targetUser    string
	status        string
	statusIsError bool
	infoText      string
	done          bool
	spinner       spinner.Model
	spinning      bool
}

// menuEntry represents a single menu item with icon and category.
type menuEntry struct {
	icon     string
	label    string
	category string // empty = same category as previous
}

var _menuEntries = []menuEntry{
	{icon: "👤", label: "Create New User", category: "Account"},
	{icon: "🗑", label: "Delete User"},
	{icon: "🔒", label: "Lock User (revoke access)", category: "Access"},
	{icon: "🔓", label: "Unlock User"},
	{icon: "📅", label: "Set Expiry Date"},
	{icon: "📋", label: "View Password Aging Info"},
	{icon: "➕", label: "Add to Group (assign role)", category: "Groups"},
	{icon: "➖", label: "Remove from Group"},
	{icon: "📁", label: "Create Group (new role)"},
	{icon: "🗑", label: "Delete Group"},
}

func newManageView() ManageView {
	s := spinner.New(spinner.WithSpinner(spinner.Dot))
	return ManageView{action: _actionMenu, spinner: s}
}

func (v *ManageView) setTarget(username string) {
	v.targetUser = username
	v.action = _actionMenu
	v.menuIndex = 0
	v.status = ""
	v.statusIsError = false
	v.infoText = ""
	v.done = false
	v.spinning = false
}

// runCmd starts the spinner and runs fn asynchronously, returning the result as cmdResultMsg.
func (v ManageView) runCmd(label, auditAction, auditTarget, auditDetail string, fn func() error) (ManageView, tea.Cmd) {
	v.spinning = true
	v.status = ""
	work := func() tea.Msg {
		if err := fn(); err != nil {
			return cmdResultMsg{
				status:  _errorStyle.Render(fmt.Sprintf("%s failed: %s", label, err)),
				isError: true,
			}
		}
		if auditAction != "" {
			audit.Log(auditAction, auditTarget, auditDetail)
		}
		return cmdResultMsg{
			status:  _successStyle.Render(fmt.Sprintf("%s succeeded!", label)),
			isError: false,
		}
	}
	return v, tea.Batch(v.spinner.Tick, work)
}

func (v ManageView) update(msg tea.Msg) (ManageView, tea.Cmd) {
	// Handle spinner ticks while spinning.
	if v.spinning {
		switch msg := msg.(type) {
		case cmdResultMsg:
			v.spinning = false
			v.status = msg.status
			v.statusIsError = msg.isError
			v.action = _actionMenu
			return v, nil
		case passwdAgingMsg:
			v.spinning = false
			v.action = _actionPasswdAging
			v.infoText = string(msg)
			return v, nil
		case tea.KeyPressMsg:
			// Block input while spinning.
			return v, nil
		default:
			var cmd tea.Cmd
			v.spinner, cmd = v.spinner.Update(msg)
			return v, cmd
		}
	}

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
	if v.spinning {
		s := v.spinner.View() + " Processing…"
		return lipgloss.NewStyle().Padding(1, 2).Render(s)
	}

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
			if v.menuIndex < len(_menuEntries)-1 {
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
				v.spinning = true
				v.status = ""
				target := v.targetUser
				return v, tea.Batch(v.spinner.Tick, func() tea.Msg {
					info, err := usermgmt.PasswordAging(target)
					if err != nil {
						return cmdResultMsg{status: _errorStyle.Render(err.Error()), isError: true}
					}
					return passwdAgingMsg(info)
				})
			case 6:
				v.action = _actionAddGroup
				v.initGroupInput()
			case 7:
				v.action = _actionRemoveGroup
				v.initGroupInput()
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
	b.WriteString(title + "\n")

	lastCategory := ""
	for i, entry := range _menuEntries {
		// Render category header
		if entry.category != "" && entry.category != lastCategory {
			b.WriteString("\n" + _menuCategoryStyle.Render("  "+entry.category) + "\n")
			lastCategory = entry.category
		}

		cursor := "  "
		style := _listItemStyle
		if i == v.menuIndex {
			cursor = "▶ "
			style = _listSelectedStyle
		}
		b.WriteString(style.Render(cursor+entry.icon+" "+entry.label) + "\n")
	}
	b.WriteString("\n" + _dividerStyle.Render("↑/↓ navigate • enter select • esc back"))
	return b.String()
}

// --- Generic confirm (y/n) with command preview ---

func (v ManageView) updateConfirm(msg tea.Msg, verb, cmdPreview string, fn func() error) (ManageView, tea.Cmd) {
	if msg, ok := msg.(tea.KeyPressMsg); ok {
		switch msg.String() {
		case "y":
			label := verb + " user '" + v.targetUser + "'"
			return v.runCmd(label, verb, v.targetUser, cmdPreview, fn)
		case "n":
			v.action = _actionMenu
		}
	}
	return v, nil
}

func (v ManageView) viewConfirmPrompt(title, prompt, cmdPreview string) string {
	bc := renderBreadcrumb("users", v.targetUser, "manage", title)
	return fmt.Sprintf("%s\n%s\n%s\n\n%s\n\n%s",
		bc,
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
		v.statusIsError = true
		return v, nil
	}

	isSSH := v.authIsSSH
	return v.runCmd("Create user '"+username+"'", "", "", "", func() error {
		if err := usermgmt.CreateUser(username, shell); err != nil {
			return err
		}
		audit.Log("create-user", username, fmt.Sprintf("useradd -m -s %s %s", shell, username))

		if secret != "" {
			if isSSH {
				if err := usermgmt.SetSSHKey(username, secret); err != nil {
					return err
				}
				audit.Log("set-ssh-key", username, "wrote authorized_keys")
			} else {
				if err := usermgmt.SetPassword(username, secret); err != nil {
					return err
				}
				audit.Log("set-password", username, "chpasswd")
			}
		}
		return nil
	})
}

func (v ManageView) viewCreate() string {
	title := _headerStyle.Render("Create New User")
	bc := renderBreadcrumb("users", v.targetUser, "manage", "Create New User")
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
	b.WriteString(bc + title + "\n\n")
	for i, input := range v.inputs {
		cursor := "  "
		if i == v.focusIdx {
			cursor = "▶ "
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
				v.statusIsError = true
				return v, nil
			}
			cmdStr := fmt.Sprintf("chage --expiredate %s %s", date, v.targetUser)
			target := v.targetUser
			return v.runCmd("Set expiry for '"+target+"' to "+date, "set-expiry", target, cmdStr, func() error {
				return usermgmt.SetExpiry(target, date)
			})
		}
	}

	var cmd tea.Cmd
	v.inputs[0], cmd = v.inputs[0].Update(msg)
	return v, cmd
}

func (v ManageView) viewExpiry() string {
	bc := renderBreadcrumb("users", v.targetUser, "manage", "Set Expiry Date")
	date := v.inputs[0].Value()
	if date == "" {
		date = "YYYY-MM-DD"
	}
	preview := fmt.Sprintf("chage --expiredate %s %s", date, v.targetUser)
	return fmt.Sprintf("%s%s\n\n  User: %s\n  Date: %s\n\n%s\n\n%s",
		bc,
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
				v.statusIsError = true
				return v, nil
			}
			target := v.targetUser
			return v.runCmd(verb+" group '"+grp+"' for '"+target+"'", verb+"-group", target, "group="+grp, func() error {
				return fn(target, grp)
			})
		}
	}

	var cmd tea.Cmd
	v.inputs[0], cmd = v.inputs[0].Update(msg)
	return v, cmd
}

func (v ManageView) viewGroupInput(title string) string {
	bc := renderBreadcrumb("users", v.targetUser, "manage", title)
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
	return fmt.Sprintf("%s%s\n\n  User: %s\n  Group: %s\n\n%s\n\n%s",
		bc,
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
				v.statusIsError = true
				return v, nil
			}
			return v.runCmd(verb+" '"+val+"'", verb, val, "", func() error {
				return fn(val)
			})
		}
	}

	var cmd tea.Cmd
	v.inputs[0], cmd = v.inputs[0].Update(msg)
	return v, cmd
}

func (v ManageView) viewSingleInput(title, label string) string {
	bc := renderBreadcrumb("users", v.targetUser, "manage", title)
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
	return fmt.Sprintf("%s%s\n\n  %s: %s\n\n%s\n\n%s",
		bc,
		_headerStyle.Render(title),
		label,
		v.inputs[0].View(),
		_previewStyle.Render("$ "+preview),
		_dividerStyle.Render("enter submit • esc back"),
	)
}

// --- Info view (read-only) ---

func (v ManageView) viewInfo(title string) string {
	bc := renderBreadcrumb("users", v.targetUser, "manage", title)
	return fmt.Sprintf("%s%s\n\n  User: %s\n\n%s\n\n%s",
		bc,
		_headerStyle.Render(title),
		_promptStyle.Render(v.targetUser),
		v.infoText,
		_dividerStyle.Render("any key to go back"),
	)
}
