package main

import "github.com/charmbracelet/lipgloss"

// Styles for lazygit-like aesthetics
var (
	focusedStyle = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("69")). // Blue border
		Padding(1, 2).
		Margin(0, 1)
	inactiveStyle = lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("240")). // Gray border
		Padding(1, 2).
		Margin(0, 1)
	selectedStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("42")) // Green for selected projects
	focusedTextStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("243")) // Darker gray for focused
	normalTextStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("252")) // Default color
	timerOnStyle  = lipgloss.NewStyle().
		Foreground(lipgloss.Color("42")).
		Border(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("42")).
		Padding(0, 1).
		Margin(0, 1).
		Height(1) // Compact status bar
	timerOffStyle = lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("240")).
		Padding(0, 1).
		Margin(0, 1).
		Height(1) // Compact status bar
	popupStyle = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("69")).
		Background(lipgloss.Color("235")).
		Padding(1, 2).
		Margin(1, 2).
		Width(50)
	errorStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("196")). // Red for error messages
		Padding(0, 1)
	totalFooterStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("252")).
		Padding(1, 0).
		Width(50)
)
