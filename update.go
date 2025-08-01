package main

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
)

func (m model) Init() tea.Cmd {
	// Force initial tick to filter records immediately
	return tea.Batch(
		tea.Tick(100*time.Millisecond, func(t time.Time) tea.Msg { return tickMsg{} }),
		tea.Tick(0, func(t time.Time) tea.Msg { return tickMsg{} }),
	)
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		// Apply container sizes from config
		sizes := calculateContainerSizes(m.width, m.height)
		m.periods.SetWidth(sizes.Periods.Width)
		m.periods.SetHeight(sizes.Periods.Height)
		m.projects.SetWidth(sizes.Projects.Width)
		m.projects.SetHeight(sizes.Projects.Height)
		m.logs.SetWidth(sizes.Logs.Width)
		m.logs.SetHeight(sizes.Logs.Height)
	case tea.KeyMsg:
		if m.popupActive {
			return m.handleProjectPopup(msg)
		}
		if m.recordEditActive {
			return m.handleRecordEditPopup(msg)
		}
		if m.newLogActive {
			return m.handleNewLogPopup(msg)
		}
		if m.helpActive {
			m.helpActive = false
			return m, nil
		}

		switch msg.String() {
		case "q", "ctrl+c":
			m.saveState()
			return m, tea.Quit
		case "tab":
			if m.focused == "periods" {
				m.focused = "projects"
				m.prevFocused = "periods"
				m.projects.Select(0)
			} else if m.focused == "projects" {
				m.focused = "logs"
				m.logs.Select(0)
			} else if m.focused == "logs" {
				m.focused = "periods"
				m.prevFocused = "periods"
				m.periods.Select(0)
			}
		case "shift+tab":
			if m.focused == "logs" {
				m.focused = "projects"
				m.prevFocused = "projects"
				m.projects.Select(0)
			} else if m.focused == "projects" {
				m.focused = "periods"
				m.prevFocused = "periods"
				m.periods.Select(0)
			} else if m.focused == "periods" {
				m.focused = "logs"
				m.logs.Select(0)
			}
		case "l", "right":
			if m.focused != "logs" {
				m.prevFocused = m.focused
				m.focused = "logs"
				m.logs.Select(0)
			}
		case "h", "left":
			if m.focused == "logs" {
				m.focused = m.prevFocused
				if m.prevFocused == "periods" {
					m.periods.Select(m.periods.Index())
				} else {
					m.projects.Select(m.projects.Index())
				}
			}
		case " ":
			if m.focused == "projects" {
				if i := m.projects.Index(); i >= 0 {
					items := m.projects.Items()
					// Deselect all projects first
					for j, it := range items {
						if p, ok := it.(item); ok {
							p.selected = false
							m.projects.SetItem(j, p)
						}
					}
					// Select the current one
					if p, ok := items[i].(item); ok {
						p.selected = true
						m.projects.SetItem(i, p)
						m.saveState()
						// Trigger immediate logs update
						logItems := make([]list.Item, len(m.filteredRecords()))
						for i, r := range m.filteredRecords() {
							logItems[i] = item{isRecord: true, record: r}
						}
						m.logs.SetItems(logItems)
					}
				}
			}
		case "n":
			if m.focused == "projects" {
				m.popupActive = true
				m.projectInput.Focus()
				m.errorMessage = ""
				return m, textinput.Blink
			} else if m.focused == "logs" {
				// Pre-fill project name with selected project
				for _, it := range m.projects.Items() {
					if p, ok := it.(item); ok && p.selected {
						m.newLogProjectInput.SetValue(p.name)
						break
					}
				}
				m.newLogActive = true
				m.newLogProjectInput.Focus()
				m.errorMessage = ""
				return m, textinput.Blink
			}
		case "d":
			if m.focused == "projects" && len(m.projects.Items()) > 1 {
				if i := m.projects.Index(); i >= 0 {
					m.projects.RemoveItem(i)
					if i >= len(m.projects.Items()) && len(m.projects.Items()) > 0 {
						m.projects.Select(0)
						// Trigger immediate logs update
						logItems := make([]list.Item, len(m.filteredRecords()))
						for i, r := range m.filteredRecords() {
							logItems[i] = item{isRecord: true, record: r}
						}
						m.logs.SetItems(logItems)
					}
					if m.timerRunning {
						if p, ok := m.projects.SelectedItem().(item); !ok || p.name != m.timerProject {
							m.timerRunning = false
							m.timerProject = ""
						}
					}
					m.saveState()
				}
			} else if m.focused == "logs" && m.logs.Index() >= 0 && len(m.records) > 0 {
				i := m.logs.Index()
				if i < len(m.logs.Items()) {
					m.records = append(m.records[:i], m.records[i+1:]...)
					m.logs.RemoveItem(i)
					if len(m.logs.Items()) == 0 {
						m.logs.Select(-1)
					} else if i >= len(m.logs.Items()) {
						m.logs.Select(len(m.logs.Items()) - 1)
					}
					m.saveState()
				}
			}
		case "e":
			if m.focused == "logs" && m.logs.Index() >= 0 && m.logs.Index() < len(m.records) {
				m.recordEditActive = true
				m.recordStartInput.SetValue(m.records[m.logs.Index()].StartTime.Format("2006-01-02 15:04:05"))
				m.recordDurationInput.SetValue(formatDuration(m.records[m.logs.Index()].Duration))
				m.newLogProjectInput.SetValue(m.records[m.logs.Index()].Project)
				m.newLogProjectInput.Focus()
				m.errorMessage = ""
				return m, textinput.Blink
			}
		case "?":
			m.helpActive = true
			return m, nil
		case "s":
			if m.timerRunning {
				duration := int64(time.Since(m.timerStart).Seconds())
				m.records = append(m.records, record{
					Project:   m.timerProject,
					Duration:  duration,
					StartTime: m.timerStart,
				})
				m.logs.InsertItem(len(m.logs.Items()), item{isRecord: true, record: m.records[len(m.records)-1]})
				m.timerRunning = false
				m.timerProject = ""
				m.saveState()
			} else {
				// Start timer for the single selected project
				for _, it := range m.projects.Items() {
					if p, ok := it.(item); ok && p.selected {
						m.timerRunning = true
						m.timerStart = time.Now()
						m.timerProject = p.name
						m.saveState()
						break
					}
				}
			}
		}
	case tickMsg:
		// Only update logs if necessary
		filtered := m.filteredRecords()
		if len(m.logs.Items()) != len(filtered) {
			logItems := make([]list.Item, len(filtered))
			for i, r := range filtered {
				logItems[i] = item{isRecord: true, record: r}
			}
			m.logs.SetItems(logItems)
		}
		return m, tea.Tick(100*time.Millisecond, func(t time.Time) tea.Msg { return tickMsg{} })
	}

	// Update periods and trigger logs update if selection changes
	if m.focused == "periods" {
		prevIndex := m.periods.Index()
		m.periods, cmd = m.periods.Update(msg)
		if m.periods.Index() != prevIndex {
			logItems := make([]list.Item, len(m.filteredRecords()))
			for i, r := range m.filteredRecords() {
				logItems[i] = item{isRecord: true, record: r}
			}
			m.logs.SetItems(logItems)
		}
	} else if m.focused == "projects" {
		m.projects, cmd = m.projects.Update(msg)
	} else {
		m.logs, cmd = m.logs.Update(msg)
	}
	return m, cmd
}

func (m model) handleProjectPopup(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg.String() {
	case "enter":
		name := m.projectInput.Value()
		if name == "" {
			m.errorMessage = "Project name cannot be empty"
			return m, nil
		}
		for _, it := range m.projects.Items() {
			if p, ok := it.(item); ok && p.name == name {
				m.errorMessage = "Project name already exists"
				return m, nil
			}
		}
		newProject := item{name: name}
		m.projects.InsertItem(len(m.projects.Items()), newProject)
		m.saveState()
		m.popupActive = false
		m.projectInput.Reset()
		m.errorMessage = ""
	case "esc":
		m.popupActive = false
		m.projectInput.Reset()
		m.errorMessage = ""
	default:
		m.projectInput, cmd = m.projectInput.Update(msg)
	}
	return m, cmd
}

func (m model) handleRecordEditPopup(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg.String() {
	case "enter":
		if i := m.logs.Index(); i >= 0 && i < len(m.records) {
			project := m.newLogProjectInput.Value()
			startTime, err1 := time.Parse("2006-01-02 15:04:05", m.recordStartInput.Value())
			duration, err2 := parseDuration(m.recordDurationInput.Value())
			if project == "" {
				m.errorMessage = "Project name cannot be empty"
				return m, nil
			}
			if err1 != nil {
				m.errorMessage = "Invalid start time format (use YYYY-MM-DD HH:MM:SS)"
				return m, nil
			}
			if err2 != nil {
				m.errorMessage = "Invalid duration format (use hh:mm:ss, non-negative, minutes/seconds <= 59)"
				return m, nil
			}
			for _, it := range m.projects.Items() {
				if p, ok := it.(item); ok && p.name == project {
					m.records[i].Project = project
					m.records[i].StartTime = startTime
					m.records[i].Duration = duration
					m.logs.SetItem(i, item{isRecord: true, record: m.records[i]})
					m.saveState()
					m.recordEditActive = false
					m.newLogProjectInput.Reset()
					m.recordStartInput.Reset()
					m.recordDurationInput.Reset()
					m.errorMessage = ""
					return m, nil
				}
			}
			m.errorMessage = "Project does not exist"
			return m, nil
		}
	case "esc":
		m.recordEditActive = false
		m.newLogProjectInput.Reset()
		m.recordStartInput.Reset()
		m.recordDurationInput.Reset()
		m.errorMessage = ""
	case "tab":
		if m.newLogProjectInput.Focused() {
			m.newLogProjectInput.Blur()
			m.recordStartInput.Focus()
		} else if m.recordStartInput.Focused() {
			m.recordStartInput.Blur()
			m.recordDurationInput.Focus()
		} else {
			m.recordDurationInput.Blur()
			m.newLogProjectInput.Focus()
		}
		return m, textinput.Blink
	case "shift+tab":
		if m.recordDurationInput.Focused() {
			m.recordDurationInput.Blur()
			m.recordStartInput.Focus()
		} else if m.recordStartInput.Focused() {
			m.recordStartInput.Blur()
			m.newLogProjectInput.Focus()
		} else {
			m.newLogProjectInput.Blur()
			m.recordDurationInput.Focus()
		}
		return m, textinput.Blink
	default:
		if m.newLogProjectInput.Focused() {
			m.newLogProjectInput, cmd = m.newLogProjectInput.Update(msg)
		} else if m.recordStartInput.Focused() {
			m.recordStartInput, cmd = m.recordStartInput.Update(msg)
		} else {
			m.recordDurationInput, cmd = m.recordDurationInput.Update(msg)
		}
	}
	return m, cmd
}

func (m model) handleNewLogPopup(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg.String() {
	case "enter":
		project := m.newLogProjectInput.Value()
		startTime, err1 := time.Parse("2006-01-02 15:04:05", m.newLogStartInput.Value())
		duration, err2 := parseDuration(m.newLogDurationInput.Value())
		if project == "" {
			m.errorMessage = "Project name cannot be empty"
			return m, nil
		}
		if err1 != nil {
			m.errorMessage = "Invalid start time format (use YYYY-MM-DD HH:MM:SS)"
			return m, nil
		}
		if err2 != nil {
			m.errorMessage = "Invalid duration format (use hh:mm:ss, non-negative, minutes/seconds <= 59)"
			return m, nil
		}
		// Validate project exists
		for _, it := range m.projects.Items() {
			if p, ok := it.(item); ok && p.name == project {
				newRecord := record{
					Project:   project,
					Duration:  duration,
					StartTime: startTime,
				}
				m.records = append(m.records, newRecord)
				m.logs.InsertItem(len(m.logs.Items()), item{isRecord: true, record: newRecord})
				m.saveState()
				m.newLogActive = false
				m.newLogProjectInput.Reset()
				m.newLogStartInput.Reset()
				m.newLogDurationInput.Reset()
				m.errorMessage = ""
				return m, nil
			}
		}
		m.errorMessage = "Project does not exist"
		return m, nil
	case "esc":
		m.newLogActive = false
		m.newLogProjectInput.Reset()
		m.newLogStartInput.Reset()
		m.newLogDurationInput.Reset()
		m.errorMessage = ""
	case "tab":
		if m.newLogProjectInput.Focused() {
			m.newLogProjectInput.Blur()
			m.newLogStartInput.Focus()
		} else if m.newLogStartInput.Focused() {
			m.newLogStartInput.Blur()
			m.newLogDurationInput.Focus()
		} else {
			m.newLogDurationInput.Blur()
			m.newLogProjectInput.Focus()
		}
		return m, textinput.Blink
	case "shift+tab":
		if m.newLogDurationInput.Focused() {
			m.newLogDurationInput.Blur()
			m.newLogStartInput.Focus()
		} else if m.newLogStartInput.Focused() {
			m.newLogStartInput.Blur()
			m.newLogProjectInput.Focus()
		} else {
			m.newLogProjectInput.Blur()
			m.newLogDurationInput.Focus()
		}
		return m, textinput.Blink
	default:
		if m.newLogProjectInput.Focused() {
			m.newLogProjectInput, cmd = m.newLogProjectInput.Update(msg)
		} else if m.newLogStartInput.Focused() {
			m.newLogStartInput, cmd = m.newLogStartInput.Update(msg)
		} else {
			m.newLogDurationInput, cmd = m.newLogDurationInput.Update(msg)
		}
	}
	return m, cmd
}
