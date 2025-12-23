package cli

import (
	"flag"
	"fmt"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	urls := flag.String("urls", "http://localhost:9000", "Comma-seprated node urls")
	flag.Parse()

	nodeURLs := strings.Split(*urls, ",")

	model := NewModel(nodeURLs)

	p := tea.NewProgram(model, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
