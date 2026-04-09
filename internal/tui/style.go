// Package tui provides the terminal user interface for ugm.
package tui

import "charm.land/lipgloss/v2"

// Styles used across the TUI views.
var (
	_listStyle = lipgloss.NewStyle().
			Width(35).
			MarginTop(1).
			PaddingRight(3).
			MarginRight(3).
			Border(lipgloss.RoundedBorder())
	_listTitleStyle = lipgloss.NewStyle().
			Background(lipgloss.Color("#3d719c"))
	_listItemStyle = lipgloss.NewStyle().
			PaddingLeft(4)
	_listSelectedStyle = lipgloss.NewStyle().
				PaddingLeft(2).
				Foreground(lipgloss.Color("#569cd6"))
	_detailStyle = lipgloss.NewStyle().
			PaddingTop(2)
	_dividerStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#5C5C5C")).
			PaddingTop(1).
			PaddingBottom(1)
	_tableStyle = lipgloss.NewStyle().
			Align(lipgloss.Center)
	_tableHeaderStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#569cd6")).
				Bold(true)
	_headerStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#569cd6")).
			PaddingBottom(1).
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
)
