package main

import (
	"fmt"
	"os"

	"nba-tui/internal/nba"
	"nba-tui/internal/ui/root"

	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	client := nba.NewClient()
	m := root.NewModel(client)

	p := tea.NewProgram(m, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Printf("there's been an error: %v", err)
		os.Exit(1)
	}
}
