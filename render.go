package main

import (
	"fmt"
	"io"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/bubbles/list"
  "github.com/charmbracelet/lipgloss"
)

// Custom delegate to control list rendering
type customDelegate struct{}

func (d customDelegate) Height() int                             { return 1 } // Each item takes 1 line
func (d customDelegate) Spacing() int                            { return 0 } // No extra spacing between items
func (d customDelegate) Update(msg tea.Msg, m *list.Model) tea.Cmd { return nil }

func (d customDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	i, ok := listItem.(item)
	if !ok {
		return
	}

	var str string
	if index == m.Index() {
		// Focused item
		if i.selected && !i.isRecord {
			// Focused and selected (Projects pane only)
			str = focusedTextStyle.Render(selectedStyle.Render(i.Title()))
		} else {
			// Focused but not selected
			str = focusedTextStyle.Render(i.Title())
		}
	} else {
		// Not focused
		if i.selected && !i.isRecord {
			// Not focused but selected (Projects pane only)
			str = selectedStyle.Render(i.Title())
		} else {
			// Not focused, not selected
			str = normalTextStyle.Render(i.Title())
		}
	}

	fmt.Fprint(w, str)
}

func (m model) View() string {
	// Skip rendering logs until initial filtering is done
	if len(m.logs.Items()) > len(m.filteredRecords()) && m.focused == "logs" {
		return ""
	}

	// Get container sizes from config
	sizes := calculateContainerSizes(m.width, m.height)

	// Left panel: Periods and Projects
	periodsStyle := inactiveStyle
	projectsStyle := inactiveStyle
	logsStyle := inactiveStyle
	if m.focused == "periods" {
		periodsStyle = focusedStyle
	} else if m.focused == "projects" {
		projectsStyle = focusedStyle
	} else {
		logsStyle = focusedStyle
	}

	periodsRendered := periodsStyle.Width(sizes.Periods.Width).Height(sizes.Periods.Height).Render(m.periods.View())
	projectsRendered := projectsStyle.Width(sizes.Projects.Width).Height(sizes.Projects.Height).Render(m.projects.View())

	// Right panel: Logs with total
	total := m.totalDuration()
	totalStr := totalFooterStyle.Render(fmt.Sprintf("Total: %s", formatDuration(int64(total.Seconds()))))
	logsContent := lipgloss.JoinVertical(lipgloss.Left, m.logs.View(), totalStr)
	logsRendered := logsStyle.Width(sizes.Logs.Width).Height(sizes.Logs.Height).Render(logsContent)

	// Status bar
	status := "Timer: Off"
	if m.timerRunning {
		elapsed := time.Since(m.timerStart)
		status = fmt.Sprintf("Timer: %s", formatDuration(int64(elapsed.Seconds())))
	}
	timerStyle := timerOffStyle
	if m.timerRunning {
		timerStyle = timerOnStyle
	}
	statusBar := timerStyle.Width(sizes.StatusBar.Width).Height(sizes.StatusBar.Height).Render(status)

	// Layout
	left := lipgloss.JoinVertical(lipgloss.Left, periodsRendered, projectsRendered)
	main := lipgloss.JoinHorizontal(lipgloss.Center, left, logsRendered)
	content := lipgloss.JoinVertical(lipgloss.Left, main, statusBar)

	// Popups
	if m.popupActive {
		popupContent := "New Project\n" + m.projectInput.View() + "\nEnter to confirm, Esc to cancel"
		if m.errorMessage != "" {
			popupContent += "\n" + errorStyle.Render("Error: "+m.errorMessage)
		}
		popup := popupStyle.Render(popupContent)
		popup = lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, popup)
		return lipgloss.JoinVertical(lipgloss.Left, content, popup)
	}
	if m.recordEditActive {
		popupContent := "Edit Record\n" +
			"Project: " + m.newLogProjectInput.View() + "\n" +
			"Start Time: " + m.recordStartInput.View() + "\n" +
			"Duration: " + m.recordDurationInput.View() + "\n" +
			"Enter to confirm, Esc to cancel, Tab to switch fields"
		if m.errorMessage != "" {
			popupContent += "\n" + errorStyle.Render("Error: "+m.errorMessage)
		}
		popup := popupStyle.Render(popupContent)
		popup = lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, popup)
		return lipgloss.JoinVertical(lipgloss.Left, content, popup)
	}
	if m.newLogActive {
		popupContent := "New Record\n" +
			"Project: " + m.newLogProjectInput.View() + "\n" +
			"Start Time: " + m.newLogStartInput.View() + "\n" +
			"Duration: " + m.newLogDurationInput.View() + "\n" +
			"Enter to confirm, Esc to cancel, Tab to switch fields"
		if m.errorMessage != "" {
			popupContent += "\n" + errorStyle.Render("Error: "+m.errorMessage)
		}
		popup := popupStyle.Render(popupContent)
		popup = lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, popup)
		return lipgloss.JoinVertical(lipgloss.Left, content, popup)
	}
	if m.helpActive {
		helpText := `Keyboard Shortcuts
?        - Show this help
q, ctrl+c - Quit
tab       - Move focus down
shift+tab - Move focus up
j, k      - Navigate items
l, right  - Focus logs pane
h, left   - Return to previous pane
space     - Select project
n         - Add new project (in Projects) or record (in Logs)
d         - Delete project (in Projects) or record (in Logs)
e         - Edit record (in Logs)
s         - Start/stop timer
Press any key to close`
		popup := popupStyle.Width(50).Render(helpText)
		popup = lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, popup)
		return lipgloss.JoinVertical(lipgloss.Left, content, popup)
	}

	return content
}
