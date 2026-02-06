package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"nba-tui/internal/nba"
	"nba-tui/internal/ui/scoreboard"
)

func main() {
	client := nba.NewClient()
	m := scoreboard.NewModel(client)

	p := tea.NewProgram(m, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Alas, there's been an error: %v", err)
		os.Exit(1)
	}
}
