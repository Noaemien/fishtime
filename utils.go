package main

import (
	"fmt"
	"time"
)

// Helper to format duration as hh:mm:ss
func formatDuration(seconds int64) string {
	hrs := seconds / 3600
	seconds %= 3600
	mins := seconds / 60
	secs := seconds % 60
	return fmt.Sprintf("%02d:%02d:%02d", hrs, mins, secs)
}

// Helper to parse hh:mm:ss duration
func parseDuration(input string) (int64, error) {
	var hrs, mins, secs int
	_, err := fmt.Sscanf(input, "%d:%d:%d", &hrs, &mins, &secs)
	if err != nil {
		return 0, fmt.Errorf("invalid duration format: use hh:mm:ss")
	}
	if hrs < 0 || mins < 0 || secs < 0 || mins > 59 || secs > 59 {
		return 0, fmt.Errorf("invalid duration: hours, minutes, and seconds must be non-negative, minutes and seconds <= 59")
	}
	return int64(hrs*3600 + mins*60 + secs), nil
}

// Helper to format record item title
func formatItemTitle(r record) string {
	return fmt.Sprintf("%s - %s @ %s", r.Project, formatDuration(r.Duration), r.StartTime.Format("2006-01-02 15:04:05"))
}

func (m model) filteredRecords() []record {
	now := time.Now()
	var period string
	if sel := m.periods.SelectedItem(); sel != nil {
		if p, ok := sel.(item); ok {
			period = p.name
		} else {
			period = "All"
		}
	} else {
		period = "All"
	}

	// Cache selected project
	var selectedProject string
	projectMap := make(map[string]bool)
	for _, it := range m.projects.Items() {
		if p, ok := it.(item); ok {
			projectMap[p.name] = true
			if p.selected {
				selectedProject = p.name
			}
		}
	}

	var filtered []record
	for _, r := range m.records {
		// Filter by selected project
		if selectedProject != "" && r.Project != selectedProject {
			continue
		}
		// Verify project exists
		if !projectMap[r.Project] {
			continue
		}
		// Filter by period
		if period != "All" {
			var maxAge time.Duration
			switch period {
			case "Year":
				maxAge = 365 * 24 * time.Hour
			case "Month":
				maxAge = 30 * 24 * time.Hour
			case "Week":
				maxAge = 7 * 24 * time.Hour
			case "Day":
				maxAge = 24 * time.Hour
			}
			if now.Sub(r.StartTime) > maxAge {
				continue
			}
		}
		filtered = append(filtered, r)
	}
	return filtered
}

func (m model) totalDuration() time.Duration {
	var total int64
	for _, r := range m.filteredRecords() {
		total += r.Duration
	}
	return time.Duration(total) * time.Second
}
