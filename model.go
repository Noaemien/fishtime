package main

import (
	"encoding/json"
	"os"
	"time"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
)

// App state
type model struct {
	periods             list.Model
	projects            list.Model
	logs                list.Model
	focused             string // "periods", "projects", or "logs"
	prevFocused         string // Tracks last left pane ("periods" or "projects")
	timerRunning        bool
	timerStart          time.Time
	timerProject        string
	records             []record
	width               int
	height              int
	popupActive         bool
	helpActive          bool
	recordEditActive    bool
	newLogActive        bool
	projectInput        textinput.Model
	recordStartInput    textinput.Model
	recordDurationInput textinput.Model
	newLogProjectInput  textinput.Model
	newLogStartInput    textinput.Model
	newLogDurationInput textinput.Model
	errorMessage        string
}

type record struct {
	Project   string    `json:"project"`
	Duration  int64     `json:"duration"` // Seconds
	StartTime time.Time `json:"start_time"`
}

type appState struct {
	Projects []struct {
		Name     string `json:"name"`
		Selected bool   `json:"selected"`
	} `json:"projects"`
	Records      []record  `json:"records"`
	TimerRunning bool      `json:"timer_running"`
	TimerStart   time.Time `json:"timer_start"`
	TimerProject string    `json:"timer_project"`
}

// Messages
type tickMsg struct{}

// Item for list.Model
type item struct {
	name     string
	selected bool
	isRecord bool
	record   record // Only used for logs pane
}

func (i item) Title() string {
	if i.isRecord {
		return formatItemTitle(i.record)
	}
	return i.name
}

func (i item) Description() string { return "" }
func (i item) FilterValue() string { return i.name }

func newModel() model {
	// Load state from file
	var state appState
	if data, err := os.ReadFile("timer_data.json"); err == nil {
		json.Unmarshal(data, &state)
	} else {
		state = appState{
			Projects: []struct {
				Name     string `json:"name"`
				Selected bool   `json:"selected"`
			}{
				{Name: "Project A", Selected: false},
				{Name: "Project B", Selected: false},
			},
			Records:      []record{},
			TimerRunning: false,
		}
	}

	// Initialize periods list
	periodItems := []list.Item{
		item{name: "All"},
		item{name: "Year"},
		item{name: "Month"},
		item{name: "Week"},
		item{name: "Day"},
	}
	periods := list.New(periodItems, customDelegate{}, 0, 0)
	periods.Title = "Period"
	periods.SetShowStatusBar(false)
	periods.SetShowHelp(false)

	// Initialize projects list
	projectItems := make([]list.Item, len(state.Projects))
	for i, p := range state.Projects {
		projectItems[i] = item{name: p.Name, selected: p.Selected}
	}
	projects := list.New(projectItems, customDelegate{}, 0, 0)
	projects.Title = "Projects (Space to select, d to delete, n to add)"
	projects.SetShowStatusBar(false)
	projects.SetShowHelp(false)

	// Initialize logs list
	logItems := make([]list.Item, len(state.Records))
	for i, r := range state.Records {
		logItems[i] = item{isRecord: true, record: r}
	}
	logs := list.New(logItems, customDelegate{}, 0, 0)
	logs.Title = "Records (e to edit, n to add, d to delete)"
	logs.SetShowStatusBar(false)
	logs.SetShowHelp(false)

	// Initialize text inputs
	projectInput := textinput.New()
	projectInput.Placeholder = "Enter project name"
	projectInput.CharLimit = 30
	projectInput.Width = 20

	recordStartInput := textinput.New()
	recordStartInput.Placeholder = "YYYY-MM-DD HH:MM:SS"
	recordStartInput.CharLimit = 19
	recordStartInput.Width = 20

	recordDurationInput := textinput.New()
	recordDurationInput.Placeholder = "hh:mm:ss"
	recordDurationInput.CharLimit = 8
	recordDurationInput.Width = 20

	newLogProjectInput := textinput.New()
	newLogProjectInput.Placeholder = "Enter project name"
	newLogProjectInput.CharLimit = 30
	newLogProjectInput.Width = 20

	newLogStartInput := textinput.New()
	newLogStartInput.Placeholder = "YYYY-MM-DD HH:MM:SS"
	newLogStartInput.CharLimit = 19
	newLogStartInput.Width = 20

	newLogDurationInput := textinput.New()
	newLogDurationInput.Placeholder = "hh:mm:ss"
	newLogDurationInput.CharLimit = 8
	newLogDurationInput.Width = 20

	// Restore timer state
	timerRunning := state.TimerRunning
	var timerStart time.Time
	timerProject := state.TimerProject
	if timerRunning {
		timerStart = state.TimerStart
		if timerStart.IsZero() {
			timerRunning = false
			timerProject = ""
		}
	}

	return model{
		periods:             periods,
		projects:            projects,
		logs:                logs,
		focused:             "periods",
		prevFocused:         "periods",
		timerRunning:        timerRunning,
		timerStart:          timerStart,
		timerProject:        timerProject,
		records:             state.Records,
		width:               80,
		height:              24,
		popupActive:         false,
		helpActive:          false,
		recordEditActive:    false,
		newLogActive:        false,
		projectInput:        projectInput,
		recordStartInput:    recordStartInput,
		recordDurationInput: recordDurationInput,
		newLogProjectInput:  newLogProjectInput,
		newLogStartInput:    newLogStartInput,
		newLogDurationInput: newLogDurationInput,
		errorMessage:        "",
	}
}

func (m model) saveState() error {
	state := appState{
		Projects: make([]struct {
			Name     string `json:"name"`
			Selected bool   `json:"selected"`
		}, len(m.projects.Items())),
		Records:      m.records,
		TimerRunning: m.timerRunning,
		TimerStart:   m.timerStart,
		TimerProject: m.timerProject,
	}
	for i, it := range m.projects.Items() {
		p, ok := it.(item)
		if ok {
			state.Projects[i] = struct {
				Name     string `json:"name"`
				Selected bool   `json:"selected"`
			}{Name: p.name, Selected: p.selected}
		}
	}
	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile("timer_data.json", data, 0644)
}
