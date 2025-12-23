package cli

import (
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type Tab int

const (
	TabCluster Tab = iota
	TabKeys
	TabMetrics
	TabHelp
)

type Model struct {
	nodeURLs    []string
	activeTab   Tab
	width       int
	height      int
	loading     bool
	spinner     spinner.Model
	keyInput    textinput.Model
	valueInput  textinput.Model
	clusterData *ClusterData
	err         error
}

type ClusterData struct {
	Nodes []NodeStatus
}

type NodeStatus struct {
	URL    string
	Status string
	Keys   int
}

func NewModel(nodeURLs []string) Model {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))

	ki := textinput.New()
	ki.Placeholder = "Enter key..."
	ki.Focus()

	vi := textinput.New()
	vi.Placeholder = "Enter value..."

	return Model{
		nodeURLs:   nodeURLs,
		activeTab:  TabCluster,
		spinner:    s,
		keyInput:   ki,
		valueInput: vi,
		loading:    true,
	}
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(
		m.spinner.Tick,
		m.fetchClusterData(),
	)
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "tab":
			m.activeTab = (m.activeTab + 1) % 4
		case "1":
			m.activeTab = TabCluster
		case "2":
			m.activeTab = TabKeys
		case "3":
			m.activeTab = TabMetrics
		case "?":
			m.activeTab = TabHelp
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case clusterDataMsg:
		m.loading = false
		m.clusterData = &msg.data
		return m, m.tick()

	case errMsg:
		m.loading = false
		m.err = msg.err
		return m, m.tick()

	case tickMsg:
		return m, m.fetchClusterData()

	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	}

	if m.activeTab == TabKeys {
		var cmd tea.Cmd
		m.keyInput, cmd = m.keyInput.Update(msg)
		return m, cmd
	}

	return m, nil
}

type clusterDataMsg struct {
	data ClusterData
}

type errMsg struct {
	err error
}

type tickMsg struct{}

func (m Model) fetchClusterData() tea.Cmd {
	return func() tea.Msg {
		data := ClusterData{
			Nodes: []NodeStatus{
				{URL: m.nodeURLs[0], Status: "healthy", Keys: 1234},
			},
		}
		return clusterDataMsg{data: data}
	}
}

func (m Model) tick() tea.Cmd {
	return tea.Tick(5*time.Second, func(t time.Time) tea.Msg {
		return tickMsg{}
	})
}
