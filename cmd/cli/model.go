package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
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
	TabMetrics // Only shown in admin mode
	TabHelp
)

type InputState int

const (
	InputNone InputState = iota
	InputGetKey
	InputSetKey
	InputSetValue
	InputDeleteKey
	InputConfirmDelete
)

type Model struct {
	nodeURLs   []string
	adminMode  bool
	activeTab  Tab
	width      int
	height     int
	loading    bool
	spinner    spinner.Model
	textInput  textinput.Model
	keyBuffer  string // Store key when entering value
	inputState InputState

	// Data
	clusterData *ClusterData
	lastResult  string
	lastError   string
	err         error
}

type ClusterData struct {
	NodeID  string       `json:"node_id"`
	Members []MemberInfo `json:"members"`
	Total   int          `json:"total"`
}

type MemberInfo struct {
	NodeID string `json:"node_id"`
	Addr   string `json:"addr"`
	Status string `json:"status"`
}

type GetKeyResponse struct {
	Key      string `json:"key"`
	Value    string `json:"value"`
	ValueRaw string `json:"value_base64"`
}

func NewModel(nodeURLs []string, adminMode bool) Model {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("#7C3AED"))

	ti := textinput.New()
	ti.CharLimit = 256
	ti.Width = 50

	return Model{
		nodeURLs:   nodeURLs,
		adminMode:  adminMode,
		activeTab:  TabCluster,
		spinner:    s,
		textInput:  ti,
		loading:    true,
		inputState: InputNone,
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
		key := msg.String()

		// Handle input states first
		if m.inputState != InputNone {
			switch key {
			case "esc":
				m.inputState = InputNone
				m.textInput.Blur()
				m.lastError = ""
				return m, nil

			case "enter":
				return m.handleInputSubmit()

			default:
				var cmd tea.Cmd
				m.textInput, cmd = m.textInput.Update(msg)
				return m, cmd
			}
		}

		// Global keys
		switch key {
		case "ctrl+c", "q":
			return m, tea.Quit

		case "1":
			m.activeTab = TabCluster
		case "2":
			m.activeTab = TabKeys
		case "3":
			if m.adminMode {
				m.activeTab = TabMetrics
			}
		case "?", "h":
			m.activeTab = TabHelp
		case "tab":
			maxTabs := 3
			if m.adminMode {
				maxTabs = 4
			}
			m.activeTab = Tab((int(m.activeTab) + 1) % maxTabs)
		}

		// Keys tab actions
		if m.activeTab == TabKeys {
			switch key {
			case "g":
				m.startInput(InputGetKey, "Enter key to GET:")
				return m, nil
			case "s":
				m.startInput(InputSetKey, "Enter key to SET:")
				return m, nil
			case "d":
				m.startInput(InputDeleteKey, "Enter key to DELETE:")
				return m, nil
			}
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case clusterDataMsg:
		m.loading = false
		m.clusterData = &msg.data
		m.err = nil
		return m, m.tick()

	case errMsg:
		m.loading = false
		m.err = msg.err
		return m, m.tick()

	case keyOperationResultMsg:
		if msg.err != nil {
			m.lastError = msg.err.Error()
			m.lastResult = ""
		} else {
			m.lastResult = msg.result
			m.lastError = ""
		}
		return m, nil

	case tickMsg:
		return m, m.fetchClusterData()

	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	}

	return m, nil
}

func (m *Model) startInput(state InputState, placeholder string) {
	m.inputState = state
	m.textInput.Placeholder = placeholder
	m.textInput.SetValue("")
	m.textInput.Focus()
	m.lastError = ""
	m.lastResult = ""
}

func (m Model) handleInputSubmit() (tea.Model, tea.Cmd) {
	value := m.textInput.Value()

	switch m.inputState {
	case InputGetKey:
		if value == "" {
			m.lastError = "Key cannot be empty"
			return m, nil
		}
		m.inputState = InputNone
		m.textInput.Blur()
		return m, m.doGet(value)

	case InputSetKey:
		if value == "" {
			m.lastError = "Key cannot be empty"
			return m, nil
		}
		m.keyBuffer = value
		m.inputState = InputSetValue
		m.textInput.Placeholder = fmt.Sprintf("Enter value for '%s':", value)
		m.textInput.SetValue("")
		return m, nil

	case InputSetValue:
		m.inputState = InputNone
		m.textInput.Blur()
		return m, m.doSet(m.keyBuffer, value)

	case InputDeleteKey:
		if value == "" {
			m.lastError = "Key cannot be empty"
			return m, nil
		}
		m.inputState = InputNone
		m.textInput.Blur()
		return m, m.doDelete(value)
	}

	return m, nil
}

// Messages
type clusterDataMsg struct {
	data ClusterData
}

type errMsg struct {
	err error
}

type tickMsg struct{}

type keyOperationResultMsg struct {
	result string
	err    error
}

// Commands
func (m Model) fetchClusterData() tea.Cmd {
	return func() tea.Msg {
		if len(m.nodeURLs) == 0 {
			return errMsg{err: fmt.Errorf("no nodes configured")}
		}

		client := &http.Client{Timeout: 2 * time.Second}

		// Query all configured nodes directly for more accurate view
		members := make([]MemberInfo, 0)
		var primaryNodeID string

		for _, nodeURL := range m.nodeURLs {
			// Check health of each node
			healthURL := nodeURL + "/health"
			resp, err := client.Get(healthURL)
			if err != nil {
				members = append(members, MemberInfo{
					Addr:   nodeURL,
					Status: "unreachable",
				})
				continue
			}
			resp.Body.Close()

			// Get node info
			statusURL := nodeURL + "/api/v1/status"
			resp, err = client.Get(statusURL)
			if err != nil {
				members = append(members, MemberInfo{
					Addr:   nodeURL,
					Status: "alive",
				})
				continue
			}

			var statusResp struct {
				NodeID string `json:"node_id"`
				Status string `json:"status"`
			}
			json.NewDecoder(resp.Body).Decode(&statusResp)
			resp.Body.Close()

			if primaryNodeID == "" {
				primaryNodeID = statusResp.NodeID
			}

			members = append(members, MemberInfo{
				NodeID: statusResp.NodeID,
				Addr:   nodeURL,
				Status: "alive",
			})
		}

		if len(members) == 0 {
			return errMsg{err: fmt.Errorf("all nodes unreachable")}
		}

		return clusterDataMsg{data: ClusterData{
			NodeID:  primaryNodeID,
			Members: members,
			Total:   len(members),
		}}
	}
}

func (m Model) doGet(key string) tea.Cmd {
	return func() tea.Msg {
		client := &http.Client{Timeout: 5 * time.Second}

		for _, nodeURL := range m.nodeURLs {
			url := fmt.Sprintf("%s/api/v1/kv/%s", nodeURL, key)
			resp, err := client.Get(url)
			if err != nil {
				continue
			}
			defer resp.Body.Close()

			// Check for 404 or other error status
			if resp.StatusCode == 404 {
				return keyOperationResultMsg{err: fmt.Errorf("key '%s' not found in database", key)}
			}
			if resp.StatusCode >= 400 {
				body, _ := io.ReadAll(resp.Body)
				return keyOperationResultMsg{err: fmt.Errorf("server error: %s", string(body))}
			}

			body, _ := io.ReadAll(resp.Body)
			var result GetKeyResponse
			json.Unmarshal(body, &result)

			// Check for empty value
			if result.Value == "" && result.Key == "" {
				return keyOperationResultMsg{err: fmt.Errorf("key '%s' not found in database", key)}
			}

			return keyOperationResultMsg{
				result: fmt.Sprintf("📥 GET '%s'\n   Value: %s", key, result.Value),
			}
		}

		return keyOperationResultMsg{err: fmt.Errorf("operation failed - no nodes available")}
	}
}

func (m Model) doSet(key, value string) tea.Cmd {
	return func() tea.Msg {
		client := &http.Client{Timeout: 5 * time.Second}

		for _, nodeURL := range m.nodeURLs {
			url := fmt.Sprintf("%s/api/v1/kv", nodeURL)
			payload := map[string]string{"key": key, "value": value}
			jsonData, _ := json.Marshal(payload)

			resp, err := client.Post(url, "application/json", bytes.NewBuffer(jsonData))
			if err != nil {
				continue
			}
			defer resp.Body.Close()

			if resp.StatusCode >= 400 {
				body, _ := io.ReadAll(resp.Body)
				return keyOperationResultMsg{err: fmt.Errorf("server error: %s", string(body))}
			}

			return keyOperationResultMsg{
				result: fmt.Sprintf("📤 SET '%s' = '%s'\n   ✅ Stored successfully!", key, value),
			}
		}

		return keyOperationResultMsg{err: fmt.Errorf("operation failed - no nodes available")}
	}
}

func (m Model) doDelete(key string) tea.Cmd {
	return func() tea.Msg {
		client := &http.Client{Timeout: 5 * time.Second}

		for _, nodeURL := range m.nodeURLs {
			// First check if key exists
			getURL := fmt.Sprintf("%s/api/v1/kv/%s", nodeURL, key)
			getResp, err := client.Get(getURL)
			if err != nil {
				continue
			}
			getResp.Body.Close()

			if getResp.StatusCode == 404 {
				return keyOperationResultMsg{err: fmt.Errorf("key '%s' does not exist - nothing to delete", key)}
			}

			// Key exists, proceed with delete
			url := fmt.Sprintf("%s/api/v1/kv/%s", nodeURL, key)
			req, _ := http.NewRequest("DELETE", url, nil)
			resp, err := client.Do(req)
			if err != nil {
				continue
			}
			resp.Body.Close()

			if resp.StatusCode >= 400 {
				return keyOperationResultMsg{err: fmt.Errorf("delete failed")}
			}

			return keyOperationResultMsg{
				result: fmt.Sprintf("🗑️  DELETE '%s'\n   ✅ Deleted successfully!", key),
			}
		}

		return keyOperationResultMsg{err: fmt.Errorf("operation failed - no nodes available")}
	}
}

func (m Model) tick() tea.Cmd {
	return tea.Tick(5*time.Second, func(t time.Time) tea.Msg {
		return tickMsg{}
	})
}
