package main

// containerConfig defines the sizing parameters for dynamic containers
type containerConfig struct {
	Periods struct {
		WidthRatio  float64 // Proportion of window width
		HeightRatio float64 // Proportion of window height
		Margin      int     // Margin around the container
	}
	Projects struct {
		WidthRatio  float64
		HeightRatio float64
		Margin      int
	}
	Logs struct {
		WidthRatio float64
		Height     int // Fixed height minus status bar and margins
		Margin     int
	}
	StatusBar struct {
		WidthRatio float64
		Height     int // Fixed height
		Margin     int
	}
}

// containerSizes holds calculated dimensions for each container
type containerSizes struct {
	Periods   containerSize
	Projects  containerSize
	Logs      containerSize
	StatusBar containerSize
}

type containerSize struct {
	Width  int
	Height int
}

// defaultContainerConfig defines the default sizing parameters
var defaultContainerConfig = containerConfig{
	Periods: struct {
		WidthRatio  float64
		HeightRatio float64
		Margin      int
	}{
		WidthRatio:  0.5,  // Half of window width
		HeightRatio: 0.5,  // Half of left panel height
		Margin:      4,    // Space around container
	},
	Projects: struct {
		WidthRatio  float64
		HeightRatio float64
		Margin      int
	}{
		WidthRatio:  0.5,  // Half of window width
		HeightRatio: 0.5,  // Half of left panel height
		Margin:      4,    // Space around container
	},
	Logs: struct {
		WidthRatio float64
		Height     int
		Margin     int
	}{
		WidthRatio: 0.5, // Half of window width
		Height:     0,   // Calculated dynamically
		Margin:     4,   // Space around container
	},
	StatusBar: struct {
		WidthRatio float64
		Height     int
		Margin     int
	}{
		WidthRatio: 1.0, // Full window width
		Height:     1,   // Fixed height
		Margin:     4,   // Space around container
	},
}

// calculateContainerSizes computes dimensions based on window size and config
func calculateContainerSizes(windowWidth, windowHeight int) containerSizes {
	config := defaultContainerConfig
	sizes := containerSizes{}

	// Calculate Periods dimensions
	periodsWidth := int(float64(windowWidth) * config.Periods.WidthRatio) - config.Periods.Margin
	periodsHeight := int(float64(windowHeight) * config.Periods.HeightRatio) - config.Periods.Margin - 2 // Adjust for title/border
	sizes.Periods = containerSize{
		Width:  periodsWidth,
		Height: periodsHeight,
	}

	// Calculate Projects dimensions
	projectsWidth := int(float64(windowWidth) * config.Projects.WidthRatio) - config.Projects.Margin
	projectsHeight := int(float64(windowHeight) * config.Projects.HeightRatio) - config.Projects.Margin - 2 // Adjust for title/border
	sizes.Projects = containerSize{
		Width:  projectsWidth,
		Height: projectsHeight,
	}

	// Calculate Logs dimensions
	logsWidth := int(float64(windowWidth) * config.Logs.WidthRatio) - config.Logs.Margin
	logsHeight := windowHeight - config.StatusBar.Height - config.StatusBar.Margin - config.Logs.Margin - 2 // Full height minus status bar
	sizes.Logs = containerSize{
		Width:  logsWidth,
		Height: logsHeight,
	}

	// Calculate Status Bar dimensions
	statusBarWidth := int(float64(windowWidth) * config.StatusBar.WidthRatio) - config.StatusBar.Margin
	statusBarHeight := config.StatusBar.Height
	sizes.StatusBar = containerSize{
		Width:  statusBarWidth,
		Height: statusBarHeight,
	}

	return sizes
}
