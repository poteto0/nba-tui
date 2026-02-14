package styles

import "github.com/charmbracelet/lipgloss"

var (
	BorderStyle = lipgloss.NewStyle().
			BorderStyle(lipgloss.NormalBorder()).
			BorderForeground(lipgloss.Color("240"))

	ActiveBorderStyle = lipgloss.NewStyle().
				BorderStyle(lipgloss.NormalBorder()).
				BorderForeground(lipgloss.Color("2")) // Green

	TableHeaderStyle = lipgloss.NewStyle().
				Bold(true).
				Border(lipgloss.NormalBorder(), false, false, true, false).
				BorderForeground(lipgloss.Color("240"))

	BoldStyle = lipgloss.NewStyle().Bold(true)

	FaintStyle = lipgloss.NewStyle().Faint(true)

	UnderlineStyle = lipgloss.NewStyle().Underline(true).Foreground(lipgloss.Color("2"))
)
