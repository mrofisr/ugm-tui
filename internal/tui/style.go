// Package tui provides the terminal user interface for ugm.
package tui

import "charm.land/lipgloss/v2"

// Badge indicators for account status.
const (
	_badgeActive = "●"
	_badgeLocked = "●"
	_badgeExpiry = "●"
	_badgeSudo   = "●"
	_badgeSys    = "●"
)

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
