package main

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

var (
	// Colors - Purple theme
	purple = lipgloss.Color("#7C3AED")
	pink   = lipgloss.Color("#EC4899")
	cyan   = lipgloss.Color("#06B6D4")
	green  = lipgloss.Color("#10B981")
	red    = lipgloss.Color("#EF4444")
	yellow = lipgloss.Color("#F59E0B")
	gray   = lipgloss.Color("#6B7280")
	darkBg = lipgloss.Color("#1F2937")
	white  = lipgloss.Color("#F9FAFB")

	// Styles
	logoStyle      = lipgloss.NewStyle().Foreground(purple).Bold(true)
	titleStyle     = lipgloss.NewStyle().Foreground(white).Bold(true)
	subtitleStyle  = lipgloss.NewStyle().Foreground(gray)
	tabStyle       = lipgloss.NewStyle().Padding(0, 2).Foreground(gray)
	activeTabStyle = lipgloss.NewStyle().Padding(0, 2).Foreground(purple).Bold(true).Background(darkBg)
	boxStyle       = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(purple).Padding(1, 2)
	successStyle   = lipgloss.NewStyle().Foreground(green)
	errorStyle     = lipgloss.NewStyle().Foreground(red)
	warningStyle   = lipgloss.NewStyle().Foreground(yellow)
	mutedStyle     = lipgloss.NewStyle().Foreground(gray)
	cyanStyle      = lipgloss.NewStyle().Foreground(cyan)
	pinkStyle      = lipgloss.NewStyle().Foreground(pink)
	inputBoxStyle  = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(pink).Padding(0, 1)
	promptStyle    = lipgloss.NewStyle().Foreground(pink).Bold(true)
)

func (m Model) View() string {
	var b strings.Builder

	// Header
	b.WriteString(m.renderHeader())
	b.WriteString("\n")

	// Tabs
	b.WriteString(m.renderTabs())
	b.WriteString("\n\n")

	// Content
	switch m.activeTab {
	case TabCluster:
		b.WriteString(m.renderClusterView())
	case TabKeys:
		b.WriteString(m.renderKeysView())
	case TabMetrics:
		b.WriteString(m.renderMetricsView())
	case TabHelp:
		b.WriteString(m.renderHelpView())
	}

	// Footer
	b.WriteString("\n")
	b.WriteString(m.renderFooter())

	return b.String()
}

func (m Model) renderHeader() string {
	// Single-line compact logo
	logo := logoStyle.Render("  🔮 StrangeDB")
	version := mutedStyle.Render(" v0.1.0")
	tagline := subtitleStyle.Render(" │ Distributed Key-Value Store")

	// Dynamic status based on actual cluster data
	status := ""
	if m.loading {
		status = warningStyle.Render(" ◐")
	} else if m.err != nil {
		status = errorStyle.Render(" ● Disconnected")
	} else if m.clusterData != nil {
		aliveCount := 0
		for _, member := range m.clusterData.Members {
			if member.Status == "alive" {
				aliveCount++
			}
		}
		if aliveCount >= 2 {
			status = successStyle.Render(fmt.Sprintf(" ● %d/%d nodes", aliveCount, len(m.clusterData.Members)))
		} else if aliveCount > 0 {
			status = warningStyle.Render(fmt.Sprintf(" ● %d/%d nodes", aliveCount, len(m.clusterData.Members)))
		} else {
			status = errorStyle.Render(" ● No nodes")
		}
	}

	topLine := logo + version + tagline + status

	// Mode indicator
	modeInfo := ""
	if m.adminMode {
		modeInfo = pinkStyle.Render("  🔧 Admin Mode")
	}

	return topLine + modeInfo + "\n" + strings.Repeat("─", 75)
}

func (m Model) renderTabs() string {
	tabs := []struct {
		key     string
		label   string
		visible bool
	}{
		{"1", "Cluster", true},
		{"2", "Keys", true},
		{"3", "Metrics", m.adminMode},
		{"?", "Help", true},
	}

	var rendered []string
	tabIndex := 0
	for _, t := range tabs {
		if !t.visible {
			continue
		}
		label := fmt.Sprintf(" %s %s ", t.key, t.label)
		if Tab(tabIndex) == m.activeTab {
			rendered = append(rendered, activeTabStyle.Render(label))
		} else {
			rendered = append(rendered, tabStyle.Render(label))
		}
		if t.visible && t.key != "?" {
			tabIndex++
		} else if t.key == "?" {
			tabIndex = 3
		}
	}

	return "  " + lipgloss.JoinHorizontal(lipgloss.Top, rendered...)
}

func (m Model) renderClusterView() string {
	var b strings.Builder

	b.WriteString("  " + titleStyle.Render("📊 CLUSTER STATUS") + "\n\n")

	if m.loading {
		b.WriteString("  " + m.spinner.View() + " Connecting to cluster...\n")
		return b.String()
	}

	if m.err != nil {
		b.WriteString("  " + errorStyle.Render("❌ "+m.err.Error()) + "\n")
		b.WriteString("  " + mutedStyle.Render("Make sure StrangeDB is running") + "\n")
		return b.String()
	}

	if m.clusterData == nil {
		b.WriteString("  " + mutedStyle.Render("No cluster data") + "\n")
		return b.String()
	}

	// Cluster info table
	lines := []string{
		fmt.Sprintf("  %-15s %s", cyanStyle.Render("Node ID:"), m.clusterData.NodeID),
		fmt.Sprintf("  %-15s %d", cyanStyle.Render("Total Nodes:"), m.clusterData.Total),
	}

	quorumStatus := "❌ No quorum"
	if m.clusterData.Total >= 2 {
		quorumStatus = "✅ Quorum ready"
	}
	lines = append(lines, fmt.Sprintf("  %-15s %s", cyanStyle.Render("Quorum:"), quorumStatus))

	for _, line := range lines {
		b.WriteString(line + "\n")
	}

	// Members
	if len(m.clusterData.Members) > 0 {
		b.WriteString("\n  " + titleStyle.Render("NODES") + "\n")
		b.WriteString("  " + strings.Repeat("─", 50) + "\n")

		for _, member := range m.clusterData.Members {
			statusIcon := successStyle.Render("●")
			if member.Status != "alive" {
				statusIcon = errorStyle.Render("●")
			}
			b.WriteString(fmt.Sprintf("  %s  %-25s %s\n",
				statusIcon,
				member.Addr,
				mutedStyle.Render(member.Status)))
		}
	}

	return b.String()
}

func (m Model) renderKeysView() string {
	var b strings.Builder

	b.WriteString("  " + titleStyle.Render("🔑 KEY OPERATIONS") + "\n\n")

	// Input mode
	if m.inputState != InputNone {
		b.WriteString(m.renderInputPrompt())
		return b.String()
	}

	// Actions menu
	b.WriteString("  " + subtitleStyle.Render("Choose an operation:") + "\n\n")

	actions := []struct {
		key   string
		label string
		desc  string
	}{
		{"g", "GET", "Retrieve a value by key"},
		{"s", "SET", "Store a key-value pair"},
		{"d", "DELETE", "Remove a key"},
	}

	for _, a := range actions {
		keyStyle := lipgloss.NewStyle().Foreground(purple).Bold(true).Width(3)
		b.WriteString(fmt.Sprintf("  %s  %-10s %s\n",
			keyStyle.Render("["+a.key+"]"),
			a.label,
			mutedStyle.Render(a.desc)))
	}

	// Results
	if m.lastResult != "" {
		b.WriteString("\n  " + strings.Repeat("─", 50) + "\n")
		b.WriteString("  " + successStyle.Render(m.lastResult) + "\n")
	}
	if m.lastError != "" {
		b.WriteString("\n  " + strings.Repeat("─", 50) + "\n")
		b.WriteString("  " + errorStyle.Render("❌ "+m.lastError) + "\n")
	}

	return b.String()
}

func (m Model) renderInputPrompt() string {
	var b strings.Builder

	// Prompt label
	var promptLabel string
	switch m.inputState {
	case InputGetKey:
		promptLabel = "📥 GET Key"
	case InputSetKey:
		promptLabel = "📤 SET Key"
	case InputSetValue:
		promptLabel = fmt.Sprintf("📤 SET Value for '%s'", m.keyBuffer)
	case InputDeleteKey:
		promptLabel = "🗑️  DELETE Key"
	}

	b.WriteString("  " + promptStyle.Render(promptLabel) + "\n\n")
	b.WriteString("  " + inputBoxStyle.Render(m.textInput.View()) + "\n\n")
	b.WriteString("  " + mutedStyle.Render("Press Enter to confirm, Esc to cancel") + "\n")

	return b.String()
}

func (m Model) renderMetricsView() string {
	var b strings.Builder

	b.WriteString("  " + titleStyle.Render("📈 METRICS (Admin Only)") + "\n\n")

	b.WriteString("  " + cyanStyle.Render("Prometheus Endpoint:") + " /metrics\n\n")

	metrics := []string{
		"strangedb_requests_total",
		"strangedb_request_duration_seconds",
		"strangedb_keys_total",
		"strangedb_storage_bytes",
		"strangedb_nodes_total",
		"strangedb_gossip_messages_total",
		"strangedb_read_repairs_total",
	}

	b.WriteString("  " + subtitleStyle.Render("Available Metrics:") + "\n")
	for _, metric := range metrics {
		b.WriteString(fmt.Sprintf("    • %s\n", metric))
	}

	b.WriteString("\n  " + mutedStyle.Render("curl http://localhost:9000/metrics | grep strangedb"))

	return b.String()
}

func (m Model) renderHelpView() string {
	var b strings.Builder

	b.WriteString("  " + titleStyle.Render("⌨️  KEYBOARD SHORTCUTS") + "\n\n")

	sections := []struct {
		title    string
		bindings []struct{ key, desc string }
	}{
		{
			"Navigation",
			[]struct{ key, desc string }{
				{"1/2/3", "Switch tabs"},
				{"Tab", "Next tab"},
				{"q", "Quit"},
			},
		},
		{
			"Key Operations",
			[]struct{ key, desc string }{
				{"g", "GET a key"},
				{"s", "SET a key"},
				{"d", "DELETE a key"},
				{"Enter", "Confirm"},
				{"Esc", "Cancel"},
			},
		},
	}

	for _, section := range sections {
		b.WriteString("  " + cyanStyle.Render(section.title) + "\n")
		for _, bind := range section.bindings {
			b.WriteString(fmt.Sprintf("    %-12s %s\n",
				pinkStyle.Render(bind.key),
				bind.desc))
		}
		b.WriteString("\n")
	}

	// About
	b.WriteString("  " + cyanStyle.Render("About StrangeDB") + "\n")
	b.WriteString("    Distributed KV store with quorum reads/writes,\n")
	b.WriteString("    self-healing clusters, and AI-powered insights.\n")

	return b.String()
}

func (m Model) renderFooter() string {
	// Connection info
	nodeInfo := ""
	if len(m.nodeURLs) > 0 {
		if len(m.nodeURLs) == 1 {
			nodeInfo = m.nodeURLs[0]
		} else {
			nodeInfo = fmt.Sprintf("%d nodes", len(m.nodeURLs))
		}
	}

	left := mutedStyle.Render(fmt.Sprintf("  Connected: %s", nodeInfo))
	right := mutedStyle.Render("Press ? for help, q to quit  ")

	gap := strings.Repeat(" ", max(0, 75-lipgloss.Width(left)-lipgloss.Width(right)))

	return left + gap + right
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
