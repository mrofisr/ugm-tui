// Package tui provides the terminal user interface for ugm.
package tui

import (
	"os"
	"strings"

	"charm.land/lipgloss/v2"
)

// Badge indicators for account status.
const (
	_badgeActive = "●"
	_badgeLocked = "●"
	_badgeExpiry = "●"
	_badgeSudo   = "●"
	_badgeSys    = "●"
)

// listWidth returns a responsive list panel width (roughly 1/3 of terminal, min 36, max 52).
func listWidth(termWidth int) int {
	w := termWidth / 3
	if w < 36 {
		w = 36
	}
	if w > 52 {
		w = 52
	}
	return w
}

// Styles used across the TUI views.
var (
	_listStyle = lipgloss.NewStyle().
			Width(44).
			PaddingRight(2).
			MarginRight(1).
			Border(lipgloss.RoundedBorder())
	_listTitleStyle = lipgloss.NewStyle().
			Background(lipgloss.Color("#3d719c"))
	_listItemStyle = lipgloss.NewStyle().
			PaddingLeft(4)
	_listSelectedStyle = lipgloss.NewStyle().
				PaddingLeft(2).
				Foreground(lipgloss.Color("#569cd6"))
	_detailStyle = lipgloss.NewStyle().
			PaddingTop(1).
			PaddingLeft(1)
	_dividerStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#5C5C5C"))
	_tableStyle = lipgloss.NewStyle().
			Align(lipgloss.Center)
	_tableHeaderStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#569cd6")).
				Bold(true)
	_headerStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#569cd6")).
			Bold(true).
			Underline(true).
			Inline(true)
	_successStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#73c991"))
	_errorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#f44747"))
	_promptStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#dcdcaa")).
			Bold(true)
	_previewStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#808080")).
			Italic(true)
	_statusBarStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#abb2bf")).
			Background(lipgloss.Color("#282c34")).
			Padding(0, 1)
	_statusKeyStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#282c34")).
			Background(lipgloss.Color("#569cd6")).
			Padding(0, 1).
			Bold(true)

	// Tab bar styles.
	_tabActiveStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#ffffff")).
			Background(lipgloss.Color("#3d719c")).
			Padding(0, 1).
			Bold(true)
	_tabInactiveStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#808080")).
				Padding(0, 1)

	// Badge color styles.
	_badgeActiveStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#73c991"))
	_badgeLockedStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#f44747"))
	_badgeExpiryStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#dcdcaa"))
	_badgeSudoStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("#c678dd"))
	_badgeSysStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("#5C5C5C"))

	// Help overlay.
	_helpOverlayStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color("#569cd6")).
				Padding(1, 2)

	// Header bar (k9s-style context bar at top).
	_headerBarStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#282c34")).
			Background(lipgloss.Color("#569cd6")).
			Bold(true).
			Padding(0, 1)
	_headerBarDimStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#abb2bf")).
				Background(lipgloss.Color("#3d719c")).
				Padding(0, 1)

	// Breadcrumb style for manage view.
	_breadcrumbStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#808080"))
	_breadcrumbActiveStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#569cd6")).
				Bold(true)

	// Menu category separator.
	_menuCategoryStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#569cd6")).
				Bold(true).
				MarginTop(1)

	// Flash/toast message styles.
	_flashSuccessStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#282c34")).
				Background(lipgloss.Color("#73c991")).
				Padding(0, 1).
				Bold(true)
	_flashErrorStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#ffffff")).
				Background(lipgloss.Color("#f44747")).
				Padding(0, 1).
				Bold(true)
)

// renderTabBar renders a horizontal tab bar with the active tab highlighted.
func renderTabBar(tabs []string, active int) string {
	var parts []string
	for i, t := range tabs {
		if i == active {
			parts = append(parts, _tabActiveStyle.Render(t))
		} else {
			parts = append(parts, _tabInactiveStyle.Render(t))
		}
	}
	return lipgloss.JoinHorizontal(lipgloss.Top, parts...) + "\n"
}

// renderHeaderBar renders a k9s-style context bar showing hostname and operator.
func renderHeaderBar(width int) string {
	hostname, _ := os.Hostname()
	operator := os.Getenv("SUDO_USER")
	if operator == "" {
		operator = os.Getenv("USER")
	}

	left := _headerBarStyle.Render("⚙ ugm")
	info := _headerBarDimStyle.Render("host:" + hostname + "  operator:" + operator)

	gap := width - lipgloss.Width(left) - lipgloss.Width(info)
	if gap < 0 {
		gap = 0
	}
	fill := _headerBarDimStyle.Render(strings.Repeat(" ", gap))
	return left + fill + info
}

// renderBreadcrumb renders a breadcrumb trail like: users › rafi › manage › Lock User
func renderBreadcrumb(parts ...string) string {
	sep := _breadcrumbStyle.Render(" › ")
	var rendered []string
	for i, p := range parts {
		if i == len(parts)-1 {
			rendered = append(rendered, _breadcrumbActiveStyle.Render(p))
		} else {
			rendered = append(rendered, _breadcrumbStyle.Render(p))
		}
	}
	return strings.Join(rendered, sep) + "\n"
}
