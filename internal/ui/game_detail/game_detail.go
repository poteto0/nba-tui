package game_detail

import (
	"fmt"
	"os/exec"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/poteto0/go-nba-sdk/types"
)

type focusArea int

const (
	boxScoreFocus focusArea = iota
	gameLogFocus
)

var (
	borderStyle = lipgloss.NewStyle().
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("240"))

	activeBorderStyle = lipgloss.NewStyle().
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("2")) // Green

	tableHeaderStyle = lipgloss.NewStyle().
		Bold(true).
		Border(lipgloss.NormalBorder(), false, false, true, false).
		BorderForeground(lipgloss.Color("240"))

	boldStyle = lipgloss.NewStyle().Bold(true)

	faintStyle = lipgloss.NewStyle().Faint(true)

	underlineStyle = lipgloss.NewStyle().Underline(true).Foreground(lipgloss.Color("2"))
)

type NbaClient interface {
	GetBoxScore(gameID string) (types.LiveBoxScoreResponse, error)
	GetPlayByPlay(gameID string) (types.LivePlayByPlayResponse, error)
}

type Model struct {
	client         NbaClient
	gameID         string
	boxScore       types.LiveBoxScoreResponse
	pbp            types.LivePlayByPlayResponse
	showingHome    bool
	lastUpdated    time.Time
	focus          focusArea
	logOffset      int
	boxOffset      int
	selectedPeriod int
}

func New(client NbaClient, gameID string) Model {
	return Model{client: client, gameID: gameID, showingHome: true, selectedPeriod: 1}
}

func (m *Model) SetLastUpdated(t time.Time) {
	m.lastUpdated = t
}

func (m Model) IsShowingHome() bool {
	return m.showingHome
}

func (m Model) GetFocus() int {
	return int(m.focus)
}

func (m Model) GetLogOffset() int {
	return m.logOffset
}

func (m Model) GetBoxOffset() int {
	return m.boxOffset
}

func (m Model) GetSelectedPeriod() int {
	return m.selectedPeriod
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(
		m.fetchBoxScore,
		m.fetchPlayByPlay,
	)
}

func (m Model) fetchBoxScore() tea.Msg {
	res, err := m.client.GetBoxScore(m.gameID)
	if err != nil {
		return ErrorMsg(err)
	}
	return BoxScoreMsg(res)
}

func (m Model) fetchPlayByPlay() tea.Msg {
	res, err := m.client.GetPlayByPlay(m.gameID)
	if err != nil {
		return ErrorMsg(err)
	}
	return PlayByPlayMsg(res)
}

type BoxScoreMsg types.LiveBoxScoreResponse
type PlayByPlayMsg types.LivePlayByPlayResponse
type ErrorMsg error

func (m Model) getCurrentTeam() types.Team {
	if m.showingHome {
		return m.boxScore.Game.HomeTeam
	}
	return m.boxScore.Game.AwayTeam
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case BoxScoreMsg:
		m.boxScore = types.LiveBoxScoreResponse(msg)
		m.lastUpdated = time.Now()
	case PlayByPlayMsg:
		m.pbp = types.LivePlayByPlayResponse(msg)
		m.lastUpdated = time.Now()
	case tea.KeyMsg:
		team := m.getCurrentTeam()
		switch msg.String() {
		case "ctrl+s":
			m.showingHome = !m.showingHome
			m.logOffset = 0
		case "ctrl+c":
			return m, tea.Quit
		case "ctrl+q":
			m.selectedPeriod++
			if m.selectedPeriod > 4 {
				m.selectedPeriod = 1
			}
			m.logOffset = 0
		case "ctrl+w":
			url := fmt.Sprintf("https://www.nba.com/game/%s", m.gameID)
			_ = exec.Command("xdg-open", url).Start()
		case "ctrl+b":
			m.focus = boxScoreFocus
		case "ctrl+l":
			m.focus = gameLogFocus
		case "h", "left":
			// Horizontal scroll if needed in future
		case "l", "right":
			// Horizontal scroll if needed in future
		case "j", "down":
			if m.focus == boxScoreFocus {
				players, err := team.Players.Take()
				if err == nil {
					if m.boxOffset < len(players)-1 {
						m.boxOffset++
					}
				}
			} else {
				var filteredActions []types.Action
				for _, action := range m.pbp.Game.Actions {
					if action.Period == m.selectedPeriod && action.TeamID == team.TeamId {
						filteredActions = append(filteredActions, action)
					}
				}
				if m.logOffset < len(filteredActions)-1 {
					m.logOffset++
				}
			}
		case "k", "up":
			if m.focus == boxScoreFocus {
				if m.boxOffset > 0 {
					m.boxOffset--
				}
			} else {
				if m.logOffset > 0 {
					m.logOffset--
				}
			}
		}
	}
	return m, nil
}

func (m Model) View() string {
	if m.boxScore.Game.GameId == "" {
		return "Loading..."
	}

	// Header: Status and Teams
	game := m.boxScore.Game
	status := game.GameStatusText
	if !game.IsFinished() {
		status = fmt.Sprintf("%dQ (%s)", game.Period, game.GameClock)
	}

	homeTricode := game.HomeTeam.TeamTricode
	homeScore := fmt.Sprintf("%d", game.HomeTeam.Score)
	awayTricode := game.AwayTeam.TeamTricode
	awayScore := fmt.Sprintf("%d", game.AwayTeam.Score)

	if game.HomeTeam.Score > game.AwayTeam.Score {
		homeTricode = boldStyle.Render(homeTricode)
		homeScore = boldStyle.Render(homeScore)
	} else if game.AwayTeam.Score > game.HomeTeam.Score {
		awayTricode = boldStyle.Render(awayTricode)
		awayScore = boldStyle.Render(awayScore)
	}

	headerStr := fmt.Sprintf("%s\n%s (%s) | %s (%s)",
		status,
		homeTricode, homeScore,
		awayTricode, awayScore,
	)
	header := borderStyle.Width(50).Align(lipgloss.Center).Render(headerStr)

	// Selected Team Display
	team := m.getCurrentTeam()
	selectedTeamView := fmt.Sprintf("Selected Team: %s", team.TeamTricode)

	// Box Score Section
	boxScoreContent := m.renderBoxScore(team)
	bsStyle := borderStyle
	if m.focus == boxScoreFocus {
		bsStyle = activeBorderStyle
	}
	// Fixed size box score
	boxScore := bsStyle.Width(50).Height(26).MaxWidth(50).MaxHeight(26).Render(boxScoreContent)

	// Game Log Section
	// Period Selector
	periods := []string{"1Q", "2Q", "3Q", "4Q"}
	var selectorParts []string
	for i, p := range periods {
		pNum := i + 1
		if pNum == m.selectedPeriod {
			selectorParts = append(selectorParts, underlineStyle.Render(p))
		} else {
			selectorParts = append(selectorParts, faintStyle.Render(p))
		}
	}
	periodSelectorContent := strings.Join(selectorParts, " | ")
	periodSelector := lipgloss.NewStyle().Width(33).Align(lipgloss.Center).Render(periodSelectorContent)

	gameLogHeader := lipgloss.NewStyle().Width(33).Align(lipgloss.Center).Render("gamelog")

	// Filter actions by team and period
	var filteredActions []types.Action
	for _, action := range m.pbp.Game.Actions {
		if action.Period == m.selectedPeriod && action.TeamID == team.TeamId {
			filteredActions = append(filteredActions, action)
		}
	}

	// Create a fixed-height log body (22 lines to fit in Height(26) minus headers)
	var logLines []string
	for i := 0; i < 22; i++ {
		idx := m.logOffset + i
		if idx < len(filteredActions) {
			action := filteredActions[idx]
			desc := action.Description
			// Truncate to avoid line wrap breaking the layout
			if len(desc) > 26 {
				desc = desc[:23] + "..."
			}
			logLines = append(logLines, fmt.Sprintf("% -5s|%s", action.Clock, desc))
		} else {
			logLines = append(logLines, "") // Pad with empty lines
		}
	}
	gameLogBody := strings.Join(logLines, "\n")
	gameLogContent := gameLogHeader + "\n" + periodSelector + "\n" + gameLogBody

	glStyle := borderStyle
	if m.focus == gameLogFocus {
		glStyle = activeBorderStyle
	}
	gameLog := glStyle.Width(35).Height(26).MaxWidth(35).MaxHeight(26).Render(gameLogContent)

	leftCol := lipgloss.JoinVertical(lipgloss.Left, header, boxScore)
	mainView := lipgloss.JoinHorizontal(lipgloss.Top, leftCol, gameLog)
	rootView := lipgloss.JoinVertical(lipgloss.Left, selectedTeamView, mainView)

	helpText := "<hjkli←↓↑→ >: move, <ctrl+s>: switch team, <ctrl+b>: box, <ctrl+l>: log, <ctrl+w>: watch, <esc>: back"
	footer := helpText
	if !m.lastUpdated.IsZero() {
		footer = fmt.Sprintf("Last updated: %s\n%s", m.lastUpdated.Format(time.RFC1123), helpText)
	}

	return lipgloss.JoinVertical(lipgloss.Left, rootView, footer)
}

func (m Model) renderBoxScore(team types.Team) string {
	s := "Box Scores\n"
	s += tableHeaderStyle.Render(fmt.Sprintf("% -15s % -5s % -3s % -3s % -3s", "PLAYER", "MIN", "PTS", "REB", "AST")) + "\n"

	players, err := team.Players.Take()
	if err != nil {
		return s + "No player data"
	}

	// Calculate how many lines we can show in the fixed height
	// Table takes up 2 lines (Title + Header)
	// Available is 26 minus borders (2) minus headers (2) = 22 lines.
	for i := 0; i < 22; i++ {
		idx := m.boxOffset + i
		if idx < len(players) {
			p := players[idx]
			name := ""
			if len(p.FirstName) > 0 {
				name = fmt.Sprintf("%s.%s", string(p.FirstName[0]), p.FamilyName)
			} else {
				name = p.FamilyName
			}
			stats, err := p.Statistics.Take()
			if err != nil {
				s += fmt.Sprintf("% -15s -\n", name)
				continue
			}

			min, _ := stats.Minutes.Take()
			pts, _ := stats.Pts.Take()
			reb, _ := stats.Reb.Take()
			ast, _ := stats.Ast.Take()

			s += fmt.Sprintf("% -15s % -5s % -3d % -3d % -3d\n",
				name, min, pts, reb, ast)
		} else {
			s += "\n" // Pad with empty lines
		}
	}
	return s
}