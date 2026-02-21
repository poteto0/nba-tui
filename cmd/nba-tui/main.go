package main

import (
	"flag"
	"fmt"
	"os"

	"nba-tui/internal/nba"
	"nba-tui/internal/ui/game_detail"
	"nba-tui/internal/ui/root"

	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	mock := flag.Bool("mock", false, "Use mock data for testing")
	noDeco := flag.Bool("no-decoration", false, "Disable color decorations")
	reload := flag.Int("reload", 30, "Reload interval in seconds (min 10s)")
	flag.Parse()

	if *reload < 10 {
		*reload = 10
	}

	var client root.Client
	if *mock {
		client = nba.NewMockClient()
	} else {
		client = nba.NewClient()
	}

	config := game_detail.Config{
		NoDecoration: *noDeco,
	}
	m := root.NewModel(client, config, *reload)

	p := tea.NewProgram(m, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Printf("there's been an error: %v", err)
		os.Exit(1)
	}
}
