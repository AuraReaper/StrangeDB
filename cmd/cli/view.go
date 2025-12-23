package cli

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

var (
	primaryColor   = lipgloss.Color("#7C3AED")
	secondaryColor = lipgloss.Color("#10B981")
	dangerColor    = lipgloss.Color("#EF4444")
	mutedColor     = lipgloss.Color("#6B7280")

	titleStyle     = lipgloss.NewStyle().Bold(true).Foreground(primaryColor)
	tabStyle       = lipgloss.NewStyle().Padding(0, 2).Foreground(mutedColor)
	activeTabStyle = lipgloss.NewStyle().Padding(0, 2).Foreground(primaryColor).Bold(true).Underline(true)
	boxStyle       = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(primaryColor).Padding(1, 2)
	healthyStyle   = lipgloss.NewStyle().Foreground(secondaryColor)
	unhealthyStyle = lipgloss.NewStyle().Foreground(dangerColor)
)

func (m Model) View() string {
	var b strings.Builder

	b.WriteString(m.renderLogo())
	b.WriteString("\n")

	b.WriteString(m.renderTabs())
	b.WriteString("\n\n")

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

	b.WriteString("\n")
	b.WriteString(m.renderFooter())

	return b.String()
}

func (m Model) renderLogo() string {
	logo := `
███████╗████████╗██████╗  █████╗ ███╗   ██╗ ██████╗ ███████╗
██╔════╝╚══██╔══╝██╔══██╗██╔══██╗████╗  ██║██╔════╝ ██╔════╝
███████╗   ██║   ██████╔╝███████║██╔██╗ ██║██║  ███╗█████╗
╚════██║   ██║   ██╔══██╗██╔══██║██║╚██╗██║██║   ██║██╔══╝
███████║   ██║   ██║  ██║██║  ██║██║ ╚████║╚██████╔╝███████╗
╚══════╝   ╚═╝   ╚═╝  ╚═╝╚═╝  ╚═╝╚═╝  ╚═══╝ ╚═════╝ ╚══════╝
                                                         DB`
	return titleStyle.Render(logo)
}

func (m Model) renderTabs() string {
	tabs := []string{"[1] Cluster", "[2] Keys", "[3] Metrics", "[?] Help"}
	var rendered []string

	for i, t := range tabs {
		if Tab(i) == m.activeTab {
			rendered = append(rendered, activeTabStyle.Render(t))
		} else {
			rendered = append(rendered, tabStyle.Render(t))
		}
	}

	return lipgloss.JoinHorizontal(lipgloss.Top, rendered...)
}

func (m Model) renderClusterView() string {
	if m.loading {
		return m.spinner.View() + " Loading cluster status..."
	}

	if m.clusterData == nil {
		return "No cluster data"
	}

	var b strings.Builder
	b.WriteString(titleStyle.Render("Cluster Status"))
	b.WriteString("\n\n")

	for _, node := range m.clusterData.Nodes {
		status := "●"
		if node.Status == "healthy" {
			status = healthyStyle.Render(status)
		} else {
			status = unhealthyStyle.Render(status)
		}

		b.WriteString(fmt.Sprintf("%s  %s  Keys: %d\n",
			status, node.URL, node.Keys))
	}

	return boxStyle.Render(b.String())
}

func (m Model) renderKeysView() string {
	var b strings.Builder
	b.WriteString(titleStyle.Render("Key Operations"))
	b.WriteString("\n\n")

	b.WriteString("Key: ")
	b.WriteString(m.keyInput.View())
	b.WriteString("\n\n")

	b.WriteString("Press Enter to GET, Ctrl+S to SET, Ctrl+D to DELETE")

	return boxStyle.Render(b.String())
}

func (m Model) renderMetricsView() string {
	return boxStyle.Render("Metrics view - Coming soon!")
}

func (m Model) renderHelpView() string {
	help := `
Keyboard Shortcuts:
  1, 2, 3    Switch tabs
  Tab        Next tab
  q          Quit
  ?          Help

Key Operations:
  Enter      Get key
  Ctrl+S     Set key
  Ctrl+D     Delete key
`
	return boxStyle.Render(help)
}

func (m Model) renderFooter() string {
	return lipgloss.NewStyle().
		Foreground(mutedColor).
		Render("Press q to quit")
}
