package scoreboard

import (
	"fmt"
	"nba-tui/internal/ui/styles"
	"os/exec"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/poteto0/go-nba-sdk/types"
)

type GotScoreboardMsg struct {
	Games []types.Game
}

type SelectGameMsg struct {
	GameId string
}

type tickMsg time.Time

func tick() tea.Cmd {
	return tea.Tick(30*time.Second, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

type ScoreboardProvider interface {
	GetScoreboard() ([]types.Game, error)
}

type Model struct {
	client      ScoreboardProvider
	Games       []types.Game
	Focus       int
	Err         error
	LastUpdated time.Time
	Width       int
	Height      int
	Columns     int
	OpenBrowser func(string) error
}

func NewModel(client ScoreboardProvider) Model {
	return Model{
		client:  client,
		Columns: 1, // Default to 1 column
		OpenBrowser: func(url string) error {
			return exec.Command("xdg-open", url).Start()
		},
	}
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(m.FetchScoreboard(), tick())
}

func (m Model) FetchScoreboard() tea.Cmd {
	return func() tea.Msg {
		games, err := m.client.GetScoreboard()
		if err != nil {
			return err
		}
		return GotScoreboardMsg{Games: games}
	}
}

func (m *Model) calculateColumns() {
	if m.Width == 0 {
		m.Columns = 1
		return
	}
	// Estimate box width.
	// Status (11) + padding + borders approx 2 chars.
	// Let's check the View implementation.
	// center(status, 11) -> 11 chars width.
	// Box adds border (2) + padding?
	// 11 chars + 2 border = 13.
	// The widest part is "POR | DEN" = 9 chars.
	// "---------" = 9 chars.
	// Actually lipgloss border adds 2 chars width.
	// Let's assume a safe width of around 18-20 chars including spacing between boxes.
	// Let's conservatively say 18 chars per box.

	boxWidth := 18
	cols := m.Width / boxWidth
	if cols < 1 {
		cols = 1
	}
	m.Columns = cols
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.Width = msg.Width
		m.Height = msg.Height
		m.calculateColumns()
		return m, nil
	case error:
		m.Err = msg
		return m, nil
	case GotScoreboardMsg:
		m.Games = msg.Games
		m.LastUpdated = time.Now()
		return m, nil
	case tickMsg:
		return m, tea.Batch(m.FetchScoreboard(), tick())
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q", "esc":
			return m, tea.Quit
		case "enter":
			if len(m.Games) > 0 {
				return m, func() tea.Msg {
					return SelectGameMsg{GameId: m.Games[m.Focus].GameId}
				}
			}
		case "ctrl+w":
			if len(m.Games) > 0 {
				game := m.Games[m.Focus]
				url := fmt.Sprintf("https://www.nba.com/game/%s", game.GameId)
				if m.OpenBrowser != nil {
					_ = m.OpenBrowser(url)
				}
			}
		case "h", "left":
			if m.Focus > 0 {
				m.Focus--
			}
		case "l", "right":
			if m.Focus < len(m.Games)-1 {
				m.Focus++
			}
		case "k", "up":
			if m.Focus >= m.Columns {
				m.Focus -= m.Columns
			}
		case "j", "down":
			if m.Focus+m.Columns < len(m.Games) {
				m.Focus += m.Columns
			}
		}
	}
	return m, nil
}

func formatScore(score int) string {
	s := fmt.Sprintf("%d", score)
	if len(s) == 1 {
		return " " + s + " "
	}
	if len(s) == 2 {
		return " " + s
	}
	return s
}

func (m Model) View() string {
	if m.Err != nil {
		return fmt.Sprintf("Error: %v", m.Err)
	}

	helpText := "<hjkli←↓↑→ >: move, <enter>: detail, <ctrl+w>: watch (browser), <q/esc>: quit"
	if !m.LastUpdated.IsZero() {
		helpText = fmt.Sprintf("Last updated: %s\n%s", m.LastUpdated.Format(time.RFC1123), helpText)
	}

	if len(m.Games) == 0 {
		return helpText + "\n\nLoading..."
	}

	// Calculate columns if not set (first render)
	if m.Columns == 0 {
		m.Columns = 1 // Safe default
	}

	var boards []string
	for i, game := range m.Games {
		style := styles.InactiveBorderStyle
		if i == m.Focus {
			style = styles.ActiveBorderStyle
		}

		status := ""
		if game.IsFinished() {
			status = "Final"
		} else {
			status = fmt.Sprintf("%dQ (%s)", game.Period, game.GameClock)
		}
		homeName := game.HomeTeam.TeamTricode
		awayName := game.AwayTeam.TeamTricode
		homeScoreStr := formatScore(game.HomeTeam.Score)
		awayScoreStr := formatScore(game.AwayTeam.Score)

		if game.HomeTeam.Score > game.AwayTeam.Score {
			homeName = styles.BoldStyle.Render(homeName)
			homeScoreStr = styles.BoldStyle.Render(homeScoreStr)
		} else if game.AwayTeam.Score > game.HomeTeam.Score {
			awayName = styles.BoldStyle.Render(awayName)
			awayScoreStr = styles.BoldStyle.Render(awayScoreStr)
		}

		content := fmt.Sprintf(
			"%s\n%s | %s\n ---------\n%s | %s",
			center(status, 11),
			homeName, awayName,
			homeScoreStr, awayScoreStr,
		)

		boards = append(boards, style.Render(content))
	}

	// Grid layout using m.Columns
	var rows []string
	for i := 0; i < len(boards); i += m.Columns {
		end := i + m.Columns
		if end > len(boards) {
			end = len(boards)
		}
		row := lipgloss.JoinHorizontal(lipgloss.Top, boards[i:end]...)
		rows = append(rows, row)
	}

	scoreboardView := lipgloss.JoinVertical(lipgloss.Left, rows...)
	return lipgloss.JoinVertical(lipgloss.Left, helpText, scoreboardView)
}

func center(s string, width int) string {
	padding := width - lipgloss.Width(s)
	if padding <= 0 {
		return s
	}
	left := padding / 2
	right := padding - left
	return strings.Repeat(" ", left) + s + strings.Repeat(" ", right)
}
