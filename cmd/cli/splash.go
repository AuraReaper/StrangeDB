package main

import (
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Purple palette for wave animation
var palette = []lipgloss.Color{
	lipgloss.Color("#7C3AED"), // deep purple
	lipgloss.Color("#9333EA"), // bright purple
	lipgloss.Color("#A855F7"), // glow
	lipgloss.Color("#9333EA"),
}

var (
	grayStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("#6B7280"))
	taglineStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#A855F7"))
)

var bigText = []string{
	"███████╗████████╗██████╗  █████╗ ███╗   ██╗ ██████╗ ███████╗",
	"██╔════╝╚══██╔══╝██╔══██╗██╔══██╗████╗  ██║██╔════╝ ██╔════╝",
	"███████╗   ██║   ██████╔╝███████║██╔██╗ ██║██║  ███╗█████╗  ",
	"╚════██║   ██║   ██╔══██╗██╔══██║██║╚██╗██║██║   ██║██╔══╝  ",
	"███████║   ██║   ██║  ██║██║  ██║██║ ╚████║╚██████╔╝███████╗",
	"╚══════╝   ╚═╝   ╚═╝  ╚═╝╚═╝  ╚═╝╚═╝  ╚═══╝ ╚═════╝ ╚══════╝",
}

// SplashModel is the animated logo screen
type SplashModel struct {
	step     int
	pulse    bool
	done     bool
	mode     string // "cluster" or "single"
	maxSteps int
}

type splashTickMsg struct{}

func NewSplashModel(mode string) SplashModel {
	return SplashModel{
		mode:     mode,
		maxSteps: 40, // About 5 seconds of animation
	}
}

func (m SplashModel) Init() tea.Cmd {
	return tea.Tick(120*time.Millisecond, func(time.Time) tea.Msg {
		return splashTickMsg{}
	})
}

func (m SplashModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case splashTickMsg:
		m.step++
		if m.step%10 == 0 {
			m.pulse = !m.pulse
		}
		// Auto-advance after animation
		if m.step >= m.maxSteps {
			m.done = true
			return m, tea.Quit
		}
		return m, tea.Tick(120*time.Millisecond, func(time.Time) tea.Msg {
			return splashTickMsg{}
		})

	case tea.KeyMsg:
		switch msg.String() {
		case "enter", " ", "q", "ctrl+c":
			m.done = true
			return m, tea.Quit
		}
	}

	return m, nil
}

func (m SplashModel) View() string {
	var b strings.Builder
	b.WriteString("\n\n")

	// Animated logo with wave effect
	for _, line := range bigText {
		for i, ch := range line {
			// Smooth wave animation
			color := palette[(i/3+m.step)%len(palette)]
			style := lipgloss.NewStyle().Foreground(color)
			if m.pulse {
				style = style.Bold(true)
			}
			b.WriteString(style.Render(string(ch)))
		}
		b.WriteString("\n")
	}

	b.WriteString("\n")
	b.WriteString(taglineStyle.Render(
		"            StrangeDB — bending consistency through time\n",
	))
	b.WriteString("\n")

	// Show startup info
	if m.mode == "cluster" {
		b.WriteString(grayStyle.Render("            🚀 Starting 3-node cluster for quorum...\n"))
		b.WriteString(grayStyle.Render("               Node 1: localhost:9000 / 9001\n"))
		b.WriteString(grayStyle.Render("               Node 2: localhost:9010 / 9011\n"))
		b.WriteString(grayStyle.Render("               Node 3: localhost:9020 / 9021\n"))
	} else {
		b.WriteString(grayStyle.Render("            🚀 Starting single node...\n"))
		b.WriteString(grayStyle.Render("               Node: localhost:9000 / 9001\n"))
	}

	b.WriteString("\n")
	progressLen := m.step * 20 / m.maxSteps
	if progressLen > 20 {
		progressLen = 20
	}
	if progressLen < 0 {
		progressLen = 0
	}
	remainingLen := 20 - progressLen
	progress := strings.Repeat("█", progressLen)
	remaining := strings.Repeat("░", remainingLen)
	b.WriteString(grayStyle.Render("            [" + progress + remaining + "] Starting...\n"))
	b.WriteString("\n")
	b.WriteString(grayStyle.Render("            Press ENTER to skip • Q to quit"))

	return b.String()
}

func (m SplashModel) IsDone() bool {
	return m.done
}
