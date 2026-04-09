// Package common provides shared styles and utilities for the TUI.
package common

import "charm.land/lipgloss/v2"

// Styles used across the TUI components.
var (
	ListStyle = lipgloss.NewStyle().
			Width(35).
			MarginTop(1).
			PaddingRight(3).
			MarginRight(3).
			Border(lipgloss.RoundedBorder())
	ListColorStyle = lipgloss.NewStyle().
			Background(lipgloss.Color("#3d719c"))
	ListItemStyle = lipgloss.NewStyle().
			PaddingLeft(4)
	ListSelectedListItemStyle = lipgloss.NewStyle().
					PaddingLeft(2).
					Foreground(lipgloss.Color("#569cd6"))
	DetailStyle = lipgloss.NewStyle().
			PaddingTop(2)
	DividerStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#5C5C5C")).
			PaddingTop(1).
			PaddingBottom(1)
	TableMainStyle = lipgloss.NewStyle().
			Align(lipgloss.Center)
	TableHeaderStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#569cd6")).
				Bold(true)
	HeaderStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#569cd6")).
			PaddingBottom(1).
			Bold(true).
			Underline(true).
			Inline(true)

	SuccessStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#73c991"))
	ErrorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#f44747"))
	PromptStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#dcdcaa")).
			Bold(true)
)
